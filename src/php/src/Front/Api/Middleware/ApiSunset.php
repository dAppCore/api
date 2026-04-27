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
use DateTimeImmutable;
use DateTimeInterface;
use DateTimeZone;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * API Sunset Middleware.
 *
 * Adds deprecation headers to a route and optionally advertises a sunset
 * date and successor endpoint. Existing header values are preserved so
 * downstream middleware and handlers can keep their own warning metadata.
 */
class ApiSunset
{
    /**
     * Handle an incoming request.
     *
     * @param  string  $sunsetDate  The sunset date (YYYY-MM-DD or RFC7231 format), or empty for deprecation-only
     * @param  string|null  $replacement  Optional successor endpoint URL
     * @param  string|null  $noticeUrl  Optional deprecation notice URL
     */
    public function handle(Request $request, Closure $next, string $sunsetDate = '', ?string $replacement = null, ?string $noticeUrl = null): Response
    {
        /** @var Response $response */
        $response = $next($request);

        $sunsetDate = trim($sunsetDate);
        $replacement = $replacement !== null ? trim($replacement) : null;
        $noticeUrl = $noticeUrl !== null ? trim($noticeUrl) : null;

        if (! (bool) config('api.headers.include_deprecation', true)) {
            return $response;
        }

        $response->headers->set('Deprecation', 'true', false);

        if ($sunsetDate !== '') {
            $response->headers->set('Sunset', $this->formatSunsetDate($sunsetDate), false);
        }

        if ($replacement !== null && $replacement !== '') {
            $response->headers->set('Link', sprintf('<%s>; rel="successor-version"', $this->successorLinkTarget($replacement)), false);
            $response->headers->set('API-Suggested-Replacement', $replacement, false);
        }

        if ($noticeUrl !== null && $noticeUrl !== '') {
            $response->headers->set('API-Deprecation-Notice-URL', $noticeUrl, false);
        }

        $warning = 'This endpoint is deprecated.';
        if ($sunsetDate !== '') {
            $warning = "This endpoint is deprecated and will be removed on {$sunsetDate}.";
        }

        $response->headers->set('X-API-Warn', $warning, false);

        return $response;
    }

    /**
     * Extract the actual successor URL/path from a replacement suggestion.
     *
     * The RFC allows human-friendly suggestions like "POST /api/v2/billing",
     * but the Link header itself must contain just the target URI.
     */
    protected function successorLinkTarget(string $replacement): string
    {
        $replacement = trim($replacement);
        if ($replacement === '') {
            return $replacement;
        }

        $parts = preg_split('/\s+/', $replacement, 2);
        if ($parts !== false && count($parts) === 2) {
            $method = strtoupper(trim($parts[0]));
            if (in_array($method, ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'TRACE', 'CONNECT'], true)) {
                $target = trim($parts[1]);
                if ($target !== '') {
                    return $target;
                }
            }
        }

        return $replacement;
    }

    /**
     * Format the sunset date to RFC7231 format when possible.
     */
    protected function formatSunsetDate(string $sunsetDate): string
    {
        $sunsetDate = trim($sunsetDate);
        if ($sunsetDate === '') {
            return $sunsetDate;
        }

        // Already RFC7231-style dates contain a comma, so preserve them.
        if (str_contains($sunsetDate, ',')) {
            return $sunsetDate;
        }

        try {
            return (new DateTimeImmutable($sunsetDate))
                ->setTimezone(new DateTimeZone('GMT'))
                ->format(DateTimeInterface::RFC7231);
        } catch (\Throwable) {
            return $sunsetDate;
        }
    }
}
