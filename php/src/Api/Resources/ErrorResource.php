<?php

declare(strict_types=1);

namespace Core\Api\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

/**
 * Standard error response format.
 *
 * Usage:
 *   return ErrorResource::make('validation_error', 'The given data was invalid.', [
 *       'name' => ['The name field is required.'],
 *   ])->response()->setStatusCode(422);
 */
class ErrorResource extends JsonResource
{
    protected string $errorCode;

    protected string $message;

    protected ?array $details;

    public function __construct(string $errorCode, string $message, ?array $details = null)
    {
        $this->errorCode = $errorCode;
        $this->message = $message;
        $this->details = $details;

        parent::__construct(null);
    }

    public static function make(...$args): static
    {
        return new static(...$args);
    }

    /**
     * Common error factory methods.
     */
    public static function unauthorized(string $message = 'Unauthorised'): static
    {
        return new static('unauthorized', $message);
    }

    public static function forbidden(string $message = 'Forbidden'): static
    {
        return new static('forbidden', $message);
    }

    public static function notFound(string $message = 'Resource not found'): static
    {
        return new static('not_found', $message);
    }

    public static function validation(array $errors): static
    {
        return new static('validation_error', 'The given data was invalid.', $errors);
    }

    public static function rateLimited(int $retryAfter): static
    {
        return new static('rate_limit_exceeded', 'Too many requests. Please slow down.', [
            'retry_after' => $retryAfter,
        ]);
    }

    public static function entitlementExceeded(string $feature): static
    {
        return new static('entitlement_exceeded', "Plan limit reached for: {$feature}");
    }

    public static function serverError(string $message = 'An unexpected error occurred'): static
    {
        return new static('internal_error', $message);
    }

    public function toArray(Request $request): array
    {
        $response = [
            'error' => $this->errorCode,
            'message' => $this->message,
        ];

        if ($this->details !== null) {
            $response['details'] = $this->details;
        }

        return $response;
    }
}
