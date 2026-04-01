<?php

declare(strict_types=1);

namespace Core\Api\Documentation;

use Core\Api\Documentation\Attributes\ApiHidden;
use Core\Api\Documentation\Attributes\ApiParameter;
use Core\Api\Documentation\Attributes\ApiResponse;
use Core\Api\Documentation\Attributes\ApiSecurity;
use Core\Api\Documentation\Attributes\ApiTag;
use Core\Api\Documentation\Extensions\ApiKeyAuthExtension;
use Core\Api\Documentation\Extensions\RateLimitExtension;
use Core\Api\Documentation\Extensions\WorkspaceHeaderExtension;
use Illuminate\Http\Resources\Json\JsonResource;
use Illuminate\Routing\Route;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Route as RouteFacade;
use Illuminate\Support\Str;
use ReflectionAttribute;
use ReflectionClass;

/**
 * Enhanced OpenAPI Specification Builder.
 *
 * Builds comprehensive OpenAPI 3.1 specification from Laravel routes,
 * with support for custom attributes, module discovery, and extensions.
 */
class OpenApiBuilder
{
    /**
     * Registered extensions.
     *
     * @var array<Extension>
     */
    protected array $extensions = [];

    /**
     * Discovered tags from modules.
     *
     * @var array<string, array>
     */
    protected array $discoveredTags = [];

    /**
     * Create a new builder instance.
     */
    public function __construct()
    {
        $this->registerDefaultExtensions();
    }

    /**
     * Register default extensions.
     */
    protected function registerDefaultExtensions(): void
    {
        $this->extensions = [
            new WorkspaceHeaderExtension,
            new RateLimitExtension,
            new ApiKeyAuthExtension,
        ];
    }

    /**
     * Add a custom extension.
     */
    public function addExtension(Extension $extension): static
    {
        $this->extensions[] = $extension;

        return $this;
    }

    /**
     * Generate the complete OpenAPI specification.
     */
    public function build(): array
    {
        $config = config('api-docs', []);

        if ($this->shouldCache($config)) {
            $cacheKey = $config['cache']['key'] ?? 'api-docs:openapi';
            $cacheTtl = $config['cache']['ttl'] ?? 3600;

            return Cache::remember($cacheKey, $cacheTtl, fn () => $this->buildSpec($config));
        }

        return $this->buildSpec($config);
    }

    /**
     * Clear the cached specification.
     */
    public function clearCache(): void
    {
        $cacheKey = config('api-docs.cache.key', 'api-docs:openapi');
        Cache::forget($cacheKey);
    }

    /**
     * Check if caching should be enabled.
     */
    protected function shouldCache(array $config): bool
    {
        if (! ($config['cache']['enabled'] ?? true)) {
            return false;
        }

        $disabledEnvs = $config['cache']['disabled_environments'] ?? ['local', 'testing'];

        return ! in_array(app()->environment(), $disabledEnvs, true);
    }

    /**
     * Build the full OpenAPI specification.
     */
    protected function buildSpec(array $config): array
    {
        $spec = [
            'openapi' => '3.1.0',
            'info' => $this->buildInfo($config),
            'servers' => $this->buildServers($config),
            'tags' => [],
            'paths' => [],
            'components' => $this->buildComponents($config),
        ];

        // Build paths and collect tags
        $spec['paths'] = $this->buildPaths($config);
        $spec['tags'] = $this->buildTags($config);

        // Apply extensions to spec
        foreach ($this->extensions as $extension) {
            $spec = $extension->extend($spec, $config);
        }

        return $spec;
    }

    /**
     * Build API info section.
     */
    protected function buildInfo(array $config): array
    {
        $info = $config['info'] ?? [];

        $result = [
            'title' => $info['title'] ?? config('app.name', 'API').' API',
            'version' => $info['version'] ?? config('api.version', '1.0.0'),
        ];

        if (! empty($info['description'])) {
            $result['description'] = $info['description'];
        }

        if (! empty($info['contact'])) {
            $contact = array_filter($info['contact']);
            if (! empty($contact)) {
                $result['contact'] = $contact;
            }
        }

        if (! empty($info['license']['name'])) {
            $result['license'] = array_filter($info['license']);
        }

        return $result;
    }

    /**
     * Build servers section.
     */
    protected function buildServers(array $config): array
    {
        $servers = $config['servers'] ?? [];

        if (empty($servers)) {
            return [
                [
                    'url' => config('app.url', 'http://localhost'),
                    'description' => 'Current Environment',
                ],
            ];
        }

        return array_map(fn ($server) => array_filter($server), $servers);
    }

