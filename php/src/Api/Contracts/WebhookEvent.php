<?php

declare(strict_types=1);

namespace Core\Api\Contracts;

/**
 * Contract for webhook events that can be rendered with templates.
 *
 * Any event that wants to use webhook templating must implement this interface.
 */
interface WebhookEvent
{
    /**
     * Get the event identifier (e.g., 'post.published', 'user.created').
     */
    public static function name(): string;

    /**
     * Get the human-readable event name for display.
     */
    public static function nameLocalised(): string;

    /**
     * Get the event payload data.
     *
     * @return array<string, mixed>
     */
    public function payload(): array;

    /**
     * Get a human-readable message describing the event.
     */
    public function message(): string;
}
