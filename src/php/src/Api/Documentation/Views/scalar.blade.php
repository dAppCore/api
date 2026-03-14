<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="API Documentation - Scalar">
    <title>{{ config('api-docs.info.title', 'API Documentation') }}</title>
    <style>
        body { margin: 0; }
    </style>
</head>
<body>
    <script
        id="api-reference"
        data-url="{{ $specUrl }}"
        data-configuration='{
            "theme": "{{ $config['theme'] ?? 'default' }}",
            "showSidebar": {{ ($config['show_sidebar'] ?? true) ? 'true' : 'false' }},
            "hideDownloadButton": {{ ($config['hide_download_button'] ?? false) ? 'true' : 'false' }},
            "hideModels": {{ ($config['hide_models'] ?? false) ? 'true' : 'false' }},
            "darkMode": false,
            "layout": "modern",
            "searchHotKey": "k"
        }'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>
