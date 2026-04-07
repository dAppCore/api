<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Attributes;

use Attribute;

/**
 * API Response attribute for documenting endpoint responses.
 *
 * Apply to controller methods to document possible responses in OpenAPI.
 *
 * Example usage:
 *
 *     #[ApiResponse(200, UserResource::class, 'User retrieved successfully')]
 *     #[ApiResponse(404, null, 'User not found')]
 *     #[ApiResponse(422, null, 'Validation failed')]
 *     public function show(User $user)
 *     {
 *         return new UserResource($user);
 *     }
 *
 *     // For paginated responses
 *     #[ApiResponse(200, UserResource::class, 'Users list', paginated: true)]
 *     public function index()
 *     {
 *         return UserResource::collection(User::paginate());
 *     }
 *
 *     // For non-JSON or binary responses
 *     #[ApiResponse(
 *         200,
 *         null,
 *         'Transparent tracking pixel',
 *         contentType: 'image/gif',
 *         schema: ['type' => 'string', 'format' => 'binary']
 *     )]
 *     public function pixel()
 *     {
 *         return response($gif, 200)->header('Content-Type', 'image/gif');
 *     }
 */
#[Attribute(Attribute::TARGET_METHOD | Attribute::IS_REPEATABLE)]
readonly class ApiResponse
{
    /**
     * @param  int  $status  HTTP status code
     * @param  string|null  $resource  Resource class for response body (null for no body)
     * @param  string|null  $description  Description of the response
     * @param  bool  $paginated  Whether this is a paginated collection response
     * @param  array<string>  $headers  Additional response headers to document
     * @param  string|null  $contentType  Explicit response media type for non-JSON responses
     * @param  array<string, mixed>|null  $schema  Explicit response schema when the body is not inferred from a resource
     */
    public function __construct(
        public int $status,
        public ?string $resource = null,
        public ?string $description = null,
        public bool $paginated = false,
        public array $headers = [],
        public ?string $contentType = null,
        public ?array $schema = null,
    ) {}

    /**
     * Get the description or generate from status code.
     */
    public function getDescription(): string
    {
        if ($this->description !== null) {
            return $this->description;
        }

        return match ($this->status) {
            200 => 'Successful response',
            201 => 'Resource created',
            202 => 'Request accepted',
            204 => 'No content',
            301 => 'Moved permanently',
            302 => 'Found (redirect)',
            304 => 'Not modified',
            400 => 'Bad request',
            401 => 'Unauthorised',
            403 => 'Forbidden',
            404 => 'Not found',
            405 => 'Method not allowed',
            410 => 'Gone',
            409 => 'Conflict',
            422 => 'Validation error',
            429 => 'Too many requests',
            500 => 'Internal server error',
            502 => 'Bad gateway',
            503 => 'Service unavailable',
            default => 'Response',
        };
    }
}
