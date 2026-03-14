<?php

/**
 * API Configuration
 *
 * Rate limiting, versioning, and API-specific settings.
 * Integrated with EntitlementService for tier-based rate limits.
 */

return [

    /*
    |--------------------------------------------------------------------------
    | API Version
    |--------------------------------------------------------------------------
    |
    | The current API version. Used in URL prefix and documentation.
    |
    */

    'version' => env('API_VERSION', '1'),

    /*
    |--------------------------------------------------------------------------
    | Rate Limiting
    |--------------------------------------------------------------------------
    |
    | Configure rate limits for API requests.
    |
    | Features:
    | - Per-endpoint limits via 'endpoints' config or #[RateLimit] attribute
    | - Per-workspace limits (when 'per_workspace' is true)
    | - Tier-based limits based on workspace subscription
    | - Burst allowance for temporary traffic spikes
    | - Sliding window algorithm for smoother rate limiting
    |
    | Priority (highest to lowest):
    | 1. Method-level #[RateLimit] attribute
    | 2. Class-level #[RateLimit] attribute
    | 3. Per-endpoint config (endpoints.{route_name})
    | 4. Tier-based limits (tiers.{tier})
    | 5. Authenticated limits
    | 6. Default limits
    |
    */

    'rate_limits' => [
        // Enable/disable rate limiting globally
        'enabled' => env('API_RATE_LIMITING_ENABLED', true),

        // Unauthenticated requests (by IP)
        'default' => [
            'limit' => 60,
            'window' => 60, // seconds
            'burst' => 1.0, // no burst allowance for unauthenticated
            // Legacy support
            'requests' => 60,
            'per_minutes' => 1,
        ],

        // Authenticated requests (by user/key)
        'authenticated' => [
            'limit' => 1000,
            'window' => 60, // seconds
            'burst' => 1.2, // 20% burst allowance
            // Legacy support
            'requests' => 1000,
            'per_minutes' => 1,
        ],

        // Enable per-workspace rate limiting (isolates limits by workspace)
        'per_workspace' => true,

        // Per-endpoint rate limits (route names)
        // Example: 'users.index' => ['limit' => 100, 'window' => 60]
        'endpoints' => [
            // High-volume endpoints may need higher limits
            // 'links.index' => ['limit' => 500, 'window' => 60],
            // 'qrcodes.index' => ['limit' => 500, 'window' => 60],

            // Sensitive endpoints may need lower limits
            // 'auth.login' => ['limit' => 10, 'window' => 60],
            // 'keys.create' => ['limit' => 20, 'window' => 60],
        ],

        // Tier-based limits (based on workspace subscription/plan)
        'tiers' => [
            'free' => [
                'limit' => 60,
                'window' => 60, // seconds
                'burst' => 1.0,
            ],
            'starter' => [
                'limit' => 1000,
                'window' => 60,
                'burst' => 1.2,
            ],
            'pro' => [
                'limit' => 5000,
                'window' => 60,
                'burst' => 1.3,
            ],
            'agency' => [
                'limit' => 20000,
                'window' => 60,
                'burst' => 1.5,
            ],
            'enterprise' => [
                'limit' => 100000,
                'window' => 60,
                'burst' => 2.0,
            ],
        ],

        // Legacy: Tier-based limits (deprecated, use 'tiers' instead)
        'by_tier' => [
            'starter' => [
                'requests' => 1000,
                'per_minutes' => 1,
            ],
            'pro' => [
                'requests' => 5000,
                'per_minutes' => 1,
            ],
            'agency' => [
                'requests' => 20000,
                'per_minutes' => 1,
            ],
            'enterprise' => [
                'requests' => 100000,
                'per_minutes' => 1,
            ],
        ],

        // Route-specific rate limiters (for named routes)
        'routes' => [
            'mcp' => 'authenticated',
            'pixel' => 'default',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Usage Alerts
    |--------------------------------------------------------------------------
    |
    | Configure notifications when API usage approaches limits.
    |
    | Thresholds define percentages of rate limit that trigger alerts:
    | - warning: First alert level (default: 80%)
    | - critical: Urgent alert level (default: 95%)
    |
    | Cooldown prevents duplicate notifications for the same level.
    |
    */

    'alerts' => [
        // Enable/disable usage alerting
        'enabled' => env('API_USAGE_ALERTS_ENABLED', true),

        // Alert thresholds (percentage of rate limit)
        'thresholds' => [
            'warning' => 80,
            'critical' => 95,
        ],

        // Hours between notifications of the same level
        'cooldown_hours' => 6,
    ],

    /*
    |--------------------------------------------------------------------------
    | API Key Settings
    |--------------------------------------------------------------------------
    |
    | Configuration for API key generation and validation.
    |
    */

    'keys' => [
        // Prefix for all API keys
        'prefix' => 'hk_',

        // Default scopes for new API keys
        'default_scopes' => ['read', 'write'],

        // Maximum API keys per workspace
        'max_per_workspace' => 10,

        // Auto-expire keys after this many days (null = never)
        'default_expiry_days' => null,
    ],

    /*
    |--------------------------------------------------------------------------
    | Webhooks
    |--------------------------------------------------------------------------
    |
    | Webhook delivery settings.
    |
    */

    'webhooks' => [
        // Maximum webhook endpoints per workspace
        'max_per_workspace' => 5,

        // Timeout for webhook delivery in seconds
        'timeout' => 30,

        // Max retries for failed deliveries
        'max_retries' => 5,

        // Disable endpoint after this many consecutive failures
        'disable_after_failures' => 10,

        // Events that are high-volume and opt-in only
        'high_volume_events' => [
            'link.clicked',
            'qrcode.scanned',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Pagination
    |--------------------------------------------------------------------------
    |
    | Default pagination settings for API responses.
    |
    */

    'pagination' => [
        'default_per_page' => 25,
        'max_per_page' => 100,
    ],

];
