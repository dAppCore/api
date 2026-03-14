<?php

declare(strict_types=1);

namespace Core\Api\Documentation;

use Core\Api\Documentation\Attributes\ApiTag;
use Illuminate\Support\Facades\Route;
use ReflectionClass;

/**
 * Module Discovery Service.
 *
 * Discovers API routes from modules and groups them by tag/module
 * for organized documentation.
 */
class ModuleDiscovery
{
    /**
     * Discovered modules with their routes.
     *
     * @var array<string, array>
     */
    protected array $modules = [];

    /**
     * Discover all API modules and their routes.
     *
     * @return array<string, array>
     */
    public function discover(): array
    {
        $this->modules = [];

        foreach (Route::getRoutes() as $route) {
            if (! $this->isApiRoute($route)) {
                continue;
            }

            $module = $this->identifyModule($route);
            $this->addRouteToModule($module, $route);
        }

        ksort($this->modules);

        return $this->modules;
    }

    /**
     * Get modules grouped by tag.
     *
     * @return array<string, array>
     */
    public function getModulesByTag(): array
    {
        $byTag = [];

        foreach ($this->discover() as $module => $data) {
            $tag = $data['tag'] ?? $module;
            $byTag[$tag] = $byTag[$tag] ?? [
                'name' => $tag,
                'description' => $data['description'] ?? null,
                'routes' => [],
            ];

            $byTag[$tag]['routes'] = array_merge(
                $byTag[$tag]['routes'],
                $data['routes']
            );
        }

        return $byTag;
    }

    /**
     * Get a summary of discovered modules.
     */
    public function getSummary(): array
    {
        $modules = $this->discover();

        return array_map(function ($data) {
            return [
                'tag' => $data['tag'],
                'description' => $data['description'],
                'route_count' => count($data['routes']),
                'endpoints' => array_map(function ($route) {
                    return [
                        'method' => $route['method'],
                        'uri' => $route['uri'],
                        'name' => $route['name'],
                    ];
                }, $data['routes']),
            ];
        }, $modules);
    }

    /**
     * Check if route is an API route.
     */
    protected function isApiRoute($route): bool
    {
        $uri = $route->uri();

        return str_starts_with($uri, 'api/') || $uri === 'api';
    }

    /**
     * Identify which module a route belongs to.
     */
    protected function identifyModule($route): string
    {
        $controller = $route->getController();

        if ($controller !== null) {
            // Check for ApiTag attribute
            $reflection = new ReflectionClass($controller);
            $tagAttrs = $reflection->getAttributes(ApiTag::class);

            if (! empty($tagAttrs)) {
                return $tagAttrs[0]->newInstance()->name;
            }

            // Infer from namespace
            $namespace = $reflection->getNamespaceName();

            // Extract module name from namespace patterns
            if (preg_match('/(?:Mod|Module|Http\\\\Controllers)\\\\([^\\\\]+)/', $namespace, $matches)) {
                return $matches[1];
            }
        }

        // Infer from route URI
        return $this->inferModuleFromUri($route->uri());
    }

    /**
     * Infer module name from URI.
     */
    protected function inferModuleFromUri(string $uri): string
    {
        // Remove api/ prefix
        $path = preg_replace('#^api/#', '', $uri);

        // Get first segment
        $parts = explode('/', $path);
        $segment = $parts[0] ?? 'general';

        // Map common segments to module names
        $mapping = [
            'bio' => 'Bio',
            'blocks' => 'Bio',
            'shortlinks' => 'Bio',
            'qr' => 'Bio',
            'commerce' => 'Commerce',
            'provisioning' => 'Commerce',
            'workspaces' => 'Tenant',
            'analytics' => 'Analytics',
            'social' => 'Social',
            'notify' => 'Notifications',
            'support' => 'Support',
            'pixel' => 'Pixel',
            'seo' => 'SEO',
            'mcp' => 'MCP',
            'content' => 'Content',
            'trust' => 'Trust',
            'webhooks' => 'Webhooks',
            'entitlements' => 'Entitlements',
        ];

        return $mapping[$segment] ?? ucfirst($segment);
    }

    /**
     * Add a route to a module.
     */
    protected function addRouteToModule(string $module, $route): void
    {
        if (! isset($this->modules[$module])) {
            $this->modules[$module] = [
                'tag' => $module,
                'description' => $this->getModuleDescription($module),
                'routes' => [],
            ];
        }

        $methods = array_filter($route->methods(), fn ($m) => $m !== 'HEAD');

        foreach ($methods as $method) {
            $this->modules[$module]['routes'][] = [
                'method' => strtoupper($method),
                'uri' => '/'.$route->uri(),
                'name' => $route->getName(),
                'action' => $route->getActionMethod(),
                'middleware' => $route->middleware(),
            ];
        }
    }

    /**
     * Get module description from config.
     */
    protected function getModuleDescription(string $module): ?string
    {
        $tags = config('api-docs.tags', []);

        return $tags[$module]['description'] ?? null;
    }
}
