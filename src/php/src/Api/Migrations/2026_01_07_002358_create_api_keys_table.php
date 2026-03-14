<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('api_keys', function (Blueprint $table) {
            $table->id();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('name');
            $table->string('key', 64)->comment('SHA256 hash of the key');
            $table->string('prefix', 16)->comment('Key prefix for identification (hk_xxxxxxxx)');
            $table->json('scopes')->default('["read","write"]');
            $table->json('server_scopes')->nullable()->comment('Per-server access: null=all, ["commerce","biohost"]=specific');
            $table->timestamp('last_used_at')->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->softDeletes();
            $table->timestamps();

            // Index for key lookup
            $table->index(['prefix', 'key']);
            $table->index('workspace_id');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('api_keys');
    }
};
