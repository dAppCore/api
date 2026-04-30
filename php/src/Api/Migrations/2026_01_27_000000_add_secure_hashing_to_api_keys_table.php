<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     *
     * Adds columns to support:
     * - Secure hashing with bcrypt/Argon2 (hash_algorithm tracks which was used)
     * - Key rotation with grace periods
     * - Tracking which key was rotated from
     */
    public function up(): void
    {
        Schema::table('api_keys', function (Blueprint $table) {
            // Track hash algorithm for backward compatibility during migration
            // 'sha256' = legacy unsalted hash, 'bcrypt' = secure hash
            $table->string('hash_algorithm', 16)->default('sha256')->after('key');

            // Grace period for key rotation - old key remains valid until this time
            $table->timestamp('grace_period_ends_at')->nullable()->after('expires_at');

            // Track key rotation lineage
            $table->foreignId('rotated_from_id')
                ->nullable()
                ->after('grace_period_ends_at')
                ->constrained('api_keys')
                ->nullOnDelete();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('api_keys', function (Blueprint $table) {
            $table->dropForeign(['rotated_from_id']);
            $table->dropColumn(['hash_algorithm', 'grace_period_ends_at', 'rotated_from_id']);
        });
    }
};