    /**
     * Build tags section from discovered modules and config.
     */
    protected function buildTags(array $config): array
    {
        $configTags = $config['tags'] ?? [];
        $tags = [];

        // Add discovered tags first
        foreach ($this->discoveredTags as $name => $data) {
            $tags[$name] = [
                'name' => $name,
                'description' => $data['description'] ?? null,
            ];
        }

        // Merge with configured tags (config takes precedence)
        foreach ($configTags as $key => $tagConfig) {
            $tagName = $tagConfig['name'] ?? $key;
            $tags[$tagName] = [
                'name' => $tagName,
                'description' => $tagConfig['description'] ?? null,
            ];
        }

        // Clean up null descriptions and sort
        $result = [];
        foreach ($tags as $tag) {
            $result[] = array_filter($tag);
        }

        usort($result, fn ($a, $b) => strcasecmp($a['name'], $b['name']));

        return $result;
    }

    /**
     * Build paths section from routes.
     */
    protected function buildPaths(array $config): array
    {
        $paths = [];
        $includePatterns = $config['routes']['include'] ?? ['api/*'];
        $excludePatterns = $config['routes']['exclude'] ?? [];

        foreach (RouteFacade::getRoutes() as $route) {
            /** @var Route $route */
            if (! $this->shouldIncludeRoute($route, $includePatterns, $excludePatterns)) {
                continue;
            }

            $path = $this->normalizePath($route->uri());
            $methods = array_filter($route->methods(), fn ($m) => $m !== 'HEAD');

            foreach ($methods as $method) {
                $method = strtolower($method);
                $operation = $this->buildOperation($route, $method, $config);

                if ($operation !== null) {
                    $paths[$path][$method] = $operation;
                }
            }
        }

        ksort($paths);

        return $paths;
    }

    /**
     * Check if a route should be included in documentation.
     */
    protected function shouldIncludeRoute(Route $route, array $include, array $exclude): bool
    {
        $uri = $route->uri();

        // Check exclusions first
        foreach ($exclude as $pattern) {
            if (fnmatch($pattern, $uri)) {
                return false;
            }
        }

        // Check inclusions
        foreach ($include as $pattern) {
            if (fnmatch($pattern, $uri)) {
                return true;
            }
        }

        return false;
    }

    /**
     * Normalize route path to OpenAPI format.
     */
    protected function normalizePath(string $uri): string
    {
        // Prepend slash if missing
        $path = '/'.ltrim($uri, '/');

        // Convert Laravel parameters to OpenAPI format: {param?} -> {param}
        $path = preg_replace('/\{([^}?]+)\?\}/', '{$1}', $path);

        return $path === '/' ? '/' : rtrim($path, '/');
    }

    /**
     * Build operation for a specific route and method.
     */
    protected function buildOperation(Route $route, string $method, array $config): ?array
    {
        $controller = $route->getController();
        $action = $route->getActionMethod();

        // Check for ApiHidden attribute
        if ($this->isHidden($controller, $action)) {
            return null;
        }

        $operation = [
            'summary' => $this->buildSummary($route, $method),
            'operationId' => $this->buildOperationId($route, $method),
            'tags' => $this->buildOperationTags($route, $controller, $action),
            'responses' => $this->buildResponses($controller, $action),
        ];

        // Add description from PHPDoc if available
        $description = $this->extractDescription($controller, $action);
        if ($description) {
            $operation['description'] = $description;
        }

        // Add parameters
        $parameters = $this->buildParameters($route, $controller, $action, $config);
        if (! empty($parameters)) {
            $operation['parameters'] = $parameters;
        }

        // Add request body for POST/PUT/PATCH
        if (in_array($method, ['post', 'put', 'patch'])) {
            $operation['requestBody'] = $this->buildRequestBody($route, $controller, $action);
        }

        // Add security requirements
        $security = $this->buildSecurity($route, $controller, $action);
        if ($security !== null) {
            $operation['security'] = $security;
        }

        // Apply extensions to operation
        foreach ($this->extensions as $extension) {
            $operation = $extension->extendOperation($operation, $route, $method, $config);
        }

        return $operation;
    }

    /**
     * Check if controller/method is hidden from docs.
     */
    protected function isHidden(?object $controller, string $action): bool
    {
        if ($controller === null) {
            return false;
        }

        $reflection = new ReflectionClass($controller);

        // Check class-level attribute
        $classAttrs = $reflection->getAttributes(ApiHidden::class);
        if (! empty($classAttrs)) {
            return true;
        }

        // Check method-level attribute
        if ($reflection->hasMethod($action)) {
            $method = $reflection->getMethod($action);
            $methodAttrs = $method->getAttributes(ApiHidden::class);
            if (! empty($methodAttrs)) {
                return true;
            }
        }

        return false;
    }

