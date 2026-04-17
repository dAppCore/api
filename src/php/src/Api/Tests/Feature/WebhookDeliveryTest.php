<?php

declare(strict_types=1);

namespace Core\Api\Models {
    function dns_get_record(string $hostname, int $type = DNS_A | DNS_AAAA, mixed ...$args): array|false
    {
        if ($hostname === 'webhook-pinned.example.test') {
            return [
                ['ip' => '1.1.1.1'],
                ['ipv6' => '2606:4700:4700::1111'],
            ];
        }

        return \dns_get_record($hostname, $type, ...$args);
    }
}

namespace {

use Core\Api\Jobs\DeliverWebhookJob;
use Core\Api\Models\WebhookDelivery;
use Core\Api\Models\WebhookEndpoint;
use Core\Api\Services\WebhookService;
use Core\Api\Services\WebhookSignature;
use Core\Tenant\Models\Workspace;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Http;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

function webhookPendingRequestOptions(PendingRequest $request): array
{
    $reflection = new ReflectionProperty($request, 'options');

    return $reflection->getValue($request);
}

beforeEach(function () {
    Http::fake();

    $this->workspace = Workspace::factory()->create();
    $this->service = app(WebhookService::class);
    $this->signatureService = app(WebhookSignature::class);
});

// -----------------------------------------------------------------------------
// Webhook Signature Service
// -----------------------------------------------------------------------------

describe('Webhook Signature Service', function () {
    it('generates a 64-character secret', function () {
        $secret = $this->signatureService->generateSecret();

        expect($secret)->toBeString();
        expect(strlen($secret))->toBe(64);
    });

    it('generates unique secrets', function () {
        $secrets = [];
        for ($i = 0; $i < 100; $i++) {
            $secrets[] = $this->signatureService->generateSecret();
        }

        expect(array_unique($secrets))->toHaveCount(100);
    });

    it('signs payload with timestamp', function () {
        $payload = '{"event":"test"}';
        $secret = 'test_secret_key';
        $timestamp = 1704067200; // Fixed timestamp for testing

        $signature = $this->signatureService->sign($payload, $secret, $timestamp);

        // Verify it's a 64-character hex string (SHA256)
        expect($signature)->toBeString();
        expect(strlen($signature))->toBe(64);
        expect(ctype_xdigit($signature))->toBeTrue();

        // Verify signature is deterministic
        $signature2 = $this->signatureService->sign($payload, $secret, $timestamp);
        expect($signature)->toBe($signature2);
    });

    it('produces different signatures for different payloads', function () {
        $secret = 'test_secret_key';
        $timestamp = 1704067200;

        $sig1 = $this->signatureService->sign('{"a":1}', $secret, $timestamp);
        $sig2 = $this->signatureService->sign('{"a":2}', $secret, $timestamp);

        expect($sig1)->not->toBe($sig2);
    });

    it('produces different signatures for different timestamps', function () {
        $payload = '{"event":"test"}';
        $secret = 'test_secret_key';

        $sig1 = $this->signatureService->sign($payload, $secret, 1704067200);
        $sig2 = $this->signatureService->sign($payload, $secret, 1704067201);

        expect($sig1)->not->toBe($sig2);
    });

    it('produces different signatures for different secrets', function () {
        $payload = '{"event":"test"}';
        $timestamp = 1704067200;

        $sig1 = $this->signatureService->sign($payload, 'secret1', $timestamp);
        $sig2 = $this->signatureService->sign($payload, 'secret2', $timestamp);

        expect($sig1)->not->toBe($sig2);
    });

    it('verifies valid signature', function () {
        $payload = '{"event":"test","data":{"id":123}}';
        $secret = 'webhook_secret_abc123';
        $timestamp = time();

        $signature = $this->signatureService->sign($payload, $secret, $timestamp);

        $isValid = $this->signatureService->verify(
            $payload,
            $signature,
            $secret,
            $timestamp
        );

        expect($isValid)->toBeTrue();
    });

    it('rejects invalid signature', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $timestamp = time();

        $isValid = $this->signatureService->verify(
            $payload,
            'invalid_signature_abc123',
            $secret,
            $timestamp
        );

        expect($isValid)->toBeFalse();
    });

    it('rejects tampered payload', function () {
        $secret = 'webhook_secret_abc123';
        $timestamp = time();

        // Sign original payload
        $signature = $this->signatureService->sign('{"event":"test"}', $secret, $timestamp);

        // Verify with tampered payload
        $isValid = $this->signatureService->verify(
            '{"event":"test","hacked":true}',
            $signature,
            $secret,
            $timestamp
        );

        expect($isValid)->toBeFalse();
    });

    it('rejects tampered timestamp', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $originalTimestamp = time();

        // Sign with original timestamp
        $signature = $this->signatureService->sign($payload, $secret, $originalTimestamp);

        // Verify with different timestamp (simulating replay attack)
        $isValid = $this->signatureService->verifySignatureOnly(
            $payload,
            $signature,
            $secret,
            $originalTimestamp + 1
        );

        expect($isValid)->toBeFalse();
    });

