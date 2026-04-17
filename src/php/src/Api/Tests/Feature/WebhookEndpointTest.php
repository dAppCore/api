<?php

declare(strict_types=1);

namespace Core\Api\Models {
    function dns_get_record(string $hostname, int $type = DNS_A | DNS_AAAA, mixed ...$args): array|false
    {
        if ($hostname === 'webhook-cname.example.test') {
            return [
                [
                    'type' => 'CNAME',
                    'target' => 'webhook-cname-target.example.test.',
                ],
            ];
        }

        if ($hostname === 'webhook-cname-target.example.test.') {
            return [
                ['ip' => '1.1.1.1'],
                ['ipv6' => '2606:4700:4700::1111'],
            ];
        }

        if ($hostname === 'webhook-private-cname.example.test') {
            return [
                [
                    'type' => 'CNAME',
                    'target' => 'webhook-private-target.example.test.',
                ],
            ];
        }

        if ($hostname === 'webhook-private-target.example.test.') {
            return [
                ['ip' => '10.0.0.1'],
            ];
        }

        return \dns_get_record($hostname, $type, ...$args);
    }
}

namespace {

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

it('WebhookEndpoint_curlResolveOptionsFor_Good_allows_public_ipv4_mapped_ipv6_literals', function () {
    $options = WebhookEndpoint::curlResolveOptionsFor('https://[::ffff:1.1.1.1]/webhooks');

    expect($options)->toBe([]);
});

it('WebhookEndpoint_curlResolveOptionsFor_Good_follows_cname_chains_to_public_ips', function () {
    if (! defined('CURLOPT_RESOLVE')) {
        $this->markTestSkipped('cURL extension is unavailable.');
    }

    $options = WebhookEndpoint::curlResolveOptionsFor('https://webhook-cname.example.test/webhooks');

    expect($options)->toHaveKey(CURLOPT_RESOLVE);
    expect($options[CURLOPT_RESOLVE])->toContain('webhook-cname.example.test:443:1.1.1.1');
    expect($options[CURLOPT_RESOLVE])->toContain('webhook-cname.example.test:443:[2606:4700:4700::1111]');
});

it('WebhookEndpoint_curlResolveOptionsFor_Bad_blocks_cname_chains_to_private_ips', function () {
    expect(fn () => WebhookEndpoint::curlResolveOptionsFor('https://webhook-private-cname.example.test/webhooks'))
        ->toThrow(\InvalidArgumentException::class, 'private, loopback, or reserved addresses');
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

}
