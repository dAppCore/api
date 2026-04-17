<?php

declare(strict_types=1);

namespace Core\Api\Services {
    function dns_get_record(string $hostname, int $type = DNS_A | DNS_AAAA, mixed ...$args): array|false
    {
        if ($hostname === 'seo-pinned.example.test') {
            return [
                ['ip' => '1.1.1.1'],
                ['ipv6' => '2606:4700:4700::1111'],
            ];
        }

        return \dns_get_record($hostname, $type, ...$args);
    }
}

namespace {

use Core\Api\Services\SeoReportService;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Http;

function seoReportService(): SeoReportService
{
    return app(SeoReportService::class);
}

function seoPendingRequestOptions(PendingRequest $request): array
{
    $reflection = new ReflectionProperty($request, 'options');

    return $reflection->getValue($request);
}

it('SeoReportService_analyse_Good_extracts_technical_signals', function () {
    Http::fake(function ($request) {
        expect($request->url())->toBe('https://1.1.1.1/article');
        expect($request->method())->toBe('GET');
        expect($request->header('User-Agent')[0])->toContain('SEO Reporter/1.0');
        expect($request->header('Accept')[0])->toBe('text/html,application/xhtml+xml');

        return Http::response(<<<'HTML'
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Example Product Landing Page</title>
    <meta name="description" content="A concise example description for the landing page.">
    <link rel="canonical" href="https://example.test/article">
    <meta name="robots" content="index,follow">
    <meta property="og:title" content="Example Product Landing Page">
    <meta property="og:description" content="A concise example description for the landing page.">
    <meta property="og:image" content="https://example.test/og-image.jpg">
    <meta property="og:type" content="article">
    <meta property="og:site_name" content="Example">
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="Example Product Landing Page">
    <meta name="twitter:description" content="A concise example description for the landing page.">
    <meta name="twitter:image" content="https://example.test/twitter.jpg">
</head>
<body>
    <h1>Example Product Landing Page</h1>
    <h2>Key Features</h2>
</body>
</html>
HTML, 200, [
            'Content-Type' => 'text/html; charset=utf-8',
        ]);
    });

    $report = seoReportService()->analyse('https://1.1.1.1/article');

    expect($report)->toMatchArray([
        'url' => 'https://1.1.1.1/article',
        'status_code' => 200,
        'content_type' => 'text/html; charset=utf-8',
        'score' => 100,
        'summary' => [
            'title' => 'Example Product Landing Page',
            'description' => 'A concise example description for the landing page.',
            'canonical' => 'https://example.test/article',
            'robots' => 'index,follow',
            'language' => 'en',
            'charset' => 'utf-8',
        ],
        'open_graph' => [
            'title' => 'Example Product Landing Page',
            'description' => 'A concise example description for the landing page.',
            'image' => 'https://example.test/og-image.jpg',
            'type' => 'article',
            'site_name' => 'Example',
        ],
        'twitter' => [
            'card' => 'summary_large_image',
            'title' => 'Example Product Landing Page',
            'description' => 'A concise example description for the landing page.',
            'image' => 'https://example.test/twitter.jpg',
        ],
        'headings' => [
            'h1' => 1,
            'h2' => 1,
            'h3' => 0,
            'h4' => 0,
            'h5' => 0,
            'h6' => 0,
        ],
        'issues' => [],
        'recommendations' => [],
    ]);
});

it('SeoReportService_analyse_Bad_rejects_oversized_responses', function () {
    Http::fake([
        'https://1.1.1.1/*' => Http::response('small-body', 200, [
            'Content-Type' => 'text/html; charset=utf-8',
            'Content-Length' => '1048577',
        ]),
    ]);

    expect(fn () => seoReportService()->analyse('https://1.1.1.1/article'))
        ->toThrow(RuntimeException::class);
});

it('SeoReportService_analyse_Ugly_caps_streamed_bodies_without_content_length', function () {
    config(['api.seo.max_body_bytes' => 16]);

    try {
        Http::fake([
            'https://1.1.1.1/*' => Http::response('abcdefghijklmnopq', 200, [
                'Content-Type' => 'text/html; charset=utf-8',
            ]),
        ]);

        expect(fn () => seoReportService()->analyse('https://1.1.1.1/article'))
            ->toThrow(RuntimeException::class);
    } finally {
        config()->offsetUnset('api.seo.max_body_bytes');
    }
});

it('SeoReportService_analyse_Ugly_blocks_unsafe_urls_before_fetching', function () {
    Http::fake();

    expect(fn () => seoReportService()->analyse('https://user:pass@1.1.1.1/article'))
        ->toThrow(\InvalidArgumentException::class);

    Http::assertNothingSent();
});

it('SeoReportService_analyse_Good_disables_redirects_and_pins_resolved_destinations', function () {
    if (! defined('CURLOPT_RESOLVE')) {
        $this->markTestSkipped('cURL extension is unavailable.');
    }

    $service = new class extends SeoReportService
    {
        public function exposePrepareUrlForSsrf(string $url): array
        {
            return $this->prepareUrlForSsrf($url);
        }

        public function exposeBuildRequest(array $curlOptions): PendingRequest
        {
            return $this->buildRequest($curlOptions);
        }
    };

    $curlOptions = $service->exposePrepareUrlForSsrf('https://seo-pinned.example.test/article');
    $request = $service->exposeBuildRequest($curlOptions['curl_options']);
    $options = seoPendingRequestOptions($request);

    expect($options['allow_redirects'] ?? null)->toBeFalse();
    expect($options['stream'] ?? null)->toBeTrue();
    expect($options['curl'][CURLOPT_RESOLVE] ?? [])->toContain('seo-pinned.example.test:443:1.1.1.1');
    expect($options['curl'][CURLOPT_RESOLVE] ?? [])->toContain('seo-pinned.example.test:443:[2606:4700:4700::1111]');
});

}