    it('rejects expired timestamp', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $oldTimestamp = time() - 600; // 10 minutes ago

        $signature = $this->signatureService->sign($payload, $secret, $oldTimestamp);

        // Default tolerance is 5 minutes
        $isValid = $this->signatureService->verify(
            $payload,
            $signature,
            $secret,
            $oldTimestamp
        );

        expect($isValid)->toBeFalse();
    });

    it('accepts timestamp within tolerance', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $recentTimestamp = time() - 60; // 1 minute ago

        $signature = $this->signatureService->sign($payload, $secret, $recentTimestamp);

        $isValid = $this->signatureService->verify(
            $payload,
            $signature,
            $secret,
            $recentTimestamp
        );

        expect($isValid)->toBeTrue();
    });

    it('allows custom tolerance', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $oldTimestamp = time() - 600; // 10 minutes ago

        $signature = $this->signatureService->sign($payload, $secret, $oldTimestamp);

        // Verify with 15-minute tolerance
        $isValid = $this->signatureService->verify(
            $payload,
            $signature,
            $secret,
            $oldTimestamp,
            tolerance: 900
        );

        expect($isValid)->toBeTrue();
    });

    it('checks timestamp validity correctly', function () {
        $now = time();

        // Within tolerance
        expect($this->signatureService->isTimestampValid($now))->toBeTrue();
        expect($this->signatureService->isTimestampValid($now - 60))->toBeTrue();
        expect($this->signatureService->isTimestampValid($now - 299))->toBeTrue();

        // Outside tolerance
        expect($this->signatureService->isTimestampValid($now - 301))->toBeFalse();
        expect($this->signatureService->isTimestampValid($now - 600))->toBeFalse();

        // Future timestamp within tolerance
        expect($this->signatureService->isTimestampValid($now + 60))->toBeTrue();

        // Future timestamp outside tolerance
        expect($this->signatureService->isTimestampValid($now + 400))->toBeFalse();
    });

    it('returns correct headers', function () {
        $payload = '{"event":"test"}';
        $secret = 'webhook_secret_abc123';
        $timestamp = 1704067200;

        $headers = $this->signatureService->getHeaders($payload, $secret, $timestamp);

        expect($headers)->toHaveKey('X-Webhook-Signature');
        expect($headers)->toHaveKey('X-Webhook-Timestamp');
        expect($headers['X-Webhook-Timestamp'])->toBe($timestamp);
        expect($headers['X-Webhook-Signature'])->toBe(
            $this->signatureService->sign($payload, $secret, $timestamp)
        );
    });
});

// -----------------------------------------------------------------------------
// Webhook URL Safety
// -----------------------------------------------------------------------------

describe('Webhook URL Safety', function () {
    it('accepts public HTTP and HTTPS destinations', function () {
        WebhookEndpoint::assertSafeUrl('https://1.1.1.1/webhooks');
        WebhookEndpoint::assertSafeUrl('http://8.8.8.8/hooks');

        expect(true)->toBeTrue();
    });

    it('rejects private and loopback destinations', function () {
        expect(fn () => WebhookEndpoint::assertSafeUrl('http://127.0.0.1/webhooks'))
            ->toThrow(\InvalidArgumentException::class);

        expect(fn () => WebhookEndpoint::assertSafeUrl('http://10.0.0.1/webhooks'))
            ->toThrow(\InvalidArgumentException::class);

        expect(fn () => WebhookEndpoint::assertSafeUrl('http://[::1]/webhooks'))
            ->toThrow(\InvalidArgumentException::class);
    });

    it('rejects reserved and multicast destinations', function () {
        expect(fn () => WebhookEndpoint::assertSafeUrl('http://224.0.0.1/webhooks'))
            ->toThrow(\InvalidArgumentException::class);

        expect(fn () => WebhookEndpoint::assertSafeUrl('http://100.64.0.1/webhooks'))
            ->toThrow(\InvalidArgumentException::class);
    });

    it('rejects unsupported schemes', function () {
        expect(fn () => WebhookEndpoint::assertSafeUrl('ftp://example.com/webhooks'))
            ->toThrow(\InvalidArgumentException::class);
    });

    it('rejects embedded credentials', function () {
        expect(fn () => WebhookEndpoint::assertSafeUrl('https://user:pass@example.com/webhooks'))
            ->toThrow(\InvalidArgumentException::class);
    });

    it('rejects unsafe webhook URLs during creation', function () {
        expect(fn () => WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'http://127.0.0.1/webhooks',
            ['workspace.created']
        ))->toThrow(\InvalidArgumentException::class);
    });
});

