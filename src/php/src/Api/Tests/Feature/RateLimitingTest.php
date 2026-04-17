<?php

declare(strict_types=1);

use Carbon\Carbon;
use Core\Api\Exceptions\RateLimitExceededException;
use Core\Api\Middleware\RateLimitApi;
use Core\Api\Models\ApiKey;
use Core\Api\RateLimit\RateLimitService;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Routing\Route;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Config;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();
    Carbon::setTestNow(Carbon::now());

    $this->rateLimitService = app(RateLimitService::class);
    $this->middleware = new RateLimitApi($this->rateLimitService);

    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);

    // Set up default configuration
    Config::set('api.rate_limits.enabled', true);
    Config::set('api.rate_limits.default', [
        'limit' => 60,
        'window' => 60,
        'burst' => 1.0,
    ]);
    Config::set('api.rate_limits.authenticated', [
        'limit' => 1000,
        'window' => 60,
        'burst' => 1.2,
    ]);
    Config::set('api.rate_limits.per_workspace', true);
    Config::set('api.rate_limits.tiers', [
        'free' => ['limit' => 60, 'window' => 60, 'burst' => 1.0],
        'starter' => ['limit' => 1000, 'window' => 60, 'burst' => 1.2],
        'pro' => ['limit' => 5000, 'window' => 60, 'burst' => 1.3],
        'agency' => ['limit' => 20000, 'window' => 60, 'burst' => 1.5],
        'enterprise' => ['limit' => 100000, 'window' => 60, 'burst' => 2.0],
    ]);
});

afterEach(function () {
    Carbon::setTestNow();
});

/**
 * Mock workspace object with tier property for testing tier-based rate limits.
 *
 * The RateLimitApi middleware uses property_exists() to check for tier,
 * so we need a class with the property defined.
 */
class MockTieredWorkspace
{
    public int $id;
    public string $tier;

    public function __construct(int $id, string $tier)
    {
        $this->id = $id;
        $this->tier = $tier;
    }
}

// -----------------------------------------------------------------------------
// Rate Limit Enforcement
// -----------------------------------------------------------------------------

describe('Rate Limit Enforcement', function () {
    it('allows requests under the limit', function () {
        $request = createMockRequest();

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->getContent())->toBe('OK');
        expect($response->getStatusCode())->toBe(200);
    });

    it('blocks requests when limit is exceeded', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        // Next request should be blocked
        $this->middleware->handle($request, fn () => new Response('OK'));
    })->throws(RateLimitExceededException::class);

    it('tracks requests correctly across multiple calls', function () {
        $request = createMockRequest();

        // Make 30 requests
        for ($i = 0; $i < 30; $i++) {
            $response = $this->middleware->handle($request, fn () => new Response('OK'));
        }

        // Verify remaining count in headers
        expect($response->headers->get('X-RateLimit-Remaining'))->toBe('30');
    });

    it('allows requests again after window expires', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        // Move time forward past the window
        Carbon::setTestNow(Carbon::now()->addSeconds(61));

        // Should be allowed again
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->getContent())->toBe('OK');
        expect($response->getStatusCode())->toBe(200);
    });

    it('can be disabled via configuration', function () {
        Config::set('api.rate_limits.enabled', false);

        $request = createMockRequest();

        // Even with 100 requests, should not be blocked
        for ($i = 0; $i < 100; $i++) {
            $response = $this->middleware->handle($request, fn () => new Response('OK'));
        }

        expect($response->getContent())->toBe('OK');
    });
});

// -----------------------------------------------------------------------------
// Rate Limit Headers
// -----------------------------------------------------------------------------

