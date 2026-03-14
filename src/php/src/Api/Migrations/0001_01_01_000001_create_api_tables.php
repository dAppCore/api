<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * API module tables.
     *
     * Creates tables for reusable webhook payload templates that can be
     * shared across different webhook configurations.
     */
    public function up(): void
    {
        Schema::disableForeignKeyConstraints();

        // Webhook Payload Templates
        // Reusable templates for customising webhook payload shapes
        Schema::create('api_webhook_payload_templates', function (Blueprint $table) {
            $table->id();
            $table->uuid('uuid')->unique();
            $table->foreignId('workspace_id')->constrained('workspaces')->cascadeOnDelete();
            $table->foreignId('namespace_id')->nullable()->constrained('namespaces')->nullOnDelete();

            $table->string('name');
            $table->text('description')->nullable();

            // Template format: simple, mustache, json
            $table->string('format', 20)->default('simple');

            // The actual template content (JSON/Twig-like syntax)
            $table->text('template');

            // Example rendered output for preview
            $table->text('example_output')->nullable();

            // Template metadata
            $table->boolean('is_default')->default(false);
            $table->unsignedSmallInteger('sort_order')->default(0);
            $table->boolean('is_active')->default(true);

            // Built-in template type (null for custom templates)
            // Values: full, minimal, slack, discord
            $table->string('builtin_type', 20)->nullable();

            $table->timestamps();

            $table->index(['workspace_id', 'is_active']);
            $table->index(['workspace_id', 'is_default']);
            $table->index(['workspace_id', 'sort_order']);
            $table->index('builtin_type');
        });

        Schema::enableForeignKeyConstraints();
    }

    public function down(): void
    {
        Schema::dropIfExists('api_webhook_payload_templates');
    }
};
