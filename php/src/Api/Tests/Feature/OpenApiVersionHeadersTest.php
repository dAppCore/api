<?php

declare(strict_types=1);

use Core\Api\Documentation\OpenApiBuilder;
use Illuminate\Support\Facades\Config;
use Illuminate\Support\Facades\Route as RouteFacade;

beforeEach(function () {
    Config::set('api.headers.include_version', true);
    Config::set('api.headers.include_deprecation', true);
    Config::set('api.versioning.deprecated', [1]);
    Config::set('api.versioning.sunset', [
        1 => '2025-06-01',
    ]);
    Config::set('api-docs.routes.include', ['api/*']);
    Config::set('api-docs.routes.exclude', []);
});

it('documents version headers and version-driven deprecation on versioned routes', function () {
    RouteFacade::prefix('api/v1')
        ->middleware(['api', 'api.version:1'])
        ->group(function () {
            RouteFacade::get('/legacy-status', fn () => response()->json(['ok' => true]));
        });

    $spec = (new OpenApiBuilder)->build();

    expect($spec['components']['headers']['xapiversion'] ?? null)->not->toBeNull();

    $operation = $spec['paths']['/api/v1/legacy-status']['get'];

    expect($operation['deprecated'] ?? null)->toBeTrue();

    foreach (['200', '400', '500'] as $status) {
        $headers = $operation['responses'][$status]['headers'] ?? [];

        expect($headers)->toHaveKey('X-API-Version');
        expect($headers)->toHaveKey('Deprecation');
        expect($headers)->toHaveKey('Sunset');
        expect($headers)->toHaveKey('X-API-Warn');
    }
});
