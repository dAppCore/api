<?php

declare(strict_types=1);

use Core\Api\Documentation\DocumentationController;
use Core\Api\Documentation\OpenApiBuilder;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

beforeEach(function () {
    if (Route::getRoutes()->getByName('api.docs.openapi.json') === null) {
        Route::get('/openapi.json', fn () => response()->json([]))
            ->name('api.docs.openapi.json');
    }
});

it('renders Stoplight Elements when selected as the default documentation ui', function () {
    $controller = new DocumentationController(new class extends OpenApiBuilder {});
    config(['api-docs.ui.default' => 'stoplight']);

    $response = $controller->index(Request::create('/api/docs', 'GET'));
    $html = $response->render();

    expect($html)->toContain('elements-api');
    expect($html)->toContain('@stoplight/elements');
});

it('renders the dedicated Stoplight documentation route', function () {
    $controller = new DocumentationController(new class extends OpenApiBuilder {});
    $response = $controller->stoplight(Request::create('/api/docs/stoplight', 'GET'));
    $html = $response->render();

    expect($html)->toContain('elements-api');
    expect($html)->toContain('@stoplight/elements');
});
