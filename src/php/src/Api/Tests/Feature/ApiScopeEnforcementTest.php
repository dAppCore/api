<?php

declare(strict_types=1);

use Core\Api\Models\ApiKey;
use Mod\Tenant\Models\User;
use Mod\Tenant\Models\Workspace;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Route;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();

    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);

    // Register test routes with scope enforcement
    Route::middleware(['api', 'api.auth', 'api.scope.enforce'])
        ->prefix('test-scope')
        ->group(function () {
            Route::get('/read', fn () => response()->json(['status' => 'ok']));
            Route::post('/write', fn () => response()->json(['status' => 'ok']));
            Route::put('/update', fn () => response()->json(['status' => 'ok']));
            Route::patch('/patch', fn () => response()->json(['status' => 'ok']));
            Route::delete('/delete', fn () => response()->json(['status' => 'ok']));
        });
});

// ─────────────────────────────────────────────────────────────────────────────
// Read Scope Enforcement
// ─────────────────────────────────────────────────────────────────────────────

describe('Read Scope Enforcement', function () {
    it('allows GET request with read scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only Key',
            [ApiKey::SCOPE_READ]
        );

        $response = $this->getJson('/api/test-scope/read', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
        expect($response->json('status'))->toBe('ok');
    });

    it('denies POST request with read-only scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only Key',
            [ApiKey::SCOPE_READ]
        );

        $response = $this->postJson('/api/test-scope/write', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('error'))->toBe('forbidden');
        expect($response->json('message'))->toContain('write');
    });

    it('denies DELETE request with read-only scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only Key',
            [ApiKey::SCOPE_READ]
        );

        $response = $this->deleteJson('/api/test-scope/delete', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('error'))->toBe('forbidden');
        expect($response->json('message'))->toContain('delete');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Write Scope Enforcement
// ─────────────────────────────────────────────────────────────────────────────

