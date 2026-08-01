<?php

declare(strict_types=1);

namespace Core\Api\Documentation;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Response;
use Illuminate\View\View;
use Symfony\Component\Yaml\Yaml;

/**
 * API Documentation Controller.
 *
 * Serves OpenAPI documentation in multiple formats and provides
 * interactive documentation UIs (Swagger, Scalar, ReDoc).
 */
class DocumentationController
{
    public function __construct(
        protected OpenApiBuilder $builder,
    ) {}

    /**
     * Show the main documentation page.
     *
     * Redirects to the configured default UI.
     */
    public function index(): View
    {
        $defaultUi = config('api-docs.ui.default', 'swagger');

        return match ($defaultUi) {
            'swagger' => $this->swagger(),
            'redoc' => $this->redoc(),
            default => $this->scalar(),
        };
    }

    /**
     * Show Swagger UI.
     */
    public function swagger(): View
    {
        $config = config('api-docs.ui.swagger', []);

        return view('api-docs::swagger', [
            'specUrl' => route('api.docs.openapi.json'),
            'config' => $config,
        ]);
    }

    /**
     * Show Scalar API Reference.
     */
    public function scalar(): View
    {
        $config = config('api-docs.ui.scalar', []);

        return view('api-docs::scalar', [
            'specUrl' => route('api.docs.openapi.json'),
            'config' => $config,
        ]);
    }

    /**
     * Show ReDoc documentation.
     */
    public function redoc(): View
    {
        return view('api-docs::redoc', [
            'specUrl' => route('api.docs.openapi.json'),
        ]);
    }

    /**
     * Get OpenAPI specification as JSON.
     */
    public function openApiJson(): JsonResponse
    {
        $spec = $this->builder->build();

        return response()->json($spec)
            ->header('Cache-Control', $this->getCacheControl());
    }

    /**
     * Get OpenAPI specification as YAML.
     */
    public function openApiYaml(): Response
    {
        $spec = $this->builder->build();

        // Convert to YAML
        $yaml = Yaml::dump($spec, 10, 2, Yaml::DUMP_MULTI_LINE_LITERAL_BLOCK);

        return response($yaml)
            ->header('Content-Type', 'application/x-yaml')
            ->header('Cache-Control', $this->getCacheControl());
    }

    /**
     * Clear the documentation cache.
     */
    public function clearCache(): JsonResponse
    {
        $this->builder->clearCache();

        return response()->json([
            'message' => 'Documentation cache cleared successfully.',
        ]);
    }

    /**
     * Get cache control header value.
     */
    protected function getCacheControl(): string
    {
        if (app()->environment('local', 'testing')) {
            return 'no-cache, no-store, must-revalidate';
        }

        $ttl = config('api-docs.cache.ttl', 3600);

        return "public, max-age={$ttl}";
    }
}
