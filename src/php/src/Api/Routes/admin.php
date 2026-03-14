<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Route;
use Mod\Api\View\Modal\Admin\WebhookTemplateManager;

/*
|--------------------------------------------------------------------------
| Admin Routes
|--------------------------------------------------------------------------
|
| Routes for the Api module's admin panel.
|
*/

Route::prefix('hub/api')->name('hub.api.')->group(function () {
    Route::get('/webhook-templates', WebhookTemplateManager::class)->name('webhook-templates');
});
