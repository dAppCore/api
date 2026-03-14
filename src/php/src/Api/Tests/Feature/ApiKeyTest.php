<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;
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
// API Key Creation
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Creation', function () {
    it('generates a new API key with correct format', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Test API Key'
        );

        expect($result)->toHaveKeys(['api_key', 'plain_key']);
        expect($result['api_key'])->toBeInstanceOf(ApiKey::class);
        expect($result['plain_key'])->toStartWith('hk_');

        // Plain key format: hk_xxxxxxxx_xxxx...
        $parts = explode('_', $result['plain_key']);
        expect($parts)->toHaveCount(3);
        expect($parts[0])->toBe('hk');
        expect(strlen($parts[1]))->toBe(8);
        expect(strlen($parts[2]))->toBe(48);
    });

    it('creates key with default read and write scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Default Scopes Key'
        );

        expect($result['api_key']->scopes)->toBe([ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]);
    });

    it('creates key with custom scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Full Access Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE, ApiKey::SCOPE_DELETE]
        );

        expect($result['api_key']->scopes)->toBe(ApiKey::ALL_SCOPES);
    });

    it('creates key with expiry date', function () {
        $expiresAt = now()->addDays(30);

        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expiring Key',
            [ApiKey::SCOPE_READ],
            $expiresAt
        );

        expect($result['api_key']->expires_at)->not->toBeNull();
        expect($result['api_key']->expires_at->timestamp)->toBe($expiresAt->timestamp);
    });

    it('stores key as bcrypt hashed value', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Hashed Key'
        );

        // Extract the key part from plain key
        $parts = explode('_', $result['plain_key'], 3);
        $keyPart = $parts[2];

        // The stored key should be a bcrypt hash (starts with $2y$)
        expect($result['api_key']->key)->toStartWith('$2y$');
        expect($result['api_key']->hash_algorithm)->toBe(ApiKey::HASH_BCRYPT);

        // Verify the key matches using Hash::check
        expect(\Illuminate\Support\Facades\Hash::check($keyPart, $result['api_key']->key))->toBeTrue();
    });

    it('sets hash_algorithm to bcrypt for new keys', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Bcrypt Key'
        );

        expect($result['api_key']->hash_algorithm)->toBe(ApiKey::HASH_BCRYPT);
        expect($result['api_key']->usesLegacyHash())->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Authentication
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Authentication', function () {
    it('finds key by valid plain key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Findable Key'
        );

        $foundKey = ApiKey::findByPlainKey($result['plain_key']);

        expect($foundKey)->not->toBeNull();
        expect($foundKey->id)->toBe($result['api_key']->id);
    });

    it('returns null for invalid key format', function () {
        expect(ApiKey::findByPlainKey('invalid-key'))->toBeNull();
        expect(ApiKey::findByPlainKey('hk_only_two_parts'))->toBeNull();
        expect(ApiKey::findByPlainKey(''))->toBeNull();
    });

    it('returns null for non-existent key', function () {
        $result = ApiKey::findByPlainKey('hk_nonexist_'.str_repeat('x', 48));

        expect($result)->toBeNull();
    });

    it('returns null for expired key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expired Key',
            [ApiKey::SCOPE_READ],
            now()->subDay() // Already expired
        );

        $foundKey = ApiKey::findByPlainKey($result['plain_key']);

        expect($foundKey)->toBeNull();
    });

    it('returns null for revoked (soft-deleted) key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Revoked Key'
        );

        $result['api_key']->revoke();

        $foundKey = ApiKey::findByPlainKey($result['plain_key']);

        expect($foundKey)->toBeNull();
    });

    it('records usage on authentication', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Tracking Key'
        );

        expect($result['api_key']->last_used_at)->toBeNull();

        $result['api_key']->recordUsage();

        expect($result['api_key']->fresh()->last_used_at)->not->toBeNull();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Scope Checking
// ─────────────────────────────────────────────────────────────────────────────

