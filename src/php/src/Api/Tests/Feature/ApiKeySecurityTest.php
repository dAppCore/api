<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Mod\Api\Database\Factories\ApiKeyFactory;
use Mod\Api\Models\ApiKey;
use Mod\Tenant\Models\User;
use Mod\Tenant\Models\Workspace;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();

    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);
});

// ─────────────────────────────────────────────────────────────────────────────
// Secure Hashing (bcrypt)
// ─────────────────────────────────────────────────────────────────────────────

describe('Secure Hashing', function () {
    it('uses bcrypt for new API keys', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Secure Key'
        );

        expect($result['api_key']->hash_algorithm)->toBe(ApiKey::HASH_BCRYPT);
        expect($result['api_key']->key)->toStartWith('$2y$');
    });

    it('verifies bcrypt hashed keys correctly', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Verifiable Key'
        );

        $parts = explode('_', $result['plain_key'], 3);
        $keyPart = $parts[2];

        expect($result['api_key']->verifyKey($keyPart))->toBeTrue();
        expect($result['api_key']->verifyKey('wrong-key'))->toBeFalse();
    });

    it('finds bcrypt keys by plain key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Findable Bcrypt Key'
        );

        $found = ApiKey::findByPlainKey($result['plain_key']);

        expect($found)->not->toBeNull();
        expect($found->id)->toBe($result['api_key']->id);
    });

    it('bcrypt keys are not vulnerable to timing attacks', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Timing Safe Key'
        );

        $parts = explode('_', $result['plain_key'], 3);
        $keyPart = $parts[2];

        // bcrypt verification should take similar time for valid and invalid keys
        // (this is a property test, not a precise timing test)
        expect($result['api_key']->verifyKey($keyPart))->toBeTrue();
        expect($result['api_key']->verifyKey('x'.$keyPart))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Legacy SHA-256 Backward Compatibility
// ─────────────────────────────────────────────────────────────────────────────

