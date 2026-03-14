<?php

declare(strict_types=1);

namespace Core\Api\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Log;
use Core\Api\Models\ApiKey;
use Core\Api\Services\ApiUsageService;
use Symfony\Component\HttpFoundation\Response;

/**
 * Track API Usage Middleware.
 *
 * Records request/response metrics for API analytics.
 * Should be applied after authentication middleware.
 */
class TrackApiUsage
{
    public function __construct(
        protected ApiUsageService $usageService
    ) {}

    /**
     * Handle an incoming request.
     */
    public function handle(Request $request, Closure $next): Response
    {
        // Record start time
        $startTime = microtime(true);

        // Process the request
        $response = $next($request);

        // Calculate response time
        $responseTimeMs = (int) ((microtime(true) - $startTime) * 1000);

        // Only track if we have an authenticated API key
        $apiKey = $request->attributes->get('api_key');

        if ($apiKey instanceof ApiKey) {
            $this->recordUsage($request, $response, $apiKey, $responseTimeMs);
        }

        return $response;
    }

    /**
     * Record the API usage.
     */
    protected function recordUsage(
        Request $request,
        Response $response,
        ApiKey $apiKey,
        int $responseTimeMs
    ): void {
        try {
            $this->usageService->record(
                apiKeyId: $apiKey->id,
                workspaceId: $apiKey->workspace_id,
                endpoint: $request->path(),
                method: $request->method(),
                statusCode: $response->getStatusCode(),
                responseTimeMs: $responseTimeMs,
                requestSize: strlen($request->getContent()),
                responseSize: strlen($response->getContent()),
                ipAddress: $request->ip(),
                userAgent: $request->userAgent()
            );
        } catch (\Throwable $e) {
            // Don't let analytics failures affect the API response
            Log::warning('Failed to record API usage', [
                'error' => $e->getMessage(),
                'api_key_id' => $apiKey->id,
                'endpoint' => $request->path(),
            ]);
        }
    }
}
