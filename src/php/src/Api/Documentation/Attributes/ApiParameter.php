<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Attributes;

use Attribute;

/**
 * API Parameter attribute for documenting endpoint parameters.
 *
 * Apply to controller methods to document query parameters, path parameters,
 * or header parameters in OpenAPI documentation.
 *
 * Example usage:
 *
 *     #[ApiParameter('page', 'query', 'integer', 'Page number', required: false, example: 1)]
 *     #[ApiParameter('per_page', 'query', 'integer', 'Items per page', required: false, example: 25)]
 *     #[ApiParameter('filter[status]', 'query', 'string', 'Filter by status', enum: ['active', 'inactive'])]
 *     public function index(Request $request)
 *     {
 *         // ...
 *     }
 *
 *     // Document header parameters
 *     #[ApiParameter('X-Custom-Header', 'header', 'string', 'Custom header value')]
 *     public function withHeader() {}
 */
#[Attribute(Attribute::TARGET_METHOD | Attribute::IS_REPEATABLE)]
readonly class ApiParameter
{
    /**
     * @param  string  $name  Parameter name
     * @param  string  $in  Parameter location: 'query', 'path', 'header', 'cookie'
     * @param  string  $type  Data type: 'string', 'integer', 'boolean', 'number', 'array'
     * @param  string|null  $description  Parameter description
     * @param  bool  $required  Whether parameter is required
     * @param  mixed  $example  Example value
     * @param  mixed  $default  Default value
     * @param  array|null  $enum  Allowed values (for enumerated parameters)
     * @param  string|null  $format  Format hint (e.g., 'date', 'email', 'uuid')
     */
    public function __construct(
        public string $name,
        public string $in = 'query',
        public string $type = 'string',
        public ?string $description = null,
        public bool $required = false,
        public mixed $example = null,
        public mixed $default = null,
        public ?array $enum = null,
        public ?string $format = null,
    ) {}

    /**
     * Convert to OpenAPI parameter schema.
     */
    public function toSchema(): array
    {
        $schema = [
            'type' => $this->type,
        ];

        if ($this->format !== null) {
            $schema['format'] = $this->format;
        }

        if ($this->enum !== null) {
            $schema['enum'] = $this->enum;
        }

        if ($this->default !== null) {
            $schema['default'] = $this->default;
        }

        if ($this->example !== null) {
            $schema['example'] = $this->example;
        }

        return $schema;
    }

    /**
     * Convert to full OpenAPI parameter object.
     */
    public function toOpenApi(): array
    {
        $param = [
            'name' => $this->name,
            'in' => $this->in,
            'required' => $this->required || $this->in === 'path',
            'schema' => $this->toSchema(),
        ];

        if ($this->description !== null) {
            $param['description'] = $this->description;
        }

        return $param;
    }
}
