<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

/**
 * API Key - authenticates SDK and REST API requests.
 *
 * Keys are prefixed with a short random identifier for identification.
 * The actual key is hashed with a slow password hash and never stored in plain text.
 *
 * Security: New keys use bcrypt; legacy SHA-256 keys remain readable only for
 * migration/rotation. The hash_algorithm column tracks which algorithm was used
 * for each key.
 */
class ApiKey extends Model
{
    use HasFactory;
    use SoftDeletes;

    /**
     * Hash algorithm identifiers.
     */
    public const HASH_SHA256 = 'sha256';

    public const HASH_BCRYPT = 'bcrypt';

    /**
     * Default grace period for key rotation (in hours).
     */
    public const DEFAULT_GRACE_PERIOD_HOURS = 24;

    /**
     * Scopes available for API keys.
     */
    public const SCOPE_READ = 'read';

    public const SCOPE_WRITE = 'write';

    public const SCOPE_DELETE = 'delete';

    public const ALL_SCOPES = [
        self::SCOPE_READ,
        self::SCOPE_WRITE,
        self::SCOPE_DELETE,
    ];

    /**
     * Build the visible key prefix root used for new API keys.
     *
     * The configured prefix is normalised to end with an underscore so
     * generated keys read as `hk_<random>_<token>`.
     */
    public static function keyPrefixRoot(): string
    {
        $prefix = trim((string) config('api.keys.prefix', 'hk_'));

        if ($prefix === '') {
            return 'hk_';
        }

        return str_ends_with($prefix, '_') ? $prefix : $prefix.'_';
    }

    /**
     * Generate the visible prefix portion of a new API key.
     *
     * Example:
     * `ApiKey::generatePrefix()` can return `hk_ab12cd34`.
     */
    public static function generatePrefix(): string
    {
        return static::keyPrefixRoot().Str::random(8);
    }

    protected $fillable = [
        'workspace_id',
        'user_id',
        'name',
        'key',
        'hash_algorithm',
        'prefix',
        'scopes',
        'server_scopes',
        'allowed_ips',
        'last_used_at',
        'expires_at',
        'grace_period_ends_at',
        'rotated_from_id',
    ];

    protected $casts = [
        'scopes' => 'array',
        'server_scopes' => 'array',
        'allowed_ips' => 'array',
        'last_used_at' => 'datetime',
        'expires_at' => 'datetime',
        'grace_period_ends_at' => 'datetime',
    ];

    protected $hidden = [
        'key', // Never expose the hashed key
    ];

    /**
     * Generate a new API key for a workspace.
     *
     * Returns both the ApiKey model and the plain key (only available once).
     * New keys use bcrypt for secure hashing with salt.
     *
     * @return array{api_key: ApiKey, plain_key: string}
     */
    public static function generate(
        int $workspaceId,
        int $userId,
        string $name,
        array $scopes = [self::SCOPE_READ, self::SCOPE_WRITE],
        ?\DateTimeInterface $expiresAt = null
    ): array {
        $plainKey = Str::random(48);
        $prefix = static::generatePrefix();

        $apiKey = static::create([
            'workspace_id' => $workspaceId,
            'user_id' => $userId,
            'name' => $name,
            'key' => Hash::driver('bcrypt')->make($plainKey),
            'hash_algorithm' => self::HASH_BCRYPT,
            'prefix' => $prefix,
            'scopes' => $scopes,
            'expires_at' => $expiresAt,
        ]);

        // Return plain key only once - never stored
        return [
            'api_key' => $apiKey,
            'plain_key' => "{$prefix}_{$plainKey}",
        ];
    }

    /**
     * Find an API key by its plain text value.
     *
     * Supports both legacy SHA-256 keys and new bcrypt keys.
     * For bcrypt keys, we must load all candidates by prefix and verify each.
     */
    public static function findByPlainKey(string $plainKey): ?static
    {
        $plainKey = trim($plainKey);
        if ($plainKey === '' || ! str_contains($plainKey, '_')) {
            return null;
        }

        $prefixRoot = static::keyPrefixRoot();
        $prefix = null;
        $key = null;

        if (str_starts_with($plainKey, $prefixRoot)) {
            $remainder = substr($plainKey, strlen($prefixRoot));
            if ($remainder !== false && $remainder !== '' && str_contains($remainder, '_')) {
                [$suffix, $key] = explode('_', $remainder, 2);
                if ($suffix !== '') {
                    $prefix = $prefixRoot.$suffix;
                }
            }
        }

        if ($prefix === null || $key === null) {
            $parts = explode('_', $plainKey, 2);
            if (count($parts) < 2) {
                return null;
            }

            [$prefix, $key] = $parts;
        }

        if ($prefix === '' || $key === '') {
            return null;
        }

        // Find potential matches by prefix
        $candidates = static::where('prefix', $prefix)
            ->whereNull('deleted_at')
            ->where(function ($query) {
                $query->whereNull('expires_at')
                    ->orWhere('expires_at', '>', now());
            })
            ->where(function ($query) {
                // Exclude keys past their grace period
                $query->whereNull('grace_period_ends_at')
                    ->orWhere('grace_period_ends_at', '>', now());
            })
            ->get();

        foreach ($candidates as $candidate) {
            if ($candidate->verifyKey($key)) {
                return $candidate;
            }
        }

        return null;
    }

