<?php

declare(strict_types=1);

use Core\Api\Controllers\Api\UnifiedPixelController;
use Core\Api\Controllers\Api\EntitlementApiController;
use Core\Api\Controllers\Api\ApiKeyController;
use Core\Api\Controllers\Api\SeoReportController;
use Core\Api\Controllers\Api\WebhookSecretController;
use Core\Api\Controllers\Api\WebhookController;
use Core\Api\Controllers\McpApiController;
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

Route::middleware(['auth.api', 'api.scope.enforce'])
    ->prefix('seo')
    ->name('api.seo.')
    ->group(function () {
        Route::get('/report', [SeoReportController::class, 'show'])
            ->name('report');
    });

// ─────────────────────────────────────────────────────────────────────────────
// Entitlements (authenticated)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['auth.api', 'api.scope.enforce'])
    ->prefix('entitlements')
    ->name('api.entitlements.')
    ->group(function () {
        Route::get('/', [EntitlementApiController::class, 'show'])
            ->name('show');
    });

// ─────────────────────────────────────────────────────────────────────────────
// Webhook secret rotation (authenticated)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['auth.api', 'api.scope.enforce'])
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

Route::middleware(['auth.api', 'api.scope.enforce'])
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
                Route::get('/', [WebhookController::class, 'index'])->name('index');
                Route::post('/', [WebhookController::class, 'store'])->name('store');
                Route::get('/{id}', [WebhookController::class, 'show'])->name('show');
                Route::patch('/{id}', [WebhookController::class, 'update'])->name('update');
                Route::delete('/{id}', [WebhookController::class, 'destroy'])->name('destroy');
                Route::get('/{id}/deliveries', [WebhookController::class, 'deliveries'])->name('deliveries');
            });
    });

// ─────────────────────────────────────────────────────────────────────────────
// MCP HTTP Bridge (API key auth)
// ─────────────────────────────────────────────────────────────────────────────

Route::middleware(['throttle:120,1', McpApiKeyAuth::class, 'api.scope.enforce'])
    ->prefix('mcp')
    ->name('api.mcp.')
    ->group(function () {
        // Scope enforcement: GET=read, POST=write
        // Server discovery (read)
        Route::get('/servers', [McpApiController::class, 'servers'])
            ->name('servers');
        Route::get('/servers/{id}', [McpApiController::class, 'server'])
            ->name('servers.show');
        Route::get('/servers/{id}/tools', [McpApiController::class, 'tools'])
            ->name('servers.tools');
        Route::get('/servers/{id}/resources', [McpApiController::class, 'resources'])
            ->name('servers.resources');

        // Tool version history (read)
        Route::get('/servers/{server}/tools/{tool}/versions', [McpApiController::class, 'toolVersions'])
            ->name('tools.versions');

        // Specific tool version (read)
        Route::get('/servers/{server}/tools/{tool}/versions/{version}', [McpApiController::class, 'toolVersion'])
            ->name('tools.version');

        // Tool execution (write)
        Route::post('/tools/call', [McpApiController::class, 'callTool'])
            ->name('tools.call');

        // Resource access (read)
        Route::get('/resources/{uri}', [McpApiController::class, 'resource'])
            ->where('uri', '.*')
            ->name('resources.show');
    });
