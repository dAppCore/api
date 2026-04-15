<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api\Concerns;

use Illuminate\Database\Eloquent\Model;

/**
 * Stable API payloads for workspace-owned resources.
 */
trait SerialisesWorkspaceResource
{
    /**
     * Convert a model into a predictable array payload.
     */
    protected function modelPayload(Model $model): array
    {
        $attributes = $model->attributesToArray();
        $attributes['id'] = $model->getKey();
        $attributes['workspace_id'] = $attributes['workspace_id'] ?? $model->getAttribute('workspace_id');
        $attributes['created_at'] = $model->getAttribute('created_at')?->toIso8601String();
        $attributes['updated_at'] = $model->getAttribute('updated_at')?->toIso8601String();

        return $attributes;
    }
}
