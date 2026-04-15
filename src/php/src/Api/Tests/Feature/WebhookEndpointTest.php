<?php

declare(strict_types=1);

use Core\Api\Models\WebhookEndpoint;

it('WebhookEndpoint_shouldReceive_Good_matches_legacy_bio_aliases', function () {
    $endpoint = new WebhookEndpoint([
        'events' => ['biolink.created'],
        'active' => true,
    ]);

    expect($endpoint->shouldReceive('bio.created'))->toBeTrue();
    expect($endpoint->shouldReceive('biolink.created'))->toBeTrue();
});

it('WebhookEndpoint_curlResolveOptionsFor_Good_returns_empty_options_for_literal_ips', function () {
    $options = WebhookEndpoint::curlResolveOptionsFor('https://1.1.1.1/webhooks');

    expect($options)->toBe([]);
});

it('WebhookEndpoint_shouldReceive_Bad_rejects_inactive_endpoints', function () {
    $endpoint = new WebhookEndpoint([
        'events' => ['biolink.created'],
        'active' => false,
    ]);

    expect($endpoint->shouldReceive('bio.created'))->toBeFalse();
});

it('WebhookEndpoint_shouldReceive_Ugly_rejects_disabled_endpoints', function () {
    $endpoint = new WebhookEndpoint([
        'events' => ['biolink.created'],
        'active' => true,
        'disabled_at' => now(),
    ]);

    expect($endpoint->shouldReceive('bio.created'))->toBeFalse();
});
