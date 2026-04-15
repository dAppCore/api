<?php

declare(strict_types=1);

namespace Mod\Api\Database\Factories;

use Illuminate\Database\Eloquent\Factories\Factory;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Mod\Api\Models\ApiKey;
use Mod\Tenant\Models\User;
use Mod\Tenant\Models\Workspace;

/**
 * Factory for generating ApiKey test instances.
 *
 * By default, creates keys with the RFC key format and bcrypt hashing.
 * Use legacyHash() to create keys with the older hk_ prefix for migration testing.
 *
 * @extends Factory<ApiKey>
 */
class ApiKeyFactory extends Factory
{
    /**
     * The name of the factory's corresponding model.
     *
     * @var class-string<ApiKey>
     */
    protected $model = ApiKey::class;

    /**
     * Store the plain key for testing.
     */
    private ?string $plainKey = null;

    /**
     * Define the model's default state.
     *
     * Creates keys with the RFC key format and bcrypt hashing by default.
     *
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        $plainKey = Str::random(48);
        $prefix = ApiKey::generatePrefix();
        $this->plainKey = "{$prefix}_{$plainKey}";

        return [
            'workspace_id' => Workspace::factory(),
            'user_id' => User::factory(),
            'name' => fake()->words(2, true).' API Key',
            'key' => Hash::driver('bcrypt')->make($plainKey),
            'hash_algorithm' => ApiKey::HASH_BCRYPT,
            'prefix' => $prefix,
            'scopes' => [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE],
            'server_scopes' => null,
            'allowed_ips' => null,
            'last_used_at' => null,
            'expires_at' => null,
            'grace_period_ends_at' => null,
            'rotated_from_id' => null,
        ];
    }

    /**
     * Get the plain key after creation.
     * Must be called immediately after create() to get the plain key.
     */
    public function getPlainKey(): ?string
    {
        return $this->plainKey;
    }

    /**
     * Create a key with specific known credentials for testing.
     *
     * This method uses ApiKey::generate() which creates secure bcrypt keys.
     *
     * @return array{api_key: ApiKey, plain_key: string}
     */
    public static function createWithPlainKey(
        ?Workspace $workspace = null,
        ?User $user = null,
        array $scopes = [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE],
        ?\DateTimeInterface $expiresAt = null
    ): array {
        $workspace ??= Workspace::factory()->create();
        $user ??= User::factory()->create();

        return ApiKey::generate(
            $workspace->id,
            $user->id,
            fake()->words(2, true).' API Key',
            $scopes,
            $expiresAt
        );
    }

    /**
     * Create a key with legacy SHA-256 hashing for migration testing.
     *
     * @return array{api_key: ApiKey, plain_key: string}
     */
    public static function createLegacyKey(
        ?Workspace $workspace = null,
        ?User $user = null,
        array $scopes = [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE],
        ?\DateTimeInterface $expiresAt = null
    ): array {
        $workspace ??= Workspace::factory()->create();
        $user ??= User::factory()->create();

        $plainKey = Str::random(48);
        $prefix = ApiKey::generatePrefix();

        $apiKey = ApiKey::create([
            'workspace_id' => $workspace->id,
            'user_id' => $user->id,
            'name' => fake()->words(2, true).' API Key',
            'key' => hash('sha256', $plainKey),
            'hash_algorithm' => ApiKey::HASH_SHA256,
            'prefix' => $prefix,
            'scopes' => $scopes,
            'expires_at' => $expiresAt,
        ]);

        return [
            'api_key' => $apiKey,
            'plain_key' => "{$prefix}_{$plainKey}",
        ];
    }

    /**
     * Create key with legacy SHA-256 hashing (for migration testing).
     */
    public function legacyHash(): static
    {
        return $this->state(function (array $attributes) {
            // Extract the plain key from the stored state
            $parts = explode('_', $this->plainKey ?? '', 3);
            $plainKey = $parts[2] ?? Str::random(48);

            return [
                'key' => hash('sha256', $plainKey),
                'hash_algorithm' => ApiKey::HASH_SHA256,
            ];
        });
    }

    /**
     * Indicate that the key has been used recently.
     */
    public function used(): static
    {
        return $this->state(fn (array $attributes) => [
            'last_used_at' => now()->subMinutes(fake()->numberBetween(1, 60)),
        ]);
    }

    /**
     * Indicate that the key expires in the future.
     *
     * @param  int  $days  Number of days until expiration
     */
    public function expiresIn(int $days = 30): static
    {
        return $this->state(fn (array $attributes) => [
            'expires_at' => now()->addDays($days),
        ]);
    }

    /**
     * Indicate that the key has expired.
     */
    public function expired(): static
    {
        return $this->state(fn (array $attributes) => [
            'expires_at' => now()->subDays(1),
        ]);
    }

    /**
     * Set specific scopes.
     *
     * @param  array<string>  $scopes
     */
    public function withScopes(array $scopes): static
    {
        return $this->state(fn (array $attributes) => [
            'scopes' => $scopes,
        ]);
    }

    /**
     * Set read-only scope.
     */
    public function readOnly(): static
    {
        return $this->withScopes([ApiKey::SCOPE_READ]);
    }

    /**
     * Set all scopes (read, write, delete).
     */
    public function fullAccess(): static
    {
        return $this->withScopes(ApiKey::ALL_SCOPES);
    }

    /**
     * Set specific server scopes.
     *
     * @param  array<string>|null  $servers
     */
    public function withServerScopes(?array $servers): static
    {
        return $this->state(fn (array $attributes) => [
            'server_scopes' => $servers,
        ]);
    }

    /**
     * Set IP whitelist restrictions.
     *
     * @param  array<string>|null  $ips  Array of IP addresses/CIDRs
     */
    public function withAllowedIps(?array $ips): static
    {
        return $this->state(fn (array $attributes) => [
            'allowed_ips' => $ips,
        ]);
    }

    /**
     * Create a revoked (soft-deleted) key.
     */
    public function revoked(): static
    {
        return $this->state(fn (array $attributes) => [
            'deleted_at' => now()->subDay(),
        ]);
    }

    /**
     * Create a key in a rotation grace period.
     *
     * @param  int  $hoursRemaining  Hours until grace period ends
     */
    public function inGracePeriod(int $hoursRemaining = 12): static
    {
        return $this->state(fn (array $attributes) => [
            'grace_period_ends_at' => now()->addHours($hoursRemaining),
        ]);
    }

    /**
     * Create a key with an expired grace period.
     */
    public function gracePeriodExpired(): static
    {
        return $this->state(fn (array $attributes) => [
            'grace_period_ends_at' => now()->subHours(1),
        ]);
    }
}
