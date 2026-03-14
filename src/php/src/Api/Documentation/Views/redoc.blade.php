<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="API Documentation - ReDoc">
    <title>{{ config('api-docs.info.title', 'API Documentation') }} - ReDoc</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        }

        /* Custom ReDoc theme overrides */
        .redoc-wrap {
            --primary-color: #3b82f6;
            --primary-color-dark: #2563eb;
            --selection-color: rgba(59, 130, 246, 0.1);
        }

        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            body {
                background: #1a1a2e;
            }
        }
    </style>
</head>
<body>
    <redoc spec-url="{{ $specUrl }}"
           expand-responses="200,201"
           path-in-middle-panel
           hide-hostname
           hide-download-button
           required-props-first
           sort-props-alphabetically
           no-auto-auth
           theme='{
               "colors": {
                   "primary": { "main": "#3b82f6" }
               },
               "typography": {
                   "fontFamily": "Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif",
                   "headings": { "fontFamily": "Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif" },
                   "code": { "fontFamily": "\"JetBrains Mono\", \"Fira Code\", Consolas, monospace" }
               },
               "sidebar": {
                   "width": "280px"
               },
               "rightPanel": {
                   "backgroundColor": "#1e293b"
               }
           }'>
    </redoc>

    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>