describe('Scope Checking', function () {
    it('checks for single scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Scoped Key',
            [ApiKey::SCOPE_READ]
        );

        $key = $result['api_key'];

        expect($key->hasScope(ApiKey::SCOPE_READ))->toBeTrue();
        expect($key->hasScope(ApiKey::SCOPE_WRITE))->toBeFalse();
        expect($key->hasScope(ApiKey::SCOPE_DELETE))->toBeFalse();
    });

    it('checks for multiple scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Multi-Scoped Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $key = $result['api_key'];

        expect($key->hasScopes([ApiKey::SCOPE_READ]))->toBeTrue();
        expect($key->hasScopes([ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]))->toBeTrue();
        expect($key->hasScopes([ApiKey::SCOPE_READ, ApiKey::SCOPE_DELETE]))->toBeFalse();
    });

    it('returns available scope constants', function () {
        expect(ApiKey::SCOPE_READ)->toBe('read');
        expect(ApiKey::SCOPE_WRITE)->toBe('write');
        expect(ApiKey::SCOPE_DELETE)->toBe('delete');
        expect(ApiKey::ALL_SCOPES)->toBe(['read', 'write', 'delete']);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Expiry Handling
// ─────────────────────────────────────────────────────────────────────────────

describe('Expiry Handling', function () {
    it('detects expired key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Past Expiry Key',
            [ApiKey::SCOPE_READ],
            now()->subDay()
        );

        expect($result['api_key']->isExpired())->toBeTrue();
    });

    it('detects non-expired key', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Future Expiry Key',
            [ApiKey::SCOPE_READ],
            now()->addDay()
        );

        expect($result['api_key']->isExpired())->toBeFalse();
    });

    it('keys without expiry are never expired', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'No Expiry Key'
        );

        expect($result['api_key']->expires_at)->toBeNull();
        expect($result['api_key']->isExpired())->toBeFalse();
    });

    it('scopes expired keys correctly', function () {
        // Create expired key
        ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expired Key 1',
            [ApiKey::SCOPE_READ],
            now()->subDays(2)
        );

        // Create active key
        ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Active Key',
            [ApiKey::SCOPE_READ],
            now()->addDays(30)
        );

        // Create no-expiry key
        ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'No Expiry Key'
        );

        $expired = ApiKey::expired()->count();
        $active = ApiKey::active()->count();

        expect($expired)->toBe(1);
        expect($active)->toBe(2);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Server Scopes (MCP Access)
// ─────────────────────────────────────────────────────────────────────────────

describe('Server Scopes', function () {
    it('allows all servers when server_scopes is null', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'All Servers Key'
        );

        $key = $result['api_key'];

        expect($key->server_scopes)->toBeNull();
        expect($key->hasServerAccess('commerce'))->toBeTrue();
        expect($key->hasServerAccess('biohost'))->toBeTrue();
        expect($key->hasServerAccess('anything'))->toBeTrue();
    });

    it('restricts to specific servers when server_scopes is set', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Limited Servers Key'
        );

        $key = $result['api_key'];
        $key->update(['server_scopes' => ['commerce', 'biohost']]);

        expect($key->hasServerAccess('commerce'))->toBeTrue();
        expect($key->hasServerAccess('biohost'))->toBeTrue();
        expect($key->hasServerAccess('analytics'))->toBeFalse();
    });

    it('returns allowed servers list', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Specific Servers Key'
        );

        $key = $result['api_key'];
        $key->update(['server_scopes' => ['commerce']]);

        expect($key->getAllowedServers())->toBe(['commerce']);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Key Revocation
// ─────────────────────────────────────────────────────────────────────────────

describe('Key Revocation', function () {
    it('revokes key via soft delete', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'To Be Revoked'
        );

        $key = $result['api_key'];
        $keyId = $key->id;

        $key->revoke();

        // Should be soft deleted
        expect(ApiKey::find($keyId))->toBeNull();
        expect(ApiKey::withTrashed()->find($keyId))->not->toBeNull();
    });

    it('revoked keys are excluded from workspace scope', function () {
        // Create active key
        ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Active Key'
        );

        // Create and revoke a key
        $revokedResult = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Revoked Key'
        );
        $revokedResult['api_key']->revoke();

        $keys = ApiKey::forWorkspace($this->workspace->id)->get();

        expect($keys)->toHaveCount(1);
        expect($keys->first()->name)->toBe('Active Key');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Masked Key Display
// ─────────────────────────────────────────────────────────────────────────────

describe('Masked Key Display', function () {
    it('provides masked key for display', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Masked Key'
        );

        $key = $result['api_key'];
        $maskedKey = $key->masked_key;

        expect($maskedKey)->toStartWith($key->prefix);
        expect($maskedKey)->toEndWith('_****');
        expect($maskedKey)->toBe("{$key->prefix}_****");
    });

    it('hides raw key in JSON serialization', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Hidden Key'
        );

        $json = $result['api_key']->toArray();

        expect($json)->not->toHaveKey('key');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Relationships
