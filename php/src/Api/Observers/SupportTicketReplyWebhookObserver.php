<?php

declare(strict_types=1);

namespace Core\Api\Observers;

use Core\Api\Models\SupportTicketReply;
use Core\Api\Services\WebhookService;

class SupportTicketReplyWebhookObserver
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function created(SupportTicketReply $reply): void
    {
        $reply->loadMissing('ticket');

        $workspaceId = (int) ($reply->ticket?->workspace_id ?? 0);
        if ($workspaceId <= 0) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, 'ticket.replied', [
            'ticket' => [
                'id' => $reply->ticket?->id,
                'workspace_id' => $reply->ticket?->workspace_id,
                'status' => $reply->ticket?->status,
                'last_replied_at' => $reply->ticket?->last_replied_at?->toIso8601String(),
            ],
            'reply' => [
                'id' => $reply->id,
                'ticket_id' => $reply->ticket_id,
                'user_id' => $reply->user_id,
                'message' => $reply->message,
                'is_staff' => (bool) $reply->is_staff,
                'metadata' => $reply->metadata ?? [],
                'created_at' => $reply->created_at?->toIso8601String(),
            ],
        ]);
    }
}
