<?php

declare(strict_types=1);

use Core\Api\Models\ApiKey;
use Core\Website\Api\Controllers\DocsController;

function renderAuthenticationGuide(): string
{
    return (new DocsController)->authentication()->render();
}

it('AuthenticationGuide_renderedPrefix_Good_uses_the_configured_api_key_prefix', function () {
    $originalPrefix = config('api.keys.prefix');

    try {
        config(['api.keys.prefix' => 'acme']);

        $html = renderAuthenticationGuide();

        expect($html)->toContain('API keys are prefixed with');
        expect($html)->toContain(ApiKey::keyPrefixRoot());
        expect($html)->toContain('Authorization: Bearer acme_your_api_key_here');
        expect($html)->not->toContain('hk_');
    } finally {
        if ($originalPrefix === null) {
            config()->offsetUnset('api.keys.prefix');
        } else {
            config(['api.keys.prefix' => $originalPrefix]);
        }
    }
});

it('AuthenticationGuide_renderedPrefix_Bad_falls_back_to_the_default_prefix_when_blank', function () {
    $originalPrefix = config('api.keys.prefix');

    try {
        config(['api.keys.prefix' => '   ']);

        $html = renderAuthenticationGuide();

        expect(ApiKey::keyPrefixRoot())->toBe('hk_');
        expect($html)->toContain('API keys are prefixed with');
        expect($html)->toContain('hk_');
        expect($html)->toContain('Authorization: Bearer hk_your_api_key_here');
    } finally {
        if ($originalPrefix === null) {
            config()->offsetUnset('api.keys.prefix');
        } else {
            config(['api.keys.prefix' => $originalPrefix]);
        }
    }
});

it('AuthenticationGuide_renderedPrefix_Ugly_trims_whitespace_before_rendering', function () {
    $originalPrefix = config('api.keys.prefix');

    try {
        config(['api.keys.prefix' => "  acme  "]);

        $html = renderAuthenticationGuide();

        expect(ApiKey::keyPrefixRoot())->toBe('acme_');
        expect($html)->toContain('API keys are prefixed with');
        expect($html)->toContain('acme_');
        expect($html)->toContain('Authorization: Bearer acme_your_api_key_here');
    } finally {
        if ($originalPrefix === null) {
            config()->offsetUnset('api.keys.prefix');
        } else {
            config(['api.keys.prefix' => $originalPrefix]);
        }
    }
});
