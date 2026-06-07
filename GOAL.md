# Sonar sweeps — core-api findings

707 findings across 26 rules. One rule per commit; fix every line listed under each rule.

## BLOCKER

### php:S2068 — Credentials should not be hard-coded (2×, vulnerability)

- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:152` — Detected URI with password, review this potentially hardcoded credential.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:310` — Detected URI with password, review this potentially hardcoded credential.

### php:S6418 — Secrets should not be hard-coded (1×, vulnerability)

- `src/php/src/Api/Documentation/Examples/CommonExamples.php:169` — 'API-Key' detected in this expression, review this potentially hard-coded secret.

## CRITICAL

### go:S1192 — String literals should not be duplicated (371×, code smell)

- `api_describable_test.go:126` — Define a constant instead of duplicating this literal "/api/widgets" 4 times.
- `api_describable_test.go:150` — Define a constant instead of duplicating this literal "expected tags array, got %T" 3 times.
- `api_renderable_test.go:72` — Define a constant instead of duplicating this literal "/api/widgets" 4 times.
- `api_renderable_test.go:107` — Define a constant instead of duplicating this literal "x-render-hints" 6 times.
- `api_test.go:24` — Define a constant instead of duplicating this literal "health-extra" 3 times.
- `api_test.go:132` — Define a constant instead of duplicating this literal "expected 200, got %d" 3 times.
- `api_test.go:137` — Define a constant instead of duplicating this literal "unmarshal error: %v" 3 times.
- `authentik_integration_test.go:149` — Define a constant instead of duplicating this literal "/v1/whoami" 4 times.
- `authentik_test.go:21` — Define a constant instead of duplicating this literal "alice@example.com" 4 times.
- `authentik_test.go:22` — Define a constant instead of duplicating this literal "Alice Smith" 3 times.
- `authentik_test.go:23` — Define a constant instead of duplicating this literal "abc-123" 3 times.
- `authentik_test.go:26` — Define a constant instead of duplicating this literal "tok.en.here" 3 times.
- `authentik_test.go:30` — Define a constant instead of duplicating this literal "expected Username=%q, got %q" 3 times.
- `authentik_test.go:33` — Define a constant instead of duplicating this literal "expected Email=%q, got %q" 3 times.
- `authentik_test.go:75` — Define a constant instead of duplicating this literal "https://auth.example.com" 3 times.
- `authentik_test.go:76` — Define a constant instead of duplicating this literal "my-client" 3 times.
- `authentik_test.go:101` — Define a constant instead of duplicating this literal "unexpected error: %v" 3 times.
- `authentik_test.go:147` — Define a constant instead of duplicating this literal "/v1/check" 6 times.
- `authentik_test.go:148` — Define a constant instead of duplicating this literal "X-authentik-username" 7 times.
- `authentik_test.go:149` — Define a constant instead of duplicating this literal "bob@example.com" 3 times.
- `authentik_test.go:149` — Define a constant instead of duplicating this literal "X-authentik-email" 4 times.
- `authentik_test.go:150` — Define a constant instead of duplicating this literal "Bob Jones" 3 times.
- `authentik_test.go:151` — Define a constant instead of duplicating this literal "uid-456" 3 times.
- `authentik_test.go:152` — Define a constant instead of duplicating this literal "jwt.tok.en" 3 times.
- `authentik_test.go:153` — Define a constant instead of duplicating this literal "X-authentik-groups" 4 times.
- `authentik_test.go:158` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `authentik_test.go:359` — Define a constant instead of duplicating this literal "carol@example.com" 3 times.
- `authentik_test.go:420` — Define a constant instead of duplicating this literal "/v1/protected/data" 3 times.
- `authz_test.go:67` — Define a constant instead of duplicating this literal "/stub/*" 5 times.
- `authz_test.go:75` — Define a constant instead of duplicating this literal "/stub/ping" 6 times.
- `bridge.go:389` — Define a constant instead of duplicating this literal "ToolBridge.Validate" 3 times.
- `bridge.go:420` — Define a constant instead of duplicating this literal "ToolBridge.ValidateResponse" 4 times.
- `bridge.go:467` — Define a constant instead of duplicating this literal "ToolBridge.ValidateSchema" 18 times.
- `bridge_test.go:24` — Define a constant instead of duplicating this literal "/tools" 32 times.
- `bridge_test.go:45` — Define a constant instead of duplicating this literal "/tools/file_read" 8 times.
- `bridge_test.go:53` — Define a constant instead of duplicating this literal "unmarshal error: %v" 23 times.
- `bridge_test.go:56` — Define a constant instead of duplicating this literal "expected Data=%q, got %q" 3 times.
- `bridge_test.go:77` — Define a constant instead of duplicating this literal "/api/v1/tools" 5 times.
- `bridge_test.go:252` — Define a constant instead of duplicating this literal "Read a file from disk" 12 times.
- `bridge_test.go:378` — Define a constant instead of duplicating this literal "expected 200, got %d" 8 times.
- `bridge_test.go:385` — Define a constant instead of duplicating this literal "/tmp/file.txt" 3 times.
- `bridge_test.go:426` — Define a constant instead of duplicating this literal "expected Success=true" 4 times.
- `bridge_test.go:469` — Define a constant instead of duplicating this literal "expected Success=false" 11 times.
- `bridge_test.go:493` — Define a constant instead of duplicating this literal "should not run" 9 times.
- `bridge_test.go:504` — Define a constant instead of duplicating this literal "expected 400, got %d" 8 times.
- `bridge_test.go:515` — Define a constant instead of duplicating this literal "expected invalid_request_body error, got %#v" 13 times.
- `bridge_test.go:646` — Define a constant instead of duplicating this literal "Publish an item" 3 times.
- `bridge_test.go:666` — Define a constant instead of duplicating this literal "/tools/publish_item" 3 times.
- `bridge_test.go:738` — Define a constant instead of duplicating this literal "^[A-Z]+$" 3 times.
- `bridge_test.go:1015` — Define a constant instead of duplicating this literal "/v1/tools" 4 times.
- `bridge_test.go:1135` — Define a constant instead of duplicating this literal "Validate array input" 3 times.
- `bridge_test.go:1154` — Define a constant instead of duplicating this literal "/tools/tags" 3 times.
- `bridge_test.go:1259` — Define a constant instead of duplicating this literal "Validate numeric input" 3 times.
- `bridge_test.go:1277` — Define a constant instead of duplicating this literal "/tools/score" 3 times.
- `brotli.go:59` — Define a constant instead of duplicating this literal "Content-Encoding" 3 times.
- `brotli.go:75` — Define a constant instead of duplicating this literal "Content-Length" 3 times.
- `brotli_test.go:24` — Define a constant instead of duplicating this literal "/stub/ping" 5 times.
- `brotli_test.go:25` — Define a constant instead of duplicating this literal "Accept-Encoding" 4 times.
- `brotli_test.go:29` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `brotli_test.go:32` — Define a constant instead of duplicating this literal "Content-Encoding" 5 times.
- `cache.go:240` — Define a constant instead of duplicating this literal "X-Request-ID" 3 times.
- `cache_control_test.go:27` — Define a constant instead of duplicating this literal "/items/{id}" 4 times.
- `cache_control_test.go:28` — Define a constant instead of duplicating this literal "public, max-age=60" 9 times.
- `cache_control_test.go:39` — Define a constant instead of duplicating this literal "GET /v1/items/:id" 5 times.
- `cache_control_test.go:123` — Define a constant instead of duplicating this literal "/v1/items/:id" 4 times.
- `cache_control_test.go:128` — Define a constant instead of duplicating this literal "/v1/items/123" 4 times.
- `cache_control_test.go:131` — Define a constant instead of duplicating this literal "Cache-Control" 6 times.
- `cache_test.go:27` — Define a constant instead of duplicating this literal "/cache" 3 times.
- `cache_test.go:72` — Define a constant instead of duplicating this literal "/cache/counter" 17 times.
- `cache_test.go:76` — Define a constant instead of duplicating this literal "expected 200, got %d" 11 times.
- `cache_test.go:80` — Define a constant instead of duplicating this literal "call-1" 12 times.
- `cache_test.go:81` — Define a constant instead of duplicating this literal "expected body to contain %q, got %q" 5 times.
- `cache_test.go:98` — Define a constant instead of duplicating this literal "X-Cache" 6 times.
- `cache_test.go:100` — Define a constant instead of duplicating this literal "expected X-Cache=HIT, got %q" 3 times.
- `cache_test.go:158` — Define a constant instead of duplicating this literal "unmarshal error: %v" 5 times.
- `cache_test.go:179` — Define a constant instead of duplicating this literal "expected counter=2, got %d" 3 times.
- `cache_test.go:207` — Define a constant instead of duplicating this literal "other-2" 4 times.
- `cache_test.go:252` — Define a constant instead of duplicating this literal "X-Request-ID" 8 times.
- `cache_test.go:277` — Define a constant instead of duplicating this literal "first-request-id" 6 times.
- `cache_test.go:288` — Define a constant instead of duplicating this literal "second-request-id" 10 times.
- `chat_completions.go:348` — Define a constant instead of duplicating this literal "models.yaml" 3 times.
- `chat_completions.go:737` — Define a constant instead of duplicating this literal "chat.completion.chunk" 3 times.
- `chat_completions.go:751` — Define a constant instead of duplicating this literal "data: %s\n\n" 3 times.
- `chat_completions_internal_test.go:76` — Define a constant instead of duplicating this literal "unexpected error: %v" 9 times.
- `chat_completions_internal_test.go:203` — Define a constant instead of duplicating this literal "<|channel>thought planning... " 3 times.
- `chat_completions_internal_test.go:214` — Define a constant instead of duplicating this literal " planning... " 3 times.
- `chat_completions_internal_test.go:278` — Define a constant instead of duplicating this literal "Content-Type" 3 times.
- `chat_completions_internal_test.go:297` — Define a constant instead of duplicating this literal "expected %s, got %s" 3 times.
- `chat_completions_internal_test.go:380` — Define a constant instead of duplicating this literal "expected %q, got %q" 3 times.
- `chat_completions_internal_test.go:385` — Define a constant instead of duplicating this literal "hello world" 4 times.
- `chat_completions_test.go:32` — Define a constant instead of duplicating this literal "unexpected error: %v" 3 times.
- `chat_completions_test.go:35` — Define a constant instead of duplicating this literal "/v1/chat/completions" 4 times.
- `chat_completions_test.go:39` — Define a constant instead of duplicating this literal "Content-Type" 4 times.
- `chat_completions_test.go:39` — Define a constant instead of duplicating this literal "application/json" 4 times.
- `client.go:301` — Define a constant instead of duplicating this literal "OpenAPIClient.Call" 4 times.
- `client.go:335` — Define a constant instead of duplicating this literal "application/json" 3 times.
- `client.go:411` — Define a constant instead of duplicating this literal "OpenAPIClient.loadSpec" 4 times.
- `client.go:505` — Define a constant instead of duplicating this literal "OpenAPIClient.buildURL" 3 times.
- `client.go:1026` — Define a constant instead of duplicating this literal "OpenAPIClient.validateOpenAPISchema" 3 times.
- `client.go:1045` — Define a constant instead of duplicating this literal "OpenAPIClient.validateOpenAPIResponse" 3 times.
- `client_test.go:53` — Define a constant instead of duplicating this literal "/hello" 3 times.
- `client_test.go:55` — Define a constant instead of duplicating this literal "expected GET, got %s" 5 times.
- `client_test.go:64` — Define a constant instead of duplicating this literal "Content-Type" 13 times.
- `client_test.go:64` — Define a constant instead of duplicating this literal "application/json" 13 times.
- `client_test.go:113` — Define a constant instead of duplicating this literal "unexpected error: %v" 12 times.
- `client_test.go:123` — Define a constant instead of duplicating this literal "expected map result, got %T" 7 times.
- `client_test.go:336` — Define a constant instead of duplicating this literal "https://api.example.com" 3 times.
- `client_test.go:530` — Define a constant instead of duplicating this literal "expected ok=true, got %#v" 3 times.
- `client_test.go:651` — Define a constant instead of duplicating this literal "expected validation to fail before the HTTP call" 3 times.
- `cmd/api/cmd_args_test.go:18` — Define a constant instead of duplicating this literal "expected %v, got %v" 4 times.
- `cmd/api/cmd_args_test.go:26` — Define a constant instead of duplicating this literal "expected nil, got %v" 3 times.
- `cmd/api/cmd_spec_test.go:145` — Define a constant instead of duplicating this literal "/api/v1/openapi.json" 7 times.
- `cmd/api/cmd_spec_test.go:147` — Define a constant instead of duplicating this literal "/api/v1/chat/completions" 7 times.
- `cmd/api/cmd_spec_test.go:180` — Define a constant instead of duplicating this literal "unexpected error: %v" 4 times.
- `codegen.go:68` — Define a constant instead of duplicating this literal "SDKGenerator.Generate" 11 times.
- `codegen_test.go:34` — Define a constant instead of duplicating this literal "spec.json" 4 times.
- `codegen_test.go:80` — Define a constant instead of duplicating this literal "failed to write spec file: %v" 3 times.
- `export_test.go:24` — Define a constant instead of duplicating this literal "Test API" 8 times.
- `export_test.go:28` — Define a constant instead of duplicating this literal "unexpected error: %v" 7 times.
- `export_test.go:33` — Define a constant instead of duplicating this literal "output is not valid JSON: %v" 3 times.
- `export_test.go:37` — Define a constant instead of duplicating this literal "expected openapi=3.1.0, got %v" 5 times.
- `expvar_test.go:24` — Define a constant instead of duplicating this literal "unexpected error: %v" 4 times.
- `expvar_test.go:30` — Define a constant instead of duplicating this literal "/debug/vars" 5 times.
- `expvar_test.go:32` — Define a constant instead of duplicating this literal "request failed: %v" 4 times.
- `graphql_test.go:58` — Define a constant instead of duplicating this literal "unexpected error: %v" 8 times.
- `graphql_test.go:65` — Define a constant instead of duplicating this literal "/graphql" 5 times.
- `graphql_test.go:65` — Define a constant instead of duplicating this literal "application/json" 7 times.
- `graphql_test.go:67` — Define a constant instead of duplicating this literal "request failed: %v" 7 times.
- `graphql_test.go:72` — Define a constant instead of duplicating this literal "expected 200, got %d" 3 times.
- `graphql_test.go:77` — Define a constant instead of duplicating this literal "failed to read body: %v" 4 times.
- `graphql_test.go:81` — Define a constant instead of duplicating this literal "expected response containing name:test, got %q" 3 times.
- `graphql_test.go:96` — Define a constant instead of duplicating this literal "/graphql/playground" 4 times.
- `graphql_test.go:175` — Define a constant instead of duplicating this literal "playground request failed: %v" 4 times.
- `group_test.go:41` — Define a constant instead of duplicating this literal "expected Name=%q, got %q" 3 times.
- `group_test.go:126` — Define a constant instead of duplicating this literal "List items" 3 times.
- `gzip_test.go:25` — Define a constant instead of duplicating this literal "/stub/ping" 5 times.
- `gzip_test.go:26` — Define a constant instead of duplicating this literal "Accept-Encoding" 4 times.
- `gzip_test.go:30` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `gzip_test.go:33` — Define a constant instead of duplicating this literal "Content-Encoding" 5 times.
- `httpsign_test.go:53` — Define a constant instead of duplicating this literal "(request-target)" 6 times.
- `httpsign_test.go:97` — Define a constant instead of duplicating this literal "/stub/ping" 5 times.
- `i18n_test.go:66` — Define a constant instead of duplicating this literal "/i18n/locale" 5 times.
- `i18n_test.go:67` — Define a constant instead of duplicating this literal "Accept-Language" 8 times.
- `i18n_test.go:71` — Define a constant instead of duplicating this literal "expected 200, got %d" 9 times.
- `i18n_test.go:76` — Define a constant instead of duplicating this literal "unmarshal error: %v" 9 times.
- `i18n_test.go:79` — Define a constant instead of duplicating this literal "expected locale=%q, got %q" 7 times.
- `i18n_test.go:215` — Define a constant instead of duplicating this literal "/i18n/greeting" 4 times.
- `location_test.go:49` — Define a constant instead of duplicating this literal "/loc/info" 5 times.
- `location_test.go:50` — Define a constant instead of duplicating this literal "X-Forwarded-Host" 3 times.
- `location_test.go:50` — Define a constant instead of duplicating this literal "api.example.com" 3 times.
- `location_test.go:54` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `location_test.go:59` — Define a constant instead of duplicating this literal "unmarshal error: %v" 5 times.
- `location_test.go:62` — Define a constant instead of duplicating this literal "expected host=%q, got %q" 3 times.
- `location_test.go:132` — Define a constant instead of duplicating this literal "proxy.example.com" 3 times.
- `location_test.go:163` — Define a constant instead of duplicating this literal "secure.example.com" 3 times.
- `middleware_test.go:25` — Define a constant instead of duplicating this literal "/secret" 3 times.
- `middleware_test.go:108` — Define a constant instead of duplicating this literal "/v1/secret" 4 times.
- `middleware_test.go:117` — Define a constant instead of duplicating this literal "unmarshal error: %v" 7 times.
- `middleware_test.go:160` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `middleware_test.go:178` — Define a constant instead of duplicating this literal "/health" 6 times.
- `middleware_test.go:230` — Define a constant instead of duplicating this literal "X-Request-ID" 9 times.
- `middleware_test.go:247` — Define a constant instead of duplicating this literal "client-id-abc" 3 times.
- `middleware_test.go:266` — Define a constant instead of duplicating this literal "client-id-xyz" 3 times.
- `middleware_test.go:289` — Define a constant instead of duplicating this literal "client-id-meta" 3 times.
- `middleware_test.go:301` — Define a constant instead of duplicating this literal "expected Meta to be present" 4 times.
- `middleware_test.go:304` — Define a constant instead of duplicating this literal "expected request_id=%q, got %q" 4 times.
- `middleware_test.go:307` — Define a constant instead of duplicating this literal "expected duration to be populated" 4 times.
- `middleware_test.go:325` — Define a constant instead of duplicating this literal "client-id-auto-meta" 5 times.
- `middleware_test.go:364` — Define a constant instead of duplicating this literal "client-id-auto-error-meta" 3 times.
- `middleware_test.go:400` — Define a constant instead of duplicating this literal "client-id-plus-json-meta" 3 times.
- `middleware_test.go:436` — Define a constant instead of duplicating this literal "Access-Control-Request-Method" 3 times.
- `middleware_test.go:444` — Define a constant instead of duplicating this literal "Access-Control-Allow-Origin" 3 times.
- `middleware_test.go:462` — Define a constant instead of duplicating this literal "https://app.example.com" 4 times.
- `modernization_test.go:25` — Define a constant instead of duplicating this literal "health-extra" 3 times.
- `modernization_test.go:99` — Define a constant instead of duplicating this literal "https://auth.example.com" 3 times.
- `modernization_test.go:102` — Define a constant instead of duplicating this literal "/public" 6 times.
- `openapi.go:302` — Define a constant instead of duplicating this literal "/health" 4 times.
- `openapi.go:363` — Define a constant instead of duplicating this literal "/debug/pprof" 3 times.
- `openapi.go:371` — Define a constant instead of duplicating this literal "/debug/vars" 3 times.
- `openapi.go:466` — Define a constant instead of duplicating this literal "application/json" 56 times.
- `openapi.go:593` — Define a constant instead of duplicating this literal "Bad request" 3 times.
- `openapi.go:602` — Define a constant instead of duplicating this literal "Too many requests" 7 times.
- `openapi.go:611` — Define a constant instead of duplicating this literal "Gateway timeout" 7 times.
- `openapi.go:620` — Define a constant instead of duplicating this literal "Internal server error" 7 times.
- `openapi_test.go:154` — Define a constant instead of duplicating this literal "unexpected error: %v" 66 times.
- `openapi_test.go:159` — Define a constant instead of duplicating this literal "invalid JSON: %v" 66 times.
- `openapi_test.go:172` — Define a constant instead of duplicating this literal "/health" 7 times.
- `openapi_test.go:173` — Define a constant instead of duplicating this literal "expected /health path in spec" 3 times.
- `openapi_test.go:191` — Define a constant instead of duplicating this literal "X-Request-ID" 6 times.
- `openapi_test.go:194` — Define a constant instead of duplicating this literal "X-RateLimit-Limit" 6 times.
- `openapi_test.go:197` — Define a constant instead of duplicating this literal "X-RateLimit-Remaining" 6 times.
- `openapi_test.go:200` — Define a constant instead of duplicating this literal "X-RateLimit-Reset" 6 times.
- `openapi_test.go:219` — Define a constant instead of duplicating this literal "X-Cache" 3 times.
- `openapi_test.go:444` — Define a constant instead of duplicating this literal "Test API" 4 times.
- `openapi_test.go:456` — Define a constant instead of duplicating this literal "https://example.com/terms" 3 times.
- `openapi_test.go:460` — Define a constant instead of duplicating this literal "API Support" 3 times.
- `openapi_test.go:463` — Define a constant instead of duplicating this literal "https://example.com/support" 3 times.
- `openapi_test.go:466` — Define a constant instead of duplicating this literal "support@example.com" 3 times.
- `openapi_test.go:470` — Define a constant instead of duplicating this literal "EUPL-1.2" 3 times.
- `openapi_test.go:473` — Define a constant instead of duplicating this literal "https://eupl.eu/1.2/en/" 3 times.
- `openapi_test.go:477` — Define a constant instead of duplicating this literal "Developer guide" 3 times.
- `openapi_test.go:480` — Define a constant instead of duplicating this literal "https://example.com/docs" 3 times.
- `openapi_test.go:483` — Define a constant instead of duplicating this literal "x-swagger-ui-path" 3 times.
- `openapi_test.go:587` — Define a constant instead of duplicating this literal "/graphql" 9 times.
- `openapi_test.go:650` — Define a constant instead of duplicating this literal "application/json" 8 times.
- `openapi_test.go:669` — Define a constant instead of duplicating this literal "/graphql/playground" 4 times.
- `openapi_test.go:784` — Define a constant instead of duplicating this literal "x-chat-completions-path" 3 times.
- `openapi_test.go:784` — Define a constant instead of duplicating this literal "/v1/chat/completions" 5 times.
- `openapi_test.go:949` — Define a constant instead of duplicating this literal "/v1/openapi.json" 5 times.
- `openapi_test.go:1053` — Define a constant instead of duplicating this literal "/events" 4 times.
- `openapi_test.go:1357` — Define a constant instead of duplicating this literal "/api/items" 3 times.
- `openapi_test.go:1374` — Define a constant instead of duplicating this literal "Create item" 4 times.
- `openapi_test.go:1471` — Define a constant instead of duplicating this literal "/status" 10 times.
- `openapi_test.go:1907` — Define a constant instead of duplicating this literal "/public" 3 times.
- `openapi_test.go:1908` — Define a constant instead of duplicating this literal "Public endpoint" 3 times.
- `openapi_test.go:1945` — Define a constant instead of duplicating this literal "/api/public" 4 times.
- `openapi_test.go:2218` — Define a constant instead of duplicating this literal "/api/users/{id}" 4 times.
- `openapi_test.go:2244` — Define a constant instead of duplicating this literal "/resources/{id}" 3 times.
- `openapi_test.go:2271` — Define a constant instead of duplicating this literal "/api/resources/{id}" 3 times.
- `openapi_test.go:2338` — Define a constant instead of duplicating this literal "Example resource" 4 times.
- `openapi_test.go:2437` — Define a constant instead of duplicating this literal "Content-Disposition" 3 times.
- `openapi_test.go:2502` — Define a constant instead of duplicating this literal "Get user" 4 times.
- `openapi_test.go:2831` — Define a constant instead of duplicating this literal "Check status" 4 times.
- `openapi_test.go:2852` — Define a constant instead of duplicating this literal "expected tags array, got %T" 5 times.
- `openapi_test.go:3358` — Define a constant instead of duplicating this literal "https://api.example.com" 6 times.
- `pkg/provider/cache_control_test.go:28` — Define a constant instead of duplicating this literal "Cache-Control" 5 times.
- `pkg/provider/proxy_internal_test.go:8` — Define a constant instead of duplicating this literal "/api/v1/cool-widget" 4 times.
- `pkg/provider/proxy_test.go:21` — Define a constant instead of duplicating this literal "cool-widget" 5 times.
- `pkg/provider/proxy_test.go:22` — Define a constant instead of duplicating this literal "/api/v1/cool-widget" 5 times.
- `pkg/provider/proxy_test.go:23` — Define a constant instead of duplicating this literal "http://127.0.0.1:9999" 5 times.
- `pkg/provider/proxy_test.go:69` — Define a constant instead of duplicating this literal "Content-Type" 3 times.
- `pkg/provider/proxy_test.go:69` — Define a constant instead of duplicating this literal "application/json" 3 times.
- `pkg/provider/registry_test.go:25` — Define a constant instead of duplicating this literal "stub.event" 6 times.
- `pkg/provider/registry_test.go:38` — Define a constant instead of duplicating this literal "core-stub-panel" 4 times.
- `pkg/provider/registry_test.go:53` — Define a constant instead of duplicating this literal "/api/full" 3 times.
- `pkg/provider/registry_test.go:60` — Define a constant instead of duplicating this literal "core-full-panel" 3 times.
- `pkg/provider/registry_test.go:316` — Define a constant instead of duplicating this literal "/tmp/a.yaml" 3 times.
- `pkg/stream/stream_group_test.go:22` — Define a constant instead of duplicating this literal "/events" 8 times.
- `pkg/stream/stream_group_test.go:23` — Define a constant instead of duplicating this literal "text/event-stream" 7 times.
- `pkg/stream/stream_group_test.go:152` — Define a constant instead of duplicating this literal "/tenant/socket" 3 times.
- `pprof_test.go:22` — Define a constant instead of duplicating this literal "unexpected error: %v" 4 times.
- `pprof_test.go:28` — Define a constant instead of duplicating this literal "/debug/pprof/" 3 times.
- `pprof_test.go:30` — Define a constant instead of duplicating this literal "request failed: %v" 4 times.
- `ratelimit_internal_test.go:28` — Define a constant instead of duplicating this literal "X-API-Key" 3 times.
- `ratelimit_internal_test.go:30` — Define a constant instead of duplicating this literal "203.0.113.10:1234" 3 times.
- `ratelimit_internal_test.go:79` — Define a constant instead of duplicating this literal "X-RateLimit-Remaining" 3 times.
- `ratelimit_test.go:37` — Define a constant instead of duplicating this literal "/rate/ping" 21 times.
- `ratelimit_test.go:38` — Define a constant instead of duplicating this literal "203.0.113.10:1234" 4 times.
- `ratelimit_test.go:43` — Define a constant instead of duplicating this literal "X-RateLimit-Limit" 3 times.
- `ratelimit_test.go:130` — Define a constant instead of duplicating this literal "203.0.113.20:1234" 3 times.
- `ratelimit_test.go:131` — Define a constant instead of duplicating this literal "X-API-Key" 5 times.
- `ratelimit_test.go:165` — Define a constant instead of duplicating this literal "203.0.113.30:1234" 3 times.
- `ratelimit_test.go:166` — Define a constant instead of duplicating this literal "Bearer token-a" 3 times.
- `ratelimit_test.go:195` — Define a constant instead of duplicating this literal "X-Principal" 3 times.
- `ratelimit_test.go:233` — Define a constant instead of duplicating this literal "X-User-ID" 4 times.
- `ratelimit_test.go:246` — Define a constant instead of duplicating this literal "203.0.113.42:1234" 3 times.
- `response_meta_test.go:91` — Define a constant instead of duplicating this literal "X-Preexisting" 4 times.
- `response_meta_test.go:100` — Define a constant instead of duplicating this literal "application/json" 3 times.
- `response_test.go:32` — Define a constant instead of duplicating this literal "expected Success=true" 3 times.
- `response_test.go:63` — Define a constant instead of duplicating this literal "marshal error: %v" 4 times.
- `response_test.go:68` — Define a constant instead of duplicating this literal "unmarshal error: %v" 7 times.
- `response_test.go:88` — Define a constant instead of duplicating this literal "resource not found" 3 times.
- `response_test.go:226` — Define a constant instead of duplicating this literal "unexpected error: %v" 3 times.
- `response_test.go:236` — Define a constant instead of duplicating this literal "/v1/meta" 3 times.
- `response_test.go:237` — Define a constant instead of duplicating this literal "client-id-meta" 6 times.
- `response_test.go:241` — Define a constant instead of duplicating this literal "expected 200, got %d" 3 times.
- `secure_test.go:24` — Define a constant instead of duplicating this literal "/health" 7 times.
- `secure_test.go:28` — Define a constant instead of duplicating this literal "expected 200, got %d" 4 times.
- `secure_test.go:52` — Define a constant instead of duplicating this literal "X-Frame-Options" 4 times.
- `secure_test.go:83` — Define a constant instead of duplicating this literal "strict-origin-when-cross-origin" 3 times.
- `servers_test.go:11` — Define a constant instead of duplicating this literal "https://api.example.com" 5 times.
- `sessions_test.go:42` — Define a constant instead of duplicating this literal "test-secret-key!" 4 times.
- `sessions_test.go:47` — Define a constant instead of duplicating this literal "/sess/set" 4 times.
- `sessions_test.go:51` — Define a constant instead of duplicating this literal "expected 200, got %d" 4 times.
- `slog_test.go:30` — Define a constant instead of duplicating this literal "/stub/ping" 3 times.
- `slog_test.go:34` — Define a constant instead of duplicating this literal "expected 200, got %d" 4 times.
- `slog_test.go:58` — Define a constant instead of duplicating this literal "/health" 3 times.
- `spec_builder_helper_test.go:22` — Define a constant instead of duplicating this literal "Engine API" 11 times.
- `spec_builder_helper_test.go:22` — Define a constant instead of duplicating this literal "Engine metadata" 11 times.
- `spec_builder_helper_test.go:23` — Define a constant instead of duplicating this literal "Engine overview" 6 times.
- `spec_builder_helper_test.go:25` — Define a constant instead of duplicating this literal "https://example.com/terms" 6 times.
- `spec_builder_helper_test.go:26` — Define a constant instead of duplicating this literal "support@example.com" 3 times.
- `spec_builder_helper_test.go:26` — Define a constant instead of duplicating this literal "API Support" 5 times.
- `spec_builder_helper_test.go:26` — Define a constant instead of duplicating this literal "https://example.com/support" 3 times.
- `spec_builder_helper_test.go:27` — Define a constant instead of duplicating this literal "https://api.example.com" 7 times.
- `spec_builder_helper_test.go:28` — Define a constant instead of duplicating this literal "https://eupl.eu/1.2/en/" 3 times.
- `spec_builder_helper_test.go:28` — Define a constant instead of duplicating this literal "EUPL-1.2" 5 times.
- `spec_builder_helper_test.go:33` — Define a constant instead of duplicating this literal "X-API-Key" 6 times.
- `spec_builder_helper_test.go:36` — Define a constant instead of duplicating this literal "Developer guide" 3 times.
- `spec_builder_helper_test.go:36` — Define a constant instead of duplicating this literal "https://example.com/docs" 5 times.
- `spec_builder_helper_test.go:43` — Define a constant instead of duplicating this literal "https://auth.example.com" 3 times.
- `spec_builder_helper_test.go:44` — Define a constant instead of duplicating this literal "core-client" 3 times.
- `spec_builder_helper_test.go:46` — Define a constant instead of duplicating this literal "/public" 4 times.
- `spec_builder_helper_test.go:48` — Define a constant instead of duplicating this literal "/socket" 7 times.
- `spec_builder_helper_test.go:52` — Define a constant instead of duplicating this literal "/events" 7 times.
- `spec_builder_helper_test.go:57` — Define a constant instead of duplicating this literal "unexpected error: %v" 27 times.
- `spec_builder_helper_test.go:68` — Define a constant instead of duplicating this literal "invalid JSON: %v" 8 times.
- `spec_builder_helper_test.go:88` — Define a constant instead of duplicating this literal "x-swagger-ui-path" 3 times.
- `spec_builder_helper_test.go:567` — Define a constant instead of duplicating this literal "/api/v1/openapi.json" 3 times.
- `spec_registry_test.go:31` — Define a constant instead of duplicating this literal "/alpha" 11 times.
- `sse_test.go:29` — Define a constant instead of duplicating this literal "unexpected error: %v" 12 times.
- `sse_test.go:35` — Define a constant instead of duplicating this literal "/events" 11 times.
- `sse_test.go:37` — Define a constant instead of duplicating this literal "request failed: %v" 11 times.
- `sse_test.go:42` — Define a constant instead of duplicating this literal "expected 200, got %d" 3 times.
- `sse_test.go:45` — Define a constant instead of duplicating this literal "Content-Type" 5 times.
- `sse_test.go:46` — Define a constant instead of duplicating this literal "text/event-stream" 5 times.
- `sse_test.go:47` — Define a constant instead of duplicating this literal "expected Content-Type starting with text/event-stream, got %q" 5 times.
- `sse_test.go:63` — Define a constant instead of duplicating this literal "/v1/events" 3 times.
- `sse_test.go:208` — Define a constant instead of duplicating this literal "event: " 4 times.
- `static_test.go:23` — Define a constant instead of duplicating this literal "hello world" 3 times.
- `static_test.go:24` — Define a constant instead of duplicating this literal "failed to write test file: %v" 4 times.
- `static_test.go:65` — Define a constant instead of duplicating this literal "<h1>Welcome</h1>" 3 times.
- `static_test.go:125` — Define a constant instead of duplicating this literal "sdk-data" 3 times.
- `static_test.go:130` — Define a constant instead of duplicating this literal "body{}" 3 times.
- `sunset_test.go:20` — Define a constant instead of duplicating this literal "/status" 3 times.
- `sunset_test.go:31` — Define a constant instead of duplicating this literal "<https://example.com/docs>; rel=\"help\"" 3 times.
- `sunset_test.go:44` — Define a constant instead of duplicating this literal "X-API-Warn" 3 times.
- `sunset_test.go:53` — Define a constant instead of duplicating this literal "/api/v2/status" 4 times.
- `sunset_test.go:53` — Define a constant instead of duplicating this literal "2025-06-01" 3 times.
- `sunset_test.go:55` — Define a constant instead of duplicating this literal "unexpected error: %v" 3 times.
- `sunset_test.go:64` — Define a constant instead of duplicating this literal "expected 200, got %d" 3 times.
- `sunset_test.go:75` — Define a constant instead of duplicating this literal "API-Suggested-Replacement" 8 times.
- `sunset_test.go:94` — Define a constant instead of duplicating this literal "Thu, 30 Apr 2026 23:59:59 GMT" 3 times.
- `sunset_test.go:105` — Define a constant instead of duplicating this literal "POST /api/v2/billing/invoices" 4 times.
- `sunset_test.go:109` — Define a constant instead of duplicating this literal "/billing" 12 times.
- `sunset_test.go:118` — Define a constant instead of duplicating this literal "</api/v2/billing/invoices>; rel=\"successor-version\"" 3 times.
- `sunset_test.go:131` — Define a constant instead of duplicating this literal "2026-04-30" 5 times.
- `swagger_test.go:23` — Define a constant instead of duplicating this literal "Test API" 13 times.
- `swagger_test.go:23` — Define a constant instead of duplicating this literal "A test API service" 8 times.
- `swagger_test.go:25` — Define a constant instead of duplicating this literal "unexpected error: %v" 23 times.
- `swagger_test.go:33` — Define a constant instead of duplicating this literal "/swagger/doc.json" 16 times.
- `swagger_test.go:35` — Define a constant instead of duplicating this literal "request failed: %v" 24 times.
- `swagger_test.go:40` — Define a constant instead of duplicating this literal "expected 200, got %d" 5 times.
- `swagger_test.go:45` — Define a constant instead of duplicating this literal "failed to read body: %v" 18 times.
- `swagger_test.go:267` — Define a constant instead of duplicating this literal "invalid JSON: %v" 16 times.
- `swagger_test.go:293` — Define a constant instead of duplicating this literal "/api/tools" 3 times.
- `swagger_test.go:296` — Define a constant instead of duplicating this literal "Query metrics data" 3 times.
- `swagger_test.go:535` — Define a constant instead of duplicating this literal "https://eupl.eu/1.2/en/" 5 times.
- `swagger_test.go:535` — Define a constant instead of duplicating this literal "EUPL-1.2" 5 times.
- `swagger_test.go:578` — Define a constant instead of duplicating this literal "support@example.com" 5 times.
- `swagger_test.go:578` — Define a constant instead of duplicating this literal "https://example.com/support" 5 times.
- `swagger_test.go:578` — Define a constant instead of duplicating this literal "API Support" 5 times.
- `swagger_test.go:624` — Define a constant instead of duplicating this literal "https://example.com/terms" 5 times.
- `swagger_test.go:660` — Define a constant instead of duplicating this literal "https://example.com/docs" 5 times.
- `swagger_test.go:660` — Define a constant instead of duplicating this literal "Developer guide" 5 times.
- `swagger_test.go:781` — Define a constant instead of duplicating this literal "https://api.example.com" 6 times.
- `swagger_test.go:950` — Define a constant instead of duplicating this literal "/v1/openapi.json" 5 times.
- `timeout_test.go:50` — Define a constant instead of duplicating this literal "/stub/ping" 3 times.
- `timeout_test.go:59` — Define a constant instead of duplicating this literal "unmarshal error: %v" 4 times.
- `timeout_test.go:65` — Define a constant instead of duplicating this literal "expected Data=%q, got %q" 3 times.
- `tracing_test.go:86` — Define a constant instead of duplicating this literal "/trace" 3 times.
- `tracing_test.go:123` — Define a constant instead of duplicating this literal "test-service" 4 times.
- `tracing_test.go:128` — Define a constant instead of duplicating this literal "/stub/ping" 5 times.
- `tracing_test.go:132` — Define a constant instead of duplicating this literal "expected 200, got %d" 8 times.
- `tracing_test.go:167` — Define a constant instead of duplicating this literal "expected at least one span" 5 times.
- `tracing_test.go:329` — Define a constant instead of duplicating this literal "tracing-test" 3 times.
- `transport_client_test.go:50` — Define a constant instead of duplicating this literal "Bearer secret" 8 times.
- `transport_client_test.go:103` — Define a constant instead of duplicating this literal "ws://example.invalid/ws" 4 times.
- `transport_client_test.go:194` — Define a constant instead of duplicating this literal "http://example.invalid/events" 3 times.
- `transport_client_test.go:204` — Define a constant instead of duplicating this literal "X-Request-ID" 5 times.
- `transport_client_test.go:231` — Define a constant instead of duplicating this literal "Content-Type" 3 times.
- `transport_client_test.go:231` — Define a constant instead of duplicating this literal "text/event-stream" 4 times.
- `webhook_test.go:404` — Define a constant instead of duplicating this literal "https://hooks.example.test/inbox" 4 times.
- `websocket_test.go:34` — Define a constant instead of duplicating this literal "wsstub.updates" 3 times.
- `websocket_test.go:34` — Define a constant instead of duplicating this literal "wsstub.events" 3 times.
- `websocket_test.go:49` — Define a constant instead of duplicating this literal "upgrade error: %v" 6 times.
- `websocket_test.go:58` — Define a constant instead of duplicating this literal "unexpected error: %v" 8 times.
- `websocket_test.go:68` — Define a constant instead of duplicating this literal "failed to dial WebSocket: %v" 3 times.
- `websocket_test.go:74` — Define a constant instead of duplicating this literal "failed to read message: %v" 5 times.
- `websocket_test.go:77` — Define a constant instead of duplicating this literal "expected message=%q, got %q" 6 times.
- `websocket_test.go:263` — Define a constant instead of duplicating this literal "gin-hello" 3 times.

