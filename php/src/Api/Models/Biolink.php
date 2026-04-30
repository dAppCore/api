<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Api\Models\Concerns\BelongsToWorkspace;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Biolink extends Model
{
    use BelongsToWorkspace;
    use SoftDeletes;

    protected $fillable = [
        'workspace_id',
        'user_id',
        'title',
        'slug',
        'url',
        'description',
        'is_published',
        'metadata',
    ];

    protected $casts = [
        'is_published' => 'boolean',
        'metadata' => 'array',
    ];
}
