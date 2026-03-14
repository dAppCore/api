<?php

declare(strict_types=1);

namespace Core\Api\View\Modal\Admin;

use Illuminate\Support\Str;
use Livewire\Attributes\Computed;
use Livewire\Attributes\Layout;
use Livewire\Attributes\Url;
use Livewire\Attributes\Validate;
use Livewire\Component;
use Livewire\WithPagination;
use Core\Api\Enums\BuiltinTemplateType;
use Core\Api\Enums\WebhookTemplateFormat;
use Core\Api\Models\WebhookPayloadTemplate;
use Core\Api\Services\WebhookTemplateService;

#[Layout('hub::admin.layouts.app')]
class WebhookTemplateManager extends Component
{
    use WithPagination;

    // List view state
    #[Url]
    public string $search = '';

    #[Url]
    public string $filter = 'all'; // all, custom, builtin, active, inactive

    public ?string $deletingId = null;

    // Editor state
    public bool $showEditor = false;

    public ?string $editingId = null;

    #[Validate('required|string|max:255')]
    public string $name = '';

    #[Validate('nullable|string|max:1000')]
    public string $description = '';

    #[Validate('required|in:simple,mustache,json')]
    public string $format = 'json';

    #[Validate('required|string|max:65535')]
    public string $template = '';

    public bool $isDefault = false;

    public bool $isActive = true;

    // Preview state
    public ?array $templatePreview = null;

    public ?array $templateErrors = null;

    public string $previewEventType = 'resource.created';

    #[Computed]
    public function workspace()
    {
        return auth()->user()?->defaultHostWorkspace();
    }

    #[Computed]
    public function templates()
    {
        if (! $this->workspace) {
            return collect();
        }

        $query = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id);

        // Apply search
        if ($this->search) {
            $escapedSearch = $this->escapeLikeWildcards($this->search);
            $query->where(function ($q) use ($escapedSearch) {
                $q->where('name', 'like', "%{$escapedSearch}%")
                    ->orWhere('description', 'like', "%{$escapedSearch}%");
            });
        }

        // Apply filter
        $query = match ($this->filter) {
            'custom' => $query->custom(),
            'builtin' => $query->builtin(),
            'active' => $query->active(),
            'inactive' => $query->where('is_active', false),
            default => $query,
        };