### php:S1192 — String literals should not be duplicated (58×, code smell)

- `src/php/src/Api/Boot.php:206` — Define a constant instead of duplicating this literal "/Routes/api.php" 4 times.
- `src/php/src/Api/Boot.php:283` — Define a constant instead of duplicating this literal "/authorize" 3 times.
- `src/php/src/Api/Controllers/Api/WebhookSecretController.php:85` — Define a constant instead of duplicating this literal "Webhook endpoint" 4 times.
- `src/php/src/Api/Controllers/McpApiController.php:109` — Define a constant instead of duplicating this literal "The selected server id is invalid." 7 times.
- `src/php/src/Api/Controllers/McpApiController.php:346` — Define a constant instead of duplicating this literal "The selected tool name is invalid." 5 times.
- `src/php/src/Api/Database/Factories/ApiKeyFactory.php:52` — Define a constant instead of duplicating this literal " API Key" 3 times.
- `src/php/src/Api/Documentation/DocumentationServiceProvider.php:26` — Define a constant instead of duplicating this literal "/config.php" 5 times.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:456` — Define a constant instead of duplicating this literal "Bio Links" 4 times.
- `src/php/src/Api/Models/WebhookEndpoint.php:227` — Define a constant instead of duplicating this literal "The webhook URL must resolve to a public IP address." 3 times.
- `src/php/src/Api/Routes/api.php:134` — Define a constant instead of duplicating this literal "/{workspace}" 4 times.
- `src/php/src/Api/Routes/api.php:161` — Define a constant instead of duplicating this literal "/{id}" 12 times.
- `src/php/src/Api/Services/SeoReportService.php:511` — Define a constant instead of duplicating this literal "The supplied URL could not be resolved to any address." 4 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:31` — Define a constant instead of duplicating this literal "192.168.1.1" 19 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:35` — Define a constant instead of duplicating this literal "10.0.0.1" 13 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:43` — Define a constant instead of duplicating this literal "192.168.1.0/24" 11 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:67` — Define a constant instead of duplicating this literal "10.0.0.0/8" 4 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:89` — Define a constant instead of duplicating this literal "2001:db8::1" 9 times.
- `src/php/src/Api/Tests/Feature/ApiKeyIpWhitelistTest.php:112` — Define a constant instead of duplicating this literal "2001:db8::/32" 4 times.
- `src/php/src/Api/Tests/Feature/ApiKeyTest.php:386` — Define a constant instead of duplicating this literal "Active Key" 3 times.
- `src/php/src/Api/Tests/Feature/ApiKeyTest.php:719` — Define a constant instead of duplicating this literal "/api/mcp/servers" 3 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:44` — Define a constant instead of duplicating this literal "Read Only Key" 5 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:64` — Define a constant instead of duplicating this literal "/api/test-scope/write" 4 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:81` — Define a constant instead of duplicating this literal "/api/test-scope/delete" 6 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:100` — Define a constant instead of duplicating this literal "Read/Write Key" 4 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:243` — Define a constant instead of duplicating this literal "Posts Admin Key" 3 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:244` — Define a constant instead of duplicating this literal "posts:*" 7 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:303` — Define a constant instead of duplicating this literal "*:read" 5 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:524` — Define a constant instead of duplicating this literal "/test-explicit/posts" 3 times.
- `src/php/src/Api/Tests/Feature/ApiScopeEnforcementTest.php:541` — Define a constant instead of duplicating this literal "/api/test-explicit/posts" 8 times.
- `src/php/src/Api/Tests/Feature/ApiUsageTest.php:37` — Define a constant instead of duplicating this literal "/api/v1/workspaces" 4 times.
- `src/php/src/Api/Tests/Feature/ApiUsageTest.php:83` — Define a constant instead of duplicating this literal "/api/v1/test" 8 times.
- `src/php/src/Api/Tests/Feature/ApiUsageTest.php:192` — Define a constant instead of duplicating this literal "/api/v1/old" 3 times.
- `src/php/src/Api/Tests/Feature/AuthenticateApiKeyTest.php:46` — Define a constant instead of duplicating this literal "/api/test-auth/scoped" 4 times.
- `src/php/src/Api/Tests/Feature/DocumentationControllerTest.php:102` — Define a constant instead of duplicating this literal "/api/docs" 3 times.
- `src/php/src/Api/Tests/Feature/McpResourceTest.php:78` — Define a constant instead of duplicating this literal "test-resource-server://documents/welcome" 4 times.
- `src/php/src/Api/Tests/Feature/McpServerAccessTest.php:51` — Define a constant instead of duplicating this literal "/allowed-server.yaml" 6 times.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:108` — Define a constant instead of duplicating this literal "/test-scan/items/{id}" 4 times.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:118` — Define a constant instead of duplicating this literal "api/*" 18 times.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:674` — Define a constant instead of duplicating this literal "Custom Tag" 3 times.
- `src/php/src/Api/Tests/Feature/PixelEndpointTest.php:16` — Define a constant instead of duplicating this literal "/api/pixel/abc12345" 3 times.
- `src/php/src/Api/Tests/Feature/PixelEndpointTest.php:17` — Define a constant instead of duplicating this literal "https://example.com" 6 times.
- `src/php/src/Api/Tests/Feature/PublicApiCorsTest.php:48` — Define a constant instead of duplicating this literal "https://example.com" 5 times.
- `src/php/src/Api/Tests/Feature/RateLimitingTest.php:706` — Define a constant instead of duplicating this literal "127.0.0.1" 3 times.
- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:45` — Define a constant instead of duplicating this literal "https://1.1.1.1/article" 5 times.
- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:75` — Define a constant instead of duplicating this literal "text/html; charset=utf-8" 4 times.
- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:87` — Define a constant instead of duplicating this literal "Example Product Landing Page" 3 times.
- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:88` — Define a constant instead of duplicating this literal "A concise example description for the landing page." 3 times.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:69` — Define a constant instead of duplicating this literal "{"event":"test"}" 13 times.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:331` — Define a constant instead of duplicating this literal "https://1.1.1.1/webhook" 16 times.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:460` — Define a constant instead of duplicating this literal "https://example.com/webhook" 13 times.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:535` — Define a constant instead of duplicating this literal "Server Error" 3 times.
- `src/php/src/Website/Api/Services/OpenApiGenerator.php:88` — Define a constant instead of duplicating this literal "Chat Widget" 3 times.
- `src/php/tests/Feature/ApiSunsetTest.php:14` — Define a constant instead of duplicating this literal "/legacy-endpoint" 10 times.
- `src/php/tests/Feature/ApiSunsetTest.php:44` — Define a constant instead of duplicating this literal "2025-06-01" 7 times.
- `src/php/tests/Feature/ApiSunsetTest.php:46` — Define a constant instead of duplicating this literal "</api/v2/users>; rel="successor-version"" 4 times.
- `src/php/tests/Feature/ApiSunsetTest.php:57` — Define a constant instead of duplicating this literal "/api/v2/users" 9 times.
- `src/php/tests/Feature/ApiVersionServiceTest.php:47` — Define a constant instead of duplicating this literal "/api/users" 6 times.
- `src/php/tests/Feature/AuthenticationGuideTest.php:21` — Define a constant instead of duplicating this literal "API keys are prefixed with" 3 times.

### go:S3776 — Cognitive Complexity of functions should not be too high (39×, code smell)

- `api.go:253` — Refactor this method to reduce its Cognitive Complexity from 19 to the 15 allowed.
- `authentik.go:171` — Refactor this method to reduce its Cognitive Complexity from 38 to the 15 allowed.
- `authentik_integration_test.go:89` — Refactor this method to reduce its Cognitive Complexity from 23 to the 15 allowed.
- `bridge.go:451` — Refactor this method to reduce its Cognitive Complexity from 97 to the 15 allowed.
- `bridge.go:566` — Refactor this method to reduce its Cognitive Complexity from 27 to the 15 allowed.
- `cache.go:90` — Refactor this method to reduce its Cognitive Complexity from 19 to the 15 allowed.
- `cache.go:191` — Refactor this method to reduce its Cognitive Complexity from 41 to the 15 allowed.
- `chat_completions.go:375` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `chat_completions.go:716` — Refactor this method to reduce its Cognitive Complexity from 33 to the 15 allowed.
- `client.go:181` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `client.go:291` — Refactor this method to reduce its Cognitive Complexity from 37 to the 15 allowed.
- `client.go:398` — Refactor this method to reduce its Cognitive Complexity from 37 to the 15 allowed.
- `client.go:502` — Refactor this method to reduce its Cognitive Complexity from 28 to the 15 allowed.
- `client.go:570` — Refactor this method to reduce its Cognitive Complexity from 19 to the 15 allowed.
- `client.go:775` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `client_test.go:749` — Refactor this method to reduce its Cognitive Complexity from 18 to the 15 allowed.
- `cmd/api/cmd_sdk.go:31` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `i18n.go:159` — Refactor this method to reduce its Cognitive Complexity from 21 to the 15 allowed.
- `openapi.go:85` — Refactor this method to reduce its Cognitive Complexity from 49 to the 15 allowed.
- `openapi.go:297` — Refactor this method to reduce its Cognitive Complexity from 88 to the 15 allowed.
- `openapi.go:943` — Refactor this method to reduce its Cognitive Complexity from 23 to the 15 allowed.
- `openapi.go:1983` — Refactor this method to reduce its Cognitive Complexity from 32 to the 15 allowed.
- `openapi.go:2214` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `openapi.go:2750` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `openapi_test.go:145` — Refactor this method to reduce its Cognitive Complexity from 28 to the 15 allowed.
- `openapi_test.go:582` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `openapi_test.go:1722` — Refactor this method to reduce its Cognitive Complexity from 21 to the 15 allowed.
- `openapi_test.go:2075` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `openapi_test.go:2914` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `pkg/provider/registry.go:213` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `pkg/stream/stream_group_test.go:168` — Refactor this method to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `ratelimit.go:63` — Refactor this method to reduce its Cognitive Complexity from 37 to the 15 allowed.
- `runtime_config_test.go:15` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `spec_builder_helper.go:238` — Refactor this method to reduce its Cognitive Complexity from 26 to the 15 allowed.
- `spec_builder_helper_test.go:17` — Refactor this method to reduce its Cognitive Complexity from 57 to the 15 allowed.
- `spec_builder_helper_test.go:247` — Refactor this method to reduce its Cognitive Complexity from 21 to the 15 allowed.
- `spec_builder_helper_test.go:347` — Refactor this method to reduce its Cognitive Complexity from 21 to the 15 allowed.
- `sse.go:149` — Refactor this method to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `transport_client.go:264` — Refactor this method to reduce its Cognitive Complexity from 18 to the 15 allowed.

### go:S1186 — Functions should not be empty (31×, code smell)

- `api_describable_test.go:23` — Add a nested comment explaining why this function is empty or complete the implementation.
- `api_renderable_test.go:23` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge.go:804` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:132` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:198` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:266` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:277` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:334` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:967` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:968` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:969` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:1026` — Add a nested comment explaining why this function is empty or complete the implementation.
- `bridge_test.go:1031` — Add a nested comment explaining why this function is empty or complete the implementation.
- `cache_control_test.go:19` — Add a nested comment explaining why this function is empty or complete the implementation.
- `cmd/api/cmd_sdk_test.go:166` — Add a nested comment explaining why this function is empty or complete the implementation.
- `cmd/api/cmd_spec_test.go:21` — Add a nested comment explaining why this function is empty or complete the implementation.
- `cmd/api/spec_groups_iter.go:51` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:28` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:36` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:46` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:66` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:81` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:102` — Add a nested comment explaining why this function is empty or complete the implementation.
- `openapi_test.go:140` — Add a nested comment explaining why this function is empty or complete the implementation.
- `pkg/provider/registry_test.go:21` — Add a nested comment explaining why this function is empty or complete the implementation.
- `pkg/stream/stream_group_test.go:83` — Add a nested comment explaining why this function is empty or complete the implementation.
- `spec_builder_helper_test.go:49` — Add a nested comment explaining why this function is empty or complete the implementation.
- `spec_builder_helper_test.go:436` — Add a nested comment explaining why this function is empty or complete the implementation.
- `spec_registry_test.go:21` — Add a nested comment explaining why this function is empty or complete the implementation.
- `swagger_internal_test.go:20` — Add a nested comment explaining why this function is empty or complete the implementation.
- `tracing_test.go:111` — Add a nested comment explaining why this function is empty or complete the implementation.

### php:S1186 — Methods should not be empty (17×, code smell)

- `src/php/src/Api/Tests/Feature/McpApiControllerTest.php:167` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/McpApiControllerTest.php:176` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:170` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:180` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:189` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1197` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1202` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1206` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1211` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1215` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1226` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1234` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1242` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1244` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1253` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1264` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.
- `src/php/src/Api/Tests/Feature/RateLimitTest.php:257` — Add a nested comment explaining why this method is empty, throw an Exception or complete the implementation.

