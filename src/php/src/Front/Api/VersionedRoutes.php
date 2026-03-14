<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Front\Api;

use Illuminate\Support\Facades\Route;

/**
 * Versioned Routes Helper.
 *
 * Provides fluent helpers for registering version-based API routes.
 *
 * ## Basic Usage
 *
 * Register routes for a specific version:
 *
 * ```php
 * use Core\Front\Api\VersionedRoutes;
 *
 * VersionedRoutes::v1(function () {
 *     Route::get('/users', [UserController::class, 'index']);
 * });
 *
 * VersionedRoutes::v2(function () {
 *     Route::get('/users', [UserControllerV2::class, 'index']);
 * });
 * ```
 *
 * ## URL Prefixed Versions
 *
 * By default, routes are prefixed with the version (e.g., /api/v1/users):
 *
 * ```php
 * VersionedRoutes::v1(function () {
 *     Route::get('/users', ...); // Accessible at /api/v1/users
 * });
 * ```
 *
 * ## Header-Only Versions
 *
 * For routes that use header-based versioning only:
 *
 * ```php
 * VersionedRoutes::version(1)
 *     ->withoutPrefix()
 *     ->routes(function () {
 *         Route::get('/users', ...); // Accessible at /api/users with Accept-Version: 1
 *     });
 * ```
 *
 * ## Multiple Versions
 *
 * Register the same routes for multiple versions:
 *
 * ```php
 * VersionedRoutes::versions([1, 2], function () {
 *     Route::get('/status', [StatusController::class, 'index']);
 * });
 * ```
 *
 * ## Deprecation
 *
 * Mark a version as deprecated with custom sunset date:
 *
 * ```php
 * VersionedRoutes::v1()
 *     ->deprecated('2025-06-01')
 *     ->routes(function () {
 *         Route::get('/legacy', ...);
 *     });
 * ```
 */
class VersionedRoutes
{
    protected int $version;

    protected bool $usePrefix = true;

    protected ?string $sunsetDate = null;

    protected bool $isDeprecated = false;

    /**
     * @var array<string>
     */
    protected array $middleware = [];

    /**
     * Create a new versioned routes instance.
     */
    public function __construct(int $version)
    {
        $this->version = $version;
    }

    /**
     * Create routes for version 1.
     */
    public static function v1(?callable $routes = null): static
    {
        $instance = new static(1);

        if ($routes !== null) {
            $instance->routes($routes);
        }

        return $instance;
    }

    /**
     * Create routes for version 2.
     */
    public static function v2(?callable $routes = null): static
    {
        $instance = new static(2);

        if ($routes !== null) {
            $instance->routes($routes);
        }

        return $instance;
    }

    /**
     * Create routes for a specific version.
     */
    public static function version(int $version): static
    {
        return new static($version);
    }

    /**
     * Register routes for multiple versions.
     *
     * @param  array<int>  $versions
     */
    public static function versions(array $versions, callable $routes): void
    {
        foreach ($versions as $version) {
            (new static($version))->routes($routes);
        }
    }

    /**
     * Don't use URL prefix for this version.
     *
     * Routes will be accessible without /v{n} prefix but will
     * still require version header for version-specific behaviour.
     */
    public function withoutPrefix(): static
    {
        $this->usePrefix = false;

        return $this;
    }

    /**
     * Use URL prefix for this version.
     *
     * This is the default behaviour.
     */
    public function withPrefix(): static
    {
        $this->usePrefix = true;

        return $this;
    }

    /**
     * Mark this version as deprecated.
     *
     * @param  string|null  $sunsetDate  Optional sunset date (YYYY-MM-DD or RFC7231 format)
     */
    public function deprecated(?string $sunsetDate = null): static
    {
        $this->isDeprecated = true;
        $this->sunsetDate = $sunsetDate;

        return $this;
    }

    /**
     * Add additional middleware to the version routes.
     *
     * @param  array<string>|string  $middleware
     */
    public function middleware(array|string $middleware): static
    {
        $this->middleware = array_merge(
            $this->middleware,
            is_array($middleware) ? $middleware : [$middleware]
        );

        return $this;
    }

    /**
     * Register the routes for this version.
     */
    public function routes(callable $routes): void
    {
        $attributes = $this->buildRouteAttributes();

        Route::group($attributes, $routes);
    }

    /**
     * Build the route group attributes.
     *
     * @return array<string, mixed>
     */
    protected function buildRouteAttributes(): array
    {
        $attributes = [
            'middleware' => $this->buildMiddleware(),
        ];

        if ($this->usePrefix) {
            $attributes['prefix'] = "v{$this->version}";
        }

        return $attributes;
    }

    /**
     * Build the middleware stack for this version.
     *
     * @return array<string>
     */
    protected function buildMiddleware(): array
    {
        $middleware = ["api.version:{$this->version}"];

        if ($this->isDeprecated && $this->sunsetDate) {
            $middleware[] = "api.sunset:{$this->sunsetDate}";
        }

        return array_merge($middleware, $this->middleware);
    }
}
