<?php

declare(strict_types=1);

namespace Mod\Api\Services;

use Carbon\Carbon;
use Mod\Api\Models\ApiUsage;
use Mod\Api\Models\ApiUsageDaily;

/**
 * API Usage Service - tracks and reports API usage metrics.
 *
 * Provides methods for recording API calls and generating reports.
 */
class ApiUsageService
{
    /**
     * Record an API request.
     */
    public function record(
        int $apiKeyId,
        int $workspaceId,
        string $endpoint,
        string $method,
        int $statusCode,
        int $responseTimeMs,
        ?int $requestSize = null,
        ?int $responseSize = null,
        ?string $ipAddress = null,
        ?string $userAgent = null
    ): ApiUsage {
        // Normalise endpoint (remove query strings, IDs)
        $normalisedEndpoint = $this->normaliseEndpoint($endpoint);

        // Record individual usage
        $usage = ApiUsage::record(
            $apiKeyId,
            $workspaceId,
            $normalisedEndpoint,
            $method,
            $statusCode,
            $responseTimeMs,
            $requestSize,
            $responseSize,
            $ipAddress,
            $userAgent
        );

        // Update daily aggregation
        ApiUsageDaily::recordFromUsage($usage);

        return $usage;
    }

    /**
     * Get usage summary for a workspace.
     */
    public function getWorkspaceSummary(
        int $workspaceId,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        $query = ApiUsageDaily::forWorkspace($workspaceId)
            ->between($startDate, $endDate);

        $totals = (clone $query)->selectRaw('
            SUM(request_count) as total_requests,
            SUM(success_count) as total_success,
            SUM(error_count) as total_errors,
            SUM(total_response_time_ms) as total_response_time,
            MIN(min_response_time_ms) as min_response_time,
            MAX(max_response_time_ms) as max_response_time,
            SUM(total_request_size) as total_request_size,
            SUM(total_response_size) as total_response_size
        ')->first();

        $totalRequests = (int) ($totals->total_requests ?? 0);
        $totalSuccess = (int) ($totals->total_success ?? 0);

        return [
            'period' => [
                'start' => $startDate->toIso8601String(),
                'end' => $endDate->toIso8601String(),
            ],
            'totals' => [
                'requests' => $totalRequests,
                'success' => $totalSuccess,
                'errors' => (int) ($totals->total_errors ?? 0),
                'success_rate' => $totalRequests > 0
                    ? round(($totalSuccess / $totalRequests) * 100, 2)
                    : 100,
            ],
            'response_time' => [
                'average_ms' => $totalRequests > 0
                    ? round((int) $totals->total_response_time / $totalRequests, 2)
                    : 0,
                'min_ms' => (int) ($totals->min_response_time ?? 0),
                'max_ms' => (int) ($totals->max_response_time ?? 0),
            ],
            'data_transfer' => [
                'request_bytes' => (int) ($totals->total_request_size ?? 0),
                'response_bytes' => (int) ($totals->total_response_size ?? 0),
            ],
        ];
    }

    /**
     * Get usage summary for a specific API key.
     */
    public function getKeySummary(
        int $apiKeyId,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        $query = ApiUsageDaily::forKey($apiKeyId)
            ->between($startDate, $endDate);

        $totals = (clone $query)->selectRaw('
            SUM(request_count) as total_requests,
            SUM(success_count) as total_success,
            SUM(error_count) as total_errors,
            SUM(total_response_time_ms) as total_response_time,
            MIN(min_response_time_ms) as min_response_time,
            MAX(max_response_time_ms) as max_response_time
        ')->first();

        $totalRequests = (int) ($totals->total_requests ?? 0);
        $totalSuccess = (int) ($totals->total_success ?? 0);

        return [
            'period' => [
                'start' => $startDate->toIso8601String(),
                'end' => $endDate->toIso8601String(),
            ],
            'totals' => [
                'requests' => $totalRequests,
                'success' => $totalSuccess,
                'errors' => (int) ($totals->total_errors ?? 0),
                'success_rate' => $totalRequests > 0
                    ? round(($totalSuccess / $totalRequests) * 100, 2)
                    : 100,
            ],
            'response_time' => [
                'average_ms' => $totalRequests > 0
                    ? round((int) $totals->total_response_time / $totalRequests, 2)
                    : 0,
                'min_ms' => (int) ($totals->min_response_time ?? 0),
                'max_ms' => (int) ($totals->max_response_time ?? 0),
            ],
        ];
    }

    /**
     * Get daily usage chart data.
     */
    public function getDailyChart(
        int $workspaceId,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        $data = ApiUsageDaily::forWorkspace($workspaceId)
            ->between($startDate, $endDate)
            ->selectRaw('
                date,
                SUM(request_count) as requests,
                SUM(success_count) as success,
                SUM(error_count) as errors,
                SUM(total_response_time_ms) / NULLIF(SUM(request_count), 0) as avg_response_time
            ')
            ->groupBy('date')
            ->orderBy('date')
            ->get();

        return $data->map(fn ($row) => [
            'date' => $row->date->toDateString(),
            'requests' => (int) $row->requests,
            'success' => (int) $row->success,
            'errors' => (int) $row->errors,
            'avg_response_time_ms' => round((float) ($row->avg_response_time ?? 0), 2),
        ])->all();
    }

    /**
     * Get top endpoints by request count.
     */
    public function getTopEndpoints(
        int $workspaceId,
        int $limit = 10,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        return ApiUsageDaily::forWorkspace($workspaceId)
            ->between($startDate, $endDate)
            ->selectRaw('
                endpoint,
                method,
                SUM(request_count) as requests,
                SUM(success_count) as success,
                SUM(error_count) as errors,
                SUM(total_response_time_ms) / NULLIF(SUM(request_count), 0) as avg_response_time
            ')
            ->groupBy('endpoint', 'method')
            ->orderByDesc('requests')
            ->limit($limit)
            ->get()
            ->map(fn ($row) => [
                'endpoint' => $row->endpoint,
                'method' => $row->method,
                'requests' => (int) $row->requests,
                'success' => (int) $row->success,
                'errors' => (int) $row->errors,
                'success_rate' => $row->requests > 0
                    ? round(($row->success / $row->requests) * 100, 2)
                    : 100,
                'avg_response_time_ms' => round((float) ($row->avg_response_time ?? 0), 2),
            ])
            ->all();
    }

    /**
     * Get error breakdown by status code.
     */
    public function getErrorBreakdown(
        int $workspaceId,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        return ApiUsage::forWorkspace($workspaceId)
            ->between($startDate, $endDate)
            ->where('status_code', '>=', 400)
            ->selectRaw('status_code, COUNT(*) as count')
            ->groupBy('status_code')
            ->orderByDesc('count')
            ->get()
            ->map(fn ($row) => [
                'status_code' => $row->status_code,
                'count' => (int) $row->count,
                'description' => $this->getStatusCodeDescription($row->status_code),
            ])
            ->all();
    }

    /**
     * Get API key usage comparison.
     */
    public function getKeyComparison(
        int $workspaceId,
        ?Carbon $startDate = null,
        ?Carbon $endDate = null
    ): array {
        $startDate = $startDate ?? now()->subDays(30);
        $endDate = $endDate ?? now();

        $aggregated = ApiUsageDaily::forWorkspace($workspaceId)
            ->between($startDate, $endDate)
            ->selectRaw('
                api_key_id,
                SUM(request_count) as requests,
                SUM(success_count) as success,
                SUM(error_count) as errors,
                SUM(total_response_time_ms) / NULLIF(SUM(request_count), 0) as avg_response_time
            ')
            ->groupBy('api_key_id')
            ->orderByDesc('requests')
            ->get();

        // Fetch API keys separately to avoid broken eager loading with aggregation
        $apiKeyIds = $aggregated->pluck('api_key_id')->filter()->unique()->all();
        $apiKeys = \Mod\Api\Models\ApiKey::whereIn('id', $apiKeyIds)
            ->select('id', 'name', 'prefix')
            ->get()
            ->keyBy('id');

        return $aggregated->map(fn ($row) => [
            'api_key_id' => $row->api_key_id,
            'api_key_name' => $apiKeys->get($row->api_key_id)?->name ?? 'Unknown',
            'api_key_prefix' => $apiKeys->get($row->api_key_id)?->prefix ?? 'N/A',
            'requests' => (int) $row->requests,
            'success' => (int) $row->success,
            'errors' => (int) $row->errors,
            'success_rate' => $row->requests > 0
                ? round(($row->success / $row->requests) * 100, 2)
                : 100,
            'avg_response_time_ms' => round((float) ($row->avg_response_time ?? 0), 2),
        ])->all();
    }

    /**
     * Normalise endpoint path for aggregation.
     *
     * Replaces dynamic IDs with placeholders for consistent grouping.
     */
    protected function normaliseEndpoint(string $endpoint): string
    {
        // Remove query string
        $path = parse_url($endpoint, PHP_URL_PATH) ?? $endpoint;

        // Replace numeric IDs with {id} placeholder
        $normalised = preg_replace('/\/\d+/', '/{id}', $path);

        // Replace UUIDs with {uuid} placeholder
        $normalised = preg_replace(
            '/\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i',
            '/{uuid}',
            $normalised
        );

        return $normalised ?? $path;
    }

    /**
     * Get human-readable status code description.
     */
    protected function getStatusCodeDescription(int $statusCode): string
    {
        return match ($statusCode) {
            400 => 'Bad Request',
            401 => 'Unauthorised',
            403 => 'Forbidden',
            404 => 'Not Found',
            405 => 'Method Not Allowed',
            422 => 'Validation Failed',
            429 => 'Rate Limit Exceeded',
            500 => 'Internal Server Error',
            502 => 'Bad Gateway',
            503 => 'Service Unavailable',
            504 => 'Gateway Timeout',
            default => 'Error',
        };
    }

    /**
     * Prune old detailed usage records.
     *
     * Keeps aggregated daily data but removes detailed logs older than retention period.
     *
     * @return int Number of records deleted
     */
    public function pruneOldRecords(int $retentionDays = 30): int
    {
        $cutoff = now()->subDays($retentionDays);

        return ApiUsage::where('created_at', '<', $cutoff)->delete();
    }
}
