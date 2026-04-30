<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Front\Api;

use Core\Front\Api\Middleware\ApiSunset;
use Core\Front\Api\Middleware\ApiVersion;
use Core\LifecycleEventProvider;
use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Foundation\Configuration\Middleware;
use Illuminate\Http\Request;
use Illuminate\Routing\Router;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\ServiceProvider;

/**
 * API frontage - API stage.
 *
 * Provides api middleware group for API routes and API versioning support.
 *
 * ## API Versioning
 *
 * This provider registers middleware for API versioning:
 * - `api.version` - Parses and validates API version from URL or headers
 * - `api.sunset` - Adds deprecation/sunset headers to endpoints
 *
 * Configure versioning in config/api.php:
 * ```php
 * 'versioning' => [
 *     'default' => 1,           // Default version when none specified
 *     'current' => 1,           // Current/latest version
 *     'supported' => [1],       // List of supported versions
 *     'deprecated' => [],       // Deprecated but still supported versions
 *     'sunset' => [],           // Sunset dates: [1 => '2025-06-01']
 * ],
 * ```
 *
 * @see ApiVersion Middleware for version parsing
 * @see ApiVersionService Service for programmatic version checks
 * @see VersionedRoutes Helper for version-based route registration
 */
class Boot extends ServiceProvider
{
    /**
     * Configure api middleware group.
     */
    public static function middleware(Middleware $middleware): void
    {
        $middleware->group('api', [
            \Illuminate\Routing\Middleware\ThrottleRequests::class.':api',
            \Illuminate\Routing\Middleware\SubstituteBindings::class,
        ]);

        // Register versioning middleware aliases
        $middleware->alias([
            'api.version' => ApiVersion::class,
            'api.sunset' => ApiSunset::class,
        ]);
    }

    public function register(): void
    {
        // Merge API configuration
        $this->mergeConfigFrom(__DIR__.'/config.php', 'api');

        // Register API version service as singleton
        $this->app->singleton(ApiVersionService::class);
    }

    public function boot(): void
    {
        $this->configureRateLimiting();
        $this->registerMiddlewareAliases();

        // Fire ApiRoutesRegistering event for lazy-loaded modules
        LifecycleEventProvider::fireApiRoutes();
    }

    /**
     * Register middleware aliases via router.
     *
     * This ensures aliases are available even if the static middleware()
     * method isn't called (e.g., in testing or custom bootstrap).
     */
    protected function registerMiddlewareAliases(): void
    {
        /** @var Router $router */
        $router = $this->app->make(Router::class);

        $router->aliasMiddleware('api.version', ApiVersion::class);
        $router->aliasMiddleware('api.sunset', ApiSunset::class);
    }

    /**
     * Configure API rate limiting.
     */
    protected function configureRateLimiting(): void
    {
        RateLimiter::for('api', function (Request $request) {
            return Limit::perMinute(60)->by($request->user()?->id ?: $request->ip());
        });
    }
}
