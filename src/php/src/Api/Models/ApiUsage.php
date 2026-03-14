<?php

declare(strict_types=1);

namespace Core\Api\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Core\Tenant\Models\Workspace;

/**
 * API Usage - individual API request log entry.
 *
 * Tracks each API call with timing, status, and size metrics.
 */
class ApiUsage extends Model
{
    public $timestamps = false;

    protected $table = 'api_usage';

    protected $fillable = [
        'api_key_id',
        'workspace_id',
        'endpoint',
        'method',
        'status_code',
        'response_time_ms',
        'request_size',
        'response_size',
        'ip_address',
        'user_agent',
        'created_at',
    ];

    protected $casts = [
        'created_at' => 'datetime',
    ];

    /**
     * Create a usage entry from request/response data.
     */
    public static function record(
        int $apiKeyId,
        int $workspaceId,
        string $endpoint,
        string $method,
        int $statusCode,
        int $responseTimeMs,
        ?int $requestSize = null,
        ?int $responseSize = null,
        ?string $ipAddress = null,
        ?string $userAgent = null
    ): static {
        return static::create([
            'api_key_id' => $apiKeyId,
            'workspace_id' => $workspaceId,
            'endpoint' => $endpoint,
            'method' => strtoupper($method),
            'status_code' => $statusCode,
            'response_time_ms' => $responseTimeMs,
            'request_size' => $requestSize,
            'response_size' => $responseSize,
            'ip_address' => $ipAddress,
            'user_agent' => $userAgent ? substr($userAgent, 0, 500) : null,
            'created_at' => now(),
        ]);
    }

    /**
     * Check if this was a successful request (2xx status).
     */
    public function isSuccess(): bool
    {
        return $this->status_code >= 200 && $this->status_code < 300;
    }

    /**
     * Check if this was a client error (4xx status).
     */
    public function isClientError(): bool
    {
        return $this->status_code >= 400 && $this->status_code < 500;
    }

    /**
     * Check if this was a server error (5xx status).
     */
    public function isServerError(): bool
    {
        return $this->status_code >= 500;
    }

    // Relationships
    public function apiKey(): BelongsTo
    {
        return $this->belongsTo(ApiKey::class);
    }

    public function workspace(): BelongsTo
    {
        return $this->belongsTo(Workspace::class);
    }

    // Scopes
    public function scopeForKey($query, int $apiKeyId)
    {
        return $query->where('api_key_id', $apiKeyId);
    }

    public function scopeForWorkspace($query, int $workspaceId)
    {
        return $query->where('workspace_id', $workspaceId);
    }

    public function scopeForEndpoint($query, string $endpoint)
    {
        return $query->where('endpoint', $endpoint);
    }

    public function scopeSuccessful($query)
    {
        return $query->whereBetween('status_code', [200, 299]);
    }

    public function scopeErrors($query)
    {
        return $query->where('status_code', '>=', 400);
    }

    public function scopeBetween($query, $startDate, $endDate)
    {
        return $query->whereBetween('created_at', [$startDate, $endDate]);
    }
}
