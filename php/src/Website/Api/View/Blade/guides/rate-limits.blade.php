@extends('api::layouts.docs')

@section('title', 'Rate Limits')

@section('content')
<div class="flex">

    <aside class="hidden lg:block fixed left-0 top-16 md:top-20 bottom-0 w-64 border-r border-zinc-200 dark:border-zinc-800">
        <div class="h-full px-4 py-8 overflow-y-auto no-scrollbar">
            <nav>
                <ul class="space-y-2">
                    <li><a href="#overview" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">Overview</a></li>
                    <li><a href="#headers" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">Headers</a></li>
                    <li><a href="#tiers" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">Tiers</a></li>
                    <li><a href="#backoff" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">Retry Strategy</a></li>
                </ul>
            </nav>
        </div>
    </aside>

    <div class="lg:pl-64 w-full">
        <div class="max-w-3xl mx-auto px-4 sm:px-6 py-12">
            <nav class="mb-8">
                <ol class="flex items-center gap-2 text-sm">
                    <li><a href="{{ route('api.guides') }}" class="text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200">Guides</a></li>
                    <li class="text-zinc-400">/</li>
                    <li class="text-zinc-800 dark:text-zinc-200">Rate Limits</li>
                </ol>
            </nav>

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">Rate Limits</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
                Rate limits apply per API key and per client IP. Every limited response includes enough metadata for clients to back off predictably.
            </p>

            <section id="overview" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Overview</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Authenticated requests are bucketed by API key or user token. Public requests fall back to the caller IP address.
                </p>
                <p class="text-zinc-600 dark:text-zinc-400">
                    When a bucket is exhausted the API responds with <code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-sm">429 Too Many Requests</code> and a retry hint.
                </p>
            </section>

            <section id="headers" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Headers</h2>
                <div class="overflow-x-auto mb-6">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Header</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Meaning</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-RateLimit-Limit</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Maximum requests available in the current window.</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-RateLimit-Remaining</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Requests still available before the bucket is exhausted.</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-RateLimit-Reset</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Unix timestamp for the next window reset.</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">Retry-After</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Seconds to wait before retrying after a 429 response.</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1776243600
Retry-After: 30</code></pre>
                </div>
            </section>

            <section id="tiers" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Plan Tiers</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Workspace subscription tiers can raise or lower request budgets. The applied tier is resolved server-side, so clients should trust response headers over hard-coded limits.
                </p>
                <ul class="list-disc list-inside space-y-2 text-zinc-600 dark:text-zinc-400">
                    <li>Unauthenticated traffic uses the default public budget.</li>
                    <li>Authenticated traffic uses the workspace or token budget when available.</li>
                    <li>Specific endpoints can define stricter caps for expensive operations.</li>
                </ul>
            </section>

            <section id="backoff" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Retry Strategy</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Clients should stop sending immediately after a <code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-sm">429</code> response and wait at least the advertised <code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-sm">Retry-After</code> duration.
                </p>
                <p class="text-zinc-600 dark:text-zinc-400">
                    For parallel workers, apply jitter on top of that delay so multiple retrying processes do not all resume at the same second.
                </p>
            </section>

            <div class="flex items-center justify-between pt-8 border-t border-zinc-200 dark:border-zinc-700">
                <a href="{{ route('api.guides.webhooks') }}" class="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">&larr; Webhooks</a>
                <a href="{{ route('api.guides.errors') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 font-medium">Error Handling &rarr;</a>
            </div>
        </div>
    </div>
</div>
@endsection
