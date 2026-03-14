<?php

declare(strict_types=1);

namespace Core\Api\Documentation;

use Core\Api\Documentation\Middleware\ProtectDocumentation;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\ServiceProvider;

/**
 * API Documentation Service Provider.
 *
 * Registers documentation routes, views, configuration, and services.
 */
class DocumentationServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        // Merge configuration
        $this->mergeConfigFrom(
            __DIR__.'/config.php',
            'api-docs'
        );

        // Register OpenApiBuilder as singleton
        $this->app->singleton(OpenApiBuilder::class, function ($app) {
            return new OpenApiBuilder;
        });
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        // Skip route registration during console commands (except route:list)
        if ($this->shouldRegisterRoutes()) {
            $this->registerRoutes();
        }

        // Register views
        $this->loadViewsFrom(__DIR__.'/Views', 'api-docs');

        // Publish configuration
        if ($this->app->runningInConsole()) {
            $this->publishes([
                __DIR__.'/config.php' => config_path('api-docs.php'),
            ], 'api-docs-config');

            $this->publishes([
                __DIR__.'/Views' => resource_path('views/vendor/api-docs'),
            ], 'api-docs-views');
        }
    }

    /**
     * Check if routes should be registered.
     */
    protected function shouldRegisterRoutes(): bool
    {
        // Always register if not in console
        if (! $this->app->runningInConsole()) {
            return true;
        }

        // Register for artisan route:list command
        $command = $_SERVER['argv'][1] ?? null;

        return $command === 'route:list' || $command === 'route:cache';
    }

    /**
     * Register documentation routes.
     */
    protected function registerRoutes(): void
    {
        $path = config('api-docs.path', '/api/docs');

        Route::middleware(['web', ProtectDocumentation::class])
            ->prefix($path)
            ->group(__DIR__.'/Routes/docs.php');
    }
}
