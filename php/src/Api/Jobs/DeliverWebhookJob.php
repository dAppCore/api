<?php

declare(strict_types=1);

namespace Core\Api\Jobs;

use Core\Api\Models\WebhookDelivery;
use Core\Api\Models\WebhookEndpoint;
use Illuminate\Bus\Queueable;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

/**
 * Delivers webhook payloads to registered endpoints.
 *
 * Implements exponential backoff retry logic:
 * - Attempt 1: Immediate
 * - Attempt 2: 1 minute delay
 * - Attempt 3: 5 minutes delay
 * - Attempt 4: 30 minutes delay
 * - Attempt 5: 2 hours delay
 * - Attempt 6 (final): 24 hours delay
 */
class DeliverWebhookJob implements ShouldQueue
{
    use Dispatchable;
    use InteractsWithQueue;
    use Queueable;
    use SerializesModels;

    /**
     * Delete the job if its models no longer exist.
     */
    public bool $deleteWhenMissingModels = true;

    /**
     * The number of times the job may be attempted.
     * We handle retries manually with exponential backoff.
     */
    public int $tries = 1;

    /**
     * Create a new job instance.
     */
    public function __construct(
        public WebhookDelivery $delivery
    ) {
        // Use dedicated webhook queue if configured
        $this->queue = config('api.webhooks.queue', 'default');

        $connection = config('api.webhooks.queue_connection');
        if ($connection) {
            $this->connection = $connection;
        }
    }

    /**
     * Execute the job.
     */
    public function handle(): void
    {
        // Don't deliver if endpoint is disabled
        $endpoint = $this->delivery->endpoint;
        if (! $endpoint || ! $endpoint->shouldReceive($this->delivery->event_type)) {
            Log::info('Webhook delivery skipped - endpoint inactive or does not receive this event', [
                'delivery_id' => $this->delivery->id,
                'event_type' => $this->delivery->event_type,
            ]);

            return;
        }

        try {
            $curlOptions = WebhookEndpoint::curlResolveOptionsFor($endpoint->url);
        } catch (\InvalidArgumentException $e) {
            $this->handleFailure(0, 'Unsafe webhook destination blocked.');
            Log::warning('Webhook delivery blocked by URL safety check', [
                'delivery_id' => $this->delivery->id,
                'endpoint_url' => $this->redactUrlForLog($endpoint->url),
            ]);

            return;
        }

        // Get delivery payload with signature headers
        $deliveryPayload = $this->delivery->getDeliveryPayload();
        $timeout = config('api.webhooks.timeout', 30);

        Log::info('Attempting webhook delivery', [
            'delivery_id' => $this->delivery->id,
            'endpoint_url' => $this->redactUrlForLog($endpoint->url),
            'event_type' => $this->delivery->event_type,
            'attempt' => $this->delivery->attempt,
        ]);

        try {
            $response = $this->buildRequest($deliveryPayload, $timeout, $curlOptions)
                ->post($endpoint->url);

            $statusCode = $response->status();
            $responseBody = $response->body();

            // Success is any 2xx status code
            if ($response->successful()) {
                $this->delivery->markSuccess($statusCode, $responseBody);

                Log::info('Webhook delivered successfully', [
                    'delivery_id' => $this->delivery->id,
                    'status_code' => $statusCode,
                ]);

                return;
            }

            // Non-2xx response - mark as failed and potentially retry
            $this->handleFailure($statusCode, $responseBody);

        } catch (\Illuminate\Http\Client\ConnectionException $e) {
            // Connection timeout or refused
            $this->handleFailure(0, 'Connection failed: '.$e->getMessage());

        } catch (\Throwable $e) {
            // Unexpected error
            $this->handleFailure(0, 'Unexpected error: '.$e->getMessage());

            Log::error('Webhook delivery unexpected error', [
                'delivery_id' => $this->delivery->id,
                'error_type' => $e::class,
                'error' => $e->getMessage(),
            ]);
        }
    }

    /**
     * Build the outbound webhook request with the full safety contract applied.
     *
     * @param  array{headers: array<string, string|int>, body: string}  $deliveryPayload
     * @param  array<string, array<int, string>>  $curlOptions
     */
    protected function buildRequest(array $deliveryPayload, int $timeout, array $curlOptions): PendingRequest
    {
        $request = Http::timeout($timeout)
            ->withoutRedirecting()
            ->withHeaders($deliveryPayload['headers'])
            ->withBody($deliveryPayload['body'], 'application/json');

        if ($curlOptions !== []) {
            $request = $request->withOptions([
                'curl' => $curlOptions,
            ]);
        }

        return $request;
    }

    /**
     * Handle a failed delivery attempt.
     */
    protected function handleFailure(int $statusCode, ?string $responseBody): void
    {
        Log::warning('Webhook delivery failed', [
            'delivery_id' => $this->delivery->id,
            'attempt' => $this->delivery->attempt,
            'status_code' => $statusCode,
            'can_retry' => $this->delivery->canRetry(),
        ]);

        // Mark as failed (this also schedules retry if attempts remain)
        $this->delivery->markFailed($statusCode, $responseBody);

        // If we can retry, dispatch a new job with the appropriate delay
        $queueConnection = $this->connection ?? config('queue.default');

        if (
            ! app()->environment('testing')
            && $queueConnection !== 'sync'
            && $this->delivery->canRetry()
            && $this->delivery->next_retry_at
        ) {
            $delay = $this->delivery->next_retry_at->diffInSeconds(now());

            Log::info('Scheduling webhook retry', [
                'delivery_id' => $this->delivery->id,
                'next_attempt' => $this->delivery->attempt,
                'delay_seconds' => $delay,
                'next_retry_at' => $this->delivery->next_retry_at->toIso8601String(),
            ]);

            $freshDelivery = $this->delivery->fresh();

            // Only reschedule when the delivery still exists. If it was deleted
            // while the job was processing, there is nothing left to retry.
            if (! $freshDelivery instanceof WebhookDelivery) {
                return;
            }

            self::dispatch($freshDelivery)->delay($delay);
        }
    }

    /**
     * Handle a job failure.
     */
    public function failed(\Throwable $exception): void
    {
        Log::error('Webhook delivery job failed completely', [
            'delivery_id' => $this->delivery->id,
            'error' => $exception->getMessage(),
        ]);
    }

    /**
     * Get the tags for the job.
     *
     * @return array<string>
     */
    public function tags(): array
    {
        return [
            'webhook',
            'webhook:'.$this->delivery->webhook_endpoint_id,
            'event:'.$this->delivery->event_type,
        ];
    }

    /**
     * Redact sensitive URL parts before logging.
     */
    protected function redactUrlForLog(string $url): string
    {
        $parsed = parse_url($url);
        if ($parsed === false) {
            return '[invalid-url]';
        }

        $scheme = $parsed['scheme'] ?? 'https';
        $host = $parsed['host'] ?? '';
        $port = isset($parsed['port']) ? ':'.$parsed['port'] : '';
        $path = $parsed['path'] ?? '';

        if ($path === '') {
            $path = '/';
        }

        return "{$scheme}://{$host}{$port}{$path}";
    }
}
