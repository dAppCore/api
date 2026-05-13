<?php

declare(strict_types=1);

namespace Core\Api\Console\Commands;

use Core\Api\Models\ApiKey;
use Core\Api\Notifications\HighApiUsageNotification;
use Core\Api\RateLimit\RateLimitService;
use Core\Tenant\Models\Workspace;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Cache;

/**
 * Check API usage levels and send alerts when approaching limits.
 *
 * Notifies workspace owners when:
 * - 80% of rate limit is used (warning)
 * - 95% of rate limit is used (critical)
 *
 * Uses cache to prevent duplicate notifications within a cooldown period.
 */
class CheckApiUsageAlerts extends Command
{
    /**
     * Cache key prefix for notification cooldowns.
     */
    protected const CACHE_PREFIX = 'api_usage_alert:';

    /**
     * Default hours between notifications of the same level.
     */
    protected const DEFAULT_COOLDOWN_HOURS = 6;

    /**
     * The name and signature of the console command.
     */
    protected $signature = 'api:check-usage-alerts
                            {--dry-run : Show what alerts would be sent without sending}
                            {--workspace= : Check a specific workspace by ID}';

    /**
     * The console command description.
     */
    protected $description = 'Check API usage levels and send alerts when approaching rate limits';

    /**
     * Alert thresholds (percentage of limit).
     * Loaded from config in constructor.
     */
    protected array $thresholds = [];

    /**
     * Cooldown hours between notifications.
     */
    protected int $cooldownHours;

    /**
     * Execute the console command.
     */
    public function handle(RateLimitService $rateLimitService): int
    {
        // Check if alerts are enabled
        if (! config('api.alerts.enabled', true)) {
            $this->info('API usage alerts are disabled.');

            return Command::SUCCESS;
        }

        // Load thresholds from config (sorted by severity, critical first)
        $this->thresholds = config('api.alerts.thresholds', [
            'critical' => 95,
            'warning' => 80,
        ]);
        arsort($this->thresholds);

        $this->cooldownHours = config('api.alerts.cooldown_hours', self::DEFAULT_COOLDOWN_HOURS);

        $dryRun = $this->option('dry-run');
        $specificWorkspace = $this->option('workspace');

        if ($dryRun) {
            $this->warn('DRY RUN MODE - No notifications will be sent');
            $this->newLine();
        }

        // Get workspaces with active API keys
        $query = Workspace::whereHas('apiKeys', function ($q) {
            $q->active();
        });

        if ($specificWorkspace) {
            $query->where('id', $specificWorkspace);
        }

        $workspaces = $query->get();

        if ($workspaces->isEmpty()) {
            $this->info('No workspaces with active API keys found.');

            return Command::SUCCESS;
        }

        $alertsSent = 0;
        $alertsSkipped = 0;

        foreach ($workspaces as $workspace) {
            $result = $this->checkWorkspaceUsage($workspace, $rateLimitService, $dryRun);
            $alertsSent += $result['sent'];
            $alertsSkipped += $result['skipped'];
        }

        $this->newLine();
        $this->info("Alerts sent: {$alertsSent}");
        $this->info("Alerts skipped (cooldown): {$alertsSkipped}");

        return Command::SUCCESS;
    }

    /**
     * Check usage for a workspace and send alerts if needed.
     *
     * @return array{sent: int, skipped: int}
     */
    protected function checkWorkspaceUsage(
        Workspace $workspace,
        RateLimitService $rateLimitService,
        bool $dryRun
    ): array {
        $sent = 0;
        $skipped = 0;

        // Get rate limit config for this workspace's tier
        $tier = $this->getWorkspaceTier($workspace);
        $limitConfig = $this->getTierLimitConfig($tier);

        if (! $limitConfig) {
            return ['sent' => 0, 'skipped' => 0];
        }

        // Check usage for each active API key
        $apiKeys = $workspace->apiKeys()->active()->get();

        foreach ($apiKeys as $apiKey) {
            $key = $rateLimitService->buildApiKeyKey($apiKey->id);
            $attempts = $rateLimitService->attempts($key, $limitConfig['window']);
            $limit = (int) floor($limitConfig['limit'] * ($limitConfig['burst'] ?? 1.0));

            if ($limit === 0) {
                continue;
            }

            $percentage = ($attempts / $limit) * 100;

            // Check thresholds (critical first, then warning)
            foreach ($this->thresholds as $level => $threshold) {
                if ($percentage >= $threshold) {
                    $cacheKey = $this->getCacheKey($workspace->id, $apiKey->id, $level);

                    if (Cache::has($cacheKey)) {
                        $this->line("  [SKIP] {$workspace->name} - Key {$apiKey->prefix}: {$level} (cooldown)");
                        $skipped++;

                        break; // Don't check lower thresholds
                    }

                    $this->line("  [ALERT] {$workspace->name} - Key {$apiKey->prefix}: {$level} ({$percentage}%)");

                    if (! $dryRun) {
                        $this->sendAlert($workspace, $apiKey, $level, $attempts, $limit, $limitConfig);
                        Cache::put($cacheKey, true, now()->addHours($this->cooldownHours));
                    }

                    $sent++;

                    break; // Only send one alert per key (highest severity)
                }
            }
        }

        return ['sent' => $sent, 'skipped' => $skipped];
    }

    /**
     * Send alert notification to workspace owner.
     */
    protected function sendAlert(
        Workspace $workspace,
        ApiKey $apiKey,
        string $level,
        int $currentUsage,
        int $limit,
        array $limitConfig
    ): void {
        $owner = $workspace->owner();

        if (! $owner) {
            $this->warn("  No owner found for workspace {$workspace->name}");

            return;
        }

        $period = $this->formatPeriod($limitConfig['window']);

        $owner->notify(new HighApiUsageNotification(
            workspace: $workspace,
            level: $level,
            currentUsage: $currentUsage,
            limit: $limit,
            period: $period,
        ));
    }

    /**
     * Get workspace tier for rate limiting.
     */
    protected function getWorkspaceTier(Workspace $workspace): string
    {
        // Check for active package
        $package = $workspace->workspacePackages()
            ->active()
            ->with('package')
            ->first();

        return $package?->package?->slug ?? 'free';
    }

    /**
     * Get rate limit config for a tier.
     *
     * @return array{limit: int, window: int, burst: float}|null
     */
    protected function getTierLimitConfig(string $tier): ?array
    {
        $config = config("api.rate_limits.tiers.{$tier}");

        if (! $config) {
            $config = config('api.rate_limits.tiers.free');
        }

        if (! $config) {
            $config = config('api.rate_limits.authenticated');
        }

        if (! $config) {
            return null;
        }

        return [
            'limit' => $config['limit'] ?? $config['requests'] ?? 60,
            'window' => $config['window'] ?? (($config['per_minutes'] ?? 1) * 60),
            'burst' => $config['burst'] ?? 1.0,
        ];
    }

    /**
     * Format window period for display.
     */
    protected function formatPeriod(int $seconds): string
    {
        if ($seconds < 60) {
            return "{$seconds} seconds";
        }

        $minutes = $seconds / 60;

        if ($minutes === 1.0) {
            return 'minute';
        }

        if ($minutes < 60) {
            return "{$minutes} minutes";
        }

        $hours = $minutes / 60;

        if ($hours === 1.0) {
            return 'hour';
        }

        return "{$hours} hours";
    }

    /**
     * Get cache key for notification cooldown.
     */
    protected function getCacheKey(int $workspaceId, int $apiKeyId, string $level): string
    {
        return self::CACHE_PREFIX."{$workspaceId}:{$apiKeyId}:{$level}";
    }
}
