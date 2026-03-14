<?php

declare(strict_types=1);

namespace Core\Api;

use Core\Events\AdminPanelBooting;
use Core\Events\ApiRoutesRegistering;
use Core\Events\ConsoleBooting;
use Core\Api\Documentation\DocumentationServiceProvider;
use Core\Api\RateLimit\RateLimitService;
use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Contracts\Cache\Repository as CacheRepository;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\ServiceProvider;

/**
 * API Module Boot.
 *
 * This module provides shared API controllers and middleware.
 * Routes are registered centrally in routes/api.php rather than
 * per-module, as API endpoints span multiple service modules.
 */
class Boot extends ServiceProvider
{
    /**
     * The module name.
     */
    protected string $moduleName = 'api';

    /**
     * Events this module listens to for lazy loading.
     *
     * @var array<class-string, string>
     */
    public static array $listens = [
        AdminPanelBooting::class => 'onAdminPanel',
        ApiRoutesRegistering::class => 'onApiRoutes',
        ConsoleBooting::class => 'onConsole',
    ];

    /**
     * Register any application services.
     */
    public function register(): void
    {
        $this->mergeConfigFrom(
            __DIR__.'/config.php',
            $this->moduleName
        );

        // Register RateLimitService as a singleton
        $this->app->singleton(RateLimitService::class, function ($app) {
            return new RateLimitService($app->make(CacheRepository::class));
        });

        // Register webhook services
        $this->app->singleton(Services\WebhookTemplateService::class);
        $this->app->singleton(Services\WebhookSecretRotationService::class);

        // Register IP restriction service for API key whitelisting
        $this->app->singleton(Services\IpRestrictionService::class);

        // Register API Documentation provider
        $this->app->register(DocumentationServiceProvider::class);
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/Migrations');
        $this->configureRateLimiting();
    }

    /**
     * Configure rate limiters for API endpoints.
     */
    protected function configureRateLimiting(): void
    {
        // Rate limit for webhook template operations: 30 per minute per user
        RateLimiter::for('api-webhook-templates', function (Request $request) {
            $user = $request->user();

            return $user
                ? Limit::perMinute(30)->by('user:'.$user->id)
                : Limit::perMinute(10)->by($request->ip());
        });

        // Rate limit for template preview/validation: 60 per minute per user
        RateLimiter::for('api-template-preview', function (Request $request) {
            $user = $request->user();

            return $user
                ? Limit::perMinute(60)->by('user:'.$user->id)
                : Limit::perMinute(20)->by($request->ip());
        });
    }

    // -------------------------------------------------------------------------
    // Event-driven handlers
    // -------------------------------------------------------------------------

    public function onAdminPanel(AdminPanelBooting $event): void
    {
        $event->views($this->moduleName, __DIR__.'/View/Blade');

        if (file_exists(__DIR__.'/Routes/admin.php')) {
            $event->routes(fn () => require __DIR__.'/Routes/admin.php');
        }

        $event->livewire('api.webhook-template-manager', View\Modal\Admin\WebhookTemplateManager::class);
    }

    public function onApiRoutes(ApiRoutesRegistering $event): void
    {
        // Middleware aliases registered via event
        $event->middleware('api.auth', Middleware\AuthenticateApiKey::class);
        $event->middleware('api.scope', Middleware\CheckApiScope::class);
        $event->middleware('api.scope.enforce', Middleware\EnforceApiScope::class);
        $event->middleware('api.rate', Middleware\RateLimitApi::class);
        $event->middleware('auth.api', Middleware\AuthenticateApiKey::class);

        // Core API routes (SEO, Pixel, Entitlements, MCP)
        if (file_exists(__DIR__.'/Routes/api.php')) {
            $event->routes(fn () => Route::middleware('api')->group(__DIR__.'/Routes/api.php'));
        }
    }

    public function onConsole(ConsoleBooting $event): void
    {
        // Register middleware aliases for CLI context (artisan route:list etc)
        $event->middleware('api.auth', Middleware\AuthenticateApiKey::class);
        $event->middleware('api.scope', Middleware\CheckApiScope::class);
        $event->middleware('api.scope.enforce', Middleware\EnforceApiScope::class);
        $event->middleware('api.rate', Middleware\RateLimitApi::class);
        $event->middleware('auth.api', Middleware\AuthenticateApiKey::class);

        // Register console commands
        $event->command(Console\Commands\CleanupExpiredGracePeriods::class);
        $event->command(Console\Commands\CheckApiUsageAlerts::class);
    }
}
