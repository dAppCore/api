<?php

declare(strict_types=1);

namespace Core\Api;

use Core\Api\Listeners\DispatchSubscriptionWebhookEvents;
use Core\Api\Models\Biolink;
use Core\Api\Models\Link;
use Core\Api\Models\SupportTicket;
use Core\Api\Models\SupportTicketReply;
use Core\Api\Observers\BiolinkWebhookObserver;
use Core\Api\Observers\LinkWebhookObserver;
use Core\Api\Observers\SupportTicketReplyWebhookObserver;
use Core\Api\Observers\SupportTicketWebhookObserver;
use Core\Api\Observers\WorkspaceWebhookObserver;
use Core\Events\AdminPanelBooting;
use Core\Events\ApiRoutesRegistering;
use Core\Events\ConsoleBooting;
use Core\Api\Documentation\DocumentationServiceProvider;
use Core\Api\RateLimit\RateLimitService;
use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Contracts\Cache\Repository as CacheRepository;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Event;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\ServiceProvider;
use Laravel\Passport\Passport;
use Laravel\Passport\Http\Controllers\AccessTokenController;
use Laravel\Passport\Http\Controllers\ApproveAuthorizationController;
use Laravel\Passport\Http\Controllers\AuthorizationController;
use Laravel\Passport\Http\Controllers\DenyAuthorizationController;
use Laravel\Passport\PassportServiceProvider;
use Core\Tenant\Models\Workspace;

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

        \Dedoc\Scramble\Scramble::ignoreDefaultRoutes();
        $this->app->register(\Dedoc\Scramble\ScrambleServiceProvider::class);

        if (class_exists(PassportServiceProvider::class)) {
            $this->app->register(PassportServiceProvider::class);
        }
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        $this->loadMigrationsFrom(__DIR__.'/Migrations');
        $this->configureRateLimiting();
        $this->configureOAuth();
        $this->registerFallbackApiRoutes();
        $this->registerWebhookHooks();
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

    /**
     * Configure OAuth2 support when Passport is installed in the host app.
     */
    protected function configureOAuth(): void
    {
        if (! class_exists(Passport::class)) {
            return;
        }

        Passport::tokensCan([
            'read' => 'Read access to resources',
            'write' => 'Write access to resources',
            'delete' => 'Delete access to resources',
        ]);
        Passport::setDefaultScope(['read']);

        if (method_exists(Passport::class, 'hashClientSecrets')) {
            Passport::hashClientSecrets();
        }
    }

    /**
     * Register webhook dispatch hooks for domain changes owned by this module.
     */
    protected function registerWebhookHooks(): void
    {
        Event::listen(\Core\Mod\Commerce\Events\SubscriptionUpdated::class, DispatchSubscriptionWebhookEvents::class);

        if (class_exists(Workspace::class)) {
            Workspace::observe(WorkspaceWebhookObserver::class);
        }

        Biolink::observe(BiolinkWebhookObserver::class);
        Link::observe(LinkWebhookObserver::class);
        SupportTicket::observe(SupportTicketWebhookObserver::class);
        SupportTicketReply::observe(SupportTicketReplyWebhookObserver::class);
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
        $this->registerMiddlewareAliases();

        // Core API routes (SEO, Pixel, Entitlements, MCP)
        if (file_exists(__DIR__.'/Routes/api.php') && ! $this->hasCoreApiRoutesRegistered()) {
            $event->routes(fn () => Route::middleware('api')->group(__DIR__.'/Routes/api.php'));
        }

        if (class_exists(Passport::class)) {
            $event->routes(fn () => $this->registerOAuthRoutes());
        }
    }

    public function onConsole(ConsoleBooting $event): void
    {
        $this->registerMiddlewareAliases();

        // Register console commands
        $event->command(Console\Commands\CleanupExpiredGracePeriods::class);
        $event->command(Console\Commands\CheckApiUsageAlerts::class);
    }

    /**
     * Register middleware aliases directly when the host event bus is not
     * wiring this package into the main route stack.
     *
     * This keeps the module usable in standalone package contexts such as
     * Testbench, route:list, and OpenAPI generation.
     */
    protected function registerMiddlewareAliases(): void
    {
        $router = $this->app['router'];

        $router->aliasMiddleware('api.auth', Middleware\AuthenticateApiKey::class);
        $router->aliasMiddleware('api.scope', Middleware\CheckApiScope::class);
        $router->aliasMiddleware('api.scope.enforce', Middleware\EnforceApiScope::class);
        $router->aliasMiddleware('api.rate', Middleware\RateLimitApi::class);
        $router->aliasMiddleware('api.cache', Middleware\ApiCacheControl::class);
        $router->aliasMiddleware('auth.api', Middleware\AuthenticateApiKey::class);
    }

    /**
     * Register the core API routes when the host application has not already
     * mounted them through the ApiRoutesRegistering event.
     */
    protected function registerFallbackApiRoutes(): void
    {
        $this->registerMiddlewareAliases();

        if (! file_exists(__DIR__.'/Routes/api.php') || $this->hasCoreApiRoutesRegistered()) {
            return;
        }

        Route::prefix('api')
            ->middleware('api')
            ->group(__DIR__.'/Routes/api.php');

        if (class_exists(Passport::class) && ! Route::has('passport.token')) {
            $this->registerOAuthRoutes();
        }
    }

    /**
     * Detect whether the package's core API routes have already been mounted.
     */
    protected function hasCoreApiRoutesRegistered(): bool
    {
        return Route::getRoutes()->getByName('api.pixel.track') !== null;
    }

    /**
     * Register OAuth routes when Passport is installed in the host app.
     */
    protected function registerOAuthRoutes(): void
    {
        Route::prefix('oauth')->group(function () {
            Route::post('/token', [AccessTokenController::class, 'issueToken'])
                ->middleware(['throttle:60,1'])
                ->name('passport.token');

            Route::middleware(['web', 'auth'])->group(function () {
                Route::get('/authorize', [AuthorizationController::class, 'authorize'])
                    ->name('passport.authorizations.authorize');
                Route::post('/authorize', [ApproveAuthorizationController::class, 'approve'])
                    ->name('passport.authorizations.approve');
                Route::delete('/authorize', [DenyAuthorizationController::class, 'deny'])
                    ->name('passport.authorizations.deny');
            });
        });
    }
}
