<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="API Documentation - Stoplight Elements">
    <title>{{ config('api-docs.info.title', 'API Documentation') }} - Stoplight</title>
    <style>
        html, body {
            margin: 0;
            min-height: 100%;
            background: #0f172a;
        }

        elements-api {
            height: 100vh;
            width: 100%;
            display: block;
        }
    </style>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
</head>
<body>
    <elements-api
        apiDescriptionUrl="{{ $specUrl }}"
        router="hash"
        layout="{{ $config['layout'] ?? 'sidebar' }}"
        theme="{{ $config['theme'] ?? 'dark' }}"
        @if($config['hide_try_it'] ?? false) hideTryIt="true" @endif
    ></elements-api>

    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
</body>
</html>
