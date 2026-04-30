<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Front\Controller;
use Core\Tenant\Models\User;
use Core\Tenant\Models\WorkspaceMember;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Workspace member management endpoints.
 */
class WorkspaceMemberController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;

    /**
     * List members for the current workspace.
     *
     * GET /api/v1/workspaces/{workspace}/members
     */
    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $members = WorkspaceMember::query()
            ->forWorkspace($workspace)
            ->with(['user', 'team', 'inviter'])
            ->orderBy('role')
            ->orderBy('created_at')
            ->get();

        return response()->json([
            'workspace_id' => $workspace->id,
            'members' => $members->map(fn (WorkspaceMember $member) => $this->serialize($member))->values()->all(),
            'count' => $members->count(),
        ]);
    }

    /**
     * Invite a member to the current workspace.
     *
     * POST /api/v1/workspaces/{workspace}/members
     */
    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $data = $request->validate([
            'email' => ['required', 'email', 'max:255'],
            'role' => ['sometimes', 'string', 'in:admin,member'],
            'expires_in_days' => ['sometimes', 'integer', 'min:1', 'max:30'],
        ]);

        $invitation = $workspace->invite(
            email: $data['email'],
            role: $data['role'] ?? 'member',
            invitedBy: $request->user(),
            expiresInDays: (int) ($data['expires_in_days'] ?? 7),
        );

        return response()->json([
            'invitation' => [
                'id' => $invitation->id,
                'workspace_id' => $invitation->workspace_id,
                'email' => $invitation->email,
                'role' => $invitation->role,
                'expires_at' => $invitation->expires_at?->toIso8601String(),
                'accepted_at' => $invitation->accepted_at?->toIso8601String(),
                'invited_by' => $invitation->invited_by,
            ],
        ], 201);
    }

    /**
     * Remove a member from the current workspace.
     *
     * DELETE /api/v1/workspaces/{workspace}/members/{user}
     */
    public function destroy(Request $request, string $user): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $member = WorkspaceMember::query()
            ->forWorkspace($workspace)
            ->forUser((int) $user)
            ->first();

        if (! $member instanceof WorkspaceMember) {
            return $this->notFoundResponse('Member');
        }

        if ($member->role === WorkspaceMember::ROLE_OWNER) {
            return $this->forbiddenResponse('Workspace owners cannot be removed through the member endpoint.');
        }

        $member->delete();

        return $this->successResponse('Member removed successfully.');
    }

    protected function serialize(WorkspaceMember $member): array
    {
        return [
            'id' => $member->id,
            'workspace_id' => $member->workspace_id,
            'user_id' => $member->user_id,
            'role' => $member->role,
            'team_id' => $member->team_id,
            'custom_permissions' => $member->custom_permissions,
            'is_default' => $member->is_default,
            'joined_at' => $member->joined_at?->toIso8601String(),
            'created_at' => $member->created_at?->toIso8601String(),
            'updated_at' => $member->updated_at?->toIso8601String(),
            'user' => $member->relationLoaded('user') && $member->user instanceof User
                ? [
                    'id' => $member->user->id,
                    'name' => $member->user->name,
                    'email' => $member->user->email,
                ]
                : null,
        ];
    }
}
