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
        Schema::create('webhook_deliveries', function (Blueprint $table) {
            $table->id();
            $table->foreignId('webhook_endpoint_id')->constrained('webhook_endpoints')->cascadeOnDelete();
            $table->string('event_id', 32)->comment('Unique event identifier (evt_xxx)');
            $table->string('event_type', 64)->index();
            $table->json('payload');
            $table->unsignedSmallInteger('response_code')->nullable();
            $table->text('response_body')->nullable();
            $table->unsignedTinyInteger('attempt')->default(1);
            $table->string('status', 16)->default('pending')->comment('pending, success, failed, retrying');
            $table->timestamp('delivered_at')->nullable();
            $table->timestamp('next_retry_at')->nullable()->index();
            $table->timestamps();

            $table->index(['webhook_endpoint_id', 'status']);
            $table->index(['status', 'next_retry_at']);
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('webhook_deliveries');
    }
};
