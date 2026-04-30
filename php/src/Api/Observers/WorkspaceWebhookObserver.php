<?php

declare(strict_types=1);

namespace Core\Api\Observers;

use Core\Api\Services\WebhookService;
use Core\Tenant\Models\Workspace;

/**
 * Dispatch workspace lifecycle webhooks from the canonical model events.
 */
class WorkspaceWebhookObserver
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function created(Workspace $workspace): void
    {
        $this->dispatch($workspace, 'workspace.created');
    }

    public function updated(Workspace $workspace): void
    {
        $this->dispatch($workspace, 'workspace.updated');
    }

    public function deleted(Workspace $workspace): void
    {
        $this->dispatch($workspace, 'workspace.deleted');
    }

    protected function dispatch(Workspace $workspace, string $event): void
    {
        $workspaceId = (int) ($workspace->id ?? 0);
        if ($workspaceId <= 0) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, $event, [
            'workspace' => [
                'id' => $workspaceId,
                'name' => $workspace->name,
                'slug' => $workspace->slug,
                'type' => $workspace->type,
                'is_active' => $workspace->is_active,
                'created_at' => $workspace->created_at?->toIso8601String(),
                'updated_at' => $workspace->updated_at?->toIso8601String(),
            ],
        ]);
    }
}
