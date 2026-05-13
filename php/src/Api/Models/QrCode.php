<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Core\Api\Models\Concerns\BelongsToWorkspace;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class QrCode extends Model
{
    use BelongsToWorkspace;
    use SoftDeletes;

    protected $table = 'qr_codes';

    protected $fillable = [
        'workspace_id',
        'user_id',
        'name',
        'target_url',
        'format',
        'size',
        'foreground_color',
        'background_color',
        'metadata',
    ];

    protected $casts = [
        'size' => 'integer',
        'metadata' => 'array',
    ];
}
