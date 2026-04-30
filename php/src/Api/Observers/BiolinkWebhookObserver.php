<?php

declare(strict_types=1);

namespace Core\Api\Observers;

use Core\Api\Models\Biolink;
use Core\Api\Services\WebhookService;

class BiolinkWebhookObserver
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function created(Biolink $biolink): void
    {
        $workspaceId = (int) $biolink->workspace_id;
        if ($workspaceId <= 0) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, 'biolink.created', [
            'biolink' => [
                'id' => $biolink->id,
                'workspace_id' => $biolink->workspace_id,
                'user_id' => $biolink->user_id,
                'title' => $biolink->title,
                'slug' => $biolink->slug,
                'url' => $biolink->url,
                'description' => $biolink->description,
                'is_published' => (bool) $biolink->is_published,
                'metadata' => $biolink->metadata ?? [],
                'created_at' => $biolink->created_at?->toIso8601String(),
            ],
        ]);
    }
}
