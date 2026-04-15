<?php

declare(strict_types=1);

use Core\Api\Controllers\Api\UnifiedPixelController;
use Core\Api\Controllers\Api\AuthController;
use Core\Api\Controllers\Api\BiolinkController;
use Core\Api\Controllers\Api\EntitlementApiController;
use Core\Api\Controllers\Api\ApiKeyController;
use Core\Api\Controllers\Api\LinkController;
use Core\Api\Controllers\Api\PaymentMethodController;
use Core\Api\Controllers\Api\QrCodeController;
use Core\Api\Controllers\Api\TicketController;
use Core\Api\Controllers\Api\WorkspaceMemberController;
use Core\Api\Controllers\Api\SeoReportController;
use Core\Api\Controllers\Api\WebhookSecretController;
use Core\Api\Controllers\Api\WebhookController;
use Core\Api\Controllers\McpApiController;
use Core\Mod\Commerce\Controllers\Api\CommerceController;
use Core\Tenant\Controllers\WorkspaceController;
use Core\Api\Middleware\PublicApiCors;
use Core\Mcp\Middleware\McpApiKeyAuth;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| Core API Routes
|--------------------------------------------------------------------------
|
| Core API routes for cross-cutting concerns.
|
| SEO, pixel tracking, entitlements, and MCP bridge endpoints.
|
*/

// ─────────────────────────────────────────────────────────────────────────────
// Unified Pixel (public tracking)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware([PublicApiCors::class, 'api.rate'])
    ->prefix('pixel')
    ->name('api.pixel.')
    ->group(function () {
        Route::match(['GET', 'POST', 'OPTIONS'], '/{pixelKey}', [UnifiedPixelController::class, 'track'])
            ->name('track');
    });

// ─────────────────────────────────────────────────────────────────────────────
// SEO analysis (authenticated)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['auth.api', 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('seo')
    ->name('api.seo.')
    ->group(function () {
        Route::get('/report', [SeoReportController::class, 'show'])
            ->name('report');
    });

// ─────────────────────────────────────────────────────────────────────────────
// Entitlements (authenticated)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['auth.api', 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('entitlements')
    ->name('api.entitlements.')
    ->group(function () {
        Route::get('/', [EntitlementApiController::class, 'show'])
            ->name('show')
            ->defaults('api_cache_control', 'cacheable');
    });

// ─────────────────────────────────────────────────────────────────────────────
// Webhook secret rotation (authenticated)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['auth.api', 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('webhooks')
    ->name('api.webhooks.')
    ->group(function () {
        Route::prefix('social/{uuid}/secret')
            ->name('social.')
            ->group(function () {
                Route::post('/rotate', [WebhookSecretController::class, 'rotateSocialSecret'])
                    ->name('rotate-secret');
                Route::get('/', [WebhookSecretController::class, 'socialSecretStatus'])
                    ->name('status');
                Route::delete('/previous', [WebhookSecretController::class, 'invalidateSocialPreviousSecret'])
                    ->name('invalidate-previous');
                Route::patch('/grace-period', [WebhookSecretController::class, 'updateSocialGracePeriod'])
                    ->name('grace-period');
            });

        Route::prefix('content/{uuid}/secret')
            ->name('content.')
            ->group(function () {
                Route::post('/rotate', [WebhookSecretController::class, 'rotateContentSecret'])
                    ->name('rotate-secret');
                Route::get('/', [WebhookSecretController::class, 'contentSecretStatus'])
                    ->name('status');
                Route::delete('/previous', [WebhookSecretController::class, 'invalidateContentPreviousSecret'])
                    ->name('invalidate-previous');
                Route::patch('/grace-period', [WebhookSecretController::class, 'updateContentGracePeriod'])
                    ->name('grace-period');
            });
    });

// ─────────────────────────────────────────────────────────────────────────────
// Versioned API surface (workspace-scoped resource management)
// ─────────────────────────────────────────────────────────────────────────────

Route::prefix('v1/auth')
    ->name('api.v1.auth.')
    ->group(function () {
        Route::post('/token', [AuthController::class, 'store'])
            ->middleware(['api.rate', 'api.cache:ephemeral'])
            ->name('token.store');
        Route::middleware(['auth.api', 'api.rate', 'api.cache:ephemeral'])->group(function () {
            Route::delete('/token', [AuthController::class, 'destroy'])->name('token.destroy');
            Route::get('/me', [AuthController::class, 'show'])->name('me');
        });
    });

