<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Str;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

/**
 * Webhook Delivery - individual delivery attempt.
 *
 * Tracks status, retries, and response details.
 */
class WebhookDelivery extends Model
{
    use HasFactory;

    public const STATUS_PENDING = 'pending';

    public const STATUS_QUEUED = 'queued';

    public const STATUS_SUCCESS = 'success';

    public const STATUS_FAILED = 'failed';

    public const STATUS_RETRYING = 'retrying';

    public const STATUS_CANCELLED = 'cancelled';

    public const MAX_RETRIES = 5;

    /**
     * Retry delays in minutes for each attempt.
     */
    public const RETRY_DELAYS = [
        1 => 1,      // 1 minute
        2 => 5,      // 5 minutes
        3 => 30,     // 30 minutes
        4 => 120,    // 2 hours
        5 => 1440,   // 24 hours
    ];

    protected $fillable = [
        'webhook_endpoint_id',
        'event_id',
        'event_type',
        'payload',
        'response_code',
        'response_body',
        'attempt',
        'status',
        'delivered_at',
        'next_retry_at',
    ];

    protected $casts = [
        'delivered_at' => 'datetime',
        'next_retry_at' => 'datetime',
    ];

    /**
     * Create a new delivery for an event.
     */
    public static function createForEvent(
        WebhookEndpoint $endpoint,
        string $eventType,
        array $data,
        ?int $workspaceId = null
    ): static {
        $eventId = 'evt_'.Str::random(24);
        $payload = [
            'id' => $eventId,
            'type' => $eventType,
            'created_at' => now()->toIso8601String(),
            'data' => $data,
            'workspace_id' => $workspaceId,
        ];

        try {
            $payloadJson = json_encode($payload, JSON_THROW_ON_ERROR);
        } catch (\JsonException $exception) {
            throw new \RuntimeException('Unable to encode webhook payload as JSON.', 0, $exception);
        }

        return static::create([
            'webhook_endpoint_id' => $endpoint->id,
            'event_id' => $eventId,
            'event_type' => $eventType,
            'payload' => $payloadJson,
            'status' => self::STATUS_PENDING,
            'attempt' => 1,
        ]);
    }

    /**
     * Mark as successfully delivered.
     */
    public function markSuccess(int $responseCode, ?string $responseBody = null): void
    {
        DB::transaction(function () use ($responseCode, $responseBody): void {
            $this->update([
                'status' => self::STATUS_SUCCESS,
                'response_code' => $responseCode,
                'response_body' => $responseBody ? Str::limit($responseBody, 10000) : null,
                'delivered_at' => now(),
                'next_retry_at' => null,
            ]);
        });

        $this->updateEndpointSuccess();
    }

    /**
     * Mark as failed and schedule retry if attempts remain.
     */
    public function markFailed(int $responseCode, ?string $responseBody = null): void
    {
        DB::transaction(function () use ($responseCode, $responseBody): void {
            if ($this->attempt >= self::MAX_RETRIES) {
                $this->update([
                    'status' => self::STATUS_FAILED,
                    'response_code' => $responseCode,
                    'response_body' => $responseBody ? Str::limit($responseBody, 10000) : null,
                ]);

                return;
            }

            // Schedule retry
            $nextAttempt = $this->attempt + 1;
            $delayMinutes = self::RETRY_DELAYS[$nextAttempt] ?? 1440;

            $this->update([
                'status' => self::STATUS_RETRYING,
                'response_code' => $responseCode,
                'response_body' => $responseBody ? Str::limit($responseBody, 10000) : null,
                'attempt' => $nextAttempt,
                'next_retry_at' => now()->addMinutes($delayMinutes),
            ]);
        });

        $this->updateEndpointFailure();
    }

    /**
     * Check if delivery can be retried.
     */
    public function canRetry(): bool
    {
        return $this->attempt < self::MAX_RETRIES
            && $this->status !== self::STATUS_SUCCESS;
    }

