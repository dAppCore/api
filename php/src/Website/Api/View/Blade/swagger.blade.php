@extends('api::layouts.docs')

@section('title', 'Swagger UI')

@push('head')
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
<style>
    .swagger-ui .topbar { display: none; }
    .swagger-ui .info { margin: 20px 0; }
    .swagger-ui .info .title { font-size: 28px; }
    .swagger-ui .scheme-container { background: transparent; box-shadow: none; padding: 0; }
    .swagger-ui .opblock-tag { font-size: 18px; }
    .swagger-ui .opblock .opblock-summary-operation-id { font-size: 13px; }
    .dark .swagger-ui { filter: invert(88%) hue-rotate(180deg); }
    .dark .swagger-ui .opblock-body pre { filter: invert(100%) hue-rotate(180deg); }
    .dark .swagger-ui img { filter: invert(100%) hue-rotate(180deg); }
</style>
@endpush

@section('content')
<div class="max-w-7xl mx-auto px-4 sm:px-6 py-8">
    <div class="mb-8">
        <h1 class="h2 mb-2 text-zinc-800 dark:text-zinc-100">Swagger UI</h1>
        <p class="text-zinc-600 dark:text-zinc-400">
            Interactive API explorer. Try out endpoints directly from your browser.
        </p>
    </div>

    <div id="swagger-ui" class="bg-white dark:bg-zinc-800 rounded-sm border border-zinc-200 dark:border-zinc-700 p-4"></div>
</div>
@endsection

@push('scripts')
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
<script>
    window.onload = function() {
        window.ui = SwaggerUIBundle({
            url: "/openapi.json",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "BaseLayout",
            defaultModelsExpandDepth: -1,
            docExpansion: 'none',
            filter: true,
            showExtensions: true,
            showCommonExtensions: true
        });
    };
</script>
@endpush
