@extends('api::layouts.docs')

@section('title', 'Authentication')

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
                        <a href="#api-keys" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            API Keys
                        </a>
                    </li>
                    <li>
                        <a href="#using-keys" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Using API Keys
                        </a>
                    </li>
                    <li>
                        <a href="#scopes" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Scopes
                        </a>
                    </li>
                    <li>
                        <a href="#security" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Security Best Practices
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
                    <li class="text-zinc-800 dark:text-zinc-200">Authentication</li>
                </ol>
            </nav>

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">Authentication</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
                Learn how to authenticate your API requests using API keys.
            </p>

            {{-- Overview --}}
            <section id="overview" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Overview</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    The API uses API keys for authentication. Each API key is scoped to a specific workspace and has configurable permissions.
                </p>
                <p class="text-zinc-600 dark:text-zinc-400">
                    API keys are prefixed with <code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-sm">hk_</code> to make them easily identifiable.
                </p>
            </section>

            {{-- API Keys --}}
            <section id="api-keys" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">API Keys</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    To create an API key:
                </p>
                <ol class="list-decimal list-inside space-y-2 text-zinc-600 dark:text-zinc-400 mb-6">
                    <li>Log in to your account</li>
                    <li>Navigate to <strong>Settings → API Keys</strong></li>
                    <li>Click <strong>Create API Key</strong></li>
                    <li>Enter a descriptive name (e.g., "Production", "Development")</li>
                    <li>Select the required scopes</li>
                    <li>Copy the generated key immediately</li>
                </ol>

                <div class="text-sm p-4 bg-amber-50 border border-amber-200 rounded-sm dark:bg-amber-900/20 dark:border-amber-800">
                    <div class="flex items-start">
                        <svg class="fill-amber-500 shrink-0 mr-3 mt-0.5" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 12a1 1 0 1 1 0-2 1 1 0 0 1 0 2zm1-4a1 1 0 0 1-2 0V5a1 1 0 0 1 2 0v3z"/>
                        </svg>
                        <p class="text-amber-800 dark:text-amber-200">
                            <strong>Important:</strong> API keys are only shown once when created. Store them securely as they cannot be retrieved later.
                        </p>
                    </div>
                </div>
            </section>

            {{-- Using Keys --}}
            <section id="using-keys" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Using API Keys</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Include your API key in the <code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-sm">Authorization</code> header as a Bearer token:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-6">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">HTTP Header</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">Authorization: Bearer hk_your_api_key_here</code></pre>
                </div>

                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Example request with cURL:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">cURL</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-teal-400">curl</span> <span class="text-zinc-500">--request</span> GET \
  <span class="text-zinc-500">--url</span> <span class="text-amber-400">'https://api.lthn.ai/v1/brain/recall'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Authorization: Bearer hk_your_api_key'</span></code></pre>
                </div>
            </section>

            {{-- Scopes --}}
            <section id="scopes" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Scopes</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    API keys can have different scopes to limit their permissions:
                </p>

                <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Scope</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Description</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">read</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Read access to resources (GET requests)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">write</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Create and update resources (POST, PUT requests)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">delete</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Delete resources (DELETE requests)</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            {{-- Security --}}
            <section id="security" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Security Best Practices</h2>
                <ul class="list-disc list-inside space-y-2 text-zinc-600 dark:text-zinc-400">
                    <li>Never commit API keys to version control</li>
                    <li>Use environment variables to store keys</li>
                    <li>Rotate keys periodically</li>
                    <li>Use the minimum required scopes</li>
                    <li>Revoke unused keys immediately</li>
                    <li>Never expose keys in client-side code</li>
                </ul>
            </section>

            {{-- Next steps --}}
            <div class="flex items-center justify-between pt-8 border-t border-zinc-200 dark:border-zinc-700">
                <a href="{{ route('api.guides.quickstart') }}" class="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">
                    &larr; Quick Start
                </a>
                <a href="{{ route('api.guides.webhooks') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 font-medium">
                    Webhooks &rarr;
                </a>
            </div>

        </div>
    </div>

</div>
@endsection
