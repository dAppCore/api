## Audit: Fail-open IDOR Pattern

Status: Complete

Scope: `src/php/src/Api/Controllers/Api/*.php`

Pattern audited:

```php
$query = Model::query();
if ($workspace !== null) { $query->forWorkspace(...); }
if ($user?->id !== null) { $query->where('user_id', ...); }
return $query->find($id);
```

### ApiKeyController.php

CLEAN. `destroy()` requires `resolveWorkspace()` to succeed before `ApiKey::query()->forWorkspace($workspace->id)->find($id)`.

### AuthController.php

CLEAN. No ID-based resource lookup with conditional workspace/user scoping; token revocation operates on the already-authenticated request attribute.

### BiolinkController.php

CLEAN. `findBiolink()` returns `null` when `resolveWorkspace()` fails, then performs `Biolink::query()->forWorkspace($workspace->id)->find($id)`.

### EntitlementApiController.php

CLEAN. No ID-based resource lookup; all workspace-specific methods require a resolved `Workspace`.

### LinkController.php

CLEAN. `findLink()` returns `null` when `resolveWorkspace()` fails, then performs `Link::query()->forWorkspace($workspace->id)->find($id)`.

### PaymentMethodController.php

CLEAN. `destroy()` and `default()` require `resolveWorkspace()` to succeed before querying `PaymentMethod::query()->where('workspace_id', $workspace->id)->find($id)`.

### QrCodeController.php

CLEAN. `findCode()` returns `null` when `resolveWorkspace()` fails, then performs `QrCode::query()->forWorkspace($workspace->id)->find($id)`.

### SeoReportController.php

CLEAN. No ID-based resource lookup; delegates URL analysis to `SeoReportService`.

### TicketController.php

VULNERABLE. `index()` fails closed when both workspace and user are absent, but `findTicket()` does not. `findTicket()` builds `SupportTicket::query()->with('replies')`, conditionally applies `forWorkspace()` and `where('user_id', ...)`, then calls `find($id)`, leaving an unscoped fallback when both context values are absent.

### UnifiedPixelController.php

CLEAN. No model lookup; returns a static tracking response.

### WebhookController.php

CLEAN. `resolveWebhook()` returns `null` when `resolveWorkspace()` fails, then performs `WebhookEndpoint::query()->forWorkspace($workspace->id)->find($id)`.

### WebhookSecretController.php

CLEAN. Secret operations require `defaultHostWorkspace()` before lookup and use mandatory `workspace_id` plus `uuid` filters with `first()`.

### WebhookTemplateController.php

CLEAN. Template UUID operations require `defaultHostWorkspace()` before lookup and use mandatory `workspace_id` plus `uuid` filters with `first()`. Validation, preview, variable, filter, and builtin endpoints do not fetch a persisted resource by caller-supplied ID.

### WorkspaceMemberController.php

CLEAN. `destroy()` requires `resolveWorkspace()` to succeed before querying `WorkspaceMember::query()->forWorkspace($workspace)->forUser((int) $user)->first()`.

## Final Classification

| Controller | Method | Status | Notes |
| --- | --- | --- | --- |
| ApiKeyController | `destroy` | CLEAN | Workspace resolution is mandatory before scoped `find($id)`. |
| AuthController | `store`, `destroy`, `show` | CLEAN | No conditional workspace/user scoped ID lookup; authenticated token/key is sourced from request context. |
| BiolinkController | `findBiolink` via `show`, `update`, `destroy` | CLEAN | Fails closed on missing workspace before `forWorkspace(...)->find($id)`. |
| EntitlementApiController | `show`, `check`, `usage` | CLEAN | Requires resolved `Workspace`; no caller-supplied ID lookup. |
| LinkController | `findLink` via `show`, `update`, `destroy`, `stats` | CLEAN | Fails closed on missing workspace before `forWorkspace(...)->find($id)`. |
| PaymentMethodController | `destroy`, `default` | CLEAN | Requires resolved workspace before `where('workspace_id', ...)->find($id)`. |
| QrCodeController | `findCode` via `show`, `download` | CLEAN | Fails closed on missing workspace before `forWorkspace(...)->find($id)`. |
| SeoReportController | `show` | CLEAN | No persisted resource ID lookup. |
| TicketController | `findTicket` via `show`, `reply` | VULNERABLE | Conditional workspace/user filters can both be skipped before `SupportTicket` `find($id)`. |
| UnifiedPixelController | `track` | CLEAN | No persisted resource ID lookup. |
| WebhookController | `resolveWebhook` via `show`, `update`, `destroy`, `deliveries` | CLEAN | Fails closed on missing workspace before `forWorkspace(...)->find($id)`. |
| WebhookSecretController | all secret rotation/status/grace-period methods | CLEAN | Requires `defaultHostWorkspace()` and mandatory `workspace_id`/`uuid` filters. |
| WebhookTemplateController | UUID-backed template methods | CLEAN | Requires `defaultHostWorkspace()` and mandatory `workspace_id`/`uuid` filters. |
| WorkspaceMemberController | `destroy` | CLEAN | Requires resolved workspace before `forWorkspace(...)->forUser(...)->first()`. |

## Recommended Mantis Tickets To File

- Fix `TicketController::findTicket()` to fail closed when both workspace and authenticated user context are absent before calling `SupportTicket::query()->find($id)`, and add regression coverage for `show`/`reply` requests without either context.
