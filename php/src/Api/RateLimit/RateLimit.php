<?php

declare(strict_types=1);

namespace Core\Api\RateLimit;

use Attribute;

/**
 * Rate limit attribute for controllers and methods.
 *
 * Apply to controller classes or individual methods to set custom rate limits.
 * Method-level attributes take precedence over class-level attributes.
 *
 * Example usage:
 *
 *     #[RateLimit(limit: 100, window: 60)]
 *     class UserController extends Controller
 *     {
 *         #[RateLimit(limit: 10, window: 60)] // Override for this method
 *         public function store() {}
 *
 *         #[RateLimit(limit: 1000, window: 60, burst: 1.5)] // Allow 50% burst
 *         public function index() {}
 *     }
 */
#[Attribute(Attribute::TARGET_CLASS | Attribute::TARGET_METHOD)]
readonly class RateLimit
{
    /**
     * @param  int  $limit  Maximum requests allowed in the window
     * @param  int  $window  Time window in seconds
     * @param  float  $burst  Burst multiplier (e.g., 1.2 for 20% burst allowance)
     * @param  string|null  $key  Custom rate limit key suffix (null uses default)
     */
    public function __construct(
        public int $limit,
        public int $window = 60,
        public float $burst = 1.0,
        public ?string $key = null,
    ) {}
}
