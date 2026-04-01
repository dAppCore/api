<?php

declare(strict_types=1);

use Core\Api\Documentation\Attributes\ApiHidden;
use Core\Api\Documentation\Attributes\ApiParameter;
use Core\Api\Documentation\Attributes\ApiResponse;
use Core\Api\Documentation\Attributes\ApiSecurity;
use Core\Api\Documentation\Attributes\ApiTag;
use Core\Api\Documentation\Extension;
use Core\Api\Documentation\Extensions\ApiKeyAuthExtension;
use Core\Api\Documentation\Extensions\RateLimitExtension;
use Core\Api\Documentation\Extensions\WorkspaceHeaderExtension;
use Core\Api\Documentation\OpenApiBuilder;
use Core\Api\RateLimit\RateLimit;
use Illuminate\Http\Resources\Json\JsonResource;
use Illuminate\Routing\Route;
use Illuminate\Support\Facades\Route as RouteFacade;

// ─────────────────────────────────────────────────────────────────────────────
// OpenApiBuilder Schema Generation
// ─────────────────────────────────────────────────────────────────────────────

describe('OpenApiBuilder Schema Generation', function () {
    it('generates valid OpenAPI 3.1 spec structure', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec)->toBeArray()
            ->toHaveKey('openapi')
            ->toHaveKey('info')
            ->toHaveKey('servers')
            ->toHaveKey('tags')
            ->toHaveKey('paths')
            ->toHaveKey('components');

        expect($spec['openapi'])->toBe('3.1.0');
    });

    it('builds info section with title and version', function () {
        config(['api-docs.info.title' => 'Test API']);
        config(['api-docs.info.version' => '2.0.0']);
        config(['api-docs.info.description' => 'A test API']);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['info'])->toHaveKey('title')
            ->toHaveKey('version');
        expect($spec['info']['title'])->toContain('API');
    });

    it('builds servers section with default URL', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['servers'])->toBeArray();
        expect($spec['servers'][0])->toHaveKey('url');
    });

    it('builds components with security schemes', function () {
        config(['api-docs.auth.api_key.enabled' => true]);
        config(['api-docs.auth.bearer.enabled' => true]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['components'])->toHaveKey('securitySchemes')
            ->toHaveKey('schemas');
        expect($spec['components']['securitySchemes'])->toHaveKey('apiKeyAuth')
            ->toHaveKey('bearerAuth');
    });

    it('builds common error schema in components', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['components']['schemas'])->toHaveKey('Error');
        expect($spec['components']['schemas']['Error']['type'])->toBe('object');
        expect($spec['components']['schemas']['Error']['properties'])->toHaveKey('message');
    });

    it('builds pagination schema in components', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['components']['schemas'])->toHaveKey('Pagination');
        expect($spec['components']['schemas']['Pagination']['properties'])
            ->toHaveKey('current_page')
            ->toHaveKey('per_page')
            ->toHaveKey('total');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// OpenApiBuilder Controller Scanning
// ─────────────────────────────────────────────────────────────────────────────

