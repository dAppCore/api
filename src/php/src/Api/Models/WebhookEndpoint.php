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
        'subscription.changed',
        'subscription.created',
        'subscription.updated',
        'subscription.cancelled',
        'subscription.renewed',

        // Invoice events
        'invoice.created',
        'invoice.paid',
        'invoice.failed',

        // BioLink events
        'biolink.created',
        'biolink.updated',
        'biolink.deleted',
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

        // Support events
        'ticket.created',
        'ticket.replied',
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
     * Ensure webhook URLs stay on public HTTP(S) destinations.
     */
    protected static function booted(): void
    {
        static::saving(function (self $webhook): void {
            if ($webhook->isDirty('url')) {
                self::assertSafeUrl((string) $webhook->url);
            }
        });
    }

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
     * Assert that a webhook URL is safe to use.
     *
     * Blocks non-HTTP(S) schemes and destinations that resolve to loopback,
     * private, link-local, or otherwise reserved addresses.
     *
     * @throws \InvalidArgumentException
     */
    public static function assertSafeUrl(string $url): void
    {
        $parsed = parse_url($url);

        if ($parsed === false || empty($parsed['scheme']) || empty($parsed['host'])) {
            throw new \InvalidArgumentException('The webhook URL must be an absolute HTTP or HTTPS URL.');
        }

        if (! in_array(strtolower($parsed['scheme']), ['http', 'https'], true)) {
            throw new \InvalidArgumentException('Only HTTP and HTTPS webhook URLs are permitted.');
        }

        if (isset($parsed['user']) || isset($parsed['pass'])) {
            throw new \InvalidArgumentException('Webhook URLs must not include embedded credentials.');
        }

        $host = $parsed['host'];
        $normalisedHost = ltrim(rtrim($host, ']'), '[');

        if (filter_var($normalisedHost, FILTER_VALIDATE_IP) !== false) {
            if (self::isPrivateIp($normalisedHost)) {
                throw new \InvalidArgumentException('Webhook URLs must not target private, loopback, or reserved addresses.');
            }

            return;
        }

        $records = dns_get_record($host, DNS_A | DNS_AAAA) ?: [];
        if ($records === []) {
            throw new \InvalidArgumentException('The webhook URL must resolve to a public IP address.');
        }

        foreach ($records as $record) {
            $ip = $record['ip'] ?? $record['ipv6'] ?? null;
            if ($ip === null) {
                continue;
            }

            if (self::isPrivateIp($ip)) {
                throw new \InvalidArgumentException('Webhook URLs must not resolve to private, loopback, or reserved addresses.');
            }
        }
    }

    /**
     * Return true when the IP falls within a private, loopback, link-local,
     * or otherwise reserved range.
     */
    protected static function isPrivateIp(string $ip): bool
    {
        return filter_var(
            $ip,
            FILTER_VALIDATE_IP,
            FILTER_FLAG_NO_PRIV_RANGE | FILTER_FLAG_NO_RES_RANGE
        ) === false;
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
        $eventTypes = self::eventTypeAliases($eventType);

        if (! $this->active) {
            return false;
        }

        if ($this->disabled_at !== null) {
            return false;
        }

        if (in_array('*', $this->events, true)) {
            return true;
        }

        foreach ($eventTypes as $type) {
            if (in_array($type, $this->events, true)) {
                return true;
            }
        }

        return false;
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
        $eventTypes = self::eventTypeAliases($eventType);

        return $query->where(function ($q) use ($eventTypes) {
            $q->whereJsonContains('events', '*');
            foreach ($eventTypes as $type) {
                $q->orWhereJsonContains('events', $type);
            }
        });
    }

    /**
     * Normalize an event type to its canonical name.
     *
     * Legacy "bio.*" event names are retained as aliases for the newer
     * "biolink.*" namespace used by the RFC.
     */
    protected static function normalizeEventType(string $eventType): string
    {
        $eventType = trim($eventType);

        return match ($eventType) {
            'bio.created' => 'biolink.created',
            'bio.updated' => 'biolink.updated',
            'bio.deleted' => 'biolink.deleted',
            default => $eventType,
        };
    }

    /**
     * Return the canonical event type and any legacy aliases that should match it.
     *
     * @return array<int, string>
     */
    protected static function eventTypeAliases(string $eventType): array
    {
        $normalized = self::normalizeEventType($eventType);

        return match ($normalized) {
            'biolink.created' => ['biolink.created', 'bio.created'],
            'biolink.updated' => ['biolink.updated', 'bio.updated'],
            'biolink.deleted' => ['biolink.deleted', 'bio.deleted'],
            default => [$normalized],
        };
    }
}
