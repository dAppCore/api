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
        $middleware = ['web', ProtectDocumentation::class];

        Route::middleware($middleware)
            ->prefix($path)
            ->group(__DIR__.'/Routes/docs.php');

        // RFC compatibility alias: expose the same documentation surface at
        // /docs/api for the public website route map while keeping the
        // canonical /api/docs route and its names intact.
        Route::middleware($middleware)
            ->prefix('/docs/api')
            ->group(function (): void {
                Route::get('/', [DocumentationController::class, 'swagger']);
                Route::get('/swagger', [DocumentationController::class, 'swagger']);
                Route::get('/scalar', [DocumentationController::class, 'scalar']);
                Route::get('/redoc', [DocumentationController::class, 'redoc']);
                Route::get('/stoplight', [DocumentationController::class, 'stoplight']);

                Route::get('/openapi.json', [DocumentationController::class, 'openApiJson'])
                    ->middleware('throttle:60,1');

                Route::get('/openapi.yaml', [DocumentationController::class, 'openApiYaml'])
                    ->middleware('throttle:60,1');
            });

        Route::middleware($middleware)
            ->get('/api/reference', [DocumentationController::class, 'redoc'])
            ->name('api.docs.reference.compat');
    }
}
