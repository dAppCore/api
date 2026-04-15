<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Controllers\Api\Concerns\SerialisesWorkspaceResource;
use Core\Api\Models\Link;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class LinkController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;
    use SerialisesWorkspaceResource;

    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $links = Link::query()
            ->forWorkspace($workspace->id)
            ->latest()
            ->paginate((int) min($request->integer('per_page', 25), 100));

        return response()->json([
            'data' => $links->getCollection()->map(fn (Link $link) => $this->modelPayload($link))->values()->all(),
            'meta' => [
                'current_page' => $links->currentPage(),
                'per_page' => $links->perPage(),
                'total' => $links->total(),
            ],
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'slug' => ['sometimes', 'nullable', 'string', 'max:255'],
            'destination_url' => ['required', 'url', 'max:2048'],
            'short_code' => ['sometimes', 'nullable', 'string', 'max:64'],
            'is_active' => ['sometimes', 'boolean'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $link = Link::query()->create([
            'workspace_id' => $workspace->id,
            'user_id' => $request->user()?->id,
            'name' => $data['name'],
            'slug' => $data['slug'] ?? Str::slug($data['name']),
            'destination_url' => $data['destination_url'],
            'short_code' => $data['short_code'] ?? Str::lower(Str::random(8)),
            'is_active' => $data['is_active'] ?? true,
            'click_count' => 0,
            'metadata' => $data['metadata'] ?? [],
        ]);

        return $this->createdResponse($this->modelPayload($link), 'Link created successfully.');
    }

    public function show(Request $request, string $workspace, string $id): JsonResponse
    {
        $link = $this->findLink($request, $id);
        if (! $link instanceof Link) {
            return $this->notFoundResponse('Link');
        }

        return response()->json(['data' => $this->modelPayload($link)]);
    }

    public function update(Request $request, string $workspace, string $id): JsonResponse
    {
        $link = $this->findLink($request, $id);
        if (! $link instanceof Link) {
            return $this->notFoundResponse('Link');
        }

        $data = $request->validate([
            'name' => ['sometimes', 'string', 'max:255'],
            'slug' => ['sometimes', 'nullable', 'string', 'max:255'],
            'destination_url' => ['sometimes', 'url', 'max:2048'],
            'short_code' => ['sometimes', 'nullable', 'string', 'max:64'],
            'is_active' => ['sometimes', 'boolean'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $link->fill($data);
        $link->save();

        return response()->json(['data' => $this->modelPayload($link->fresh())]);
    }

    public function destroy(Request $request, string $workspace, string $id): JsonResponse
    {
        $link = $this->findLink($request, $id);
        if (! $link instanceof Link) {
            return $this->notFoundResponse('Link');
        }

        $link->delete();

        return $this->successResponse('Link deleted successfully.');
    }

    public function stats(Request $request, string $workspace, string $id): JsonResponse
    {
        $link = $this->findLink($request, $id);
        if (! $link instanceof Link) {
            return $this->notFoundResponse('Link');
        }

        $metadata = $link->metadata ?? [];

        return response()->json([
            'data' => [
                'id' => $link->id,
                'workspace_id' => $link->workspace_id,
                'click_count' => $link->click_count ?? 0,
                'last_clicked_at' => $link->last_clicked_at?->toIso8601String(),
                'stats' => is_array($metadata['stats'] ?? null) ? $metadata['stats'] : [],
            ],
        ]);
    }

    protected function findLink(Request $request, string $id): ?Link
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return null;
        }

        return Link::query()
            ->forWorkspace($workspace->id)
            ->find($id);
    }
}