// -----------------------------------------------------------------------------
// Webhook Endpoint Signing
// -----------------------------------------------------------------------------

describe('Webhook Endpoint Signing', function () {
    it('generates signature for payload with timestamp', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $payload = '{"event":"test"}';
        $timestamp = time();

        $signature = $endpoint->generateSignature($payload, $timestamp);

        expect($signature)->toBeString();
        expect(strlen($signature))->toBe(64);
    });

    it('verifies valid signature', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $payload = '{"event":"test","data":{"id":123}}';
        $timestamp = time();

        $signature = $endpoint->generateSignature($payload, $timestamp);

        $isValid = $endpoint->verifySignature($payload, $signature, $timestamp);

        expect($isValid)->toBeTrue();
    });

    it('rejects invalid signature', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $isValid = $endpoint->verifySignature(
            '{"event":"test"}',
            'invalid_signature',
            time()
        );

        expect($isValid)->toBeFalse();
    });

    it('rotates secret and invalidates old signatures', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $payload = '{"event":"test"}';
        $timestamp = time();

        // Sign with original secret
        $originalSignature = $endpoint->generateSignature($payload, $timestamp);

        // Rotate secret
        $newSecret = $endpoint->rotateSecret();
        $endpoint->refresh();

        // Old signature should be invalid
        $isValid = $endpoint->verifySignature($payload, $originalSignature, $timestamp);
        expect($isValid)->toBeFalse();

        // New signature should be valid
        $newSignature = $endpoint->generateSignature($payload, $timestamp);
        $isValid = $endpoint->verifySignature($payload, $newSignature, $timestamp);
        expect($isValid)->toBeTrue();

        // New secret should be 64 characters
        expect(strlen($newSecret))->toBe(64);
    });
});

// -----------------------------------------------------------------------------
// Webhook Service
// -----------------------------------------------------------------------------

describe('Webhook Service', function () {
    it('dispatches event to subscribed endpoints', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $deliveries = $this->service->dispatch(
            $this->workspace->id,
            'bio.created',
            ['bio_id' => 123, 'name' => 'Test Bio']
        );

        expect($deliveries)->toHaveCount(1);
        expect($deliveries[0]->event_type)->toBe('bio.created');
        expect($deliveries[0]->webhook_endpoint_id)->toBe($endpoint->id);
        expect($deliveries[0]->status)->toBe(WebhookDelivery::STATUS_PENDING);
    });

    it('does not dispatch to endpoints not subscribed to event', function () {
        WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.updated'] // Different event
        );

        $deliveries = $this->service->dispatch(
            $this->workspace->id,
            'bio.created',
            ['bio_id' => 123]
        );

        expect($deliveries)->toBeEmpty();
    });

    it('dispatches to wildcard subscribed endpoints', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['*'] // Subscribe to all events
        );

        $deliveries = $this->service->dispatch(
            $this->workspace->id,
            'any.event.type',
            ['data' => 'test']
        );

        expect($deliveries)->toHaveCount(1);
    });

    it('does not dispatch to inactive endpoints', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );
        $endpoint->update(['active' => false]);

        $deliveries = $this->service->dispatch(
            $this->workspace->id,
            'bio.created',
            ['bio_id' => 123]
        );

        expect($deliveries)->toBeEmpty();
    });

    it('does not dispatch to disabled endpoints', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );
        $endpoint->update(['disabled_at' => now()]);

        $deliveries = $this->service->dispatch(
            $this->workspace->id,
            'bio.created',
            ['bio_id' => 123]
        );

        expect($deliveries)->toBeEmpty();
    });

    it('returns webhook stats for workspace', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        // Create some deliveries
        WebhookDelivery::createForEvent($endpoint, 'bio.created', ['id' => 1]);
        $delivery2 = WebhookDelivery::createForEvent($endpoint, 'bio.created', ['id' => 2]);
        $delivery2->markSuccess(200);
        $delivery3 = WebhookDelivery::createForEvent($endpoint, 'bio.created', ['id' => 3]);
        $delivery3->markFailed(500, 'Server Error');

        $stats = $this->service->getStats($this->workspace->id);

        expect($stats['total'])->toBe(3);
        expect($stats['pending'])->toBe(1);
        expect($stats['success'])->toBe(1);
        expect($stats['retrying'])->toBe(1); // Failed with retries remaining
    });
});

