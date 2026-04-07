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