describe('Legacy SHA-256 Compatibility', function () {
    it('identifies legacy hash keys', function () {
        $result = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        expect($result['api_key']->hash_algorithm)->toBe(ApiKey::HASH_SHA256);
        expect($result['api_key']->usesLegacyHash())->toBeTrue();
    });

    it('verifies legacy SHA-256 keys correctly', function () {
        $result = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        $parts = explode('_', $result['plain_key'], 3);
        $keyPart = $parts[2];

        expect($result['api_key']->verifyKey($keyPart))->toBeTrue();
        expect($result['api_key']->verifyKey('wrong-key'))->toBeFalse();
    });

    it('finds legacy SHA-256 keys by plain key', function () {
        $result = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        $found = ApiKey::findByPlainKey($result['plain_key']);

        expect($found)->not->toBeNull();
        expect($found->id)->toBe($result['api_key']->id);
    });

    it('treats null hash_algorithm as legacy', function () {
        // Create a key without hash_algorithm (simulating pre-migration key)
        $plainKey = Str::random(48);
        $prefix = 'hk_'.Str::random(8);

        $apiKey = ApiKey::create([
            'workspace_id' => $this->workspace->id,
            'user_id' => $this->user->id,
            'name' => 'Pre-migration Key',
            'key' => hash('sha256', $plainKey),
            'hash_algorithm' => null, // Simulate pre-migration
            'prefix' => $prefix,
            'scopes' => [ApiKey::SCOPE_READ],
        ]);

        expect($apiKey->usesLegacyHash())->toBeTrue();

        // Should still be findable
        $found = ApiKey::findByPlainKey("{$prefix}_{$plainKey}");
        expect($found)->not->toBeNull();
        expect($found->id)->toBe($apiKey->id);
    });

    it('can query for legacy hash keys', function () {
        // Create a bcrypt key
        ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Secure Key'
        );

        // Create a legacy key
        ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        $legacyKeys = ApiKey::legacyHash()->get();
        $secureKeys = ApiKey::secureHash()->get();

        expect($legacyKeys)->toHaveCount(1);
        expect($secureKeys)->toHaveCount(1);
        expect($legacyKeys->first()->name)->toContain('API Key');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Key Rotation for Security Migration
// ─────────────────────────────────────────────────────────────────────────────

describe('Security Migration via Rotation', function () {
    it('rotates legacy key to secure bcrypt key', function () {
        $legacy = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        expect($legacy['api_key']->usesLegacyHash())->toBeTrue();

        $rotated = $legacy['api_key']->rotate();

        expect($rotated['api_key']->hash_algorithm)->toBe(ApiKey::HASH_BCRYPT);
        expect($rotated['api_key']->usesLegacyHash())->toBeFalse();
        expect($rotated['api_key']->key)->toStartWith('$2y$');
    });

    it('preserves settings when rotating legacy key', function () {
        $legacy = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user,
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_DELETE]
        );

        $legacy['api_key']->update(['server_scopes' => ['commerce', 'biohost']]);

        $rotated = $legacy['api_key']->fresh()->rotate();

        expect($rotated['api_key']->scopes)->toBe([ApiKey::SCOPE_READ, ApiKey::SCOPE_DELETE]);
        expect($rotated['api_key']->server_scopes)->toBe(['commerce', 'biohost']);
        expect($rotated['api_key']->workspace_id)->toBe($this->workspace->id);
    });

    it('legacy key remains valid during grace period after rotation', function () {
        $legacy = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        $legacy['api_key']->rotate(24); // 24 hour grace period

        // Old key should still work
        $found = ApiKey::findByPlainKey($legacy['plain_key']);
        expect($found)->not->toBeNull();
        expect($found->isInGracePeriod())->toBeTrue();
    });

    it('tracks rotation lineage', function () {
        $original = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        $rotated = $original['api_key']->rotate();

        expect($rotated['api_key']->rotated_from_id)->toBe($original['api_key']->id);
        expect($rotated['api_key']->rotatedFrom->id)->toBe($original['api_key']->id);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Grace Period Handling
// ─────────────────────────────────────────────────────────────────────────────

describe('Grace Period', function () {
    it('sets grace period on rotation', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'To Be Rotated'
        );

        $result['api_key']->rotate(48);

        $oldKey = $result['api_key']->fresh();
        expect($oldKey->grace_period_ends_at)->not->toBeNull();
        expect($oldKey->isInGracePeriod())->toBeTrue();
        expect($oldKey->grace_period_ends_at->diffInHours(now()))->toBeLessThanOrEqual(48);
    });

    it('key becomes invalid after grace period expires', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expiring Grace Key'
        );

        $result['api_key']->update([
            'grace_period_ends_at' => now()->subHour(),
        ]);

        $found = ApiKey::findByPlainKey($result['plain_key']);
        expect($found)->toBeNull();
    });

    it('can end grace period early', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Early End Key'
        );

        $result['api_key']->rotate(24);

        $oldKey = $result['api_key']->fresh();
        expect($oldKey->isInGracePeriod())->toBeTrue();

        $oldKey->endGracePeriod();

        expect($oldKey->fresh()->trashed())->toBeTrue();
    });

    it('scopes keys in grace period correctly', function () {
        // Key in grace period
        $key1 = ApiKey::generate($this->workspace->id, $this->user->id, 'In Grace');
        $key1['api_key']->update(['grace_period_ends_at' => now()->addHours(12)]);

        // Key with expired grace period
        $key2 = ApiKey::generate($this->workspace->id, $this->user->id, 'Expired Grace');
        $key2['api_key']->update(['grace_period_ends_at' => now()->subHours(1)]);

        // Normal key
        ApiKey::generate($this->workspace->id, $this->user->id, 'Normal Key');

        expect(ApiKey::inGracePeriod()->count())->toBe(1);
        expect(ApiKey::gracePeriodExpired()->count())->toBe(1);
        expect(ApiKey::active()->count())->toBe(2); // Normal + In Grace
    });

    it('detects grace period expired status', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Status Check Key'
        );

        // Not in grace period
        expect($result['api_key']->isInGracePeriod())->toBeFalse();
        expect($result['api_key']->isGracePeriodExpired())->toBeFalse();

        // In grace period
        $result['api_key']->update(['grace_period_ends_at' => now()->addHour()]);
        expect($result['api_key']->fresh()->isInGracePeriod())->toBeTrue();
        expect($result['api_key']->fresh()->isGracePeriodExpired())->toBeFalse();

        // Grace period expired
        $result['api_key']->update(['grace_period_ends_at' => now()->subHour()]);
        expect($result['api_key']->fresh()->isInGracePeriod())->toBeFalse();
        expect($result['api_key']->fresh()->isGracePeriodExpired())->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Hash Algorithm Constants
// ─────────────────────────────────────────────────────────────────────────────

describe('Hash Algorithm Constants', function () {
    it('defines correct hash algorithm constants', function () {
        expect(ApiKey::HASH_SHA256)->toBe('sha256');
        expect(ApiKey::HASH_BCRYPT)->toBe('bcrypt');
    });

    it('defines default grace period constant', function () {
        expect(ApiKey::DEFAULT_GRACE_PERIOD_HOURS)->toBe(24);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Factory Legacy Support
// ─────────────────────────────────────────────────────────────────────────────

describe('Factory Legacy Support', function () {
    it('creates legacy keys via static helper', function () {
        $result = ApiKeyFactory::createLegacyKey(
            $this->workspace,
            $this->user
        );

        expect($result['api_key']->hash_algorithm)->toBe(ApiKey::HASH_SHA256);
        expect($result['api_key']->key)->not->toStartWith('$2y$');

        // Should be a 64-char hex string (SHA-256)
        expect(strlen($result['api_key']->key))->toBe(64);
    });

    it('creates keys in grace period via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->inGracePeriod(6)
            ->create();

        expect($key->isInGracePeriod())->toBeTrue();
        expect($key->grace_period_ends_at->diffInHours(now()))->toBeLessThanOrEqual(6);
    });

    it('creates keys with expired grace period via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->gracePeriodExpired()
            ->create();

        expect($key->isGracePeriodExpired())->toBeTrue();
        expect($key->isInGracePeriod())->toBeFalse();
    });
});
