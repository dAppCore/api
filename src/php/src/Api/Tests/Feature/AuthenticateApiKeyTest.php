<?php

declare(strict_types=1);

use Core\Api\Models\ApiKey;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;
use Illuminate\Http\Request;
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

    Route::middleware(['api', 'auth.api:read'])
        ->prefix('test-auth')
        ->group(function () {
            Route::get('/scoped', function (Request $request) {
                return response()->json([
                    'auth_type' => $request->attributes->get('auth_type'),
                    'principal' => $request->attributes->get('principal'),
                    'user_id' => $request->attributes->get('userID'),
                ]);
            });
        });
});

it('AuthenticateApiKey_handle_Good authenticates scoped api keys', function () {
    $result = ApiKey::generate(
        $this->workspace->id,
        $this->user->id,
        'Scoped Key',
        [ApiKey::SCOPE_READ]
    );

    $response = $this->getJson('/api/test-auth/scoped', [
        'Authorization' => "Bearer {$result['plain_key']}",
    ]);

    $response->assertOk();
    expect($response->json())->toMatchArray([
        'auth_type' => 'api_key',
        'principal' => 'api-key:'.$result['api_key']->id,
        'user_id' => (string) $this->user->id,
    ]);
    expect($result['api_key']->fresh()->last_used_at)->not->toBeNull();
});

it('AuthenticateApiKey_handle_Good authenticates unscoped bearer tokens', function () {
    $result = $this->user->createToken('Dashboard Token');

    $response = $this->getJson('/api/v1/auth/me', [
        'Authorization' => "Bearer {$result['token']}",
    ]);

    $response->assertOk();
    expect($response->json('auth_type'))->toBe('access_token');
    expect($response->json('user.id'))->toBe($this->user->id);
    expect($result['model']->fresh()->last_used_at)->not->toBeNull();
});

it('AuthenticateApiKey_handle_Bad rejects scoped bearer tokens without api-key scopes', function () {
    $result = $this->user->createToken('Dashboard Token');

    $response = $this->getJson('/api/test-auth/scoped', [
        'Authorization' => "Bearer {$result['token']}",
    ]);

    $response->assertForbidden();
    expect($response->json('error'))->toBe('forbidden');
    expect($response->json('message'))->toBe('Token missing required scope: read');
    expect($result['model']->fresh()->last_used_at)->toBeNull();
});

it('AuthenticateApiKey_handle_Ugly rejects malformed bearer tokens and unauthenticated requests', function () {
    $response = $this->getJson('/api/test-auth/scoped', [
        'Authorization' => 'Bearer hk_invalid_'.str_repeat('x', 48),
    ]);

    $response->assertUnauthorized();
    expect($response->json('error'))->toBe('unauthorized');
    expect($response->json('message'))->toBe('Invalid API key');

    $noAuthResponse = $this->getJson('/api/v1/auth/me');

    $noAuthResponse->assertUnauthorized();
    expect($noAuthResponse->json('error'))->toBe('unauthorized');
    expect($noAuthResponse->json('message'))->toBe('Invalid authentication token');
});