        return $query
            ->ordered()
            ->paginate(20);
    }

    #[Computed]
    public function templateFormats(): array
    {
        return [
            'simple' => WebhookTemplateFormat::SIMPLE->label(),
            'mustache' => WebhookTemplateFormat::MUSTACHE->label(),
            'json' => WebhookTemplateFormat::JSON->label(),
        ];
    }

    #[Computed]
    public function templateFormatDescriptions(): array
    {
        return [
            'simple' => WebhookTemplateFormat::SIMPLE->description(),
            'mustache' => WebhookTemplateFormat::MUSTACHE->description(),
            'json' => WebhookTemplateFormat::JSON->description(),
        ];
    }

    #[Computed]
    public function availableVariables(): array
    {
        $service = app(WebhookTemplateService::class);

        return $service->getAvailableVariables($this->previewEventType);
    }

    #[Computed]
    public function availableFilters(): array
    {
        $service = app(WebhookTemplateService::class);

        return $service->getAvailableFilters();
    }

    #[Computed]
    public function builtinTemplates(): array
    {
        $service = app(WebhookTemplateService::class);

        return $service->getBuiltinTemplates();
    }

    public function mount(): void
    {
        // Ensure builtin templates exist for this workspace
        if ($this->workspace) {
            WebhookPayloadTemplate::createBuiltinTemplates(
                $this->workspace->id,
                $this->workspace->default_namespace_id ?? null
            );
        }
    }

    // -------------------------------------------------------------------------
    // List Actions
    // -------------------------------------------------------------------------

    public function confirmDelete(string $uuid): void
    {
        $this->deletingId = $uuid;
    }

    public function cancelDelete(): void
    {
        $this->deletingId = null;
    }

    public function delete(): void
    {
        if (! $this->deletingId || ! $this->workspace) {
            return;
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
            ->where('uuid', $this->deletingId)
            ->first();

        if ($template) {
            // Don't allow deleting builtin templates
            if ($template->isBuiltin()) {
                $this->dispatch('notify', type: 'error', message: 'Built-in templates cannot be deleted.');
                $this->deletingId = null;

                return;
            }

            $template->delete();
            $this->dispatch('notify', type: 'success', message: 'Template deleted.');
            unset($this->templates);
        }

        $this->deletingId = null;
    }

    public function toggleActive(string $uuid): void
    {
        if (! $this->workspace) {
            return;
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if ($template) {
            $template->update(['is_active' => ! $template->is_active]);
            unset($this->templates);
            $this->dispatch('notify', type: 'success', message: $template->is_active ? 'Template enabled.' : 'Template disabled.');
        }
    }

    public function setDefault(string $uuid): void
    {
        if (! $this->workspace) {
            return;
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if ($template) {
            $template->setAsDefault();
            unset($this->templates);
            $this->dispatch('notify', type: 'success', message: 'Default template updated.');
        }
    }

    public function duplicate(string $uuid): void
    {
        if (! $this->workspace) {
            return;
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if ($template) {
            $template->duplicate();
            unset($this->templates);
            $this->dispatch('notify', type: 'success', message: 'Template duplicated.');
        }
    }

    // -------------------------------------------------------------------------
    // Editor Actions
    // -------------------------------------------------------------------------

    public function create(): void
    {
        $this->resetEditor();
        $this->template = $this->getDefaultTemplateContent();
        $this->showEditor = true;
    }

    public function edit(string $uuid): void
    {
        if (! $this->workspace) {
            return;
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if ($template) {
            $this->editingId = $uuid;
            $this->name = $template->name;
            $this->description = $template->description ?? '';
            $this->format = $template->format->value;
            $this->template = $template->template;
            $this->isDefault = $template->is_default;
            $this->isActive = $template->is_active;
            $this->templatePreview = null;
            $this->templateErrors = null;
            $this->showEditor = true;
        }
    }

    public function closeEditor(): void
    {
        $this->showEditor = false;
        $this->resetEditor();
    }

    public function save(): void
    {
        // Validate template first
        $this->validateTemplate();
        if (! empty($this->templateErrors)) {
            return;
        }

        $this->validate();

        if (! $this->workspace) {
            return;
        }

        $data = [
            'name' => $this->name,
            'description' => $this->description ?: null,
            'format' => $this->format,
            'template' => $this->template,
            'is_default' => $this->isDefault,
            'is_active' => $this->isActive,
        ];

        if ($this->editingId) {
            $template = WebhookPayloadTemplate::where('workspace_id', $this->workspace->id)
                ->where('uuid', $this->editingId)
                ->first();

            if ($template) {
                // Don't allow modifying builtin templates' core properties
                if ($template->isBuiltin()) {
                    unset($data['format']);
                }

                $template->update($data);
                $this->dispatch('notify', type: 'success', message: 'Template updated.');
            }
        } else {
            $data['uuid'] = Str::uuid()->toString();
            $data['workspace_id'] = $this->workspace->id;
            $data['namespace_id'] = $this->workspace->default_namespace_id ?? null;

            WebhookPayloadTemplate::create($data);
            $this->dispatch('notify', type: 'success', message: 'Template created.');
        }

        unset($this->templates);
        $this->closeEditor();
    }

    public function validateTemplate(): void
    {
        $this->templateErrors = null;

        if (empty($this->template)) {
            $this->templateErrors = ['Template cannot be empty.'];

            return;
        }

        $service = app(WebhookTemplateService::class);
        $format = WebhookTemplateFormat::tryFrom($this->format) ?? WebhookTemplateFormat::SIMPLE;

        $result = $service->validateTemplate($this->template, $format);

        if (! $result['valid']) {
            $this->templateErrors = $result['errors'];
        }
    }

    public function previewTemplate(): void
    {
        $this->templatePreview = null;
        $this->templateErrors = null;

        if (empty($this->template)) {
            $this->templateErrors = ['Template cannot be empty.'];

            return;
        }

        $service = app(WebhookTemplateService::class);
        $format = WebhookTemplateFormat::tryFrom($this->format) ?? WebhookTemplateFormat::SIMPLE;

        $result = $service->previewPayload($this->template, $format, $this->previewEventType);

        if ($result['success']) {
            $this->templatePreview = $result['output'];
            $this->templateErrors = null;
        } else {
            $this->templatePreview = null;
            $this->templateErrors = $result['errors'];
        }
    }

    public function insertVariable(string $variable): void
    {
        $this->dispatch('insert-variable', variable: '{{'.$variable.'}}');
    }

    public function loadBuiltinTemplate(string $type): void
    {
        $builtinType = BuiltinTemplateType::tryFrom($type);
        if ($builtinType) {
            $this->template = $builtinType->template();
            $this->format = $builtinType->format()->value;
            $this->templatePreview = null;
            $this->templateErrors = null;
        }
    }

    public function resetTemplate(): void
    {
        $this->template = $this->getDefaultTemplateContent();
        $this->templatePreview = null;
        $this->templateErrors = null;
    }

    public function render()
    {
        return view('api::admin.webhook-template-manager');
    }

    // -------------------------------------------------------------------------
    // Protected Methods
    // -------------------------------------------------------------------------

    protected function resetEditor(): void
    {
        $this->editingId = null;
        $this->name = '';
        $this->description = '';
        $this->format = 'json';
        $this->template = '';
        $this->isDefault = false;
        $this->isActive = true;
        $this->templatePreview = null;
        $this->templateErrors = null;
    }

    protected function getDefaultTemplateContent(): string
    {
        return <<<'JSON'
{
    "event": "{{event.type}}",
    "timestamp": "{{timestamp}}",
    "data": {{data | json}}
}
JSON;
    }

    protected function escapeLikeWildcards(string $value): string
    {
        return str_replace(['%', '_'], ['\\%', '\\_'], $value);
    }
}
