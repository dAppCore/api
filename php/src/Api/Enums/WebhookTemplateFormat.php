<?php

declare(strict_types=1);

namespace Core\Api\Enums;

/**
 * Webhook payload template formats.
 *
 * Defines supported template syntaxes for customising webhook payloads.
 */
enum WebhookTemplateFormat: string
{
    /**
     * Simple variable substitution: {{variable.path}}
     */
    case SIMPLE = 'simple';

    /**
     * Mustache-style templates with conditionals and loops.
     */
    case MUSTACHE = 'mustache';

    /**
     * Raw JSON with variable interpolation.
     */
    case JSON = 'json';

    /**
     * Get human-readable label for the format.
     */
    public function label(): string
    {
        return match ($this) {
            self::SIMPLE => 'Simple (variable substitution)',
            self::MUSTACHE => 'Mustache (conditionals and loops)',
            self::JSON => 'JSON (structured template)',
        };
    }

    /**
     * Get description for the format.
     */
    public function description(): string
    {
        return match ($this) {
            self::SIMPLE => 'Basic {{variable}} replacement. Best for simple payloads.',
            self::MUSTACHE => 'Full Mustache syntax with {{#if}}, {{#each}}, and filters.',
            self::JSON => 'JSON template with embedded {{variables}}. Validates structure.',
        };
    }

    /**
     * Get example template for the format.
     */
    public function example(): string
    {
        return match ($this) {
            self::SIMPLE => '{"event": "{{event.type}}", "id": "{{data.id}}"}',
            self::MUSTACHE => '{"event": "{{event.type}}"{{#if data.user}}, "user": "{{data.user.name}}"{{/if}}}',
            self::JSON => <<<'JSON'
{
    "event": "{{event.type}}",
    "timestamp": "{{timestamp | iso8601}}",
    "data": {
        "id": "{{data.id}}",
        "name": "{{data.name}}"
    }
}
JSON,
        };
    }
}
