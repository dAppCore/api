<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Api\Controllers\Api\Concerns\SerialisesWorkspaceResource;
use Core\Api\Models\QrCode;
use Core\Front\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class QrCodeController extends Controller
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

        $codes = QrCode::query()
            ->forWorkspace($workspace->id)
            ->latest()
            ->paginate((int) min($request->integer('per_page', 25), 100));

        return response()->json([
            'data' => $codes->getCollection()->map(fn (QrCode $code) => $this->modelPayload($code))->values()->all(),
            'meta' => [
                'current_page' => $codes->currentPage(),
                'per_page' => $codes->perPage(),
                'total' => $codes->total(),
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
            'target_url' => ['required', 'url', 'max:2048'],
            // png is intentionally not accepted: download() cannot produce
            // PNG output until a real QR encoder is wired in. SVG remains
            // accepted at the record level so callers can still register
            // QR codes; download() returns 501 Not Implemented.
            'format' => ['sometimes', 'string', 'in:svg'],
            'size' => ['sometimes', 'integer', 'min:64', 'max:2048'],
            'foreground_color' => ['sometimes', 'string', 'max:32'],
            'background_color' => ['sometimes', 'string', 'max:32'],
            'metadata' => ['sometimes', 'array'],
        ]);

        $code = QrCode::query()->create([
            'workspace_id' => $workspace->id,
            'user_id' => $request->user()?->id,
            'name' => $data['name'],
            'target_url' => $data['target_url'],
            'format' => $data['format'] ?? 'svg',
            'size' => $data['size'] ?? 256,
            'foreground_color' => $data['foreground_color'] ?? '#000000',
            'background_color' => $data['background_color'] ?? '#ffffff',
            'metadata' => $data['metadata'] ?? [],
        ]);

        return $this->createdResponse($this->modelPayload($code), 'QR code created successfully.');
    }

    public function show(Request $request, string $workspace, string $id): JsonResponse
    {
        $code = $this->findCode($request, $id);
        if (! $code instanceof QrCode) {
            return $this->notFoundResponse('QR code');
        }

        return response()->json(['data' => $this->modelPayload($code)]);
    }

    public function download(Request $request, string $workspace, string $id): Response|JsonResponse
    {
        $code = $this->findCode($request, $id);
        if (! $code instanceof QrCode) {
            return $this->notFoundResponse('QR code');
        }

        // No real QR encoder is wired in yet — renderSvg only emits a hash-
        // derived placeholder pattern that scanners cannot decode. Fail loud
        // rather than serve a broken SVG that mislabels itself as a QR code.
        // TODO: integrate a real QR encoder (e.g. endroid/qr-code or
        // bacon/bacon-qr-code) and return the encoded SVG/PNG here.
        return response()->json([
            'message' => 'QR code rendering is not yet implemented.',
            'code' => $code->id,
        ], 501);
    }

    protected function findCode(Request $request, string $id): ?QrCode
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return null;
        }

        return QrCode::query()
            ->forWorkspace($workspace->id)
            ->find($id);
    }
}
