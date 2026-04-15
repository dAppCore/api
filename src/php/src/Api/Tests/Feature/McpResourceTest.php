<?php

declare(strict_types=1);

use Illuminate\Support\Facades\Cache;
use Mod\Api\Models\ApiKey;
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
        'MCP Resource Key',
        [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
    );

    $this->plainKey = $result['plain_key'];

    $this->serverId = 'test-resource-server';
    $this->serverDir = resource_path('mcp/servers');
    $this->serverFile = $this->serverDir.'/'.$this->serverId.'.yaml';

    if (! is_dir($this->serverDir)) {
        mkdir($this->serverDir, 0777, true);
    }

    file_put_contents($this->serverFile, <<<YAML
id: test-resource-server
name: Test Resource Server
status: available
resources:
  - uri: test-resource-server://documents/welcome
    path: documents/welcome
    name: welcome
    content:
      message: Hello from the MCP resource bridge
      version: 1
YAML);
});

afterEach(function () {
    Cache::flush();

    if (isset($this->serverFile) && is_file($this->serverFile)) {
        unlink($this->serverFile);
    }

    if (isset($this->serverDir) && is_dir($this->serverDir)) {
        @rmdir($this->serverDir);
    }

    $mcpDir = dirname($this->serverDir ?? '');
    if (is_dir($mcpDir)) {
        @rmdir($mcpDir);
    }
});

it('reads a resource from the server definition', function () {
    $encodedUri = rawurlencode('test-resource-server://documents/welcome');

    $response = $this->getJson("/api/mcp/resources/{$encodedUri}", [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJson([
        'uri' => 'test-resource-server://documents/welcome',
        'server' => 'test-resource-server',
        'resource' => 'documents/welcome',
    ]);

    expect($response->json('content'))->toBe([
        'message' => 'Hello from the MCP resource bridge',
        'version' => 1,
    ]);
});

it('lists resources for a server', function () {
    $response = $this->getJson('/api/mcp/servers/test-resource-server/resources', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('server', 'test-resource-server');
    $response->assertJsonPath('count', 1);
    $response->assertJsonPath('resources.0.uri', 'test-resource-server://documents/welcome');
    $response->assertJsonPath('resources.0.path', 'documents/welcome');
    $response->assertJsonPath('resources.0.name', 'welcome');
    $response->assertJsonMissingPath('resources.0.content');
});

it('rejects unsafe resource paths before dispatching to the server', function () {
    $encodedUri = rawurlencode('test-resource-server://../secrets');

    $response = $this->getJson("/api/mcp/resources/{$encodedUri}", [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertStatus(422);
    $response->assertJsonValidationErrors(['uri']);
});
