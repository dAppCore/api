<?php

declare(strict_types=1);

namespace Core\Api\Controllers\Api;

use Core\Api\Concerns\HasApiResponses;
use Core\Api\Concerns\ResolvesWorkspace;
use Core\Front\Controller;
use Core\Mod\Commerce\Models\PaymentMethod;
use Core\Mod\Commerce\Services\PaymentMethodService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Payment method management endpoints.
 */
class PaymentMethodController extends Controller
{
    use HasApiResponses;
    use ResolvesWorkspace;

    public function __construct(
        protected PaymentMethodService $service
    ) {
    }

    /**
     * List payment methods for the current workspace.
     *
     * GET /api/v1/commerce/payment-methods
     */
    public function index(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $methods = $this->service->getPaymentMethods($workspace);

        return response()->json([
            'workspace_id' => $workspace->id,
            'payment_methods' => $methods->map(fn (PaymentMethod $method) => $this->serialize($method))->values()->all(),
            'count' => $methods->count(),
        ]);
    }

    /**
     * Add a payment method to the current workspace.
     *
     * POST /api/v1/commerce/payment-methods
     */
    public function store(Request $request): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $data = $request->validate([
            'gateway_payment_method_id' => ['required', 'string', 'max:255'],
            'gateway' => ['sometimes', 'string', 'max:50'],
        ]);

        $paymentMethod = $this->service->addPaymentMethod(
            workspace: $workspace,
            gatewayPaymentMethodId: $data['gateway_payment_method_id'],
            user: $request->user(),
            gateway: $data['gateway'] ?? 'stripe',
        );
        $paymentMethod = $this->refreshPaymentMethod($paymentMethod);

        return response()->json([
            'payment_method' => $this->serialize($paymentMethod),
        ], 201);
    }

    /**
     * Remove a payment method from the current workspace.
     *
     * DELETE /api/v1/commerce/payment-methods/{id}
     */
    public function destroy(Request $request, string $id): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $paymentMethod = PaymentMethod::query()
            ->where('workspace_id', $workspace->id)
            ->find($id);

        if (! $paymentMethod instanceof PaymentMethod) {
            return $this->notFoundResponse('Payment method');
        }

        try {
            $this->service->removePaymentMethod($workspace, $paymentMethod);
        } catch (\RuntimeException $e) {
            return $this->errorResponse(
                errorCode: 'invalid_status',
                message: $e->getMessage(),
                status: 422,
            );
        }

        return $this->successResponse('Payment method removed successfully.');
    }

    /**
     * Set a payment method as the workspace default.
     *
     * POST /api/v1/commerce/payment-methods/{id}/default
     */
    public function default(Request $request, string $id): JsonResponse
    {
        $workspace = $this->resolveWorkspace($request);
        if ($workspace === null) {
            return $this->noWorkspaceResponse();
        }

        $paymentMethod = PaymentMethod::query()
            ->where('workspace_id', $workspace->id)
            ->find($id);

        if (! $paymentMethod instanceof PaymentMethod) {
            return $this->notFoundResponse('Payment method');
        }

        $this->service->setDefaultPaymentMethod($workspace, $paymentMethod);
        $paymentMethod = $this->refreshPaymentMethod($paymentMethod);

        return response()->json([
            'payment_method' => $this->serialize($paymentMethod),
        ]);
    }

    /**
     * Serialise a payment method for API responses.
     */
    protected function serialize(PaymentMethod $paymentMethod): array
    {
        return [
            'id' => $paymentMethod->id,
            'workspace_id' => $paymentMethod->workspace_id,
            'gateway' => $paymentMethod->gateway,
            'type' => $paymentMethod->type,
            'brand' => $paymentMethod->brand,
            'last_four' => $paymentMethod->last_four,
            'exp_month' => $paymentMethod->exp_month,
            'exp_year' => $paymentMethod->exp_year,
            'is_default' => $paymentMethod->is_default,
            'is_active' => $paymentMethod->is_active,
            'display_name' => $paymentMethod->getDisplayName(),
            'created_at' => $paymentMethod->created_at?->toIso8601String(),
            'updated_at' => $paymentMethod->updated_at?->toIso8601String(),
        ];
    }

    /**
     * Refresh a payment method if it still exists.
     */
    protected function refreshPaymentMethod(PaymentMethod $paymentMethod): PaymentMethod
    {
        $paymentMethod->refresh();

        return $paymentMethod;
    }
}
