<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Tenant\Concerns\BelongsToNamespace;
use Core\Tenant\Models\Workspace;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Str;
use Core\Api\Enums\BuiltinTemplateType;
use Core\Api\Enums\WebhookTemplateFormat;
use Spatie\Activitylog\LogOptions;
use Spatie\Activitylog\Traits\LogsActivity;

/**
 * Reusable webhook payload template.
 *
 * Allows users to define custom templates for webhook payloads,
 * supporting variable substitution, conditionals, and loops.
 *
 * @property int $id
 * @property string $uuid
 * @property int $workspace_id
 * @property int|null $namespace_id
 * @property string $name
 * @property string|null $description
 * @property WebhookTemplateFormat $format
 * @property string $template
 * @property string|null $example_output
 * @property bool $is_default
 * @property int $sort_order
 * @property bool $is_active
 * @property BuiltinTemplateType|null $builtin_type
 * @property \Carbon\Carbon $created_at
 * @property \Carbon\Carbon $updated_at
 */
class WebhookPayloadTemplate extends Model
{
    use BelongsToNamespace;
    use HasFactory;
    use LogsActivity;

    protected $table = 'api_webhook_payload_templates';

    protected $fillable = [
        'uuid',
        'workspace_id',
        'namespace_id',
        'name',
        'description',
        'format',
        'template',
        'example_output',
        'is_default',
        'sort_order',
        'is_active',
        'builtin_type',
    ];

    protected $casts = [
        'format' => WebhookTemplateFormat::class,
        'is_default' => 'boolean',
        'sort_order' => 'integer',
        'is_active' => 'boolean',
        'builtin_type' => BuiltinTemplateType::class,
    ];

    protected static function boot(): void
    {
        parent::boot();

        static::creating(function (WebhookPayloadTemplate $template) {
            if (empty($template->uuid)) {
                $template->uuid = (string) Str::uuid();
            }
        });

        // Ensure only one default template per workspace
        static::saving(function (WebhookPayloadTemplate $template) {
            if ($template->is_default) {
                static::where('workspace_id', $template->workspace_id)
                    ->where('id', '!=', $template->id ?? 0)
                    ->where('is_default', true)
                    ->update(['is_default' => false]);
            }
        });
    }

    // -------------------------------------------------------------------------
    // Relationships
    // -------------------------------------------------------------------------

    public function workspace(): BelongsTo
    {
        return $this->belongsTo(Workspace::class);
    }

    // -------------------------------------------------------------------------
    // Scopes
    // -------------------------------------------------------------------------

    public function scopeActive(Builder $query): Builder
    {
        return $query->where('is_active', true);
    }

    public function scopeDefault(Builder $query): Builder
    {
        return $query->where('is_default', true);
    }

    public function scopeBuiltin(Builder $query): Builder
    {
        return $query->whereNotNull('builtin_type');
    }

    public function scopeCustom(Builder $query): Builder
    {
        return $query->whereNull('builtin_type');
    }

    public function scopeOrdered(Builder $query): Builder
    {
        return $query->orderBy('sort_order')->orderBy('name');
    }

    public function scopeForWorkspace(Builder $query, int $workspaceId): Builder
    {
        return $query->where('workspace_id', $workspaceId);
    }

    // -------------------------------------------------------------------------
    // State Checks
    // -------------------------------------------------------------------------

    public function isActive(): bool
    {
        return $this->is_active === true;
    }

    public function isDefault(): bool
    {
        return $this->is_default === true;
    }

    public function isBuiltin(): bool
    {
        return $this->builtin_type !== null;
    }

    public function isCustom(): bool
    {
        return $this->builtin_type === null;
    }

    // -------------------------------------------------------------------------
    // Template Methods
    // -------------------------------------------------------------------------

    /**
     * Get the template format enum.
     */
    public function getFormat(): WebhookTemplateFormat
    {
        return $this->format ?? WebhookTemplateFormat::SIMPLE;
    }

    /**
     * Get the builtin type if this is a builtin template.
     */
    public function getBuiltinType(): ?BuiltinTemplateType
    {
        return $this->builtin_type;
    }

    /**
     * Update the example output preview.
     */
    public function updateExampleOutput(string $output): void
    {
        $this->update(['example_output' => $output]);
    }

    /**
     * Set this template as the workspace default.
     */
    public function setAsDefault(): void
    {
        $this->update(['is_default' => true]);
    }

    /**
     * Duplicate this template with a new name.
     */
    public function duplicate(?string $newName = null): static
    {
        $duplicate = $this->replicate(['uuid', 'is_default']);
        $duplicate->uuid = (string) Str::uuid();
        $duplicate->name = $newName ?? $this->name.' (copy)';
        $duplicate->is_default = false;
        $duplicate->builtin_type = null; // Custom copy
        $duplicate->save();

        return $duplicate;
    }

    // -------------------------------------------------------------------------
    // Utilities
    // -------------------------------------------------------------------------

    public function getRouteKeyName(): string
    {
        return 'uuid';
    }

    public function getActivitylogOptions(): LogOptions
    {
        return LogOptions::defaults()
            ->logOnly(['name', 'description', 'format', 'template', 'is_default', 'is_active'])
            ->logOnlyDirty()
            ->dontSubmitEmptyLogs()
            ->setDescriptionForEvent(fn (string $eventName) => "Webhook template {$eventName}");
    }

    /**
     * Get Flux badge colour for status.
     */
    public function getStatusColorAttribute(): string
    {
        if (! $this->is_active) {
            return 'zinc';
        }

        if ($this->is_default) {
            return 'green';
        }

        return 'blue';
    }

    /**
     * Get status label.
     */
    public function getStatusLabelAttribute(): string
    {
        if (! $this->is_active) {
            return 'Inactive';
        }

        if ($this->is_default) {
            return 'Default';
        }

        return 'Active';
    }

    /**
     * Get icon for template type.
     */
    public function getTypeIconAttribute(): string
    {
        if ($this->isBuiltin()) {
            return match ($this->builtin_type) {
                BuiltinTemplateType::SLACK => 'slack',
                BuiltinTemplateType::DISCORD => 'discord',
                BuiltinTemplateType::FULL => 'code-bracket',
                BuiltinTemplateType::MINIMAL => 'minus',
                default => 'document-text',
            };
        }

        return 'document';
    }

    // -------------------------------------------------------------------------
    // Factory Methods
    // -------------------------------------------------------------------------

    /**
     * Create all builtin templates for a workspace.
     */
    public static function createBuiltinTemplates(int $workspaceId, ?int $namespaceId = null): void
    {
        $sortOrder = 0;

        foreach (BuiltinTemplateType::cases() as $type) {
            static::firstOrCreate(
                [
                    'workspace_id' => $workspaceId,
                    'builtin_type' => $type,
                ],
                [
                    'uuid' => (string) Str::uuid(),
                    'namespace_id' => $namespaceId,
                    'name' => $type->label(),
                    'description' => $type->description(),
                    'format' => $type->format(),
                    'template' => $type->template(),
                    'is_default' => $type === BuiltinTemplateType::FULL,
                    'sort_order' => $sortOrder++,
                    'is_active' => true,
                ]
            );
        }
    }

    /**
     * Get or create the default template for a workspace.
     */
    public static function getDefaultForWorkspace(int $workspaceId): ?static
    {
        return static::forWorkspace($workspaceId)
            ->active()
            ->default()
            ->first();
    }
}
