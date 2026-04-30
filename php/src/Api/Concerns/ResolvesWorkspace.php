<?php

declare(strict_types=1);

namespace Core\Api\Concerns;

use Illuminate\Http\Request;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;

/**
 * Resolve workspace from request context.
 *
 * Supports both API key authentication (workspace from key) and
 * session authentication (workspace from user default).
 */
trait ResolvesWorkspace
{
    /**
     * Get the workspace from request context.
     *
     * Priority:
     * 1. API key workspace (set by AuthenticateApiKey middleware)
     * 2. Explicit workspace_id parameter
     * 3. User's default workspace
     */
    protected function resolveWorkspace(Request $request): ?Workspace
    {
        // API key auth provides workspace directly
        $workspace = $request->attributes->get('workspace');
        if ($workspace instanceof Workspace) {
            return $workspace;
        }

        // Check for explicit workspace_id
        $workspaceId = $request->attributes->get('workspace_id')
            ?? $request->input('workspace_id')
            ?? $request->header('X-Workspace-Id');

        if ($workspaceId) {
            return $this->findWorkspaceForUser($request, (int) $workspaceId);
        }

        $workspaceParam = $request->route('workspace');
        if ($workspaceParam instanceof Workspace) {
            return $workspaceParam;
        }
        if ($workspaceParam !== null) {
            if (is_scalar($workspaceParam) && $workspaceParam !== '' && ctype_digit((string) $workspaceParam)) {
                return $this->findWorkspaceForUser($request, (int) $workspaceParam);
            }

            // Fail closed: an explicit but invalid workspace route value must
            // NOT silently fall back to the user's default. That could route
            // a workspace-scoped request onto the wrong tenant.
            return null;
        }

        // Fall back to user's default workspace
        $user = $request->user();
        if ($user instanceof User) {
            return $user->defaultHostWorkspace();
        }

        return null;
    }

    /**
     * Find a workspace by ID that the user has access to.
     */
    protected function findWorkspaceForUser(Request $request, int $workspaceId): ?Workspace
    {
        $user = $request->user();

        if (! $user instanceof User) {
            return null;
        }

        return $user->workspaces()
            ->where('workspaces.id', $workspaceId)
            ->first();
    }

    /**
     * Get the authentication type.
     */
    protected function getAuthType(Request $request): string
    {
        return $request->attributes->get('auth_type', 'session');
    }

    /**
     * Check if authenticated via API key.
     */
    protected function isApiKeyAuth(Request $request): bool
    {
        return $this->getAuthType($request) === 'api_key';
    }
}
