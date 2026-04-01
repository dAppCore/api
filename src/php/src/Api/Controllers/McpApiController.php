<?php

declare(strict_types=1);

namespace Core\Api\Controllers;

use Core\Front\Controller;
use Core\Api\Concerns\HasApiResponses;
use Core\Api\Documentation\Attributes\ApiParameter;
use Core\Api\Models\ApiKey;
use Core\Mod\Mcp\Models\McpApiRequest;
use Core\Mod\Mcp\Models\McpToolCall;
use Core\Mod\Mcp\Models\McpToolVersion;
use Core\Mod\Mcp\Services\McpWebhookDispatcher;
use Core\Mod\Mcp\Services\ToolVersionService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Cache;
use Symfony\Component\Yaml\Yaml;

/**
 * MCP HTTP API Controller.
 *
 * Provides HTTP bridge to MCP servers for external integrations.
 */
class McpApiController extends Controller
{
    use HasApiResponses;

    /**
     * List all available MCP servers.
     *
     * GET /api/v1/mcp/servers
     */
    public function servers(Request $request): JsonResponse
    {
        $registry = $this->loadRegistry();

        $servers = collect($registry['servers'] ?? [])
            ->map(fn ($ref) => $this->loadServerSummary($ref['id']))
            ->filter()
            ->values();

        return response()->json([
            'servers' => $servers,
            'count' => $servers->count(),
        ]);
    }

    /**
     * Get server details with tools and resources.
     *
     * GET /api/v1/mcp/servers/{id}
     */
    public function server(Request $request, string $id): JsonResponse
    {
        $server = $this->loadServerFull($id);

        if (! $server) {
            return $this->notFoundResponse('Server');
        }

        return response()->json($server);
    }

