<?php

declare(strict_types=1);

namespace Core\Api\Services;

use DOMDocument;
use DOMXPath;
use Illuminate\Http\Client\Response as HttpResponse;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Str;
use RuntimeException;
use Throwable;

/**
 * SEO report service.
 *
 * Fetches a page and extracts the most useful technical SEO signals from it.
 */
class SeoReportService
{
    /**
     * Maximum number of bytes to read from the fetched page.
     */
    protected const MAX_BODY_BYTES = 1_048_576;

    /**
     * Analyse a URL and return a technical SEO report.
     *
     * @throws \InvalidArgumentException when the URL is blocked for SSRF reasons.
     * @throws RuntimeException when the fetch fails or the response is too large.
     */
    public function analyse(string $url): array
    {
        $curlOptions = $this->prepareUrlForSsrf($url);

        try {
            /** @var HttpResponse $response */
            $response = $this->buildRequest($curlOptions['curl_options'] ?? [])->get($url)->throw();
            $html = $this->readBodyWithLimit($response);
        } catch (RuntimeException $exception) {
            throw $exception;
        } catch (Throwable $exception) {
            throw new RuntimeException('Unable to fetch the requested URL.', 0, $exception);
        }

        $xpath = $this->loadXPath($html);

        $title = $this->extractSingleText($xpath, '//title');
        $description = $this->extractMetaContent($xpath, 'description');
        $canonical = $this->extractLinkHref($xpath, 'canonical');
        $robots = $this->extractMetaContent($xpath, 'robots');
        $language = $this->extractHtmlAttribute($xpath, 'lang');
        $charset = $this->extractCharset($xpath);

        $openGraph = [
            'title' => $this->extractMetaContent($xpath, 'og:title', 'property'),
            'description' => $this->extractMetaContent($xpath, 'og:description', 'property'),
            'image' => $this->extractMetaContent($xpath, 'og:image', 'property'),
            'type' => $this->extractMetaContent($xpath, 'og:type', 'property'),
            'site_name' => $this->extractMetaContent($xpath, 'og:site_name', 'property'),
        ];

        $twitterCard = [
            'card' => $this->extractMetaContent($xpath, 'twitter:card', 'name'),
            'title' => $this->extractMetaContent($xpath, 'twitter:title', 'name'),
            'description' => $this->extractMetaContent($xpath, 'twitter:description', 'name'),
            'image' => $this->extractMetaContent($xpath, 'twitter:image', 'name'),
        ];

        $headings = $this->countHeadings($xpath);
        $issues = $this->buildIssues($title, $description, $canonical, $robots, $openGraph, $headings);

        return [
            'url' => $url,
            'status_code' => $response->status(),
            'content_type' => $response->header('Content-Type'),
            'score' => $this->calculateScore($issues),
            'summary' => [
                'title' => $title,
                'description' => $description,
                'canonical' => $canonical,
                'robots' => $robots,
                'language' => $language,
                'charset' => $charset,
            ],
            'open_graph' => $openGraph,
            'twitter' => $twitterCard,
            'headings' => $headings,
            'issues' => $issues,
            'recommendations' => $this->buildRecommendations($issues),
        ];
    }

    /**
     * Build the SEO fetch request with SSRF-safe client options.
     *
     * @param  array<string, array<int, string>>  $curlOptions
     */
    protected function buildRequest(array $curlOptions): PendingRequest
    {
        $request = Http::withHeaders([
            'User-Agent' => config('app.name', 'Core API').' SEO Reporter/1.0',
            'Accept' => 'text/html,application/xhtml+xml',
        ])
            ->timeout((int) config('api.seo.timeout', 10))
            ->withoutRedirecting()
            ->withOptions([
                'stream' => true,
            ]);

        if ($curlOptions !== []) {
            $request = $request->withOptions([
                'curl' => $curlOptions,
            ]);
        }

        return $request;
    }

    /**
     * Read the fetched page without allowing unbounded memory growth.
     */
    protected function readBodyWithLimit(HttpResponse $response): string
    {
        $maxBytes = (int) config('api.seo.max_body_bytes', self::MAX_BODY_BYTES);
        if ($maxBytes < 1) {
            throw new RuntimeException('SEO response size limit is invalid.');
        }

        $stream = $response->toPsrResponse()->getBody();
        try {
            $contentLength = $response->header('Content-Length');
            if (is_numeric($contentLength) && (int) $contentLength > $maxBytes) {
                throw new RuntimeException('The requested URL returned a response that is too large.');
            }

            $body = '';
            while (! $stream->eof()) {
                $chunk = $stream->read(8192);
                if ($chunk === '' && ! $stream->eof()) {
                    throw new RuntimeException('Unable to read the requested URL response body.');
                }

                $body .= $chunk;

                if (strlen($body) > $maxBytes) {
                    throw new RuntimeException('The requested URL returned a response that is too large.');
                }
            }

            return $body;
        } finally {
            $stream->close();
        }
    }

