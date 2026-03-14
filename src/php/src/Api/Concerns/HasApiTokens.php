<?php

declare(strict_types=1);

namespace Core\Api\Concerns;

use Core\Tenant\Models\UserToken;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Support\Str;

/**
 * Trait for models that can have API tokens.
 *
 * Provides methods to create and manage personal access tokens
 * for API authentication.
 */
trait HasApiTokens
{
    /**
     * Get all API tokens for this user.
     *
     * @return HasMany<UserToken>
     */
    public function tokens(): HasMany
    {
        return $this->hasMany(UserToken::class);
    }

    /**
     * Create a new personal access token for the user.
     *
     * @param  string  $name  Human-readable name for the token
     * @param  \DateTimeInterface|null  $expiresAt  Optional expiration date
     * @return array{token: string, model: UserToken} Plain-text token and model instance
     */
    public function createToken(string $name, ?\DateTimeInterface $expiresAt = null): array
    {
        // Generate a random 40-character token
        $plainTextToken = Str::random(40);

        // Hash it for storage
        $hashedToken = hash('sha256', $plainTextToken);

        // Create the token record
        $token = $this->tokens()->create([
            'name' => $name,
            'token' => $hashedToken,
            'expires_at' => $expiresAt,
        ]);

        return [
            'token' => $plainTextToken,
            'model' => $token,
        ];
    }

    /**
     * Revoke all tokens for this user.
     *
     * @return int Number of tokens deleted
     */
    public function revokeAllTokens(): int
    {
        return $this->tokens()->delete();
    }

    /**
     * Revoke a specific token by its ID.
     *
     * @return bool True if the token was deleted
     */
    public function revokeToken(int $tokenId): bool
    {
        return (bool) $this->tokens()->where('id', $tokenId)->delete();
    }
}
