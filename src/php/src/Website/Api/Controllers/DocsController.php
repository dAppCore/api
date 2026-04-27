<?php

declare(strict_types=1);

namespace Core\Website\Api\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Response;
use Illuminate\View\View;
use Core\Website\Api\Services\OpenApiGenerator;
use Symfony\Component\Yaml\Yaml;

class DocsController
{
    public function index(): View
    {
        return view('api::index');
    }

    /**
     * RFC-compatible interactive docs entrypoint.
     */
    public function docs(): View
    {
        return $this->swagger();
    }

    public function guides(): View
    {
        return view('api::guides.index');
    }

    public function quickstart(): View
    {
        return view('api::guides.quickstart');
    }

    public function authentication(): View
    {
        return view('api::guides.authentication');
    }

    public function qrcodes(): View
    {
        return view('api::guides.qrcodes');
    }

    public function webhooks(): View
    {
        return view('api::guides.webhooks');
    }

    public function rateLimits(): View
    {
        return view('api::guides.rate-limits');
    }

    public function errors(): View
    {
        return view('api::guides.errors');
    }

    public function changelog(): View
    {
        return view('api::changelog');
    }

    public function reference(): View
    {
        return view('api::reference');
    }

    public function api(): View
    {
        return view('api::docs');
    }

    public function swagger(): View
    {
        return view('api::swagger');
    }

    public function scalar(): View
    {
        return view('api::scalar');
    }

    public function redoc(): View
    {
        return view('api::redoc');
    }

    public function stoplight(): View
    {
        return view('api::stoplight');
    }

    public function openapi(OpenApiGenerator $generator): JsonResponse
    {
        return response()->json($generator->generate());
    }

    public function openapiYaml(OpenApiGenerator $generator): Response
    {
        return response(
            Yaml::dump($generator->generate(), 20, 2, Yaml::DUMP_MULTI_LINE_LITERAL_BLOCK),
            200,
            ['Content-Type' => 'application/x-yaml; charset=utf-8']
        );
    }

    public function sdks(): View
    {
        return view('api::sdks', [
            'language' => null,
        ]);
    }

    public function sdkDownload(string $language): View
    {
        return view('api::sdks', [
            'language' => $language,
        ]);
    }
}
