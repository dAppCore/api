<?php

declare(strict_types=1);

use Core\Front\Api\ApiVersionService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Config;

beforeEach(function () {
    Config::set('api.versioning.default', 1);
    Config::set('api.versioning.current', 2);
    Config::set('api.versioning.supported', [1, 2]);
    Config::set('api.versioning.deprecated', []);
    Config::set('api.versioning.sunset', []);
});

it('normalises configured versions before reading them', function () {
    Config::set('api.versioning.supported', ['1', '02', '2', 'ignored', 0, -1]);
    Config::set('api.versioning.deprecated', ['1', '3', '3']);
    Config::set('api.versioning.sunset', [
        '1' => '2025-06-01',
        '02' => '2025-12-31',
        'ignored' => '2026-01-01',
        0 => '2024-01-01',
        -1 => '2024-06-01',
        3 => '',
    ]);

    $versions = new ApiVersionService();

    expect($versions->supportedVersions())->toBe([1, 2]);
    expect($versions->deprecatedVersions())->toBe([1, 3]);
    expect($versions->sunsetDates())->toBe([
        1 => '2025-06-01',
        2 => '2025-12-31',
    ]);
    expect($versions->isSupported(1))->toBeTrue();
    expect($versions->isSupported(2))->toBeTrue();
    expect($versions->isSupported(3))->toBeFalse();
    expect($versions->isDeprecated())->toBeFalse();
});

it('ApiVersionService_current_Good_reads_request_version_attributes', function () {
    Config::set('api.versioning.deprecated', [2]);

    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');
    $request->attributes->set('api_version', 2);
    $request->attributes->set('api_version_string', 'v2');

    expect($versions->current($request))->toBe(2);
    expect($versions->currentString($request))->toBe('v2');
    expect($versions->is(2, $request))->toBeTrue();
    expect($versions->isV1($request))->toBeFalse();
    expect($versions->isV2($request))->toBeTrue();
    expect($versions->isAtLeast(1, $request))->toBeTrue();
    expect($versions->isAtLeast(3, $request))->toBeFalse();
    expect($versions->isDeprecated($request))->toBeTrue();
});

it('ApiVersionService_defaults_Good_uses_configured_versions_when_request_has_none', function () {
    Config::set('api.versioning.default', 1);
    Config::set('api.versioning.current', 3);

    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');

    expect($versions->current($request))->toBeNull();
    expect($versions->defaultVersion())->toBe(1);
    expect($versions->latestVersion())->toBe(3);
    expect($versions->isAtLeast(1, $request))->toBeFalse();
    expect($versions->isDeprecated($request))->toBeFalse();
});

it('ApiVersionService_negotiate_Good_picks_the_best_available_handler', function () {
    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');

    $request->attributes->set('api_version', 3);

    $exact = $versions->negotiate($request, [
        1 => fn () => 'v1',
        2 => fn () => 'v2',
        3 => fn () => 'v3',
    ]);

    expect($exact)->toBe('v3');

    $request->attributes->set('api_version', 4);
    $fallback = $versions->negotiate($request, [
        1 => fn () => 'v1',
        2 => fn () => 'v2',
    ]);

    expect($fallback)->toBe('v2');
});

it('ApiVersionService_negotiate_Bad_throws_when_no_handler_matches', function () {
    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');
    $request->attributes->set('api_version', 1);

    expect(fn () => $versions->negotiate($request, [
        2 => fn () => 'v2',
    ]))->toThrow(InvalidArgumentException::class);
});

it('ApiVersionService_transform_Good_applies_exact_or_fallback_transformers', function () {
    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');
    $payload = ['name' => 'Ada', 'legacy' => true];

    $request->attributes->set('api_version', 2);

    $exact = $versions->transform($request, $payload, [
        1 => fn (array $data) => ['name' => $data['name']],
        2 => fn (array $data) => ['name' => $data['name'], 'version' => 2],
    ]);

    expect($exact)->toBe(['name' => 'Ada', 'version' => 2]);

    $request->attributes->set('api_version', 3);
    $fallback = $versions->transform($request, $payload, [
        1 => fn (array $data) => ['name' => $data['name']],
        2 => fn (array $data) => ['name' => $data['name'], 'version' => 2],
    ]);

    expect($fallback)->toBe(['name' => 'Ada', 'version' => 2]);
});

it('ApiVersionService_transform_Ugly_returns_original_data_without_a_matching_transformer', function () {
    $versions = new ApiVersionService();
    $request = Request::create('/api/users', 'GET');
    $request->attributes->set('api_version', 1);
    $payload = ['name' => 'Ada', 'legacy' => true];

    $result = $versions->transform($request, $payload, [
        2 => fn (array $data) => ['name' => $data['name'], 'version' => 2],
    ]);

    expect($result)->toBe($payload);
});
