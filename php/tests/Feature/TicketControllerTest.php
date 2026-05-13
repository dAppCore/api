<?php

declare(strict_types=1);

use Core\Api\Controllers\Api\TicketController;
use Core\Api\Models\SupportTicket;
use Core\Tenant\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Route;

uses(RefreshDatabase::class);

beforeEach(function () {
    Route::get('/api/test-support/tickets/{id}', [TicketController::class, 'show']);
});

function ticketControllerTestTicket(array $attributes = []): SupportTicket
{
    return SupportTicket::query()->create(array_merge([
        'workspace_id' => null,
        'user_id' => null,
        'subject' => 'Private support issue',
        'message' => 'Customer-only support conversation',
        'status' => 'open',
        'priority' => 'normal',
        'metadata' => [],
        'last_replied_at' => now(),
    ], $attributes));
}

it('TicketController_findTicket_AnonymousAccess_Bad_blocks_unscoped_lookup', function () {
    $ticket = ticketControllerTestTicket();

    $response = $this->getJson("/api/test-support/tickets/{$ticket->id}");

    $response
        ->assertStatus(403)
        ->assertJsonMissing(['subject' => 'Private support issue']);
});

it('TicketController_findTicket_FailOpenAttempt_Ugly_logs_warning_context', function () {
    $ticket = ticketControllerTestTicket();

    Log::shouldReceive('warning')
        ->once()
        ->with('TicketController.findTicket fail-open attempt', \Mockery::on(function (array $context) use ($ticket): bool {
            return ($context['ticket_id'] ?? null) === (string) $ticket->id
                && isset($context['actor_ip'])
                && ($context['route'] ?? null) === "api/test-support/tickets/{$ticket->id}";
        }));

    $this->getJson("/api/test-support/tickets/{$ticket->id}")
        ->assertStatus(403);
});

it('TicketController_findTicket_AuthenticatedUser_Good_returns_owned_ticket', function () {
    $user = User::query()->create([
        'name' => fake()->name(),
        'email' => fake()->unique()->safeEmail(),
        'password' => 'password',
    ]);
    $ticket = ticketControllerTestTicket([
        'user_id' => $user->id,
        'subject' => 'Owned support issue',
    ]);

    $this->actingAs($user);

    $response = $this->getJson("/api/test-support/tickets/{$ticket->id}");

    $response
        ->assertOk()
        ->assertJsonPath('data.id', $ticket->id)
        ->assertJsonPath('data.subject', 'Owned support issue');
});
