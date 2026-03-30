<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Illuminate\Support\Str;
use Core\Api\Enums\WebhookTemplateFormat;
use Core\Api\Models\WebhookPayloadTemplate;
use Core\Api\Services\WebhookTemplateService;

/**
 * API controller for managing webhook payload templates.
 */
class WebhookTemplateController extends Controller
{
    use HasApiResponses;

    public function __construct(
        protected WebhookTemplateService $templateService
    ) {}

    /**
     * List all templates for the workspace.
     */
    public function index(Request $request): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $query = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->active()
            ->ordered();

        // Optional filtering
        if ($request->has('builtin')) {
            $request->boolean('builtin')
                ? $query->builtin()
                : $query->custom();
        }

        $templates = $query->get()->map(fn ($template) => $this->formatTemplate($template));

        return response()->json([
            'data' => $templates,
            'meta' => [
                'total' => $templates->count(),
            ],
        ]);
    }

    /**
     * Get a single template by UUID.
     */
    public function show(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $template) {
            return $this->notFoundResponse('Template');
        }

        return response()->json([
            'data' => $this->formatTemplate($template, true),
        ]);
    }

    /**
     * Create a new template.
     */
    public function store(Request $request): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $validated = $request->validate([
            'name' => 'required|string|max:255',
            'description' => 'nullable|string|max:1000',
            'format' => 'required|in:simple,mustache,json',
            'template' => 'required|string|max:65535',
            'is_default' => 'boolean',
            'is_active' => 'boolean',
        ]);

        // Validate template syntax
        $format = WebhookTemplateFormat::from($validated['format']);
        $validation = $this->templateService->validateTemplate($validated['template'], $format);

        if (! $validation['valid']) {
            return $this->validationErrorResponse([
                'template' => $validation['errors'],
            ]);
        }

        $template = WebhookPayloadTemplate::create([
            'uuid' => Str::uuid()->toString(),
            'workspace_id' => $workspace->id,
            'namespace_id' => $workspace->default_namespace_id ?? null,
            'name' => $validated['name'],
            'description' => $validated['description'] ?? null,
            'format' => $validated['format'],
            'template' => $validated['template'],
            'is_default' => $validated['is_default'] ?? false,
            'is_active' => $validated['is_active'] ?? true,
        ]);

        return response()->json([
            'data' => $this->formatTemplate($template, true),
        ], 201);
    }

    /**
     * Update an existing template.
     */
    public function update(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $template) {
            return $this->notFoundResponse('Template');
        }

        $validated = $request->validate([
            'name' => 'sometimes|string|max:255',
            'description' => 'nullable|string|max:1000',
            'format' => 'sometimes|in:simple,mustache,json',
            'template' => 'sometimes|string|max:65535',
            'is_default' => 'boolean',
            'is_active' => 'boolean',
        ]);

        // Validate template syntax if template is being updated
        if (isset($validated['template'])) {
            $format = WebhookTemplateFormat::from($validated['format'] ?? $template->format->value);
            $validation = $this->templateService->validateTemplate($validated['template'], $format);

            if (! $validation['valid']) {
                return $this->validationErrorResponse([
                    'template' => $validation['errors'],
                ]);
            }
        }

        // Don't allow modifying builtin templates' format
        if ($template->isBuiltin()) {
            unset($validated['format']);
        }

        $template->update($validated);

        return response()->json([
            'data' => $this->formatTemplate($template->fresh(), true),
        ]);
    }

    /**
     * Delete a template.
     */
    public function destroy(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $template) {
            return $this->notFoundResponse('Template');
        }

        // Don't allow deleting builtin templates
        if ($template->isBuiltin()) {
            return $this->forbiddenResponse('Built-in templates cannot be deleted');
        }

        $template->delete();

        return response()->json(null, 204);
    }

    /**
     * Validate a template without saving.
     */
    public function validate(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'format' => 'required|in:simple,mustache,json',
            'template' => 'required|string|max:65535',
        ]);

        $format = WebhookTemplateFormat::from($validated['format']);
        $validation = $this->templateService->validateTemplate($validated['template'], $format);

        return response()->json([
            'valid' => $validation['valid'],
            'errors' => $validation['errors'],
        ]);
    }

    /**
     * Preview a template with sample data.
     */
    public function preview(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'format' => 'required|in:simple,mustache,json',
            'template' => 'required|string|max:65535',
            'event_type' => 'nullable|string|max:100',
        ]);

        $format = WebhookTemplateFormat::from($validated['format']);
        $result = $this->templateService->previewPayload(
            $validated['template'],
            $format,
            $validated['event_type'] ?? null
        );

        return response()->json($result);
    }

    /**
     * Duplicate an existing template.
     */
    public function duplicate(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $template) {
            return $this->notFoundResponse('Template');
        }

        $newName = $request->input('name', $template->name.' (copy)');
        $duplicate = $template->duplicate($newName);

        return response()->json([
            'data' => $this->formatTemplate($duplicate, true),
        ], 201);
    }

    /**
     * Set a template as the workspace default.
     */
    public function setDefault(Request $request, string $uuid): JsonResponse
    {
        $workspace = $request->user()?->defaultHostWorkspace();

        if (! $workspace) {
            return $this->noWorkspaceResponse();
        }

        $template = WebhookPayloadTemplate::where('workspace_id', $workspace->id)
            ->where('uuid', $uuid)
            ->first();

        if (! $template) {
            return $this->notFoundResponse('Template');
        }

        $template->setAsDefault();

        return response()->json([
            'data' => $this->formatTemplate($template->fresh(), true),
        ]);
    }

    /**
     * Get available template variables.
     */
    public function variables(Request $request): JsonResponse
    {
        $eventType = $request->input('event_type');
        $variables = $this->templateService->getAvailableVariables($eventType);

        return response()->json([
            'data' => $variables,
        ]);
    }

    /**
     * Get available template filters.
     */
    public function filters(): JsonResponse
    {
        $filters = $this->templateService->getAvailableFilters();

        return response()->json([
            'data' => $filters,
        ]);
    }

    /**
     * Get builtin template definitions.
     */
    public function builtins(): JsonResponse
    {
        $templates = $this->templateService->getBuiltinTemplates();

        return response()->json([
            'data' => $templates,
        ]);
    }

    // -------------------------------------------------------------------------
    // Protected Methods
    // -------------------------------------------------------------------------

    /**
     * Format a template for API response.
     */
    protected function formatTemplate(WebhookPayloadTemplate $template, bool $includeContent = false): array
    {
        $data = [
            'uuid' => $template->uuid,
            'name' => $template->name,
            'description' => $template->description,
            'format' => $template->format->value,
            'is_default' => $template->is_default,
            'is_active' => $template->is_active,
            'is_builtin' => $template->isBuiltin(),
            'builtin_type' => $template->builtin_type?->value,
            'created_at' => $template->created_at?->toIso8601String(),
            'updated_at' => $template->updated_at?->toIso8601String(),
        ];

        if ($includeContent) {
            $data['template'] = $template->template;
            $data['example_output'] = $template->example_output;
        }

        return $data;
    }
}
