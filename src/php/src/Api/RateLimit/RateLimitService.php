<?php

declare(strict_types=1);

namespace Core\Api\RateLimit;

use Carbon\Carbon;
use Illuminate\Contracts\Cache\Repository as CacheRepository;

/**
 * Rate limiting service with sliding window algorithm.
 *
 * Provides granular rate limiting with support for:
 * - Per-key rate limiting (API keys, users, IPs, etc.)
 * - Sliding window algorithm for smoother rate limiting
 * - Burst allowance configuration
 * - Tier-based limits
 */
class RateLimitService
{
    /**
     * Cache prefix for rate limit keys.
     */
    protected const CACHE_PREFIX = 'rate_limit:';

    /**
     * TTL for advisory bucket locks when the cache backend does not expose a
     * native lock primitive.
     */
    protected const LOCK_TTL_SECONDS = 10;

    public function __construct(
        protected CacheRepository $cache,
    ) {}

    /**
     * Check if a request would be allowed without incrementing the counter.
     *
     * @param  string  $key  Unique identifier for the rate limit bucket
     * @param  int  $limit  Maximum requests allowed
     * @param  int  $window  Time window in seconds
     * @param  float  $burst  Burst multiplier (e.g., 1.2 for 20% burst allowance)
     */
    public function check(string $key, int $limit, int $window, float $burst = 1.0): RateLimitResult
    {
        $cacheKey = $this->getCacheKey($key);
        $effectiveLimit = (int) floor($limit * $burst);
        $now = Carbon::now();
        $windowStart = $now->timestamp - $window;

        // Get current window data
        $hits = $this->getWindowHits($cacheKey, $windowStart);
        if ($hits === null) {
            return RateLimitResult::denied($limit, 1, $now->copy()->addSecond());
        }

        $currentCount = count($hits);
        $remaining = max(0, $effectiveLimit - $currentCount);

        // Calculate reset time
        $resetsAt = $this->calculateResetTime($hits, $window, $effectiveLimit);

        if ($currentCount >= $effectiveLimit) {
            // Find oldest hit to determine retry after
            $oldestHit = $hits !== [] ? min($hits) : $now->timestamp;
            $retryAfter = max(1, ($oldestHit + $window) - $now->timestamp);

            return RateLimitResult::denied($limit, $retryAfter, $resetsAt);
        }

        return RateLimitResult::allowed($limit, $remaining, $resetsAt);
    }

    /**
     * Record a hit and check if the request is allowed.
     *
     * @param  string  $key  Unique identifier for the rate limit bucket
     * @param  int  $limit  Maximum requests allowed
     * @param  int  $window  Time window in seconds
     * @param  float  $burst  Burst multiplier (e.g., 1.2 for 20% burst allowance)
     */
    public function hit(string $key, int $limit, int $window, float $burst = 1.0): RateLimitResult
    {
        $cacheKey = $this->getCacheKey($key);
        $lock = $this->acquireBucketLock($cacheKey);
        if ($lock === null) {
            // Fail closed only when neither the native lock path nor the
            // advisory lock fallback can safely protect the bucket.
            return RateLimitResult::denied($limit, 1, Carbon::now()->addSecond());
        }

        try {
            return $this->hitWithoutLock($cacheKey, $limit, $window, $burst);
        } finally {
            $lock->release();
        }
    }

    /**
     * Get remaining attempts for a key.
     *
     * @param  string  $key  Unique identifier for the rate limit bucket
     * @param  int  $limit  Maximum requests allowed (needed to calculate remaining)
     * @param  int  $window  Time window in seconds
     * @param  float  $burst  Burst multiplier
     */
    public function remaining(string $key, int $limit, int $window, float $burst = 1.0): int
    {
        $cacheKey = $this->getCacheKey($key);
        $effectiveLimit = (int) floor($limit * $burst);
        $windowStart = Carbon::now()->timestamp - $window;

        $hits = $this->getWindowHits($cacheKey, $windowStart);
        if ($hits === null) {
            return 0;
        }

        return max(0, $effectiveLimit - count($hits));
    }

    /**
     * Reset (clear) a rate limit bucket.
     */
    public function reset(string $key): void
    {
        $cacheKey = $this->getCacheKey($key);

        try {
            $this->cache->forget($cacheKey);
        } catch (\Throwable) {
            // Best-effort cleanup only.
        }
    }

    /**
     * Get the current hit count for a key.
     */
    public function attempts(string $key, int $window): int
    {
        $cacheKey = $this->getCacheKey($key);
        $windowStart = Carbon::now()->timestamp - $window;

        $hits = $this->getWindowHits($cacheKey, $windowStart);

        return $hits === null ? 0 : count($hits);
    }

    /**
     * Build a rate limit key for an endpoint.
     */
    public function buildEndpointKey(string $identifier, string $endpoint): string
    {
        return "endpoint:{$identifier}:{$endpoint}";
    }

    /**
     * Build a rate limit key for a workspace.
     */
    public function buildWorkspaceKey(int $workspaceId, ?string $suffix = null): string
    {
        $key = "workspace:{$workspaceId}";

        if ($suffix !== null) {
            $key .= ":{$suffix}";
        }

        return $key;
    }

    /**
     * Build a rate limit key for an API key.
     */
    public function buildApiKeyKey(int|string $apiKeyId, ?string $suffix = null): string
    {
        $key = "api_key:{$apiKeyId}";

        if ($suffix !== null) {
            $key .= ":{$suffix}";
        }

        return $key;
    }

