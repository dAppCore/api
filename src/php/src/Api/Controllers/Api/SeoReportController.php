<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Documentation\Attributes\ApiParameter;
use Core\Api\Documentation\Attributes\ApiTag;
use Core\Api\Services\SeoReportService;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use InvalidArgumentException;
use RuntimeException;

/**
 * SEO report and analysis controller.
 */
#[ApiTag('SEO', 'SEO report and analysis endpoints')]
class SeoReportController extends Controller
{
    use HasApiResponses;

    public function __construct(
        protected SeoReportService $seoReportService
    ) {
    }

    /**
     * Analyse a URL and return a technical SEO report.
     *
     * GET /api/seo/report?url=https://example.com
     */
    #[ApiParameter(
        name: 'url',
        in: 'query',
        type: 'string',
        description: 'URL to analyse',
        required: true,
        format: 'uri'
    )]
    public function show(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'url' => ['required', 'url'],
        ]);

        try {
            $report = $this->seoReportService->analyse($validated['url']);
        } catch (InvalidArgumentException) {
            return $this->validationErrorResponse([
                'url' => ['The requested URL must be a public HTTP or HTTPS endpoint.'],
            ]);
        } catch (RuntimeException) {
            return $this->errorResponse(
                errorCode: 'seo_unavailable',
                message: 'Unable to fetch the requested URL.',
                meta: [
                    'provider' => 'seo',
                ],
                status: 502,
            );
        }

        return response()->json([
            'data' => $report,
        ]);
    }
}