    /**
     * Build operation summary.
     */
    protected function buildSummary(Route $route, string $method): string
    {
        $name = $route->getName();

        if ($name) {
            // Convert route name to human-readable summary
            $parts = explode('.', $name);
            $action = array_pop($parts);

            return Str::title(str_replace(['-', '_'], ' ', $action));
        }

        // Generate from URI and method
        $uri = Str::afterLast($route->uri(), '/');

        return Str::title($method.' '.str_replace(['-', '_'], ' ', $uri));
    }

    /**
     * Build operation ID from route name.
     */
    protected function buildOperationId(Route $route, string $method): string
    {
        $name = $route->getName();

        if ($name) {
            return Str::camel(str_replace(['.', '-'], '_', $name));
        }

        return Str::camel($method.'_'.str_replace(['/', '-', '.'], '_', $route->uri()));
    }

    /**
     * Build tags for an operation.
     */
    protected function buildOperationTags(Route $route, ?object $controller, string $action): array
    {
        // Check for ApiTag attribute
        if ($controller !== null) {
            $tagAttr = $this->getAttribute($controller, $action, ApiTag::class);
            if ($tagAttr !== null) {
                $tag = $tagAttr->newInstance();
                $this->discoveredTags[$tag->name] = ['description' => $tag->description];

                return [$tag->name];
            }
        }

        // Infer tag from route
        return [$this->inferTag($route)];
    }

    /**
     * Infer tag from route.
     */
    protected function inferTag(Route $route): string
    {
        $uri = $route->uri();
        $name = $route->getName() ?? '';

        // Common tag mappings by route prefix
        $tagMap = [
            'api/bio' => 'Bio Links',
            'api/blocks' => 'Bio Links',
            'api/shortlinks' => 'Bio Links',
            'api/qr' => 'Bio Links',
            'api/commerce' => 'Commerce',
            'api/provisioning' => 'Commerce',
            'api/workspaces' => 'Workspaces',
            'api/analytics' => 'Analytics',
            'api/social' => 'Social',
            'api/notify' => 'Notifications',
            'api/support' => 'Support',
            'api/pixel' => 'Pixel',
            'api/seo' => 'SEO',
            'api/mcp' => 'MCP',
            'api/content' => 'Content',
            'api/trust' => 'Trust',
            'api/webhooks' => 'Webhooks',
            'api/entitlements' => 'Entitlements',
        ];

        foreach ($tagMap as $prefix => $tag) {
            if (str_starts_with($uri, $prefix)) {
                $this->discoveredTags[$tag] = $this->discoveredTags[$tag] ?? [];

                return $tag;
            }
        }

        $this->discoveredTags['General'] = $this->discoveredTags['General'] ?? [];

        return 'General';
    }

    /**
     * Extract description from PHPDoc.
     */
    protected function extractDescription(?object $controller, string $action): ?string
    {
        if ($controller === null) {
            return null;
        }

        $reflection = new ReflectionClass($controller);
        if (! $reflection->hasMethod($action)) {
            return null;
        }

        $method = $reflection->getMethod($action);
        $doc = $method->getDocComment();

        if (! $doc) {
            return null;
        }

        // Extract description from PHPDoc (first paragraph before @tags)
        preg_match('/\/\*\*\s*\n\s*\*\s*(.+?)(?:\n\s*\*\s*\n|\n\s*\*\s*@)/s', $doc, $matches);

        if (! empty($matches[1])) {
            $description = preg_replace('/\n\s*\*\s*/', ' ', $matches[1]);

            return trim($description);
        }

        return null;
    }

    /**
     * Build parameters for operation.
     */
    protected function buildParameters(Route $route, ?object $controller, string $action, array $config): array
    {
        $parameters = [];

        // Add path parameters
        preg_match_all('/\{([^}?]+)\??}/', $route->uri(), $matches);
        foreach ($matches[1] as $param) {
            $parameters[] = [
                'name' => $param,
                'in' => 'path',
                'required' => true,
                'schema' => ['type' => 'string'],
            ];
        }

        // Add parameters from ApiParameter attributes
        if ($controller !== null) {
            $reflection = new ReflectionClass($controller);
            if ($reflection->hasMethod($action)) {
                $method = $reflection->getMethod($action);
                $paramAttrs = $method->getAttributes(ApiParameter::class, ReflectionAttribute::IS_INSTANCEOF);

                foreach ($paramAttrs as $attr) {
                    $param = $attr->newInstance();
                    $parameters[] = $param->toOpenApi();
                }
            }
        }

        return $parameters;
    }

