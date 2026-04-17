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

$_SERVER['argv'][1] = $_SERVER['argv'][1] ?? 'route:list';

abstract class TestCase extends Orchestra\Testbench\TestCase
{
    protected function getPackageProviders(\Illuminate\Contracts\Foundation\Application $app): array
    {
        return [
            Core\Api\Boot::class,
        ];
    }
}

uses(TestCase::class)
    ->in(
        __DIR__.'/../src/php/tests/Feature',
        __DIR__.'/../src/php/src/Api/Tests/Feature',
    );