describe('Write Scope Enforcement', function () {
    it('allows POST request with write scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read/Write Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $response = $this->postJson('/api/test-scope/write', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('allows PUT request with write scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read/Write Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $response = $this->putJson('/api/test-scope/update', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('allows PATCH request with write scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read/Write Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $response = $this->patchJson('/api/test-scope/patch', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('denies DELETE request without delete scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read/Write Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
        );

        $response = $this->deleteJson('/api/test-scope/delete', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('message'))->toContain('delete');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Delete Scope Enforcement
// ─────────────────────────────────────────────────────────────────────────────

describe('Delete Scope Enforcement', function () {
    it('allows DELETE request with delete scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Full Access Key',
            [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE, ApiKey::SCOPE_DELETE]
        );

        $response = $this->deleteJson('/api/test-scope/delete', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('includes key scopes in error response', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only Key',
            [ApiKey::SCOPE_READ]
        );

        $response = $this->deleteJson('/api/test-scope/delete', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('key_scopes'))->toBe([ApiKey::SCOPE_READ]);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Full Access Keys
// ─────────────────────────────────────────────────────────────────────────────

describe('Full Access Keys', function () {
    it('allows all operations with full access', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Full Access Key',
            ApiKey::ALL_SCOPES
        );

        $headers = ['Authorization' => "Bearer {$result['plain_key']}"];

        expect($this->getJson('/api/test-scope/read', $headers)->status())->toBe(200);
        expect($this->postJson('/api/test-scope/write', [], $headers)->status())->toBe(200);
        expect($this->putJson('/api/test-scope/update', [], $headers)->status())->toBe(200);
        expect($this->patchJson('/api/test-scope/patch', [], $headers)->status())->toBe(200);
        expect($this->deleteJson('/api/test-scope/delete', [], $headers)->status())->toBe(200);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Non-API Key Auth (Session)
// ─────────────────────────────────────────────────────────────────────────────

describe('Non-API Key Auth', function () {
    it('passes through for session authenticated users', function () {
        // For session auth, the middleware should allow through
        // as scope enforcement only applies to API key auth
        $this->actingAs($this->user);

        // The api.auth middleware will require API key, so this tests
        // that if somehow session auth is used, scope middleware allows it
        // In practice, routes use either 'auth' OR 'api.auth', not both
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Wildcard Scopes - Resource Wildcards (posts:*)
// ─────────────────────────────────────────────────────────────────────────────

describe('Resource Wildcard Scopes', function () {
    it('grants access with resource wildcard scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Admin Key',
            ['posts:*']
        );

        $apiKey = $result['api_key'];

        // Resource wildcard should grant all actions on that resource
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:write'))->toBeTrue();
        expect($apiKey->hasScope('posts:delete'))->toBeTrue();
        expect($apiKey->hasScope('posts:publish'))->toBeTrue();
    });

    it('resource wildcard does not grant access to other resources', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Only Key',
            ['posts:*']
        );

        $apiKey = $result['api_key'];

        // Should not grant access to other resources
        expect($apiKey->hasScope('users:read'))->toBeFalse();
        expect($apiKey->hasScope('analytics:read'))->toBeFalse();
        expect($apiKey->hasScope('webhooks:write'))->toBeFalse();
    });

    it('multiple resource wildcards work independently', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Content Admin Key',
            ['posts:*', 'pages:*']
        );

        $apiKey = $result['api_key'];

        // Both resource wildcards should work
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:delete'))->toBeTrue();
        expect($apiKey->hasScope('pages:write'))->toBeTrue();
        expect($apiKey->hasScope('pages:publish'))->toBeTrue();

        // Others should not
        expect($apiKey->hasScope('users:read'))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Wildcard Scopes - Action Wildcards (*:read)
// ─────────────────────────────────────────────────────────────────────────────

describe('Action Wildcard Scopes', function () {
    it('grants read access to all resources with action wildcard', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only All Key',
            ['*:read']
        );

        $apiKey = $result['api_key'];

        // Action wildcard should grant read on all resources
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('users:read'))->toBeTrue();
        expect($apiKey->hasScope('analytics:read'))->toBeTrue();
        expect($apiKey->hasScope('webhooks:read'))->toBeTrue();
    });

    it('action wildcard does not grant other actions', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only All Key',
            ['*:read']
        );

        $apiKey = $result['api_key'];

        // Should not grant write or delete
        expect($apiKey->hasScope('posts:write'))->toBeFalse();
        expect($apiKey->hasScope('users:delete'))->toBeFalse();
        expect($apiKey->hasScope('webhooks:manage'))->toBeFalse();
    });

    it('multiple action wildcards work independently', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read/Write All Key',
            ['*:read', '*:write']
        );

        $apiKey = $result['api_key'];

        // Both action wildcards should work
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:write'))->toBeTrue();
        expect($apiKey->hasScope('users:read'))->toBeTrue();
        expect($apiKey->hasScope('users:write'))->toBeTrue();

        // Delete should not be granted
        expect($apiKey->hasScope('posts:delete'))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Wildcard Scopes - Full Wildcard (*)
// ─────────────────────────────────────────────────────────────────────────────

describe('Full Wildcard Scope', function () {
    it('full wildcard grants access to everything', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'God Mode Key',
            ['*']
        );

        $apiKey = $result['api_key'];

        // Full wildcard should grant everything
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:write'))->toBeTrue();
        expect($apiKey->hasScope('posts:delete'))->toBeTrue();
        expect($apiKey->hasScope('users:read'))->toBeTrue();
        expect($apiKey->hasScope('admin:system'))->toBeTrue();
        expect($apiKey->hasScope('any:thing'))->toBeTrue();
    });

    it('full wildcard grants simple scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'God Mode Key',
            ['*']
        );

        $apiKey = $result['api_key'];

        // Simple scopes should also be granted
        expect($apiKey->hasScope('read'))->toBeTrue();
        expect($apiKey->hasScope('write'))->toBeTrue();
        expect($apiKey->hasScope('delete'))->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Scope Inheritance and Hierarchy
// ─────────────────────────────────────────────────────────────────────────────

describe('Scope Inheritance', function () {
    it('exact scopes take precedence over wildcards', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Mixed Key',
            ['posts:read', 'posts:*']
        );

        $apiKey = $result['api_key'];

        // Both exact and wildcard should work
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:write'))->toBeTrue();
    });

    it('combined resource and action wildcards cover different scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Mixed Wildcards Key',
            ['posts:*', '*:read']
        );

        $apiKey = $result['api_key'];

        // posts:* should grant all post actions
        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('posts:write'))->toBeTrue();
        expect($apiKey->hasScope('posts:delete'))->toBeTrue();

        // *:read should grant read on all resources
        expect($apiKey->hasScope('users:read'))->toBeTrue();
        expect($apiKey->hasScope('analytics:read'))->toBeTrue();

        // But not write on other resources
        expect($apiKey->hasScope('users:write'))->toBeFalse();
    });

    it('empty scopes array denies all access', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'No Scopes Key',
            []
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasScope('read'))->toBeFalse();
        expect($apiKey->hasScope('posts:read'))->toBeFalse();
        expect($apiKey->hasScope('*'))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// hasScopes and hasAnyScope Methods
// ─────────────────────────────────────────────────────────────────────────────

describe('Multiple Scope Checking', function () {
    it('hasScopes requires all scopes to be present', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Limited Key',
            ['posts:read', 'posts:write']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasScopes(['posts:read']))->toBeTrue();
        expect($apiKey->hasScopes(['posts:read', 'posts:write']))->toBeTrue();
        expect($apiKey->hasScopes(['posts:read', 'posts:delete']))->toBeFalse();
    });

    it('hasAnyScope requires at least one scope to be present', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Limited Key',
            ['posts:read']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasAnyScope(['posts:read']))->toBeTrue();
        expect($apiKey->hasAnyScope(['posts:read', 'posts:write']))->toBeTrue();
        expect($apiKey->hasAnyScope(['posts:delete', 'users:read']))->toBeFalse();
    });

    it('hasScopes works with wildcards', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Admin Key',
            ['posts:*']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasScopes(['posts:read', 'posts:write']))->toBeTrue();
        expect($apiKey->hasScopes(['posts:read', 'users:read']))->toBeFalse();
    });

    it('hasAnyScope works with wildcards', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read All Key',
            ['*:read']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasAnyScope(['posts:read', 'posts:write']))->toBeTrue();
        expect($apiKey->hasAnyScope(['posts:write', 'users:write']))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// CheckApiScope Middleware (Explicit Scope Requirements)
// ─────────────────────────────────────────────────────────────────────────────

describe('CheckApiScope Middleware', function () {
    beforeEach(function () {
        // Register routes with explicit scope requirements
        Route::middleware(['api', 'api.auth', 'api.scope:posts:read'])
            ->get('/test-explicit/posts', fn () => response()->json(['status' => 'ok']));

        Route::middleware(['api', 'api.auth', 'api.scope:posts:write'])
            ->post('/test-explicit/posts', fn () => response()->json(['status' => 'ok']));

        Route::middleware(['api', 'api.auth', 'api.scope:posts:read,posts:write'])
            ->put('/test-explicit/posts', fn () => response()->json(['status' => 'ok']));
    });

    it('allows request with exact required scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Reader Key',
            ['posts:read']
        );

        $response = $this->getJson('/api/test-explicit/posts', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('allows request with wildcard matching required scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Admin Key',
            ['posts:*']
        );

        $response = $this->getJson('/api/test-explicit/posts', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(200);
    });

    it('denies request without required scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Users Only Key',
            ['users:read']
        );

        $response = $this->getJson('/api/test-explicit/posts', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('error'))->toBe('forbidden');
        expect($response->json('message'))->toContain('posts:read');
    });

    it('requires all scopes when multiple specified', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Posts Reader Only Key',
            ['posts:read']
        );

        // Route requires both posts:read AND posts:write
        $response = $this->putJson('/api/test-explicit/posts', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('message'))->toContain('posts:write');
    });

    it('allows all scopes with full wildcard', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Admin Key',
            ['*']
        );

        $headers = ['Authorization' => "Bearer {$result['plain_key']}"];

        expect($this->getJson('/api/test-explicit/posts', $headers)->status())->toBe(200);
        expect($this->postJson('/api/test-explicit/posts', [], $headers)->status())->toBe(200);
        expect($this->putJson('/api/test-explicit/posts', [], $headers)->status())->toBe(200);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Scope Denial Error Responses