### php:S3776 — Cognitive Complexity of functions should not be too high (17×, code smell)

- `src/php/src/Api/Console/Commands/CheckApiUsageAlerts.php:125` — Refactor this function to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:411` — Refactor this function to reduce its Cognitive Complexity from 20 to the 15 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:598` — Refactor this function to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:754` — Refactor this function to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:944` — Refactor this function to reduce its Cognitive Complexity from 89 to the 15 allowed.
- `src/php/src/Api/Documentation/Extensions/SunsetExtension.php:76` — Refactor this function to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `src/php/src/Api/Documentation/Extensions/SunsetExtension.php:139` — Refactor this function to reduce its Cognitive Complexity from 19 to the 15 allowed.
- `src/php/src/Api/Documentation/Middleware/ProtectDocumentation.php:22` — Refactor this function to reduce its Cognitive Complexity from 21 to the 15 allowed.
- `src/php/src/Api/Middleware/AuthenticateApiKey.php:120` — Refactor this function to reduce its Cognitive Complexity from 18 to the 15 allowed.
- `src/php/src/Api/Models/ApiKey.php:162` — Refactor this function to reduce its Cognitive Complexity from 18 to the 15 allowed.
- `src/php/src/Api/Models/WebhookEndpoint.php:178` — Refactor this function to reduce its Cognitive Complexity from 16 to the 15 allowed.
- `src/php/src/Api/Services/SeoReportService.php:456` — Refactor this function to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `src/php/src/Api/Services/SeoReportService.php:536` — Refactor this function to reduce its Cognitive Complexity from 17 to the 15 allowed.
- `src/php/src/Api/Services/WebhookSecretRotationService.php:274` — Refactor this function to reduce its Cognitive Complexity from 18 to the 15 allowed.
- `src/php/src/Api/Tests/Feature/RateLimitingTest.php:459` — Refactor this function to reduce its Cognitive Complexity from 24 to the 15 allowed.
- `src/php/src/Front/Api/Middleware/ApiVersion.php:75` — Refactor this function to reduce its Cognitive Complexity from 22 to the 15 allowed.
- `src/php/src/Front/Api/VersionedRoutes.php:252` — Refactor this function to reduce its Cognitive Complexity from 20 to the 15 allowed.

