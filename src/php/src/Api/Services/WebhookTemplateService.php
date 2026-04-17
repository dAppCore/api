<?php

declare(strict_types=1);

namespace Core\Api\Services;

use Illuminate\Support\Arr;
use Illuminate\Support\Carbon;
use Core\Api\Contracts\WebhookEvent;
use Core\Api\Enums\BuiltinTemplateType;
use Core\Api\Enums\WebhookTemplateFormat;
use Core\Api\Models\WebhookPayloadTemplate;

/**
 * Service for rendering and validating webhook payload templates.
 *
 * Supports multiple template formats:
 * - Simple: Basic {{variable}} substitution
 * - Mustache: Conditionals and loops with {{#if}}, {{#each}}
 * - JSON: Structured JSON with embedded variables
 */
class WebhookTemplateService
{
    /**
     * Available filters for template variables.
     */
    protected const FILTERS = [
        'iso8601' => 'formatIso8601',
        'timestamp' => 'formatTimestamp',
        'currency' => 'formatCurrency',
        'json' => 'formatJson',
        'upper' => 'formatUpper',
        'lower' => 'formatLower',
        'default' => 'formatDefault',
        'truncate' => 'formatTruncate',
        'escape' => 'formatEscape',
        'urlencode' => 'formatUrlencode',
    ];

    /**
     * Render a template with the given event data.
     *
     * @param  WebhookPayloadTemplate  $template  The template containing the pattern
     * @param  WebhookEvent  $event  The event providing data
     * @return array The rendered payload
     *
     * @throws \InvalidArgumentException If template is invalid
     */
    public function render(WebhookPayloadTemplate $template, WebhookEvent $event): array
    {
        $templateContent = $template->template;
        $format = $template->getFormat();
        $context = $this->buildContext($event);

        return $this->renderTemplate($templateContent, $format, $context);
    }

    /**
     * Render a template string with context data.
     *
     * @param  string  $templateContent  The template content
     * @param  WebhookTemplateFormat  $format  The template format
     * @param  array  $context  The context data
     * @return array The rendered payload
     *
     * @throws \InvalidArgumentException If template renders to invalid JSON
     */
    public function renderTemplate(string $templateContent, WebhookTemplateFormat $format, array $context): array
    {
        $rendered = match ($format) {
            WebhookTemplateFormat::SIMPLE => $this->renderSimple($templateContent, $context),
            WebhookTemplateFormat::MUSTACHE => $this->renderMustache($templateContent, $context),
            WebhookTemplateFormat::JSON => $this->renderJson($templateContent, $context),
        };

        // Parse as JSON if it's a string
        if (is_string($rendered)) {
            $decoded = json_decode($rendered, true);
            if (json_last_error() !== JSON_ERROR_NONE) {
                throw new \InvalidArgumentException('Template rendered to invalid JSON: '.json_last_error_msg());
            }

            return $decoded;
        }

        return $rendered;
    }

    /**
     * Build the default payload structure for an event.
     */
    public function buildDefaultPayload(WebhookEvent $event): array
    {
        return [
            'event' => $event::name(),
            'data' => $event->payload(),
            'timestamp' => now()->toIso8601String(),
        ];
    }

    /**
     * Validate a template for syntax errors.
     *
     * @param  string  $template  The template content
     * @param  WebhookTemplateFormat  $format  The template format
     * @return array{valid: bool, errors: array<string>}
     */
    public function validateTemplate(string $template, WebhookTemplateFormat $format): array
    {
        $errors = [];

        // Check for empty template
        if (empty(trim($template))) {
            return ['valid' => false, 'errors' => ['Template cannot be empty.']];
        }

        // Format-specific validation
        $errors = match ($format) {
            WebhookTemplateFormat::SIMPLE => $this->validateSimple($template),
            WebhookTemplateFormat::MUSTACHE => $this->validateMustache($template),
            WebhookTemplateFormat::JSON => $this->validateJson($template),
        };

        return [
            'valid' => empty($errors),
            'errors' => $errors,
        ];
    }

