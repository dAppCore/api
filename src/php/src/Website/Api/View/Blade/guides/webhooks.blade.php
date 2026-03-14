@extends('api::layouts.docs')

@section('title', 'Webhooks')

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
                        <a href="#setup" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Setup
                        </a>
                    </li>
                    <li>
                        <a href="#events" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Event Types
                        </a>
                    </li>
                    <li>
                        <a href="#payload" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Payload Format
                        </a>
                    </li>
                    <li>
                        <a href="#headers" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Request Headers
                        </a>
                    </li>
                    <li>
                        <a href="#verification" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Signature Verification
                        </a>
                    </li>
                    <li>
                        <a href="#retry-policy" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Retry Policy
                        </a>
                    </li>
                    <li>
                        <a href="#best-practices" data-scrollspy-link class="block px-3 py-2 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200 rounded-sm relative before:absolute before:inset-y-1 before:left-0 before:w-0.5 before:rounded-full">
                            Best Practices
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
                    <li class="text-zinc-800 dark:text-zinc-200">Webhooks</li>
                </ol>
            </nav>

            <h1 class="h1 mb-4 text-zinc-800 dark:text-zinc-100">Webhooks</h1>
            <p class="text-xl text-zinc-600 dark:text-zinc-400 mb-12">
                Receive real-time notifications for events in your workspace with cryptographically signed payloads.
            </p>

            {{-- Overview --}}
            <section id="overview" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Overview</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Webhooks allow your application to receive real-time HTTP callbacks when events occur in your workspace. Instead of polling the API, webhooks push data to your server as events happen.
                </p>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    All webhook requests are cryptographically signed using HMAC-SHA256, allowing you to verify that requests genuinely came from our platform and haven't been tampered with.
                </p>
                <div class="text-sm p-4 bg-amber-50 border border-amber-200 rounded-sm dark:bg-amber-900/20 dark:border-amber-800">
                    <div class="flex items-start">
                        <svg class="fill-amber-500 shrink-0 mr-3 mt-0.5" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 12a1 1 0 1 1 0-2 1 1 0 0 1 0 2zm1-3H7V4h2v5z"/>
                        </svg>
                        <p class="text-amber-800 dark:text-amber-200">
                            <strong>Security:</strong> Always verify webhook signatures before processing. Never trust unverified webhook requests.
                        </p>
                    </div>
                </div>
            </section>

            {{-- Setup --}}
            <section id="setup" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Setup</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    To configure webhooks:
                </p>
                <ol class="list-decimal list-inside space-y-2 text-zinc-600 dark:text-zinc-400 mb-4">
                    <li>Go to <strong>Settings &rarr; Webhooks</strong> in your workspace</li>
                    <li>Click <strong>Add Webhook</strong></li>
                    <li>Enter your endpoint URL (must be HTTPS in production)</li>
                    <li>Select the events you want to receive</li>
                    <li>Save and securely store your webhook secret</li>
                </ol>
                <div class="text-sm p-4 bg-blue-50 border border-blue-200 rounded-sm dark:bg-blue-900/20 dark:border-blue-800">
                    <div class="flex items-start">
                        <svg class="fill-blue-500 shrink-0 mr-3 mt-0.5" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm1 12H7V7h2v5zm0-6H7V4h2v2z"/>
                        </svg>
                        <p class="text-blue-800 dark:text-blue-200">
                            Your webhook secret is only shown once when you create the endpoint. Store it securely - you'll need it to verify incoming webhooks.
                        </p>
                    </div>
                </div>
            </section>

            {{-- Events --}}
            <section id="events" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Event Types</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Available webhook events:
                </p>

                <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Event</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Description</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">bio.created</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A new biolink was created</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">bio.updated</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A biolink was updated</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">bio.deleted</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A biolink was deleted</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">link.created</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A new link was created</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">link.clicked</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A link was clicked (high volume)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">qrcode.created</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A QR code was generated</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">qrcode.scanned</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">A QR code was scanned (high volume)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">*</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Subscribe to all events (wildcard)</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            {{-- Payload --}}
            <section id="payload" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Payload Format</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Webhook payloads are sent as JSON with a consistent structure:
                </p>

                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300">{
  <span class="text-blue-400">"id"</span>: <span class="text-green-400">"evt_abc123xyz456"</span>,
  <span class="text-blue-400">"type"</span>: <span class="text-green-400">"bio.created"</span>,
  <span class="text-blue-400">"created_at"</span>: <span class="text-green-400">"2024-01-15T10:30:00Z"</span>,
  <span class="text-blue-400">"workspace_id"</span>: <span class="text-amber-400">1</span>,
  <span class="text-blue-400">"data"</span>: {
    <span class="text-blue-400">"id"</span>: <span class="text-amber-400">123</span>,
    <span class="text-blue-400">"url"</span>: <span class="text-green-400">"mypage"</span>,
    <span class="text-blue-400">"type"</span>: <span class="text-green-400">"biolink"</span>
  }
}</code></pre>
                </div>
            </section>

            {{-- Headers --}}
            <section id="headers" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Request Headers</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    Every webhook request includes the following headers:
                </p>

                <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Header</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Description</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-Webhook-Signature</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">HMAC-SHA256 signature for verification</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-Webhook-Timestamp</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Unix timestamp when the webhook was sent</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-Webhook-Event</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">The event type (e.g., <code>bio.created</code>)</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">X-Webhook-Id</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Unique delivery ID for idempotency</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4"><code class="px-1.5 py-0.5 bg-zinc-100 dark:bg-zinc-800 rounded text-xs">Content-Type</code></td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Always <code>application/json</code></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            {{-- Verification --}}
            <section id="verification" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Signature Verification</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    To verify a webhook signature, compute the HMAC-SHA256 of the timestamp concatenated with the raw request body using your webhook secret. The signature includes the timestamp to prevent replay attacks.
                </p>

                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">Verification Algorithm</h3>
                <ol class="list-decimal list-inside space-y-2 text-zinc-600 dark:text-zinc-400 mb-6">
                    <li>Extract <code>X-Webhook-Signature</code> and <code>X-Webhook-Timestamp</code> headers</li>
                    <li>Concatenate: <code>timestamp + "." + raw_request_body</code></li>
                    <li>Compute: <code>HMAC-SHA256(concatenated_string, your_webhook_secret)</code></li>
                    <li>Compare using timing-safe comparison (prevents timing attacks)</li>
                    <li>Verify timestamp is within 5 minutes of current time (prevents replay attacks)</li>
                </ol>

                {{-- PHP Example --}}
                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">PHP</h3>
                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-6">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">webhook-handler.php</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-purple-400">&lt;?php</span>

