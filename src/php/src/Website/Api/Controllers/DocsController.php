<?php

declare(strict_types=1);

namespace Core\Website\Api\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\View\View;
use Core\Website\Api\Services\OpenApiGenerator;

class DocsController
{
    public function index(): View
    {
        return view('api::index');
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

    public function errors(): View
    {
        return view('api::guides.errors');
    }

    public function reference(): View
    {
        return view('api::reference');
    }

    public function api(): View
    {
        return $this->reference();
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
}
