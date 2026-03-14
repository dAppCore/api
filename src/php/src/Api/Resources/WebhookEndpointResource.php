<?php

declare(strict_types=1);

namespace Core\Api\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\JsonResource;

/**
 * Webhook endpoint resource for API responses.
 *
 * @property int $id
 * @property string $url
 * @property array $events
 * @property bool $active
 * @property string|null $description
 * @property \Carbon\Carbon|null $last_triggered_at
 * @property int $failure_count
 * @property \Carbon\Carbon|null $disabled_at
 * @property \Carbon\Carbon $created_at
 * @property \Carbon\Carbon $updated_at
 * @property string $secret
 */
class WebhookEndpointResource extends JsonResource
{
    /**
     * Include secret in response (only on creation).
     */
    public bool $includeSecret = false;

    /**
     * Create resource with secret visible.
     */
    public static function withSecret($resource): static
    {
        $instance = new static($resource);
        $instance->includeSecret = true;

        return $instance;
    }

    public function toArray(Request $request): array
    {
        return [
            'id' => $this->id,
            'url' => $this->url,
            'events' => $this->events,
            'active' => $this->active,
            'description' => $this->description,
            'last_triggered_at' => $this->last_triggered_at?->toIso8601String(),
            'failure_count' => $this->failure_count,
            'disabled_at' => $this->disabled_at?->toIso8601String(),
            'created_at' => $this->created_at->toIso8601String(),
            'updated_at' => $this->updated_at->toIso8601String(),

            // Only on creation
            'secret' => $this->when($this->includeSecret, $this->secret),

            // Links
            'links' => [
                'self' => route('api.v1.webhooks.show', $this->id, false),
                'deliveries' => route('api.v1.webhooks.deliveries', $this->id, false),
            ],
        ];
    }
}
