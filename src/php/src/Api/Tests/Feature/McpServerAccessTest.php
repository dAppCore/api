<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;
use Core\Api\Models\ApiKey;
use Mod\Tenant\Models\User;
use Mod\Tenant\Models\Workspace;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();

    $this->user = User::factory()->create();
    $this->workspace = Workspace::factory()->create();
    $this->workspace->users()->attach($this->user->id, [
        'role' => 'owner',
        'is_default' => true,
    ]);

    $result = ApiKey::generate(
        $this->workspace->id,
        $this->user->id,
        'Restricted MCP Key',
        [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
    );

    $this->plainKey = $result['plain_key'];
    $result['api_key']->update(['server_scopes' => ['allowed-server']]);

    $this->mcpDir = resource_path('mcp');
    $this->serverDir = $this->mcpDir.'/servers';
    $this->registryFile = $this->mcpDir.'/registry.yaml';

    if (! is_dir($this->serverDir)) {
        mkdir($this->serverDir, 0777, true);
    }

    file_put_contents($this->registryFile, <<<YAML
servers:
  - id: allowed-server
  - id: blocked-server
YAML);

    file_put_contents($this->serverDir.'/allowed-server.yaml', <<<YAML
id: allowed-server
name: Allowed Server
status: available
tools:
  - name: ping
    description: Ping the server
    inputSchema:
      type: object
      properties:
        message:
          type: string
      required:
        - message
YAML);

    file_put_contents($this->serverDir.'/blocked-server.yaml', <<<YAML
id: blocked-server
name: Blocked Server
status: available
tools:
  - name: ping
    description: Ping the server
    inputSchema:
      type: object
      properties:
        message:
          type: string
      required:
        - message
YAML);
});

afterEach(function () {
    Cache::flush();

    $paths = [];

    if (isset($this->serverDir)) {
        $paths[] = $this->serverDir.'/allowed-server.yaml';
        $paths[] = $this->serverDir.'/blocked-server.yaml';
    }

    if (isset($this->registryFile)) {
        $paths[] = $this->registryFile;
    }

    foreach ($paths as $path) {
        if ($path && is_file($path)) {
            unlink($path);
        }
    }

    if (isset($this->serverDir) && is_dir($this->serverDir)) {
        @rmdir($this->serverDir);
    }

    if (isset($this->mcpDir) && is_dir($this->mcpDir)) {
        @rmdir($this->mcpDir);
    }
});

it('filters server discovery to accessible mcp servers', function () {
    $response = $this->getJson('/api/mcp/servers', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('count', 1);
    $response->assertJsonPath('servers.0.id', 'allowed-server');
    $response->assertJsonMissingPath('servers.1');
});

it('denies access to disallowed mcp servers', function () {
    $response = $this->postJson('/api/mcp/tools/call', [
        'server' => 'blocked-server',
        'tool' => 'ping',
        'arguments' => [
            'message' => 'hello',
        ],
    ], [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertStatus(403);
    $response->assertJsonPath('error', 'forbidden');
});

it('falls back to an empty registry when the registry file cannot be parsed', function () {
    file_put_contents($this->registryFile, <<<YAML
servers:
  - id: allowed-server
    name: [
YAML);

    $response = $this->getJson('/api/mcp/servers', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('count', 0);
    $response->assertJsonPath('servers', []);
});

it('returns a not found response when a server definition cannot be parsed', function () {
    file_put_contents($this->serverDir.'/allowed-server.yaml', <<<YAML
id: allowed-server
name: Allowed Server
status: available
tools:
  - name: ping
    description: Ping the server
    inputSchema:
      type: object
      properties:
        message:
          type: string
      required:
        - message
    description: [
YAML);

    $response = $this->getJson('/api/mcp/servers/allowed-server', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertNotFound();
    $response->assertJsonPath('error', 'not_found');
});
