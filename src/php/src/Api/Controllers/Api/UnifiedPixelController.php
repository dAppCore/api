<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\RateLimit\RateLimit;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Routing\Controller;

/**
 * Unified tracking pixel controller.
 *
 * GET /api/pixel/{pixelKey} returns a transparent 1x1 GIF for image embeds.
 * POST /api/pixel/{pixelKey} returns 204 No Content for fetch-based tracking.
 */
class UnifiedPixelController extends Controller
{
    /**
     * Transparent 1x1 GIF used by browser pixel embeds.
     */
    private const TRANSPARENT_GIF = 'R0lGODlhAQABAPAAAP///wAAACH5BAAAAAAALAAAAAABAAEAAAICRAEAOw==';

    /**
     * Track a pixel hit.
     *
     * GET /api/pixel/abc12345 -> transparent GIF
     * POST /api/pixel/abc12345 -> 204 No Content
     */
    #[RateLimit(limit: 10000, window: 60)]
    public function track(Request $request, string $pixelKey): Response
    {
        if ($request->isMethod('post')) {
            return response()->noContent()
                ->header('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
                ->header('Pragma', 'no-cache')
                ->header('Expires', '0');
        }

        $pixel = base64_decode(self::TRANSPARENT_GIF);

        return response($pixel, 200)
            ->header('Content-Type', 'image/gif')
            ->header('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
            ->header('Pragma', 'no-cache')
            ->header('Expires', '0')
            ->header('Content-Length', (string) strlen($pixel));
    }
}
