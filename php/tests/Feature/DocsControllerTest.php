<?php

declare(strict_types=1);

use Core\Website\Api\Controllers\DocsController;

function assertDocsView(object $view, string $expectedName): void
{
    expect($view->name())->toBe($expectedName);
}

it('DocsController_index_Good_returns_the_landing_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->index(), 'api::index');
});

it('DocsController_docs_Good_delegates_to_the_swagger_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->docs(), 'api::swagger');
});

it('DocsController_guides_Good_returns_the_guides_index_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->guides(), 'api::guides.index');
});

it('DocsController_quickstart_Good_returns_the_quickstart_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->quickstart(), 'api::guides.quickstart');
});

it('DocsController_authentication_Good_returns_the_authentication_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->authentication(), 'api::guides.authentication');
});

it('DocsController_qrcodes_Good_returns_the_qrcodes_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->qrcodes(), 'api::guides.qrcodes');
});

it('DocsController_webhooks_Good_returns_the_webhooks_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->webhooks(), 'api::guides.webhooks');
});

it('DocsController_rateLimits_Good_returns_the_rate_limits_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->rateLimits(), 'api::guides.rate-limits');
});

it('DocsController_errors_Good_returns_the_errors_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->errors(), 'api::guides.errors');
});

it('DocsController_changelog_Good_returns_the_changelog_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->changelog(), 'api::changelog');
});

it('DocsController_reference_Good_returns_the_reference_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->reference(), 'api::reference');
});

it('DocsController_api_Good_returns_the_docs_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->api(), 'api::docs');
});

it('DocsController_swagger_Good_returns_the_swagger_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->swagger(), 'api::swagger');
});

it('DocsController_scalar_Good_returns_the_scalar_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->scalar(), 'api::scalar');
});

it('DocsController_redoc_Good_returns_the_redoc_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->redoc(), 'api::redoc');
});

it('DocsController_stoplight_Good_returns_the_stoplight_view', function () {
    $controller = new DocsController;

    assertDocsView($controller->stoplight(), 'api::stoplight');
});

it('DocsController_sdks_Good_returns_the_sdk_landing_view_without_a_language', function () {
    $controller = new DocsController;

    $view = $controller->sdks();

    expect($view->name())->toBe('api::sdks');
    expect($view->getData()['language'])->toBeNull();
});

it('DocsController_sdkDownload_Ugly_preserves_unusual_language_values', function () {
    $controller = new DocsController;

    $view = $controller->sdkDownload('../../../../ruby');

    expect($view->name())->toBe('api::sdks');
    expect($view->getData()['language'])->toBe('../../../../ruby');
});
