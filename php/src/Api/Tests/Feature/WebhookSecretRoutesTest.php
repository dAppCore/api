<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;

it('registers webhook secret management routes', function () {
    $socialRotate = Route::getRoutes()->getByName('api.webhooks.social.rotate-secret');
    $socialStatus = Route::getRoutes()->getByName('api.webhooks.social.status');
    $contentRotate = Route::getRoutes()->getByName('api.webhooks.content.rotate-secret');
    $contentStatus = Route::getRoutes()->getByName('api.webhooks.content.status');

    expect($socialRotate)->not->toBeNull();
    expect($socialRotate->uri())->toBe('api/webhooks/social/{uuid}/secret/rotate');
    expect($socialRotate->methods())->toContain('POST');

    expect($socialStatus)->not->toBeNull();
    expect($socialStatus->uri())->toBe('api/webhooks/social/{uuid}/secret');
    expect($socialStatus->methods())->toContain('GET');

    expect($contentRotate)->not->toBeNull();
    expect($contentRotate->uri())->toBe('api/webhooks/content/{uuid}/secret/rotate');
    expect($contentRotate->methods())->toContain('POST');

    expect($contentStatus)->not->toBeNull();
    expect($contentStatus->uri())->toBe('api/webhooks/content/{uuid}/secret');
    expect($contentStatus->methods())->toContain('GET');
});
