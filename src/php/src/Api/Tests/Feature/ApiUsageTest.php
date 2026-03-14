<?php

declare(strict_types=1);

use Mod\Api\Models\ApiKey;
use Mod\Api\Models\ApiUsage;
use Mod\Api\Models\ApiUsageDaily;
use Mod\Api\Services\ApiUsageService;
use Mod\Tenant\Models\User;
use Mod\Tenant\Models\Workspace;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);

    $result = ApiKey::generate($this->workspace->id, $this->user->id, 'Test Key');
    $this->apiKey = $result['api_key'];

    $this->service = app(ApiUsageService::class);
});

// ─────────────────────────────────────────────────────────────────────────────
// Recording Usage
// ─────────────────────────────────────────────────────────────────────────────

describe('Recording API Usage', function () {
    it('records individual usage entries', function () {
        $usage = $this->service->record(
            apiKeyId: $this->apiKey->id,
            workspaceId: $this->workspace->id,
            endpoint: '/api/v1/workspaces',
            method: 'GET',
            statusCode: 200,
            responseTimeMs: 150,
            requestSize: 0,
            responseSize: 1024
        );

        expect($usage)->toBeInstanceOf(ApiUsage::class);
        expect($usage->api_key_id)->toBe($this->apiKey->id);
        expect($usage->endpoint)->toBe('/api/v1/workspaces');
        expect($usage->method)->toBe('GET');
        expect($usage->status_code)->toBe(200);
        expect($usage->response_time_ms)->toBe(150);
    });

    it('normalises endpoint paths with IDs', function () {
        $usage = $this->service->record(
            apiKeyId: $this->apiKey->id,
            workspaceId: $this->workspace->id,
            endpoint: '/api/v1/workspaces/123/users/456',
            method: 'GET',
            statusCode: 200,
            responseTimeMs: 100
        );

        expect($usage->endpoint)->toBe('/api/v1/workspaces/{id}/users/{id}');
    });

    it('normalises endpoint paths with UUIDs', function () {
        $usage = $this->service->record(
            apiKeyId: $this->apiKey->id,
            workspaceId: $this->workspace->id,
            endpoint: '/api/v1/resources/550e8400-e29b-41d4-a716-446655440000',
            method: 'GET',
            statusCode: 200,
            responseTimeMs: 100
        );

        expect($usage->endpoint)->toBe('/api/v1/resources/{uuid}');
    });

    it('updates daily aggregation on record', function () {
        $this->service->record(
            apiKeyId: $this->apiKey->id,
            workspaceId: $this->workspace->id,
            endpoint: '/api/v1/test',
            method: 'GET',
            statusCode: 200,
            responseTimeMs: 100
        );

        $daily = ApiUsageDaily::forKey($this->apiKey->id)
            ->where('date', now()->toDateString())
            ->first();

        expect($daily)->not->toBeNull();
        expect($daily->request_count)->toBe(1);
        expect($daily->success_count)->toBe(1);
    });

    it('increments daily counts correctly', function () {
        // Record multiple requests
        for ($i = 0; $i < 5; $i++) {
            $this->service->record(
                apiKeyId: $this->apiKey->id,
                workspaceId: $this->workspace->id,
                endpoint: '/api/v1/test',
                method: 'GET',
                statusCode: 200,
                responseTimeMs: 100 + ($i * 10)
            );
        }

        // Record some errors
        for ($i = 0; $i < 2; $i++) {
            $this->service->record(
                apiKeyId: $this->apiKey->id,
                workspaceId: $this->workspace->id,
                endpoint: '/api/v1/test',
                method: 'GET',
                statusCode: 500,
                responseTimeMs: 50
            );
        }

        $daily = ApiUsageDaily::forKey($this->apiKey->id)
            ->where('date', now()->toDateString())
            ->first();

        expect($daily->request_count)->toBe(7);
        expect($daily->success_count)->toBe(5);
        expect($daily->error_count)->toBe(2);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Usage Summaries
// ─────────────────────────────────────────────────────────────────────────────

describe('Usage Summaries', function () {
    beforeEach(function () {
        // Create some usage data
        for ($i = 0; $i < 10; $i++) {
            $this->service->record(
                apiKeyId: $this->apiKey->id,
                workspaceId: $this->workspace->id,
                endpoint: '/api/v1/workspaces',
                method: 'GET',
                statusCode: 200,
                responseTimeMs: 100 + $i
            );
        }

        for ($i = 0; $i < 3; $i++) {
            $this->service->record(
                apiKeyId: $this->apiKey->id,
                workspaceId: $this->workspace->id,
                endpoint: '/api/v1/workspaces',
                method: 'POST',
                statusCode: 422,
                responseTimeMs: 50
            );
        }
    });

    it('returns workspace summary', function () {
        $summary = $this->service->getWorkspaceSummary($this->workspace->id);

        expect($summary)->toHaveKeys(['period', 'totals', 'response_time', 'data_transfer']);
        expect($summary['totals']['requests'])->toBe(13);
        expect($summary['totals']['success'])->toBe(10);
        expect($summary['totals']['errors'])->toBe(3);
    });

    it('returns key summary', function () {
        $summary = $this->service->getKeySummary($this->apiKey->id);

        expect($summary['totals']['requests'])->toBe(13);
        expect($summary['totals']['success_rate'])->toBeGreaterThan(70);
    });

    it('calculates average response time', function () {
        $summary = $this->service->getWorkspaceSummary($this->workspace->id);

        // (100+101+102+...+109 + 50*3) / 13
        expect($summary['response_time']['average_ms'])->toBeGreaterThan(0);
    });

    it('filters by date range', function () {
        // Create usage for 2 days ago with correct timestamp upfront
        $oldDate = now()->subDays(2);
        $usage = ApiUsage::create([
            'api_key_id' => $this->apiKey->id,
            'workspace_id' => $this->workspace->id,
            'endpoint' => '/api/v1/old',
            'method' => 'GET',
            'status_code' => 200,
            'response_time_ms' => 100,
            'created_at' => $oldDate,
            'updated_at' => $oldDate,
        ]);

        // Also create a backdated daily aggregate for consistency
        ApiUsageDaily::updateOrCreate(
            [
                'api_key_id' => $this->apiKey->id,
                'date' => $oldDate->toDateString(),
            ],
            [
                'request_count' => 1,
                'success_count' => 1,
                'error_count' => 0,
                'total_response_time_ms' => 100,
                'total_request_size' => 0,
                'total_response_size' => 0,
            ]
        );

        // Summary for last 24 hours should not include old data
        $summary = $this->service->getWorkspaceSummary(
            $this->workspace->id,
            now()->subDay(),
            now()
        );

        expect($summary['totals']['requests'])->toBe(13); // Only today's requests
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Charts and Reports
// ─────────────────────────────────────────────────────────────────────────────

describe('Charts and Reports', function () {
    beforeEach(function () {
        // Create usage spread across days
        for ($day = 0; $day < 7; $day++) {
            $date = now()->subDays($day);
            $requests = 10 - $day;

            for ($i = 0; $i < $requests; $i++) {
                $usage = ApiUsage::record(
                    $this->apiKey->id,
                    $this->workspace->id,
                    '/api/v1/test',
                    'GET',
                    200,
                    100
                );
                $usage->update(['created_at' => $date]);

                ApiUsageDaily::recordFromUsage($usage);
            }
        }
    });

    it('returns daily chart data', function () {
        $chart = $this->service->getDailyChart($this->workspace->id);

        expect($chart)->toBeArray();
        expect(count($chart))->toBeGreaterThan(0);
        expect($chart[0])->toHaveKeys(['date', 'requests', 'success', 'errors', 'avg_response_time_ms']);
    });

    it('returns top endpoints', function () {
        // Add some variety
        $this->service->record(
            $this->apiKey->id,
            $this->workspace->id,
            '/api/v1/popular',
            'GET',
            200,
            100
        );

        $endpoints = $this->service->getTopEndpoints($this->workspace->id, 5);

        expect($endpoints)->toBeArray();
        expect($endpoints[0])->toHaveKeys(['endpoint', 'method', 'requests', 'success_rate', 'avg_response_time_ms']);
    });

    it('returns error breakdown', function () {
        // Add some errors
        $this->service->record($this->apiKey->id, $this->workspace->id, '/api/v1/test', 'GET', 401, 50);
        $this->service->record($this->apiKey->id, $this->workspace->id, '/api/v1/test', 'GET', 404, 50);
        $this->service->record($this->apiKey->id, $this->workspace->id, '/api/v1/test', 'GET', 500, 50);

        $errors = $this->service->getErrorBreakdown($this->workspace->id);

        expect($errors)->toBeArray();
        expect(count($errors))->toBe(3);
        expect($errors[0])->toHaveKeys(['status_code', 'count', 'description']);
    });

    it('returns key comparison', function () {
        // Create another key with usage
        $key2 = ApiKey::generate($this->workspace->id, $this->user->id, 'Second Key');
        $this->service->record($key2['api_key']->id, $this->workspace->id, '/api/v1/test', 'GET', 200, 100);

        $comparison = $this->service->getKeyComparison($this->workspace->id);

        expect($comparison)->toBeArray();
        expect(count($comparison))->toBe(2);
        expect($comparison[0])->toHaveKeys(['api_key_id', 'api_key_name', 'requests', 'success_rate']);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Data Retention
// ─────────────────────────────────────────────────────────────────────────────

describe('Data Retention', function () {
    it('prunes old detailed records', function () {
        // Create old records
        for ($i = 0; $i < 5; $i++) {
            $usage = ApiUsage::record(
                $this->apiKey->id,
                $this->workspace->id,
                '/api/v1/old',
                'GET',
                200,
                100
            );
            $usage->update(['created_at' => now()->subDays(60)]);
        }

        // Create recent records
        for ($i = 0; $i < 3; $i++) {
            ApiUsage::record(
                $this->apiKey->id,
                $this->workspace->id,
                '/api/v1/recent',
                'GET',
                200,
                100
            );
        }

        $deleted = $this->service->pruneOldRecords(30);

        expect($deleted)->toBe(5);
        expect(ApiUsage::count())->toBe(3);
    });

    it('keeps daily aggregates when pruning detailed records', function () {
        // Create and aggregate old record
        $usage = ApiUsage::record(
            $this->apiKey->id,
            $this->workspace->id,
            '/api/v1/old',
            'GET',
            200,
            100
        );
        $usage->update(['created_at' => now()->subDays(60)]);
        ApiUsageDaily::recordFromUsage($usage);

        $dailyCountBefore = ApiUsageDaily::count();

        $this->service->pruneOldRecords(30);

        // Daily aggregates should remain
        expect(ApiUsageDaily::count())->toBe($dailyCountBefore);
    });
});
