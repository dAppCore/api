<?php

declare(strict_types=1);

it('renders Stoplight Elements when selected as the default documentation ui', function () {
    config(['api-docs.ui.default' => 'stoplight']);

    $response = $this->get('/api/docs');

    $response->assertOk();
    $response->assertSee('elements-api', false);
    $response->assertSee('@stoplight/elements', false);
});

it('renders the dedicated Stoplight documentation route', function () {
    $response = $this->get('/api/docs/stoplight');

    $response->assertOk();
    $response->assertSee('elements-api', false);
    $response->assertSee('@stoplight/elements', false);
});
