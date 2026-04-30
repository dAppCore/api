@extends('api::layouts.docs')

@section('title', 'API Reference')

@section('content')
<div class="flex">

    {{-- Sidebar --}}
    <aside class="hidden lg:block fixed left-0 top-16 md:top-20 bottom-0 w-64 border-r border-zinc-200 dark:border-zinc-800">
        <div class="h-full px-4 py-8 overflow-y-auto no-scrollbar">
            <nav>
                <h3 class="text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-3">Resources</h3>
                <ul class="space-y-1">
                    <li>
                        <a href="#brain" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Brain Memory
                        </a>
                    </li>
                    <li>
                        <a href="#score" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Content Scoring
                        </a>
                    </li>
                    <li>
                        <a href="#collections" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Collections
                        </a>
                    </li>
                    <li>
                        <a href="#health" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Health
                        </a>
                    </li>
                </ul>
            </nav>
        </div>
    </aside>

    {{-- Main content --}}
    <div class="lg:pl-64 w-full">
        <div class="max-w-4xl mx-auto px-4 sm:px-6 py-12">

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">API Reference</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-4">
                Complete reference for all API endpoints.
            </p>
            <p class="text-zinc-600 dark:text-zinc-400 mb-12">
                Base URL: <code class="px-2 py-1 bg-zinc-100 dark:bg-zinc-800 rounded text-sm font-pt-mono">https://api.lthn.ai/v1</code>
            </p>

            {{-- Brain Memory --}}
            <section id="brain" data-scrollspy-target class="mb-16">
                <h2 class="h2 mb-6 text-zinc-800 dark:text-zinc-100 pb-2 border-b border-zinc-200 dark:border-zinc-700">Brain Memory</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-6">
                    Store and retrieve agent memories with vector search. Powered by Qdrant for semantic retrieval.
                </p>

                @include('api::partials.endpoint', [
                    'method' => 'POST',
                    'path' => '/brain/remember',
                    'description' => 'Store a new memory in the vector database.',
                    'body' => '{"content": "Go uses structural typing", "type": "fact", "project": "go-agentic", "tags": ["go", "typing"]}',
                    'response' => '{"id": "mem-abc-123", "type": "fact", "project": "go-agentic", "created_at": "2026-03-03T12:00:00+00:00"}'
                ])

                @include('api::partials.endpoint', [
                    'method' => 'POST',
                    'path' => '/brain/recall',
                    'description' => 'Search memories by semantic query. Returns ranked results with confidence scores.',
                    'body' => '{"query": "how does typing work in Go", "top_k": 5, "project": "go-agentic"}',
                    'response' => '{"memories": [{"id": "mem-abc-123", "type": "fact", "content": "Go uses structural typing", "confidence": 0.95}], "scores": {"mem-abc-123": 0.87}}'
                ])

                @include('api::partials.endpoint', [
                    'method' => 'DELETE',
                    'path' => '/brain/forget/{id}',
                    'description' => 'Delete a specific memory by ID.',
                    'response' => '{"deleted": true}'
                ])
            </section>

            {{-- Content Scoring --}}
            <section id="score" data-scrollspy-target class="mb-16">
                <h2 class="h2 mb-6 text-zinc-800 dark:text-zinc-100 pb-2 border-b border-zinc-200 dark:border-zinc-700">Content Scoring</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-6">
                    Score text for AI patterns and analyse linguistic imprints via the EaaS scoring engine.
                </p>

                @include('api::partials.endpoint', [
                    'method' => 'POST',
                    'path' => '/score/content',
                    'description' => 'Score text for AI-generated content patterns. Returns a score (0-1), confidence, and label.',
                    'body' => '{"text": "The text to analyse for AI patterns", "prompt": "Optional scoring prompt"}',
                    'response' => '{"score": 0.23, "confidence": 0.91, "label": "human"}'
                ])

                @include('api::partials.endpoint', [
                    'method' => 'POST',
                    'path' => '/score/imprint',
                    'description' => 'Perform linguistic imprint analysis on text. Returns a unique imprint fingerprint.',
                    'body' => '{"text": "The text to analyse for linguistic patterns"}',
                    'response' => '{"imprint": "abc123def456", "confidence": 0.88}'
                ])
            </section>

            {{-- Collections --}}
            <section id="collections" data-scrollspy-target class="mb-16">
                <h2 class="h2 mb-6 text-zinc-800 dark:text-zinc-100 pb-2 border-b border-zinc-200 dark:border-zinc-700">Collections</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-6">
                    Manage vector database collections for brain memory storage.
                </p>

                @include('api::partials.endpoint', [
                    'method' => 'POST',
                    'path' => '/brain/collections',
                    'description' => 'Ensure the workspace collection exists in the vector database. Creates it if missing.',
                    'response' => '{"status": "ok"}'
                ])
            </section>

            {{-- Health --}}
            <section id="health" data-scrollspy-target class="mb-16">
                <h2 class="h2 mb-6 text-zinc-800 dark:text-zinc-100 pb-2 border-b border-zinc-200 dark:border-zinc-700">Health</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-6">
                    Health check endpoints. These do not require authentication.
                </p>

                @include('api::partials.endpoint', [
                    'method' => 'GET',
                    'path' => '/score/health',
                    'description' => 'Check the health of the scoring engine and its upstream services.',
                    'response' => '{"status": "healthy", "upstream_status": 200}'
                ])
            </section>

            {{-- CTA --}}
            <div class="mt-12 p-6 bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm text-center">
                <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100">Try it out</h3>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">Test endpoints interactively with Swagger UI.</p>
                <a href="{{ route('api.swagger') }}" class="btn text-white bg-cyan-600 hover:bg-cyan-700 px-6 py-2 rounded-sm font-medium">
                    Open Swagger UI
                </a>
            </div>

        </div>
    </div>

</div>
@endsection