    /**
     * Load an HTML document into an XPath query object.
     */
    protected function loadXPath(string $html): DOMXPath
    {
        $previous = libxml_use_internal_errors(true);
        try {
            $document = new DOMDocument();
            $document->loadHTML($html, LIBXML_NOERROR | LIBXML_NOWARNING);

            libxml_clear_errors();

            return new DOMXPath($document);
        } finally {
            libxml_use_internal_errors($previous);
        }
    }

    /**
     * Extract the first text node matched by an XPath query.
     */
    protected function extractSingleText(DOMXPath $xpath, string $query): ?string
    {
        $nodes = $xpath->query($query);

        if (! $nodes || $nodes->length === 0) {
            return null;
        }

        $node = $nodes->item(0);

        if (! $node) {
            return null;
        }

        $value = trim($node->textContent ?? '');

        return $value !== '' ? $value : null;
    }

    /**
     * Extract a meta tag content value.
     */
    protected function extractMetaContent(DOMXPath $xpath, string $name, string $attribute = 'name'): ?string
    {
        $query = sprintf('//meta[@%s=%s]/@content', $attribute, $this->quoteForXPath($name));
        $nodes = $xpath->query($query);

        if (! $nodes || $nodes->length === 0) {
            return null;
        }

        $node = $nodes->item(0);

        if (! $node) {
            return null;
        }

        $value = trim($node->textContent ?? '');

        return $value !== '' ? $value : null;
    }

    /**
     * Extract a link href value.
     */
    protected function extractLinkHref(DOMXPath $xpath, string $rel): ?string
    {
        $query = sprintf('//link[@rel=%s]/@href', $this->quoteForXPath($rel));
        $nodes = $xpath->query($query);

        if (! $nodes || $nodes->length === 0) {
            return null;
        }

        $node = $nodes->item(0);

        if (! $node) {
            return null;
        }

        $value = trim($node->textContent ?? '');

        return $value !== '' ? $value : null;
    }

    /**
     * Extract the HTML lang attribute.
     */
    protected function extractHtmlAttribute(DOMXPath $xpath, string $attribute): ?string
    {
        $nodes = $xpath->query(sprintf('//html/@%s', $attribute));

        if (! $nodes || $nodes->length === 0) {
            return null;
        }

        $node = $nodes->item(0);

        if (! $node) {
            return null;
        }

        $value = trim($node->textContent ?? '');

        return $value !== '' ? $value : null;
    }

    /**
     * Extract a charset declaration.
     */
    protected function extractCharset(DOMXPath $xpath): ?string
    {
        $nodes = $xpath->query('//meta[@charset]/@charset');

        if ($nodes && $nodes->length > 0) {
            $node = $nodes->item(0);

            if ($node) {
                $value = trim($node->textContent ?? '');

                if ($value !== '') {
                    return $value;
                }
            }
        }

        // The http-equiv Content-Type meta returns a full value such as
        // "text/html; charset=utf-8". Extract only the charset token so that
        // callers receive a bare encoding label (e.g. "utf-8"), not the whole
        // content-type string.
        $contentType = $this->extractMetaContent($xpath, 'content-type', 'http-equiv');
        if ($contentType !== null) {
            if (preg_match('/charset\s*=\s*["\']?([^\s;"\']+)/i', $contentType, $matches)) {
                return $matches[1];
            }
        }

        return null;
    }

    /**
     * Count headings by level.
     *
     * @return array<string, int>
     */
    protected function countHeadings(DOMXPath $xpath): array
    {
        $counts = [];

        for ($level = 1; $level <= 6; $level++) {
            $nodes = $xpath->query('//h'.$level);
            $counts['h'.$level] = $nodes ? $nodes->length : 0;
        }

        return $counts;
    }

