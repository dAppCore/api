<?php

declare(strict_types=1);

namespace Core\Api\Exceptions;

use Core\Api\RateLimit\RateLimitResult;
use Illuminate\Http\JsonResponse;
use Symfony\Component\HttpKernel\Exception\HttpException;

/**
 * Exception thrown when API rate limit is exceeded.
 *
 * Renders as a proper JSON response with rate limit headers.
 */
class RateLimitExceededException extends HttpException
{
    public function __construct(
        protected RateLimitResult $rateLimitResult,
        string $message = 'Too many requests. Please slow down.',
    ) {
        parent::__construct(429, $message);
    }

    /**
     * Get the rate limit result.
     */
    public function getRateLimitResult(): RateLimitResult
    {
        return $this->rateLimitResult;
    }

    /**
     * Render the exception as a JSON response.
     */
    public function render(): JsonResponse
    {
        return response()->json([
            'error' => 'rate_limit_exceeded',
            'message' => $this->getMessage(),
            'retry_after' => $this->rateLimitResult->retryAfter,
            'limit' => $this->rateLimitResult->limit,
            'resets_at' => $this->rateLimitResult->resetsAt->toIso8601String(),
        ], 429, $this->rateLimitResult->headers());
    }

    /**
     * Get headers for the response.
     *
     * @return array<string, string|int>
     */
    public function getHeaders(): array
    {
        return array_map(fn ($value) => (string) $value, $this->rateLimitResult->headers());
    }
}
