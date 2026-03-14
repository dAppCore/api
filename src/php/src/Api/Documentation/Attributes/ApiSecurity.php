<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Attributes;

use Attribute;

/**
 * API Security attribute for documenting authentication requirements.
 *
 * Apply to controller classes or methods to specify authentication requirements.
 *
 * Example usage:
 *
 *     // Require API key authentication
 *     #[ApiSecurity('apiKey')]
 *     class ProtectedController extends Controller {}
 *
 *     // Require bearer token
 *     #[ApiSecurity('bearer')]
 *     public function profile() {}
 *
 *     // Require specific scopes
 *     #[ApiSecurity('apiKey', scopes: ['read', 'write'])]
 *     public function update() {}
 *
 *     // Mark endpoint as public (no auth required)
 *     #[ApiSecurity(null)]
 *     public function publicEndpoint() {}
 */
#[Attribute(Attribute::TARGET_CLASS | Attribute::TARGET_METHOD | Attribute::IS_REPEATABLE)]
readonly class ApiSecurity
{
    /**
     * @param  string|null  $scheme  Security scheme name (null for no auth)
     * @param  array<string>  $scopes  Required OAuth2 scopes (if applicable)
     */
    public function __construct(
        public ?string $scheme,
        public array $scopes = [],
    ) {}

    /**
     * Check if this marks the endpoint as public.
     */
    public function isPublic(): bool
    {
        return $this->scheme === null;
    }
}
