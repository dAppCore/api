<?php

declare(strict_types=1);

use Core\Api\Controllers\Api\UnifiedPixelController;
use Core\Api\Controllers\Api\EntitlementApiController;
use Core\Api\Controllers\Api\SeoReportController;
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
