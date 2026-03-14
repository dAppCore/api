<?php

declare(strict_types=1);

use Core\Api\Middleware\PublicApiCors;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

// ─────────────────────────────────────────────────────────────────────────────
// OPTIONS Preflight Requests
// ─────────────────────────────────────────────────────────────────────────────

describe('OPTIONS Preflight Requests', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('returns 204 status for OPTIONS request', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response('Should not reach'));

        expect($response->getStatusCode())->toBe(204);
    });

    it('returns empty body for OPTIONS preflight', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response('Should not reach'));

        expect($response->getContent())->toBe('');
    });

    it('does not call the next handler for OPTIONS requests', function () {
        $request = createCorsRequest('OPTIONS');
        $called = false;

        $this->middleware->handle($request, function () use (&$called) {
            $called = true;

            return new Response('');
        });

        expect($called)->toBeFalse();
    });

    it('includes all required CORS headers on OPTIONS response', function () {
        $request = createCorsRequest('OPTIONS', ['Origin' => 'https://example.com']);

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->has('Access-Control-Allow-Origin'))->toBeTrue();
        expect($response->headers->has('Access-Control-Allow-Methods'))->toBeTrue();
        expect($response->headers->has('Access-Control-Allow-Headers'))->toBeTrue();
        expect($response->headers->has('Access-Control-Max-Age'))->toBeTrue();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// CORS Headers on Regular Requests
// ─────────────────────────────────────────────────────────────────────────────

describe('CORS Headers on Regular Requests', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('adds CORS headers to GET response', function () {
        $request = createCorsRequest('GET', ['Origin' => 'https://example.com']);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->has('Access-Control-Allow-Origin'))->toBeTrue();
        expect($response->getContent())->toBe('OK');
    });

    it('adds CORS headers to POST response', function () {
        $request = createCorsRequest('POST', ['Origin' => 'https://example.com']);

        $response = $this->middleware->handle($request, fn () => new Response('Created', 201));

        expect($response->headers->has('Access-Control-Allow-Origin'))->toBeTrue();
        expect($response->getStatusCode())->toBe(201);
    });

    it('passes through request to next handler', function () {
        $request = createCorsRequest('GET');
        $nextCalled = false;

        $this->middleware->handle($request, function () use (&$nextCalled) {
            $nextCalled = true;

            return new Response('OK');
        });

        expect($nextCalled)->toBeTrue();
    });

    it('preserves original response content and status', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('{"data":"value"}', 200));

        expect($response->getContent())->toBe('{"data":"value"}');
        expect($response->getStatusCode())->toBe(200);
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Origin Handling
// ─────────────────────────────────────────────────────────────────────────────

describe('Origin Handling', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('echoes back the request Origin header on GET', function () {
        $request = createCorsRequest('GET', ['Origin' => 'https://customer-site.com']);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Origin'))->toBe('https://customer-site.com');
    });

    it('echoes back the Origin on OPTIONS preflight', function () {
        $request = createCorsRequest('OPTIONS', ['Origin' => 'https://app.example.org']);

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->get('Access-Control-Allow-Origin'))->toBe('https://app.example.org');
    });

    it('uses wildcard when no Origin header is present on GET', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Origin'))->toBe('*');
    });

    it('uses wildcard on OPTIONS when no Origin header is present', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->get('Access-Control-Allow-Origin'))->toBe('*');
    });

    it('accepts any origin (public API design for customer websites)', function () {
        $origins = [
            'https://customer1.com',
            'https://shop.example.org',
            'http://localhost:3000',
            'https://subdomain.company.co.uk',
        ];

        $middleware = new PublicApiCors();

        foreach ($origins as $origin) {
            $request = createCorsRequest('GET', ['Origin' => $origin]);
            $response = $middleware->handle($request, fn () => new Response('OK'));

            expect($response->headers->get('Access-Control-Allow-Origin'))->toBe($origin);
        }
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Allowed Methods
// ─────────────────────────────────────────────────────────────────────────────

describe('Allowed Methods', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('sets allowed methods header to GET, POST, OPTIONS', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Methods'))->toBe('GET, POST, OPTIONS');
    });

    it('includes OPTIONS in allowed methods on preflight response', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        $methods = $response->headers->get('Access-Control-Allow-Methods');
        expect($methods)->toContain('OPTIONS');
        expect($methods)->toContain('GET');
        expect($methods)->toContain('POST');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Allowed Request Headers
// ─────────────────────────────────────────────────────────────────────────────

describe('Allowed Request Headers', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('sets allowed headers to Content-Type, Accept, X-Requested-With', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->get('Access-Control-Allow-Headers'))
            ->toBe('Content-Type, Accept, X-Requested-With');
    });

    it('includes Content-Type in allowed headers on regular request', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Headers'))->toContain('Content-Type');
    });

    it('includes Accept in allowed headers', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Headers'))->toContain('Accept');
    });

    it('includes X-Requested-With in allowed headers', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Allow-Headers'))->toContain('X-Requested-With');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Exposed Response Headers
// ─────────────────────────────────────────────────────────────────────────────

describe('Exposed Response Headers', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('exposes rate limit headers to the browser', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        $exposed = $response->headers->get('Access-Control-Expose-Headers');
        expect($exposed)->toContain('X-RateLimit-Limit');
        expect($exposed)->toContain('X-RateLimit-Remaining');
        expect($exposed)->toContain('X-RateLimit-Reset');
        expect($exposed)->toContain('Retry-After');
    });

    it('exposes rate limit headers on OPTIONS preflight too', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        $exposed = $response->headers->get('Access-Control-Expose-Headers');
        expect($exposed)->toContain('X-RateLimit-Limit');
        expect($exposed)->toContain('X-RateLimit-Remaining');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Preflight Cache and Vary Headers
// ─────────────────────────────────────────────────────────────────────────────

describe('Preflight Cache and Vary Headers', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('sets Max-Age to 3600 seconds for preflight caching', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->get('Access-Control-Max-Age'))->toBe('3600');
    });

    it('sets Max-Age on regular responses', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Access-Control-Max-Age'))->toBe('3600');
    });

    it('sets Vary header to Origin for correct cache keying', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->get('Vary'))->toBe('Origin');
    });

    it('sets Vary header on OPTIONS preflight', function () {
        $request = createCorsRequest('OPTIONS');

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->get('Vary'))->toBe('Origin');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Security Boundaries
// ─────────────────────────────────────────────────────────────────────────────

describe('Security Boundaries', function () {
    beforeEach(function () {
        $this->middleware = new PublicApiCors();
    });

    it('does not set Access-Control-Allow-Credentials on regular requests', function () {
        $request = createCorsRequest('GET', ['Origin' => 'https://example.com']);

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->headers->has('Access-Control-Allow-Credentials'))->toBeFalse();
    });

    it('does not set Access-Control-Allow-Credentials on OPTIONS preflight', function () {
        $request = createCorsRequest('OPTIONS', ['Origin' => 'https://example.com']);

        $response = $this->middleware->handle($request, fn () => new Response(''));

        expect($response->headers->has('Access-Control-Allow-Credentials'))->toBeFalse();
    });

    it('allows requests without Origin header (non-browser clients)', function () {
        $request = createCorsRequest('GET');

        $response = $this->middleware->handle($request, fn () => new Response('OK'));

        expect($response->getStatusCode())->toBe(200);
        expect($response->headers->get('Access-Control-Allow-Origin'))->toBe('*');
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Helper Functions
// ─────────────────────────────────────────────────────────────────────────────

function createCorsRequest(string $method = 'GET', array $headers = []): Request
{
    $request = Request::create('/api/public/test', $method);

    foreach ($headers as $key => $value) {
        $request->headers->set($key, $value);
    }

    return $request;
}
