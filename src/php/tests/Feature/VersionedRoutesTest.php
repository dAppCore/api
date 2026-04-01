<?php

declare(strict_types=1);

use Core\Front\Api\VersionedRoutes;

it('passes a replacement url through deprecated versioned routes', function () {
    $routes = new class (2) extends VersionedRoutes {
        public function attributes(): array
        {
            return $this->buildRouteAttributes();
        }
    };

    $attributes = $routes->deprecated('2025-06-01', '/api/v3/users')->attributes();

    expect($attributes)->toHaveKey('middleware');
    expect($attributes['middleware'])->toContain('api.version:2');
    expect($attributes['middleware'])->toContain('api.sunset:2025-06-01,/api/v3/users');
});

it('preserves the existing deprecated signature without a replacement url', function () {
    $routes = new class (1) extends VersionedRoutes {
        public function attributes(): array
        {
            return $this->buildRouteAttributes();
        }
    };

    $attributes = $routes->deprecated('2025-06-01')->attributes();

    expect($attributes['middleware'])->toContain('api.sunset:2025-06-01');
    expect($attributes['middleware'])->not->toContain('api.sunset:2025-06-01,/api/v3/users');
});

it('keeps deprecated routes active without a sunset date', function () {
    $routes = new class (3) extends VersionedRoutes {
        public function attributes(): array
        {
            return $this->buildRouteAttributes();
        }
    };

    $attributes = $routes->deprecated()->attributes();

    expect($attributes['middleware'])->toContain('api.version:3');
    expect($attributes['middleware'])->toContain('api.sunset');
});

it('passes a replacement url through deprecated versioned routes without a sunset date', function () {
    $routes = new class (4) extends VersionedRoutes {
        public function attributes(): array
        {
            return $this->buildRouteAttributes();
        }
    };

    $attributes = $routes->deprecated(null, '/api/v4/users')->attributes();

    expect($attributes['middleware'])->toContain('api.version:4');
    expect($attributes['middleware'])->toContain('api.sunset:,/api/v4/users');
});
