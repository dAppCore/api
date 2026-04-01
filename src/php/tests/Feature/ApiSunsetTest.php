<?php

declare(strict_types=1);

use Core\Front\Api\Middleware\ApiSunset;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

it('adds deprecation headers without a sunset date', function () {
    $middleware = new ApiSunset();
    $request = Request::create('/legacy-endpoint', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->has('Link'))->toBeFalse();
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated.');
});

it('adds a replacement link without a sunset date', function () {
    $middleware = new ApiSunset();
    $request = Request::create('/old-endpoint', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), '', '/api/v4/users');

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->get('Link'))->toBe('</api/v4/users>; rel="successor-version"');
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated.');
});