<span class="text-zinc-500">// Get request data</span>
<span class="text-purple-400">$payload</span> = <span class="text-teal-400">file_get_contents</span>(<span class="text-green-400">'php://input'</span>);
<span class="text-purple-400">$signature</span> = <span class="text-purple-400">$_SERVER</span>[<span class="text-green-400">'HTTP_X_WEBHOOK_SIGNATURE'</span>] ?? <span class="text-green-400">''</span>;
<span class="text-purple-400">$timestamp</span> = <span class="text-purple-400">$_SERVER</span>[<span class="text-green-400">'HTTP_X_WEBHOOK_TIMESTAMP'</span>] ?? <span class="text-green-400">''</span>;
<span class="text-purple-400">$secret</span> = <span class="text-teal-400">getenv</span>(<span class="text-green-400">'WEBHOOK_SECRET'</span>);

<span class="text-zinc-500">// Verify timestamp (5 minute tolerance)</span>
<span class="text-purple-400">$tolerance</span> = <span class="text-amber-400">300</span>;
<span class="text-pink-400">if</span> (<span class="text-teal-400">abs</span>(<span class="text-teal-400">time</span>() - (<span class="text-pink-400">int</span>)<span class="text-purple-400">$timestamp</span>) > <span class="text-purple-400">$tolerance</span>) {
    <span class="text-teal-400">http_response_code</span>(<span class="text-amber-400">401</span>);
    <span class="text-pink-400">die</span>(<span class="text-green-400">'Webhook timestamp expired'</span>);
}

