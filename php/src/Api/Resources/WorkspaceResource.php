<?php

declare(strict_types=1);

namespace Core\Api\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

/**
 * Workspace API resource.
 *
 * Transforms Workspace models into API responses.
 *
 * @mixin \Core\Tenant\Models\Workspace
 */
class WorkspaceResource extends JsonResource
{
    /**
     * Transform the resource into an array.
     *
     * @return array<string, mixed>
     */
    public function toArray(Request $request): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'slug' => $this->slug,
            'icon' => $this->icon,
            'color' => $this->color,
            'description' => $this->description,
            'type' => $this->type,
            'is_active' => $this->is_active,

            // Stats
            'users_count' => $this->whenCounted('users'),
            'bio_pages_count' => $this->whenCounted('bioPages'),

            // Role (when available via pivot)
            'role' => $this->whenPivotLoaded('user_workspace', fn () => $this->pivot->role),
            'is_default' => $this->whenPivotLoaded('user_workspace', fn () => $this->pivot->is_default),

            // Settings (public only)
            'settings' => $this->when($this->settings, fn () => $this->getPublicSettings()),

            // Timestamps
            'created_at' => $this->created_at?->toIso8601String(),
            'updated_at' => $this->updated_at?->toIso8601String(),
        ];
    }

    /**
     * Get public settings (filter sensitive data).
     */
    protected function getPublicSettings(): array
    {
        $settings = $this->settings ?? [];

        // Remove sensitive keys
        unset(
            $settings['wp_connector_secret'],
            $settings['api_secrets']
        );

        return $settings;
    }
}
