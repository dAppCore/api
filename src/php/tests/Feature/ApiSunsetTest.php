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

it('formats the sunset date and keeps the replacement link', function () {
    $middleware = new ApiSunset();
    $request = Request::create('/legacy-endpoint', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), '2025-06-01', '/api/v2/users');

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->get('Sunset'))->toBe('Sun, 01 Jun 2025 00:00:00 GMT');
    expect($response->headers->get('Link'))->toBe('</api/v2/users>; rel="successor-version"');
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated and will be removed on 2025-06-01.');
});
