<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Controllers\Api\Concerns\SerialisesWorkspaceResource;
use Core\Api\Models\SupportTicket;
use Core\Api\Models\SupportTicketReply;
use Core\Front\Controller;
use Illuminate\Auth\Access\AuthorizationException;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Log;

class TicketController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;
    use SerialisesWorkspaceResource;

    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        $user = $request->user();

        if ($workspace === null && $user === null) {
            return $this->forbiddenResponse('Authentication or workspace context required.');
        }

        $query = SupportTicket::query()
            ->with('replies')
            ->latest();

        if ($workspace !== null) {
            $query->forWorkspace($workspace->id);
        }

        if ($user?->id !== null) {
            $query->where('user_id', $user->id);
        }

        $tickets = $query->paginate((int) min($request->integer('per_page', 25), 100));

        return response()->json([
            'data' => $tickets->getCollection()->map(fn (SupportTicket $ticket) => $this->ticketPayload($ticket))->values()->all(),
            'meta' => [
                'current_page' => $tickets->currentPage(),
                'per_page' => $tickets->perPage(),
                'total' => $tickets->total(),
            ],
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);

        $data = $request->validate([
            'subject' => ['required', 'string', 'max:255'],
            'message' => ['required', 'string'],
            'priority' => ['sometimes', 'string', 'in:low,normal,high,urgent'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $ticket = SupportTicket::query()->create([
            'workspace_id' => $workspace?->id,
            'user_id' => $request->user()?->id,
            'subject' => $data['subject'],
            'message' => $data['message'],
            'status' => 'open',
            'priority' => $data['priority'] ?? 'normal',
            'metadata' => $data['metadata'] ?? [],
            'last_replied_at' => now(),
        ]);

        return $this->createdResponse($this->ticketPayload($ticket), 'Support ticket created successfully.');
    }

    public function show(Request $request, string $id): JsonResponse
    {
        $ticket = $this->findTicket($request, $id);
        if (! $ticket instanceof SupportTicket) {
            return $this->notFoundResponse('Support ticket');
        }

        return response()->json(['data' => $this->ticketPayload($ticket->load('replies'))]);
    }

    public function reply(Request $request, string $id): JsonResponse
    {
        $ticket = $this->findTicket($request, $id);
        if (! $ticket instanceof SupportTicket) {
            return $this->notFoundResponse('Support ticket');
        }

        $data = $request->validate([
            'message' => ['required', 'string'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $reply = SupportTicketReply::query()->create([
            'ticket_id' => $ticket->id,
            'user_id' => $request->user()?->id,
            'message' => $data['message'],
            'is_staff' => false,
            'metadata' => $data['metadata'] ?? [],
        ]);

        $ticket->forceFill([
            'status' => 'open',
            'last_replied_at' => now(),
        ])->save();

        $ticket->refresh();
        $ticket->load('replies');

        return response()->json([
            'data' => [
                'ticket' => $this->ticketPayload($ticket),
                'reply' => $reply->attributesToArray(),
            ],
        ]);
    }

    protected function findTicket(Request $request, string $id): ?SupportTicket
    {
        $query = SupportTicket::query()->with('replies');

        $workspace = $this->resolveWorkspace($request);
        $user = $request->user();

        if ($workspace === null && $user === null) {
            Log::warning('TicketController.findTicket fail-open attempt', [
                'ticket_id' => $id,
                'actor_ip' => $request->ip(),
                'route' => $request->path(),
            ]);

            throw new AuthorizationException('Authentication context required');
        }

        if ($workspace !== null) {
            $query->forWorkspace($workspace->id);
        }

        if ($user?->id !== null) {
            $query->where('user_id', $user->id);
        }

        return $query->find($id);
    }

    protected function ticketPayload(SupportTicket $ticket): array
    {
        $payload = $this->modelPayload($ticket);
        $payload['replies'] = $ticket->relationLoaded('replies')
            ? $ticket->replies->map(fn (SupportTicketReply $reply) => $reply->attributesToArray())->values()->all()
            : [];

        return $payload;
    }
}
