<?php

declare(strict_types=1);

use Core\Front\Api\VersionedRoutes;

class InspectableVersionedRoutes extends VersionedRoutes
{
    public function attributes(): array
    {
        return $this->buildRouteAttributes();
    }

    public function middlewareStack(): array
    {
        return $this->buildMiddleware();
    }
}

it('passes a replacement url through deprecated versioned routes', function () {
    $routes = new InspectableVersionedRoutes(2);

    $attributes = $routes->deprecated('2025-06-01', '/api/v3/users')->attributes();

    expect($attributes)->toHaveKey('middleware');
    expect($attributes['middleware'])->toContain('api.version:2');
    expect($attributes['middleware'])->toContain('api.sunset:2025-06-01,/api/v3/users');
});

it('preserves the existing deprecated signature without a replacement url', function () {
    $routes = new InspectableVersionedRoutes(1);

    $attributes = $routes->deprecated('2025-06-01')->attributes();

    expect($attributes['middleware'])->toContain('api.sunset:2025-06-01');
    expect($attributes['middleware'])->not->toContain('api.sunset:2025-06-01,/api/v3/users');
});

it('keeps deprecated routes active without a sunset date', function () {
    $routes = new InspectableVersionedRoutes(3);

    $attributes = $routes->deprecated()->attributes();

    expect($attributes['middleware'])->toContain('api.version:3');
    expect($attributes['middleware'])->toContain('api.sunset');
});

it('passes a replacement url through deprecated versioned routes without a sunset date', function () {
    $routes = new InspectableVersionedRoutes(4);

    $attributes = $routes->deprecated(null, '/api/v4/users')->attributes();

    expect($attributes['middleware'])->toContain('api.version:4');
    expect($attributes['middleware'])->toContain('api.sunset:,/api/v4/users');
});

it('VersionedRoutes_withoutPrefix_Good_omits_the_url_prefix', function () {
    $routes = new InspectableVersionedRoutes(2);

    $attributes = $routes->withoutPrefix()->attributes();

    expect($attributes)->not->toHaveKey('prefix');
    expect($attributes['middleware'])->toContain('api.version:2');
});

it('VersionedRoutes_withPrefix_Good_restores_the_url_prefix', function () {
    $routes = new InspectableVersionedRoutes(2);

    $attributes = $routes->withoutPrefix()->withPrefix()->attributes();

    expect($attributes['prefix'])->toBe('v2');
    expect($attributes['middleware'])->toContain('api.version:2');
});

it('VersionedRoutes_middleware_Good_appends_extra_middleware_after_the_version_guard', function () {
    $routes = new InspectableVersionedRoutes(5);

    $middleware = $routes->middleware(['auth:sanctum', 'throttle:60,1'])->middlewareStack();

    expect($middleware[0])->toBe('api.version:5');
    expect($middleware)->toContain('auth:sanctum');
    expect($middleware)->toContain('throttle:60,1');
});

it('VersionedRoutes_deprecated_Ugly_supports_notice_url_only', function () {
    $routes = new InspectableVersionedRoutes(6);

    $attributes = $routes->deprecated(null, null, 'https://docs.example.com/deprecation/v6')->attributes();

    expect($attributes['middleware'])->toContain('api.version:6');
    expect($attributes['middleware'])->toContain('api.sunset:,,https://docs.example.com/deprecation/v6');
});
