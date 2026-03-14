<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Attributes;

use Attribute;

/**
 * API Tag attribute for grouping endpoints in documentation.
 *
 * Apply to controller classes to group their endpoints under a specific tag
 * in the OpenAPI documentation.
 *
 * Example usage:
 *
 *     #[ApiTag('Users', 'User management endpoints')]
 *     class UserController extends Controller
 *     {
 *         // All methods will be tagged with 'Users'
 *     }
 *
 *     // Or use on specific methods to override class-level tag
 *     #[ApiTag('Admin')]
 *     public function adminOnly() {}
 */
#[Attribute(Attribute::TARGET_CLASS | Attribute::TARGET_METHOD)]
readonly class ApiTag
{
    /**
     * @param  string  $name  The tag name displayed in documentation
     * @param  string|null  $description  Optional description of the tag
     */
    public function __construct(
        public string $name,
        public ?string $description = null,
    ) {}
}
