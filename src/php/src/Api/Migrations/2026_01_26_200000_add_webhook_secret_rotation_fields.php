<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

/**
 * Add secret rotation grace period fields to webhook tables.
 *
 * This migration adds support for webhook secret rotation with a grace period,
 * allowing both old and new secrets to be accepted during the transition.
 */
return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        // Add grace period fields to social_webhooks
        if (Schema::hasTable('social_webhooks')) {
            Schema::table('social_webhooks', function (Blueprint $table) {
                $table->text('previous_secret')->nullable()->after('secret');
                $table->timestamp('secret_rotated_at')->nullable()->after('previous_secret');
                $table->unsignedInteger('grace_period_seconds')->default(86400)->after('secret_rotated_at'); // 24 hours
            });
        }

        // Add grace period fields to content_webhook_endpoints
        if (Schema::hasTable('content_webhook_endpoints')) {
            Schema::table('content_webhook_endpoints', function (Blueprint $table) {
                $table->text('previous_secret')->nullable()->after('secret');
                $table->timestamp('secret_rotated_at')->nullable()->after('previous_secret');
                $table->unsignedInteger('grace_period_seconds')->default(86400)->after('secret_rotated_at'); // 24 hours
            });
        }
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        if (Schema::hasTable('social_webhooks')) {
            Schema::table('social_webhooks', function (Blueprint $table) {
                $table->dropColumn(['previous_secret', 'secret_rotated_at', 'grace_period_seconds']);
            });
        }

        if (Schema::hasTable('content_webhook_endpoints')) {
            Schema::table('content_webhook_endpoints', function (Blueprint $table) {
                $table->dropColumn(['previous_secret', 'secret_rotated_at', 'grace_period_seconds']);
            });
        }
    }
};