<span class="text-zinc-500">// Compute expected signature</span>
<span class="text-purple-400">$signedPayload</span> = <span class="text-purple-400">$timestamp</span> . <span class="text-green-400">'.'</span> . <span class="text-purple-400">$payload</span>;
<span class="text-purple-400">$expectedSignature</span> = <span class="text-teal-400">hash_hmac</span>(<span class="text-green-400">'sha256'</span>, <span class="text-purple-400">$signedPayload</span>, <span class="text-purple-400">$secret</span>);

<span class="text-zinc-500">// Verify signature (timing-safe comparison)</span>
<span class="text-pink-400">if</span> (!<span class="text-teal-400">hash_equals</span>(<span class="text-purple-400">$expectedSignature</span>, <span class="text-purple-400">$signature</span>)) {
    <span class="text-teal-400">http_response_code</span>(<span class="text-amber-400">401</span>);
    <span class="text-pink-400">die</span>(<span class="text-green-400">'Invalid webhook signature'</span>);
}

<span class="text-zinc-500">// Signature valid - process the webhook</span>
<span class="text-purple-400">$event</span> = <span class="text-teal-400">json_decode</span>(<span class="text-purple-400">$payload</span>, <span class="text-pink-400">true</span>);
<span class="text-teal-400">processWebhook</span>(<span class="text-purple-400">$event</span>);</code></pre>
                </div>

                {{-- Node.js Example --}}
                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">Node.js</h3>
                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-6">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">webhook-handler.js</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-pink-400">const</span> crypto = <span class="text-teal-400">require</span>(<span class="text-green-400">'crypto'</span>);
<span class="text-pink-400">const</span> express = <span class="text-teal-400">require</span>(<span class="text-green-400">'express'</span>);

<span class="text-pink-400">const</span> app = <span class="text-teal-400">express</span>();
app.<span class="text-teal-400">use</span>(express.<span class="text-teal-400">raw</span>({ type: <span class="text-green-400">'application/json'</span> }));

<span class="text-pink-400">const</span> WEBHOOK_SECRET = process.env.WEBHOOK_SECRET;
<span class="text-pink-400">const</span> TOLERANCE = <span class="text-amber-400">300</span>; <span class="text-zinc-500">// 5 minutes</span>

app.<span class="text-teal-400">post</span>(<span class="text-green-400">'/webhook'</span>, (req, res) => {
    <span class="text-pink-400">const</span> signature = req.headers[<span class="text-green-400">'x-webhook-signature'</span>];
    <span class="text-pink-400">const</span> timestamp = req.headers[<span class="text-green-400">'x-webhook-timestamp'</span>];
    <span class="text-pink-400">const</span> payload = req.body;

    <span class="text-zinc-500">// Verify timestamp</span>
    <span class="text-pink-400">const</span> now = Math.<span class="text-teal-400">floor</span>(Date.<span class="text-teal-400">now</span>() / <span class="text-amber-400">1000</span>);
    <span class="text-pink-400">if</span> (Math.<span class="text-teal-400">abs</span>(now - <span class="text-teal-400">parseInt</span>(timestamp)) > TOLERANCE) {
        <span class="text-pink-400">return</span> res.<span class="text-teal-400">status</span>(<span class="text-amber-400">401</span>).<span class="text-teal-400">send</span>(<span class="text-green-400">'Webhook timestamp expired'</span>);
    }

    <span class="text-zinc-500">// Compute expected signature</span>
    <span class="text-pink-400">const</span> signedPayload = <span class="text-green-400">`${</span>timestamp<span class="text-green-400">}.${</span>payload<span class="text-green-400">}`</span>;
    <span class="text-pink-400">const</span> expectedSignature = crypto
        .<span class="text-teal-400">createHmac</span>(<span class="text-green-400">'sha256'</span>, WEBHOOK_SECRET)
        .<span class="text-teal-400">update</span>(signedPayload)
        .<span class="text-teal-400">digest</span>(<span class="text-green-400">'hex'</span>);

    <span class="text-zinc-500">// Verify signature (timing-safe comparison)</span>
    <span class="text-pink-400">if</span> (!crypto.<span class="text-teal-400">timingSafeEqual</span>(
        Buffer.<span class="text-teal-400">from</span>(expectedSignature),
        Buffer.<span class="text-teal-400">from</span>(signature)
    )) {
        <span class="text-pink-400">return</span> res.<span class="text-teal-400">status</span>(<span class="text-amber-400">401</span>).<span class="text-teal-400">send</span>(<span class="text-green-400">'Invalid webhook signature'</span>);
    }

    <span class="text-zinc-500">// Signature valid - process the webhook</span>
    <span class="text-pink-400">const</span> event = JSON.<span class="text-teal-400">parse</span>(payload);
    <span class="text-teal-400">processWebhook</span>(event);
    res.<span class="text-teal-400">status</span>(<span class="text-amber-400">200</span>).<span class="text-teal-400">send</span>(<span class="text-green-400">'OK'</span>);
});</code></pre>
                </div>

                {{-- Python Example --}}
                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">Python</h3>
                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-6">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">webhook_handler.py</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-pink-400">import</span> hmac
