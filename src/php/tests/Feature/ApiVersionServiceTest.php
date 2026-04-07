<?php

declare(strict_types=1);

use Core\Front\Api\ApiVersionService;
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