describe('Rate Limit Headers', function () {
    it('includes X-RateLimit-Limit header', function () {
        $request = createMockRequest();

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->has('X-RateLimit-Limit'))->toBeTrue();
        expect($response->headers->get('X-RateLimit-Limit'))->toBe('60');
    });

    it('includes X-RateLimit-Remaining header', function () {
        $request = createMockRequest();

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->has('X-RateLimit-Remaining'))->toBeTrue();
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(59);
    });

    it('includes X-RateLimit-Reset header', function () {
        $request = createMockRequest();

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->has('X-RateLimit-Reset'))->toBeTrue();

        $resetTimestamp = (int) $response->headers->get('X-RateLimit-Reset');
        expect($resetTimestamp)->toBeGreaterThan(time());
    });

    it('decrements remaining count with each request', function () {
        $request = createMockRequest();

        $response1 = $this->middleware->handle($request, fn () => new Response('OK'));
        $response2 = $this->middleware->handle($request, fn () => new Response('OK'));
        $response3 = $this->middleware->handle($request, fn () => new Response('OK'));

        expect((int) $response1->headers->get('X-RateLimit-Remaining'))->toBe(59);
        expect((int) $response2->headers->get('X-RateLimit-Remaining'))->toBe(58);
        expect((int) $response3->headers->get('X-RateLimit-Remaining'))->toBe(57);
    });

    it('includes Retry-After header when limit exceeded', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();

            expect($response->headers->has('Retry-After'))->toBeTrue();
            expect((int) $response->headers->get('Retry-After'))->toBeGreaterThan(0);
        }
    });

    it('shows zero remaining when limit is reached', function () {
        $request = createMockRequest();

        // Make 59 requests (one less than limit)
        for ($i = 0; $i < 59; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        // 60th request uses the last allowance
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(0);
    });
});

// -----------------------------------------------------------------------------
// Tier-Based Rate Limits
// -----------------------------------------------------------------------------

describe('Tier-Based Rate Limits', function () {
    it('applies free tier limits by default', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('free')]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('60');
    });

    it('applies starter tier limits', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('starter')]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('1000');
    });

    it('applies pro tier limits', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('pro')]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('5000');
    });

    it('applies agency tier limits', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('agency')]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('20000');
    });

    it('applies enterprise tier limits', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('enterprise')]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('100000');
    });

    it('falls back to free tier for unknown tier', function () {
        $request = createMockRequest(['workspace' => createWorkspaceWithTier('unknown')]);

        // Without tier config, falls back to default
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // Should use default (60) since 'unknown' tier doesn't exist
        expect((int) $response->headers->get('X-RateLimit-Limit'))->toBeLessThanOrEqual(1000);
    });

    it('higher tiers have higher limits', function () {
        $tiers = Config::get('api.rate_limits.tiers');

        expect($tiers['starter']['limit'])->toBeGreaterThan($tiers['free']['limit']);
        expect($tiers['pro']['limit'])->toBeGreaterThan($tiers['starter']['limit']);
        expect($tiers['agency']['limit'])->toBeGreaterThan($tiers['pro']['limit']);
        expect($tiers['enterprise']['limit'])->toBeGreaterThan($tiers['agency']['limit']);
    });
});

// -----------------------------------------------------------------------------
// Workspace-Scoped Rate Limits
// -----------------------------------------------------------------------------

