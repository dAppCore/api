@extends('api::layouts.docs')

@section('title', 'API Changelog')

@section('content')
<div class="max-w-4xl mx-auto px-4 sm:px-6 py-12">
    <div class="max-w-3xl">
        <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">API Changelog</h1>
        <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
            Track endpoint additions, compatibility notes, and announced deprecations from one stable location.
        </p>
    </div>

    <div class="space-y-8">
        <section class="bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm p-6">
            <div class="flex items-center justify-between gap-4 mb-4">
                <h2 class="h4 text-zinc-800 dark:text-zinc-100">Current</h2>
                <span class="text-xs font-medium uppercase tracking-wider text-emerald-700 dark:text-emerald-400">Active</span>
            </div>
            <ul class="list-disc list-inside space-y-2 text-zinc-600 dark:text-zinc-400">
                <li>`/docs` now serves the interactive API explorer route described by the RFC.</li>
                <li>Dedicated documentation pages are available for rate limits and changelog history.</li>
                <li>OpenAPI exports remain available at <code>/openapi.json</code> and <code>/openapi.yaml</code>.</li>
            </ul>
        </section>

        <section class="bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm p-6">
            <div class="flex items-center justify-between gap-4 mb-4">
                <h2 class="h4 text-zinc-800 dark:text-zinc-100">Deprecation Signals</h2>
                <span class="text-xs font-medium uppercase tracking-wider text-amber-700 dark:text-amber-400">Headers</span>
            </div>
            <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                Deprecated endpoints communicate migration metadata through standard headers so SDKs and custom clients can warn early.
            </p>
            <div class="bg-zinc-800 rounded-sm overflow-hidden">
                <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">Deprecation: true
Sunset: Wed, 30 Apr 2026 23:59:59 GMT
API-Deprecation-Notice-URL: https://docs.api.dappco.re/deprecation/example
API-Suggested-Replacement: POST /api/v2/example</code></pre>
            </div>
        </section>
    </div>
</div>
@endsection