// -----------------------------------------------------------------------------
// Webhook Delivery Job
// -----------------------------------------------------------------------------

describe('Webhook Delivery Job', function () {
    it('marks delivery as success on 2xx response', function () {
        Http::fake([
            'example.com/*' => Http::response(['received' => true], 200),
        ]);

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        $delivery->refresh();
        expect($delivery->status)->toBe(WebhookDelivery::STATUS_SUCCESS);
        expect($delivery->response_code)->toBe(200);
        expect($delivery->delivered_at)->not->toBeNull();
    });

    it('marks delivery as retrying on 5xx response', function () {
        Http::fake([
            'example.com/*' => Http::response('Server Error', 500),
        ]);

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        $delivery->refresh();
        expect($delivery->status)->toBe(WebhookDelivery::STATUS_RETRYING);
        expect($delivery->response_code)->toBe(500);
        expect($delivery->attempt)->toBe(2);
        expect($delivery->next_retry_at)->not->toBeNull();
    });

    it('marks delivery as failed after max retries', function () {
        Http::fake([
            'example.com/*' => Http::response('Server Error', 500),
        ]);

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );
        $delivery->update(['attempt' => WebhookDelivery::MAX_RETRIES]);

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        $delivery->refresh();
        expect($delivery->status)->toBe(WebhookDelivery::STATUS_FAILED);
    });

    it('includes correct signature and timestamp headers', function () {
        Http::fake(function ($request) {
            // Verify all required headers exist
            expect($request->hasHeader('X-Webhook-Signature'))->toBeTrue();
            expect($request->hasHeader('X-Webhook-Timestamp'))->toBeTrue();
            expect($request->hasHeader('X-Webhook-Event'))->toBeTrue();
            expect($request->hasHeader('X-Webhook-Id'))->toBeTrue();

            // Verify timestamp is a valid Unix timestamp
            $timestamp = $request->header('X-Webhook-Timestamp')[0];
            expect(is_numeric($timestamp))->toBeTrue();
            expect((int) $timestamp)->toBeGreaterThan(0);

            // Verify signature is a 64-character hex string
            $signature = $request->header('X-Webhook-Signature')[0];
            expect(strlen($signature))->toBe(64);
            expect(ctype_xdigit($signature))->toBeTrue();

            return Http::response(['ok' => true], 200);
        });

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        Http::assertSent(function ($request) {
            return $request->url() === 'https://example.com/webhook';
        });
    });

    it('sends verifiable signature', function () {
        $capturedRequest = null;

        Http::fake(function ($request) use (&$capturedRequest) {
            $capturedRequest = $request;

            return Http::response(['ok' => true], 200);
        });

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        // Verify the signature can be verified by a recipient
        $body = $capturedRequest->body();
        $signature = $capturedRequest->header('X-Webhook-Signature')[0];
        $timestamp = (int) $capturedRequest->header('X-Webhook-Timestamp')[0];

        $isValid = $endpoint->verifySignature($body, $signature, $timestamp);
        expect($isValid)->toBeTrue();
    });

    it('skips delivery if endpoint becomes inactive', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        // Deactivate endpoint after delivery created
        $endpoint->update(['active' => false]);

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        // Should not have made any HTTP requests
        Http::assertNothingSent();

        // Delivery should remain pending (skipped)
        $delivery->refresh();
        expect($delivery->status)->toBe(WebhookDelivery::STATUS_PENDING);
    });

    it('revalidates the destination before each outbound request', function () {
        Http::fake();

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        WebhookEndpoint::query()
            ->whereKey($endpoint->id)
            ->update(['url' => 'http://127.0.0.1/webhook']);

        $job = new DeliverWebhookJob($delivery);
        $job->handle();

        Http::assertNothingSent();

        $delivery->refresh();
        expect($delivery->status)->toBe(WebhookDelivery::STATUS_RETRYING);
        expect($delivery->response_code)->toBe(0);
    });

    it('disables redirects and pins validated webhook destinations', function () {
        if (! defined('CURLOPT_RESOLVE')) {
            $this->markTestSkipped('cURL extension is unavailable.');
        }

        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://webhook-pinned.example.test/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $job = new class($delivery) extends DeliverWebhookJob
        {
            public function exposeBuildRequest(array $deliveryPayload, int $timeout, array $curlOptions): PendingRequest
            {
                return $this->buildRequest($deliveryPayload, $timeout, $curlOptions);
            }
        };

        $request = $job->exposeBuildRequest(
            $delivery->getDeliveryPayload(),
            (int) config('api.webhooks.timeout', 30),
            WebhookEndpoint::curlResolveOptionsFor($endpoint->url)
        );

        $options = webhookPendingRequestOptions($request);

        expect($options['allow_redirects'] ?? null)->toBeFalse();
        expect($options['curl'][CURLOPT_RESOLVE] ?? [])->toContain('webhook-pinned.example.test:443:1.1.1.1');
        expect($options['curl'][CURLOPT_RESOLVE] ?? [])->toContain('webhook-pinned.example.test:443:[2606:4700:4700::1111]');
    });
});

