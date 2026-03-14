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
        $currentCount = count($hits);
        $remaining = max(0, $effectiveLimit - $currentCount);

        // Calculate reset time
        $resetsAt = $this->calculateResetTime($hits, $window, $effectiveLimit);

        if ($currentCount >= $effectiveLimit) {
            // Find oldest hit to determine retry after
            $oldestHit = min($hits);
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
            $oldestHit = min($hits);
            $retryAfter = max(1, ($oldestHit + $window) - $now->timestamp);

            return RateLimitResult::denied($limit, $retryAfter, $resetsAt);
        }

        // Record the hit
        $hits[] = $now->timestamp;
        $this->storeWindowHits($cacheKey, $hits, $window);

        $remaining = max(0, $effectiveLimit - count($hits));

        return RateLimitResult::allowed($limit, $remaining, $resetsAt);
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

        return max(0, $effectiveLimit - count($hits));
    }

    /**
     * Reset (clear) a rate limit bucket.
     */
    public function reset(string $key): void
    {
        $cacheKey = $this->getCacheKey($key);
        $this->cache->forget($cacheKey);
    }

    /**
     * Get the current hit count for a key.
     */
    public function attempts(string $key, int $window): int
    {
        $cacheKey = $this->getCacheKey($key);
        $windowStart = Carbon::now()->timestamp - $window;

        return count($this->getWindowHits($cacheKey, $windowStart));
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
     * @return array<int> Array of timestamps
     */
    protected function getWindowHits(string $cacheKey, int $windowStart): array
    {
        /** @var array<int> $hits */
        $hits = $this->cache->get($cacheKey, []);

        // Filter to only include hits within the window
        return array_values(array_filter($hits, fn (int $timestamp) => $timestamp >= $windowStart));
    }

    /**
     * Store hits in cache.
     *
     * @param  array<int>  $hits  Array of timestamps
     */
    protected function storeWindowHits(string $cacheKey, array $hits, int $window): void
    {
        // Add buffer to TTL to handle clock drift
        $ttl = $window + 60;
        $this->cache->put($cacheKey, $hits, $ttl);
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
