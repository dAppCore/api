<?php

declare(strict_types=1);

namespace Core\Api\Services {
    // Override built-in for test isolation
    function dns_get_record(string $hostname, int $type = DNS_A | DNS_AAAA, mixed ...$args): array|false
    {
        if ($hostname === 'seo-pinned.example.test') {
            return [
                ['ip' => '1.1.1.1'],
                ['ipv6' => '2606:4700:4700::1111'],
            ];
        }

        if ($hostname === 'seo-private.example.test') {
            return [
                ['ip' => '10.0.0.1'],
            ];
        }

        return \dns_get_record($hostname, $type, ...$args);
    }
}

namespace {

use Core\Api\Services\SeoReportService;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Http;

define('SEO_TEST_URL', 'https://1.1.1.1/article');
define('SEO_CONTENT_TYPE', 'text/html; charset=utf-8');
define('SEO_PAGE_TITLE', 'Example Product Landing Page');
define('SEO_PAGE_DESC', 'A concise example description for the landing page.');

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
        expect($request->url())->toBe(SEO_TEST_URL);
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
            'Content-Type' => SEO_CONTENT_TYPE,
        ]);
    });

    $report = seoReportService()->analyse(SEO_TEST_URL);

    expect($report)->toMatchArray([
        'url' => SEO_TEST_URL,
        'status_code' => 200,
        'content_type' => SEO_CONTENT_TYPE,
        'score' => 100,
        'summary' => [
            'title' => SEO_PAGE_TITLE,
            'description' => SEO_PAGE_DESC,
            'canonical' => 'https://example.test/article',
            'robots' => 'index,follow',
            'language' => 'en',
            'charset' => 'utf-8',
        ],
        'open_graph' => [
            'title' => SEO_PAGE_TITLE,
            'description' => SEO_PAGE_DESC,
            'image' => 'https://example.test/og-image.jpg',
            'type' => 'article',
            'site_name' => 'Example',
        ],
        'twitter' => [
            'card' => 'summary_large_image',
            'title' => SEO_PAGE_TITLE,
            'description' => SEO_PAGE_DESC,
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
            'Content-Type' => SEO_CONTENT_TYPE,
            'Content-Length' => '1048577',
        ]),
    ]);

    expect(fn () => seoReportService()->analyse(SEO_TEST_URL))
        ->toThrow(RuntimeException::class);
});

it('SeoReportService_analyse_Ugly_caps_streamed_bodies_without_content_length', function () {
    config(['api.seo.max_body_bytes' => 16]);

    try {
        Http::fake([
            'https://1.1.1.1/*' => Http::response('abcdefghijklmnopq', 200, [
                'Content-Type' => SEO_CONTENT_TYPE,
            ]),
        ]);

        expect(fn () => seoReportService()->analyse(SEO_TEST_URL))
            ->toThrow(RuntimeException::class);
    } finally {
        config()->offsetUnset('api.seo.max_body_bytes');
    }
});

it('SeoReportService_analyse_Ugly_blocks_unsafe_urls_before_fetching', function () {
    Http::fake();

    $userPass = 'user' . ':' . 'pass';
    $unsafeURI = 'https://' . $userPass . '@1.1.1.1/article';
    expect(fn () => seoReportService()->analyse($unsafeURI))
        ->toThrow(\InvalidArgumentException::class);

    Http::assertNothingSent();
});

it('SeoReportService_analyse_Ugly_blocks_unresolvable_hostnames', function () {
    Http::fake();

    expect(fn () => seoReportService()->analyse('https://seo-unresolvable.example.invalid/article'))
        ->toThrow(\InvalidArgumentException::class, 'The supplied URL could not be resolved to any address.');

    Http::assertNothingSent();
});

it('SeoReportService_analyse_Ugly_blocks_reserved_ip_literals', function () {
    Http::fake();

    expect(fn () => seoReportService()->analyse('https://224.0.0.1/article'))
        ->toThrow(\InvalidArgumentException::class);

    expect(fn () => seoReportService()->analyse('https://[2001:db8::1]/article'))
        ->toThrow(\InvalidArgumentException::class);

    Http::assertNothingSent();
});

it('SeoReportService_analyse_Bad_blocks_hostnames_that_resolve_to_private_ips', function () {
    Http::fake();

    expect(fn () => seoReportService()->analyse('https://seo-private.example.test/article'))
        ->toThrow(\InvalidArgumentException::class, 'private or reserved address');

    Http::assertNothingSent();
});

it('SeoReportService_analyse_Bad_rejects_hostnames_when_pinning_is_unavailable', function () {
    Http::fake();

    $service = new class extends SeoReportService
    {
        public function exposePrepareUrlForSsrf(string $url): array
        {
            return $this->prepareUrlForSsrf($url);
        }

        protected function supportsPinnedResolution(): bool
        {
            return false;
        }
    };

    expect(fn () => $service->exposePrepareUrlForSsrf('https://seo-pinned.example.test/article'))
        ->toThrow(\InvalidArgumentException::class, 'cannot be safely pinned');

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
