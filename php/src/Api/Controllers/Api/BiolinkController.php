<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Controllers\Api\Concerns\SerialisesWorkspaceResource;
use Core\Api\Models\Biolink;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class BiolinkController extends Controller
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

        $biolinks = Biolink::query()
            ->forWorkspace($workspace->id)
            ->latest()
            ->paginate((int) min($request->integer('per_page', 25), 100));

        return response()->json([
            'data' => $biolinks->getCollection()->map(fn (Biolink $biolink) => $this->modelPayload($biolink))->values()->all(),
            'meta' => [
                'current_page' => $biolinks->currentPage(),
                'per_page' => $biolinks->perPage(),
                'total' => $biolinks->total(),
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
            'title' => ['required', 'string', 'max:255'],
            'slug' => ['sometimes', 'nullable', 'string', 'max:255'],
            'url' => ['required', 'url', 'max:2048'],
            'description' => ['sometimes', 'nullable', 'string'],
            'is_published' => ['sometimes', 'boolean'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $biolink = Biolink::query()->create([
            'workspace_id' => $workspace->id,
            'user_id' => $request->user()?->id,
            'title' => $data['title'],
            'slug' => $data['slug'] ?? Str::slug($data['title']),
            'url' => $data['url'],
            'description' => $data['description'] ?? null,
            'is_published' => $data['is_published'] ?? true,
            'metadata' => $data['metadata'] ?? [],
        ]);

        return $this->createdResponse($this->modelPayload($biolink), 'Biolink created successfully.');
    }

    public function show(Request $request, string $workspace, string $id): JsonResponse
    {
        $biolink = $this->findBiolink($request, $id);
        if (! $biolink instanceof Biolink) {
            return $this->notFoundResponse('Biolink');
        }

        return response()->json(['data' => $this->modelPayload($biolink)]);
    }

    public function update(Request $request, string $workspace, string $id): JsonResponse
    {
        $biolink = $this->findBiolink($request, $id);
        if (! $biolink instanceof Biolink) {
            return $this->notFoundResponse('Biolink');
        }

        $data = $request->validate([
            'title' => ['sometimes', 'string', 'max:255'],
            'slug' => ['sometimes', 'nullable', 'string', 'max:255'],
            'url' => ['sometimes', 'url', 'max:2048'],
            'description' => ['sometimes', 'nullable', 'string'],
            'is_published' => ['sometimes', 'boolean'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $biolink->fill($data);
        if (array_key_exists('title', $data) && ! array_key_exists('slug', $data) && ! $biolink->slug) {
            $biolink->slug = Str::slug((string) $data['title']);
        }
        $biolink->save();

        $freshBiolink = $biolink->fresh();
        if ($freshBiolink instanceof Biolink) {
            $biolink = $freshBiolink;
        }

        return response()->json(['data' => $this->modelPayload($biolink)]);
    }

    public function destroy(Request $request, string $workspace, string $id): JsonResponse
    {
        $biolink = $this->findBiolink($request, $id);
        if (! $biolink instanceof Biolink) {
            return $this->notFoundResponse('Biolink');
        }

        $biolink->delete();

        return $this->successResponse('Biolink deleted successfully.');
    }

    protected function findBiolink(Request $request, string $id): ?Biolink
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return null;
        }

        return Biolink::query()
            ->forWorkspace($workspace->id)
            ->find($id);
    }
}
