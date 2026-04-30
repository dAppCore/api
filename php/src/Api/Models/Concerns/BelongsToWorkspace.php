<?php

declare(strict_types=1);

namespace Core\Api\Models\Concerns;

use Core\Tenant\Models\Workspace;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * Shared workspace scoping for package-owned API models.
 */
trait BelongsToWorkspace
{
    /**
     * Limit the query to one workspace.
     *
     * Example:
     * Biolink::query()->forWorkspace($workspace->id)->latest()->get()
     */
    public function scopeForWorkspace(Builder $query, int $workspaceId): Builder
    {
        return $query->where($this->getTable().'.workspace_id', $workspaceId);
    }

    /**
     * Workspace that owns the resource.
     */
    public function workspace(): BelongsTo
    {
        return $this->belongsTo(Workspace::class);
    }
}
