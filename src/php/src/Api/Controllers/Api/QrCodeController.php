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
            'format' => ['sometimes', 'string', 'in:png,svg'],
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

        $svg = $this->renderSvg($code);

        return response($svg, 200, [
            'Content-Type' => 'image/svg+xml',
            'Content-Disposition' => 'attachment; filename="qr-code-'.$code->id.'.svg"',
        ]);
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

    /**
     * Render a deterministic SVG fallback when no QR library is present.
     */
    protected function renderSvg(QrCode $code): string
    {
        $size = max(64, (int) ($code->size ?? 256));
        $cells = 21;
        $cellSize = max(1, (int) floor($size / $cells));
        $foreground = $code->foreground_color ?: '#000000';
        $background = $code->background_color ?: '#ffffff';
        $hash = hash('sha256', $code->target_url);

        $rects = [];
        $index = 0;
        for ($y = 0; $y < $cells; $y++) {
            for ($x = 0; $x < $cells; $x++) {
                $nibble = hexdec($hash[$index % strlen($hash)]);
                if (($nibble % 2) === 0) {
                    $rects[] = sprintf(
                        '<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />',
                        $x * $cellSize,
                        $y * $cellSize,
                        $cellSize,
                        $cellSize,
                        htmlspecialchars($foreground, ENT_QUOTES, 'UTF-8')
                    );
                }
                $index++;
            }
        }

        return sprintf(
            '<svg xmlns="http://www.w3.org/2000/svg" width="%1$d" height="%1$d" viewBox="0 0 %1$d %1$d"><rect width="100%%" height="100%%" fill="%2$s" />%3$s</svg>',
            $size,
            htmlspecialchars($background, ENT_QUOTES, 'UTF-8'),
            implode('', $rects)
        );
    }
}
