<?php

declare(strict_types=1);

namespace Core\Api\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\Log;
use Core\Api\Services\WebhookSecretRotationService;
use Core\Content\Models\ContentWebhookEndpoint;
use Core\Social\Models\Webhook;

/**
 * Clean up expired webhook secret grace periods.
 *
 * Removes previous_secret values from webhooks where the grace period has expired.
 * This command should be run periodically (e.g., daily via scheduler).
 */
class CleanupExpiredSecrets extends Command
{
    protected $signature = 'webhook:cleanup-secrets
                            {--dry-run : Show what would be cleaned up without making changes}
                            {--model= : Only process a specific model (social, content)}';

    protected $description = 'Clean up expired webhook secret grace periods';

    /**
     * Webhook model classes to process.
     *
     * @var array<string, string>
     */
    protected array $webhookModels = [
        'social' => Webhook::class,
        'content' => ContentWebhookEndpoint::class,
    ];

    public function handle(WebhookSecretRotationService $service): int
    {
        $dryRun = $this->option('dry-run');
        $modelFilter = $this->option('model');

        $this->info('Starting webhook secret cleanup...');

        if ($dryRun) {
            $this->warn('DRY RUN MODE - No data will be modified');
        }

        $startTime = microtime(true);
        $totalCleaned = 0;

        $modelsToProcess = $this->getModelsToProcess($modelFilter);

        if (empty($modelsToProcess)) {
            $this->error('No valid models to process.');

            return Command::FAILURE;
        }

        foreach ($modelsToProcess as $name => $modelClass) {
            if (! class_exists($modelClass)) {
                $this->warn("Model class {$modelClass} not found, skipping...");

                continue;
            }

            $this->info("Processing {$name} webhooks...");

            if ($dryRun) {
                $count = $this->countExpiredGracePeriods($modelClass, $service);
                $this->line("  Would clean up: {$count} webhook(s)");
                $totalCleaned += $count;
            } else {
                $count = $service->cleanupExpiredGracePeriods($modelClass);
                $this->line("  Cleaned up: {$count} webhook(s)");
                $totalCleaned += $count;
            }
        }

        $elapsed = round(microtime(true) - $startTime, 2);

        $this->newLine();
        $this->info('Cleanup Summary:');
        $this->line("  Total cleaned: {$totalCleaned} webhook(s)");
        $this->line("  Time elapsed: {$elapsed}s");

        if (! $dryRun && $totalCleaned > 0) {
            Log::info('Webhook secret cleanup completed', [
                'total_cleaned' => $totalCleaned,
                'elapsed_seconds' => $elapsed,
            ]);
        }

        $this->info('Webhook secret cleanup complete.');

        return Command::SUCCESS;
    }

    /**
     * Get the webhook models to process based on the filter.
     *
     * @return array<string, string>
     */
    protected function getModelsToProcess(?string $filter): array
    {
        if ($filter === null) {
            return $this->webhookModels;
        }

        $filter = strtolower($filter);

        if (! isset($this->webhookModels[$filter])) {
            $this->error("Invalid model filter: {$filter}");
            $this->line('Available models: '.implode(', ', array_keys($this->webhookModels)));

            return [];
        }

        return [$filter => $this->webhookModels[$filter]];
    }

    /**
     * Count webhooks with expired grace periods (for dry run).
     */
    protected function countExpiredGracePeriods(string $modelClass, WebhookSecretRotationService $service): int
    {
        $count = 0;

        $modelClass::query()
            ->whereNotNull('previous_secret')
            ->whereNotNull('secret_rotated_at')
            ->chunkById(100, function ($webhooks) use ($service, &$count) {
                foreach ($webhooks as $webhook) {
                    if (! $service->isInGracePeriod($webhook)) {
                        $count++;
                    }
                }
            });

        return $count;
    }
}
