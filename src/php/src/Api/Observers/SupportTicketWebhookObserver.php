<?php

declare(strict_types=1);

namespace Core\Api\Observers;

use Core\Api\Models\SupportTicket;
use Core\Api\Services\WebhookService;

class SupportTicketWebhookObserver
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function created(SupportTicket $ticket): void
    {
        $workspaceId = (int) ($ticket->workspace_id ?? 0);
        if ($workspaceId <= 0) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, 'ticket.created', [
            'ticket' => [
                'id' => $ticket->id,
                'workspace_id' => $ticket->workspace_id,
                'user_id' => $ticket->user_id,
                'subject' => $ticket->subject,
                'message' => $ticket->message,
                'status' => $ticket->status,
                'priority' => $ticket->priority,
                'metadata' => $ticket->metadata ?? [],
                'created_at' => $ticket->created_at?->toIso8601String(),
                'last_replied_at' => $ticket->last_replied_at?->toIso8601String(),
            ],
        ]);
    }
}
