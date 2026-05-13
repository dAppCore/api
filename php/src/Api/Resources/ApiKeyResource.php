<?php

declare(strict_types=1);

namespace Core\Api\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

/**
 * API Key resource for API responses.
 *
 * @property int $id
 * @property string $name
 * @property string $prefix
 * @property array|null $scopes
 * @property array|null $server_scopes
 * @property array|null $allowed_ips
 * @property \Carbon\Carbon|null $last_used_at
 * @property \Carbon\Carbon|null $expires_at
 * @property \Carbon\Carbon|null $grace_period_ends_at
 * @property \Carbon\Carbon $created_at
 * @property \Carbon\Carbon $updated_at
 * @property string $masked_key
 */
class ApiKeyResource extends JsonResource
{
    /**
     * The plain key to include in creation response.
     * Only set when key is first created.
     */
    public ?string $plainKey = null;

    /**
     * Create a new resource instance with plain key.
     */
    public static function withPlainKey($resource, string $plainKey): static
    {
        $instance = new static($resource);
        $instance->plainKey = $plainKey;

        return $instance;
    }

    public function toArray(Request $request): array
    {
        return [
            'id' => $this->id,
            'workspace_id' => $this->workspace_id,
            'name' => $this->name,
            'prefix' => $this->prefix,
            'scopes' => $this->scopes,
            'server_scopes' => $this->server_scopes,
            'allowed_ips' => $this->allowed_ips,
            'last_used_at' => $this->last_used_at?->toIso8601String(),
            'expires_at' => $this->expires_at?->toIso8601String(),
            'grace_period_ends_at' => $this->grace_period_ends_at?->toIso8601String(),
            'rotated_from_id' => $this->rotated_from_id,
            'created_at' => $this->created_at->toIso8601String(),
            'updated_at' => $this->updated_at->toIso8601String(),

            // Only included on creation
            'key' => $this->when($this->plainKey !== null, $this->plainKey),

            // Masked display key
            'display_key' => $this->masked_key,
        ];
    }
}