<span class="text-pink-400">import</span> hashlib
<span class="text-pink-400">import</span> time
<span class="text-pink-400">import</span> os
<span class="text-pink-400">from</span> flask <span class="text-pink-400">import</span> Flask, request, abort

app = Flask(__name__)
WEBHOOK_SECRET = os.environ[<span class="text-green-400">'WEBHOOK_SECRET'</span>]
TOLERANCE = <span class="text-amber-400">300</span>  <span class="text-zinc-500"># 5 minutes</span>

<span class="text-pink-400">@</span>app.route(<span class="text-green-400">'/webhook'</span>, methods=[<span class="text-green-400">'POST'</span>])
<span class="text-pink-400">def</span> <span class="text-teal-400">webhook</span>():
    signature = request.headers.get(<span class="text-green-400">'X-Webhook-Signature'</span>, <span class="text-green-400">''</span>)
    timestamp = request.headers.get(<span class="text-green-400">'X-Webhook-Timestamp'</span>, <span class="text-green-400">''</span>)
    payload = request.get_data(as_text=<span class="text-pink-400">True</span>)

    <span class="text-zinc-500"># Verify timestamp</span>
    <span class="text-pink-400">if</span> abs(time.time() - int(timestamp)) > TOLERANCE:
        abort(<span class="text-amber-400">401</span>, <span class="text-green-400">'Webhook timestamp expired'</span>)

    <span class="text-zinc-500"># Compute expected signature</span>
    signed_payload = <span class="text-green-400">f'{timestamp}.{payload}'</span>
    expected_signature = hmac.new(
        WEBHOOK_SECRET.encode(),
        signed_payload.encode(),
        hashlib.sha256
    ).hexdigest()

    <span class="text-zinc-500"># Verify signature (timing-safe comparison)</span>
    <span class="text-pink-400">if not</span> hmac.compare_digest(expected_signature, signature):
        abort(<span class="text-amber-400">401</span>, <span class="text-green-400">'Invalid webhook signature'</span>)

    <span class="text-zinc-500"># Signature valid - process the webhook</span>
    event = request.get_json()
    process_webhook(event)
    <span class="text-pink-400">return</span> <span class="text-green-400">'OK'</span>, <span class="text-amber-400">200</span></code></pre>
                </div>

                {{-- Ruby Example --}}
                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">Ruby</h3>
                <div class="bg-zinc-800 rounded-sm overflow-hidden mb-6">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">webhook_handler.rb</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-pink-400">require</span> <span class="text-green-400">'sinatra'</span>
<span class="text-pink-400">require</span> <span class="text-green-400">'openssl'</span>
<span class="text-pink-400">require</span> <span class="text-green-400">'json'</span>

WEBHOOK_SECRET = ENV[<span class="text-green-400">'WEBHOOK_SECRET'</span>]
TOLERANCE = <span class="text-amber-400">300</span>  <span class="text-zinc-500"># 5 minutes</span>

