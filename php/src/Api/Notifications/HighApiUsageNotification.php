<?php

declare(strict_types=1);

namespace Core\Api\Notifications;

use Core\Tenant\Models\Workspace;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Notifications\Messages\MailMessage;
use Illuminate\Notifications\Notification;

/**
 * Notification sent when API usage approaches rate limits.
 *
 * Levels:
 * - warning: 80% of limit used
 * - critical: 95% of limit used
 */
class HighApiUsageNotification extends Notification implements ShouldQueue
{
    use Queueable;

    public function __construct(
        public Workspace $workspace,
        public string $level,
        public int $currentUsage,
        public int $limit,
        public string $period,
    ) {}

    /**
     * Get the notification's delivery channels.
     *
     * @return array<int, string>
     */
    public function via(object $notifiable): array
    {
        return ['mail'];
    }

    /**
     * Get the mail representation of the notification.
     */
    public function toMail(object $notifiable): MailMessage
    {
        $percentage = round(($this->currentUsage / $this->limit) * 100, 1);

        $subject = match ($this->level) {
            'critical' => "API Usage Critical - {$percentage}% of limit reached",
            default => "API Usage Warning - {$percentage}% of limit reached",
        };

        $message = (new MailMessage)
            ->subject($subject)
            ->greeting($this->getGreeting())
            ->line($this->getMainMessage())
            ->line("**Workspace:** {$this->workspace->name}")
            ->line("**Current usage:** {$this->currentUsage} requests")
            ->line("**Rate limit:** {$this->limit} requests per {$this->period}")
            ->line("**Usage:** {$percentage}%");

        if ($this->level === 'critical') {
            $message->line('If you exceed your rate limit, API requests will be temporarily blocked until the limit resets.');
        }

        $message->action('View API Usage', url('/developer/api'))
            ->line('Consider upgrading your plan if you regularly approach these limits.');

        return $message;
    }

    /**
     * Get the greeting based on level.
     */
    protected function getGreeting(): string
    {
        return match ($this->level) {
            'critical' => 'Warning: API Usage Critical',
            default => 'Notice: API Usage High',
        };
    }

    /**
     * Get the main message based on level.
     */
    protected function getMainMessage(): string
    {
        return match ($this->level) {
            'critical' => 'Your API usage has reached a critical level and is approaching the rate limit.',
            default => 'Your API usage is high and approaching the rate limit threshold.',
        };
    }

    /**
     * Get the array representation of the notification.
     *
     * @return array<string, mixed>
     */
    public function toArray(object $notifiable): array
    {
        return [
            'level' => $this->level,
            'workspace_id' => $this->workspace->id,
            'workspace_name' => $this->workspace->name,
            'current_usage' => $this->currentUsage,
            'limit' => $this->limit,
            'period' => $this->period,
        ];
    }
}
