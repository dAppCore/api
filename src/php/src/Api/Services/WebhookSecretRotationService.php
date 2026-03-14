<?php

declare(strict_types=1);

namespace Core\Api\Services;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

/**
 * Service for managing webhook secret rotation with grace periods.
 *
 * Provides functionality for:
 * - Rotating webhook secrets while preserving the old secret during a grace period
 * - Verifying signatures against both current and previous secrets
 * - Cleaning up expired grace periods
 * - Getting the current rotation status
 */
class WebhookSecretRotationService
{
    /**
     * Default grace period in seconds (24 hours).
     */
    public const DEFAULT_GRACE_PERIOD = 86400;

    /**
     * Minimum grace period in seconds (5 minutes).
     */
    public const MIN_GRACE_PERIOD = 300;

    /**
     * Maximum grace period in seconds (7 days).
     */
    public const MAX_GRACE_PERIOD = 604800;

    /**
     * Rotate the secret for a webhook model.
     *
     * Generates a new secret, stores the old one for the grace period,
     * and records the rotation timestamp.
     *
     * @param  Model  $webhook  The webhook model (must have secret, previous_secret, secret_rotated_at fields)
     * @param  int|null  $gracePeriodSeconds  Custom grace period (uses model's default if null)
     * @return string The new secret
     */
    public function rotateSecret(Model $webhook, ?int $gracePeriodSeconds = null): string
    {
        $newSecret = Str::random(64);
        $currentSecret = $webhook->secret;

        // Determine grace period
        $gracePeriod = $gracePeriodSeconds ?? $webhook->grace_period_seconds ?? self::DEFAULT_GRACE_PERIOD;
        $gracePeriod = max(self::MIN_GRACE_PERIOD, min(self::MAX_GRACE_PERIOD, $gracePeriod));

        DB::transaction(function () use ($webhook, $currentSecret, $newSecret, $gracePeriod) {
            $webhook->update([
                'previous_secret' => $currentSecret,
                'secret' => $newSecret,
                'secret_rotated_at' => now(),
                'grace_period_seconds' => $gracePeriod,
            ]);
        });

        Log::info('Webhook secret rotated', [
            'webhook_id' => $webhook->id,
            'webhook_type' => class_basename($webhook),
            'grace_period_seconds' => $gracePeriod,
        ]);

        return $newSecret;
    }

    /**
     * Verify a signature against both current and previous secrets during grace period.
     *
     * @param  Model  $webhook  The webhook model
     * @param  string  $payload  The raw payload to verify
     * @param  string|null  $signature  The provided signature
     * @param  string  $algorithm  Hash algorithm (default: sha256)
     * @return array{valid: bool, used_previous: bool, message: string}
     */
    public function verifySignature(
        Model $webhook,
        string $payload,
        ?string $signature,
        string $algorithm = 'sha256'
    ): array {
        // If no secret configured, skip verification
        if (empty($webhook->secret)) {
            return [
                'valid' => true,
                'used_previous' => false,
                'message' => 'No secret configured, verification skipped',
            ];
        }

        // Signature required when secret is set
        if (empty($signature)) {
            return [
                'valid' => false,
                'used_previous' => false,
                'message' => 'Signature required but not provided',
            ];
        }

        // Normalise signature (strip prefix like sha256= if present)
        $signature = $this->normaliseSignature($signature, $algorithm);

        // Check against current secret
        $expectedSignature = hash_hmac($algorithm, $payload, $webhook->secret);
        if (hash_equals($expectedSignature, $signature)) {
            return [
                'valid' => true,
                'used_previous' => false,
                'message' => 'Signature verified with current secret',
            ];
        }

        // Check against previous secret if in grace period
        if ($this->isInGracePeriod($webhook) && ! empty($webhook->previous_secret)) {
            $previousExpectedSignature = hash_hmac($algorithm, $payload, $webhook->previous_secret);
            if (hash_equals($previousExpectedSignature, $signature)) {
                return [
                    'valid' => true,
                    'used_previous' => true,
                    'message' => 'Signature verified with previous secret (grace period)',
                ];
            }
        }

        return [
            'valid' => false,
            'used_previous' => false,
            'message' => 'Signature verification failed',
        ];
    }

    /**
     * Check if the webhook is currently in a grace period.
     */
    public function isInGracePeriod(Model $webhook): bool
    {
        if (empty($webhook->secret_rotated_at)) {
            return false;
        }

        $rotatedAt = Carbon::parse($webhook->secret_rotated_at);
        $gracePeriodSeconds = $webhook->grace_period_seconds ?? self::DEFAULT_GRACE_PERIOD;
        $graceEndsAt = $rotatedAt->copy()->addSeconds($gracePeriodSeconds);

        return now()->isBefore($graceEndsAt);
    }

