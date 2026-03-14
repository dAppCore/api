<?php

/*
 * Core PHP Framework
 *
 * Licensed under the European Union Public Licence (EUPL) v1.2.
 * See LICENSE file for details.
 */

declare(strict_types=1);

namespace Core\Front\Api\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * API Sunset Middleware.
 *
 * Adds the HTTP Sunset header to responses to indicate when an endpoint
 * will be deprecated or removed.
 *
 * The Sunset header is defined in RFC 8594 and indicates that a resource
 * will become unresponsive at the specified date.
 *
 * ## Usage
 *
 * Apply to routes that will be sunset:
 *
 * ```php
 * Route::middleware('api.sunset:2025-06-01')->group(function () {
 *     Route::get('/legacy-endpoint', LegacyController::class);
 * });
 * ```
 *
 * Or with a replacement link:
 *
 * ```php
 * Route::middleware('api.sunset:2025-06-01,/api/v2/new-endpoint')->group(function () {
 *     Route::get('/old-endpoint', OldController::class);
 * });
 * ```
 *
 * ## Response Headers
 *
 * The middleware adds these headers:
 * - Sunset: <date in RFC7231 format>
 * - Deprecation: true
 * - Link: <replacement-url>; rel="successor-version" (if replacement provided)
 *
 * @see https://datatracker.ietf.org/doc/html/rfc8594 RFC 8594: The "Sunset" HTTP Header Field
 */
class ApiSunset
{
    /**
     * Handle an incoming request.
     *
     * @param  string  $sunsetDate  The sunset date (YYYY-MM-DD or RFC7231 format)
     * @param  string|null  $replacement  Optional replacement endpoint URL
     */
    public function handle(Request $request, Closure $next, string $sunsetDate, ?string $replacement = null): Response
    {
        /** @var Response $response */
        $response = $next($request);

        // Convert date to RFC7231 format if needed
        $formattedDate = $this->formatSunsetDate($sunsetDate);

        // Add Sunset header
        $response->headers->set('Sunset', $formattedDate);

        // Add Deprecation header
        $response->headers->set('Deprecation', 'true');

        // Add warning header
        $version = $request->attributes->get('api_version', 'unknown');
        $response->headers->set(
            'X-API-Warn',
            "This endpoint is deprecated and will be removed on {$sunsetDate}."
        );

        // Add Link header for replacement if provided
        if ($replacement !== null) {
            $response->headers->set('Link', "<{$replacement}>; rel=\"successor-version\"");
        }

        return $response;
    }

    /**
     * Format the sunset date to RFC7231 format.
     *
     * Accepts dates in YYYY-MM-DD format or already-formatted RFC7231 dates.
     */
    protected function formatSunsetDate(string $date): string
    {
        // Check if already in RFC7231 format (contains comma, day name)
        if (str_contains($date, ',')) {
            return $date;
        }

        try {
            return (new \DateTimeImmutable($date))
                ->setTimezone(new \DateTimeZone('GMT'))
                ->format(\DateTimeInterface::RFC7231);
        } catch (\Exception) {
            // If parsing fails, return as-is
            return $date;
        }
    }
}
