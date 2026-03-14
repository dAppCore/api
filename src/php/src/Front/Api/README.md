# API Versioning

Core PHP Framework provides built-in API versioning support with deprecation handling and sunset headers.

## Quick Start

### 1. Configure Versions

Add to your `config/api.php`:

```php
'versioning' => [
    'default' => 1,           // Version when none specified
    'current' => 2,           // Latest/current version
    'supported' => [1, 2],    // All supported versions
    'deprecated' => [1],      // Deprecated but still working
    'sunset' => [             // Removal dates
        1 => '2025-12-31',
    ],
],
```

### 2. Apply Middleware

The `api.version` middleware is automatically available. Apply it to routes:

```php
// Version-agnostic routes (uses default version)
Route::middleware('api.version')->group(function () {
    Route::get('/status', StatusController::class);
});

// Version-specific routes with URL prefix
use Core\Front\Api\VersionedRoutes;

VersionedRoutes::v1(function () {
    Route::get('/users', [UserController::class, 'indexV1']);
});

VersionedRoutes::v2(function () {
    Route::get('/users', [UserController::class, 'indexV2']);
});
```

### 3. Version Negotiation in Controllers

```php
use Core\Front\Api\ApiVersionService;

class UserController
{
    public function __construct(
        protected ApiVersionService $versions
    ) {}

    public function index(Request $request)
    {
        return $this->versions->negotiate($request, [
            1 => fn() => $this->indexV1(),
            2 => fn() => $this->indexV2(),
        ]);
    }
}
```

## Version Resolution

The middleware resolves the API version from (in priority order):

1. **URL Path**: `/api/v1/users` or `/v2/users`
2. **Accept-Version Header**: `Accept-Version: v1` or `Accept-Version: 2`
3. **Accept Header**: `Accept: application/vnd.hosthub.v1+json`
4. **Default**: Falls back to configured default version

## Response Headers

Successful responses include:

```
X-API-Version: 2
```

Deprecated versions also include:

```
Deprecation: true
X-API-Warn: API version 1 is deprecated. Please upgrade to v2.
Sunset: Wed, 31 Dec 2025 00:00:00 GMT
```

## Error Responses

### Unsupported Version (400)

```json
{
  "error": "unsupported_api_version",
  "message": "API version 99 is not supported.",
  "requested_version": 99,
  "supported_versions": [1, 2],
  "current_version": 2,
  "hint": "Use Accept-Version header or URL prefix (e.g., /api/v1/) to specify version."
}
```

### Version Too Low (400)

```json
{
  "error": "api_version_too_low",
  "message": "This endpoint requires API version 2 or higher.",
  "requested_version": 1,
  "minimum_version": 2
}
```

## Versioned Routes Helper

The `VersionedRoutes` class provides a fluent API for registering version-specific routes:

```php
use Core\Front\Api\VersionedRoutes;

// Simple version registration
VersionedRoutes::v1(function () {
    Route::get('/users', UserController::class);
});

// With URL prefix (default)
VersionedRoutes::v2(function () {
    Route::get('/users', UserControllerV2::class);
}); // Accessible at /api/v2/users

// Header-only versioning (no URL prefix)
VersionedRoutes::version(2)
    ->withoutPrefix()
    ->routes(function () {
        Route::get('/users', UserControllerV2::class);
    }); // Accessible at /api/users with Accept-Version: 2

// Multiple versions for the same routes
VersionedRoutes::versions([1, 2], function () {
    Route::get('/health', HealthController::class);
});

// Deprecated version with sunset
VersionedRoutes::v1()
    ->deprecated('2025-06-01')
    ->routes(function () {
        Route::get('/legacy', LegacyController::class);
    });
```

## ApiVersionService

Inject `ApiVersionService` for programmatic version checks:

```php
use Core\Front\Api\ApiVersionService;

class UserController
{
    public function __construct(
        protected ApiVersionService $versions
    ) {}

    public function show(Request $request, User $user)
    {
        $data = $user->toArray();

        // Version-specific transformations
        return $this->versions->transform($request, $data, [
            1 => fn($d) => Arr::except($d, ['created_at', 'metadata']),
            2 => fn($d) => $d,
        ]);
    }
}
```

### Available Methods

| Method | Description |
|--------|-------------|
| `current($request)` | Get version number (e.g., 1, 2) |
| `currentString($request)` | Get version string (e.g., 'v1') |
| `is($version, $request)` | Check exact version |
| `isV1($request)` | Check if version 1 |
| `isV2($request)` | Check if version 2 |
| `isAtLeast($version, $request)` | Check minimum version |
| `isDeprecated($request)` | Check if version is deprecated |
| `defaultVersion()` | Get configured default |
| `latestVersion()` | Get current/latest version |
| `supportedVersions()` | Get all supported versions |
| `deprecatedVersions()` | Get deprecated versions |
| `sunsetDates()` | Get sunset dates map |
| `isSupported($version)` | Check if version is supported |
| `negotiate($request, $handlers)` | Call version-specific handler |
| `transform($request, $data, $transformers)` | Transform data per version |

## Sunset Middleware

For endpoint-specific deprecation, use the `api.sunset` middleware:

```php
Route::middleware('api.sunset:2025-06-01')->group(function () {
    Route::get('/legacy-endpoint', LegacyController::class);
});

// With replacement hint
Route::middleware('api.sunset:2025-06-01,/api/v2/new-endpoint')->group(function () {
    Route::get('/old-endpoint', OldController::class);
});
```

Adds headers:

```
Sunset: Sun, 01 Jun 2025 00:00:00 GMT
Deprecation: true
X-API-Warn: This endpoint is deprecated and will be removed on 2025-06-01.
Link: </api/v2/new-endpoint>; rel="successor-version"
```

## Versioning Strategy

### Guidelines

1. **Add, don't remove**: New fields can be added to any version
2. **New version for breaking changes**: Removing/renaming fields requires new version
3. **Deprecate before removal**: Give clients time to migrate
4. **Document changes**: Maintain changelog per version

### Version Lifecycle

```
v1: Active -> Deprecated (with sunset) -> Removed from supported
v2: Active (current)
v3: Future
```

### Environment Variables

```env
API_VERSION_DEFAULT=1
API_VERSION_CURRENT=2
API_VERSIONS_SUPPORTED=1,2
API_VERSIONS_DEPRECATED=1
```

## Testing

Test versioned endpoints by setting the Accept-Version header:

```php
$response = $this->withHeaders([
    'Accept-Version' => 'v2',
])->getJson('/api/users');

$response->assertHeader('X-API-Version', '2');
```

Or use URL prefix:

```php
$response = $this->getJson('/api/v2/users');
```