    /**
     * List tools for a specific server.
     *
     * GET /api/v1/mcp/servers/{id}/tools
     *
     * Query params:
     * - include_versions: bool - include version info for each tool
     */
    #[ApiParameter(
        name: 'include_versions',
        in: 'query',
        type: 'boolean',
        description: 'Include version information for each tool',
        required: false,
        example: false,
        default: false
    )]
    public function tools(Request $request, string $id): JsonResponse
    {
        $server = $this->loadServerFull($id);

        if (! $server) {
            return $this->notFoundResponse('Server');
        }

        $tools = $server['tools'] ?? [];
        $includeVersions = $request->boolean('include_versions', false);

        // Optionally enrich tools with version information
        if ($includeVersions) {
            $versionService = app(ToolVersionService::class);
            $tools = collect($tools)->map(function ($tool) use ($id, $versionService) {
                $toolName = $tool['name'] ?? '';
                $latestVersion = $versionService->getLatestVersion($id, $toolName);

                $tool['versioning'] = [
                    'latest_version' => $latestVersion?->version ?? ToolVersionService::DEFAULT_VERSION,
                    'is_versioned' => $latestVersion !== null,
                    'deprecated' => $latestVersion?->is_deprecated ?? false,
                ];

                // If version exists, use its schema (may differ from YAML)
                if ($latestVersion?->input_schema) {
                    $tool['inputSchema'] = $latestVersion->input_schema;
                }

                return $tool;
            })->all();
        }

        return response()->json([
            'server' => $id,
            'tools' => $tools,
            'count' => count($tools),
        ]);
    }

    /**
     * List resources for a specific server.
     *
     * GET /api/v1/mcp/servers/{id}/resources
     *
     * Query params:
     * - include_content: bool - include resource content when the definition already contains it
     */
    #[ApiParameter(
        name: 'include_content',
        in: 'query',
        type: 'boolean',
        description: 'Include resource content when the definition already contains it',
        required: false,
        example: false,
        default: false
    )]
    public function resources(Request $request, string $id): JsonResponse
    {
        $server = $this->loadServerFull($id);

        if (! $server) {
            return $this->notFoundResponse('Server');
        }

        $includeContent = $request->boolean('include_content', false);

        $resources = collect($server['resources'] ?? [])
            ->filter(fn ($resource) => is_array($resource))
            ->map(function (array $resource) use ($includeContent) {
                $payload = array_filter([
                    'uri' => $resource['uri'] ?? null,
                    'path' => $resource['path'] ?? null,
                    'name' => $resource['name'] ?? null,
                    'description' => $resource['description'] ?? null,
                    'mime_type' => $resource['mime_type'] ?? ($resource['mimeType'] ?? null),
                ], static fn ($value) => $value !== null);

                if ($includeContent && $this->resourceDefinitionHasContent($resource)) {
                    $payload['content'] = $this->normaliseResourceContent($resource);
                }

                return $payload;
            })
            ->values();

        return response()->json([
            'server' => $id,
            'resources' => $resources,
            'count' => $resources->count(),
        ]);
    }

    /**
     * Execute a tool on an MCP server.
     *
     * POST /api/v1/mcp/tools/call
     *
     * Request body:
     * - server: string (required)
     * - tool: string (required)
     * - arguments: array (optional)
     * - version: string (optional) - semver version to use, defaults to latest
     */
    public function callTool(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'server' => 'required|string|max:64',
            'tool' => 'required|string|max:128',
            'arguments' => 'nullable|array',
            'version' => 'nullable|string|max:32',
        ]);

        $server = $this->loadServerFull($validated['server']);
        if (! $server) {
            return $this->notFoundResponse('Server');
        }

        // Verify tool exists in server definition
        $toolDef = collect($server['tools'] ?? [])->firstWhere('name', $validated['tool']);
        if (! $toolDef) {
            return $this->notFoundResponse('Tool');
        }

        // Version resolution
        $versionService = app(ToolVersionService::class);
        $versionResult = $versionService->resolveVersion(
            $validated['server'],
            $validated['tool'],
            $validated['version'] ?? null
        );

        // If version was requested but is sunset, block the call
        if ($versionResult['error']) {
            $error = $versionResult['error'];

            // Sunset versions return 410 Gone
            $status = ($error['code'] ?? '') === 'TOOL_VERSION_SUNSET' ? 410 : 400;

            return $this->errorResponse(
                errorCode: $error['code'] ?? 'VERSION_ERROR',
                message: $error['message'] ?? 'Version error',
                meta: [
                    'server' => $validated['server'],
                    'tool' => $validated['tool'],
                    'requested_version' => $validated['version'] ?? null,
                    'latest_version' => $error['latest_version'] ?? null,
                    'migration_notes' => $error['migration_notes'] ?? null,
                ],
                status: $status,
            );
        }

        /** @var McpToolVersion|null $toolVersion */
        $toolVersion = $versionResult['version'];
        $deprecationWarning = $versionResult['warning'];

        // Use versioned schema if available for validation
        $schemaForValidation = $toolVersion?->input_schema ?? $toolDef['inputSchema'] ?? null;
        if ($schemaForValidation) {
            $validationErrors = $this->validateToolArguments(
                ['inputSchema' => $schemaForValidation],
                $validated['arguments'] ?? []
            );

            if (! empty($validationErrors)) {
                return $this->errorResponse(
                    errorCode: 'VALIDATION_ERROR',
                    message: 'Validation failed',
                    meta: [
                        'validation_errors' => $validationErrors,
                        'server' => $validated['server'],
                        'tool' => $validated['tool'],
                        'version' => $toolVersion?->version ?? 'unversioned',
                    ],
                    status: 422,
                );
            }
        }

        // Get API key for logging
        $apiKey = $request->attributes->get('api_key');
        $workspace = $apiKey?->workspace;

        $startTime = microtime(true);

        try {
            // Execute the tool via artisan command
            $result = $this->executeToolViaArtisan(
                $validated['server'],
                $validated['tool'],
                $validated['arguments'] ?? [],
                $toolVersion?->version
            );

            $durationMs = (int) ((microtime(true) - $startTime) * 1000);

            // Log the call
            $this->logToolCall($apiKey, $validated, $result, $durationMs, true);

            // Dispatch webhooks
            $this->dispatchWebhook($apiKey, $validated, true, $durationMs);

            $response = [
                'success' => true,
                'server' => $validated['server'],
                'tool' => $validated['tool'],
                'version' => $toolVersion?->version ?? ToolVersionService::DEFAULT_VERSION,
                'result' => $result,
                'duration_ms' => $durationMs,
            ];

            // Include deprecation warning if applicable
            if ($deprecationWarning) {
                $response['_warnings'] = [$deprecationWarning];
            }

            // Log full request for debugging/replay
            $this->logApiRequest($request, $validated, 200, $response, $durationMs, $apiKey);

            // Build response with deprecation headers if needed
            $jsonResponse = response()->json($response);

            if ($deprecationWarning) {
                $jsonResponse->header('X-MCP-Deprecation-Warning', $deprecationWarning['message'] ?? 'Version deprecated');
                if (isset($deprecationWarning['sunset_at'])) {
                    $jsonResponse->header('X-MCP-Sunset-At', $deprecationWarning['sunset_at']);
                }
                if (isset($deprecationWarning['latest_version'])) {
                    $jsonResponse->header('X-MCP-Latest-Version', $deprecationWarning['latest_version']);
                }
            }

            return $jsonResponse;
        } catch (\Throwable $e) {
            $durationMs = (int) ((microtime(true) - $startTime) * 1000);

            $this->logToolCall($apiKey, $validated, null, $durationMs, false, $e->getMessage());

            // Dispatch webhooks (even on failure)
            $this->dispatchWebhook($apiKey, $validated, false, $durationMs, $e->getMessage());

            $response = [
                'success' => false,
                'error' => $e->getMessage(),
                'server' => $validated['server'],
                'tool' => $validated['tool'],
                'version' => $toolVersion?->version ?? ToolVersionService::DEFAULT_VERSION,
            ];

            // Log full request for debugging/replay
            $this->logApiRequest($request, $validated, 500, $response, $durationMs, $apiKey, $e->getMessage());

            return $this->errorResponse(
                errorCode: 'tool_execution_error',
                message: $e->getMessage(),
                meta: array_filter([
                    'server' => $validated['server'],
                    'tool' => $validated['tool'],
                    'version' => $toolVersion?->version ?? ToolVersionService::DEFAULT_VERSION,
                ]),
                status: 500,
            );
        }
    }

    /**
     * Validate tool arguments against a JSON schema.
     *
     * @return array<string> Validation error messages
     */
    protected function validateToolArguments(array $toolDef, array $arguments): array
    {
        $inputSchema = $toolDef['inputSchema'] ?? null;

        if (! $inputSchema || ! is_array($inputSchema)) {
            return [];
        }

        $errors = [];
        $properties = $inputSchema['properties'] ?? [];
        $required = $inputSchema['required'] ?? [];

        // Check required properties
        foreach ($required as $requiredProp) {
            if (! array_key_exists($requiredProp, $arguments)) {
                $errors[] = "Missing required argument: {$requiredProp}";
            }
        }

        // Type validation for provided arguments
        foreach ($arguments as $key => $value) {
            if (! isset($properties[$key])) {
                if (($inputSchema['additionalProperties'] ?? true) === false) {
                    $errors[] = "Unknown argument: {$key}";
                }

                continue;
            }

            $propSchema = $properties[$key];
            $expectedType = $propSchema['type'] ?? null;

            if ($expectedType && ! $this->validateType($value, $expectedType)) {
                $errors[] = "Argument '{$key}' must be of type {$expectedType}";
            }

            // Validate enum values
            if (isset($propSchema['enum']) && ! in_array($value, $propSchema['enum'], true)) {
                $allowedValues = implode(', ', $propSchema['enum']);
                $errors[] = "Argument '{$key}' must be one of: {$allowedValues}";
            }
        }

        return $errors;
    }

    /**
     * Validate a value against a JSON Schema type.
     */
    protected function validateType(mixed $value, string $type): bool
    {
        return match ($type) {
            'string' => is_string($value),
            'integer' => is_int($value) || (is_numeric($value) && floor((float) $value) == $value),
            'number' => is_numeric($value),
            'boolean' => is_bool($value),
            'array' => is_array($value) && array_is_list($value),
            'object' => is_array($value) && ! array_is_list($value),
            'null' => is_null($value),
            default => true,
        };
    }

    /**
     * Get version history for a specific tool.
     *
     * GET /api/v1/mcp/servers/{server}/tools/{tool}/versions
     */
    public function toolVersions(Request $request, string $server, string $tool): JsonResponse
    {
        $serverConfig = $this->loadServerFull($server);
        if (! $serverConfig) {
            return $this->notFoundResponse('Server');
        }

        // Verify tool exists in server definition
        $toolDef = collect($serverConfig['tools'] ?? [])->firstWhere('name', $tool);
        if (! $toolDef) {
            return $this->notFoundResponse('Tool');
        }

        $versionService = app(ToolVersionService::class);
        $versions = $versionService->getVersionHistory($server, $tool);

        return response()->json([
            'server' => $server,
            'tool' => $tool,
            'versions' => $versions->map(fn (McpToolVersion $v) => $v->toApiArray())->values(),
            'count' => $versions->count(),
        ]);
    }

    /**
     * Get a specific version of a tool.
     *
     * GET /api/v1/mcp/servers/{server}/tools/{tool}/versions/{version}
     */
    public function toolVersion(Request $request, string $server, string $tool, string $version): JsonResponse
    {
        $versionService = app(ToolVersionService::class);
        $toolVersion = $versionService->getToolAtVersion($server, $tool, $version);

        if (! $toolVersion) {
            return $this->notFoundResponse('Version');
        }

        $response = response()->json($toolVersion->toApiArray());

        // Add deprecation headers if applicable
        if ($deprecationWarning = $toolVersion->getDeprecationWarning()) {
            $response->header('X-MCP-Deprecation-Warning', $deprecationWarning['message'] ?? 'Version deprecated');
            if (isset($deprecationWarning['sunset_at'])) {
                $response->header('X-MCP-Sunset-At', $deprecationWarning['sunset_at']);
            }
        }

        return $response;
    }

    /**
     * Read a resource from an MCP server.
     *
     * GET /api/v1/mcp/resources/{uri}
     */
    public function resource(Request $request, string $uri): JsonResponse
    {
        $uri = rawurldecode($uri);

        // Parse URI format: server://resource/path
        if (! preg_match('/^([a-z0-9-]+):\/\/(.+)$/', $uri, $matches)) {
            return $this->validationErrorResponse([
                'uri' => ['Invalid resource URI format. Expected pattern server://resource/path'],
            ], 400);
        }

        $serverId = $matches[1];
        $resourcePath = $matches[2];

        $server = $this->loadServerFull($serverId);
        if (! $server) {
            return $this->notFoundResponse('Server');
        }

        $resourceDef = $this->findResourceDefinition($server, $uri, $resourcePath);
        if ($resourceDef !== null && $this->resourceDefinitionHasContent($resourceDef)) {
            return response()->json([
                'uri' => $uri,
                'server' => $serverId,
                'resource' => $resourcePath,
                'content' => $this->normaliseResourceContent($resourceDef),
            ]);
        }

        try {
            $result = $this->readResourceViaArtisan($serverId, $resourcePath);
            if ($result === null) {
                return $this->notFoundResponse('Resource');
            }

            if (is_array($result) && array_key_exists('content', $result)) {
                $content = $result['content'];
            } elseif (is_array($result) && array_key_exists('contents', $result)) {
                $content = $result['contents'];
            } else {
                $content = $result;
            }

            return response()->json([
                'uri' => $uri,
                'server' => $serverId,
                'resource' => $resourcePath,
                'content' => $content,
            ]);
        } catch (\Throwable $e) {
            return $this->errorResponse(
                errorCode: 'resource_read_error',
                message: $e->getMessage(),
                meta: [
                    'uri' => $uri,
                ],
                status: 500,
            );
        }
    }

    /**
     * Execute tool via artisan MCP server command.
     */
    protected function executeToolViaArtisan(string $server, string $tool, array $arguments, ?string $version = null): mixed
    {
        $command = $this->resolveMcpServerCommand($server);
        if (! $command) {
            throw new \RuntimeException("Unknown server: {$server}");
        }

        $mcpRequest = $this->buildToolCallRequest($tool, $arguments, $version);

        // Execute via process
        $process = proc_open(
            ['php', 'artisan', $command],
            [
                0 => ['pipe', 'r'],
                1 => ['pipe', 'w'],
                2 => ['pipe', 'w'],
            ],
            $pipes,
            base_path()
        );

        if (! is_resource($process)) {
            throw new \RuntimeException('Failed to start MCP server process');
        }

        fwrite($pipes[0], json_encode($mcpRequest)."\n");
        fclose($pipes[0]);

        $output = stream_get_contents($pipes[1]);
        fclose($pipes[1]);
        fclose($pipes[2]);

        proc_close($process);

        $response = json_decode($output, true);

        if (isset($response['error'])) {
            throw new \RuntimeException($response['error']['message'] ?? 'Tool execution failed');
        }

        return $response['result'] ?? null;
    }

    /**
     * Build the JSON-RPC payload for an MCP tool call.
     */
    protected function buildToolCallRequest(string $tool, array $arguments, ?string $version = null): array
    {
        $params = [
            'name' => $tool,
            'arguments' => $arguments,
        ];

        if ($version !== null && $version !== '') {
            $params['version'] = $version;
        }

        return [
            'jsonrpc' => '2.0',
            'id' => uniqid(),
            'method' => 'tools/call',
            'params' => $params,
        ];
    }

    /**
     * Read resource via artisan MCP server command.
     */
    protected function readResourceViaArtisan(string $server, string $path): mixed
    {
        $command = $this->resolveMcpServerCommand($server);
        if (! $command) {
            throw new \RuntimeException("Unknown server: {$server}");
        }

        $mcpRequest = [
            'jsonrpc' => '2.0',
            'id' => uniqid(),
            'method' => 'resources/read',
            'params' => [
                'uri' => "{$server}://{$path}",
                'path' => $path,
            ],
        ];

        $process = proc_open(
            ['php', 'artisan', $command],
            [
                0 => ['pipe', 'r'],
                1 => ['pipe', 'w'],
                2 => ['pipe', 'w'],
            ],
            $pipes,
            base_path()
        );

        if (! is_resource($process)) {
            throw new \RuntimeException('Failed to start MCP server process');
        }

        fwrite($pipes[0], json_encode($mcpRequest)."\n");
        fclose($pipes[0]);

        $output = stream_get_contents($pipes[1]);
        fclose($pipes[1]);
        fclose($pipes[2]);

        proc_close($process);

        $response = json_decode($output, true);
        if (! is_array($response)) {
            throw new \RuntimeException('Invalid MCP resource response');
        }

        if (isset($response['error'])) {
            throw new \RuntimeException($response['error']['message'] ?? 'Resource read failed');
        }

        return $response['result'] ?? null;
    }

    /**
     * Resolve the artisan command used for a given MCP server.
     */
    protected function resolveMcpServerCommand(string $server): ?string
    {
        $commandMap = [
            'hosthub-agent' => 'mcp:agent-server',
            'socialhost' => 'mcp:socialhost-server',
            'biohost' => 'mcp:biohost-server',
            'commerce' => 'mcp:commerce-server',
            'supporthost' => 'mcp:support-server',
            'upstream' => 'mcp:upstream-server',
        ];

        return $commandMap[$server] ?? null;
    }

    /**
     * Find a resource definition within the loaded server config.
     */
    protected function findResourceDefinition(array $server, string $uri, string $path): mixed
    {
        foreach ($server['resources'] ?? [] as $resource) {
            if (! is_array($resource)) {
                continue;
            }

            $resourceUri = $resource['uri'] ?? null;
            $resourcePath = $resource['path'] ?? null;
            $resourceName = $resource['name'] ?? null;

            if ($resourceUri === $uri || $resourcePath === $path || $resourceName === basename($path)) {
                return $resource;
            }
        }

        return null;
    }

    /**
     * Normalise a resource definition into a response payload.
     */
    protected function normaliseResourceContent(mixed $resource): mixed
    {
        if (! is_array($resource)) {
            return $resource;
        }

        foreach (['content', 'contents', 'body', 'text', 'value'] as $field) {
            if (array_key_exists($field, $resource)) {
                return $resource[$field];
            }
        }

        return $resource;
    }

    /**
     * Determine whether a resource definition already carries readable content.
     */
    protected function resourceDefinitionHasContent(mixed $resource): bool
    {
        if (! is_array($resource)) {
            return true;
        }

        foreach (['content', 'contents', 'body', 'text', 'value'] as $field) {
            if (array_key_exists($field, $resource)) {
                return true;
            }
        }

        return false;
    }

    /**
     * Log full API request for debugging and replay.
     */
    protected function logApiRequest(
        Request $request,
        array $validated,
        int $status,
        array $response,
        int $durationMs,
        ?ApiKey $apiKey,
        ?string $error = null
    ): void {
        try {
            McpApiRequest::log(
                method: $request->method(),
                path: '/tools/call',
                requestBody: $validated,
                responseStatus: $status,
                responseBody: $response,
                durationMs: $durationMs,
                workspaceId: $apiKey?->workspace_id,
                apiKeyId: $apiKey?->id,
                serverId: $validated['server'],
                toolName: $validated['tool'],
                errorMessage: $error,
                ipAddress: $request->ip(),
                headers: $request->headers->all()
            );
        } catch (\Throwable $e) {
            // Don't let logging failures affect API response
            report($e);
        }
    }

    /**
     * Dispatch webhook for tool execution.
     */
    protected function dispatchWebhook(
        ?ApiKey $apiKey,
        array $request,
        bool $success,
        int $durationMs,
        ?string $error = null
    ): void {
        if (! $apiKey?->workspace_id) {
            return;
        }

        try {
            $dispatcher = new McpWebhookDispatcher;
            $dispatcher->dispatchToolExecuted(
                workspaceId: $apiKey->workspace_id,
                serverId: $request['server'],
                toolName: $request['tool'],
                arguments: $request['arguments'] ?? [],
                success: $success,
                durationMs: $durationMs,
                errorMessage: $error
            );
        } catch (\Throwable $e) {
            // Don't let webhook failures affect API response
            report($e);
        }
    }

    /**
     * Log tool call for analytics.
     */
    protected function logToolCall(
        ?ApiKey $apiKey,
        array $request,
        mixed $result,
        int $durationMs,
        bool $success,
        ?string $error = null
    ): void {
        McpToolCall::log(
            serverId: $request['server'],
            toolName: $request['tool'],
            params: $request['arguments'] ?? [],
            success: $success,
            durationMs: $durationMs,
            errorMessage: $error,
            workspaceId: $apiKey?->workspace_id
        );
    }

    // Registry loading methods (shared with McpRegistryController)

    protected function loadRegistry(): array
    {
        return Cache::remember('mcp:registry', 600, function () {
            $path = resource_path('mcp/registry.yaml');

            return file_exists($path) ? Yaml::parseFile($path) : ['servers' => []];
        });
    }

    protected function loadServerFull(string $id): ?array
    {
        return Cache::remember("mcp:server:{$id}", 600, function () use ($id) {
            $path = resource_path("mcp/servers/{$id}.yaml");

            return file_exists($path) ? Yaml::parseFile($path) : null;
        });
    }

    protected function loadServerSummary(string $id): ?array
    {
        $server = $this->loadServerFull($id);
        if (! $server) {
            return null;
        }

        return [
            'id' => $server['id'],
            'name' => $server['name'],
            'tagline' => $server['tagline'] ?? '',
            'status' => $server['status'] ?? 'available',
            'tool_count' => count($server['tools'] ?? []),
            'resource_count' => count($server['resources'] ?? []),
        ];
    }
}
