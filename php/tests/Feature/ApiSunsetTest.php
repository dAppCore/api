<?php

declare(strict_types=1);

use Core\Front\Api\Middleware\ApiSunset;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Config;
use Symfony\Component\HttpFoundation\Response;

define('LEGACY_ENDPOINT', '/legacy-endpoint');
define('SUNSET_DATE', '2025-06-01');
define('SUNSET_LINK_REL', '</api/v2/users>; rel="successor-version"');
define('API_V2_USERS', '/api/v2/users');

it('adds deprecation headers without a sunset date', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'));

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->has('Link'))->toBeFalse();
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated.');
});

it('adds a replacement link without a sunset date', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create('/old-endpoint', 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), '', '/api/v4/users');

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->get('Link'))->toBe('</api/v4/users>; rel="successor-version"');
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated.');
});

it('ApiSunset_successorLinkTarget_Good_strips_a_method_prefix_from_the_replacement_target', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), SUNSET_DATE, 'POST /api/v2/users');

    expect($response->headers->get('Link'))->toBe(SUNSET_LINK_REL);
    expect($response->headers->get('API-Suggested-Replacement'))->toBe('POST /api/v2/users');
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated and will be removed on 2025-06-01.');
});

it('ApiSunset_successorLinkTarget_Bad_keeps_plain_replacement_paths_unchanged', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), SUNSET_DATE, API_V2_USERS);

    expect($response->headers->get('Link'))->toBe(SUNSET_LINK_REL);
    expect($response->headers->get('API-Suggested-Replacement'))->toBe(API_V2_USERS);
});

it('ApiSunset_successorLinkTarget_Ugly_preserves_unrecognised_prefixes_verbatim', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), SUNSET_DATE, 'FETCH /api/v2/users');

    expect($response->headers->get('Link'))->toBe('<FETCH /api/v2/users>; rel="successor-version"');
    expect($response->headers->get('API-Suggested-Replacement'))->toBe('FETCH /api/v2/users');
});

it('preserves existing deprecation headers while appending sunset metadata', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, function () {
        $response = new Response('OK');
        $response->headers->set('Deprecation', 'false');
        $response->headers->set('Sunset', 'Wed, 01 Jan 2025 00:00:00 GMT');
        $response->headers->set('Link', '<https://example.com/docs>; rel="help"');
        $response->headers->set('X-API-Warn', 'Existing warning');

        return $response;
    }, SUNSET_DATE, API_V2_USERS);

    expect($response->headers->all('Deprecation'))->toHaveCount(2);
    expect($response->headers->all('Sunset'))->toHaveCount(2);
    expect($response->headers->all('Link'))->toHaveCount(2);
    expect($response->headers->all('X-API-Warn'))->toHaveCount(2);
    expect($response->headers->all('Link'))->toContain('<https://example.com/docs>; rel="help"');
    expect($response->headers->all('Link'))->toContain(SUNSET_LINK_REL);
});

it('formats the sunset date and keeps the replacement link', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), SUNSET_DATE, API_V2_USERS);

    expect($response->headers->get('Deprecation'))->toBe('true');
    expect($response->headers->get('Sunset'))->toBe('Sun, 01 Jun 2025 00:00:00 GMT');
    expect($response->headers->get('Link'))->toBe(SUNSET_LINK_REL);
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated and will be removed on 2025-06-01.');
});

it('adds a deprecation notice url when provided', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle(
        $request,
        fn () => new Response('OK'),
        SUNSET_DATE,
        API_V2_USERS,
        'https://docs.example.com/deprecation/users'
    );

    expect($response->headers->get('API-Deprecation-Notice-URL'))->toBe('https://docs.example.com/deprecation/users');
    expect($response->headers->get('API-Suggested-Replacement'))->toBe(API_V2_USERS);
});

it('preserves already formatted sunset dates', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');
    $sunset = 'Wed, 01 Jan 2025 00:00:00 GMT';

    $response = $middleware->handle($request, fn () => new Response('OK'), $sunset, API_V2_USERS);

    expect($response->headers->get('Sunset'))->toBe($sunset);
    expect($response->headers->get('X-API-Warn'))->toBe("This endpoint is deprecated and will be removed on {$sunset}.");
});

it('ApiSunset_formatSunsetDate_Ugly_preserves_invalid_sunset_values', function () {
    Config::set('api.headers.include_deprecation', true);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), 'not-a-date', API_V2_USERS);

    expect($response->headers->get('Sunset'))->toBe('not-a-date');
    expect($response->headers->get('X-API-Warn'))->toBe('This endpoint is deprecated and will be removed on not-a-date.');
});

it('skips deprecation headers when they are disabled in configuration', function () {
    Config::set('api.headers.include_deprecation', false);

    $middleware = new ApiSunset();
    $request = Request::create(LEGACY_ENDPOINT, 'GET');

    $response = $middleware->handle($request, fn () => new Response('OK'), SUNSET_DATE, API_V2_USERS);

    expect($response->headers->has('Deprecation'))->toBeFalse();
    expect($response->headers->has('Sunset'))->toBeFalse();
    expect($response->headers->has('Link'))->toBeFalse();
    expect($response->headers->has('X-API-Warn'))->toBeFalse();
});
