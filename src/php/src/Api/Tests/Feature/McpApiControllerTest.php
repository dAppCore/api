<?php

declare(strict_types=1);

use Core\Api\Controllers\McpApiController;

it('includes the requested tool version in the MCP JSON-RPC payload', function () {
    $controller = new class extends McpApiController
    {
        public function payload(string $tool, array $arguments, ?string $version = null): array
        {
            return $this->buildToolCallRequest($tool, $arguments, $version);
        }
    };

    $payload = $controller->payload('search', ['query' => 'status'], '1.2.3');

    expect($payload['jsonrpc'])->toBe('2.0');
    expect($payload['method'])->toBe('tools/call');
    expect($payload['params'])->toMatchArray([
        'name' => 'search',
        'arguments' => ['query' => 'status'],
        'version' => '1.2.3',
    ]);
});

it('omits the version field when one is not requested', function () {
    $controller = new class extends McpApiController
    {
        public function payload(string $tool, array $arguments, ?string $version = null): array
        {
            return $this->buildToolCallRequest($tool, $arguments, $version);
        }
    };

    $payload = $controller->payload('search', ['query' => 'status']);

    expect($payload['params'])->toMatchArray([
        'name' => 'search',
        'arguments' => ['query' => 'status'],
    ]);
    expect($payload['params'])->not->toHaveKey('version');
});

it('rejects malformed MCP payloads before writing to a subprocess', function () {
    $controller = new class extends McpApiController
    {
        public function encode(array $payload): string
        {
            return $this->encodeMcpRequest($payload, 'MCP tool call');
        }
    };

    expect(fn () => $controller->encode([
        'jsonrpc' => '2.0',
        'id' => 'test',
        'method' => 'tools/call',
        'params' => [
            'name' => 'search',
            'arguments' => [
                'query' => "\xB1",
            ],
        ],
    ]))->toThrow(RuntimeException::class, 'Unable to encode MCP tool call request as JSON.');
});

it('rejects malformed tool responses from the subprocess bridge', function () {
    $controller = new class extends McpApiController
    {
        public function call(string $server, string $tool, array $arguments = []): mixed
        {
            return $this->executeToolViaArtisan($server, $tool, $arguments, null);
        }

        protected function runMcpServerCommand(string $command, string $payload, string $context): string
        {
            return 'not-json';
        }
    };

    expect(fn () => $controller->call('hosthub-agent', 'search'))
        ->toThrow(RuntimeException::class, 'Invalid MCP tool response');
});
