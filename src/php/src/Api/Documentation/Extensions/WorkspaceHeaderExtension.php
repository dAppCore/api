<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Extensions;

use Core\Api\Documentation\Extension;
use Illuminate\Routing\Route;

/**
 * Workspace Header Extension.
 *
 * Adds documentation for the X-Workspace-ID header used in multi-tenant
 * API operations.
 */
class WorkspaceHeaderExtension implements Extension
{
    /**
     * Extend the complete OpenAPI specification.
     */
    public function extend(array $spec, array $config): array
    {
        // Add workspace header parameter to components
        $workspaceConfig = $config['workspace'] ?? [];

        if (! empty($workspaceConfig)) {
            $spec['components']['parameters']['workspaceId'] = [
                'name' => $workspaceConfig['header_name'] ?? 'X-Workspace-ID',
                'in' => 'header',
                'required' => $workspaceConfig['required'] ?? false,
                'description' => $workspaceConfig['description'] ?? 'Workspace identifier for multi-tenant operations',
                'schema' => [
                    'type' => 'string',
                    'format' => 'uuid',
                    'example' => '550e8400-e29b-41d4-a716-446655440000',
                ],
            ];
        }

        return $spec;
    }

    /**
     * Extend an individual operation.
     */
    public function extendOperation(array $operation, Route $route, string $method, array $config): array
    {
        // Check if route requires workspace context
        if (! $this->requiresWorkspace($route)) {
            return $operation;
        }

        $workspaceConfig = $config['workspace'] ?? [];
        $headerName = $workspaceConfig['header_name'] ?? 'X-Workspace-ID';

        // Add workspace header parameter reference
        $operation['parameters'] = $operation['parameters'] ?? [];

        // Check if already added
        foreach ($operation['parameters'] as $param) {
            if (isset($param['name']) && $param['name'] === $headerName) {
                return $operation;
            }
        }

        // Add as reference to component
        $operation['parameters'][] = [
            '$ref' => '#/components/parameters/workspaceId',
        ];

        return $operation;
    }

    /**
     * Check if route requires workspace context.
     */
    protected function requiresWorkspace(Route $route): bool
    {
        $middleware = $route->middleware();

        // Check for workspace-related middleware
        foreach ($middleware as $m) {
            if (str_contains($m, 'workspace') ||
                str_contains($m, 'api.auth') ||
                str_contains($m, 'auth.api')) {
                return true;
            }
        }

        // Check route name patterns that typically need workspace
        $name = $route->getName() ?? '';
        $workspaceRoutes = [
            'api.key.',
            'api.bio.',
            'api.blocks.',
            'api.shortlinks.',
            'api.qr.',
            'api.workspaces.',
            'api.webhooks.',
            'api.content.',
        ];

        foreach ($workspaceRoutes as $pattern) {
            if (str_starts_with($name, $pattern)) {
                return true;
            }
        }

        return false;
    }
}
