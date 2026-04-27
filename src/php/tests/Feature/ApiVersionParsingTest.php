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
    Config::set('api.versioning.deprecated', []);
    Config::set('api.versioning.sunset', []);
    Config::set('api.headers.include_version', true);
    Config::set('api.headers.include_deprecation', true);
});

it('resolves the api version from an accept-version header with parameters', function () {
    $middleware = new ApiVersion();
    $request = Request::create('/api/users', 'GET');
    $request->headers->set('Accept-Version', 'v2; q=1.0');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('X-API-Version'))->toBe('2');
    expect($request->attributes->get('api_version'))->toBe(2);
    expect($request->attributes->get('api_version_string'))->toBe('v2');
});

it('resolves the api version from a vendor accept header inside a list', function () {
    $middleware = new ApiVersion();
    $request = Request::create('/api/users', 'GET');
    $request->headers->set(
        'Accept',
        'text/html;q=0.8, application/json, application/vnd.hosthub.v2+json; charset=utf-8'
    );

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('X-API-Version'))->toBe('2');
    expect($request->attributes->get('api_version'))->toBe(2);
    expect($request->attributes->get('api_version_string'))->toBe('v2');
});

it('ApiVersionParsingTest_handle_Good_resolves_the_api_version_from_the_url_path_prefix', function () {
    $middleware = new ApiVersion();
    $request = Request::create('/api/v2/users', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('X-API-Version'))->toBe('2');
    expect($request->attributes->get('api_version'))->toBe(2);
    expect($request->attributes->get('api_version_string'))->toBe('v2');
});

it('ApiVersionParsingTest_handle_Bad_rejects_unsupported_path_versions', function () {
    $middleware = new ApiVersion();
    $request = Request::create('/api/v9/users', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->getStatusCode())->toBe(400);
    expect($response->headers->get('X-API-Version'))->toBe('2');
    expect($response->getData(true))->toMatchArray([
        'error' => 'unsupported_api_version',
        'requested_version' => 9,
        'supported_versions' => [1, 2],
        'current_version' => 2,
    ]);
});

it('ApiVersionParsingTest_handle_Ugly_prefers_the_url_path_over_version_headers', function () {
    $middleware = new ApiVersion();
    $request = Request::create('/api/v1/users', 'GET');
    $request->headers->set('Accept-Version', 'v2');
    $request->headers->set('Accept', 'application/vnd.hosthub.v2+json');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('X-API-Version'))->toBe('1');
    expect($request->attributes->get('api_version'))->toBe(1);
    expect($request->attributes->get('api_version_string'))->toBe('v1');
});
