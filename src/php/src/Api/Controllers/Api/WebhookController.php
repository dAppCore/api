<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Models\WebhookDelivery;
use Core\Api\Models\WebhookEndpoint;
use Core\Api\Resources\WebhookEndpointResource;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;
use Illuminate\Validation\Rule;

/**
 * Webhook management endpoints.
 */
class WebhookController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;

    /**
     * List webhooks for the current workspace.
     *
     * GET /api/v1/webhooks
     */
    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $webhooks = WebhookEndpoint::query()
            ->forWorkspace($workspace->id)
            ->latest()
            ->get();

        return response()->json([
            'workspace_id' => $workspace->id,
            'webhooks' => $webhooks->map(fn (WebhookEndpoint $webhook) => (new WebhookEndpointResource($webhook))->toArray($request))->values()->all(),
            'count' => $webhooks->count(),
        ]);
    }

    /**
     * Create a webhook endpoint.
     *
     * POST /api/v1/webhooks
     */
    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $data = $request->validate([
            'url' => ['required', 'url', 'max:2048'],
            'events' => ['required', 'array', 'min:1'],
            'events.*' => ['string', Rule::in(WebhookEndpoint::EVENTS)],
            'description' => ['sometimes', 'nullable', 'string', 'max:255'],
            'active' => ['sometimes', 'boolean'],
        ]);

        $webhook = WebhookEndpoint::createForWorkspace(
            workspaceId: $workspace->id,
            url: $data['url'],
            events: $data['events'],
            description: $data['description'] ?? null,
        );

        if (array_key_exists('active', $data) && ! (bool) $data['active']) {
            $webhook->update(['active' => false]);
        }

        return response()->json([
            'webhook' => WebhookEndpointResource::withSecret($webhook)->toArray($request),
        ], 201);
    }

    /**
     * Show a single webhook endpoint.
     *
     * GET /api/v1/webhooks/{id}
     */
    public function show(Request $request, string $id): JsonResponse
    {
        $webhook = $this->resolveWebhook($request, $id);
        if (! $webhook instanceof WebhookEndpoint) {
            return $this->notFoundResponse('Webhook');
        }

        return response()->json([
            'webhook' => (new WebhookEndpointResource($webhook))->toArray($request),
        ]);
    }

    /**
     * Update a webhook endpoint.
     *
     * PATCH /api/v1/webhooks/{id}
     */
    public function update(Request $request, string $id): JsonResponse
    {
        $webhook = $this->resolveWebhook($request, $id);
        if (! $webhook instanceof WebhookEndpoint) {
            return $this->notFoundResponse('Webhook');
        }

        $data = $request->validate([
            'url' => ['sometimes', 'required', 'url', 'max:2048'],
            'events' => ['sometimes', 'required', 'array', 'min:1'],
            'events.*' => ['string', Rule::in(WebhookEndpoint::EVENTS)],
            'description' => ['sometimes', 'nullable', 'string', 'max:255'],
            'active' => ['sometimes', 'boolean'],
        ]);

        $updates = array_intersect_key($data, array_flip(['url', 'events', 'description', 'active']));

        if ($updates !== []) {
            $webhook->update($updates);
        }

        return response()->json([
            'webhook' => (new WebhookEndpointResource($webhook->fresh()))->toArray($request),
        ]);
    }

    /**
     * Delete a webhook endpoint.
     *
     * DELETE /api/v1/webhooks/{id}
     */
    public function destroy(Request $request, string $id): JsonResponse
    {
        $webhook = $this->resolveWebhook($request, $id);
        if (! $webhook instanceof WebhookEndpoint) {
            return $this->notFoundResponse('Webhook');
        }

        $webhook->delete();

        return $this->successResponse('Webhook deleted successfully.');
    }

    /**
     * List delivery history for a webhook endpoint.
     *
     * GET /api/v1/webhooks/{id}/deliveries
     */
    public function deliveries(Request $request, string $id): JsonResponse
    {
        $webhook = $this->resolveWebhook($request, $id);
        if (! $webhook instanceof WebhookEndpoint) {
            return $this->notFoundResponse('Webhook');
        }

        $deliveries = $webhook->deliveries()
            ->latest()
            ->get();

        return response()->json([
            'webhook_id' => $webhook->id,
            'deliveries' => $deliveries->map(fn (WebhookDelivery $delivery) => $this->serializeDelivery($delivery))->values()->all(),
            'count' => $deliveries->count(),
        ]);
    }

    protected function resolveWebhook(Request $request, string $id): ?WebhookEndpoint
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return null;
        }

        return WebhookEndpoint::query()
            ->forWorkspace($workspace->id)
            ->find($id);
    }

    protected function serializeDelivery(WebhookDelivery $delivery): array
    {
        return [
            'id' => $delivery->id,
            'event_id' => $delivery->event_id,
            'event_type' => $delivery->event_type,
            'status' => $delivery->status,
            'attempt' => $delivery->attempt,
            'response_code' => $delivery->response_code,
            'response_body' => $delivery->response_body,
            'delivered_at' => $delivery->delivered_at?->toIso8601String(),
            'next_retry_at' => $delivery->next_retry_at?->toIso8601String(),
            'created_at' => $delivery->created_at?->toIso8601String(),
            'updated_at' => $delivery->updated_at?->toIso8601String(),
        ];
    }
}
