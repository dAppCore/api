<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Reference - {{ config('core.app.name', 'Core PHP') }}</title>
    <meta name="description" content="{{ config('core.app.name', 'Core PHP') }} API Reference - ReDoc documentation">
    <link rel="icon" type="image/png" href="/favicon.png">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
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
        #redoc-container { padding-top: 40px; }
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
    <div id="redoc-container"></div>
    <script src="https://unpkg.com/redoc@latest/bundles/redoc.standalone.js"></script>
    <script>
        Redoc.init('{{ route('api.openapi.json') }}', {
            theme: {
                typography: {
                    fontFamily: 'Inter, system-ui, sans-serif',
                    headings: { fontFamily: 'Inter, system-ui, sans-serif' }
                },
                colors: {
                    primary: { main: '#3b82f6' }
                },
                sidebar: {
                    backgroundColor: '#1e293b',
                    textColor: '#94a3b8'
                },
                rightPanel: {
                    backgroundColor: '#0f172a'
                }
            },
            scrollYOffset: 40
        }, document.getElementById('redoc-container'));
    </script>
</body>
</html>