    /**
     * Get available variables for an event type.
     *
     * @param  string|null  $eventType  The event type (e.g., 'post.published')
     * @return array<string, array{type: string, description: string, example: mixed}>
     */
    public function getAvailableVariables(?string $eventType = null): array
    {
        // Base variables available for all events
        $variables = [
            'event.type' => [
                'type' => 'string',
                'description' => 'The event identifier',
                'example' => $eventType ?? 'resource.action',
            ],
            'event.name' => [
                'type' => 'string',
                'description' => 'Human-readable event name',
                'example' => 'Resource Updated',
            ],
            'message' => [
                'type' => 'string',
                'description' => 'Human-readable event message',
                'example' => 'A resource was updated successfully.',
            ],
            'timestamp' => [
                'type' => 'datetime',
                'description' => 'When the event occurred (ISO 8601)',
                'example' => now()->toIso8601String(),
            ],
            'timestamp_unix' => [
                'type' => 'integer',
                'description' => 'Unix timestamp of the event',
                'example' => now()->timestamp,
            ],
            'data' => [
                'type' => 'object',
                'description' => 'Event-specific data payload',
                'example' => ['id' => 1, 'name' => 'Example'],
            ],
            'data.id' => [
                'type' => 'mixed',
                'description' => 'Primary identifier of the resource',
                'example' => 123,
            ],
            'data.uuid' => [
                'type' => 'string',
                'description' => 'UUID of the resource (if available)',
                'example' => '550e8400-e29b-41d4-a716-446655440000',
            ],
        ];

        return $variables;
    }

    /**
     * Get available filters for template variables.
     *
     * @return array<string, string>
     */
    public function getAvailableFilters(): array
    {
        return [
            'iso8601' => 'Format datetime as ISO 8601',
            'timestamp' => 'Format datetime as Unix timestamp',
            'currency' => 'Format number as currency (2 decimal places)',
            'json' => 'Encode value as JSON',
            'upper' => 'Convert to uppercase',
            'lower' => 'Convert to lowercase',
            'default' => 'Provide default value if empty (e.g., {{value | default:N/A}})',
            'truncate' => 'Truncate to specified length (e.g., {{text | truncate:100}})',
            'escape' => 'HTML escape the value',
            'urlencode' => 'URL encode the value',
        ];
    }

    /**
     * Preview a template with sample data.
     *
     * @param  string  $template  The template content
     * @param  WebhookTemplateFormat  $format  The template format
     * @param  string|null  $eventType  The event type for sample data
     * @return array{success: bool, output: mixed, errors: array<string>}
     */
    public function previewPayload(string $template, WebhookTemplateFormat $format, ?string $eventType = null): array
    {
        // Validate first
        $validation = $this->validateTemplate($template, $format);
        if (! $validation['valid']) {
            return [
                'success' => false,
                'output' => null,
                'errors' => $validation['errors'],
            ];
        }

        // Build sample context
        $context = $this->buildSampleContext($eventType);

        try {
            $rendered = match ($format) {
                WebhookTemplateFormat::SIMPLE => $this->renderSimple($template, $context),
                WebhookTemplateFormat::MUSTACHE => $this->renderMustache($template, $context),
                WebhookTemplateFormat::JSON => $this->renderJson($template, $context),
            };

            // Parse as JSON
            if (is_string($rendered)) {
                $decoded = json_decode($rendered, true);
                if (json_last_error() !== JSON_ERROR_NONE) {
                    return [
                        'success' => false,
                        'output' => $rendered,
                        'errors' => ['Rendered template is not valid JSON: '.json_last_error_msg()],
                    ];
                }

                return [
                    'success' => true,
                    'output' => $decoded,
                    'errors' => [],
                ];
            }

            return [
                'success' => true,
                'output' => $rendered,
                'errors' => [],
            ];
        } catch (\Exception $e) {
            return [
                'success' => false,
                'output' => null,
                'errors' => [$e->getMessage()],
            ];
        }
    }

    /**
     * Get builtin template content by type.
     */
    public function getBuiltinTemplate(BuiltinTemplateType $type): string
    {
        return $type->template();
    }

    /**
     * Get all builtin templates.
     *
     * @return array<string, array{name: string, description: string, template: string, format: WebhookTemplateFormat}>
     */
    public function getBuiltinTemplates(): array
    {
        $templates = [];

        foreach (BuiltinTemplateType::cases() as $type) {
            $templates[$type->value] = [
                'name' => $type->label(),
                'description' => $type->description(),
                'template' => $type->template(),
                'format' => $type->format(),
            ];
        }

        return $templates;
    }

    // -------------------------------------------------------------------------
    // Protected Methods
    // -------------------------------------------------------------------------

    /**
     * Build template context from an event.
     */
    protected function buildContext(WebhookEvent $event): array
    {
        return [
            'event' => [
                'type' => $event::name(),
                'name' => $event::nameLocalised(),
            ],
            'data' => $event->payload(),
            'message' => $event->message(),
            'timestamp' => now()->toIso8601String(),
            'timestamp_unix' => now()->timestamp,
        ];
    }