## MAJOR

### php:S1142 — Functions should not contain too many return statements (62×, code smell)

- `src/php/src/Api/Concerns/ResolvesWorkspace.php:27` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Console/Commands/CheckApiUsageAlerts.php:259` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/Api/ApiKeyController.php:59` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/Api/PaymentMethodController.php:84` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/Api/WebhookTemplateController.php:133` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/Api/WebhookTemplateController.php:190` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/Api/WorkspaceMemberController.php:92` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:105` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:154` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:221` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:373` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:411` — This method has 8 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:666` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:711` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:754` — This method has 9 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:1308` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:1362` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:1429` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:1498` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Controllers/McpApiController.php:1520` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/Extensions/RateLimitExtension.php:152` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/Extensions/RateLimitExtension.php:188` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/Extensions/RateLimitExtension.php:234` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/Extensions/VersionExtension.php:98` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/Middleware/ProtectDocumentation.php:22` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:356` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:492` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:749` — This method has 8 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:959` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:1103` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/ApiCacheControl.php:23` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/AuthenticateApiKey.php:31` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/AuthenticateApiKey.php:77` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/AuthenticateApiKey.php:120` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/AuthenticateApiKey.php:181` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/RateLimitApi.php:82` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/RateLimitApi.php:134` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/RateLimitApi.php:192` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/RateLimitApi.php:313` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Middleware/RateLimitApi.php:345` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Models/ApiKey.php:162` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Models/ApiKey.php:331` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Models/WebhookDelivery.php:203` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Models/WebhookEndpoint.php:296` — This method has 7 returns, which is more than the 3 allowed.
- `src/php/src/Api/Models/WebhookEndpoint.php:424` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/RateLimit/RateLimitService.php:208` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/RateLimit/RateLimitService.php:264` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/RateLimit/RateLimitService.php:342` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/IpRestrictionService.php:26` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/IpRestrictionService.php:66` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/IpRestrictionService.php:128` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/IpRestrictionService.php:208` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/IpRestrictionService.php:234` — This method has 6 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/SeoReportService.php:591` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/WebhookSecretRotationService.php:85` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/WebhookSecretRotationService.php:274` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/WebhookTemplateService.php:214` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/WebhookTemplateService.php:547` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Services/WebhookTemplateService.php:568` — This method has 5 returns, which is more than the 3 allowed.
- `src/php/src/Api/Tests/Feature/WebhookEndpointTest.php:6` — This function has 5 returns, which is more than the 3 allowed.
- `src/php/src/Front/Api/Middleware/ApiSunset.php:105` — This method has 4 returns, which is more than the 3 allowed.
- `src/php/src/Website/Api/Services/OpenApiGenerator.php:308` — This method has 4 returns, which is more than the 3 allowed.

### php:S112 — Generic exceptions ErrorException, RuntimeException and Exception should not be thrown (33×, code smell)

- `src/php/src/Api/Controllers/McpApiController.php:854` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:863` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:867` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:903` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:914` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:931` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:935` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:960` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:970` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:980` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1018` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1028` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1052` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1069` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1096` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Controllers/McpApiController.php:1102` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Models/WebhookDelivery.php:87` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Models/WebhookDelivery.php:184` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/ApiKeyService.php:51` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/ApiKeyService.php:90` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/ApiKeyService.php:97` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/ApiKeyService.php:133` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:43` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:50` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:134` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:141` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:148` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Services/SeoReportService.php:154` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Tests/Feature/ApiKeyTest.php:279` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Tests/Feature/AuthenticateApiKeyTest.php:106` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Tests/Feature/McpApiControllerTest.php:164` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Tests/Feature/RateLimitTest.php:281` — Define and throw a dedicated exception instead of using a generic one.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:443` — Define and throw a dedicated exception instead of using a generic one.

### php:S1172 — Unused function parameters should be removed (29×, code smell)

- `src/php/src/Api/Database/Factories/ApiKeyFactory.php:139` — Remove the unused function parameter "$attributes".
- `src/php/src/Api/Documentation/DocumentationController.php:45` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:58` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:71` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:81` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:94` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:105` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationController.php:120` — Remove the unused function parameter "$request".
- `src/php/src/Api/Documentation/DocumentationServiceProvider.php:40` — Remove the unused function parameter "$app".
- `src/php/src/Api/Documentation/Examples/CommonExamples.php:121` — Remove the unused function parameter "$status".
- `src/php/src/Api/Documentation/OpenApiBuilder.php:525` — Remove the unused function parameter "$config".
- `src/php/src/Api/Documentation/OpenApiBuilder.php:836` — Remove the unused function parameter "$value".
- `src/php/src/Api/Documentation/OpenApiBuilder.php:907` — Remove the unused function parameter "$route".
- `src/php/src/Api/Services/WebhookTemplateService.php:547` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:568` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:596` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:601` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:606` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:632` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Services/WebhookTemplateService.php:637` — Remove the unused function parameter "$arg".
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:72` — Remove the unused function parameter "$serverId".
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:72` — Remove the unused function parameter "$toolName".
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:143` — Remove the unused function parameter "$server".
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:143` — Remove the unused function parameter "$version".
- `src/php/src/Api/Tests/Feature/McpServerDetailTest.php:143` — Remove the unused function parameter "$tool".
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1204` — Remove the unused function parameter "$id".
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1213` — Remove the unused function parameter "$id".
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1217` — Remove the unused function parameter "$id".
- `src/php/src/Api/Tests/Feature/OpenApiDocumentationComprehensiveTest.php:1255` — Remove the unused function parameter "$id".

