<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ config('core.app.name', 'Core PHP') }} API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html {
            box-sizing: border-box;
            overflow-y: scroll;
        }
        *,
        *:before,
        *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            display: none;
        }
        .swagger-ui .info {
            margin: 20px 0;
        }
        .swagger-ui .info .title {
            font-size: 32px;
        }
        /* Custom header */
        .api-header {
            background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
            padding: 24px 32px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .api-header-title {
            color: white;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            font-size: 20px;
            font-weight: 600;
            margin: 0;
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .api-header-title svg {
            width: 24px;
            height: 24px;
        }
        .api-header-links {
            display: flex;
            gap: 16px;
        }
        .api-header-links a {
            color: #94a3b8;
            text-decoration: none;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            font-size: 14px;
            transition: color 0.2s;
        }
        .api-header-links a:hover {
            color: white;
        }
    </style>
</head>
<body>
    <header class="api-header">
        <h1 class="api-header-title">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" />
            </svg>
            {{ config('core.app.name', 'Core PHP') }} API
        </h1>
        <nav class="api-header-links">
            <a href="{{ config('app.url') }}">{{ config('core.app.name', 'Core PHP') }}</a>
            <a href="/openapi.json" target="_blank">OpenAPI JSON</a>
            <a href="{{ str_replace('api.', 'mcp.', request()->getSchemeAndHttpHost()) }}">MCP Portal</a>
        </nav>
    </header>

    <div id="swagger-ui"></div>

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
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: -1,
                docExpansion: 'list',
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
        };
    </script>
</body>
</html>
