<?php

declare(strict_types=1);

use Mod\Api\Models\ApiKey;
use Mod\Api\Services\ApiKeyService;
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
    $this->service = app(ApiKeyService::class);
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Rotation
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Rotation', function () {
    it('rotates a key creating new key with same settings', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Original Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $result = $this->service->rotate($original['api_key']);

        expect($result)->toHaveKeys(['api_key', 'plain_key', 'old_key']);
        expect($result['api_key']->name)->toBe('Original Key');
        expect($result['api_key']->scopes)->toBe([ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]);
        expect($result['api_key']->workspace_id)->toBe($this->workspace->id);
        expect($result['api_key']->rotated_from_id)->toBe($original['api_key']->id);
    });

    it('sets grace period on old key during rotation', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Grace Period Key'
        );

        $result = $this->service->rotate($original['api_key'], 24);

        $oldKey = $result['old_key']->fresh();
        expect($oldKey->grace_period_ends_at)->not->toBeNull();
        expect($oldKey->isInGracePeriod())->toBeTrue();
    });

    it('old key remains valid during grace period', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Still Valid Key'
        );

        $this->service->rotate($original['api_key'], 24);

        // Old key should still be findable
        $foundKey = ApiKey::findByPlainKey($original['plain_key']);
        expect($foundKey)->not->toBeNull();
        expect($foundKey->id)->toBe($original['api_key']->id);
    });

    it('old key becomes invalid after grace period expires', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expired Grace Key'
        );

        $original['api_key']->update([
            'grace_period_ends_at' => now()->subHour(),
        ]);

        $foundKey = ApiKey::findByPlainKey($original['plain_key']);
        expect($foundKey)->toBeNull();
    });

    it('prevents rotating key already in grace period', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Already Rotating Key'
        );

        $this->service->rotate($original['api_key']);

        expect(fn () => $this->service->rotate($original['api_key']->fresh()))
            ->toThrow(\RuntimeException::class);
    });

    it('can end grace period early', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Early End Key'
        );

        $this->service->rotate($original['api_key'], 24);
        $this->service->endGracePeriod($original['api_key']->fresh());

        expect($original['api_key']->fresh()->trashed())->toBeTrue();
    });

    it('preserves server scopes during rotation', function () {
        $original = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Server Scoped Key'
        );
        $original['api_key']->update(['server_scopes' => ['commerce', 'biohost']]);

        $result = $this->service->rotate($original['api_key']->fresh());

        expect($result['api_key']->server_scopes)->toBe(['commerce', 'biohost']);
    });

    it('cleans up keys with expired grace periods', function () {
        // Create keys with expired grace periods
        $key1 = ApiKey::generate($this->workspace->id, $this->user->id, 'Expired 1');
        $key1['api_key']->update(['grace_period_ends_at' => now()->subDay()]);

        $key2 = ApiKey::generate($this->workspace->id, $this->user->id, 'Expired 2');
        $key2['api_key']->update(['grace_period_ends_at' => now()->subHour()]);

        // Create key still in grace period
        $key3 = ApiKey::generate($this->workspace->id, $this->user->id, 'Still Active');
        $key3['api_key']->update(['grace_period_ends_at' => now()->addDay()]);

        $cleaned = $this->service->cleanupExpiredGracePeriods();

        expect($cleaned)->toBe(2);
        expect($key1['api_key']->fresh()->trashed())->toBeTrue();
        expect($key2['api_key']->fresh()->trashed())->toBeTrue();
        expect($key3['api_key']->fresh()->trashed())->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Scopes via Service
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Service Scopes', function () {
    it('updates key scopes', function () {
        $result = $this->service->create(
            $this->workspace->id,
            $this->user->id,
            'Scoped Key'
        );

        $this->service->updateScopes($result['api_key'], [ApiKey::SCOPE_READ]);

        expect($result['api_key']->fresh()->scopes)->toBe([ApiKey::SCOPE_READ]);
    });

    it('requires at least one valid scope', function () {
        $result = $this->service->create(
            $this->workspace->id,
            $this->user->id,
            'Invalid Scopes Key'
        );

        expect(fn () => $this->service->updateScopes($result['api_key'], ['invalid']))
            ->toThrow(\InvalidArgumentException::class);
    });

    it('updates server scopes', function () {
        $result = $this->service->create(
            $this->workspace->id,
            $this->user->id,
            'Server Scoped Key'
        );

        $this->service->updateServerScopes($result['api_key'], ['commerce']);

        expect($result['api_key']->fresh()->server_scopes)->toBe(['commerce']);
    });

    it('clears server scopes with null', function () {
        $result = $this->service->create(
            $this->workspace->id,
            $this->user->id,
            'Clear Server Scopes Key',
            serverScopes: ['commerce']
        );

        $this->service->updateServerScopes($result['api_key'], null);

        expect($result['api_key']->fresh()->server_scopes)->toBeNull();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Service Limits
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Service Limits', function () {
    it('enforces max keys per workspace limit', function () {
        config(['api.keys.max_per_workspace' => 2]);

        $this->service->create($this->workspace->id, $this->user->id, 'Key 1');
        $this->service->create($this->workspace->id, $this->user->id, 'Key 2');

        expect(fn () => $this->service->create($this->workspace->id, $this->user->id, 'Key 3'))
            ->toThrow(\RuntimeException::class);
    });

    it('returns workspace key statistics', function () {
        $key1 = $this->service->create($this->workspace->id, $this->user->id, 'Active Key');
        $key2 = $this->service->create($this->workspace->id, $this->user->id, 'Expired Key');
        $key2['api_key']->update(['expires_at' => now()->subDay()]);

        $key3 = $this->service->create($this->workspace->id, $this->user->id, 'Rotating Key');
        $this->service->rotate($key3['api_key']);

        $stats = $this->service->getStats($this->workspace->id);

        expect($stats)->toHaveKeys(['total', 'active', 'expired', 'in_grace_period', 'revoked']);
        expect($stats['total'])->toBe(4); // 3 original + 1 rotated
        expect($stats['expired'])->toBe(1);
        expect($stats['in_grace_period'])->toBe(1);
    });
});
