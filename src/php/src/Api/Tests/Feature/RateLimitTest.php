<?php

declare(strict_types=1);

namespace Core\Api\Tests\Feature;

use Carbon\Carbon;
use Core\LifecycleEventProvider;
use Core\Api\Exceptions\RateLimitExceededException;
use Core\Api\RateLimit\RateLimit;
use Core\Api\RateLimit\RateLimitResult;
use Core\Api\RateLimit\RateLimitService;
use Illuminate\Contracts\Cache\Repository as CacheRepository;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Config;
use Orchestra\Testbench\TestCase;

/**
 * Rate Limiting Tests
 *
 * Tests for the rate limiting service, result DTO, attribute, exception,
 * and configuration.
 */
class RateLimitTest extends TestCase
{
    use RefreshDatabase;

    protected RateLimitService $rateLimitService;

    protected function getPackageProviders($app): array
    {
        return [
            LifecycleEventProvider::class,
        ];
    }

    protected function setUp(): void
    {
        parent::setUp();

        Cache::flush();
        Carbon::setTestNow(Carbon::now());

        $this->rateLimitService = new RateLimitService($this->app->make(CacheRepository::class));
    }

    protected function tearDown(): void
    {
        Carbon::setTestNow();
        parent::tearDown();
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitResult DTO Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_rate_limit_result_creates_allowed_result(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::allowed(100, 99, $resetsAt);

        $this->assertTrue($result->allowed);
        $this->assertSame(100, $result->limit);
        $this->assertSame(99, $result->remaining);
        $this->assertSame(0, $result->retryAfter);
        $this->assertSame($resetsAt->timestamp, $result->resetsAt->timestamp);
    }

    public function test_rate_limit_result_creates_denied_result(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);

        $this->assertFalse($result->allowed);
        $this->assertSame(100, $result->limit);
        $this->assertSame(0, $result->remaining);
        $this->assertSame(30, $result->retryAfter);
        $this->assertSame($resetsAt->timestamp, $result->resetsAt->timestamp);
    }

    public function test_rate_limit_result_generates_correct_headers_for_allowed(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::allowed(100, 99, $resetsAt);

        $headers = $result->headers();

        $this->assertArrayHasKey('X-RateLimit-Limit', $headers);
        $this->assertArrayHasKey('X-RateLimit-Remaining', $headers);
        $this->assertArrayHasKey('X-RateLimit-Reset', $headers);
        $this->assertSame(100, $headers['X-RateLimit-Limit']);
        $this->assertSame(99, $headers['X-RateLimit-Remaining']);
        $this->assertSame($resetsAt->timestamp, $headers['X-RateLimit-Reset']);
        $this->assertArrayNotHasKey('Retry-After', $headers);
    }

    public function test_rate_limit_result_generates_correct_headers_for_denied(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);

        $headers = $result->headers();

        $this->assertArrayHasKey('X-RateLimit-Limit', $headers);
        $this->assertArrayHasKey('X-RateLimit-Remaining', $headers);
        $this->assertArrayHasKey('X-RateLimit-Reset', $headers);
        $this->assertArrayHasKey('Retry-After', $headers);
        $this->assertSame(100, $headers['X-RateLimit-Limit']);
        $this->assertSame(0, $headers['X-RateLimit-Remaining']);
        $this->assertSame(30, $headers['Retry-After']);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitService - Basic Rate Limiting Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_service_allows_requests_under_the_limit(): void
    {
        $result = $this->rateLimitService->hit('test-key', 10, 60);

        $this->assertTrue($result->allowed);
        $this->assertSame(9, $result->remaining);
        $this->assertSame(10, $result->limit);
    }

    public function test_service_tracks_requests_correctly(): void
    {
        // Make 5 requests
        for ($i = 0; $i < 5; $i++) {
            $result = $this->rateLimitService->hit('test-key', 10, 60);
        }

        $this->assertTrue($result->allowed);
        $this->assertSame(5, $result->remaining);
    }

    public function test_service_blocks_requests_when_limit_exceeded(): void
    {
        // Make 10 requests (at limit)
        for ($i = 0; $i < 10; $i++) {
            $this->rateLimitService->hit('test-key', 10, 60);
        }

        // 11th request should be blocked
        $result = $this->rateLimitService->hit('test-key', 10, 60);

        $this->assertFalse($result->allowed);
        $this->assertSame(0, $result->remaining);
        $this->assertGreaterThan(0, $result->retryAfter);
    }

