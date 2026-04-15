<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Api\Models\Concerns\BelongsToWorkspace;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Link extends Model
{
    use BelongsToWorkspace;
    use SoftDeletes;

    protected $fillable = [
        'workspace_id',
        'user_id',
        'name',
        'slug',
        'destination_url',
        'short_code',
        'is_active',
        'click_count',
        'last_clicked_at',
        'metadata',
    ];

    protected $casts = [
        'is_active' => 'boolean',
        'click_count' => 'integer',
        'last_clicked_at' => 'datetime',
        'metadata' => 'array',
    ];
}
