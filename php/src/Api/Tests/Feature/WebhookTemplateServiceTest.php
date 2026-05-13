<?php

declare(strict_types=1);

use Carbon\Carbon;
use Core\Api\Contracts\WebhookEvent;
use Core\Api\Enums\BuiltinTemplateType;
use Core\Api\Enums\WebhookTemplateFormat;
use Core\Api\Services\WebhookTemplateService;

final class WebhookTemplateServiceTestEvent implements WebhookEvent
{
    public static function name(): string
    {
        return 'workspace.created';
    }

    public static function nameLocalised(): string
    {
        return 'Workspace Created';
    }

    public function payload(): array
    {
        return [
            'id' => 123,
            'name' => 'Acme',
            'tags' => ['alpha', 'beta'],
            'user' => [
                'name' => 'Ada',
            ],
        ];
    }

    public function message(): string
    {
        return 'Workspace created successfully';
    }
}

beforeEach(function () {
    Carbon::setTestNow(Carbon::create(2026, 4, 17, 12, 0, 0, 'UTC'));
    $this->service = app(WebhookTemplateService::class);
});

afterEach(function () {
    Carbon::setTestNow();
});

it('WebhookTemplateService_getBuiltinTemplates_Good_exposes_all_builtin_templates', function () {
    $templates = $this->service->getBuiltinTemplates();

    foreach (BuiltinTemplateType::cases() as $type) {
        expect($templates)->toHaveKey($type->value);
        expect($templates[$type->value]['name'])->toBe($type->label());
        expect($templates[$type->value]['description'])->toBe($type->description());
        expect($templates[$type->value]['template'])->toBe($type->template());
        expect($templates[$type->value]['format'])->toBe($type->format());
    }
});

it('WebhookTemplateService_validateTemplate_Good_accepts_structurally_valid_json_templates', function () {
    $result = $this->service->validateTemplate(
        '{"event": {{event.type}}, "name": {{event.name}}}',
        WebhookTemplateFormat::JSON,
    );

    expect($result)->toBe([
        'valid' => true,
        'errors' => [],
    ]);
});

it('WebhookTemplateService_validateTemplate_Bad_rejects_invalid_json_structure', function () {
    $result = $this->service->validateTemplate(
        '{"event": {{event.type}}, "name": }',
        WebhookTemplateFormat::JSON,
    );

    expect($result['valid'])->toBeFalse();
    expect($result['errors'])->toContain(
        'Template is not valid JSON structure: Syntax error'
    );
});

it('WebhookTemplateService_validateTemplate_Ugly_reports_unknown_filters_and_mismatched_braces', function () {
    $result = $this->service->validateTemplate(
        '{{value | mystery}',
        WebhookTemplateFormat::SIMPLE,
    );

    expect($result['valid'])->toBeFalse();
    expect($result['errors'])->toContain('Mismatched template braces. Found 1 opening and 0 closing.');
    expect($result['errors'])->toContain(
        'Unknown filter: mystery. Available: iso8601, timestamp, currency, json, upper, lower, default, truncate, escape, urlencode'
    );
});

it('WebhookTemplateService_validateTemplate_Good_accepts_supported_mustache_blocks', function () {
    $result = $this->service->validateTemplate(
        '{"labels":[{{#each data.tags}}"{{this}}"{{/each}}],"has_user":{{#if data.user}}true{{/if}},"has_optional":{{#unless data.optional}}false{{/unless}}}',
        WebhookTemplateFormat::MUSTACHE,
    );

    expect($result)->toBe([
        'valid' => true,
        'errors' => [],
    ]);
});

it('WebhookTemplateService_validateTemplate_Good_accepts_loop_context_variables_in_mustache_blocks', function () {
    $result = $this->service->validateTemplate(
        '{"is_first":{{#if @first}}true{{/if}}}',
        WebhookTemplateFormat::MUSTACHE,
    );

    expect($result)->toBe([
        'valid' => true,
        'errors' => [],
    ]);
});

