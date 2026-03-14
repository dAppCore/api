<?php

declare(strict_types=1);

namespace Mod\Api\Console\Commands;

use Illuminate\Console\Command;
use Mod\Api\Services\ApiKeyService;

/**
 * Clean up API keys with expired grace periods.
 *
 * When an API key is rotated, the old key enters a grace period where
 * both keys are valid. This command revokes keys whose grace period
 * has ended.
 */
class CleanupExpiredGracePeriods extends Command
{
    /**
     * The name and signature of the console command.
     */
    protected $signature = 'api:cleanup-grace-periods
                            {--dry-run : Show what would be revoked without actually revoking}';

    /**
     * The console command description.
     */
    protected $description = 'Revoke API keys with expired grace periods after rotation';

    /**
     * Execute the console command.
     */
    public function handle(ApiKeyService $service): int
    {
        $dryRun = $this->option('dry-run');

        if ($dryRun) {
            $this->warn('DRY RUN MODE - No keys will be revoked');
            $this->newLine();

            // Count keys that would be cleaned up
            $count = \Mod\Api\Models\ApiKey::gracePeriodExpired()
                ->whereNull('deleted_at')
                ->count();

            if ($count === 0) {
                $this->info('No API keys with expired grace periods found.');
            } else {
                $this->info("Would revoke {$count} API key(s) with expired grace periods.");
            }

            return Command::SUCCESS;
        }

        $this->info('Cleaning up API keys with expired grace periods...');

        $count = $service->cleanupExpiredGracePeriods();

        if ($count === 0) {
            $this->info('No API keys with expired grace periods found.');
        } else {
            $this->info("Revoked {$count} API key(s) with expired grace periods.");
        }

        return Command::SUCCESS;
    }
}
