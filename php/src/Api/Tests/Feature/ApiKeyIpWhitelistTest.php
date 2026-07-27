<?php

declare(strict_types=1);

use Core\Api\Services\IpRestrictionService;
use Illuminate\Support\Facades\Cache;
use Core\Api\Models\ApiKey;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;

define('IP_192_168_1_1', '192.168.1.1');
define('IP_10_0_0_1', '10.0.0.1');
define('CIDR_192_168_1_0_24', '192.168.1.0/24');
define('CIDR_10_0_0_0_8', '10.0.0.0/8');
define('IP_2001_DB8_1', '2001:db8::1');
define('CIDR_2001_DB8_32', '2001:db8::/32');

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();

    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);
    $this->ipService = app(IpRestrictionService::class);
});

// ─────────────────────────────────────────────────────────────────────────────
// IP Restriction Service - IPv4
// ─────────────────────────────────────────────────────────────────────────────

describe('IP Restriction Service - IPv4', function () {
    it('allows IP when whitelist is empty', function () {
        expect($this->ipService->isIpAllowed(IP_192_168_1_1, []))->toBeTrue();
    });

    it('matches exact IPv4 address', function () {
        $whitelist = [IP_192_168_1_1, IP_10_0_0_1];

        expect($this->ipService->isIpAllowed(IP_192_168_1_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed(IP_10_0_0_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('192.168.1.2', $whitelist))->toBeFalse();
    });

    it('matches IPv4 CIDR /24 range', function () {
        $whitelist = [CIDR_192_168_1_0_24];

        expect($this->ipService->isIpAllowed('192.168.1.0', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed(IP_192_168_1_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('192.168.1.255', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('192.168.2.1', $whitelist))->toBeFalse();
    });

    it('matches IPv4 CIDR /16 range', function () {
        $whitelist = ['10.0.0.0/16'];

        expect($this->ipService->isIpAllowed(IP_10_0_0_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('10.0.255.255', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('10.1.0.1', $whitelist))->toBeFalse();
    });

    it('matches IPv4 CIDR /32 single host', function () {
        $whitelist = ['192.168.1.100/32'];

        expect($this->ipService->isIpAllowed('192.168.1.100', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('192.168.1.101', $whitelist))->toBeFalse();
    });

    it('matches IPv4 CIDR /8 class A range', function () {
        $whitelist = [CIDR_10_0_0_0_8];

        expect($this->ipService->isIpAllowed(IP_10_0_0_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('10.255.255.255', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('11.0.0.1', $whitelist))->toBeFalse();
    });

    it('rejects invalid IPv4 addresses', function () {
        $whitelist = [CIDR_192_168_1_0_24];

        expect($this->ipService->isIpAllowed('invalid', $whitelist))->toBeFalse();
        expect($this->ipService->isIpAllowed('256.256.256.256', $whitelist))->toBeFalse();
        expect($this->ipService->isIpAllowed('', $whitelist))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// IP Restriction Service - IPv6
// ─────────────────────────────────────────────────────────────────────────────

describe('IP Restriction Service - IPv6', function () {
    it('matches exact IPv6 address', function () {
        $whitelist = ['::1', IP_2001_DB8_1];

        expect($this->ipService->isIpAllowed('::1', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db8::2', $whitelist))->toBeFalse();
    });

    it('normalises IPv6 for comparison', function () {
        $whitelist = ['2001:db8:0000:0000:0000:0000:0000:0001'];

        // Shortened form should match expanded form
        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeTrue();
    });

    it('matches IPv6 CIDR /64 range', function () {
        $whitelist = ['2001:db8::/64'];

        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db8::ffff', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db8:0:1::1', $whitelist))->toBeFalse();
    });

    it('matches IPv6 CIDR /32 range', function () {
        $whitelist = [CIDR_2001_DB8_32];

        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db8:ffff::1', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db9::1', $whitelist))->toBeFalse();
    });

    it('matches IPv6 loopback', function () {
        $whitelist = ['::1/128'];

        expect($this->ipService->isIpAllowed('::1', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('::2', $whitelist))->toBeFalse();
    });

    it('does not match IPv4 against IPv6 CIDR', function () {
        $whitelist = [CIDR_2001_DB8_32];

        expect($this->ipService->isIpAllowed(IP_192_168_1_1, $whitelist))->toBeFalse();
    });

    it('does not match IPv6 against IPv4 CIDR', function () {
        $whitelist = [CIDR_192_168_1_0_24];

        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// IP Restriction Service - Validation
// ─────────────────────────────────────────────────────────────────────────────

describe('IP Restriction Service - Validation', function () {
    it('validates correct IPv4 addresses', function () {
        $result = $this->ipService->validateEntry(IP_192_168_1_1);

        expect($result['valid'])->toBeTrue();
        expect($result['error'])->toBeNull();
    });

    it('validates correct IPv6 addresses', function () {
        $result = $this->ipService->validateEntry(IP_2001_DB8_1);

        expect($result['valid'])->toBeTrue();
        expect($result['error'])->toBeNull();
    });

    it('validates correct IPv4 CIDR', function () {
        $result = $this->ipService->validateEntry(CIDR_192_168_1_0_24);

        expect($result['valid'])->toBeTrue();
        expect($result['error'])->toBeNull();
    });

    it('validates correct IPv6 CIDR', function () {
        $result = $this->ipService->validateEntry(CIDR_2001_DB8_32);

        expect($result['valid'])->toBeTrue();
        expect($result['error'])->toBeNull();
    });

    it('rejects invalid IP addresses', function () {
        $result = $this->ipService->validateEntry('not-an-ip');

        expect($result['valid'])->toBeFalse();
        expect($result['error'])->toBe('Invalid IP address');
    });

    it('rejects invalid CIDR prefix for IPv4', function () {
        $result = $this->ipService->validateEntry('192.168.1.0/33');

        expect($result['valid'])->toBeFalse();
        expect($result['error'])->toBe('IPv4 prefix must be between 0 and 32');
    });

    it('rejects invalid CIDR prefix for IPv6', function () {
        $result = $this->ipService->validateEntry('2001:db8::/129');

        expect($result['valid'])->toBeFalse();
        expect($result['error'])->toBe('IPv6 prefix must be between 0 and 128');
    });

    it('rejects empty entries', function () {
        $result = $this->ipService->validateEntry('');

        expect($result['valid'])->toBeFalse();
        expect($result['error'])->toBe('Empty entry');
    });

    it('parses multi-line whitelist input', function () {
        $input = "192.168.1.1\n10.0.0.0/8\n# Comment line\n2001:db8::1\ninvalid-ip";

        $result = $this->ipService->parseWhitelistInput($input);

        expect($result['entries'])->toBe([IP_192_168_1_1, CIDR_10_0_0_0_8, IP_2001_DB8_1]);
        expect($result['errors'])->toHaveCount(1);
        expect($result['errors'][0])->toContain('invalid-ip');
    });

    it('handles comma-separated whitelist input', function () {
        $input = '192.168.1.1, 10.0.0.1, 172.16.0.0/12';

        $result = $this->ipService->parseWhitelistInput($input);

        expect($result['entries'])->toBe([IP_192_168_1_1, IP_10_0_0_1, '172.16.0.0/12']);
        expect($result['errors'])->toBeEmpty();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key IP Whitelist Model Methods
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key IP Whitelist Model', function () {
    it('reports no restrictions when allowed_ips is null', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'No Restrictions Key'
        );

        expect($result['api_key']->hasIpRestrictions())->toBeFalse();
        expect($result['api_key']->getAllowedIps())->toBeNull();
    });

    it('reports no restrictions when allowed_ips is empty', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Empty Whitelist Key'
        );
        $result['api_key']->update(['allowed_ips' => []]);

        expect($result['api_key']->fresh()->hasIpRestrictions())->toBeFalse();
    });

    it('reports restrictions when allowed_ips has entries', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Restricted Key'
        );
        $result['api_key']->update(['allowed_ips' => [CIDR_192_168_1_0_24]]);

        $key = $result['api_key']->fresh();
        expect($key->hasIpRestrictions())->toBeTrue();
        expect($key->getAllowedIps())->toBe([CIDR_192_168_1_0_24]);
    });

    it('updates allowed IPs', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Update IPs Key'
        );

        $result['api_key']->updateAllowedIps([CIDR_10_0_0_0_8, IP_192_168_1_1]);

        expect($result['api_key']->fresh()->getAllowedIps())->toBe([CIDR_10_0_0_0_8, IP_192_168_1_1]);
    });

    it('adds IP to whitelist', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Add IP Key'
        );
        $result['api_key']->update(['allowed_ips' => [IP_192_168_1_1]]);

        $result['api_key']->addAllowedIp(IP_10_0_0_1);

        expect($result['api_key']->fresh()->getAllowedIps())->toBe([IP_192_168_1_1, IP_10_0_0_1]);
    });

    it('does not add duplicate IPs', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Duplicate IP Key'
        );
        $result['api_key']->update(['allowed_ips' => [IP_192_168_1_1]]);

        $result['api_key']->addAllowedIp(IP_192_168_1_1);

        expect($result['api_key']->fresh()->getAllowedIps())->toBe([IP_192_168_1_1]);
    });

    it('removes IP from whitelist', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Remove IP Key'
        );
        $result['api_key']->update(['allowed_ips' => [IP_192_168_1_1, IP_10_0_0_1]]);

        $result['api_key']->removeAllowedIp(IP_192_168_1_1);

        expect($result['api_key']->fresh()->getAllowedIps())->toBe([IP_10_0_0_1]);
    });

    it('sets allowed_ips to null when removing last IP', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'Remove Last IP Key'
        );
        $result['api_key']->update(['allowed_ips' => [IP_192_168_1_1]]);

        $result['api_key']->removeAllowedIp(IP_192_168_1_1);

        expect($result['api_key']->fresh()->getAllowedIps())->toBeNull();
        expect($result['api_key']->fresh()->hasIpRestrictions())->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Rotation with IP Whitelist
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Rotation with IP Whitelist', function () {
    it('preserves IP whitelist during rotation', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'IP Restricted Key'
        );
        $result['api_key']->update(['allowed_ips' => [CIDR_192_168_1_0_24, IP_10_0_0_1]]);

        $rotated = $result['api_key']->fresh()->rotate();

        expect($rotated['api_key']->getAllowedIps())->toBe([CIDR_192_168_1_0_24, IP_10_0_0_1]);
    });

    it('preserves empty IP whitelist during rotation', function () {
        $result = ApiKey::generate(
            $this->workspace->id,
            $this->user->id,
            'No Restrictions Key'
        );

        $rotated = $result['api_key']->rotate();

        expect($rotated['api_key']->getAllowedIps())->toBeNull();
        expect($rotated['api_key']->hasIpRestrictions())->toBeFalse();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// API Key Factory IP Whitelist
// ─────────────────────────────────────────────────────────────────────────────

describe('API Key Factory IP Whitelist', function () {
    it('creates keys with IP whitelist via factory', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->withAllowedIps([CIDR_192_168_1_0_24, '::1'])
            ->create();

        expect($key->hasIpRestrictions())->toBeTrue();
        expect($key->getAllowedIps())->toBe([CIDR_192_168_1_0_24, '::1']);
    });

    it('creates keys without IP restrictions by default', function () {
        $key = ApiKey::factory()
            ->for($this->workspace)
            ->for($this->user)
            ->create();

        expect($key->hasIpRestrictions())->toBeFalse();
        expect($key->getAllowedIps())->toBeNull();
    });
});

// ─────────────────────────────────────────────────────────────────────────────
// Mixed IP Versions
// ─────────────────────────────────────────────────────────────────────────────

describe('Mixed IP Versions in Whitelist', function () {
    it('handles mixed IPv4 and IPv6 entries', function () {
        $whitelist = [CIDR_192_168_1_0_24, CIDR_2001_DB8_32, IP_10_0_0_1, '::1'];

        // IPv4 matching
        expect($this->ipService->isIpAllowed('192.168.1.100', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed(IP_10_0_0_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('172.16.0.1', $whitelist))->toBeFalse();

        // IPv6 matching
        expect($this->ipService->isIpAllowed(IP_2001_DB8_1, $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('::1', $whitelist))->toBeTrue();
        expect($this->ipService->isIpAllowed('2001:db9::1', $whitelist))->toBeFalse();
    });
});
