<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Api\Models\Concerns\BelongsToWorkspace;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class SupportTicket extends Model
{
    use BelongsToWorkspace;

    protected $table = 'support_tickets';

    protected $fillable = [
        'workspace_id',
        'user_id',
        'subject',
        'message',
        'status',
        'priority',
        'last_replied_at',
        'closed_at',
        'metadata',
    ];

    protected $casts = [
        'last_replied_at' => 'datetime',
        'closed_at' => 'datetime',
        'metadata' => 'array',
    ];

    /**
     * Replies appended to the ticket conversation.
     */
    public function replies(): HasMany
    {
        return $this->hasMany(SupportTicketReply::class, 'ticket_id');
    }
}