    /**
     * Verify if the provided key matches this API key's stored hash.
     *
     * Handles both legacy SHA-256 and secure bcrypt algorithms.
     */
    public function verifyKey(string $plainKey): bool
    {
        if ($this->hash_algorithm === self::HASH_BCRYPT) {
            return Hash::check($plainKey, $this->key);
        }

        // Legacy SHA-256 verification (for backward compatibility)
        return hash_equals($this->key, hash('sha256', $plainKey));
    }

    /**
     * Check if this key uses legacy (insecure) SHA-256 hashing.
     *
     * Keys using SHA-256 should be rotated to use bcrypt.
     */
    public function usesLegacyHash(): bool
    {
        return $this->hash_algorithm === self::HASH_SHA256
            || $this->hash_algorithm === null;
    }

    /**
     * Rotate this API key, creating a new secure key.
     *
     * The old key remains valid during the grace period to allow
     * seamless migration of integrations.
     *
     * @param  int  $gracePeriodHours  Hours the old key remains valid
     * @return array{api_key: ApiKey, plain_key: string, old_key: ApiKey}
     */
    public function rotate(int $gracePeriodHours = self::DEFAULT_GRACE_PERIOD_HOURS): array
    {
        // Create new key with same settings
        $result = static::generate(
            $this->workspace_id,
            $this->user_id,
            $this->name,
            $this->scopes ?? [self::SCOPE_READ, self::SCOPE_WRITE],
            $this->expires_at
        );

        // Copy server scopes and IP restrictions to new key
        $result['api_key']->update([
            'server_scopes' => $this->server_scopes,
            'allowed_ips' => $this->allowed_ips,
            'rotated_from_id' => $this->id,
        ]);

        // Set grace period on old key
        $this->update([
            'grace_period_ends_at' => now()->addHours($gracePeriodHours),
        ]);

        return [
            'api_key' => $result['api_key'],
            'plain_key' => $result['plain_key'],
            'old_key' => $this,
        ];
    }

    /**
     * Check if this key is currently in a rotation grace period.
     */
    public function isInGracePeriod(): bool
    {
        return $this->grace_period_ends_at !== null
            && $this->grace_period_ends_at->isFuture();
    }

    /**
     * Check if the grace period has expired (key should be revoked).
     */
    public function isGracePeriodExpired(): bool
    {
        return $this->grace_period_ends_at !== null
            && $this->grace_period_ends_at->isPast();
    }

    /**
     * End the grace period early and revoke this key.
     */
    public function endGracePeriod(): void
    {
        $this->update(['grace_period_ends_at' => now()]);
        $this->revoke();
    }

    /**
     * Record API key usage.
     */
    public function recordUsage(): void
    {
        $this->update(['last_used_at' => now()]);
    }

    /**
     * Check if key has a specific scope.
     *
     * Supports wildcard matching:
     * - `posts:*` grants all actions on posts resource
     * - `*:read` grants read action on all resources
     * - `*` grants full access to everything
     */
    public function hasScope(string $scope): bool
    {
        $scopes = $this->scopes ?? [];

        // Exact match
        if (in_array($scope, $scopes, true)) {
            return true;
        }

        // Full wildcard (grants everything)
        if (in_array('*', $scopes, true)) {
            return true;
        }

        // Check for resource:action pattern
        if (! str_contains($scope, ':')) {
            // Simple scope (read, write, delete) - no wildcard matching
            return false;
        }

        [$resource, $action] = explode(':', $scope, 2);

        // Resource wildcard (e.g., posts:* grants posts:read, posts:write, etc.)
        if (in_array("{$resource}:*", $scopes, true)) {
            return true;
        }

        // Action wildcard (e.g., *:read grants posts:read, users:read, etc.)
        if (in_array("*:{$action}", $scopes, true)) {
            return true;
        }

        return false;
    }