    public function test_check_method_does_not_increment_counter(): void
    {
        // Hit once
        $this->rateLimitService->hit('test-key', 10, 60);

        // Check multiple times (should not count)
        $this->rateLimitService->check('test-key', 10, 60);
        $this->rateLimitService->check('test-key', 10, 60);
        $this->rateLimitService->check('test-key', 10, 60);

        // Verify only 1 hit was recorded
        $this->assertSame(9, $this->rateLimitService->remaining('test-key', 10, 60));
    }

    public function test_service_resets_correctly(): void
    {
        // Make some requests
        for ($i = 0; $i < 5; $i++) {
            $this->rateLimitService->hit('test-key', 10, 60);
        }

        $this->assertSame(5, $this->rateLimitService->remaining('test-key', 10, 60));

        // Reset
        $this->rateLimitService->reset('test-key');

        $this->assertSame(10, $this->rateLimitService->remaining('test-key', 10, 60));
    }

    public function test_service_returns_correct_attempts_count(): void
    {
        $this->assertSame(0, $this->rateLimitService->attempts('test-key', 60));

        $this->rateLimitService->hit('test-key', 10, 60);
        $this->rateLimitService->hit('test-key', 10, 60);
        $this->rateLimitService->hit('test-key', 10, 60);

        $this->assertSame(3, $this->rateLimitService->attempts('test-key', 60));
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitService - Sliding Window Algorithm Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_sliding_window_expires_old_requests(): void
    {
        // Make 5 requests now
        for ($i = 0; $i < 5; $i++) {
            $this->rateLimitService->hit('test-key', 10, 60);
        }

        $this->assertSame(5, $this->rateLimitService->remaining('test-key', 10, 60));

        // Move time forward 61 seconds (past the window)
        Carbon::setTestNow(Carbon::now()->addSeconds(61));

        // Old requests should have expired
        $this->assertSame(10, $this->rateLimitService->remaining('test-key', 10, 60));
    }

    public function test_sliding_window_maintains_requests_within_window(): void
    {
        // Make 5 requests now
        for ($i = 0; $i < 5; $i++) {
            $this->rateLimitService->hit('test-key', 10, 60);
        }

        // Move time forward 30 seconds (still within window)
        Carbon::setTestNow(Carbon::now()->addSeconds(30));

        // Requests should still count
        $this->assertSame(5, $this->rateLimitService->remaining('test-key', 10, 60));

        // Make 3 more requests
        for ($i = 0; $i < 3; $i++) {
            $this->rateLimitService->hit('test-key', 10, 60);
        }

        $this->assertSame(2, $this->rateLimitService->remaining('test-key', 10, 60));
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitService - Burst Allowance Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_burst_allows_when_configured(): void
    {
        // With 20% burst, limit of 10 becomes effective limit of 12
        for ($i = 0; $i < 12; $i++) {
            $result = $this->rateLimitService->hit('test-key', 10, 60, 1.2);
            $this->assertTrue($result->allowed);
        }

        // 13th request should be blocked
        $result = $this->rateLimitService->hit('test-key', 10, 60, 1.2);
        $this->assertFalse($result->allowed);
    }

    public function test_burst_reports_base_limit_not_burst_limit(): void
    {
        $result = $this->rateLimitService->hit('test-key', 10, 60, 1.5);

        // Limit shown should be the base limit (10), not the burst limit (15)
        $this->assertSame(10, $result->limit);
    }

    public function test_burst_calculates_remaining_based_on_burst_limit(): void
    {
        // With 50% burst, limit of 10 becomes effective limit of 15
        $result = $this->rateLimitService->hit('test-key', 10, 60, 1.5);

        // After 1 hit, remaining should be 14 (15 - 1)
        $this->assertSame(14, $result->remaining);
    }

    public function test_burst_works_without_burst(): void
    {
        for ($i = 0; $i < 10; $i++) {
            $result = $this->rateLimitService->hit('test-key', 10, 60, 1.0);
            $this->assertTrue($result->allowed);
        }

        $result = $this->rateLimitService->hit('test-key', 10, 60, 1.0);
        $this->assertFalse($result->allowed);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitService - Key Builders Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_builds_endpoint_keys_correctly(): void
    {
        $key = $this->rateLimitService->buildEndpointKey('api_key:123', 'users.index');
        $this->assertSame('endpoint:api_key:123:users.index', $key);
    }

    public function test_builds_workspace_keys_correctly(): void
    {
        $key = $this->rateLimitService->buildWorkspaceKey(456);
        $this->assertSame('workspace:456', $key);

        $keyWithSuffix = $this->rateLimitService->buildWorkspaceKey(456, 'users.index');
        $this->assertSame('workspace:456:users.index', $keyWithSuffix);
    }

    public function test_builds_api_key_keys_correctly(): void
    {
        $key = $this->rateLimitService->buildApiKeyKey(789);
        $this->assertSame('api_key:789', $key);

        $keyWithSuffix = $this->rateLimitService->buildApiKeyKey(789, 'users.index');
        $this->assertSame('api_key:789:users.index', $keyWithSuffix);
    }

    public function test_builds_ip_keys_correctly(): void
    {
        $key = $this->rateLimitService->buildIpKey('192.168.1.1');
        $this->assertSame('ip:192.168.1.1', $key);

        $keyWithSuffix = $this->rateLimitService->buildIpKey('192.168.1.1', 'users.index');
        $this->assertSame('ip:192.168.1.1:users.index', $keyWithSuffix);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimit Attribute Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_attribute_instantiates_with_required_parameters(): void
    {
        $attribute = new RateLimit(limit: 100);

        $this->assertSame(100, $attribute->limit);
        $this->assertSame(60, $attribute->window); // default
        $this->assertSame(1.0, $attribute->burst); // default
        $this->assertNull($attribute->key); // default
    }

    public function test_attribute_instantiates_with_all_parameters(): void
    {
        $attribute = new RateLimit(
            limit: 200,
            window: 120,
            burst: 1.5,
            key: 'custom-key'
        );

        $this->assertSame(200, $attribute->limit);
        $this->assertSame(120, $attribute->window);
        $this->assertSame(1.5, $attribute->burst);
        $this->assertSame('custom-key', $attribute->key);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // RateLimitExceededException Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_exception_creates_with_rate_limit_result(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);
        $exception = new RateLimitExceededException($result);

        $this->assertSame(429, $exception->getStatusCode());
        $this->assertSame($result, $exception->getRateLimitResult());
    }

    public function test_exception_renders_as_json_response(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);
        $exception = new RateLimitExceededException($result);

        $response = $exception->render();

        $this->assertSame(429, $response->getStatusCode());

        $content = json_decode($response->getContent(), true);
        $this->assertSame('rate_limit_exceeded', $content['error']);
        $this->assertSame(30, $content['retry_after']);
        $this->assertSame(100, $content['limit']);
    }

    public function test_exception_includes_rate_limit_headers_in_response(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);
        $exception = new RateLimitExceededException($result);

        $response = $exception->render();

        $this->assertSame('100', $response->headers->get('X-RateLimit-Limit'));
        $this->assertSame('0', $response->headers->get('X-RateLimit-Remaining'));
        $this->assertSame('30', $response->headers->get('Retry-After'));
    }

    public function test_exception_allows_custom_message(): void
    {
        $resetsAt = Carbon::now()->addMinute();
        $result = RateLimitResult::denied(100, 30, $resetsAt);
        $exception = new RateLimitExceededException($result, 'Custom rate limit message');

        $response = $exception->render();
        $content = json_decode($response->getContent(), true);

        $this->assertSame('Custom rate limit message', $content['message']);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // Per-Workspace Rate Limiting Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_isolates_rate_limits_by_workspace(): void
    {
        // Create two different workspace keys
        $key1 = $this->rateLimitService->buildWorkspaceKey(1, 'endpoint');
        $key2 = $this->rateLimitService->buildWorkspaceKey(2, 'endpoint');

        // Hit rate limit for workspace 1
        for ($i = 0; $i < 10; $i++) {
            $this->rateLimitService->hit($key1, 10, 60);
        }

        // Workspace 1 should be blocked
        $result1 = $this->rateLimitService->hit($key1, 10, 60);
        $this->assertFalse($result1->allowed);

        // Workspace 2 should still be allowed
        $result2 = $this->rateLimitService->hit($key2, 10, 60);
        $this->assertTrue($result2->allowed);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // Rate Limit Configuration Tests
    // ─────────────────────────────────────────────────────────────────────────

    public function test_config_has_enabled_flag(): void
    {
        Config::set('api.rate_limits.enabled', true);
        $this->assertTrue(config('api.rate_limits.enabled'));
    }

    public function test_config_has_default_limits(): void
    {
        Config::set('api.rate_limits.default', [
            'limit' => 60,
            'window' => 60,
            'burst' => 1.0,
        ]);

        $default = config('api.rate_limits.default');

        $this->assertArrayHasKey('limit', $default);
        $this->assertArrayHasKey('window', $default);
        $this->assertArrayHasKey('burst', $default);
    }

    public function test_config_has_authenticated_limits(): void
    {
        Config::set('api.rate_limits.authenticated', [
            'limit' => 1000,
            'window' => 60,
            'burst' => 1.2,
        ]);

        $authenticated = config('api.rate_limits.authenticated');

        $this->assertArrayHasKey('limit', $authenticated);
        $this->assertSame(1000, $authenticated['limit']);
    }

    public function test_config_has_per_workspace_flag(): void
    {
        Config::set('api.rate_limits.per_workspace', true);
        $this->assertTrue(config('api.rate_limits.per_workspace'));
    }

    public function test_config_has_endpoints_configuration(): void
    {
        Config::set('api.rate_limits.endpoints', []);
        $this->assertIsArray(config('api.rate_limits.endpoints'));
    }

    public function test_config_has_tier_based_limits(): void
    {
        Config::set('api.rate_limits.tiers', [
            'free' => ['limit' => 60, 'window' => 60, 'burst' => 1.0],
            'starter' => ['limit' => 1000, 'window' => 60, 'burst' => 1.2],
            'pro' => ['limit' => 5000, 'window' => 60, 'burst' => 1.3],
            'agency' => ['limit' => 20000, 'window' => 60, 'burst' => 1.5],
            'enterprise' => ['limit' => 100000, 'window' => 60, 'burst' => 2.0],
        ]);

        $tiers = config('api.rate_limits.tiers');

        $this->assertArrayHasKey('free', $tiers);
        $this->assertArrayHasKey('starter', $tiers);
        $this->assertArrayHasKey('pro', $tiers);
        $this->assertArrayHasKey('agency', $tiers);
        $this->assertArrayHasKey('enterprise', $tiers);

        foreach ($tiers as $tier => $tierConfig) {
            $this->assertArrayHasKey('limit', $tierConfig);
            $this->assertArrayHasKey('window', $tierConfig);
            $this->assertArrayHasKey('burst', $tierConfig);
        }
    }

    public function test_tier_limits_increase_with_tier_level(): void
    {
        Config::set('api.rate_limits.tiers', [
            'free' => ['limit' => 60, 'window' => 60, 'burst' => 1.0],
            'starter' => ['limit' => 1000, 'window' => 60, 'burst' => 1.2],
            'pro' => ['limit' => 5000, 'window' => 60, 'burst' => 1.3],
            'agency' => ['limit' => 20000, 'window' => 60, 'burst' => 1.5],
            'enterprise' => ['limit' => 100000, 'window' => 60, 'burst' => 2.0],
        ]);

        $tiers = config('api.rate_limits.tiers');

        $this->assertGreaterThan($tiers['free']['limit'], $tiers['starter']['limit']);
        $this->assertGreaterThan($tiers['starter']['limit'], $tiers['pro']['limit']);
        $this->assertGreaterThan($tiers['pro']['limit'], $tiers['agency']['limit']);
        $this->assertGreaterThan($tiers['agency']['limit'], $tiers['enterprise']['limit']);
    }

    public function test_higher_tiers_have_higher_burst_allowance(): void
    {
        Config::set('api.rate_limits.tiers', [
            'free' => ['limit' => 60, 'window' => 60, 'burst' => 1.0],
            'pro' => ['limit' => 5000, 'window' => 60, 'burst' => 1.3],
            'agency' => ['limit' => 20000, 'window' => 60, 'burst' => 1.5],
            'enterprise' => ['limit' => 100000, 'window' => 60, 'burst' => 2.0],
        ]);

        $tiers = config('api.rate_limits.tiers');

        $this->assertGreaterThanOrEqual($tiers['pro']['burst'], $tiers['agency']['burst']);
        $this->assertGreaterThanOrEqual($tiers['agency']['burst'], $tiers['enterprise']['burst']);
    }
}