    /**
     * Get formatted payload with signature headers.
     *
     * Includes all required headers for webhook verification:
     * - X-Webhook-Signature: HMAC-SHA256 signature of timestamp.payload
     * - X-Webhook-Timestamp: Unix timestamp (for replay protection)
     * - X-Webhook-Event: The event type (e.g., 'bio.created')
     * - X-Webhook-Id: Unique delivery ID for idempotency
     *
     * ## Verification Instructions (for recipients)
     *
     * 1. Get the signature and timestamp from headers
     * 2. Compute: HMAC-SHA256(timestamp + "." + rawBody, yourSecret)
     * 3. Compare with X-Webhook-Signature using timing-safe comparison
     * 4. Verify timestamp is within 5 minutes of current time
     *
     * @param  int|null  $timestamp  Unix timestamp (defaults to current time)
     * @return array{headers: array<string, string|int>, body: string}
     */
    public function getDeliveryPayload(?int $timestamp = null): array
    {
        $timestamp ??= time();
        try {
            $jsonPayload = json_encode($this->payload, JSON_THROW_ON_ERROR);
        } catch (\JsonException $exception) {
            throw new \RuntimeException('Unable to encode webhook payload as JSON.', 0, $exception);
        }

        return [
            'headers' => [
                'Content-Type' => 'application/json',
                'X-Webhook-Id' => $this->event_id,
                'X-Webhook-Event' => $this->event_type,
                'X-Webhook-Timestamp' => (string) $timestamp,
                'X-Webhook-Signature' => $this->endpoint->generateSignature($jsonPayload, $timestamp),
            ],
            'body' => $jsonPayload,
        ];
    }

    /**
     * Decode the stored payload lazily so invalid in-memory payloads can be
     * rejected at delivery formatting time instead of during model hydration.
     */
    public function getPayloadAttribute(mixed $value): array
    {
        if (is_array($value)) {
            return $value;
        }

        if ($value === null || $value === '') {
            return [];
        }

        if (is_string($value)) {
            $decoded = json_decode($value, true, 512, JSON_THROW_ON_ERROR);

            return is_array($decoded) ? $decoded : [];
        }

        return (array) $value;
    }

    // Relationships
    public function endpoint(): BelongsTo
    {
        return $this->belongsTo(WebhookEndpoint::class, 'webhook_endpoint_id');
    }

    // Scopes
    public function scopePending($query)
    {
        return $query->where('status', self::STATUS_PENDING);
    }

    public function scopeRetrying($query)
    {
        return $query->where('status', self::STATUS_RETRYING)
            ->where('next_retry_at', '<=', now());
    }

    public function scopeNeedsDelivery($query)
    {
        return $query->where(function ($q) {
            $q->where('status', self::STATUS_PENDING)
                ->orWhere(function ($q2) {
                    $q2->where('status', self::STATUS_RETRYING)
                        ->where('next_retry_at', '<=', now());
                });
        });
    }

    /**
     * Best-effort bookkeeping for a successful delivery.
     *
     * The delivery status must not be rolled back if the endpoint record has
     * been deleted or its counter update fails.
     */
    protected function updateEndpointSuccess(): void
    {
        $this->updateEndpointState('recordSuccess', 'success');
    }

    /**
     * Best-effort bookkeeping for a failed delivery.
     *
     * The delivery status must not be rolled back if the endpoint record has
     * been deleted or its counter update fails.
     */
    protected function updateEndpointFailure(): void
    {
        $this->updateEndpointState('recordFailure', 'failure');
    }

    /**
     * Apply endpoint bookkeeping without risking the delivery state update.
     */
    protected function updateEndpointState(string $method, string $outcome): void
    {
        try {
            $endpoint = $this->endpoint;

            if (! $endpoint instanceof WebhookEndpoint) {
                Log::warning('Webhook delivery endpoint bookkeeping skipped', [
                    'delivery_id' => $this->id,
                    'webhook_endpoint_id' => $this->webhook_endpoint_id,
                    'outcome' => $outcome,
                    'reason' => 'missing_endpoint',
                ]);

                return;
            }

            $endpoint->{$method}();
        } catch (\Throwable $exception) {
            report($exception);

            Log::warning('Webhook delivery endpoint bookkeeping failed', [
                'delivery_id' => $this->id,
                'webhook_endpoint_id' => $this->webhook_endpoint_id,
                'outcome' => $outcome,
                'error_type' => $exception::class,
                'error' => $exception->getMessage(),
            ]);
        }
    }
}
