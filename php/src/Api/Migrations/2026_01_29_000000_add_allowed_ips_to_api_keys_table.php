<?php

declare(strict_types=1);

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     *
     * Adds IP whitelisting support to API keys:
     * - allowed_ips: JSON array of IP addresses and/or CIDR ranges
     *
     * When allowed_ips is null or empty, no IP restrictions apply.
     * When populated, only requests from whitelisted IPs are accepted.
     *
     * Supports both IPv4 and IPv6 addresses and CIDR notation:
     * - Individual IPs: "192.168.1.1", "::1"
     * - CIDR ranges: "192.168.0.0/24", "2001:db8::/32"
     */
    public function up(): void
    {
        Schema::table('api_keys', function (Blueprint $table) {
            $table->json('allowed_ips')
                ->nullable()
                ->after('server_scopes')
                ->comment('IP whitelist: null=all IPs allowed, ["192.168.1.0/24"]=specific');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('api_keys', function (Blueprint $table) {
            $table->dropColumn('allowed_ips');
        });
    }
};