// ─────────────────────────────────────────────────────────────────────────────

describe('Relationships', function () {
    it('belongs to workspace', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Workspace Key'
        );

        expect($result['api_key']->workspace->id)->toBe($this->workspace->id);
    });

    it('belongs to user', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'User Key'
        );

        expect($result['api_key']->user->id)->toBe($this->user->id);
    });

    it('is deleted when workspace is deleted', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Cascade Key'
        );

        $keyId = $result['api_key']->id;

        $this->workspace->delete();

        expect(ApiKey::withTrashed()->find($keyId))->toBeNull();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Factory Tests
// ─────────────────────────────────────────────────────────────────────────────

describe('Factory', function () {
    it('creates key via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->create();

        expect($key)->toBeInstanceOf(ApiKey::class);
        expect($key->workspace_id)->toBe($this->workspace->id);
        expect($key->user_id)->toBe($this->user->id);
    });

    it('creates read-only key via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->readOnly()
            ->create();

        expect($key->scopes)->toBe([ApiKey::SCOPE_READ]);
    });

    it('creates full access key via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->fullAccess()
            ->create();

        expect($key->scopes)->toBe(ApiKey::ALL_SCOPES);
    });

    it('creates expired key via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->expired()
            ->create();

        expect($key->isExpired())->toBeTrue();
    });

    it('creates key with known credentials via helper', function () {
        $result = ApiKeyFactory::createWithPlainKey(
            $this->workspace,
            $this->user,
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        expect($result)->toHaveKeys(['api_key', 'plain_key']);

        // Verify the plain key works for lookup
        $foundKey = ApiKey::findByPlainKey($result['plain_key']);
        expect($foundKey)->not->toBeNull();
        expect($foundKey->id)->toBe($result['api_key']->id);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Rate Limiting (Integration)
// ─────────────────────────────────────────────────────────────────────────────

describe('Rate Limiting Configuration', function () {
    it('has default rate limits configured', function () {
        $default = config('api.rate_limits.default');

        expect($default)->toHaveKeys(['requests', 'per_minutes']);
        expect($default['requests'])->toBeInt();
        expect($default['per_minutes'])->toBeInt();
    });

    it('has authenticated rate limits configured', function () {
        $authenticated = config('api.rate_limits.authenticated');

        expect($authenticated)->toHaveKeys(['requests', 'per_minutes']);
        expect($authenticated['requests'])->toBeGreaterThan(config('api.rate_limits.default.requests'));
    });

    it('has tier-based rate limits configured', function () {
        $tiers = ['starter', 'pro', 'agency', 'enterprise'];

        foreach ($tiers as $tier) {
            $limits = config("api.rate_limits.by_tier.{$tier}");
            expect($limits)->toHaveKeys(['requests', 'per_minutes']);
        }
    });

    it('tier limits increase with tier level', function () {
        $starter = config('api.rate_limits.by_tier.starter.requests');
        $pro = config('api.rate_limits.by_tier.pro.requests');
        $agency = config('api.rate_limits.by_tier.agency.requests');
        $enterprise = config('api.rate_limits.by_tier.enterprise.requests');

        expect($pro)->toBeGreaterThan($starter);
        expect($agency)->toBeGreaterThan($pro);
        expect($enterprise)->toBeGreaterThan($agency);
    });

    it('has route-level rate limit names configured', function () {
        $routeLimits = config('api.rate_limits.routes');

        expect($routeLimits)->toBeArray();
        expect($routeLimits)->toHaveKeys(['mcp', 'pixel']);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// HTTP Authentication Tests
// ─────────────────────────────────────────────────────────────────────────────

describe('HTTP Authentication', function () {
    it('requires authorization header', function () {
        $response = $this->getJson('/api/mcp/servers');

        expect($response->status())->toBe(401);
        expect($response->json('error'))->toBe('unauthorized');
    });

    it('rejects invalid API key', function () {
        $response = $this->getJson('/api/mcp/servers', [
            'Authorization' => 'Bearer hk_invalid_'.str_repeat('x', 48),
        ]);

        expect($response->status())->toBe(401);
    });

    it('rejects expired API key via HTTP', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Expired HTTP Key',
            [ApiKey::SCOPE_READ],
            now()->subDay()
        );

        $response = $this->getJson('/api/mcp/servers', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(401);
    });
});
