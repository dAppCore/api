<?php

declare(strict_types=1);

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
        __DIR__.'/../src/php/tests/Feature',
        __DIR__.'/../src/php/src/Api/Tests/Feature',
    );
