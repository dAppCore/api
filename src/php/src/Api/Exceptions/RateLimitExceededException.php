<?php

declare(strict_types=1);

namespace Core\Api\Exceptions;

use Core\Api\RateLimit\RateLimitResult;
use Core\Api\Concerns\HasApiResponses;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Symfony\Component\HttpKernel\Exception\HttpException;

/**
 * Exception thrown when API rate limit is exceeded.
 *
 * Renders as a proper JSON response with rate limit headers.
 */
class RateLimitExceededException extends HttpException
{
    use HasApiResponses;

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
    public function render(?Request $request = null): JsonResponse
    {
        $response = $this->errorResponse(
            errorCode: 'rate_limit_exceeded',
            message: $this->getMessage(),
            meta: [
                'retry_after' => $this->rateLimitResult->retryAfter,
                'limit' => $this->rateLimitResult->limit,
                'resets_at' => $this->rateLimitResult->resetsAt->toIso8601String(),
            ],
            status: 429,
        )->withHeaders($this->rateLimitResult->headers());

        if ($request !== null) {
            $origin = $request->headers->get('Origin');
            $allowedOrigins = (array) config('cors.allowed_origins', []);
            if ($origin !== null && in_array($origin, $allowedOrigins, true)) {
                $response->headers->set('Access-Control-Allow-Origin', $origin);
            }

            $existingVary = $response->headers->get('Vary');
            $response->headers->set(
                'Vary',
                $existingVary ? $existingVary.', Origin' : 'Origin'
            );
        }

        return $response;
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
