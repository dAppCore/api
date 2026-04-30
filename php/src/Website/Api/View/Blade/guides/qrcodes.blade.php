@extends('api::layouts.docs')

@section('title', 'QR Code Generation')

@section('content')
<div class="flex">

    {{-- Sidebar --}}
    <aside class="hidden lg:block fixed left-0 top-16 md:top-20 bottom-0 w-64 border-r border-zinc-200 dark:border-zinc-800">
        <div class="h-full px-4 py-8 overflow-y-auto no-scrollbar">
            <nav>
                <ul class="space-y-2">
                    <li>
                        <a href="#overview" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Overview
                        </a>
                    </li>
                    <li>
                        <a href="#biolink-qr" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Biolink QR Codes
                        </a>
                    </li>
                    <li>
                        <a href="#custom-qr" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Custom URL QR Codes
                        </a>
                    </li>
                    <li>
                        <a href="#options" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Customisation Options
                        </a>
                    </li>
                    <li>
                        <a href="#download" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Download Formats
                        </a>
                    </li>
                </ul>
            </nav>
        </div>
    </aside>

    {{-- Main content --}}
    <div class="lg:pl-64 w-full">
        <div class="max-w-3xl mx-auto px-4 sm:px-6 py-12">

            {{-- Breadcrumb --}}
            <nav class="mb-8">
                <ol class="flex items-center gap-2 text-sm">
                    <li><a href="{{ route('api.guides') }}" class="text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200">Guides</a></li>
                    <li class="text-zinc-400">/</li>
                    <li class="text-zinc-800 dark:text-zinc-200">QR Code Generation</li>
                </ol>
            </nav>

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">QR Code Generation</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
                Generate customisable QR codes for your biolinks or any URL.
            </p>

            {{-- Overview --}}
            <section id="overview" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Overview</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    The API provides two ways to generate QR codes:
                </p>
                <ul class="list-disc list-inside space-y-2 text-zinc-600 dark:text-zinc-400">
                    <li><strong>Biolink QR codes</strong> - Generate QR codes for your existing biolinks</li>
                    <li><strong>Custom URL QR codes</strong> - Generate QR codes for any URL</li>
                </ul>
            </section>

            {{-- Biolink QR --}}
            <section id="biolink-qr" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Biolink QR Codes</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Get QR code data for an existing biolink:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-4">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">cURL</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-teal-400">curl</span> <span class="text-zinc-500">--request</span> GET \
  <span class="text-zinc-500">--url</span> <span class="text-amber-400">'https://api.lthn.ai/api/v1/bio/1/qr'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Authorization: Bearer YOUR_API_KEY'</span></code></pre>
                </div>

                <p class="text-zinc-600 dark:text-zinc-400 mb-4">Response:</p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">{
  <span class="text-blue-400">"data"</span>: {
    <span class="text-blue-400">"svg"</span>: <span class="text-green-400">"&lt;svg&gt;...&lt;/svg&gt;"</span>,
    <span class="text-blue-400">"url"</span>: <span class="text-green-400">"https://example.com/mypage"</span>
  }
}</code></pre>
                </div>
            </section>

            {{-- Custom QR --}}
            <section id="custom-qr" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Custom URL QR Codes</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Generate a QR code for any URL:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-4">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">cURL</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-teal-400">curl</span> <span class="text-zinc-500">--request</span> POST \
  <span class="text-zinc-500">--url</span> <span class="text-amber-400">'https://api.lthn.ai/api/v1/qr/generate'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Authorization: Bearer YOUR_API_KEY'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Content-Type: application/json'</span> \
  <span class="text-zinc-500">--data</span> <span class="text-amber-400">'{
    "url": "https://example.com",
    "format": "svg",
    "size": 300
  }'</span></code></pre>
                </div>
            </section>

            {{-- Options --}}
            <section id="options" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Customisation Options</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Available customisation parameters:
                </p>

                <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Parameter</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Type</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Description</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">format</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">string</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Output format: <code>svg</code> or <code>png</code></td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">size</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">integer</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Size in pixels (100-2000)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">color</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">string</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Foreground colour (hex)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">background</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">string</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Background colour (hex)</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            {{-- Download --}}
            <section id="download" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Download Formats</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Download QR codes directly as image files:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-4">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">cURL</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-teal-400">curl</span> <span class="text-zinc-500">--request</span> GET \
  <span class="text-zinc-500">--url</span> <span class="text-amber-400">'https://api.lthn.ai/api/v1/bio/1/qr/download?format=png&size=500'</span> \
  <span class="text-zinc-500">--header</span> <span class="text-amber-400">'Authorization: Bearer YOUR_API_KEY'</span> \
  <span class="text-zinc-500">--output</span> <span class="text-amber-400">qrcode.png</span></code></pre>
                </div>

                <p class="text-zinc-600 dark:text-zinc-400">
                    The response is binary image data with appropriate Content-Type header.
                </p>
            </section>

            {{-- Next steps --}}
            <div class="flex items-center justify-between pt-8 border-t border-zinc-200 dark:border-zinc-700">
                <a href="{{ route('api.guides.authentication') }}" class="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">
                    &larr; Authentication
                </a>
                <a href="{{ route('api.guides.webhooks') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-blue-500 font-medium">
                    Webhooks &rarr;
                </a>
            </div>

        </div>
    </div>

</div>
@endsection