it('WebhookTemplateService_buildDefaultPayload_Good_uses_the_event_contract', function () {
    $result = $this->service->buildDefaultPayload(new WebhookTemplateServiceTestEvent());

    expect($result)->toBe([
        'event' => 'workspace.created',
        'data' => [
            'id' => 123,
            'name' => 'Acme',
            'tags' => ['alpha', 'beta'],
            'user' => [
                'name' => 'Ada',
            ],
        ],
        'timestamp' => '2026-04-17T12:00:00+00:00',
    ]);
});

it('WebhookTemplateService_renderTemplate_Good_renders_simple_templates_with_filters', function () {
    $result = $this->service->renderTemplate(
        '{"event":"{{event.type}}","message":"{{message | upper}}","tags":{{data.tags | json}}}',
        WebhookTemplateFormat::SIMPLE,
        [
            'event' => [
                'type' => 'workspace.created',
            ],
            'message' => 'sample message',
            'data' => [
                'tags' => ['alpha', 'beta'],
            ],
        ],
    );

    expect($result)->toBe([
        'event' => 'workspace.created',
        'message' => 'SAMPLE MESSAGE',
        'tags' => ['alpha', 'beta'],
    ]);
});

it('WebhookTemplateService_renderTemplate_Good_renders_mustache_conditionals_and_loops', function () {
    $result = $this->service->renderTemplate(
        '{"labels":[{{#each data.tags}}"{{this}}"{{/each}}],"has_user":{{#if data.user}}true{{/if}},"has_optional":{{#unless data.optional}}false{{/unless}}}',
        WebhookTemplateFormat::MUSTACHE,
        [
            'data' => [
                'tags' => ['alpha'],
                'user' => [
                    'name' => 'Ada',
                ],
            ],
        ],
    );

    expect($result)->toBe([
        'labels' => ['alpha'],
        'has_user' => true,
        'has_optional' => false,
    ]);
});

it('WebhookTemplateService_renderTemplate_Bad_throws_when_rendered_json_is_invalid', function () {
    expect(fn () => $this->service->renderTemplate(
        '{"event": {{event.type}}}',
        WebhookTemplateFormat::JSON,
        [
            'event' => [
                'type' => 'workspace.created',
            ],
        ],
    ))->toThrow(InvalidArgumentException::class, 'Template rendered to invalid JSON');
});

it('WebhookTemplateService_renderTemplate_Ugly_preserves_arrays_and_missing_values_in_valid_json', function () {
    $result = $this->service->renderTemplate(
        '{"missing":"{{missing}}","data":{{data}}}',
        WebhookTemplateFormat::JSON,
        [
            'data' => [
                'id' => 123,
            ],
        ],
    );

    expect($result)->toBe([
        'missing' => '',
        'data' => [
            'id' => 123,
        ],
    ]);
});

it('WebhookTemplateService_previewPayload_Good_returns_a_rendered_sample_payload', function () {
    $result = $this->service->previewPayload(
        '{"event": {{event.type | json}}, "message": {{message | json}}, "id": {{data.id}}}',
        WebhookTemplateFormat::JSON,
        'workspace.created',
    );

    expect($result['success'])->toBeTrue();
    expect($result['errors'])->toBe([]);
    expect($result['output'])->toBe([
        'event' => 'workspace.created',
        'message' => 'Sample webhook event message',
        'id' => '123',
    ]);
});

it('WebhookTemplateService_previewPayload_Bad_rejects_empty_templates', function () {
    $result = $this->service->previewPayload(
        '   ',
        WebhookTemplateFormat::SIMPLE,
    );

    expect($result)->toBe([
        'success' => false,
        'output' => null,
        'errors' => ['Template cannot be empty.'],
    ]);
});

it('WebhookTemplateService_previewPayload_Ugly_surfaces_render_errors_for_interpolation_that_breaks_json', function () {
    $result = $this->service->previewPayload(
        '{"event": {{event.type}}}',
        WebhookTemplateFormat::JSON,
        'workspace.created',
    );

    expect($result['success'])->toBeFalse();
    expect($result['output'])->toBe('{"event": workspace.created}');
    expect($result['errors'])->toContain(
        'Rendered template is not valid JSON: Syntax error'
    );
});
