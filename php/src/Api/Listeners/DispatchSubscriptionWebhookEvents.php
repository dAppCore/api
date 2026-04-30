<?php

declare(strict_types=1);

namespace Core\Api\Listeners;

use Core\Api\Services\WebhookService;
use Core\Mod\Commerce\Events\SubscriptionUpdated;
use Core\Mod\Commerce\Models\Subscription;

class DispatchSubscriptionWebhookEvents
{
    public function __construct(
        protected WebhookService $webhookService,
    ) {
    }

    public function handle(SubscriptionUpdated $event): void
    {
        $subscription = $event->subscription;
        $workspaceId = (int) ($subscription->workspace_id ?? 0);

        if ($workspaceId <= 0) {
            return;
        }

        if ($this->isCancellationUpdate($subscription)) {
            $this->webhookService->dispatch($workspaceId, 'subscription.cancelled', [
                'subscription' => $this->subscriptionPayload($subscription),
            ]);

            return;
        }

        if (! $this->isPlanChangeUpdate($subscription)) {
            return;
        }

        $this->webhookService->dispatch($workspaceId, 'subscription.changed', [
            'subscription' => $this->subscriptionPayload($subscription),
        ]);
    }

    protected function isCancellationUpdate(Subscription $subscription): bool
    {
        return ($subscription->wasChanged('cancelled_at') && $subscription->cancelled_at !== null)
            || ($subscription->wasChanged('status') && $subscription->status === 'cancelled');
    }

    protected function isPlanChangeUpdate(Subscription $subscription): bool
    {
        if ($subscription->wasChanged('workspace_package_id') || $subscription->wasChanged('gateway_price_id')) {
            return true;
        }

        if (! $subscription->wasChanged('metadata')) {
            return false;
        }

        $metadata = $subscription->metadata ?? [];

        return is_array($metadata)
            && (array_key_exists('plan_change', $metadata) || array_key_exists('pending_plan_change', $metadata));
    }

    protected function subscriptionPayload(Subscription $subscription): array
    {
        $subscription->loadMissing('workspacePackage.package');

        $metadata = is_array($subscription->metadata) ? $subscription->metadata : [];
        $currentPackage = $subscription->workspacePackage?->package;
        $planChange = is_array($metadata['plan_change'] ?? null) ? $metadata['plan_change'] : [];
        $pendingPlanChange = is_array($metadata['pending_plan_change'] ?? null) ? $metadata['pending_plan_change'] : [];

        return [
            'id' => $subscription->id,
            'workspace_id' => $subscription->workspace_id,
            'workspace_package_id' => $subscription->workspace_package_id,
            'status' => $subscription->status,
            'billing_cycle' => $subscription->billing_cycle,
            'cancel_at_period_end' => (bool) $subscription->cancel_at_period_end,
            'cancelled_at' => $subscription->cancelled_at?->toIso8601String(),
            'current_period_start' => $subscription->current_period_start?->toIso8601String(),
            'current_period_end' => $subscription->current_period_end?->toIso8601String(),
            'current_package' => [
                'id' => $currentPackage?->id,
                'code' => $currentPackage?->code,
                'name' => $currentPackage?->name,
            ],
            'plan_change' => [
                'from' => $planChange['from'] ?? null,
                'to' => $planChange['to'] ?? $pendingPlanChange['to_package_code'] ?? null,
                'scheduled_for' => $pendingPlanChange['scheduled_for'] ?? null,
                'changed_at' => $planChange['changed_at'] ?? null,
            ],
        ];
    }
}
