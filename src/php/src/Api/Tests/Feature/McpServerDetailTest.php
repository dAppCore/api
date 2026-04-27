<?php

declare(strict_types=1);

use Core\Mod\Mcp\Services\ToolVersionService;
use Illuminate\Support\Facades\Cache;
use Core\Api\Controllers\McpApiController;
use Core\Api\Models\ApiKey;
use Core\Tenant\Models\User;
use Core\Tenant\Models\Workspace;
use Illuminate\Http\Request;

uses(\Illuminate\Foundation\Testing\RefreshDatabase::class);

beforeEach(function () {
    Cache::flush();

    $this->user = User::query()->create([
        'name' => fake()->name(),
        'email' => fake()->unique()->safeEmail(),
        'password' => 'password',
    ]);
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
    $this->apiKey = $result['api_key'];

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

it('McpApiController_callToolByRoute_Good_uses_route_parameters_with_a_test_seam', function () {
    app()->instance(ToolVersionService::class, new class
    {
        public function resolveVersion(string $server, string $tool, ?string $version): array
        {
            return [
                'version' => null,
                'warning' => null,
                'error' => null,
            ];
        }
    });

    $controller = new class extends McpApiController
    {
        public array $calls = [];

        protected function executeToolViaArtisan(string $server, string $tool, array $arguments, ?string $version = null): mixed
        {
            $this->calls[] = compact('server', 'tool', 'arguments', 'version');

            return [
                'ok' => true,
                'server' => $server,
                'tool' => $tool,
                'arguments' => $arguments,
                'version' => $version,
            ];
        }

        protected function logToolCall(
            ?ApiKey $apiKey,
            array $request,
            mixed $result,
            int $durationMs,
            bool $success,
            ?string $error = null
        ): void {
        }

        protected function dispatchWebhook(
            ?ApiKey $apiKey,
            array $request,
            bool $success,
            int $durationMs,
            ?string $error = null
        ): void {
        }

        protected function logApiRequest(
            Request $request,
            string $path,
            array $validated,
            int $status,
            array $response,
            int $durationMs,
            ?ApiKey $apiKey,
            ?string $error = null
        ): void {
        }
    };

    $request = Request::create('/api/mcp/servers/test-detail-server/tools/search', 'POST', [
        'arguments' => [
            'query' => 'status',
        ],
    ]);
    $request->attributes->set('api_key', $this->apiKey);

    $response = $controller->callToolByRoute($request, 'test-detail-server', 'search');

    expect($response->getStatusCode())->toBe(200);
    expect($response->getData(true))->toMatchArray([
        'success' => true,
        'server' => 'test-detail-server',
        'tool' => 'search',
        'result' => [
            'ok' => true,
            'server' => 'test-detail-server',
            'tool' => 'search',
            'arguments' => [
                'query' => 'status',
            ],
            'version' => null,
        ],
    ]);
    expect($controller->calls)->toHaveCount(1);
    expect($controller->calls[0])->toMatchArray([
        'server' => 'test-detail-server',
        'tool' => 'search',
        'arguments' => [
            'query' => 'status',
        ],
        'version' => null,
    ]);
});

it('McpApiController_callToolByRoute_Bad_rejects_invalid_server_ids', function () {
    $controller = new class extends McpApiController
    {
    };

    $request = Request::create('/api/mcp/servers/_bad/tools/search', 'POST', [
        'arguments' => [
            'query' => 'status',
        ],
    ]);
    $request->attributes->set('api_key', $this->apiKey);

    $response = $controller->callToolByRoute($request, '_bad', 'search');

    expect($response->getStatusCode())->toBe(422);
    expect($response->getData(true))->toMatchArray([
        'error' => 'validation_error',
        'errors' => [
            'server' => ['The selected server id is invalid.'],
        ],
    ]);
});

it('McpApiController_callToolByRoute_Ugly_rejects_invalid_tool_names', function () {
    $controller = new class extends McpApiController
    {
    };

    $request = Request::create('/api/mcp/servers/test-detail-server/tools/_bad', 'POST', [
        'arguments' => [
            'query' => 'status',
        ],
    ]);
    $request->attributes->set('api_key', $this->apiKey);

    $response = $controller->callToolByRoute($request, 'test-detail-server', '_bad');

    expect($response->getStatusCode())->toBe(422);
    expect($response->getData(true))->toMatchArray([
        'error' => 'validation_error',
        'errors' => [
            'tool' => ['The selected tool name is invalid.'],
        ],
    ]);
});

it('McpApiController_callTool_Bad_rejects_traversal_style_tool_names', function () {
    $controller = new class extends McpApiController
    {
    };

    $request = Request::create('/api/mcp/tools/call', 'POST', [
        'server' => 'test-detail-server',
        'tool' => 'foo..bar',
        'arguments' => [
            'query' => 'status',
        ],
    ]);
    $request->attributes->set('api_key', $this->apiKey);

    $response = $controller->callTool($request);

    expect($response->getStatusCode())->toBe(422);
    expect($response->getData(true))->toMatchArray([
        'error' => 'validation_error',
        'errors' => [
            'tool' => ['The selected tool name is invalid.'],
        ],
    ]);
});
