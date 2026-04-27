<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Extensions;

use Core\Api\Documentation\Extension;
use Illuminate\Routing\Route;

/**
 * API Key Authentication Extension.
 *
 * Enhances API key authentication documentation with examples
 * and detailed instructions.
 */
class ApiKeyAuthExtension implements Extension
{
    /**
     * Extend the complete OpenAPI specification.
     */
    public function extend(array $spec, array $config): array
    {
        $apiKeyConfig = $config['auth']['api_key'] ?? [];

        if (! ($apiKeyConfig['enabled'] ?? true)) {
            return $spec;
        }

        // Enhance API key security scheme description
        if (isset($spec['components']['securitySchemes']['apiKeyAuth'])) {
            $spec['components']['securitySchemes']['apiKeyAuth']['description'] = $this->buildApiKeyDescription($apiKeyConfig);
        }

        // Add authentication guide to info.description
        $authGuide = $this->buildAuthenticationGuide($config);
        if (! empty($authGuide)) {
            $spec['info']['description'] = ($spec['info']['description'] ?? '')."\n\n".$authGuide;
        }

        // Add example schemas for authentication-related responses
        $spec['components']['schemas']['UnauthorizedError'] = [
            'type' => 'object',
            'properties' => [
                'message' => [
                    'type' => 'string',
                    'example' => 'Unauthenticated.',
                ],
            ],
        ];

        $spec['components']['schemas']['ForbiddenError'] = [
            'type' => 'object',
            'properties' => [
                'message' => [
                    'type' => 'string',
                    'example' => 'This action is unauthorised.',
                ],
            ],
        ];

        // Add common auth error responses to components
        $spec['components']['responses']['Unauthorized'] = [
            'description' => 'Authentication required or invalid credentials',
            'content' => [
                'application/json' => [
                    'schema' => [
                        '$ref' => '#/components/schemas/UnauthorizedError',
                    ],
                    'examples' => [
                        'missing_key' => [
                            'summary' => 'Missing API Key',
                            'value' => ['message' => 'API key is required.'],
                        ],
                        'invalid_key' => [
                            'summary' => 'Invalid API Key',
                            'value' => ['message' => 'Invalid API key.'],
                        ],
                        'expired_key' => [
                            'summary' => 'Expired API Key',
                            'value' => ['message' => 'API key has expired.'],
                        ],
                    ],
                ],
            ],
        ];

        $spec['components']['responses']['Forbidden'] = [
            'description' => 'Insufficient permissions for this action',
            'content' => [
                'application/json' => [
                    'schema' => [
                        '$ref' => '#/components/schemas/ForbiddenError',
                    ],
                    'examples' => [
                        'insufficient_scope' => [
                            'summary' => 'Missing Required Scope',
                            'value' => ['message' => 'API key lacks required scope: write'],
                        ],
                        'workspace_access' => [
                            'summary' => 'Workspace Access Denied',
                            'value' => ['message' => 'API key does not have access to this workspace.'],
                        ],
                    ],
                ],
            ],
        ];

        return $spec;
    }

    /**
     * Extend an individual operation.
     */
    public function extendOperation(array $operation, Route $route, string $method, array $config): array
    {
        // Add 401/403 responses to authenticated endpoints
        if (! empty($operation['security'])) {
            $hasApiKeyAuth = false;
            foreach ($operation['security'] as $security) {
                if (isset($security['apiKeyAuth'])) {
                    $hasApiKeyAuth = true;
                    break;
                }
            }

            if ($hasApiKeyAuth) {
                // Add 401 response if not present
                if (! isset($operation['responses']['401'])) {
                    $operation['responses']['401'] = [
                        '$ref' => '#/components/responses/Unauthorized',
                    ];
                }

                // Add 403 response if not present
                if (! isset($operation['responses']['403'])) {
                    $operation['responses']['403'] = [
                        '$ref' => '#/components/responses/Forbidden',
                    ];
                }
            }
        }

        return $operation;
    }

    /**
     * Build detailed API key description.
     */
    protected function buildApiKeyDescription(array $config): string
    {
        $headerName = $config['name'] ?? 'X-API-Key';
        $baseDescription = $config['description'] ?? 'API key for authentication.';

        return <<<MARKDOWN
$baseDescription

## Usage

Include your API key in the `$headerName` header:

```
$headerName: your_api_key_here
```

## Key Format

API keys follow the format: `prefix_xxxxxxxxxxxxxxxx`

- The visible prefix is a short random identifier used for lookup
- Keys are 32+ characters long
- Keys should be kept secret and never committed to version control

## Scopes

API keys can be created with specific scopes:

- `read` - Read access to resources
- `write` - Create and update resources
- `delete` - Delete resources

## Key Management

- Create and manage API keys in your workspace settings
- Keys can be revoked at any time
- Set expiration dates for temporary access
- Monitor usage via the API dashboard
MARKDOWN;
    }

    /**
     * Build authentication guide for API description.
     */
    protected function buildAuthenticationGuide(array $config): string
    {
        $apiKeyConfig = $config['auth']['api_key'] ?? [];
        $bearerConfig = $config['auth']['bearer'] ?? [];

        $sections = [];

        $sections[] = '## Authentication';
        $sections[] = '';
        $sections[] = 'This API supports multiple authentication methods:';
        $sections[] = '';

        if ($apiKeyConfig['enabled'] ?? true) {
            $headerName = $apiKeyConfig['name'] ?? 'X-API-Key';
            $sections[] = '### API Key Authentication';
            $sections[] = '';
            $sections[] = "For server-to-server integration, use API key authentication via the `$headerName` header.";
            $sections[] = '';
            $sections[] = '```http';
            $sections[] = 'GET /api/endpoint HTTP/1.1';
            $sections[] = 'Host: api.example.com';
            $sections[] = "$headerName: prefix_your_api_key_here";
            $sections[] = '```';
            $sections[] = '';
        }

        if ($bearerConfig['enabled'] ?? true) {
            $sections[] = '### Bearer Token Authentication';
            $sections[] = '';
            $sections[] = 'For user-authenticated requests (SPAs, mobile apps), use bearer token authentication.';
            $sections[] = '';
            $sections[] = '```http';
            $sections[] = 'GET /api/endpoint HTTP/1.1';
            $sections[] = 'Host: api.example.com';
            $sections[] = 'Authorization: Bearer your_token_here';
            $sections[] = '```';
            $sections[] = '';
        }

        return implode("\n", $sections);
    }
}
