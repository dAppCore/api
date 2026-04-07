<?php

declare(strict_types=1);

use Core\Api\Models\ApiKey;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;
use Illuminate\Support\Facades\Http;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);

    $result = ApiKey::generate(
        $this->workspace->id,
        $this->user->id,
        'SEO Key',
        [ApiKey::SCOPE_READ]
    );

    $this->plainKey = $result['plain_key'];
});

it('returns a technical SEO report for a URL', function () {
    Http::fake([
        'https://example.com*' => Http::response(<<<'HTML'
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Example Product Landing Page</title>
    <meta name="description" content="A concise example description for the landing page.">
    <link rel="canonical" href="https://example.com/landing-page">
    <meta property="og:title" content="Example Product Landing Page">
    <meta property="og:description" content="A concise example description for the landing page.">
    <meta property="og:image" content="https://example.com/og-image.jpg">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="Example">
    <meta name="twitter:card" content="summary_large_image">
</head>
<body>
    <h1>Example Product Landing Page</h1>
    <h2>Key Features</h2>
</body>
</html>
HTML, 200, [
            'Content-Type' => 'text/html; charset=utf-8',
        ]),
    ]);

    $response = $this->getJson('/api/seo/report?url=https://example.com', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('data.url', 'https://example.com');
    $response->assertJsonPath('data.status_code', 200);
    $response->assertJsonPath('data.summary.title', 'Example Product Landing Page');
    $response->assertJsonPath('data.summary.description', 'A concise example description for the landing page.');
    $response->assertJsonPath('data.headings.h1', 1);
    $response->assertJsonPath('data.open_graph.site_name', 'Example');
    $response->assertJsonPath('data.score', 100);
    $response->assertJsonPath('data.issues', []);
});
