<?php

declare(strict_types=1);

namespace Core\Api\Enums;

/**
 * Built-in webhook template types.
 *
 * Pre-defined template configurations for common webhook destinations.
 */
enum BuiltinTemplateType: string
{
    /**
     * Full event data - sends everything.
     */
    case FULL = 'full';

    /**
     * Minimal payload - essential fields only.
     */
    case MINIMAL = 'minimal';

    /**
     * Slack-formatted message.
     */
    case SLACK = 'slack';

    /**
     * Discord-formatted message.
     */
    case DISCORD = 'discord';

    /**
     * Get human-readable label for the type.
     */
    public function label(): string
    {
        return match ($this) {
            self::FULL => 'Full payload',
            self::MINIMAL => 'Minimal payload',
            self::SLACK => 'Slack message',
            self::DISCORD => 'Discord message',
        };
    }

    /**
     * Get description for the type.
     */
    public function description(): string
    {
        return match ($this) {
            self::FULL => 'Sends all event data in a structured format.',
            self::MINIMAL => 'Sends only essential fields: event type, ID, and timestamp.',
            self::SLACK => 'Formats payload for Slack incoming webhooks with blocks.',
            self::DISCORD => 'Formats payload for Discord webhooks with embeds.',
        };
    }

    /**
     * Get the default template content for this type.
     */
    public function template(): string
    {
        return match ($this) {
            self::FULL => <<<'JSON'
{
    "event": "{{event.type}}",
    "timestamp": "{{timestamp}}",
    "timestamp_unix": {{timestamp_unix}},
    "data": {{data | json}}
}
JSON,
            self::MINIMAL => <<<'JSON'
{
    "event": "{{event.type}}",
    "id": "{{data.id}}",
    "timestamp": "{{timestamp}}"
}
JSON,
            self::SLACK => <<<'JSON'
{
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "{{event.name}}"
            }
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "{{message}}"
            }
        },
        {
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": "*Event:* `{{event.type}}` | *Time:* {{timestamp | iso8601}}"
                }
            ]
        }
    ]
}
JSON,
            self::DISCORD => <<<'JSON'
{
    "embeds": [
        {
            "title": "{{event.name}}",
            "description": "{{message}}",
            "color": 5814783,
            "fields": [
                {
                    "name": "Event Type",
                    "value": "`{{event.type}}`",
                    "inline": true
                },
                {
                    "name": "ID",
                    "value": "{{data.id | default:N/A}}",
                    "inline": true
                }
            ],
            "timestamp": "{{timestamp}}"
        }
    ]
}
JSON,
        };
    }

    /**
     * Get the template format for this type.
     */
    public function format(): WebhookTemplateFormat
    {
        return WebhookTemplateFormat::JSON;
    }
}
