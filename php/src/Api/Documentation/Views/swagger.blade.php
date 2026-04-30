<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="API Documentation - Swagger UI">
    <title>{{ config('api-docs.info.title', 'API Documentation') }} - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info { margin: 20px 0; }
        .swagger-ui .info .title { font-size: 28px; }
        .swagger-ui .scheme-container { background: transparent; box-shadow: none; padding: 0; }
        .swagger-ui .opblock-tag { font-size: 18px; }
        .swagger-ui .opblock .opblock-summary-operation-id { font-size: 13px; }

        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            body { background: #1a1a2e; }
            .swagger-ui { filter: invert(88%) hue-rotate(180deg); }
            .swagger-ui .opblock-body pre { filter: invert(100%) hue-rotate(180deg); }
            .swagger-ui img { filter: invert(100%) hue-rotate(180deg); }
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>

    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: @json($specUrl),
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
                docExpansion: @json($config['doc_expansion'] ?? 'none'),
                filter: @json($config['filter'] ?? true),
                showExtensions: @json($config['show_extensions'] ?? true),
                showCommonExtensions: @json($config['show_common_extensions'] ?? true),
                syntaxHighlight: {
                    activated: true,
                    theme: "monokai"
                },
                requestInterceptor: function(request) {
                    // Add any default headers here
                    return request;
                }
            });
        };
    </script>
</body>
</html>