### Web:S5255 — "aria-label" or "aria-labelledby" attributes should be used to differentiate similar elements (12×, code smell)

- `src/php/src/Website/Api/View/Blade/guides/authentication.blade.php:12` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/authentication.blade.php:49` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/errors.blade.php:11` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/errors.blade.php:48` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/qrcodes.blade.php:11` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/qrcodes.blade.php:48` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/quickstart.blade.php:12` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/quickstart.blade.php:49` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/rate-limits.blade.php:10` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/rate-limits.blade.php:23` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/webhooks.blade.php:11` — Add an "aria-label" or "aria-labbelledby" attribute to this element.
- `src/php/src/Website/Api/View/Blade/guides/webhooks.blade.php:63` — Add an "aria-label" or "aria-labbelledby" attribute to this element.

### php:S1448 — Classes should not have too many methods (8×, code smell)

- `src/php/src/Api/Controllers/McpApiController.php:27` — Class "McpApiController" has 37 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Documentation/OpenApiBuilder.php:31` — Class "OpenApiBuilder" has 38 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Models/ApiKey.php:26` — Class "ApiKey" has 35 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Models/WebhookEndpoint.php:32` — Class "WebhookEndpoint" has 25 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Models/WebhookPayloadTemplate.php:41` — Class "WebhookPayloadTemplate" has 24 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Services/ApiSnippetService.php:12` — Class "ApiSnippetService" has 21 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/Services/WebhookTemplateService.php:22` — Class "WebhookTemplateService" has 28 methods, which is greater than 20 authorized. Split it into smaller classes.
- `src/php/src/Api/View/Modal/Admin/WebhookTemplateManager.php:20` — Class "WebhookTemplateManager" has 27 methods, which is greater than 20 authorized. Split it into smaller classes.

