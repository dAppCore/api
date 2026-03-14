<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Api\Services\WebhookSignature;
use Core\Tenant\Models\Workspace;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\SoftDeletes;

/**
 * Webhook Endpoint - receives event notifications.
 *
 * Uses HMAC-SHA256 signatures with timestamps for security:
 * - All outbound webhooks are signed with a per-endpoint secret
 * - Timestamps prevent replay attacks (5-minute tolerance)
 * - Auto-disables after 10 consecutive delivery failures
 *
 * ## Signature Verification (for webhook recipients)
 *
 * Recipients should verify webhooks using:
 * 1. Compute: HMAC-SHA256(timestamp + "." + payload, secret)
 * 2. Compare with X-Webhook-Signature header (timing-safe)
 * 3. Verify X-Webhook-Timestamp is within 5 minutes of current time
 *
 * See WebhookSignature service for full documentation.
 */
class WebhookEndpoint extends Model
{
    use HasFactory;
    use SoftDeletes;

    /**
     * Available webhook events.
     */
    public const EVENTS = [
        // Workspace events
        'workspace.created',
        'workspace.updated',
        'workspace.deleted',

        // Subscription events
        'subscription.created',
        'subscription.updated',
        'subscription.cancelled',
        'subscription.renewed',

        // Invoice events
        'invoice.created',
        'invoice.paid',
        'invoice.failed',

        // BioLink events
        'bio.created',
        'bio.updated',
        'bio.deleted',

        // Link events
        'link.created',
        'link.updated',
        'link.deleted',
        'link.clicked', // High volume - opt-in only

        // QR Code events
        'qrcode.created',
        'qrcode.scanned', // High volume - opt-in only

        // MCP events
        'mcp.tool.executed', // Tool execution completed
    ];

    protected $fillable = [
        'workspace_id',
        'url',
        'secret',
        'events',
        'active',
        'description',
        'last_triggered_at',
        'failure_count',
        'disabled_at',
    ];

    protected $casts = [
        'events' => 'array',
        'active' => 'boolean',
        'last_triggered_at' => 'datetime',
        'disabled_at' => 'datetime',
    ];

    protected $hidden = [
        'secret',
    ];

    /**
     * Create a new webhook endpoint with auto-generated secret.
     */
    public static function createForWorkspace(
        int $workspaceId,
        string $url,
        array $events,
        ?string $description = null
    ): static {
        $signatureService = app(WebhookSignature::class);

        return static::create([
            'workspace_id' => $workspaceId,
            'url' => $url,
            'secret' => $signatureService->generateSecret(),
            'events' => $events,
            'description' => $description,
            'active' => true,
        ]);
    }

    /**
     * Generate signature for payload with timestamp.
     *
     * The signature includes the timestamp to prevent replay attacks.
     * Format: HMAC-SHA256(timestamp + "." + payload, secret)
     *
     * @param  string  $payload  The JSON-encoded webhook payload
     * @param  int  $timestamp  Unix timestamp of the request
     * @return string The hex-encoded HMAC-SHA256 signature
     */
    public function generateSignature(string $payload, int $timestamp): string
    {
        $signatureService = app(WebhookSignature::class);

        return $signatureService->sign($payload, $this->secret, $timestamp);
    }

    /**
     * Verify a signature from an incoming request (for testing endpoints).
     *
     * @param  string  $payload  The raw request body
     * @param  string  $signature  The signature from the header
     * @param  int  $timestamp  The timestamp from the header
     * @param  int  $tolerance  Maximum age in seconds (default: 300)
     * @return bool True if the signature is valid
     */
    public function verifySignature(
        string $payload,
        string $signature,
        int $timestamp,
        int $tolerance = WebhookSignature::DEFAULT_TOLERANCE
    ): bool {
        $signatureService = app(WebhookSignature::class);

        return $signatureService->verify($payload, $signature, $this->secret, $timestamp, $tolerance);
    }

    /**
     * Check if endpoint should receive an event.
     */
    public function shouldReceive(string $eventType): bool
    {
        if (! $this->active) {
            return false;
        }

        if ($this->disabled_at !== null) {
            return false;
        }

        return in_array($eventType, $this->events, true)
            || in_array('*', $this->events, true);
    }

    /**
     * Record successful delivery.
     */
    public function recordSuccess(): void
    {
        $this->update([
            'last_triggered_at' => now(),
            'failure_count' => 0,
        ]);
    }

    /**
     * Record failed delivery.
     * Auto-disables after 10 consecutive failures.
     */
    public function recordFailure(): void
    {
        $failureCount = $this->failure_count + 1;

        $updates = [
            'failure_count' => $failureCount,
            'last_triggered_at' => now(),
        ];

        // Auto-disable after 10 consecutive failures
        if ($failureCount >= 10) {
            $updates['disabled_at'] = now();
            $updates['active'] = false;
        }

        $this->update($updates);
    }

    /**
     * Re-enable a disabled endpoint.
     */
    public function enable(): void
    {
        $this->update([
            'active' => true,
            'disabled_at' => null,
            'failure_count' => 0,
        ]);
    }

    /**
     * Rotate the webhook secret.
     *
     * Generates a new cryptographically secure secret. The old secret
     * immediately becomes invalid - recipients must update their configuration.
     *
     * @return string The new secret (only returned once, store securely)
     */
    public function rotateSecret(): string
    {
        $signatureService = app(WebhookSignature::class);
        $newSecret = $signatureService->generateSecret();
        $this->update(['secret' => $newSecret]);

        return $newSecret;
    }

    // Relationships
    public function workspace(): BelongsTo
    {
        return $this->belongsTo(Workspace::class, 'workspace_id');
    }

    public function deliveries(): HasMany
    {
        return $this->hasMany(WebhookDelivery::class);
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->where('active', true)
            ->whereNull('disabled_at');
    }

    public function scopeForWorkspace($query, int $workspaceId)
    {
        return $query->where('workspace_id', $workspaceId);
    }

    public function scopeForEvent($query, string $eventType)
    {
        return $query->where(function ($q) use ($eventType) {
            $q->whereJsonContains('events', $eventType)
                ->orWhereJsonContains('events', '*');
        });
    }
}
