<?php

declare(strict_types=1);

use Core\Front\Api\Middleware\ApiVersion;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Config;
use Symfony\Component\HttpFoundation\Response;

beforeEach(function () {
    Config::set('api.versioning.default', 1);
    Config::set('api.versioning.current', 2);
    Config::set('api.versioning.supported', [1, 2]);
    Config::set('api.versioning.deprecated', [1]);
    Config::set('api.versioning.sunset', [
        1 => '2025-06-01',
    ]);
    Config::set('api.headers.include_version', true);
    Config::set('api.headers.include_deprecation', true);
});

it('skips the api version header when it is disabled in configuration', function () {
    Config::set('api.headers.include_version', false);

    $middleware = new ApiVersion();
    $request = Request::create('/api/users', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->has('X-API-Version'))->toBeFalse();
});

it('skips deprecation headers when they are disabled in configuration', function () {
    Config::set('api.headers.include_deprecation', false);

    $middleware = new ApiVersion();
    $request = Request::create('/api/v1/users', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('X-API-Version'))->toBe('1');
    expect($response->headers->has('Deprecation'))->toBeFalse();
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->has('X-API-Warn'))->toBeFalse();
});
