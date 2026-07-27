<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;

define('PIXEL_ENDPOINT', '/api/pixel/abc12345');
define('EXAMPLE_COM', 'https://example.com');

beforeEach(function () {
    Cache::flush();
});

afterEach(function () {
    Cache::flush();
});

it('returns a transparent gif for get requests', function () {
    $response = $this->get(PIXEL_ENDPOINT, [
        'Origin' => EXAMPLE_COM,
    ]);

    $response->assertOk();
    $response->assertHeader('Content-Type', 'image/gif');
    $response->assertHeader('Access-Control-Allow-Origin', EXAMPLE_COM);
    $response->assertHeader('X-RateLimit-Limit', '10000');
    $response->assertHeader('X-RateLimit-Remaining', '9999');

    expect($response->getContent())->toBe(base64_decode('R0lGODlhAQABAPAAAP///wAAACH5BAAAAAAALAAAAAABAAEAAAICRAEAOw=='));
});

it('accepts post tracking requests without a body', function () {
    $response = $this->post(PIXEL_ENDPOINT, [], [
        'Origin' => EXAMPLE_COM,
    ]);

    $response->assertNoContent();
    $response->assertHeader('Access-Control-Allow-Origin', EXAMPLE_COM);
    $response->assertHeader('X-RateLimit-Limit', '10000');
    $response->assertHeader('X-RateLimit-Remaining', '9999');
});

it('handles preflight requests for public pixel tracking', function () {
    $response = $this->call('OPTIONS', PIXEL_ENDPOINT, [], [], [], [
        'HTTP_ORIGIN' => EXAMPLE_COM,
    ]);

    $response->assertNoContent();
    $response->assertHeader('Access-Control-Allow-Origin', EXAMPLE_COM);
    $response->assertHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
});