post <span class="text-green-400">'/webhook'</span> <span class="text-pink-400">do</span>
  signature = request.env[<span class="text-green-400">'HTTP_X_WEBHOOK_SIGNATURE'</span>] || <span class="text-green-400">''</span>
  timestamp = request.env[<span class="text-green-400">'HTTP_X_WEBHOOK_TIMESTAMP'</span>] || <span class="text-green-400">''</span>
  payload = request.body.read

  <span class="text-zinc-500"># Verify timestamp</span>
  <span class="text-pink-400">if</span> (Time.now.to_i - timestamp.to_i).<span class="text-teal-400">abs</span> > TOLERANCE
    halt <span class="text-amber-400">401</span>, <span class="text-green-400">'Webhook timestamp expired'</span>
  <span class="text-pink-400">end</span>

  <span class="text-zinc-500"># Compute expected signature</span>
  signed_payload = <span class="text-green-400">"#{timestamp}.#{payload}"</span>
  expected_signature = OpenSSL::HMAC.hexdigest(
    <span class="text-green-400">'sha256'</span>,
    WEBHOOK_SECRET,
    signed_payload
  )

  <span class="text-zinc-500"># Verify signature (timing-safe comparison)</span>
  <span class="text-pink-400">unless</span> Rack::Utils.secure_compare(expected_signature, signature)
    halt <span class="text-amber-400">401</span>, <span class="text-green-400">'Invalid webhook signature'</span>
  <span class="text-pink-400">end</span>

  <span class="text-zinc-500"># Signature valid - process the webhook</span>
  event = JSON.parse(payload)
  process_webhook(event)
  <span class="text-amber-400">200</span>
<span class="text-pink-400">end</span></code></pre>
                </div>

                {{-- Go Example --}}
                <h3 class="h4 mb-3 text-zinc-800 dark:text-zinc-100">Go</h3>
                <div class="bg-zinc-800 rounded-sm overflow-hidden">
                    <div class="flex items-center justify-between px-4 py-2 border-b border-zinc-700">
                        <span class="text-sm text-zinc-400">webhook_handler.go</span>
                    </div>
                    <pre class="overflow-x-auto p-4 text-sm"><code class="font-pt-mono text-zinc-300"><span class="text-pink-400">package</span> main

<span class="text-pink-400">import</span> (
    <span class="text-green-400">"crypto/hmac"</span>
    <span class="text-green-400">"crypto/sha256"</span>
    <span class="text-green-400">"crypto/subtle"</span>
    <span class="text-green-400">"encoding/hex"</span>
    <span class="text-green-400">"io"</span>
    <span class="text-green-400">"math"</span>
    <span class="text-green-400">"net/http"</span>
    <span class="text-green-400">"os"</span>
    <span class="text-green-400">"strconv"</span>
    <span class="text-green-400">"time"</span>
)

<span class="text-pink-400">const</span> tolerance = <span class="text-amber-400">300</span> <span class="text-zinc-500">// 5 minutes</span>

