<?php

declare(strict_types=1);

namespace Core\Api\Tests\Feature;

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
use Orchestra\Testbench\TestCase;

/**
 * Test OpenAPI documentation generation.
 */
class OpenApiDocumentationTest extends TestCase
{
    public function test_openapi_builder_can_be_instantiated(): void
    {
        $builder = new OpenApiBuilder;

        $this->assertInstanceOf(OpenApiBuilder::class, $builder);
    }

    public function test_extensions_implement_interface(): void
    {
        $this->assertInstanceOf(Extension::class, new WorkspaceHeaderExtension);
        $this->assertInstanceOf(Extension::class, new RateLimitExtension);
        $this->assertInstanceOf(Extension::class, new ApiKeyAuthExtension);
    }

    public function test_api_tag_attribute(): void
    {
        $tag = new ApiTag('Users', 'User management');

        $this->assertEquals('Users', $tag->name);
        $this->assertEquals('User management', $tag->description);
    }

    public function test_api_response_attribute(): void
    {
        $response = new ApiResponse(200, null, 'Success');

        $this->assertEquals(200, $response->status);
        $this->assertEquals('Success', $response->getDescription());
        $this->assertFalse($response->paginated);
    }

    public function test_api_response_generates_description_from_status(): void
    {
        $response = new ApiResponse(404);

        $this->assertEquals('Not found', $response->getDescription());

        $goneResponse = new ApiResponse(410);

        $this->assertEquals('Gone', $goneResponse->getDescription());
    }

    public function test_api_security_attribute(): void
    {
        $security = new ApiSecurity('apiKey', ['read', 'write']);

        $this->assertEquals('apiKey', $security->scheme);
        $this->assertEquals(['read', 'write'], $security->scopes);
        $this->assertFalse($security->isPublic());
    }

    public function test_api_security_public(): void
    {
        $security = new ApiSecurity(null);

        $this->assertTrue($security->isPublic());
    }

    public function test_api_parameter_attribute(): void
    {
        $param = new ApiParameter(
            name: 'page',
            in: 'query',
            type: 'integer',
            description: 'Page number',
            required: false,
            example: 1
        );

        $this->assertEquals('page', $param->name);
        $this->assertEquals('query', $param->in);
        $this->assertEquals('integer', $param->type);
        $this->assertEquals(1, $param->example);
    }

    public function test_api_parameter_to_openapi(): void
    {
        $param = new ApiParameter(
            name: 'page',
            in: 'query',
            type: 'integer',
            description: 'Page number',
            required: false,
            example: 1
        );

        $openApi = $param->toOpenApi();

        $this->assertEquals('page', $openApi['name']);
        $this->assertEquals('query', $openApi['in']);
        $this->assertFalse($openApi['required']);
        $this->assertEquals('integer', $openApi['schema']['type']);
    }

    public function test_api_hidden_attribute(): void
    {
        $hidden = new ApiHidden('Internal only');

        $this->assertEquals('Internal only', $hidden->reason);
    }
}
