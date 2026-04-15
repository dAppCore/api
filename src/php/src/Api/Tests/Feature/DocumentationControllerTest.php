<?php

declare(strict_types=1);

use Core\Api\Documentation\DocumentationController;
use Core\Api\Documentation\OpenApiBuilder;
use Illuminate\Http\Request;

class StubDocumentationBuilder extends OpenApiBuilder
{
    public bool $cleared = false;

    public function build(): array
    {
        return [
            'openapi' => '3.1.0',
            'info' => [
                'title' => 'Core API',
                'version' => '1.0.0',
            ],
            'servers' => [],
            'tags' => [],
            'paths' => [],
            'components' => [],
        ];
    }

    public function clearCache(): void
    {
        $this->cleared = true;
    }
}

it('DocumentationController_openApiJson_Good_returns_json_with_cache_headers', function () {
    $builder = new StubDocumentationBuilder;
    $controller = new DocumentationController($builder);

    $response = $controller->openApiJson(Request::create('/api/docs/openapi.json', 'GET'));

    expect($response->getStatusCode())->toBe(200);
    expect($response->headers->get('Cache-Control'))->toContain('no-cache');
    expect($response->headers->get('Cache-Control'))->toContain('no-store');
    expect($response->headers->get('Cache-Control'))->toContain('must-revalidate');
    expect($response->getData(true))->toMatchArray([
        'openapi' => '3.1.0',
        'info' => [
            'title' => 'Core API',
            'version' => '1.0.0',
        ],
    ]);
});

it('DocumentationController_openApiYaml_Good_returns_yaml_with_cache_headers', function () {
    $builder = new StubDocumentationBuilder;
    $controller = new DocumentationController($builder);

    $response = $controller->openApiYaml(Request::create('/api/docs/openapi.yaml', 'GET'));

    expect($response->getStatusCode())->toBe(200);
    expect($response->headers->get('Content-Type'))->toBe('application/x-yaml');
    expect($response->headers->get('Cache-Control'))->toContain('no-cache');
    expect($response->headers->get('Cache-Control'))->toContain('no-store');
    expect($response->headers->get('Cache-Control'))->toContain('must-revalidate');
    expect($response->getContent())->toContain('openapi: 3.1.0');
    expect($response->getContent())->toContain("title: 'Core API'");
});

it('DocumentationController_clearCache_Bad_marks_the_builder_as_cleared', function () {
    $builder = new StubDocumentationBuilder;
    $controller = new DocumentationController($builder);

    $response = $controller->clearCache(Request::create('/api/docs/cache/clear', 'POST'));

    expect($builder->cleared)->toBeTrue();
    expect($response->getStatusCode())->toBe(200);
    expect($response->getData(true))->toBe([
        'message' => 'Documentation cache cleared successfully.',
    ]);
});