    /**
     * Build issue list from the extracted SEO data.
     *
     * @return array<int, array<string, string>>
     */
    protected function buildIssues(
        ?string $title,
        ?string $description,
        ?string $canonical,
        ?string $robots,
        array $openGraph,
        array $headings
    ): array {
        $issues = [];

        if ($title === null) {
            $issues[] = $this->issue('missing_title', 'No <title> tag was found.', 'high');
        } elseif (Str::length($title) < 10) {
            $issues[] = $this->issue('title_too_short', 'The page title is shorter than 10 characters.', 'medium');
        } elseif (Str::length($title) > 60) {
            $issues[] = $this->issue('title_too_long', 'The page title is longer than 60 characters.', 'medium');
        }

        if ($description === null) {
            $issues[] = $this->issue('missing_description', 'No meta description was found.', 'high');
        }

        if ($canonical === null) {
            $issues[] = $this->issue('missing_canonical', 'No canonical URL was found.', 'medium');
        }

        if (($headings['h1'] ?? 0) === 0) {
            $issues[] = $this->issue('missing_h1', 'The page does not contain an H1 heading.', 'high');
        } elseif (($headings['h1'] ?? 0) > 1) {
            $issues[] = $this->issue('multiple_h1', 'The page contains multiple H1 headings.', 'medium');
        }

        if (($openGraph['title'] ?? null) === null) {
            $issues[] = $this->issue('missing_og_title', 'No Open Graph title was found.', 'low');
        }

        if (($openGraph['description'] ?? null) === null) {
            $issues[] = $this->issue('missing_og_description', 'No Open Graph description was found.', 'low');
        }

        if ($robots !== null && Str::contains(Str::lower($robots), ['noindex', 'nofollow'])) {
            $issues[] = $this->issue('robots_restricted', 'Robots directives block indexing or following links.', 'high');
        }

        return $issues;
    }

    /**
     * Convert a list of issues to a report score.
     */
    protected function calculateScore(array $issues): int
    {
        $penalties = [
            'missing_title' => 20,
            'title_too_short' => 5,
            'title_too_long' => 5,
            'missing_description' => 15,
            'missing_canonical' => 10,
            'missing_h1' => 15,
            'multiple_h1' => 5,
            'missing_og_title' => 5,
            'missing_og_description' => 5,
            'robots_restricted' => 20,
        ];

        $score = 100;

        foreach ($issues as $issue) {
            $score -= $penalties[$issue['code']] ?? 0;
        }

        return max(0, $score);
    }

    /**
     * Build recommendations from issues.
     *
     * @return array<int, string>
     */
    protected function buildRecommendations(array $issues): array
    {
        $recommendations = [];

        foreach ($issues as $issue) {
            $recommendations[] = match ($issue['code']) {
                'missing_title' => 'Add a concise page title that describes the page content.',
                'title_too_short' => 'Expand the page title so it is more descriptive.',
                'title_too_long' => 'Shorten the page title to keep it under 60 characters.',
                'missing_description' => 'Add a meta description to improve search snippets.',
                'missing_canonical' => 'Add a canonical URL to prevent duplicate content issues.',
                'missing_h1' => 'Add a single, descriptive H1 heading.',
                'multiple_h1' => 'Reduce the page to a single primary H1 heading.',
                'missing_og_title' => 'Add an Open Graph title for better social sharing.',
                'missing_og_description' => 'Add an Open Graph description for better social sharing.',
                'robots_restricted' => 'Remove noindex or nofollow directives if the page should be indexed.',
                default => $issue['message'],
            };
        }

        return array_values(array_unique($recommendations));
    }

    /**
     * Build an issue record.
     *
     * @return array{code: string, message: string, severity: string}
     */
    protected function issue(string $code, string $message, string $severity): array
    {
        return [
            'code' => $code,
            'message' => $message,
            'severity' => $severity,
        ];
    }

    /**
     * Validate that a URL is safe to fetch and does not target internal/private
     * network resources (SSRF protection).
     *
     * Blocks:
     *  - Non-HTTP/HTTPS schemes
     *  - Loopback addresses (127.0.0.0/8, ::1)
     *  - RFC-1918 private ranges (10/8, 172.16/12, 192.168/16)
     *  - Link-local ranges (169.254.0.0/16, fe80::/10)
     *  - IPv6 ULA (fc00::/7)
     *
     * @throws \InvalidArgumentException when the URL fails SSRF validation.
     */
    protected function prepareUrlForSsrf(string $url): array
    {
        $parsed = parse_url($url);

        if ($parsed === false || empty($parsed['scheme']) || empty($parsed['host'])) {
            throw new \InvalidArgumentException('The supplied URL is not valid.');
        }

        $scheme = strtolower((string) $parsed['scheme']);

        if (! in_array($scheme, ['http', 'https'], true)) {
            throw new \InvalidArgumentException('Only HTTP and HTTPS URLs are permitted.');
        }

        $host = $parsed['host'];
        $port = isset($parsed['port'])
            ? (int) $parsed['port']
            : ($scheme === 'https' ? 443 : 80);
        $resolveEntries = [];

        if (isset($parsed['user']) || isset($parsed['pass'])) {
            throw new \InvalidArgumentException('The supplied URL is not valid.');
        }

        // If the host is an IP literal (IPv4 or bracketed IPv6), validate it
        // directly. dns_get_record returns nothing for IP literals and
        // gethostbyname returns the same value, so both would silently skip
        // the private-range check without this explicit guard.
        $normalised = ltrim(rtrim($host, ']'), '['); // strip IPv6 brackets
        if (filter_var($normalised, FILTER_VALIDATE_IP) !== false) {
            if ($this->isPrivateIp($normalised)) {
                throw new \InvalidArgumentException('The supplied URL resolves to a private or reserved address.');
            }

            // Valid public IP literal — no DNS lookup required.
            return [
                'curl_options' => [],
            ];
        }

        $ips = $this->resolvePublicIps($host);
        foreach ($ips as $ip) {
            $resolveEntries[] = sprintf(
                '%s:%d:%s',
                $host,
                $port,
                str_contains($ip, ':') ? '['.$ip.']' : $ip
            );
        }

        if ($resolveEntries === []) {
            throw new \InvalidArgumentException('The supplied URL could not be resolved to any address.');
        }

        return [
            'curl_options' => defined('CURLOPT_RESOLVE')
                ? [
                    CURLOPT_RESOLVE => array_values(array_unique($resolveEntries)),
                ]
                : [],
        ];
    }

