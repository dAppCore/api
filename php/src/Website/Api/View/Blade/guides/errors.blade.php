@extends('api::layouts.docs')

@section('title', 'Error Handling')

@section('content')
<div class="flex">

    {{-- Sidebar --}}
    <aside class="hidden lg:block fixed left-0 top-16 md:top-20 bottom-0 w-64 border-r border-zinc-200 dark:border-zinc-800">
        <div class="h-full px-4 py-8 overflow-y-auto no-scrollbar">
            <nav>
                <ul class="space-y-2">
                    <li>
                        <a href="#overview" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Overview
                        </a>
                    </li>
                    <li>
                        <a href="#http-codes" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            HTTP Status Codes
                        </a>
                    </li>
                    <li>
                        <a href="#error-format" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Error Format
                        </a>
                    </li>
                    <li>
                        <a href="#common-errors" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Common Errors
                        </a>
                    </li>
                    <li>
                        <a href="#rate-limiting" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Rate Limiting
                        </a>
                    </li>
                </ul>
            </nav>
        </div>
    </aside>

    {{-- Main content --}}
    <div class="lg:pl-64 w-full">
        <div class="max-w-3xl mx-auto px-4 sm:px-6 py-12">

            {{-- Breadcrumb --}}
            <nav class="mb-8">
                <ol class="flex items-center gap-2 text-sm">
                    <li><a href="{{ route('api.guides') }}" class="text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200">Guides</a></li>
                    <li class="text-zinc-400">/</li>
                    <li class="text-zinc-800 dark:text-zinc-200">Error Handling</li>
                </ol>
            </nav>

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">Error Handling</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
                Understand API error codes and how to handle them gracefully.
            </p>

            {{-- Overview --}}
            <section id="overview" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Overview</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    The API uses conventional HTTP response codes to indicate success or failure. Codes in the 2xx range indicate success, 4xx indicate client errors, and 5xx indicate server errors.
                </p>
            </section>

            {{-- HTTP Codes --}}
            <section id="http-codes" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">HTTP Status Codes</h2>

                <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Code</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Meaning</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded text-xs font-medium">200</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Success - Request completed successfully</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded text-xs font-medium">201</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Created - Resource was created successfully</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">400</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Bad Request - Invalid request parameters</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">401</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Unauthorised - Invalid or missing API key</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">403</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Forbidden - Insufficient permissions</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">404</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Not Found - Resource doesn't exist</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">422</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Unprocessable - Validation failed</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-xs font-medium">429</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Too Many Requests - Rate limit exceeded</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><span class="px-2 py-1 bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 rounded text-xs font-medium">500</span></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Server Error - Something went wrong on our end</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            {{-- Error Format --}}
            <section id="error-format" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Error Format</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Error responses include a JSON body with details:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">{
  <span class="text-blue-400">"message"</span>: <span class="text-green-400">"The given data was invalid."</span>,
  <span class="text-blue-400">"errors"</span>: {
    <span class="text-blue-400">"url"</span>: [
      <span class="text-green-400">"The url has already been taken."</span>
    ]
  }
}</code></pre>
                </div>
            </section>

            {{-- Common Errors --}}
            <section id="common-errors" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Common Errors</h2>

                <div class="space-y-4">
                    <div class="p-4 border border-zinc-200 dark:border-zinc-700 rounded-sm">
                        <h4 class="font-medium text-zinc-800 dark:text-zinc-200 mb-2">Invalid API Key</h4>
                        <p class="text-sm text-zinc-600 dark:text-zinc-400 mb-2">
                            Returned when the API key is missing, malformed, or revoked.
                        </p>
                        <code class="text-xs px-2 py-1 bg-zinc-100 dark:bg-zinc-800 rounded">401 Unauthorised</code>
                    </div>

                    <div class="p-4 border border-zinc-200 dark:border-zinc-700 rounded-sm">
                        <h4 class="font-medium text-zinc-800 dark:text-zinc-200 mb-2">Resource Not Found</h4>
                        <p class="text-sm text-zinc-600 dark:text-zinc-400 mb-2">
                            The requested resource (biolink, workspace, etc.) doesn't exist or you don't have access.
                        </p>
                        <code class="text-xs px-2 py-1 bg-zinc-100 dark:bg-zinc-800 rounded">404 Not Found</code>
                    </div>

                    <div class="p-4 border border-zinc-200 dark:border-zinc-700 rounded-sm">
                        <h4 class="font-medium text-zinc-800 dark:text-zinc-200 mb-2">Validation Failed</h4>
                        <p class="text-sm text-zinc-600 dark:text-zinc-400 mb-2">
                            Request data failed validation. Check the <code>errors</code> object for specific fields.
                        </p>
                        <code class="text-xs px-2 py-1 bg-zinc-100 dark:bg-zinc-800 rounded">422 Unprocessable Entity</code>
                    </div>
                </div>
            </section>

            {{-- Rate Limiting --}}
            <section id="rate-limiting" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Rate Limiting</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    API requests are rate limited to ensure fair usage. Rate limit headers are included in all responses:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-4">
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">X-RateLimit-Limit: 60
X-RateLimit-Remaining: 58
X-RateLimit-Reset: 1705320000</code></pre>
                </div>

                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    When rate limited, you'll receive a 429 response. Wait until the reset timestamp before retrying.
                </p>

                <div class="text-sm p-4 bg-zinc-50 border border-zinc-200 rounded-sm dark:bg-zinc-800 dark:border-zinc-700">
                    <p class="text-zinc-600 dark:text-zinc-400">
                        <strong>Tip:</strong> Implement exponential backoff in your retry logic. Start with a 1-second delay and double it with each retry, up to a maximum of 32 seconds.
                    </p>
                </div>
            </section>

            {{-- Next steps --}}
            <div class="flex items-center justify-between pt-8 border-t border-zinc-200 dark:border-zinc-700">
                <a href="{{ route('api.guides.webhooks') }}" class="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">
                    &larr; Webhooks
                </a>
                <a href="{{ route('api.reference') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 font-medium">
                    API Reference &rarr;
                </a>
            </div>

        </div>
    </div>

</div>
@endsection