### php:S3358 — Ternary operators should not be nested (2×, code smell)

- `src/php/src/Api/Models/WebhookEndpoint.php:198` — Extract this nested ternary operation into an independent statement.
- `src/php/src/Api/Services/SeoReportService.php:473` — Extract this nested ternary operation into an independent statement.

### php:S138 — Functions should not have too many lines of code (2×, code smell)

- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:51` — This function expression has 158 lines, which is greater than the 150 lines authorized. Split it into smaller functions.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:550` — This function expression has 215 lines, which is greater than the 150 lines authorized. Split it into smaller functions.

### Web:S6853 — Label elements should have a text label and an associated control (2×, code smell)

- `src/php/src/Api/View/Blade/admin/webhook-template-manager.blade.php:264` — A form label must be associated with a control and have accessible text.
- `src/php/src/Api/View/Blade/admin/webhook-template-manager.blade.php:268` — A form label must be associated with a control and have accessible text.

### php:S3011 — Reflection should not be used to increase accessibility of classes, methods, or fields (2×, code smell)

- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:40` — Make sure that this accessibility bypass is safe here.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:36` — Make sure that this accessibility bypass is safe here.

### php:S107 — Functions should not have too many parameters (1×, code smell)

- `src/php/src/Api/Services/ApiUsageService.php:22` — This function has 10 parameters, which is greater than the 7 authorized.

