<?php

declare(strict_types=1);

use Core\Api\Documentation\DocumentationController;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| API Documentation Routes
|--------------------------------------------------------------------------
|
| These routes serve the OpenAPI documentation and interactive API explorers.
| Protected by the ProtectDocumentation middleware for production environments.
|
*/

// Documentation UI routes
Route::get('/', [DocumentationController::class, 'index'])->name('api.docs');
Route::get('/swagger', [DocumentationController::class, 'swagger'])->name('api.docs.swagger');
Route::get('/scalar', [DocumentationController::class, 'scalar'])->name('api.docs.scalar');
Route::get('/redoc', [DocumentationController::class, 'redoc'])->name('api.docs.redoc');
Route::get('/stoplight', [DocumentationController::class, 'stoplight'])->name('api.docs.stoplight');

// OpenAPI specification routes
Route::get('/openapi.json', [DocumentationController::class, 'openApiJson'])
    ->name('api.docs.openapi.json')
    ->middleware('throttle:60,1');

Route::get('/openapi.yaml', [DocumentationController::class, 'openApiYaml'])
    ->name('api.docs.openapi.yaml')
    ->middleware('throttle:60,1');

// Cache management (admin only)
Route::post('/cache/clear', [DocumentationController::class, 'clearCache'])
    ->name('api.docs.cache.clear')
    ->middleware('auth');