    /**
     * Get the secret rotation status for a webhook.
     *
     * @return array{
     *     has_previous_secret: bool,
     *     in_grace_period: bool,
     *     grace_period_seconds: int,
     *     rotated_at: ?string,
     *     grace_ends_at: ?string,
     *     time_remaining_seconds: ?int,
     *     time_remaining_human: ?string
     * }
     */
    public function getSecretStatus(Model $webhook): array
    {
        $hasPreviousSecret = ! empty($webhook->previous_secret);
        $inGracePeriod = $this->isInGracePeriod($webhook);
        $gracePeriodSeconds = $webhook->grace_period_seconds ?? self::DEFAULT_GRACE_PERIOD;

        $rotatedAt = $webhook->secret_rotated_at ? Carbon::parse($webhook->secret_rotated_at) : null;
        $graceEndsAt = $rotatedAt ? $rotatedAt->copy()->addSeconds($gracePeriodSeconds) : null;
        $timeRemaining = ($inGracePeriod && $graceEndsAt) ? now()->diffInSeconds($graceEndsAt, false) : null;
        $timeRemainingHuman = $timeRemaining > 0 ? $this->humanReadableTime($timeRemaining) : null;

        return [
            'has_previous_secret' => $hasPreviousSecret,
            'in_grace_period' => $inGracePeriod,
            'grace_period_seconds' => $gracePeriodSeconds,
            'rotated_at' => $rotatedAt?->toIso8601String(),
            'grace_ends_at' => $graceEndsAt?->toIso8601String(),
            'time_remaining_seconds' => $timeRemaining > 0 ? (int) $timeRemaining : null,
            'time_remaining_human' => $timeRemainingHuman,
        ];
    }

    /**
     * Immediately invalidate the previous secret.
     *
     * Use this to end the grace period early (e.g., if the old secret was compromised).
     */
    public function invalidatePreviousSecret(Model $webhook): void
    {
        $webhook->update([
            'previous_secret' => null,
            'secret_rotated_at' => null,
        ]);

        Log::info('Webhook previous secret invalidated', [
            'webhook_id' => $webhook->id,
            'webhook_type' => class_basename($webhook),
        ]);
    }

    /**
     * Clean up expired grace periods for a specific model class.
     *
     * @param  string  $modelClass  The webhook model class to clean up
     * @return int Number of webhooks cleaned up
     */
    public function cleanupExpiredGracePeriods(string $modelClass): int
    {
        $count = 0;

        $modelClass::query()
            ->whereNotNull('previous_secret')
            ->whereNotNull('secret_rotated_at')
            ->chunkById(100, function ($webhooks) use (&$count) {
                foreach ($webhooks as $webhook) {
                    if (! $this->isInGracePeriod($webhook)) {
                        $webhook->update([
                            'previous_secret' => null,
                            'secret_rotated_at' => null,
                        ]);
                        $count++;
                    }
                }
            });

        if ($count > 0) {
            Log::info('Cleaned up expired webhook secret grace periods', [
                'model_class' => $modelClass,
                'count' => $count,
            ]);
        }

        return $count;
    }

    /**
     * Update the grace period duration for a webhook.
     */
    public function updateGracePeriod(Model $webhook, int $gracePeriodSeconds): void
    {
        $gracePeriodSeconds = max(self::MIN_GRACE_PERIOD, min(self::MAX_GRACE_PERIOD, $gracePeriodSeconds));

        $webhook->update([
            'grace_period_seconds' => $gracePeriodSeconds,
        ]);
    }

    /**
     * Normalise a signature by removing common prefixes.
     */
    protected function normaliseSignature(string $signature, string $algorithm): string
    {
        // Handle sha256=... format (GitHub, WordPress)
        $prefix = $algorithm.'=';
        if (str_starts_with($signature, $prefix)) {
            return substr($signature, strlen($prefix));
        }

        return $signature;
    }

    /**
     * Convert seconds to human-readable time string.
     */
    protected function humanReadableTime(int $seconds): string
    {
        if ($seconds < 60) {
            return $seconds.' second'.($seconds !== 1 ? 's' : '');
        }

        if ($seconds < 3600) {
            $minutes = (int) floor($seconds / 60);

            return $minutes.' minute'.($minutes !== 1 ? 's' : '');
        }

        if ($seconds < 86400) {
            $hours = (int) floor($seconds / 3600);
            $minutes = (int) floor(($seconds % 3600) / 60);

            $result = $hours.' hour'.($hours !== 1 ? 's' : '');
            if ($minutes > 0) {
                $result .= ' '.$minutes.' minute'.($minutes !== 1 ? 's' : '');
            }

            return $result;
        }

        $days = (int) floor($seconds / 86400);
        $hours = (int) floor(($seconds % 86400) / 3600);

        $result = $days.' day'.($days !== 1 ? 's' : '');
        if ($hours > 0) {
            $result .= ' '.$hours.' hour'.($hours !== 1 ? 's' : '');
        }

        return $result;
    }
}