    /**
     * Resolve a hostname to public IPs, following CNAME chains.
     *
     * @return array<int, string>
     */
    protected function resolvePublicIps(string $host, array &$visitedHosts = [], int $depth = 0): array
    {
        $normalisedHost = strtolower(rtrim($host, '.'));

        if ($normalisedHost === '' || isset($visitedHosts[$normalisedHost])) {
            throw new \InvalidArgumentException('The supplied URL could not be resolved to any address.');
        }

        if ($depth > 8) {
            throw new \InvalidArgumentException('The supplied URL could not be resolved to any address.');
        }

        $visitedHosts[$normalisedHost] = true;

        $records = dns_get_record($host, DNS_A | DNS_AAAA | DNS_CNAME) ?: [];
        if ($records === []) {
            $resolved = gethostbyname($host);
            if ($resolved !== $host) {
                $records[] = ['ip' => $resolved];
            }
        }

        if ($records === []) {
            throw new \InvalidArgumentException('The supplied URL could not be resolved to any address.');
        }

        $ips = [];

        foreach ($records as $record) {
            $ip = $record['ip'] ?? $record['ipv6'] ?? null;
            if ($ip !== null) {
                if ($this->isPrivateIp($ip)) {
                    throw new \InvalidArgumentException('The supplied URL resolves to a private or reserved address.');
                }

                $ips[] = $ip;

                continue;
            }

            if (($record['type'] ?? null) === 'CNAME' && ! empty($record['target'])) {
                $ips = array_merge(
                    $ips,
                    $this->resolvePublicIps((string) $record['target'], $visitedHosts, $depth + 1)
                );
            }
        }

        return array_values(array_unique($ips));
    }

    /**
     * Return true when an IP address falls within a private, loopback, or
     * link-local range.
     */
    protected function isPrivateIp(string $ip): bool
    {
        if (filter_var($ip, FILTER_VALIDATE_IP) === false) {
            return true; // Treat unresolvable as unsafe.
        }

        $packed = inet_pton($ip);
        if ($packed === false) {
            return true;
        }

        // Preserve support for public IPv4-mapped IPv6 literals while still
        // applying the reserved-range guard to the embedded IPv4 address.
        if (strlen($packed) === 16 && str_repeat("\x00", 10)."\xff\xff" === substr($packed, 0, 12)) {
            $embeddedIpv4 = inet_ntop(substr($packed, 12, 4));
            if ($embeddedIpv4 === false) {
                return true;
            }

            return filter_var(
                $embeddedIpv4,
                FILTER_VALIDATE_IP,
                FILTER_FLAG_NO_PRIV_RANGE | FILTER_FLAG_NO_RES_RANGE
            ) === false;
        }

        return filter_var(
            $ip,
            FILTER_VALIDATE_IP,
            FILTER_FLAG_NO_PRIV_RANGE | FILTER_FLAG_NO_RES_RANGE
        ) === false;
    }

    /**
     * Quote a literal for XPath queries.
     */
    protected function quoteForXPath(string $value): string
    {
        if (! str_contains($value, "'")) {
            return "'{$value}'";
        }

        if (! str_contains($value, '"')) {
            return '"'.$value.'"';
        }

        $parts = array_map(
            fn (string $part) => "'{$part}'",
            explode("'", $value)
        );

        return 'concat('.implode(", \"'\", ", $parts).')';
    }
}
