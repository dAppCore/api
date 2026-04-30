<?php

declare(strict_types=1);

namespace Core\Api\Documentation;

use Illuminate\Routing\Route;

/**
 * OpenAPI Extension Interface.
 *
 * Extensions allow customizing the OpenAPI specification generation
 * by modifying the spec or individual operations.
 */
interface Extension
{
    /**
     * Extend the complete OpenAPI specification.
     *
     * Called after the spec is built but before it's cached or returned.
     *
     * @param  array  $spec  The OpenAPI specification array
     * @param  array  $config  Documentation configuration
     * @return array Modified specification
     */
    public function extend(array $spec, array $config): array;

    /**
     * Extend an individual operation.
     *
     * Called for each route operation during path building.
     *
     * @param  array  $operation  The operation array
     * @param  Route  $route  The Laravel route
     * @param  string  $method  HTTP method (lowercase)
     * @param  array  $config  Documentation configuration
     * @return array Modified operation
     */
    public function extendOperation(array $operation, Route $route, string $method, array $config): array;
}