    /**
     * Build a rate limit key for an IP address.
     */
    public function buildIpKey(string $ip, ?string $suffix = null): string
    {
        $key = "ip:{$ip}";

        if ($suffix !== null) {
            $key .= ":{$suffix}";
        }

        return $key;
    }

    /**
     * Get hits within the sliding window.
     *
     * @return array<int>|null Array of timestamps, or null when the cache value
     *                         is unavailable or malformed.
     */
    protected function getWindowHits(string $cacheKey, int $windowStart): ?array
    {
        try {
            $hits = $this->cache->get($cacheKey, []);
        } catch (\Throwable) {
            return null;
        }

        if (! is_array($hits)) {
            return null;
        }

        $filtered = [];

        foreach ($hits as $hit) {
            if (is_int($hit)) {
                $timestamp = $hit;
            } elseif (is_string($hit) && preg_match('/^-?\d+$/', $hit)) {
                $timestamp = (int) $hit;
            } else {
                return null;
            }

            if ($timestamp >= $windowStart) {
                $filtered[] = $timestamp;
            }
        }

        return array_values($filtered);
    }

    /**
     * Store hits in cache.
     *
     * @param  array<int>  $hits  Array of timestamps
     */
    protected function storeWindowHits(string $cacheKey, array $hits, int $window): bool
    {
        // Add buffer to TTL to handle clock drift
        $ttl = $window + 60;

        try {
            $this->cache->put($cacheKey, $hits, $ttl);

            return true;
        } catch (\Throwable) {
            return false;
        }
    }

    /**
     * Perform the rate-limit update without acquiring a lock first.
     *
     * Callers should use this only when the cache backend cannot provide an
     * atomic lock. The public `hit()` method guards this path.
     */
    protected function hitWithoutLock(string $cacheKey, int $limit, int $window, float $burst): RateLimitResult
    {
        $effectiveLimit = (int) floor($limit * $burst);
        $now = Carbon::now();
        $windowStart = $now->timestamp - $window;

        // Get current window data and clean up old entries
        $hits = $this->getWindowHits($cacheKey, $windowStart);
        $currentCount = count($hits);

        // Calculate reset time
        $resetsAt = $this->calculateResetTime($hits, $window, $effectiveLimit);

        if ($currentCount >= $effectiveLimit) {
            // Find oldest hit to determine retry after
            $oldestHit = $hits !== [] ? min($hits) : $now->timestamp;
            $retryAfter = max(1, ($oldestHit + $window) - $now->timestamp);

            return RateLimitResult::denied($limit, $retryAfter, $resetsAt);
        }

        // Record the hit
        $hits[] = $now->timestamp;
        if (! $this->storeWindowHits($cacheKey, $hits, $window)) {
            return RateLimitResult::denied($limit, 1, $now->copy()->addSecond());
        }

        $remaining = max(0, $effectiveLimit - count($hits));

        return RateLimitResult::allowed($limit, $remaining, $resetsAt);
    }

    /**
     * Determine whether the cache backend exposes atomic locks.
     */
    protected function supportsAtomicLocks(): bool
    {
        return method_exists($this->cache, 'lock');
    }

    /**
     * Try to acquire a lock for a specific rate-limit bucket.
     */
    protected function acquireBucketLock(string $cacheKey): mixed
    {
        if ($this->supportsAtomicLocks()) {
            try {
                $lock = $this->cache->lock($cacheKey.':lock', 10);

                if (method_exists($lock, 'get') && $lock->get()) {
                    return $lock;
                }
            } catch (\Throwable) {
                // Fall through to the advisory lock fallback below.
            }
        }

        return $this->acquireAdvisoryLock($cacheKey);
    }

    /**
     * Acquire a best-effort cache-backed lock using atomic add semantics.
     *
     * This path is used on cache drivers that do not expose a dedicated lock
     * primitive but still support atomic "add" operations. It prevents
     * parallel requests from oversubscribing the same bucket on common cache
     * stores such as the array, file, and Redis drivers.
     */
    protected function acquireAdvisoryLock(string $cacheKey): mixed
    {
        if (! method_exists($this->cache, 'add')) {
            return null;
        }

        $lockKey = $cacheKey.':lock';
        try {
            $token = bin2hex(random_bytes(8));
        } catch (\Throwable) {
            return null;
        }

        try {
            if (! $this->cache->add($lockKey, $token, self::LOCK_TTL_SECONDS)) {
                return null;
            }
        } catch (\Throwable) {
            return null;
        }

        return new class($this->cache, $lockKey, $token) {
            public function __construct(
                protected CacheRepository $cache,
                protected string $lockKey,
                protected string $token,
            ) {
            }

            public function release(): void
            {
                try {
                    if ($this->cache->get($this->lockKey) === $this->token) {
                        $this->cache->forget($this->lockKey);
                    }
                } catch (\Throwable) {
                    // Best-effort cleanup only.
                }
            }
        };
    }

    /**
     * Calculate when the rate limit resets.
     *
     * @param  array<int>  $hits  Array of timestamps
     */
    protected function calculateResetTime(array $hits, int $window, int $limit): Carbon
    {
        if (empty($hits)) {
            return Carbon::now()->addSeconds($window);
        }

        // If under limit, reset is at the end of the window
        if (count($hits) < $limit) {
            return Carbon::now()->addSeconds($window);
        }

        // If at or over limit, reset when the oldest hit expires
        $oldestHit = min($hits);

        return Carbon::createFromTimestamp($oldestHit + $window);
    }

    /**
     * Generate the cache key.
     */
    protected function getCacheKey(string $key): string
    {
        return self::CACHE_PREFIX.$key;
    }
}