    /**
     * Build responses section.
     */
    protected function buildResponses(?object $controller, string $action): array
    {
        $responses = [];

        // Get ApiResponse attributes
        if ($controller !== null) {
            $reflection = new ReflectionClass($controller);
            if ($reflection->hasMethod($action)) {
                $method = $reflection->getMethod($action);
                $responseAttrs = $method->getAttributes(ApiResponse::class, ReflectionAttribute::IS_INSTANCEOF);

                foreach ($responseAttrs as $attr) {
                    $response = $attr->newInstance();
                    $responses[(string) $response->status] = $this->buildResponseSchema($response);
                }
            }
        }

        // Default 200 response if none specified
        if (empty($responses)) {
            $responses['200'] = ['description' => 'Successful response'];
        }

        return $responses;
    }

    /**
     * Build response schema from ApiResponse attribute.
     */
    protected function buildResponseSchema(ApiResponse $response): array
    {
        $result = [
            'description' => $response->getDescription(),
        ];

        if ($response->resource !== null && class_exists($response->resource)) {
            $schema = $this->extractResourceSchema($response->resource);

            if ($response->paginated) {
                $schema = $this->wrapPaginatedSchema($schema);
            }

            $result['content'] = [
                'application/json' => [
                    'schema' => $schema,
                ],
            ];
        }

        if (! empty($response->headers)) {
            $result['headers'] = [];
            foreach ($response->headers as $header => $description) {
                $result['headers'][$header] = [
                    'description' => $description,
                    'schema' => ['type' => 'string'],
                ];
            }
        }

        return $result;
    }

    /**
     * Extract schema from JsonResource class.
     */
    protected function extractResourceSchema(string $resourceClass): array
    {
        if (! is_subclass_of($resourceClass, JsonResource::class)) {
            return ['type' => 'object'];
        }

        // For now, return a generic object schema
        // A more sophisticated implementation would analyze the resource's toArray method
        return [
            'type' => 'object',
            'additionalProperties' => true,
        ];
    }

    /**
     * Wrap schema in pagination structure.
     */
    protected function wrapPaginatedSchema(array $itemSchema): array
    {
        return [
            'type' => 'object',
            'properties' => [
                'data' => [
                    'type' => 'array',
                    'items' => $itemSchema,
                ],
                'links' => [
                    'type' => 'object',
                    'properties' => [
                        'first' => ['type' => 'string', 'format' => 'uri'],
                        'last' => ['type' => 'string', 'format' => 'uri'],
                        'prev' => ['type' => 'string', 'format' => 'uri', 'nullable' => true],
                        'next' => ['type' => 'string', 'format' => 'uri', 'nullable' => true],
                    ],
                ],
                'meta' => [
                    'type' => 'object',
                    'properties' => [
                        'current_page' => ['type' => 'integer'],
                        'from' => ['type' => 'integer', 'nullable' => true],
                        'last_page' => ['type' => 'integer'],
                        'per_page' => ['type' => 'integer'],
                        'to' => ['type' => 'integer', 'nullable' => true],
                        'total' => ['type' => 'integer'],
                    ],
                ],
            ],
        ];
    }

    /**
     * Build request body schema.
     */
    protected function buildRequestBody(Route $route, ?object $controller, string $action): array
    {
        if ($controller instanceof \Core\Api\Controllers\McpApiController && $action === 'callTool') {
            return [
                'required' => true,
                'content' => [
                    'application/json' => [
                        'schema' => [
                            'type' => 'object',
                            'properties' => [
                                'server' => [
                                    'type' => 'string',
                                    'maxLength' => 64,
                                    'description' => 'MCP server identifier.',
                                ],
                                'tool' => [
                                    'type' => 'string',
                                    'maxLength' => 128,
                                    'description' => 'Tool name to invoke on the selected server.',
                                ],
                                'arguments' => [
                                    'type' => 'object',
                                    'description' => 'Tool arguments passed through to MCP.',
                                    'additionalProperties' => true,
                                ],
                                'version' => [
                                    'type' => 'string',
                                    'maxLength' => 32,
                                    'description' => 'Optional tool version to execute. Defaults to the latest supported version.',
                                ],
                            ],
                            'required' => ['server', 'tool'],
                            'additionalProperties' => true,
                        ],
                    ],
                ],
            ];
        }

        return [
            'required' => true,
            'content' => [
                'application/json' => [
                    'schema' => ['type' => 'object'],
                ],
            ],
        ];
    }