    /**
     * Build sample context for preview.
     */
    protected function buildSampleContext(?string $eventType): array
    {
        $variables = $this->getAvailableVariables($eventType);
        $context = [];

        foreach ($variables as $path => $info) {
            Arr::set($context, $path, $info['example']);
        }

        // Add message
        $context['message'] = 'Sample webhook event message';

        return $context;
    }

    /**
     * Render simple template with variable substitution.
     */
    protected function renderSimple(string $template, array $context): string
    {
        // Match {{variable}} or {{variable | filter}} or {{variable | filter:arg}}
        return preg_replace_callback(
            '/\{\{\s*([a-zA-Z0-9_\.]+)(?:\s*\|\s*([a-zA-Z0-9_]+)(?::([^\}]+))?)?\s*\}\}/',
            function ($matches) use ($context) {
                $path = $matches[1];
                $filter = $matches[2] ?? null;
                $filterArg = $matches[3] ?? null;

                $value = Arr::get($context, $path);

                // Apply filter if specified
                if ($filter && isset(self::FILTERS[$filter])) {
                    $method = self::FILTERS[$filter];
                    $value = $this->$method($value, $filterArg);
                }

                // Convert arrays/objects to JSON strings
                if (is_array($value) || is_object($value)) {
                    return json_encode($value);
                }

                return (string) ($value ?? '');
            },
            $template
        );
    }

    /**
     * Render mustache-style template.
     */
    protected function renderMustache(string $template, array $context): string
    {
        // Process conditionals: {{#if variable}}...{{/if}}
        $template = preg_replace_callback(
            '/\{\{#if\s+([a-zA-Z0-9_\.]+)\s*\}\}(.*?)\{\{\/if\}\}/s',
            function ($matches) use ($context) {
                $path = $matches[1];
                $content = $matches[2];
                $value = Arr::get($context, $path);

                // Check if value is truthy
                if ($value && (! is_array($value) || ! empty($value))) {
                    return $this->renderMustache($content, $context);
                }

                return '';
            },
            $template
        );

        // Process negative conditionals: {{#unless variable}}...{{/unless}}
        $template = preg_replace_callback(
            '/\{\{#unless\s+([a-zA-Z0-9_\.]+)\s*\}\}(.*?)\{\{\/unless\}\}/s',
            function ($matches) use ($context) {
                $path = $matches[1];
                $content = $matches[2];
                $value = Arr::get($context, $path);

                // Check if value is falsy
                if (! $value || (is_array($value) && empty($value))) {
                    return $this->renderMustache($content, $context);
                }

                return '';
            },
            $template
        );

        // Process loops: {{#each variable}}...{{/each}}
        $template = preg_replace_callback(
            '/\{\{#each\s+([a-zA-Z0-9_\.]+)\s*\}\}(.*?)\{\{\/each\}\}/s',
            function ($matches) use ($context) {
                $path = $matches[1];
                $content = $matches[2];
                $items = Arr::get($context, $path, []);

                if (! is_array($items)) {
                    return '';
                }

                $output = '';
                foreach ($items as $index => $item) {
                    $itemContext = array_merge($context, [
                        'this' => $item,
                        '@index' => $index,
                        '@first' => $index === 0,
                        '@last' => $index === count($items) - 1,
                    ]);

                    // Also allow direct access to item properties
                    if (is_array($item)) {
                        $itemContext = array_merge($itemContext, $item);
                    }

                    $output .= $this->renderMustache($content, $itemContext);
                }

                return $output;
            },
            $template
        );

        // Finally, do simple variable replacement
        return $this->renderSimple($template, $context);
    }

    /**
     * Render JSON template.
     */
    protected function renderJson(string $template, array $context): string
    {
        // For JSON format, we do simple rendering then validate the result
        return $this->renderSimple($template, $context);
    }

