@extends('api::layouts.docs')

@section('title', 'SDK Downloads')
@section('description', 'Download generated API SDKs for common languages and platforms.')

@section('content')
@php
    $sdkTargets = [
        ['name' => 'PHP', 'slug' => 'php', 'install' => 'composer require dappcore/core-sdk'],
        ['name' => 'TypeScript / JS', 'slug' => 'js', 'install' => 'npm install @dappcore/core-sdk'],
        ['name' => 'Python', 'slug' => 'python', 'install' => 'pip install core'],
        ['name' => 'Go', 'slug' => 'go', 'install' => 'go get github.com/dappcore/core-go'],
    ];
@endphp

<div class="max-w-6xl mx-auto px-4 sm:px-6 py-12 md:py-20">
    <div class="max-w-3xl mx-auto text-center mb-12">
        <span class="font-nycd text-xl text-cyan-600">SDK Distribution</span>
        <h1 class="h1 mt-4 mb-4 text-zinc-800 dark:text-zinc-100">Generated client libraries</h1>
        <p class="text-xl text-zinc-600 dark:text-zinc-400">
            Export OpenAPI once and generate client SDKs for the languages your integrations use most.
        </p>
    </div>

    <div class="grid md:grid-cols-2 gap-6">
        @foreach($sdkTargets as $sdk)
            <a href="{{ route('api.sdks.language', $sdk['slug']) }}" class="block rounded-sm border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-6 hover:border-cyan-300 dark:hover:border-cyan-600 transition-colors">
                <div class="flex items-center justify-between gap-4">
                    <div>
                        <h2 class="h3 text-zinc-800 dark:text-zinc-100">{{ $sdk['name'] }}</h2>
                        <p class="mt-2 text-sm text-zinc-500 dark:text-zinc-400">{{ $sdk['install'] }}</p>
                    </div>
                    <span class="inline-flex h-10 w-10 items-center justify-center rounded-sm bg-cyan-100 dark:bg-cyan-900/30 text-cyan-700 dark:text-cyan-400">
                        <i class="fa-solid fa-download"></i>
                    </span>
                </div>
            </a>
        @endforeach
    </div>

    @if($language !== null)
        <div class="mt-10 rounded-sm border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-900 p-6">
            <p class="text-sm font-medium text-zinc-500 dark:text-zinc-400">Selected SDK</p>
            <h2 class="h3 mt-2 text-zinc-800 dark:text-zinc-100">{{ strtoupper($language) }}</h2>
            <p class="mt-3 text-zinc-600 dark:text-zinc-400">
                SDK bundles are generated from the current OpenAPI export. If you need the latest spec, fetch <code>/openapi.json</code> or <code>/openapi.yaml</code> first.
            </p>
        </div>
    @endif
</div>
@endsection
