<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Core\Api\Services\WebhookSecretRotationService;
use Core\Content\Models\ContentWebhookEndpoint;
use Core\Social\Models\Webhook;

/**
 * API controller for webhook secret rotation operations.
 */
class WebhookSecretController extends Controller
{
    public function __construct(
        protected WebhookSecretRotationService $rotationService
    ) {}

    /**
     * Rotate a social webhook secret.
     */
    public function rotateSocialSecret(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $webhook = Webhook::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $webhook) {
            return response()->json(['error' => 'Webhook not found'], 404);
        }

        $validated = $request->validate([
            'grace_period_seconds' => 'nullable|integer|min:300|max:604800', // 5 min to 7 days
        ]);

        $newSecret = $this->rotationService->rotateSecret(
            $webhook,
            $validated['grace_period_seconds'] ?? null
        );

        return response()->json([
            'success' => true,
            'message' => 'Secret rotated successfully',
            'data' => [
                'secret' => $newSecret,
                'status' => $this->rotationService->getSecretStatus($webhook->fresh()),
            ],
        ]);
    }

    /**
     * Rotate a content webhook endpoint secret.
     */
    public function rotateContentSecret(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $endpoint = ContentWebhookEndpoint::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $endpoint) {
            return response()->json(['error' => 'Webhook endpoint not found'], 404);
        }

        $validated = $request->validate([
            'grace_period_seconds' => 'nullable|integer|min:300|max:604800',
        ]);

        $newSecret = $this->rotationService->rotateSecret(
            $endpoint,
            $validated['grace_period_seconds'] ?? null
        );

        return response()->json([
            'success' => true,
            'message' => 'Secret rotated successfully',
            'data' => [
                'secret' => $newSecret,
                'status' => $this->rotationService->getSecretStatus($endpoint->fresh()),
            ],
        ]);
    }

    /**
     * Get secret rotation status for a social webhook.
     */
    public function socialSecretStatus(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $webhook = Webhook::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $webhook) {
            return response()->json(['error' => 'Webhook not found'], 404);
        }

        return response()->json([
            'data' => $this->rotationService->getSecretStatus($webhook),
        ]);
    }

    /**
     * Get secret rotation status for a content webhook endpoint.
     */
    public function contentSecretStatus(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $endpoint = ContentWebhookEndpoint::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $endpoint) {
            return response()->json(['error' => 'Webhook endpoint not found'], 404);
        }

        return response()->json([
            'data' => $this->rotationService->getSecretStatus($endpoint),
        ]);
    }

    /**
     * Invalidate the previous secret for a social webhook.
     */
    public function invalidateSocialPreviousSecret(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $webhook = Webhook::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $webhook) {
            return response()->json(['error' => 'Webhook not found'], 404);
        }

        $this->rotationService->invalidatePreviousSecret($webhook);

        return response()->json([
            'success' => true,
            'message' => 'Previous secret invalidated',
        ]);
    }

    /**
     * Invalidate the previous secret for a content webhook endpoint.
     */
    public function invalidateContentPreviousSecret(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $endpoint = ContentWebhookEndpoint::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $endpoint) {
            return response()->json(['error' => 'Webhook endpoint not found'], 404);
        }

        $this->rotationService->invalidatePreviousSecret($endpoint);

        return response()->json([
            'success' => true,
            'message' => 'Previous secret invalidated',
        ]);
    }

    /**
     * Update the grace period for a social webhook.
     */
    public function updateSocialGracePeriod(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $webhook = Webhook::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $webhook) {
            return response()->json(['error' => 'Webhook not found'], 404);
        }

        $validated = $request->validate([
            'grace_period_seconds' => 'required|integer|min:300|max:604800',
        ]);

        $this->rotationService->updateGracePeriod($webhook, $validated['grace_period_seconds']);

        return response()->json([
            'success' => true,
            'message' => 'Grace period updated',
            'data' => [
                'grace_period_seconds' => $webhook->fresh()->grace_period_seconds,
            ],
        ]);
    }

    /**
     * Update the grace period for a content webhook endpoint.
     */
    public function updateContentGracePeriod(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return response()->json(['error' => 'Workspace not found'], 404);
        }

        $endpoint = ContentWebhookEndpoint::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $endpoint) {
            return response()->json(['error' => 'Webhook endpoint not found'], 404);
        }

        $validated = $request->validate([
            'grace_period_seconds' => 'required|integer|min:300|max:604800',
        ]);

        $this->rotationService->updateGracePeriod($endpoint, $validated['grace_period_seconds']);

        return response()->json([
            'success' => true,
            'message' => 'Grace period updated',
            'data' => [
                'grace_period_seconds' => $endpoint->fresh()->grace_period_seconds,
            ],
        ]);
    }
}
