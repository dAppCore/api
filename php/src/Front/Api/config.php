<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

/**
 * API Configuration.
 *
 * Settings for the REST API infrastructure including versioning,
 * rate limiting, and deprecation handling.
 */

return [
    /*
    |--------------------------------------------------------------------------
    | API Versioning
    |--------------------------------------------------------------------------
    |
    | Configure how API versions are handled. The middleware supports:
    | - URL path versioning: /api/v1/users
    | - Header versioning: Accept-Version: v1
    | - Accept header: application/vnd.hosthub.v1+json
    |
    | Version Strategy:
    | - Add new fields to existing versions (backwards compatible)
    | - Use new version for breaking changes (removing/renaming fields)
    | - Deprecate old versions with sunset dates before removal
    |
    */
    'versioning' => [
        // Default version when no version specified in request
        // Clients should always specify version explicitly
        'default' => (int) env('API_VERSION_DEFAULT', 1),

        // Current/latest API version
        // Used in deprecation warnings to suggest upgrade path
        'current' => (int) env('API_VERSION_CURRENT', 1),

        // Supported API versions (all still functional)
        // Remove versions from this list to disable them entirely
        'supported' => array_map('intval', array_filter(
            explode(',', env('API_VERSIONS_SUPPORTED', '1'))
        )),

        // Deprecated versions (still work but warn clients)
        // Responses include Deprecation: true header
        'deprecated' => array_map('intval', array_filter(
            explode(',', env('API_VERSIONS_DEPRECATED', ''))
        )),

        // Sunset dates for deprecated versions
        // Format: [version => 'YYYY-MM-DD']
        // After this date, version should be removed from 'supported'
        'sunset' => [
            // Example: 1 => '2025-12-31',
        ],

        // Optional successor endpoints for deprecated versions
        // Format: [version => '/api/v2/...']
        'replacement' => [
            // Example: 1 => '/api/v2/users',
        ],

        // Optional deprecation notice URLs for deprecated versions
        // Format: [version => 'https://docs.example.com/deprecation/v1']
        'notice_url' => [
            // Example: 1 => 'https://docs.example.com/deprecation/v1',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Response Headers
    |--------------------------------------------------------------------------
    |
    | Standard headers added to API responses.
    |
    */
    'headers' => [
        // Add X-API-Version header to all responses
        'include_version' => true,

        // Add deprecation warnings for old versions
        'include_deprecation' => true,
    ],
];