    /**
     * Validate simple template syntax.
     */
    protected function validateSimple(string $template): array
    {
        $errors = [];

        // Check for unclosed braces
        $openCount = substr_count($template, '{{');
        $closeCount = substr_count($template, '}}');

        if ($openCount !== $closeCount) {
            $errors[] = 'Mismatched template braces. Found '.$openCount.' opening and '.$closeCount.' closing.';
        }

        // Check for invalid variable names
        preg_match_all('/\{\{\s*([^}|]+)/', $template, $matches);
        foreach ($matches[1] as $varName) {
            $varName = trim($varName);
            // Allow supported mustache block syntax so validateMustache() can reuse this check.
            if ($this->isSupportedMustacheBlockTag($varName)) {
                continue;
            }

            if (! preg_match('/^[#\/]?[a-zA-Z0-9_\.@]+$/', $varName)) {
                $errors[] = "Invalid variable name: {$varName}";
            }
        }

        // Check for unknown filters
        preg_match_all('/\|\s*([a-zA-Z0-9_]+)/', $template, $filterMatches);
        foreach ($filterMatches[1] as $filter) {
            if (! isset(self::FILTERS[$filter])) {
                $errors[] = "Unknown filter: {$filter}. Available: ".implode(', ', array_keys(self::FILTERS));
            }
        }

        return $errors;
    }

    /**
     * Determine whether a tag is a supported mustache block delimiter.
     */
    protected function isSupportedMustacheBlockTag(string $tag): bool
    {
        return (bool) preg_match('/^(?:#(?:if|unless|each)\s+[a-zA-Z0-9_\.]+|\/(?:if|unless|each))$/', $tag);
    }

    /**
     * Validate mustache template syntax.
     */
    protected function validateMustache(string $template): array
    {
        $errors = $this->validateSimple($template);

        // Check for unclosed blocks
        $blocks = ['if', 'unless', 'each'];
        foreach ($blocks as $block) {
            $openCount = preg_match_all('/\{\{#'.$block.'\s/', $template);
            $closeCount = preg_match_all('/\{\{\/'.$block.'\}\}/', $template);

            if ($openCount !== $closeCount) {
                $errors[] = "Unclosed {{#{$block}}} block. Found {$openCount} opening and {$closeCount} closing.";
            }
        }

        return $errors;
    }

    /**
     * Validate JSON template syntax.
     */
    protected function validateJson(string $template): array
    {
        $errors = $this->validateSimple($template);

        // Try to parse as JSON after replacing variables with placeholders
        $testTemplate = preg_replace('/\{\{[^}]+\}\}/', '"__placeholder__"', $template);

        json_decode($testTemplate);
        if (json_last_error() !== JSON_ERROR_NONE) {
            $errors[] = 'Template is not valid JSON structure: '.json_last_error_msg();
        }

        return $errors;
    }

    // -------------------------------------------------------------------------
    // Filter methods
    // -------------------------------------------------------------------------

    protected function formatIso8601(mixed $value, ?string $arg = null): string
    {
        if ($value instanceof Carbon) {
            return $value->toIso8601String();
        }

        if (is_numeric($value)) {
            return Carbon::createFromTimestamp($value)->toIso8601String();
        }

        if (is_string($value)) {
            try {
                return Carbon::parse($value)->toIso8601String();
            } catch (\Exception) {
                return (string) $value;
            }
        }

        return (string) $value;
    }

    protected function formatTimestamp(mixed $value, ?string $arg = null): int
    {
        if ($value instanceof Carbon) {
            return $value->timestamp;
        }

        if (is_numeric($value)) {
            return (int) $value;
        }

        if (is_string($value)) {
            try {
                return Carbon::parse($value)->timestamp;
            } catch (\Exception) {
                return 0;
            }
        }

        return 0;
    }

    protected function formatCurrency(mixed $value, ?string $arg = null): string
    {
        $decimals = $arg ? (int) $arg : 2;

        return number_format((float) $value, $decimals);
    }

    protected function formatJson(mixed $value, ?string $arg = null): string
    {
        return json_encode($value) ?: '""';
    }

    protected function formatUpper(mixed $value, ?string $arg = null): string
    {
        return mb_strtoupper((string) $value);
    }

    protected function formatLower(mixed $value, ?string $arg = null): string
    {
        return mb_strtolower((string) $value);
    }

    protected function formatDefault(mixed $value, ?string $arg = null): mixed
    {
        if ($value === null || $value === '' || (is_array($value) && empty($value))) {
            return $arg ?? '';
        }

        return $value;
    }

    protected function formatTruncate(mixed $value, ?string $arg = null): string
    {
        $length = $arg ? (int) $arg : 100;
        $string = (string) $value;

        if (mb_strlen($string) <= $length) {
            return $string;
        }

        return mb_substr($string, 0, $length - 3).'...';
    }

    protected function formatEscape(mixed $value, ?string $arg = null): string
    {
        return htmlspecialchars((string) $value, ENT_QUOTES | ENT_HTML5, 'UTF-8');
    }

    protected function formatUrlencode(mixed $value, ?string $arg = null): string
    {
        return urlencode((string) $value);
    }
}
