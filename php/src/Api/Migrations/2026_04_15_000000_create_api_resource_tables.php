<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('biolinks', function (Blueprint $table) {
            $table->id();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('title');
            $table->string('slug')->index();
            $table->string('url', 2048);
            $table->text('description')->nullable();
            $table->boolean('is_published')->default(true);
            $table->json('metadata')->nullable();
            $table->timestamps();
            $table->softDeletes();
            $table->unique(['workspace_id', 'slug']);
        });

        Schema::create('links', function (Blueprint $table) {
            $table->id();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('name');
            $table->string('slug')->nullable()->index();
            $table->string('destination_url', 2048);
            $table->string('short_code', 64)->unique();
            $table->boolean('is_active')->default(true);
            $table->unsignedBigInteger('click_count')->default(0);
            $table->timestamp('last_clicked_at')->nullable();
            $table->json('metadata')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('qr_codes', function (Blueprint $table) {
            $table->id();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('name');
            $table->string('target_url', 2048);
            $table->string('format', 16)->default('svg');
            $table->unsignedInteger('size')->default(256);
            $table->string('foreground_color', 32)->default('#000000');
            $table->string('background_color', 32)->default('#ffffff');
            $table->json('metadata')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('qr_codes');
        Schema::dropIfExists('links');
        Schema::dropIfExists('biolinks');
    }
};
