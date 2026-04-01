<?php

declare(strict_types=1);

use Mod\Api\Models\ApiKey;
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

    $result = ApiKey::generate(
        $this->workspace->id,
        $this->user->id,
        'Entitlements Key',
        [ApiKey::SCOPE_READ]
    );

    $this->plainKey = $result['plain_key'];
    $this->apiKey = $result['api_key'];
});

it('returns entitlement limits and usage for the current workspace', function () {
    app(ApiUsageService::class)->record(
        apiKeyId: $this->apiKey->id,
        workspaceId: $this->workspace->id,
        endpoint: '/api/entitlements',
        method: 'GET',
        statusCode: 200,
        responseTimeMs: 42,
        ipAddress: '127.0.0.1',
        userAgent: 'Pest'
    );

    $response = $this->getJson('/api/entitlements', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('workspace_id', $this->workspace->id);
    $response->assertJsonPath('authentication.type', 'api_key');
    $response->assertJsonPath('limits.api_keys.maximum', config('api.keys.max_per_workspace'));
    $response->assertJsonPath('limits.api_keys.active', 1);
    $response->assertJsonPath('usage.totals.requests', 1);
    $response->assertJsonPath('features.mcp', true);
});