    /**
     * Check if key has all specified scopes.
     */
    public function hasScopes(array $scopes): bool
    {
        foreach ($scopes as $scope) {
            if (! $this->hasScope($scope)) {
                return false;
            }
        }

        return true;
    }

    /**
     * Check if key has any of the specified scopes.
     */
    public function hasAnyScope(array $scopes): bool
    {
        foreach ($scopes as $scope) {
            if ($this->hasScope($scope)) {
                return true;
            }
        }

        return false;
    }

    /**
     * Check if key is expired.
     */
    public function isExpired(): bool
    {
        return $this->expires_at !== null && $this->expires_at->isPast();
    }

    /**
     * Check if key has access to a specific MCP server.
     */
    public function hasServerAccess(string $serverId): bool
    {
        // Null means all servers
        if ($this->server_scopes === null) {
            return true;
        }

        return in_array($serverId, $this->server_scopes, true);
    }

    /**
     * Get list of allowed servers (null = all).
     */
    public function getAllowedServers(): ?array
    {
        return $this->server_scopes;
    }

    /**
     * Check if this key has IP restrictions enabled.
     */
    public function hasIpRestrictions(): bool
    {
        return ! empty($this->allowed_ips);
    }

    /**
     * Get the allowed IPs list (null = all IPs allowed).
     *
     * @return array<string>|null
     */
    public function getAllowedIps(): ?array
    {
        return $this->allowed_ips;
    }

    /**
     * Update the IP whitelist.
     *
     * @param  array<string>|null  $ips  Array of IP addresses/CIDRs, or null to allow all
     */
    public function updateAllowedIps(?array $ips): void
    {
        $this->update(['allowed_ips' => $ips]);
    }

    /**
     * Add an IP or CIDR to the whitelist.
     */
    public function addAllowedIp(string $ipOrCidr): void
    {
        $whitelist = $this->allowed_ips ?? [];

        if (! in_array($ipOrCidr, $whitelist, true)) {
            $whitelist[] = $ipOrCidr;
            $this->update(['allowed_ips' => $whitelist]);
        }
    }

    /**
     * Remove an IP or CIDR from the whitelist.
     */
    public function removeAllowedIp(string $ipOrCidr): void
    {
        $whitelist = $this->allowed_ips ?? [];
        $whitelist = array_values(array_filter($whitelist, fn ($entry) => $entry !== $ipOrCidr));
        $this->update(['allowed_ips' => empty($whitelist) ? null : $whitelist]);
    }

    /**
     * Revoke this API key.
     */
    public function revoke(): void
    {
        $this->delete();
    }

    /**
     * Get the masked key for display.
     * Shows prefix and last 4 characters.
     */
    public function getMaskedKeyAttribute(): string
    {
        return "{$this->prefix}_****";
    }

    // Relationships
    public function workspace(): BelongsTo
    {
        return $this->belongsTo(Workspace::class, 'workspace_id');
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    /**
     * Get the key this one was rotated from.
     */
    public function rotatedFrom(): BelongsTo
    {
        return $this->belongsTo(static::class, 'rotated_from_id');
    }

    // Query Scopes
    public function scopeForWorkspace($query, int $workspaceId)
    {
        return $query->where('workspace_id', $workspaceId);
    }

    public function scopeActive($query)
    {
        return $query->whereNull('deleted_at')
            ->where(function ($q) {
                $q->whereNull('expires_at')
                    ->orWhere('expires_at', '>', now());
            })
            ->where(function ($q) {
                $q->whereNull('grace_period_ends_at')
                    ->orWhere('grace_period_ends_at', '>', now());
            });
    }

    public function scopeExpired($query)
    {
        return $query->whereNotNull('expires_at')
            ->where('expires_at', '<=', now());
    }

    /**
     * Keys currently in a rotation grace period.
     */
    public function scopeInGracePeriod($query)
    {
        return $query->whereNotNull('grace_period_ends_at')
            ->where('grace_period_ends_at', '>', now());
    }

    /**
     * Keys with expired grace periods (should be cleaned up).
     */
    public function scopeGracePeriodExpired($query)
    {
        return $query->whereNotNull('grace_period_ends_at')
            ->where('grace_period_ends_at', '<=', now());
    }

    /**
     * Keys using legacy SHA-256 hashing (should be rotated).
     */
    public function scopeLegacyHash($query)
    {
        return $query->where(function ($q) {
            $q->where('hash_algorithm', self::HASH_SHA256)
                ->orWhereNull('hash_algorithm');
        });
    }

    /**
     * Keys using secure bcrypt hashing.
     */
    public function scopeSecureHash($query)
    {
        return $query->where('hash_algorithm', self::HASH_BCRYPT);
    }
}
