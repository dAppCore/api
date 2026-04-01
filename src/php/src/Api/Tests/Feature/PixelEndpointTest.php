<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;

beforeEach(function () {
    Cache::flush();
});

afterEach(function () {
    Cache::flush();
});

it('returns a transparent gif for get requests', function () {
    $response = $this->get('/api/pixel/abc12345', [
        'Origin' => 'https://example.com',
    ]);

    $response->assertOk();
    $response->assertHeader('Content-Type', 'image/gif');
    $response->assertHeader('Access-Control-Allow-Origin', 'https://example.com');
    $response->assertHeader('X-RateLimit-Limit', '10000');
    $response->assertHeader('X-RateLimit-Remaining', '9999');

    expect($response->getContent())->toBe(base64_decode('R0lGODlhAQABAPAAAP///wAAACH5BAAAAAAALAAAAAABAAEAAAICRAEAOw=='));
});

it('accepts post tracking requests without a body', function () {
    $response = $this->post('/api/pixel/abc12345', [], [
        'Origin' => 'https://example.com',
    ]);

    $response->assertNoContent();
    $response->assertHeader('Access-Control-Allow-Origin', 'https://example.com');
    $response->assertHeader('X-RateLimit-Limit', '10000');
    $response->assertHeader('X-RateLimit-Remaining', '9999');
});

it('handles preflight requests for public pixel tracking', function () {
    $response = $this->call('OPTIONS', '/api/pixel/abc12345', [], [], [], [
        'HTTP_ORIGIN' => 'https://example.com',
    ]);

    $response->assertNoContent();
    $response->assertHeader('Access-Control-Allow-Origin', 'https://example.com');
    $response->assertHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
});