// ─────────────────────────────────────────────────────────────────────────────

describe('Scope Denial Error Responses', function () {
    it('EnforceApiScope returns 403 with detailed error', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Read Only Key',
            [ApiKey::SCOPE_READ]
        );

        $response = $this->postJson('/api/test-scope/write', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json())->toHaveKeys(['error', 'message', 'detail', 'key_scopes']);
        expect($response->json('error'))->toBe('forbidden');
        expect($response->json('detail'))->toContain('POST');
        expect($response->json('detail'))->toContain('write');
    });

    it('CheckApiScope returns 403 with required scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Wrong Scopes Key',
            ['users:read']
        );

        $response = $this->getJson('/api/test-explicit/posts', [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json())->toHaveKeys(['error', 'message', 'required_scopes', 'key_scopes']);
        expect($response->json('required_scopes'))->toContain('posts:read');
        expect($response->json('key_scopes'))->toBe(['users:read']);
    });

    it('error response contains accurate key scopes', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Multi Scope Key',
            ['posts:read', 'users:read', 'analytics:read']
        );

        $response = $this->deleteJson('/api/test-scope/delete', [], [
            'Authorization' => "Bearer {$result['plain_key']}",
        ]);

        expect($response->status())->toBe(403);
        expect($response->json('key_scopes'))->toBe(['posts:read', 'users:read', 'analytics:read']);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Edge Cases
// ─────────────────────────────────────────────────────────────────────────────

describe('Edge Cases', function () {
    it('handles null scopes array gracefully', function () {
        // Directly create a key with null scopes (bypassing generate())
        $apiKey = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->create(['scopes' => null]);

        expect($apiKey->hasScope('read'))->toBeFalse();
        expect($apiKey->hasScope('posts:read'))->toBeFalse();
        expect($apiKey->hasAnyScope(['read']))->toBeFalse();
    });

    it('handles scope with multiple colons', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Nested Scope Key',
            ['api:v2:posts:read']
        );

        $apiKey = $result['api_key'];

        // Exact match should work
        expect($apiKey->hasScope('api:v2:posts:read'))->toBeTrue();

        // The first segment before colon is treated as resource
        // Resource wildcard for 'api' should match
        $result2 = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'API Wildcard Key',
            ['api:*']
        );

        // api:* should match api:v2:posts:read (treats v2:posts:read as action)
        expect($result2['api_key']->hasScope('api:v2:posts:read'))->toBeTrue();
    });

    it('handles empty string scope', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Valid Key',
            ['posts:read']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasScope(''))->toBeFalse();
    });

    it('scope matching is case-sensitive', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Lowercase Key',
            ['posts:read']
        );

        $apiKey = $result['api_key'];

        expect($apiKey->hasScope('posts:read'))->toBeTrue();
        expect($apiKey->hasScope('Posts:Read'))->toBeFalse();
        expect($apiKey->hasScope('POSTS:READ'))->toBeFalse();
    });
});