// -----------------------------------------------------------------------------
// Webhook Endpoint Auto-Disable
// -----------------------------------------------------------------------------

describe('Webhook Endpoint Auto-Disable', function () {
    it('disables endpoint after consecutive failures', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        // Simulate 10 consecutive failures
        for ($i = 0; $i < 10; $i++) {
            $endpoint->recordFailure();
        }

        $endpoint->refresh();
        expect($endpoint->active)->toBeFalse();
        expect($endpoint->disabled_at)->not->toBeNull();
        expect($endpoint->failure_count)->toBe(10);
    });

    it('resets failure count on success', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        // Record some failures
        $endpoint->recordFailure();
        $endpoint->recordFailure();
        $endpoint->recordFailure();
        expect($endpoint->fresh()->failure_count)->toBe(3);

        // Record success
        $endpoint->recordSuccess();

        $endpoint->refresh();
        expect($endpoint->failure_count)->toBe(0);
    });

    it('can be re-enabled after being disabled', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        // Disable it
        $endpoint->update([
            'active' => false,
            'disabled_at' => now(),
            'failure_count' => 10,
        ]);

        // Re-enable
        $endpoint->enable();

        $endpoint->refresh();
        expect($endpoint->active)->toBeTrue();
        expect($endpoint->disabled_at)->toBeNull();
        expect($endpoint->failure_count)->toBe(0);
    });
});

// -----------------------------------------------------------------------------
// Delivery Payload Headers
// -----------------------------------------------------------------------------

describe('Delivery Payload Headers', function () {
    it('includes all required headers', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $payload = $delivery->getDeliveryPayload();

        expect($payload)->toHaveKey('headers');
        expect($payload)->toHaveKey('body');
        expect($payload['headers'])->toHaveKey('Content-Type');
        expect($payload['headers'])->toHaveKey('X-Webhook-Id');
        expect($payload['headers'])->toHaveKey('X-Webhook-Event');
        expect($payload['headers'])->toHaveKey('X-Webhook-Timestamp');
        expect($payload['headers'])->toHaveKey('X-Webhook-Signature');
    });

    it('uses provided timestamp', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $fixedTimestamp = 1704067200;
        $payload = $delivery->getDeliveryPayload($fixedTimestamp);

        expect($payload['headers']['X-Webhook-Timestamp'])->toBe((string) $fixedTimestamp);
    });

    it('generates valid signature in payload', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = WebhookDelivery::createForEvent(
            $endpoint,
            'bio.created',
            ['bio_id' => 123]
        );

        $payload = $delivery->getDeliveryPayload();

        $timestamp = (int) $payload['headers']['X-Webhook-Timestamp'];
        $signature = $payload['headers']['X-Webhook-Signature'];
        $body = $payload['body'];

        // Verify the signature is valid
        $isValid = $endpoint->verifySignature($body, $signature, $timestamp);
        expect($isValid)->toBeTrue();
    });

    it('rejects payloads that cannot be encoded as json', function () {
        $endpoint = WebhookEndpoint::createForWorkspace(
            $this->workspace->id,
            'https://example.com/webhook',
            ['bio.created']
        );

        $delivery = new WebhookDelivery([
            'event_id' => 'evt_invalid_payload',
            'event_type' => 'bio.created',
            'payload' => ['message' => "\xB1"],
        ]);
        $delivery->setRelation('endpoint', $endpoint);

        expect(fn () => $delivery->getDeliveryPayload())
            ->toThrow(RuntimeException::class, 'Unable to encode webhook payload as JSON.');
    });
});

}