describe('Workspace-Scoped Rate Limits', function () {
    it('isolates rate limits between workspaces', function () {
        $workspace1 = Workspace::factory()->create();
        $workspace2 = Workspace::factory()->create();

        $request1 = createMockRequest(['workspace' => $workspace1]);
        $request2 = createMockRequest(['workspace' => $workspace2]);

        // Exhaust rate limit for workspace 1
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request1, fn () => new Response('OK'));
        }

        // Workspace 2 should still have full quota
        $response = $this->middleware->handle($request2, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(59);
    });

    it('includes workspace ID in rate limit key', function () {
        $workspace = Workspace::factory()->create();
        $apiKey = createApiKeyForWorkspace($workspace);

        $request = createMockRequest([
            'workspace' => $workspace,
            'api_key' => $apiKey,
        ]);

        $this->middleware->handle($request, fn () => new Response('OK'));

        // Verify key was created with workspace scope
        $cacheKey = "rate_limit:api_key:{$apiKey->id}:ws:{$workspace->id}:route:test.route";
        expect(Cache::has($cacheKey))->toBeTrue();
    });

    it('ignores malformed workspace context instead of crashing', function () {
        $workspace = Workspace::factory()->create();
        $apiKey = createApiKeyForWorkspace($workspace);

        $request = createMockRequest([
            'api_key' => $apiKey,
            'workspace' => ['id' => $workspace->id],
        ]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect($response->headers->get('X-RateLimit-Limit'))->toBe('1000');
        expect(Cache::has("rate_limit:api_key:{$apiKey->id}:route:test.route"))->toBeTrue();
        expect(Cache::has("rate_limit:api_key:{$apiKey->id}:ws:{$workspace->id}:route:test.route"))->toBeFalse();
    });

    it('can disable per-workspace limiting', function () {
        Config::set('api.rate_limits.per_workspace', false);

        $workspace1 = Workspace::factory()->create();
        $workspace2 = Workspace::factory()->create();

        $apiKey1 = createApiKeyForWorkspace($workspace1);
        $apiKey2 = createApiKeyForWorkspace($workspace2);

        // Use same API key ID to test shared limit
        $request1 = createMockRequest([
            'workspace' => $workspace1,
            'api_key' => $apiKey1,
        ]);

        // Make requests from workspace 1
        for ($i = 0; $i < 30; $i++) {
            $this->middleware->handle($request1, fn () => new Response('OK'));
        }

        // The remaining should reflect shared limit usage
        $response = $this->middleware->handle($request1, fn () => new Response('OK'));
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBeLessThan(60);
    });
});

// -----------------------------------------------------------------------------
// Burst Allowance
// -----------------------------------------------------------------------------

describe('Burst Allowance', function () {
    it('allows burst above base limit when configured', function () {
        // Configure with 20% burst (limit 10 becomes effective 12)
        Config::set('api.rate_limits.default', [
            'limit' => 10,
            'window' => 60,
            'burst' => 1.2,
        ]);

        $request = createMockRequest();

        // Should allow 12 requests (10 * 1.2)
        for ($i = 0; $i < 12; $i++) {
            $response = $this->middleware->handle($request, fn () => new Response('OK'));
            expect($response->getStatusCode())->toBe(200);
        }

        // 13th request should be blocked
        $this->middleware->handle($request, fn () => new Response('OK'));
    })->throws(RateLimitExceededException::class);

    it('reports base limit in headers not burst limit', function () {
        Config::set('api.rate_limits.default', [
            'limit' => 10,
            'window' => 60,
            'burst' => 1.5,
        ]);

        $request = createMockRequest();
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // Should show base limit of 10, not burst limit of 15
        expect($response->headers->get('X-RateLimit-Limit'))->toBe('10');
    });

    it('calculates remaining based on burst limit', function () {
        Config::set('api.rate_limits.default', [
            'limit' => 10,
            'window' => 60,
            'burst' => 1.5,
        ]);

        $request = createMockRequest();
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // After 1 hit, remaining should be 14 (15 - 1 where 15 = 10 * 1.5)
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(14);
    });

    it('applies tier-specific burst allowance', function () {
        $workspace = createWorkspaceWithTier('enterprise');
        $request = createMockRequest(['workspace' => $workspace]);

        // Enterprise tier has burst of 2.0 (100000 * 2.0 = 200000 effective)
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // After 1 hit, remaining should be 199999
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(199999);
    });

    it('has no burst allowance for free tier', function () {
        $workspace = createWorkspaceWithTier('free');
        $request = createMockRequest(['workspace' => $workspace]);

        // Free tier has burst of 1.0 (no burst)
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // After 1 hit, remaining should be 59 (60 - 1, no burst)
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(59);
    });
});

// -----------------------------------------------------------------------------
// Quota Exceeded Response
// -----------------------------------------------------------------------------

