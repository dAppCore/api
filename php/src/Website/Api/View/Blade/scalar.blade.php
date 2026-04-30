<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Reference - {{ config('core.app.name', 'Core PHP') }}</title>
    <meta name="description" content="{{ config('core.app.name', 'Core PHP') }} API Reference - Interactive documentation with code samples">
    <link rel="icon" type="image/png" href="/favicon.png">
    <style>
        * { box-sizing: border-box; }
        body { margin: 0; }
        .api-nav {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            z-index: 9999;
            height: 40px;
            background: #1a1a2e;
            border-bottom: 1px solid #2d2d44;
            display: flex;
            align-items: center;
            padding: 0 16px;
            font-family: system-ui, sans-serif;
            font-size: 14px;
        }
        .api-nav a {
            color: #a0a0b8;
            text-decoration: none;
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .api-nav a:hover { color: #fff; }
        .api-nav svg { width: 16px; height: 16px; }
        .scalar-wrapper { padding-top: 40px; height: 100vh; }
        .scalar-app { --scalar-font: 'Inter', system-ui, sans-serif; }
    </style>
</head>
<body>
    <nav class="api-nav">
        <a href="{{ route('api.docs') }}">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back to API Docs
        </a>
    </nav>
    <div class="scalar-wrapper">
    <script
        id="api-reference"
        data-url="/openapi.json"
        data-configuration='{
            "theme": "kepler",
            "layout": "modern",
            "darkMode": true,
            "hiddenClients": ["unirest"],
            "defaultHttpClient": {
                "targetKey": "php",
                "clientKey": "guzzle"
            },
            "metaData": {
                "title": "{{ config('core.app.name', 'Core PHP') }} API",
                "description": "API documentation for {{ config('core.app.name', 'Core PHP') }}"
            }
        }'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    </div>
</body>
</html>
