<?php

declare(strict_types=1);

namespace Core\Api\Resources;

use Illuminate\Http\Request;
use Illuminate\Http\Resources\Json\ResourceCollection;

/**
 * Base paginated collection with standard pagination metadata.
 *
 * Usage:
 *   return new PaginatedCollection($paginator, WorkspaceResource::class);
 */
class PaginatedCollection extends ResourceCollection
{
    /**
     * The resource class to use for items.
     */
    protected string $resourceClass;

    public function __construct($resource, string $resourceClass)
    {
        $this->resourceClass = $resourceClass;
        parent::__construct($resource);
    }

    public function toArray(Request $request): array
    {
        return [
            'data' => $this->resourceClass::collection($this->collection),
            'meta' => [
                'current_page' => $this->currentPage(),
                'from' => $this->firstItem(),
                'last_page' => $this->lastPage(),
                'per_page' => $this->perPage(),
                'to' => $this->lastItem(),
                'total' => $this->total(),
            ],
            'links' => [
                'first' => $this->url(1),
                'last' => $this->url($this->lastPage()),
                'prev' => $this->previousPageUrl(),
                'next' => $this->nextPageUrl(),
            ],
        ];
    }
}
