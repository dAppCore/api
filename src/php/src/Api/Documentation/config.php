<?php

/**
 * API Documentation Configuration
 *
 * Configuration for OpenAPI/Swagger documentation powered by Scramble.
 */

return [

    /*
    |--------------------------------------------------------------------------
    | Documentation Enabled
    |--------------------------------------------------------------------------
    |
    | Enable or disable API documentation. When disabled, the /api/docs
    | endpoint will return 404.
    |
    */

    'enabled' => env('API_DOCS_ENABLED', true),

    /*
    |--------------------------------------------------------------------------
    | Documentation Path
    |--------------------------------------------------------------------------
    |
    | The URL path where API documentation is served.
    |
    */

    'path' => '/api/docs',

    /*
    |--------------------------------------------------------------------------
    | API Information
    |--------------------------------------------------------------------------
    |
    | Basic information about your API displayed in the documentation.
    |
    */

    'info' => [
        'title' => env('API_DOCS_TITLE', 'API Documentation'),
        'description' => env('API_DOCS_DESCRIPTION', 'REST API for programmatic access to services.'),
        'version' => env('API_DOCS_VERSION', '1.0.0'),
        'contact' => [
            'name' => env('API_DOCS_CONTACT_NAME'),
            'email' => env('API_DOCS_CONTACT_EMAIL'),
            'url' => env('API_DOCS_CONTACT_URL'),
        ],
        'license' => [
            'name' => env('API_DOCS_LICENSE_NAME', 'Proprietary'),
            'url' => env('API_DOCS_LICENSE_URL'),
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Servers
    |--------------------------------------------------------------------------
    |
    | List of API servers displayed in the documentation.
    |
    */

    'servers' => [
        [
            'url' => env('APP_URL', 'http://localhost'),
            'description' => 'Current Environment',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Authentication Schemes
    |--------------------------------------------------------------------------
    |
    | Configure how authentication is documented in OpenAPI.
    |
    */

    'auth' => [
        // API Key authentication via header
        'api_key' => [
            'enabled' => true,
            'name' => 'X-API-Key',
            'in' => 'header',
            'description' => 'API key for authentication. Create keys in your workspace settings.',
        ],

        // Bearer token authentication
        'bearer' => [
            'enabled' => true,
            'scheme' => 'bearer',
            'format' => 'JWT',
            'description' => 'Bearer token authentication for user sessions.',
        ],

        // OAuth2 (if applicable)
        'oauth2' => [
            'enabled' => false,
            'flows' => [
                'authorizationCode' => [
                    'authorizationUrl' => '/oauth/authorize',
                    'tokenUrl' => '/oauth/token',
                    'refreshUrl' => '/oauth/token',
                    'scopes' => [
                        'read' => 'Read access to resources',
                        'write' => 'Write access to resources',
                        'delete' => 'Delete access to resources',
                    ],
                ],
            ],
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Workspace Header
    |--------------------------------------------------------------------------
    |
    | Configure the workspace header documentation.
    |
    */

    'workspace' => [
        'header_name' => 'X-Workspace-ID',
        'required' => false,
        'description' => 'Optional workspace identifier for multi-tenant operations. If not provided, the default workspace associated with the API key will be used.',
    ],

    /*
    |--------------------------------------------------------------------------
    | Rate Limiting Documentation
    |--------------------------------------------------------------------------
    |
    | Configure how rate limits are documented in responses.
    |
    */

    'rate_limits' => [
        'enabled' => true,
        'headers' => [
            'X-RateLimit-Limit' => 'Maximum number of requests allowed per window',
            'X-RateLimit-Remaining' => 'Number of requests remaining in the current window',
            'X-RateLimit-Reset' => 'Unix timestamp when the rate limit window resets',
            'Retry-After' => 'Seconds to wait before retrying (only on 429 responses)',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Module Tags
    |--------------------------------------------------------------------------
    |
    | Map module namespaces to documentation tags for grouping endpoints.
    |
    */

    'tags' => [
        // Module namespace => Tag configuration
        'Bio' => [
            'name' => 'Bio Links',
            'description' => 'Bio link pages, blocks, and customization',
        ],
        'Commerce' => [
            'name' => 'Commerce',
            'description' => 'Billing, subscriptions, orders, and invoices',
        ],
        'Analytics' => [
            'name' => 'Analytics',
            'description' => 'Website and link analytics tracking',
        ],
        'Social' => [
            'name' => 'Social',
            'description' => 'Social media management and scheduling',
        ],
        'Notify' => [
            'name' => 'Notifications',
            'description' => 'Push notifications and alerts',
        ],
        'Support' => [
            'name' => 'Support',
            'description' => 'Helpdesk and customer support',
        ],
        'Tenant' => [
            'name' => 'Workspaces',
            'description' => 'Workspace and team management',
        ],
        'Pixel' => [
            'name' => 'Pixel',
            'description' => 'Unified tracking pixel endpoints',
        ],
        'SEO' => [
            'name' => 'SEO',
            'description' => 'SEO analysis and reporting',
        ],
        'MCP' => [
            'name' => 'MCP',
            'description' => 'Model Context Protocol HTTP bridge',
        ],
        'Content' => [
            'name' => 'Content',
            'description' => 'AI content generation',
        ],
        'Trust' => [
            'name' => 'Trust',
            'description' => 'Social proof and testimonials',
        ],
        'Webhooks' => [
            'name' => 'Webhooks',
            'description' => 'Webhook endpoints and management',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Route Filtering
    |--------------------------------------------------------------------------
    |
    | Configure which routes are included in the documentation.
    |
    */

    'routes' => [
        // Only include routes matching these patterns
        'include' => [
            'api/*',
        ],

        // Exclude routes matching these patterns
        'exclude' => [
            'api/sanctum/*',
            'api/telescope/*',
            'api/horizon/*',
        ],

        // Hide internal/admin routes from public docs
        'hide_internal' => true,
    ],

    /*
    |--------------------------------------------------------------------------
    | Documentation UI
    |--------------------------------------------------------------------------
    |
    | Configure the documentation UI appearance.
    |
    */

    'ui' => [
        // Default UI renderer: 'swagger', 'scalar', 'redoc', 'stoplight'
        'default' => 'scalar',

        // Swagger UI specific options
        'swagger' => [
            'doc_expansion' => 'none', // 'list', 'full', 'none'
            'filter' => true,
            'show_extensions' => true,
            'show_common_extensions' => true,
        ],

        // Scalar specific options
        'scalar' => [
            'theme' => 'default', // 'default', 'alternate', 'moon', 'purple', 'solarized'
            'show_sidebar' => true,
            'hide_download_button' => false,
            'hide_models' => false,
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Access Control
    |--------------------------------------------------------------------------
    |
    | Configure who can access the documentation.
    |
    */

    'access' => [
        // Require authentication to view docs
        'require_auth' => env('API_DOCS_REQUIRE_AUTH', false),

        // Only allow these roles to view docs (empty = all authenticated users)
        'allowed_roles' => [],

        // Allow unauthenticated access in these environments
        'public_environments' => ['local', 'testing', 'staging'],

        // IP whitelist for production (empty = no restriction)
        'ip_whitelist' => [],
    ],

    /*
    |--------------------------------------------------------------------------
    | Caching
    |--------------------------------------------------------------------------
    |
    | Configure documentation caching.
    |
    */

    'cache' => [
        // Enable caching of generated OpenAPI spec
        'enabled' => env('API_DOCS_CACHE_ENABLED', true),

        // Cache key prefix
        'key' => 'api-docs:openapi',

        // Cache duration in seconds (1 hour default)
        'ttl' => env('API_DOCS_CACHE_TTL', 3600),

        // Disable cache in these environments
        'disabled_environments' => ['local', 'testing'],
    ],

];
