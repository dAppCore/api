<?php

declare(strict_types=1);

namespace Core\Api\Documentation\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Protect Documentation Middleware.
 *
 * Controls access to API documentation based on environment,
 * authentication, and IP whitelist.
 */
class ProtectDocumentation
{
    /**
     * Handle an incoming request.
     */
    public function handle(Request $request, Closure $next): Response
    {
        // Check if documentation is enabled
        if (! config('api-docs.enabled', true)) {
            abort(404);
        }

        $config = config('api-docs.access', []);

        // Check if public access is allowed in current environment
        $publicEnvironments = $config['public_environments'] ?? ['local', 'testing', 'staging'];
        if (in_array(app()->environment(), $publicEnvironments, true)) {
            return $next($request);
        }

        // Check IP whitelist
        $ipWhitelist = $config['ip_whitelist'] ?? [];
        if (! empty($ipWhitelist)) {
            $clientIp = $request->ip();
            if (! in_array($clientIp, $ipWhitelist, true)) {
                abort(403, 'Access denied.');
            }

            return $next($request);
        }

        // Check if authentication is required
        if ($config['require_auth'] ?? false) {
            if (! $request->user()) {
                return redirect()->route('login');
            }

            // Check allowed roles
            $allowedRoles = $config['allowed_roles'] ?? [];
            if (! empty($allowedRoles)) {
                $user = $request->user();

                // Check if user has any of the allowed roles
                $hasRole = false;
                foreach ($allowedRoles as $role) {
                    if (method_exists($user, 'hasRole') && $user->hasRole($role)) {
                        $hasRole = true;
                        break;
                    }
                }

                if (! $hasRole) {
                    abort(403, 'Insufficient permissions to view documentation.');
                }
            }
        }

        return $next($request);
    }
}
