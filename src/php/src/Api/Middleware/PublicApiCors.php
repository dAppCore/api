<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * CORS middleware for public API endpoints.
 *
 * Public endpoints like the unified pixel need to be accessible from any
 * customer website, so we allow all origins. These endpoints are rate-limited
 * and do not expose sensitive data.
 */
class PublicApiCors
{
    /**
     * Handle an incoming request.
     */
    public function handle(Request $request, Closure $next): Response
    {
        // Handle preflight OPTIONS request
        if ($request->isMethod('OPTIONS')) {
            return $this->buildPreflightResponse($request);
        }

        $response = $next($request);

        return $this->addCorsHeaders($response, $request);
    }

    /**
     * Build preflight response for OPTIONS requests.
     */
    protected function buildPreflightResponse(Request $request): Response
    {
        $response = response('', 204);

        return $this->addCorsHeaders($response, $request);
    }

    /**
     * Add CORS headers to response.
     */
    protected function addCorsHeaders(Response $response, Request $request): Response
    {
        $origin = $request->header('Origin', '*');

        // Allow any origin for public widget/pixel endpoints
        $response->headers->set('Access-Control-Allow-Origin', $origin);
        $response->headers->set('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
        $response->headers->set('Access-Control-Allow-Headers', 'Content-Type, Accept, X-Requested-With');
        $response->headers->set('Access-Control-Expose-Headers', 'X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After');
        $response->headers->set('Access-Control-Max-Age', '3600');

        // Vary on Origin for proper caching
        $response->headers->set('Vary', 'Origin');

        return $response;
    }
}