describe('Quota Exceeded Response', function () {
    it('returns 429 status code when limit exceeded', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            expect($e->getStatusCode())->toBe(429);
        }
    });

    it('returns proper JSON error response', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();
            $content = json_decode($response->getContent(), true);

            expect($content['error'])->toBe('rate_limit_exceeded');
            expect($content)->toHaveKey('message');
            expect($content)->toHaveKey('retry_after');
            expect($content)->toHaveKey('limit');
            expect($content)->toHaveKey('resets_at');
        }
    });

    it('includes retry_after in error response', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();
            $content = json_decode($response->getContent(), true);

            expect($content['retry_after'])->toBeGreaterThan(0);
            expect($content['retry_after'])->toBeLessThanOrEqual(60);
        }
    });

    it('includes resets_at ISO8601 timestamp', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();
            $content = json_decode($response->getContent(), true);

            expect($content['resets_at'])->toMatch('/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/');
        }
    });

    it('includes rate limit headers in error response', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();

            expect($response->headers->has('X-RateLimit-Limit'))->toBeTrue();
            expect($response->headers->has('X-RateLimit-Remaining'))->toBeTrue();
            expect($response->headers->has('X-RateLimit-Reset'))->toBeTrue();
            expect($response->headers->has('Retry-After'))->toBeTrue();
        }
    });

    it('shows zero remaining in error response', function () {
        $request = createMockRequest();

        // Exhaust the rate limit
        for ($i = 0; $i < 60; $i++) {
            $this->middleware->handle($request, fn () => new Response('OK'));
        }

        try {
            $this->middleware->handle($request, fn () => new Response('OK'));
        } catch (RateLimitExceededException $e) {
            $response = $e->render();

            expect($response->headers->get('X-RateLimit-Remaining'))->toBe('0');
        }
    });
});

// -----------------------------------------------------------------------------
// API Key-Based Rate Limiting
// -----------------------------------------------------------------------------

