<?php

declare(strict_types=1);

namespace Core\Api\RateLimit;

use Carbon\Carbon;

/**
 * Rate limit check result DTO.
 *
 * Contains information about the current rate limit status for a request.
 */
readonly class RateLimitResult
{
    public function __construct(
        public bool $allowed,
        public int $limit,
        public int $remaining,
        public int $retryAfter,
        public Carbon $resetsAt,
    ) {}

    /**
     * Create a successful result (request allowed).
     */
    public static function allowed(int $limit, int $remaining, Carbon $resetsAt): self
    {
        return new self(
            allowed: true,
            limit: $limit,
            remaining: $remaining,
            retryAfter: 0,
            resetsAt: $resetsAt,
        );
    }

    /**
     * Create a denied result (rate limit exceeded).
     */
    public static function denied(int $limit, int $retryAfter, Carbon $resetsAt): self
    {
        return new self(
            allowed: false,
            limit: $limit,
            remaining: 0,
            retryAfter: $retryAfter,
            resetsAt: $resetsAt,
        );
    }

    /**
     * Get headers for the response.
     *
     * @return array<string, string|int>
     */
    public function headers(): array
    {
        $headers = [
            'X-RateLimit-Limit' => $this->limit,
            'X-RateLimit-Remaining' => $this->remaining,
            'X-RateLimit-Reset' => $this->resetsAt->timestamp,
        ];

        if (! $this->allowed) {
            $headers['Retry-After'] = $this->retryAfter;
        }

        return $headers;
    }
}
