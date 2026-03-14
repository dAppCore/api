<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Examples;

/**
 * Common API Examples.
 *
 * Provides example requests and responses for documentation.
 */
class CommonExamples
{
    /**
     * Get example for pagination parameters.
     */
    public static function paginationParams(): array
    {
        return [
            'page' => [
                'name' => 'page',
                'in' => 'query',
                'description' => 'Page number for pagination',
                'required' => false,
                'schema' => [
                    'type' => 'integer',
                    'minimum' => 1,
                    'default' => 1,
                    'example' => 1,
                ],
            ],
            'per_page' => [
                'name' => 'per_page',
                'in' => 'query',
                'description' => 'Number of items per page',
                'required' => false,
                'schema' => [
                    'type' => 'integer',
                    'minimum' => 1,
                    'maximum' => 100,
                    'default' => 25,
                    'example' => 25,
                ],
            ],
        ];
    }

    /**
     * Get example for sorting parameters.
     */
    public static function sortingParams(): array
    {
        return [
            'sort' => [
                'name' => 'sort',
                'in' => 'query',
                'description' => 'Field to sort by (prefix with - for descending)',
                'required' => false,
                'schema' => [
                    'type' => 'string',
                    'example' => '-created_at',
                ],
            ],
        ];
    }

    /**
     * Get example for filtering parameters.
     */
    public static function filteringParams(): array
    {
        return [
            'filter' => [
                'name' => 'filter',
                'in' => 'query',
                'description' => 'Filter parameters in the format filter[field]=value',
                'required' => false,
                'style' => 'deepObject',
                'explode' => true,
                'schema' => [
                    'type' => 'object',
                    'additionalProperties' => [
                        'type' => 'string',
                    ],
                ],
                'example' => [
                    'status' => 'active',
                    'created_at[gte]' => '2024-01-01',
                ],
            ],
        ];
    }

    /**
     * Get example paginated response.
     */
    public static function paginatedResponse(string $dataExample = '[]'): array
    {
        return [
            'data' => json_decode($dataExample, true) ?? [],
            'links' => [
                'first' => 'https://api.example.com/resource?page=1',
                'last' => 'https://api.example.com/resource?page=10',
                'prev' => null,
                'next' => 'https://api.example.com/resource?page=2',
            ],
            'meta' => [
                'current_page' => 1,
                'from' => 1,
                'last_page' => 10,
                'per_page' => 25,
                'to' => 25,
                'total' => 250,
            ],
        ];
    }

    /**
     * Get example error response.
     */
    public static function errorResponse(int $status, string $message, ?array $errors = null): array
    {
        $response = ['message' => $message];

        if ($errors !== null) {
            $response['errors'] = $errors;
        }

        return $response;
    }

    /**
     * Get example validation error response.
     */
    public static function validationErrorResponse(): array
    {
        return [
            'message' => 'The given data was invalid.',
            'errors' => [
                'email' => [
                    'The email field is required.',
                ],
                'name' => [
                    'The name field must be at least 2 characters.',
                ],
            ],
        ];
    }

    /**
     * Get example rate limit headers.
     */
    public static function rateLimitHeaders(int $limit = 1000, int $remaining = 999): array
    {
        return [
            'X-RateLimit-Limit' => (string) $limit,
            'X-RateLimit-Remaining' => (string) $remaining,
            'X-RateLimit-Reset' => (string) (time() + 60),
        ];
    }

    /**
     * Get example authentication headers.
     */
    public static function authHeaders(string $type = 'api_key'): array
    {
        return match ($type) {
            'api_key' => [
                'X-API-Key' => 'hk_1234567890abcdefghijklmnop',
            ],
            'bearer' => [
                'Authorization' => 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
            ],
            default => [],
        };
    }

    /**
     * Get example workspace header.
     */
    public static function workspaceHeader(): array
    {
        return [
            'X-Workspace-ID' => '550e8400-e29b-41d4-a716-446655440000',
        ];
    }

    /**
     * Get example CURL request.
     */
    public static function curlExample(
        string $method,
        string $endpoint,
        ?array $body = null,
        array $headers = []
    ): string {
        $curl = "curl -X {$method} \\\n";
        $curl .= "  'https://api.example.com{$endpoint}' \\\n";

        foreach ($headers as $name => $value) {
            $curl .= "  -H '{$name}: {$value}' \\\n";
        }

        if ($body !== null) {
            $curl .= "  -H 'Content-Type: application/json' \\\n";
            $curl .= "  -d '".json_encode($body, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES)."'";
        }

        return rtrim($curl, " \\\n");
    }

    /**
     * Get example JavaScript fetch request.
     */
    public static function fetchExample(
        string $method,
        string $endpoint,
        ?array $body = null,
        array $headers = []
    ): string {
        $allHeaders = array_merge([
            'Content-Type' => 'application/json',
        ], $headers);

        $options = [
            'method' => strtoupper($method),
            'headers' => $allHeaders,
        ];

        if ($body !== null) {
            $options['body'] = 'JSON.stringify('.json_encode($body, JSON_PRETTY_PRINT).')';
        }

        $code = "const response = await fetch('https://api.example.com{$endpoint}', {\n";
        $code .= "  method: '{$options['method']}',\n";
        $code .= '  headers: '.json_encode($allHeaders, JSON_PRETTY_PRINT).",\n";

        if ($body !== null) {
            $code .= '  body: JSON.stringify('.json_encode($body, JSON_PRETTY_PRINT)."),\n";
        }

        $code .= "});\n\n";
        $code .= 'const data = await response.json();';

        return $code;
    }

    /**
     * Get example PHP request.
     */
    public static function phpExample(
        string $method,
        string $endpoint,
        ?array $body = null,
        array $headers = []
    ): string {
        $code = "<?php\n\n";
        $code .= "\$client = new \\GuzzleHttp\\Client();\n\n";
        $code .= "\$response = \$client->request('{$method}', 'https://api.example.com{$endpoint}', [\n";

        if (! empty($headers)) {
            $code .= "    'headers' => [\n";
            foreach ($headers as $name => $value) {
                $code .= "        '{$name}' => '{$value}',\n";
            }
            $code .= "    ],\n";
        }

        if ($body !== null) {
            $code .= "    'json' => ".var_export($body, true).",\n";
        }

        $code .= "]);\n\n";
        $code .= '$data = json_decode($response->getBody(), true);';

        return $code;
    }
}