<span class="text-pink-400">func</span> <span class="text-teal-400">webhookHandler</span>(w http.ResponseWriter, r *http.Request) {
    signature := r.Header.Get(<span class="text-green-400">"X-Webhook-Signature"</span>)
    timestamp := r.Header.Get(<span class="text-green-400">"X-Webhook-Timestamp"</span>)
    secret := os.Getenv(<span class="text-green-400">"WEBHOOK_SECRET"</span>)

    payload, _ := io.ReadAll(r.Body)

    <span class="text-zinc-500">// Verify timestamp</span>
    ts, _ := strconv.ParseInt(timestamp, <span class="text-amber-400">10</span>, <span class="text-amber-400">64</span>)
    <span class="text-pink-400">if</span> math.Abs(<span class="text-teal-400">float64</span>(time.Now().Unix()-ts)) > tolerance {
        http.Error(w, <span class="text-green-400">"Webhook timestamp expired"</span>, <span class="text-amber-400">401</span>)
        <span class="text-pink-400">return</span>
    }

    <span class="text-zinc-500">// Compute expected signature</span>
    signedPayload := timestamp + <span class="text-green-400">"."</span> + <span class="text-teal-400">string</span>(payload)
    mac := hmac.New(sha256.New, []<span class="text-teal-400">byte</span>(secret))
    mac.Write([]<span class="text-teal-400">byte</span>(signedPayload))
    expectedSignature := hex.EncodeToString(mac.Sum(<span class="text-pink-400">nil</span>))

    <span class="text-zinc-500">// Verify signature (timing-safe comparison)</span>
    <span class="text-pink-400">if</span> subtle.ConstantTimeCompare(
        []<span class="text-teal-400">byte</span>(expectedSignature),
        []<span class="text-teal-400">byte</span>(signature),
    ) != <span class="text-amber-400">1</span> {
        http.Error(w, <span class="text-green-400">"Invalid webhook signature"</span>, <span class="text-amber-400">401</span>)
        <span class="text-pink-400">return</span>
    }

    <span class="text-zinc-500">// Signature valid - process the webhook</span>
    processWebhook(payload)
    w.WriteHeader(http.StatusOK)
}</code></pre>
                </div>
            </section>

            {{-- Retry Policy --}}
            <section id="retry-policy" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Retry Policy</h2>
                <p class="text-zinc-600 dark:text-zinc-400 mb-4">
                    If your endpoint returns a non-2xx status code or times out, we'll retry with exponential backoff:
                </p>

                <div class="overflow-x-auto mb-4">
                    <table class="w-full text-sm">
                        <thead>
                            <tr class="border-b border-zinc-200 dark:border-zinc-700">
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Attempt</th>
                                <th class="text-left py-3 px-4 font-medium text-zinc-800 dark:text-zinc-200">Delay</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                            <tr>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">1 (initial)</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">Immediate</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">2</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">1 minute</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">3</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">5 minutes</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">4</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">30 minutes</td>
                            </tr>
                            <tr>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">5 (final)</td>
                                <td class="py-3 px-4 text-zinc-600 dark:text-zinc-400">2 hours</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <p class="text-zinc-600 dark:text-zinc-400">
                    After 5 failed attempts, the delivery is marked as failed. If your endpoint fails 10 consecutive deliveries, it will be automatically disabled. You can re-enable it from your webhook settings.
                </p>
            </section>

            {{-- Best Practices --}}
            <section id="best-practices" data-scrollspy-target class="mb-12">
                <h2 class="h3 mb-4 text-zinc-800 dark:text-zinc-100">Best Practices</h2>
                <ul class="space-y-3 text-zinc-600 dark:text-zinc-400">
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Always verify signatures</strong> - Never process webhooks without verification</span>
                    </li>
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Respond quickly</strong> - Return 200 within 30 seconds to avoid timeouts</span>
                    </li>
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Process asynchronously</strong> - Queue webhook processing for long-running tasks</span>
                    </li>
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Handle duplicates</strong> - Use <code>X-Webhook-Id</code> for idempotency</span>
                    </li>
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Use HTTPS</strong> - Always use HTTPS endpoints in production</span>
                    </li>
                    <li class="flex items-start">
                        <svg class="fill-green-500 shrink-0 mr-3 mt-1" width="16" height="16" viewBox="0 0 16 16">
                            <path d="M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm3.78 5.22a.75.75 0 0 1 0 1.06l-4.5 4.5a.75.75 0 0 1-1.06 0l-2-2a.75.75 0 1 1 1.06-1.06L6.75 9.19l3.97-3.97a.75.75 0 0 1 1.06 0z"/>
                        </svg>
                        <span><strong>Rotate secrets regularly</strong> - Rotate your webhook secret periodically</span>
                    </li>
                </ul>
            </section>

            {{-- Next steps --}}
            <div class="flex items-center justify-between pt-8 border-t border-zinc-200 dark:border-zinc-700">
                <a href="{{ route('api.guides.qrcodes') }}" class="text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200">
                    &larr; QR Code Generation
                </a>
                <a href="{{ route('api.guides.errors') }}" class="text-cyan-600 hover:text-cyan-700 dark:hover:text-cyan-500 font-medium">
                    Error Handling &rarr;
                </a>
            </div>

        </div>
    </div>

</div>
@endsection
