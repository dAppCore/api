<?php

declare(strict_types=1);

namespace Core\Api\Observers;

use Core\Api\Models\Link;
use Core\Api\Services\WebhookService;

class LinkWebhookObserver
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function updated(Link $link): void
    {
        $workspaceId = (int) $link->workspace_id;
        if ($workspaceId <= 0 || ! $this->isClickUpdate($link)) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, 'link.clicked', [
            'link' => [
                'id' => $link->id,
                'workspace_id' => $link->workspace_id,
                'user_id' => $link->user_id,
                'name' => $link->name,
                'slug' => $link->slug,
                'short_code' => $link->short_code,
                'destination_url' => $link->destination_url,
                'click_count' => (int) ($link->click_count ?? 0),
                'last_clicked_at' => $link->last_clicked_at?->toIso8601String(),
                'metadata' => $link->metadata ?? [],
            ],
        ]);
    }

    protected function isClickUpdate(Link $link): bool
    {
        return $link->wasChanged('click_count') || $link->wasChanged('last_clicked_at');
    }
}