### php:S1066 — Mergeable "if" statements should be combined (1×, code smell)

- `src/php/src/Api/Services/SeoReportService.php:297` — Merge this if statement with the enclosing one.

### go:S107 — Functions should not have too many parameters (1×, code smell)

- `openapi.go:554` — This function has 12 parameters, which is greater than the 7 authorized.

### php:S1068 — Unused "private" fields should be removed (1×, code smell)

- `src/php/src/Api/Services/WebhookSignature.php:55` — Remove this unused "SECRET_LENGTH" private field.

## MINOR

### php:S1481 — Unused local variables should be removed (7×, code smell)

- `src/php/src/Api/Documentation/OpenApiBuilder.php:452` — Remove this unused "$name" local variable.
- `src/php/src/Api/Tests/Feature/ApiKeyRotationTest.php:218` — Remove this unused "$key1" local variable.
- `src/php/src/Api/Tests/Feature/ApiUsageTest.php:189` — Remove this unused "$usage" local variable.
- `src/php/src/Api/Tests/Feature/RateLimitTest.php:606` — Remove this unused "$tier" local variable.
- `src/php/src/Api/Tests/Feature/RateLimitingTest.php:360` — Remove this unused "$apiKey2" local variable.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:433` — Remove this unused "$endpoint" local variable.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:474` — Remove this unused "$endpoint" local variable.

