@extends('api::layouts.docs')

@section('title', 'Guides')

@section('content')
<div class="max-w-7xl mx-auto px-4 sm:px-6 py-12">
    <div class="max-w-3xl">
        <h1 class="h2 mb-4 text-zinc-800 dark:text-zinc-100">Guides</h1>
        <p class="text-lg text-zinc-600 dark:text-zinc-400 mb-12">
            Step-by-step tutorials and best practices for integrating with the API.
        </p>
    </div>

    <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">

        {{-- Quick Start --}}
        <a href="{{ route('api.guides.quickstart') }}" class="group block p-6 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
            <div class="flex items-center gap-3 mb-3">
                <div class="w-8 h-8 flex items-center justify-center bg-cyan-100 dark:bg-cyan-900/30 rounded-sm">
                    <i class="fa-solid fa-bolt text-cyan-600 text-sm"></i>
                </div>
                <span class="text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider">Getting Started</span>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100 group-hover:text-cyan-600 dark:group-hover:text-cyan-500">Quick Start</h3>
            <p class="text-sm text-zinc-600 dark:text-zinc-400">Get up and running with the API in under 5 minutes.</p>
        </a>

        {{-- Authentication --}}
        <a href="{{ route('api.guides.authentication') }}" class="group block p-6 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
            <div class="flex items-center gap-3 mb-3">
                <div class="w-8 h-8 flex items-center justify-center bg-amber-100 dark:bg-amber-900/30 rounded-sm">
                    <i class="fa-solid fa-key text-amber-600 text-sm"></i>
                </div>
                <span class="text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider">Security</span>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100 group-hover:text-cyan-600 dark:group-hover:text-cyan-500">Authentication</h3>
            <p class="text-sm text-zinc-600 dark:text-zinc-400">Learn how to authenticate your API requests using API keys.</p>
        </a>

        {{-- Webhooks --}}
        <a href="{{ route('api.guides.webhooks') }}" class="group block p-6 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
            <div class="flex items-center gap-3 mb-3">
                <div class="w-8 h-8 flex items-center justify-center bg-rose-100 dark:bg-rose-900/30 rounded-sm">
                    <i class="fa-solid fa-satellite-dish text-rose-600 text-sm"></i>
                </div>
                <span class="text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider">Advanced</span>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100 group-hover:text-cyan-600 dark:group-hover:text-cyan-500">Webhooks</h3>
            <p class="text-sm text-zinc-600 dark:text-zinc-400">Receive real-time notifications for events in your workspace.</p>
        </a>

        {{-- Error Handling --}}
        <a href="{{ route('api.guides.errors') }}" class="group block p-6 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-sm hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
            <div class="flex items-center gap-3 mb-3">
                <div class="w-8 h-8 flex items-center justify-center bg-zinc-100 dark:bg-zinc-700 rounded-sm">
                    <i class="fa-solid fa-triangle-exclamation text-zinc-600 dark:text-zinc-400 text-sm"></i>
                </div>
                <span class="text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider">Reference</span>
            </div>
            <h3 class="h4 mb-2 text-zinc-800 dark:text-zinc-100 group-hover:text-cyan-600 dark:group-hover:text-cyan-500">Error Handling</h3>
            <p class="text-sm text-zinc-600 dark:text-zinc-400">Understand API error codes and how to handle them gracefully.</p>
        </a>

    </div>
</div>
@endsection