Route::middleware(['auth.api', 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('v1')
    ->name('api.v1.')
    ->group(function () {
        Route::prefix('workspaces')
            ->name('workspaces.')
            ->group(function () {
                Route::get('/', [WorkspaceController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                Route::get('/current', [WorkspaceController::class, 'current'])->name('current')->defaults('api_cache_control', 'cacheable');
                Route::post('/', [WorkspaceController::class, 'store'])->name('store');
                Route::get('/{workspace}', [WorkspaceController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                Route::put('/{workspace}', [WorkspaceController::class, 'update'])->name('update');
                Route::patch('/{workspace}', [WorkspaceController::class, 'update'])->name('patch');
                Route::delete('/{workspace}', [WorkspaceController::class, 'destroy'])->name('destroy');
                Route::post('/{workspace}/switch', [WorkspaceController::class, 'switch'])->name('switch');

                Route::prefix('{workspace}/members')
                    ->name('members.')
                    ->group(function () {
                        Route::get('/', [WorkspaceMemberController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                        Route::post('/', [WorkspaceMemberController::class, 'store'])->name('store');
                        Route::delete('/{user}', [WorkspaceMemberController::class, 'destroy'])->name('destroy');
                    });

                Route::prefix('{workspace}/entitlements')
                    ->name('entitlements.')
                    ->group(function () {
                        Route::get('/', [EntitlementApiController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                        Route::get('/check/{feature}', [EntitlementApiController::class, 'check'])->name('check')->defaults('api_cache_control', 'cacheable');
                        Route::get('/usage', [EntitlementApiController::class, 'usage'])->name('usage')->defaults('api_cache_control', 'cacheable');
                    });

                Route::prefix('{workspace}/biolinks')
                    ->name('biolinks.')
                    ->group(function () {
                        Route::get('/', [BiolinkController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                        Route::post('/', [BiolinkController::class, 'store'])->name('store');
                        Route::get('/{id}', [BiolinkController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                        Route::patch('/{id}', [BiolinkController::class, 'update'])->name('update');
                        Route::delete('/{id}', [BiolinkController::class, 'destroy'])->name('destroy');
                    });

                Route::prefix('{workspace}/links')
                    ->name('links.')
                    ->group(function () {
                        Route::get('/', [LinkController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                        Route::post('/', [LinkController::class, 'store'])->name('store');
                        Route::get('/{id}', [LinkController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                        Route::patch('/{id}', [LinkController::class, 'update'])->name('update');
                        Route::delete('/{id}', [LinkController::class, 'destroy'])->name('destroy');
                        Route::get('/{id}/stats', [LinkController::class, 'stats'])->name('stats')->defaults('api_cache_control', 'cacheable');
                    });

                Route::prefix('{workspace}/qr-codes')
                    ->name('qr-codes.')
                    ->group(function () {
                        Route::get('/', [QrCodeController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                        Route::post('/', [QrCodeController::class, 'store'])->name('store');
                        Route::get('/{id}', [QrCodeController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                        Route::get('/{id}/download', [QrCodeController::class, 'download'])->name('download')->defaults('api_cache_control', 'cacheable');
                    });
            });

        Route::prefix('commerce')
            ->name('commerce.')
            ->group(function () {
                Route::get('/subscriptions', [CommerceController::class, 'subscription'])->name('subscriptions.show')->defaults('api_cache_control', 'cacheable');
                Route::post('/subscriptions/change', [CommerceController::class, 'previewUpgrade'])->name('subscriptions.change');
                Route::post('/subscriptions/change/confirm', [CommerceController::class, 'executeUpgrade'])->name('subscriptions.change.confirm');
                Route::post('/subscriptions/cancel', [CommerceController::class, 'cancelSubscription'])->name('subscriptions.cancel');
                Route::post('/subscriptions/resume', [CommerceController::class, 'resumeSubscription'])->name('subscriptions.resume');

                Route::get('/invoices', [CommerceController::class, 'invoices'])->name('invoices.index')->defaults('api_cache_control', 'cacheable');
                Route::get('/invoices/{invoice}', [CommerceController::class, 'showInvoice'])->name('invoices.show')->defaults('api_cache_control', 'cacheable');
                Route::get('/invoices/{invoice}/pdf', [CommerceController::class, 'downloadInvoice'])->name('invoices.pdf')->defaults('api_cache_control', 'cacheable');

                Route::get('/payment-methods', [PaymentMethodController::class, 'index'])->name('payment-methods.index')->defaults('api_cache_control', 'cacheable');
                Route::post('/payment-methods', [PaymentMethodController::class, 'store'])->name('payment-methods.store');
                Route::delete('/payment-methods/{id}', [PaymentMethodController::class, 'destroy'])->name('payment-methods.destroy');
                Route::post('/payment-methods/{id}/default', [PaymentMethodController::class, 'default'])->name('payment-methods.default');

                // Compatibility aliases for older route shapes.
                Route::get('/subscription', [CommerceController::class, 'subscription'])->name('subscription')->defaults('api_cache_control', 'cacheable');
                Route::post('/cancel', [CommerceController::class, 'cancelSubscription'])->name('cancel');
                Route::post('/resume', [CommerceController::class, 'resumeSubscription'])->name('resume');
                Route::get('/invoices/{invoice}/download', [CommerceController::class, 'downloadInvoice'])->name('invoices.download')->defaults('api_cache_control', 'cacheable');
                Route::post('/upgrade/preview', [CommerceController::class, 'previewUpgrade'])->name('upgrade.preview');
                Route::post('/upgrade', [CommerceController::class, 'executeUpgrade'])->name('upgrade');
            });

        Route::prefix('support/tickets')
            ->name('support.tickets.')
            ->group(function () {
                Route::get('/', [TicketController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                Route::post('/', [TicketController::class, 'store'])->name('store');
                Route::get('/{id}', [TicketController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                Route::post('/{id}/reply', [TicketController::class, 'reply'])->name('reply');
            });
    });

Route::middleware(['web', 'auth:sanctum', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('v1')
    ->name('api.v1.')
    ->group(function () {
        Route::prefix('api-keys')
            ->name('api-keys.')
            ->group(function () {
                Route::get('/', [ApiKeyController::class, 'index'])->name('index');
                Route::post('/', [ApiKeyController::class, 'store'])->name('store');
                Route::delete('/{id}', [ApiKeyController::class, 'destroy'])->name('destroy');
            });

        Route::prefix('webhooks')
            ->name('webhooks.')
            ->group(function () {
                Route::get('/', [WebhookController::class, 'index'])->name('index')->defaults('api_cache_control', 'cacheable');
                Route::post('/', [WebhookController::class, 'store'])->name('store');
                Route::get('/{id}', [WebhookController::class, 'show'])->name('show')->defaults('api_cache_control', 'cacheable');
                Route::patch('/{id}', [WebhookController::class, 'update'])->name('update');
                Route::delete('/{id}', [WebhookController::class, 'destroy'])->name('destroy');
                Route::get('/{id}/deliveries', [WebhookController::class, 'deliveries'])->name('deliveries')->defaults('api_cache_control', 'cacheable');
            });
    });

// ─────────────────────────────────────────────────────────────────────────────
// MCP HTTP Bridge (API key auth)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['throttle:120,1', McpApiKeyAuth::class, 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('mcp')
    ->name('api.mcp.')
    ->group(function () {
        // Scope enforcement: GET=read, POST=write
        // Server discovery (read)
        Route::get('/servers', [McpApiController::class, 'servers'])
            ->name('servers')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}', [McpApiController::class, 'server'])
            ->name('servers.show')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}/tools', [McpApiController::class, 'tools'])
            ->name('servers.tools')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}/resources', [McpApiController::class, 'resources'])
            ->name('servers.resources')
            ->defaults('api_cache_control', 'cacheable');

        // Tool version history (read)
        Route::get('/servers/{server}/tools/{tool}/versions', [McpApiController::class, 'toolVersions'])
            ->name('tools.versions')
            ->defaults('api_cache_control', 'cacheable');

        // Specific tool version (read)
        Route::get('/servers/{server}/tools/{tool}/versions/{version}', [McpApiController::class, 'toolVersion'])
            ->name('tools.version')
            ->defaults('api_cache_control', 'cacheable');

        // Tool execution (write)
        Route::post('/servers/{server}/tools/{tool}', [McpApiController::class, 'callToolByRoute'])
            ->name('tools.call.route');
        Route::post('/tools/call', [McpApiController::class, 'callTool'])
            ->name('tools.call');

        // Resource access (read)
        Route::get('/resources/{uri}', [McpApiController::class, 'resource'])
            ->where('uri', '.*')
            ->name('resources.show')
            ->defaults('api_cache_control', 'cacheable');
    });

// Versioned MCP bridge aliases for RFC compatibility.
Route::middleware(['throttle:120,1', McpApiKeyAuth::class, 'api.scope.enforce', 'api.rate', 'api.cache:ephemeral'])
    ->prefix('v1/mcp')
    ->name('api.v1.mcp.')
    ->group(function () {
        Route::get('/servers', [McpApiController::class, 'servers'])
            ->name('servers')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}', [McpApiController::class, 'server'])
            ->name('servers.show')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}/tools', [McpApiController::class, 'tools'])
            ->name('servers.tools')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{id}/resources', [McpApiController::class, 'resources'])
            ->name('servers.resources')
            ->defaults('api_cache_control', 'cacheable');

        Route::get('/servers/{server}/tools/{tool}/versions', [McpApiController::class, 'toolVersions'])
            ->name('tools.versions')
            ->defaults('api_cache_control', 'cacheable');
        Route::get('/servers/{server}/tools/{tool}/versions/{version}', [McpApiController::class, 'toolVersion'])
            ->name('tools.version')
            ->defaults('api_cache_control', 'cacheable');

        Route::post('/servers/{server}/tools/{tool}', [McpApiController::class, 'callToolByRoute'])
            ->name('tools.call.route');
        Route::post('/tools/call', [McpApiController::class, 'callTool'])
            ->name('tools.call');

        Route::get('/resources/{uri}', [McpApiController::class, 'resource'])
            ->where('uri', '.*')
            ->name('resources.show')
            ->defaults('api_cache_control', 'cacheable');
    });
