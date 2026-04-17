<?php

declare(strict_types=1);

namespace Core\Api\Services;

use Core\Api\Jobs\DeliverWebhookJob;
use Core\Api\Models\WebhookDelivery;
use Core\Api\Models\WebhookEndpoint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

/**
 * Webhook Service - dispatches events to registered webhook endpoints.
 *
 * Finds all active endpoints subscribed to an event type and queues
 * delivery jobs with proper payload formatting and signature generation.
 */
class WebhookService
{
    /**
     * Dispatch an event to all subscribed webhook endpoints.
     *
     * @param  int  $workspaceId  The workspace that owns the webhooks
     * @param  string  $eventType  The event type (e.g., 'bio.created')
     * @param  array  $data  The event payload data
     * @return array<WebhookDelivery> The created delivery records
     */
    public function dispatch(int $workspaceId, string $eventType, array $data): array
    {
        try {
            // Find all active endpoints for this workspace that subscribe to this event
            $endpoints = WebhookEndpoint::query()
                ->forWorkspace($workspaceId)
                ->active()
                ->forEvent($eventType)
                ->get();

            if ($endpoints->isEmpty()) {
                Log::debug('No webhook endpoints found for event', [
                    'workspace_id' => $workspaceId,
                    'event_type' => $eventType,
                ]);

                return [];
            }

            $deliveries = [];

            foreach ($endpoints as $endpoint) {
                try {
                    // Keep each endpoint isolated so one failure does not stop
                    // the rest of the workspace's subscriptions from receiving
                    // the event. Return the delivery from the transaction so
                    // rollbacks cannot leak phantom records into the result.
                    $delivery = DB::transaction(function () use (
                        $endpoint,
                        $eventType,
                        $data,
                        $workspaceId
                    ) {
                        $delivery = WebhookDelivery::createForEvent(
                            $endpoint,
                            $eventType,
                            $data,
                            $workspaceId
                        );

                        $this->queueDelivery($delivery);

                        Log::info('Webhook delivery queued', [
                            'delivery_id' => $delivery->id,
                            'endpoint_id' => $endpoint->id,
                            'event_type' => $eventType,
                        ]);

                        return $delivery;
                    });

                    if ($delivery instanceof WebhookDelivery) {
                        $deliveries[] = $delivery;
                    }
                } catch (\Throwable $endpointException) {
                    report($endpointException);
                    Log::warning('Webhook delivery failed for endpoint', [
                        'workspace_id' => $workspaceId,
                        'endpoint_id' => $endpoint->id,
                        'event_type' => $eventType,
                        'error_type' => $endpointException::class,
                        'error' => $endpointException->getMessage(),
                    ]);
                }
            }

            return $deliveries;
        } catch (\Throwable $exception) {
            report($exception);
            Log::error('Webhook dispatch failed', [
                'workspace_id' => $workspaceId,
                'event_type' => $eventType,
                'error_type' => $exception::class,
                'error' => $exception->getMessage(),
            ]);

            return [];
        }
    }

    /**
     * Retry a specific failed delivery.
     *
     * @return bool True if retry was queued, false if not eligible
     */
    public function retry(WebhookDelivery $delivery): bool
    {
        if (! $delivery->canRetry()) {
            return false;
        }

        try {
            DB::transaction(function () use ($delivery) {
                // Reset status for manual retry but preserve attempt history
                $delivery->update([
                    'status' => WebhookDelivery::STATUS_PENDING,
                    'next_retry_at' => null,
                ]);

                $this->queueDelivery($delivery);

                Log::info('Manual webhook retry queued', [
                    'delivery_id' => $delivery->id,
                    'attempt' => $delivery->attempt,
                ]);
            });

            return true;
        } catch (\Throwable $exception) {
            report($exception);
            Log::error('Manual webhook retry failed', [
                'delivery_id' => $delivery->id,
                'error_type' => $exception::class,
                'error' => $exception->getMessage(),
            ]);

            return false;
        }
    }

    /**
     * Process all pending and retryable deliveries.
     *
     * This method is typically called by a scheduled command.
     *
     * @return int Number of deliveries queued
     */
    public function processQueue(): int
    {
        $count = 0;

        // Process deliveries one at a time with row locking to prevent race conditions
        $deliveryIds = WebhookDelivery::query()
            ->needsDelivery()
            ->limit(100)
            ->pluck('id');

        foreach ($deliveryIds as $deliveryId) {
            try {
                DB::transaction(function () use ($deliveryId, &$count) {
                // Lock the row for update to prevent concurrent processing
                $delivery = WebhookDelivery::query()
                    ->with('endpoint')
                    ->where('id', $deliveryId)
                    ->lockForUpdate()
                    ->first();

                if (! $delivery) {
                    return;
                }

                // Skip if already being processed (status changed since initial query)
                if (! in_array($delivery->status, [WebhookDelivery::STATUS_PENDING, WebhookDelivery::STATUS_RETRYING])) {
                    return;
                }

                // Handle inactive endpoints by cancelling the delivery
                if (! $delivery->endpoint?->shouldReceive($delivery->event_type)) {
                    $delivery->update(['status' => WebhookDelivery::STATUS_CANCELLED]);

                    return;
                }

                // Mark as queued to prevent duplicate processing
                $delivery->update(['status' => WebhookDelivery::STATUS_QUEUED]);

                $this->queueDelivery($delivery);
                $count++;
            });
            } catch (\Throwable $exception) {
                report($exception);
                Log::warning('Webhook queue processing failed for delivery', [
                    'delivery_id' => $deliveryId,
                    'error_type' => $exception::class,
                    'error' => $exception->getMessage(),
                ]);
            }
        }

        if ($count > 0) {
            Log::info('Processed webhook queue', ['count' => $count]);
        }

        return $count;
    }

    /**
     * Queue a webhook delivery job after the surrounding transaction commits.
     */
    protected function queueDelivery(WebhookDelivery $delivery): void
    {
        DeliverWebhookJob::dispatch($delivery)->afterCommit();
    }

    /**
     * Get delivery statistics for a workspace.
     */
    public function getStats(int $workspaceId): array
    {
        $endpointIds = WebhookEndpoint::query()
            ->forWorkspace($workspaceId)
            ->pluck('id');

        if ($endpointIds->isEmpty()) {
            return [
                'total' => 0,
                'pending' => 0,
                'success' => 0,
                'failed' => 0,
                'retrying' => 0,
            ];
        }

        $deliveries = WebhookDelivery::query()
            ->whereIn('webhook_endpoint_id', $endpointIds);

        return [
            'total' => (clone $deliveries)->count(),
            'pending' => (clone $deliveries)->where('status', WebhookDelivery::STATUS_PENDING)->count(),
            'success' => (clone $deliveries)->where('status', WebhookDelivery::STATUS_SUCCESS)->count(),
            'failed' => (clone $deliveries)->where('status', WebhookDelivery::STATUS_FAILED)->count(),
            'retrying' => (clone $deliveries)->where('status', WebhookDelivery::STATUS_RETRYING)->count(),
        ];
    }
}
