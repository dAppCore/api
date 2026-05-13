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
        Schema::create('webhook_endpoints', function (Blueprint $table) {
            $table->id();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->string('url');
            $table->string('secret', 64)->comment('HMAC signing secret');
            $table->json('events')->comment('Event types to receive, or ["*"] for all');
            $table->boolean('active')->default(true);
            $table->string('description')->nullable();
            $table->timestamp('last_triggered_at')->nullable();
            $table->unsignedInteger('failure_count')->default(0);
            $table->timestamp('disabled_at')->nullable()->comment('Auto-disabled after 10 consecutive failures');
            $table->timestamps();
            $table->softDeletes();

            $table->index(['workspace_id', 'active']);
            $table->index(['active', 'disabled_at']);
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('webhook_endpoints');
    }
};
