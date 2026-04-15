<?php

declare(strict_types=1);

use Core\Mod\Mcp\Services\ToolVersionService;
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
        'MCP Server Detail Key',
        [ApiKey::SCOPE_READ, ApiKey::SCOPE_WRITE]
    );

    $this->plainKey = $result['plain_key'];

    $this->serverId = 'test-detail-server';
    $this->serverDir = resource_path('mcp/servers');
    $this->serverFile = $this->serverDir.'/'.$this->serverId.'.yaml';

    if (! is_dir($this->serverDir)) {
        mkdir($this->serverDir, 0777, true);
    }

    file_put_contents($this->serverFile, <<<YAML
id: test-detail-server
name: Test Detail Server
status: available
tools:
  - name: search
    description: Search records
    inputSchema:
      type: object
      properties:
        query:
          type: string
      required:
        - query
resources:
  - uri: test-detail-server://documents/welcome
    path: documents/welcome
    name: welcome
    content:
      message: Hello from the server detail endpoint
      version: 2
YAML);

    app()->instance(ToolVersionService::class, new class
    {
        public function getLatestVersion(string $serverId, string $toolName): object
        {
            return (object) [
                'version' => '2.1.0',
                'is_deprecated' => false,
                'input_schema' => [
                    'type' => 'object',
                    'properties' => [
                        'query' => [
                            'type' => 'string',
                        ],
                    ],
                    'required' => ['query'],
                ],
            ];
        }
    });
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

it('includes tool versions and resource content on server detail requests when requested', function () {
    $response = $this->getJson('/api/mcp/servers/test-detail-server?include_versions=1&include_content=1', [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertOk();
    $response->assertJsonPath('id', 'test-detail-server');
    $response->assertJsonPath('tools.0.name', 'search');
    $response->assertJsonPath('tools.0.versioning.latest_version', '2.1.0');
    $response->assertJsonPath('tools.0.inputSchema.required.0', 'query');
    $response->assertJsonPath('resources.0.uri', 'test-detail-server://documents/welcome');
    $response->assertJsonPath('resources.0.content.message', 'Hello from the server detail endpoint');
    $response->assertJsonPath('resources.0.content.version', 2);
    $response->assertJsonPath('tool_count', 1);
    $response->assertJsonPath('resource_count', 1);
});

it('rejects unsafe server identifiers before filesystem-backed lookup', function () {
    $response = $this->postJson('/api/mcp/tools/call', [
        'server' => '../secrets',
        'tool' => 'search',
        'arguments' => [
            'query' => 'status',
        ],
    ], [
        'Authorization' => "Bearer {$this->plainKey}",
    ]);

    $response->assertStatus(422);
    $response->assertJsonValidationErrors(['server']);
});