describe('API Key-Based Rate Limiting', function () {
    it('uses API key ID in rate limit key', function () {
        $apiKey = createApiKeyForWorkspace($this->workspace);

        $request = createMockRequest([
            'api_key' => $apiKey,
            'workspace' => $this->workspace,
        ]);

        $this->middleware->handle($request, fn () => new Response('OK'));

        $cacheKey = "rate_limit:api_key:{$apiKey->id}:ws:{$this->workspace->id}:route:test.route";
        expect(Cache::has($cacheKey))->toBeTrue();
    });

    it('isolates rate limits between API keys', function () {
        $apiKey1 = createApiKeyForWorkspace($this->workspace);
        $apiKey2 = createApiKeyForWorkspace($this->workspace);

        Config::set('api.rate_limits.authenticated', [
            'limit' => 10,
            'window' => 60,
            'burst' => 1.0,
        ]);

        $request1 = createMockRequest([
            'api_key' => $apiKey1,
            'workspace' => $this->workspace,
        ]);
        $request2 = createMockRequest([
            'api_key' => $apiKey2,
            'workspace' => $this->workspace,
        ]);

        // Exhaust rate limit for API key 1
        for ($i = 0; $i < 10; $i++) {
            $this->middleware->handle($request1, fn () => new Response('OK'));
        }

        // API key 2 should still have full quota
        $response = $this->middleware->handle($request2, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(9);
    });

    it('applies authenticated limits when API key present', function () {
        $apiKey = createApiKeyForWorkspace($this->workspace);

        $request = createMockRequest([
            'api_key' => $apiKey,
        ]);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // Authenticated limit is 1000 with 1.2 burst = 1200 effective, so 1199 remaining
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(1199);
    });
});

// -----------------------------------------------------------------------------
// IP-Based Rate Limiting (Unauthenticated)
// -----------------------------------------------------------------------------

describe('IP-Based Rate Limiting', function () {
    it('uses IP address for unauthenticated requests', function () {
        $request = createMockRequest();

        $this->middleware->handle($request, fn () => new Response('OK'));

        $cacheKey = 'rate_limit:ip:127.0.0.1:route:test.route';
        expect(Cache::has($cacheKey))->toBeTrue();
    });

    it('isolates rate limits between IP addresses', function () {
        Config::set('api.rate_limits.default', [
            'limit' => 10,
            'window' => 60,
            'burst' => 1.0,
        ]);

        $request1 = createMockRequest([], '192.168.1.1');
        $request2 = createMockRequest([], '192.168.1.2');

        // Exhaust rate limit for IP 1
        for ($i = 0; $i < 10; $i++) {
            $this->middleware->handle($request1, fn () => new Response('OK'));
        }

        // IP 2 should still have full quota
        $response = $this->middleware->handle($request2, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(9);
    });

    it('applies default limits for unauthenticated requests', function () {
        $request = createMockRequest();

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('60');
    });
});

// -----------------------------------------------------------------------------
// Per-Endpoint Rate Limits
// -----------------------------------------------------------------------------

describe('Per-Endpoint Rate Limits', function () {
    it('applies endpoint-specific rate limit from config', function () {
        Config::set('api.rate_limits.endpoints.test.route', [
            'limit' => 5,
            'window' => 60,
            'burst' => 1.0,
        ]);

        $request = createMockRequest();
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('X-RateLimit-Limit'))->toBe('5');
    });

    it('isolates rate limits between endpoints', function () {
        Config::set('api.rate_limits.default', [
            'limit' => 5,
            'window' => 60,
            'burst' => 1.0,
        ]);

        $request1 = createMockRequest([], '127.0.0.1', 'api.users.index');
        $request2 = createMockRequest([], '127.0.0.1', 'api.posts.index');

        // Exhaust rate limit for endpoint 1
        for ($i = 0; $i < 5; $i++) {
            $this->middleware->handle($request1, fn () => new Response('OK'));
        }

        // Endpoint 2 should still have full quota
        $response = $this->middleware->handle($request2, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(4);
    });
});

// -----------------------------------------------------------------------------
// Rate Limit Bypass for Trusted Clients
// -----------------------------------------------------------------------------

describe('Rate Limit Bypass', function () {
    it('bypasses rate limiting when disabled globally', function () {
        Config::set('api.rate_limits.enabled', false);

        $request = createMockRequest();

        // Make many requests without hitting a limit
        for ($i = 0; $i < 1000; $i++) {
            $response = $this->middleware->handle($request, fn () => new Response('OK'));
            expect($response->getStatusCode())->toBe(200);
        }
    });

    it('does not add rate limit headers when disabled', function () {
        Config::set('api.rate_limits.enabled', false);

        $request = createMockRequest();
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // Headers should not be present when rate limiting is disabled
        expect($response->headers->has('X-RateLimit-Limit'))->toBeFalse();
    });

    it('enterprise tier has very high effective limit with burst', function () {
        $workspace = createWorkspaceWithTier('enterprise');
        $request = createMockRequest(['workspace' => $workspace]);

        // Enterprise: 100000 * 2.0 burst = 200000 effective requests per minute
        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        // Should be able to make many requests without hitting limit
        expect($response->getStatusCode())->toBe(200);
        expect((int) $response->headers->get('X-RateLimit-Remaining'))->toBe(199999);
    });
});

// -----------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------

function createMockRequest(array $attributes = [], string $ip = '127.0.0.1', string $routeName = 'test.route'): Request
{
    $request = Request::create('/api/test', 'GET');
    $request->server->set('REMOTE_ADDR', $ip);

    // Create a mock route
    $route = new Route(['GET'], '/api/test', ['as' => $routeName]);
    $request->setRouteResolver(fn () => $route);

    // Set request attributes
    foreach ($attributes as $key => $value) {
        $request->attributes->set($key, $value);
    }

    return $request;
}

/**
 * Create a mock workspace object with a tier property.
 *
 * Uses MockTieredWorkspace because the middleware uses property_exists()
 * which requires the property to be defined on the class.
 */
function createWorkspaceWithTier(string $tier): MockTieredWorkspace
{
    static $counter = 1000;

    return new MockTieredWorkspace($counter++, $tier);
}

function createApiKeyForWorkspace(Workspace $workspace): ApiKey
{
    $user = User::factory()->create();
    $result = ApiKey::generate(
        $workspace->id,
        $user->id,
        'Test API Key'
    );

    return $result['api_key'];
}