    /**
     * Build security requirements.
     */
    protected function buildSecurity(Route $route, ?object $controller, string $action): ?array
    {
        // Check for ApiSecurity attribute
        if ($controller !== null) {
            $securityAttr = $this->getAttribute($controller, $action, ApiSecurity::class);
            if ($securityAttr !== null) {
                $security = $securityAttr->newInstance();
                if ($security->isPublic()) {
                    return []; // Empty array means no auth required
                }

                return [[$security->scheme => $security->scopes]];
            }
        }

        // Infer from route middleware
        $middleware = $route->middleware();

        if (in_array('auth:sanctum', $middleware) || in_array('auth', $middleware)) {
            return [['bearerAuth' => []]];
        }

        if (in_array('api.auth', $middleware) || in_array('auth.api', $middleware)) {
            return [['apiKeyAuth' => []]];
        }

        foreach ($middleware as $m) {
            if (str_contains($m, 'ApiKeyAuth') || str_contains($m, 'AuthenticateApiKey')) {
                return [['apiKeyAuth' => []]];
            }
        }

        return null;
    }

    /**
     * Build components section.
     */
    protected function buildComponents(array $config): array
    {
        $components = [
            'securitySchemes' => [],
            'schemas' => $this->buildCommonSchemas(),
        ];

        // Add API Key security scheme
        $apiKeyConfig = $config['auth']['api_key'] ?? [];
        if ($apiKeyConfig['enabled'] ?? true) {
            $components['securitySchemes']['apiKeyAuth'] = [
                'type' => 'apiKey',
                'in' => $apiKeyConfig['in'] ?? 'header',
                'name' => $apiKeyConfig['name'] ?? 'X-API-Key',
                'description' => $apiKeyConfig['description'] ?? 'API key for authentication',
            ];
        }

        // Add Bearer token security scheme
        $bearerConfig = $config['auth']['bearer'] ?? [];
        if ($bearerConfig['enabled'] ?? true) {
            $components['securitySchemes']['bearerAuth'] = [
                'type' => 'http',
                'scheme' => $bearerConfig['scheme'] ?? 'bearer',
                'bearerFormat' => $bearerConfig['format'] ?? 'JWT',
                'description' => $bearerConfig['description'] ?? 'Bearer token authentication',
            ];
        }

        // Add OAuth2 security scheme
        $oauth2Config = $config['auth']['oauth2'] ?? [];
        if ($oauth2Config['enabled'] ?? false) {
            $components['securitySchemes']['oauth2'] = [
                'type' => 'oauth2',
                'flows' => $oauth2Config['flows'] ?? [],
            ];
        }

        return $components;
    }

    /**
     * Build common reusable schemas.
     */
    protected function buildCommonSchemas(): array
    {
        return [
            'Error' => [
                'type' => 'object',
                'required' => ['message'],
                'properties' => [
                    'message' => ['type' => 'string', 'description' => 'Error message'],
                    'errors' => [
                        'type' => 'object',
                        'description' => 'Validation errors (field => messages)',
                        'additionalProperties' => [
                            'type' => 'array',
                            'items' => ['type' => 'string'],
                        ],
                    ],
                ],
            ],
            'Pagination' => [
                'type' => 'object',
                'properties' => [
                    'current_page' => ['type' => 'integer'],
                    'from' => ['type' => 'integer', 'nullable' => true],
                    'last_page' => ['type' => 'integer'],
                    'per_page' => ['type' => 'integer'],
                    'to' => ['type' => 'integer', 'nullable' => true],
                    'total' => ['type' => 'integer'],
                ],
            ],
        ];
    }

    /**
     * Get attribute from controller class or method.
     *
     * @template T
     *
     * @param  class-string<T>  $attributeClass
     * @return ReflectionAttribute<T>|null
     */
    protected function getAttribute(object $controller, string $action, string $attributeClass): ?ReflectionAttribute
    {
        $reflection = new ReflectionClass($controller);

        // Check method first (method takes precedence)
        if ($reflection->hasMethod($action)) {
            $method = $reflection->getMethod($action);
            $attrs = $method->getAttributes($attributeClass);
            if (! empty($attrs)) {
                return $attrs[0];
            }
        }

        // Fall back to class
        $attrs = $reflection->getAttributes($attributeClass);

        return $attrs[0] ?? null;
    }
}
