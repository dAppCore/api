@extends('api::layouts.docs')

@section('title', 'API Documentation')
@section('description', 'Build powerful integrations with the API. Access brain memory, content scoring, and more.')
@php($apiKeyPrefix = \Core\Api\Models\ApiKey::keyPrefixRoot())

@section('content')
<div class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-20">

    {{-- Hero --}}
    <div class="max-w-3xl mx-auto text-center mb-16">
        <div class="mb-4">
            <span class="font-nycd text-xl text-cyan-600">Developer Documentation</span>
        </div>
        <h1 class="h1 mb-6 text-zinc-800 dark:text-zinc-100">Build with the API</h1>
        <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-8">
            Store and retrieve agent memories, score content for AI patterns,
            and integrate intelligent tooling into your applications.
        </p>
        <div class="flex flex-wrap justify-center gap-4">
            <a href="{{ route('api.guides.quickstart') }}" class="btn text-white bg-cyan-600 hover:bg-cyan-700 px-6 py-3 rounded-sm font-medium">
                Get Started
            </a>
            <a href="{{ route('api.reference') }}" class="btn text-zinc-600 bg-white border border-zinc-200 hover:border-zinc-300 dark:text-zinc-300 dark:bg-zinc-800 dark:border-zinc-700 dark:hover:border-zinc-600 px-6 py-3 rounded-sm font-medium">
                API Reference
            </a>
        </div>
    </div>

    {{-- Features grid --}}
    <div class="grid md:grid-cols-3 gap-8 mb-16">

        {{-- Brain Memory --}}
        <div class="bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm p-6">
            <div class="w-10 h-10 flex items-center justify-center bg-cyan-100 dark:bg-cyan-900/30 rounded-sm mb-4">
                <i class="fa-solid fa-brain text-cyan-600"></i>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100">Brain Memory</h3>
            <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                Store and retrieve agent memories with vector search. Powered by Qdrant for semantic retrieval.
            </p>
            <a href="{{ route('api.reference') }}#brain" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 text-sm font-medium">
                Learn more &rarr;
            </a>
        </div>

        {{-- Content Scoring --}}
        <div class="bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm p-6">
            <div class="w-10 h-10 flex items-center justify-center bg-purple-100 dark:bg-purple-900/30 rounded-sm mb-4">
                <i class="fa-solid fa-chart-line text-purple-600"></i>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100">Content Scoring</h3>
            <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                Score text for AI-generated patterns and analyse linguistic imprints via the EaaS scoring engine.
            </p>
            <a href="{{ route('api.reference') }}#score" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 text-sm font-medium">
                Learn more &rarr;
            </a>
        </div>

        {{-- Authentication --}}
        <div class="bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm p-6">
            <div class="w-10 h-10 flex items-center justify-center bg-amber-100 dark:bg-amber-900/30 rounded-sm mb-4">
                <i class="fa-solid fa-key text-amber-600"></i>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100">Authentication</h3>
            <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                Secure API key authentication with scoped permissions. Generate keys from your workspace settings.
            </p>
            <a href="{{ route('api.guides.authentication') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 text-sm font-medium">
                Learn more &rarr;
            </a>
        </div>

    </div>

    {{-- Quick start code example --}}
    <div class="max-w-4xl mx-auto">
        <h2 class="h3 mb-6 text-center text-zinc-800 dark:text-zinc-100">Quick Start</h2>
        <div class="bg-zinc-800 rounded-sm overflow-hidden">
            <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                <span class="text-sm text-zinc-400">cURL</span>
                <button class="text-xs text-zinc-500 hover:text-zinc-300" onclick="navigator.clipboard.writeText(this.closest('.bg-zinc-800').querySelector('code').textContent)">
                    Copy
                </button>
            </div>
            <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-teal-400">curl</span> <span class="text-zinc-500">--request</span> POST \
  <span class="text-zinc-500">--url</span> <span class="text-amber-400">'https://api.lthn.ai/v1/brain/recall'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Authorization: Bearer {{ $apiKeyPrefix }}your_api_key'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Content-Type: application/json'</span> \
  <span class="text-zinc-500">--data</span> <span class="text-amber-400">'{"query": "hello world"}'</span></code></pre>
        </div>

        <div class="mt-4 text-center">
            <a href="{{ route('api.guides.quickstart') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 text-sm font-medium">
                View full quick start guide &rarr;
            </a>
        </div>
    </div>

    {{-- API endpoints preview --}}
    <div class="mt-16">
        <h2 class="h3 mb-8 text-center text-zinc-800 dark:text-zinc-100">API Endpoints</h2>
        <div class="grid md:grid-cols-2 gap-4 max-w-4xl mx-auto">
            @foreach([
                ['method' => 'POST', 'path' => '/v1/brain/remember', 'desc' => 'Store a memory'],
                ['method' => 'POST', 'path' => '/v1/brain/recall', 'desc' => 'Search memories by query'],
                ['method' => 'DELETE', 'path' => '/v1/brain/forget/{id}', 'desc' => 'Delete a memory'],
                ['method' => 'POST', 'path' => '/v1/score/content', 'desc' => 'Score text for AI patterns'],
                ['method' => 'POST', 'path' => '/v1/score/imprint', 'desc' => 'Linguistic imprint analysis'],
                ['method' => 'GET', 'path' => '/v1/score/health', 'desc' => 'Scoring engine health check'],
            ] as $endpoint)
            <a href="{{ route('api.reference') }}" class="flex items-center gap-4 p-4 bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
                <span class="inline-flex items-center justify-center px-2 py-1 text-xs font-medium rounded {{ $endpoint['method'] === 'GET' ? 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400' : ($endpoint['method'] === 'DELETE' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400') }}">
                    {{ $endpoint['method'] }}
                </span>
                <div class="flex-1 min-w-0">
                    <code class="text-sm font-pt-mono text-zinc-800 dark:text-zinc-200 truncate block">{{ $endpoint['path'] }}</code>
                    <span class="text-xs text-zinc-500 dark:text-zinc-400">{{ $endpoint['desc'] }}</span>
                </div>
            </a>
            @endforeach
        </div>

        <div class="mt-8 text-center">
            <a href="{{ route('api.swagger') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 font-medium">
                View all endpoints in Swagger UI &rarr;
            </a>
        </div>
    </div>

</div>
@endsection