### php:S100 — Function names should comply with a naming convention (3×, code smell)

- `src/php/src/Api/Tests/Feature/SeoReportServiceTest.php:6` — Rename function "dns_get_record" to match the regular expression ^[a-z][a-zA-Z0-9]*$.
- `src/php/src/Api/Tests/Feature/WebhookDeliveryTest.php:6` — Rename function "dns_get_record" to match the regular expression ^[a-z][a-zA-Z0-9]*$.
- `src/php/src/Api/Tests/Feature/WebhookEndpointTest.php:6` — Rename function "dns_get_record" to match the regular expression ^[a-z][a-zA-Z0-9]*$.

### go:S1940 — Boolean checks should not be inverted (2×, code smell)

- `client.go:687` — Use the opposite operator ("!=") instead.
- `client.go:729` — Use the opposite operator ("!=") instead.

### php:S6353 — Regular expression quantifiers and character classes should be used concisely (2×, code smell)

- `src/php/src/Api/Services/WebhookTemplateService.php:343` — Use concise character class syntax '\w' instead of '[a-zA-Z0-9_]'.
- `src/php/src/Api/Services/WebhookTemplateService.php:486` — Use concise character class syntax '\w' instead of '[a-zA-Z0-9_]'.

### php:S1488 — Local variables should not be declared and then immediately returned or thrown (1×, code smell)

- `src/php/src/Api/Services/WebhookTemplateService.php:139` — Immediately return this expression instead of assigning it to the temporary variable "$variables".

