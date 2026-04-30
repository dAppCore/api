<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Attributes;

use Attribute;

/**
 * API Hidden attribute for excluding endpoints from documentation.
 *
 * Apply to controller classes or methods to hide them from the generated
 * OpenAPI documentation.
 *
 * Example usage:
 *
 *     // Hide entire controller
 *     #[ApiHidden]
 *     class InternalController extends Controller {}
 *
 *     // Hide specific method
 *     class UserController extends Controller
 *     {
 *         #[ApiHidden]
 *         public function internalMethod() {}
 *     }
 *
 *     // Hide with reason (for code documentation)
 *     #[ApiHidden('Internal use only')]
 *     public function debug() {}
 */
#[Attribute(Attribute::TARGET_CLASS | Attribute::TARGET_METHOD)]
readonly class ApiHidden
{
    /**
     * @param  string|null  $reason  Optional reason for hiding (documentation only)
     */
    public function __construct(
        public ?string $reason = null,
    ) {}
}
