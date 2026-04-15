<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Core\Website\Api\Controllers\DocsController;

// Documentation landing
Route::get('/', [DocsController::class, 'index'])->name('api.docs');

// Guides
Route::get('/guides', [DocsController::class, 'guides'])->name('api.guides');
Route::get('/guides/quickstart', [DocsController::class, 'quickstart'])->name('api.guides.quickstart');
Route::get('/guides/authentication', [DocsController::class, 'authentication'])->name('api.guides.authentication');
Route::get('/guides/qrcodes', [DocsController::class, 'qrcodes'])->name('api.guides.qrcodes');
Route::get('/guides/webhooks', [DocsController::class, 'webhooks'])->name('api.guides.webhooks');
Route::get('/guides/errors', [DocsController::class, 'errors'])->name('api.guides.errors');

// API Reference
Route::get('/reference', [DocsController::class, 'reference'])->name('api.reference');
Route::get('/docs/api', [DocsController::class, 'api'])->name('api.docs.api');
Route::get('/openapi.yaml', [DocsController::class, 'openapiYaml'])
    ->middleware('throttle:60,1')
    ->name('api.openapi.yaml');
Route::get('/sdks', [DocsController::class, 'sdks'])->name('api.sdks');
Route::get('/sdks/{language}', [DocsController::class, 'sdkDownload'])->name('api.sdks.language');

// Swagger UI
Route::get('/swagger', [DocsController::class, 'swagger'])->name('api.swagger');

// Scalar (modern API reference with sidebar)
Route::get('/scalar', [DocsController::class, 'scalar'])->name('api.scalar');

// ReDoc (three-panel API reference)
Route::get('/redoc', [DocsController::class, 'redoc'])->name('api.redoc');

// Stoplight Elements API reference
Route::get('/stoplight', [DocsController::class, 'stoplight'])->name('api.stoplight');

// OpenAPI spec (rate limited - expensive to generate)
Route::get('/openapi.json', [DocsController::class, 'openapi'])
    ->middleware('throttle:60,1')
    ->name('api.openapi.json');
