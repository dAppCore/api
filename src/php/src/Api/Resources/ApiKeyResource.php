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
 * @property \Carbon\Carbon|null $last_used_at
 * @property \Carbon\Carbon|null $expires_at
 * @property \Carbon\Carbon $created_at
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
            'name' => $this->name,
            'prefix' => $this->prefix,
            'scopes' => $this->scopes,
            'last_used_at' => $this->last_used_at?->toIso8601String(),
            'expires_at' => $this->expires_at?->toIso8601String(),
            'created_at' => $this->created_at->toIso8601String(),

            // Only included on creation
            'key' => $this->when($this->plainKey !== null, $this->plainKey),

            // Masked display key
            'display_key' => $this->masked_key,
        ];
    }
}
