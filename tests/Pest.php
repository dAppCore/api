<?php

declare(strict_types=1);

if (! class_exists(\Dedoc\Scramble\Scramble::class)) {
    eval(<<<'PHP'
namespace Dedoc\Scramble;

final class Scramble
{
    public static function ignoreDefaultRoutes(): void
    {
    }
}
PHP);
}

if (! class_exists(\Dedoc\Scramble\ScrambleServiceProvider::class)) {
    eval(<<<'PHP'
namespace Dedoc\Scramble;

final class ScrambleServiceProvider extends \Illuminate\Support\ServiceProvider
{
    public function register(): void
    {
    }
}
PHP);
}

if (! class_exists(\Core\Mcp\Middleware\McpApiKeyAuth::class)) {
    class_alias(\Core\Api\Middleware\AuthenticateApiKey::class, \Core\Mcp\Middleware\McpApiKeyAuth::class);
}

if (! class_exists(\Core\Mod\Mcp\Services\ToolVersionService::class)) {
    eval(<<<'PHP'
namespace Core\Mod\Mcp\Services;

final class ToolVersionService
{
    public const DEFAULT_VERSION = 'latest';

    public function getLatestVersion(string $serverId, string $toolName): ?object
    {
        return null;
    }

    public function resolveVersion(string $server, string $tool, ?string $version): array
    {
        return [
            'version' => null,
            'warning' => null,
            'error' => null,
        ];
    }

    public function getVersionHistory(string $server, string $tool): \Illuminate\Support\Collection
    {
        return collect();
    }

    public function getToolAtVersion(string $server, string $tool, string $version): ?object
    {
        return null;
    }
}
PHP);
}

$_SERVER['argv'][1] = $_SERVER['argv'][1] ?? 'route:list';

abstract class TestCase extends Orchestra\Testbench\TestCase
{
    protected function getPackageProviders($app): array
    {
        return [
            Core\Api\Boot::class,
        ];
    }
}

uses(TestCase::class)
    ->in(
        __DIR__.'/../php/tests/Feature',
        __DIR__.'/../php/src/Api/Tests/Feature',
    );