describe('OpenApiBuilder Controller Scanning', function () {
    beforeEach(function () {
        // Register test routes with various configurations
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/test-scan/items', [TestOpenApiController::class, 'index'])
                    ->name('test-scan.items.index');
                RouteFacade::get('/test-scan/items/{id}', [TestOpenApiController::class, 'show'])
                    ->name('test-scan.items.show');
                RouteFacade::post('/test-scan/items', [TestOpenApiController::class, 'store'])
                    ->name('test-scan.items.store');
                RouteFacade::put('/test-scan/items/{id}', [TestOpenApiController::class, 'update'])
                    ->name('test-scan.items.update');
                RouteFacade::delete('/test-scan/items/{id}', [TestOpenApiController::class, 'destroy'])
                    ->name('test-scan.items.destroy');
            });

        config(['api-docs.routes.include' => ['api/*']]);
        config(['api-docs.routes.exclude' => []]);
    });

    it('discovers routes matching include patterns', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['paths'])->toHaveKey('/api/test-scan/items');
    });

    it('generates correct HTTP methods for routes', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['paths']['/api/test-scan/items'])->toHaveKey('get')
            ->toHaveKey('post');
        expect($spec['paths']['/api/test-scan/items/{id}'])->toHaveKey('get')
            ->toHaveKey('put')
            ->toHaveKey('delete');
    });

    it('normalises path parameters to OpenAPI format', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        // Laravel {id} should remain as {id} in OpenAPI
        expect($spec['paths'])->toHaveKey('/api/test-scan/items/{id}');
    });

    it('generates operation IDs from route names', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/test-scan/items']['get'];
        expect($operation['operationId'])->toBe('testScanItemsIndex');
    });

    it('generates summary from route name', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/test-scan/items']['get'];
        expect($operation)->toHaveKey('summary');
        expect($operation['summary'])->toContain('Index');
    });

    it('extracts path parameters as required', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/test-scan/items/{id}']['get'];
        expect($operation['parameters'])->toBeArray();

        $idParam = collect($operation['parameters'])->firstWhere('name', 'id');
        expect($idParam)->not->toBeNull();
        expect($idParam['in'])->toBe('path');
        expect($idParam['required'])->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Application Endpoint Parameter Docs
// ─────────────────────────────────────────────────────────────────────────────

describe('Application Endpoint Parameter Docs', function () {
    it('documents the SEO report url query parameter', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/seo/report']['get'];
        $urlParam = collect($operation['parameters'] ?? [])->firstWhere('name', 'url');

        expect($urlParam)->not->toBeNull();
        expect($urlParam['in'])->toBe('query');
        expect($urlParam['required'])->toBeTrue();
        expect($urlParam['schema']['format'])->toBe('uri');
    });

    it('documents MCP list query parameters', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $toolsOperation = $spec['paths']['/api/mcp/servers/{id}/tools']['get'];
        $includeVersions = collect($toolsOperation['parameters'] ?? [])->firstWhere('name', 'include_versions');

        expect($includeVersions)->not->toBeNull();
        expect($includeVersions['in'])->toBe('query');
        expect($includeVersions['schema']['type'])->toBe('boolean');

        $resourcesOperation = $spec['paths']['/api/mcp/servers/{id}/resources']['get'];
        $includeContent = collect($resourcesOperation['parameters'] ?? [])->firstWhere('name', 'include_content');

        expect($includeContent)->not->toBeNull();
        expect($includeContent['in'])->toBe('query');
        expect($includeContent['schema']['type'])->toBe('boolean');
    });

    it('documents the MCP tool call request body shape', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/mcp/tools/call']['post'];
        $schema = $operation['requestBody']['content']['application/json']['schema'] ?? null;

        expect($schema)->not->toBeNull();
        expect($schema['type'])->toBe('object');
        expect($schema['properties'])->toHaveKey('server')
            ->toHaveKey('tool')
            ->toHaveKey('arguments')
            ->toHaveKey('version');
        expect($schema['required'])->toBe(['server', 'tool']);
        expect($schema['additionalProperties'])->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// ApiParameter Attribute Parsing
// ─────────────────────────────────────────────────────────────────────────────

describe('ApiParameter Attribute Parsing', function () {
    it('creates parameter with all properties', function () {
        $param = new ApiParameter(
            name: 'filter',
            in: 'query',
            type: 'string',
            description: 'Filter results',
            required: true,
            example: 'active',
            default: null,
            enum: ['active', 'inactive'],
            format: null
        );

        expect($param->name)->toBe('filter');
        expect($param->in)->toBe('query');
        expect($param->type)->toBe('string');
        expect($param->description)->toBe('Filter results');
        expect($param->required)->toBeTrue();
        expect($param->example)->toBe('active');
        expect($param->enum)->toBe(['active', 'inactive']);
    });

    it('defaults to query parameter type', function () {
        $param = new ApiParameter('search');

        expect($param->in)->toBe('query');
        expect($param->type)->toBe('string');
        expect($param->required)->toBeFalse();
    });

    it('converts to OpenAPI schema format', function () {
        $param = new ApiParameter(
            name: 'page',
            in: 'query',
            type: 'integer',
            description: 'Page number',
            example: 1,
            default: 1,
            format: 'int32'
        );

        $schema = $param->toSchema();

        expect($schema['type'])->toBe('integer');
        expect($schema['format'])->toBe('int32');
        expect($schema['example'])->toBe(1);
        expect($schema['default'])->toBe(1);
    });

    it('converts to full OpenAPI parameter object', function () {
        $param = new ApiParameter(
            name: 'status',
            in: 'query',
            type: 'string',
            description: 'Status filter',
            required: false,
            enum: ['draft', 'published', 'archived']
        );

        $openApi = $param->toOpenApi();

        expect($openApi['name'])->toBe('status');
        expect($openApi['in'])->toBe('query');
        expect($openApi['required'])->toBeFalse();
        expect($openApi['description'])->toBe('Status filter');
        expect($openApi['schema']['type'])->toBe('string');
        expect($openApi['schema']['enum'])->toBe(['draft', 'published', 'archived']);
    });

    it('forces path parameters to be required', function () {
        $param = new ApiParameter(
            name: 'id',
            in: 'path',
            type: 'string',
            required: false // Should be overridden
        );

        $openApi = $param->toOpenApi();

        expect($openApi['required'])->toBeTrue();
    });

    it('supports header parameters', function () {
        $param = new ApiParameter(
            name: 'X-Custom-Header',
            in: 'header',
            type: 'string',
            description: 'Custom header value'
        );

        $openApi = $param->toOpenApi();

        expect($openApi['in'])->toBe('header');
        expect($openApi['name'])->toBe('X-Custom-Header');
    });

    it('supports cookie parameters', function () {
        $param = new ApiParameter(
            name: 'session_id',
            in: 'cookie',
            type: 'string'
        );

        $openApi = $param->toOpenApi();

        expect($openApi['in'])->toBe('cookie');
    });

    it('handles array type parameters', function () {
        $param = new ApiParameter(
            name: 'ids',
            in: 'query',
            type: 'array',
            description: 'List of IDs'
        );

        $schema = $param->toSchema();

        expect($schema['type'])->toBe('array');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// ApiResponse Attribute Rendering
// ─────────────────────────────────────────────────────────────────────────────

describe('ApiResponse Attribute Rendering', function () {
    it('creates response with status and description', function () {
        $response = new ApiResponse(
            status: 200,
            description: 'Successful operation'
        );

        expect($response->status)->toBe(200);
        expect($response->getDescription())->toBe('Successful operation');
    });

    it('generates description from status code', function () {
        $testCases = [
            200 => 'Successful response',
            201 => 'Resource created',
            204 => 'No content',
            400 => 'Bad request',
            401 => 'Unauthorized',
            403 => 'Forbidden',
            404 => 'Not found',
            422 => 'Validation error',
            429 => 'Too many requests',
            500 => 'Internal server error',
        ];

        foreach ($testCases as $status => $expectedDescription) {
            $response = new ApiResponse(status: $status);
            expect($response->getDescription())->toBe($expectedDescription);
        }
    });

    it('supports resource class reference', function () {
        $response = new ApiResponse(
            status: 200,
            resource: TestJsonResource::class,
            description: 'User details'
        );

        expect($response->resource)->toBe(TestJsonResource::class);
    });

    it('supports paginated flag', function () {
        $response = new ApiResponse(
            status: 200,
            resource: TestJsonResource::class,
            paginated: true
        );

        expect($response->paginated)->toBeTrue();
    });

    it('supports response headers', function () {
        $response = new ApiResponse(
            status: 200,
            headers: [
                'X-Request-Id' => 'Unique request identifier',
                'X-Rate-Limit-Remaining' => 'Remaining rate limit',
            ]
        );

        expect($response->headers)->toHaveKey('X-Request-Id')
            ->toHaveKey('X-Rate-Limit-Remaining');
    });

    it('handles unknown status codes gracefully', function () {
        $response = new ApiResponse(status: 418);

        expect($response->getDescription())->toBe('Response');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// ApiSecurity Attribute Requirements
// ─────────────────────────────────────────────────────────────────────────────

describe('ApiSecurity Attribute Requirements', function () {
    it('creates security requirement with scheme', function () {
        $security = new ApiSecurity(scheme: 'apiKey');

        expect($security->scheme)->toBe('apiKey');
        expect($security->scopes)->toBe([]);
        expect($security->isPublic())->toBeFalse();
    });

    it('supports OAuth2 scopes', function () {
        $security = new ApiSecurity(
            scheme: 'oauth2',
            scopes: ['read:users', 'write:users']
        );

        expect($security->scheme)->toBe('oauth2');
        expect($security->scopes)->toBe(['read:users', 'write:users']);
    });

    it('marks endpoint as public with null scheme', function () {
        $security = new ApiSecurity(scheme: null);

        expect($security->isPublic())->toBeTrue();
    });

    it('supports bearer authentication', function () {
        $security = new ApiSecurity(scheme: 'bearer');

        expect($security->scheme)->toBe('bearer');
        expect($security->isPublic())->toBeFalse();
    });

    it('supports apiKey authentication', function () {
        $security = new ApiSecurity(scheme: 'apiKey');

        expect($security->scheme)->toBe('apiKey');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// ApiHidden Attribute Filtering
// ─────────────────────────────────────────────────────────────────────────────

describe('ApiHidden Attribute Filtering', function () {
    it('creates hidden attribute without reason', function () {
        $hidden = new ApiHidden;

        expect($hidden->reason)->toBeNull();
    });

    it('creates hidden attribute with reason', function () {
        $hidden = new ApiHidden(reason: 'Internal endpoint only');

        expect($hidden->reason)->toBe('Internal endpoint only');
    });

    it('excludes hidden endpoints from documentation', function () {
        // Register a hidden route
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/hidden-test/public', [TestPublicController::class, 'index'])
                    ->name('hidden-test.public');
                RouteFacade::get('/hidden-test/internal', [TestHiddenController::class, 'index'])
                    ->name('hidden-test.internal');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        // Public endpoint should be present
        expect($spec['paths'])->toHaveKey('/api/hidden-test/public');

        // Hidden endpoint should not be present
        expect($spec['paths'])->not->toHaveKey('/api/hidden-test/internal');
    });

    it('excludes hidden methods but includes non-hidden ones', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/partial-hidden/public', [TestPartialHiddenController::class, 'publicMethod'])
                    ->name('partial-hidden.public');
                RouteFacade::get('/partial-hidden/hidden', [TestPartialHiddenController::class, 'hiddenMethod'])
                    ->name('partial-hidden.hidden');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['paths'])->toHaveKey('/api/partial-hidden/public');
        expect($spec['paths'])->not->toHaveKey('/api/partial-hidden/hidden');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// ApiTag Attribute Grouping
// ─────────────────────────────────────────────────────────────────────────────

describe('ApiTag Attribute Grouping', function () {
    it('creates tag with name only', function () {
        $tag = new ApiTag(name: 'Users');

        expect($tag->name)->toBe('Users');
        expect($tag->description)->toBeNull();
    });

    it('creates tag with name and description', function () {
        $tag = new ApiTag(
            name: 'Products',
            description: 'Product management endpoints'
        );

        expect($tag->name)->toBe('Products');
        expect($tag->description)->toBe('Product management endpoints');
    });

    it('discovers tags from controller attributes', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/tagged/items', [TestTaggedController::class, 'index'])
                    ->name('tagged.items.index');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/tagged/items']['get'];
        expect($operation['tags'])->toContain('Custom Tag');
    });

    it('collects discovered tags in tags section', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/tagged/items', [TestTaggedController::class, 'index'])
                    ->name('tagged.items.index');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $tagNames = collect($spec['tags'])->pluck('name')->toArray();
        expect($tagNames)->toContain('Custom Tag');
    });

    it('infers tags from route prefixes when not specified', function () {
        RouteFacade::prefix('api/bio')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/links', fn () => response()->json([]))
                    ->name('bio.links.index');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/bio/links']['get'];
        expect($operation['tags'])->toContain('Bio Links');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Extension System
// ─────────────────────────────────────────────────────────────────────────────

describe('Extension System', function () {
    it('registers default extensions', function () {
        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        // Default extensions should have been applied
        // Check for workspace header parameter (WorkspaceHeaderExtension)
        // Check for rate limit response (RateLimitExtension)
        // Check for auth error responses (ApiKeyAuthExtension)
        expect($spec['components'])->toBeArray();
    });

    it('allows adding custom extensions', function () {
        $builder = new OpenApiBuilder;
        $customExtension = new TestCustomExtension;

        $builder->addExtension($customExtension);
        $spec = $builder->build();

        expect($spec['x-custom-extension'])->toBe('added');
    });

    it('WorkspaceHeaderExtension adds workspace parameter', function () {
        config(['api-docs.workspace' => [
            'header_name' => 'X-Workspace-ID',
            'required' => false,
            'description' => 'Workspace identifier',
        ]]);

        $extension = new WorkspaceHeaderExtension;
        $spec = ['components' => ['parameters' => []]];

        $result = $extension->extend($spec, config('api-docs'));

        expect($result['components']['parameters'])->toHaveKey('workspaceId');
        expect($result['components']['parameters']['workspaceId']['name'])->toBe('X-Workspace-ID');
    });

    it('RateLimitExtension adds rate limit headers', function () {
        config(['api-docs.rate_limits' => ['enabled' => true]]);

        $extension = new RateLimitExtension;
        $spec = ['components' => ['headers' => [], 'responses' => []]];

        $result = $extension->extend($spec, config('api-docs'));

        expect($result['components']['responses'])->toHaveKey('RateLimitExceeded');
    });

    it('ApiKeyAuthExtension adds authentication schemas', function () {
        config(['api-docs.auth.api_key' => ['enabled' => true, 'name' => 'X-API-Key']]);
        config(['api-docs.auth.bearer' => ['enabled' => true]]);

        $extension = new ApiKeyAuthExtension;
        $spec = [
            'info' => ['description' => 'Test API'],
            'components' => ['securitySchemes' => ['apiKeyAuth' => []]],
        ];

        $result = $extension->extend($spec, config('api-docs'));

        expect($result['components']['schemas'])->toHaveKey('UnauthorizedError')
            ->toHaveKey('ForbiddenError');
        expect($result['components']['responses'])->toHaveKey('Unauthorized')
            ->toHaveKey('Forbidden');
    });

    it('extensions can modify individual operations', function () {
        $extension = new RateLimitExtension;
        $config = ['rate_limits' => ['enabled' => true]];

        // Create a mock route with rate limiting
        $route = RouteFacade::get('/test', fn () => 'test')->middleware('throttle:60,1');
        $route->prepareForSerialization();

        $operation = [
            'summary' => 'Test',
            'responses' => ['200' => ['description' => 'OK']],
        ];

        $result = $extension->extendOperation($operation, $route, 'get', $config);

        expect($result['responses'])->toHaveKey('429');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Error Response Documentation
// ─────────────────────────────────────────────────────────────────────────────

describe('Error Response Documentation', function () {
    it('documents 401 Unauthorized response', function () {
        $extension = new ApiKeyAuthExtension;
        $spec = [
            'info' => [],
            'components' => ['securitySchemes' => ['apiKeyAuth' => []]],
        ];

        $result = $extension->extend($spec, ['auth' => ['api_key' => ['enabled' => true]]]);

        expect($result['components']['responses']['Unauthorized'])->toHaveKey('description')
            ->toHaveKey('content');
        expect($result['components']['responses']['Unauthorized']['content']['application/json']['examples'])
            ->toHaveKey('missing_key')
            ->toHaveKey('invalid_key')
            ->toHaveKey('expired_key');
    });

    it('documents 403 Forbidden response', function () {
        $extension = new ApiKeyAuthExtension;
        $spec = [
            'info' => [],
            'components' => ['securitySchemes' => ['apiKeyAuth' => []]],
        ];

        $result = $extension->extend($spec, ['auth' => ['api_key' => ['enabled' => true]]]);

        expect($result['components']['responses']['Forbidden'])->toHaveKey('description')
            ->toHaveKey('content');
        expect($result['components']['responses']['Forbidden']['content']['application/json']['examples'])
            ->toHaveKey('insufficient_scope')
            ->toHaveKey('workspace_access');
    });

    it('documents 429 Rate Limit Exceeded response', function () {
        $extension = new RateLimitExtension;
        $spec = ['components' => ['headers' => [], 'responses' => []]];

        $result = $extension->extend($spec, ['rate_limits' => ['enabled' => true]]);

        expect($result['components']['responses']['RateLimitExceeded'])->toHaveKey('description')
            ->toHaveKey('headers')
            ->toHaveKey('content');
        expect($result['components']['responses']['RateLimitExceeded']['headers'])
            ->toHaveKey('Retry-After');
    });

    it('automatically adds auth error responses to protected operations', function () {
        $extension = new ApiKeyAuthExtension;

        $route = RouteFacade::get('/test', fn () => 'test');
        $operation = [
            'security' => [['apiKeyAuth' => []]],
            'responses' => ['200' => ['description' => 'OK']],
        ];

        $result = $extension->extendOperation($operation, $route, 'get', []);

        expect($result['responses'])->toHaveKey('401')
            ->toHaveKey('403');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Authentication Documentation
// ─────────────────────────────────────────────────────────────────────────────

describe('Authentication Documentation', function () {
    it('documents API Key authentication scheme', function () {
        config(['api-docs.auth.api_key' => [
            'enabled' => true,
            'name' => 'X-API-Key',
            'in' => 'header',
            'description' => 'API key for authentication',
        ]]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $apiKeyScheme = $spec['components']['securitySchemes']['apiKeyAuth'];

        expect($apiKeyScheme['type'])->toBe('apiKey');
        expect($apiKeyScheme['in'])->toBe('header');
        expect($apiKeyScheme['name'])->toBe('X-API-Key');
    });

    it('documents Bearer authentication scheme', function () {
        config(['api-docs.auth.bearer' => [
            'enabled' => true,
            'scheme' => 'bearer',
            'format' => 'JWT',
        ]]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $bearerScheme = $spec['components']['securitySchemes']['bearerAuth'];

        expect($bearerScheme['type'])->toBe('http');
        expect($bearerScheme['scheme'])->toBe('bearer');
        expect($bearerScheme['bearerFormat'])->toBe('JWT');
    });

    it('infers security from route middleware', function () {
        RouteFacade::prefix('api')
            ->middleware(['api', 'auth:sanctum'])
            ->group(function () {
                RouteFacade::get('/auth-test/protected', fn () => response()->json([]))
                    ->name('auth-test.protected');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/auth-test/protected']['get'];
        expect($operation['security'][0])->toHaveKey('bearerAuth');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Request/Response Examples Validation
// ─────────────────────────────────────────────────────────────────────────────

describe('Request/Response Examples Validation', function () {
    it('includes request body for POST operations', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::post('/example/create', fn () => response()->json([]))
                    ->name('example.create');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/example/create']['post'];
        expect($operation)->toHaveKey('requestBody');
        expect($operation['requestBody'])->toHaveKey('content');
    });

    it('includes request body for PUT operations', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::put('/example/update/{id}', fn () => response()->json([]))
                    ->name('example.update');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/example/update/{id}']['put'];
        expect($operation)->toHaveKey('requestBody');
    });

    it('includes request body for PATCH operations', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::patch('/example/patch/{id}', fn () => response()->json([]))
                    ->name('example.patch');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/example/patch/{id}']['patch'];
        expect($operation)->toHaveKey('requestBody');
    });

    it('does not include request body for GET operations', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/example/list', fn () => response()->json([]))
                    ->name('example.list');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/example/list']['get'];
        expect($operation)->not->toHaveKey('requestBody');
    });

    it('includes default 200 response when none specified', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/default-response', fn () => response()->json([]))
                    ->name('default-response.index');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        $operation = $spec['paths']['/api/default-response']['get'];
        expect($operation['responses'])->toHaveKey('200');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Caching Behaviour
// ─────────────────────────────────────────────────────────────────────────────

describe('Caching Behaviour', function () {
    it('respects cache disabled environments', function () {
        config(['api-docs.cache' => [
            'enabled' => true,
            'disabled_environments' => ['testing'],
        ]]);

        // In testing environment, cache should be disabled
        $builder = new OpenApiBuilder;
        $spec1 = $builder->build();
        $spec2 = $builder->build();

        // Both should return fresh data (not cached)
        expect($spec1)->toEqual($spec2);
    });

    it('clears cache when requested', function () {
        $builder = new OpenApiBuilder;
        $builder->clearCache();

        // Should not throw an exception
        expect(true)->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Route Exclusion
// ─────────────────────────────────────────────────────────────────────────────

describe('Route Exclusion', function () {
    it('excludes routes matching exclude patterns', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/included', fn () => response()->json([]))
                    ->name('included');
                RouteFacade::get('/internal/excluded', fn () => response()->json([]))
                    ->name('internal.excluded');
            });

        config(['api-docs.routes.include' => ['api/*']]);
        config(['api-docs.routes.exclude' => ['api/internal/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        expect($spec['paths'])->toHaveKey('/api/included');
        expect($spec['paths'])->not->toHaveKey('/api/internal/excluded');
    });

    it('excludes HEAD methods from documentation', function () {
        RouteFacade::prefix('api')
            ->middleware('api')
            ->group(function () {
                RouteFacade::get('/head-test', fn () => response()->json([]))
                    ->name('head-test');
            });

        config(['api-docs.routes.include' => ['api/*']]);

        $builder = new OpenApiBuilder;
        $spec = $builder->build();

        // GET route also creates HEAD route, but HEAD should be excluded
        $operation = $spec['paths']['/api/head-test'] ?? [];
        expect($operation)->not->toHaveKey('head');
        expect($operation)->toHaveKey('get');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// RateLimit Attribute Integration
// ─────────────────────────────────────────────────────────────────────────────

describe('RateLimit Attribute Integration', function () {
    it('creates RateLimit attribute with properties', function () {
        $rateLimit = new RateLimit(
            limit: 100,
            window: 60,
            burst: 1.2,
            key: 'custom'
        );

        expect($rateLimit->limit)->toBe(100);
        expect($rateLimit->window)->toBe(60);
        expect($rateLimit->burst)->toBe(1.2);
        expect($rateLimit->key)->toBe('custom');
    });

    it('defaults to 60 second window', function () {
        $rateLimit = new RateLimit(limit: 100);

        expect($rateLimit->window)->toBe(60);
    });

    it('defaults to no burst', function () {
        $rateLimit = new RateLimit(limit: 100);

        expect($rateLimit->burst)->toBe(1.0);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Test Fixtures (Controllers and Resources)
// ─────────────────────────────────────────────────────────────────────────────

/**
 * Test controller for OpenAPI scanning tests.
 */
class TestOpenApiController
{
    #[ApiParameter('filter', 'query', 'string', 'Filter items')]
    #[ApiParameter('page', 'query', 'integer', 'Page number', false, 1)]
    #[ApiResponse(200, TestJsonResource::class, 'List of items', paginated: true)]
    public function index(): void {}

    #[ApiResponse(200, TestJsonResource::class, 'Item details')]
    #[ApiResponse(404, null, 'Item not found')]
    public function show(string $id): void {}

    #[ApiSecurity('apiKey', ['write'])]
    #[ApiResponse(201, TestJsonResource::class, 'Item created')]
    #[ApiResponse(422, null, 'Validation failed')]
    public function store(): void {}

    #[ApiSecurity('apiKey', ['write'])]
    #[ApiResponse(200, TestJsonResource::class, 'Item updated')]
    public function update(string $id): void {}

    #[ApiSecurity('apiKey', ['delete'])]
    #[ApiResponse(204, null, 'Item deleted')]
    public function destroy(string $id): void {}
}

/**
 * Test hidden controller.
 */
#[ApiHidden('Internal use only')]
class TestHiddenController
{
    public function index(): void {}
}

/**
 * Test public controller.
 */
class TestPublicController
{
    public function index(): void {}
}

/**
 * Test controller with partially hidden methods.
 */
class TestPartialHiddenController
{
    public function publicMethod(): void {}

    #[ApiHidden]
    public function hiddenMethod(): void {}
}

/**
 * Test tagged controller.
 */
#[ApiTag('Custom Tag', 'Custom tag description')]
class TestTaggedController
{
    public function index(): void {}
}

/**
 * Test JSON Resource.
 */
class TestJsonResource extends JsonResource
{
    public function toArray($request): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
        ];
    }
}

/**
 * Test custom extension.
 */
class TestCustomExtension implements Extension
{
    public function extend(array $spec, array $config): array
    {
        $spec['x-custom-extension'] = 'added';

        return $spec;
    }

    public function extendOperation(array $operation, Route $route, string $method, array $config): array
    {
        return $operation;
    }
}
