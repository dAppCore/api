<?php

declare(strict_types=1);

namespace Core\Api\Concerns;

use Illuminate\Http\JsonResponse;

/**
 * Standardised API response helpers.
 *
 * Provides consistent error response format across all API endpoints.
 */
trait HasApiResponses
{
    /**
     * Return a standard error response.
     */
    protected function errorResponse(
        string $errorCode,
        string $message,
        array $meta = [],
        int $status = 400,
    ): JsonResponse {
        $response = [
            'success' => false,
            'error' => $errorCode,
            'message' => $message,
            'error_code' => $errorCode,
        ];

        if ($meta !== []) {
            $response['details'] = $meta;
        }

        return response()->json(array_merge($response, $meta), $status);
    }

    /**
     * Return a no workspace response.
     */
    protected function noWorkspaceResponse(): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'no_workspace',
            message: 'No workspace found. Please select a workspace first.',
            status: 404,
        );
    }

    /**
     * Return a resource not found response.
     */
    protected function notFoundResponse(string $resource = 'Resource'): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'not_found',
            message: "{$resource} not found.",
            meta: [
                'resource' => $resource,
            ],
            status: 404,
        );
    }

    /**
     * Return a feature limit reached response.
     */
    protected function limitReachedResponse(string $feature, ?string $message = null): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'entitlement_exceeded',
            message: $message ?? 'You have reached your limit for this feature.',
            meta: [
                'feature' => $feature,
                'upgrade_url' => route('hub.usage'),
            ],
            status: 403,
        );
    }

    /**
     * Return an access denied response.
     */
    protected function accessDeniedResponse(string $message = 'Access denied.'): JsonResponse
    {
        return $this->forbiddenResponse($message, status: 403);
    }

    /**
     * Return a forbidden response.
     */
    protected function forbiddenResponse(string $message, array $meta = [], int $status = 403): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'forbidden',
            message: $message,
            meta: $meta,
            status: $status,
        );
    }

    /**
     * Return a success response with message.
     */
    protected function successResponse(string $message, array $data = []): JsonResponse
    {
        return response()->json(array_merge([
            'success' => true,
            'message' => $message,
        ], $data));
    }

    /**
     * Return a created response.
     */
    protected function createdResponse(mixed $resource, string $message = 'Created successfully.'): JsonResponse
    {
        return response()->json([
            'success' => true,
            'message' => $message,
            'data' => $resource,
        ], 201);
    }

    /**
     * Return a validation error response.
     */
    protected function validationErrorResponse(array $errors, int $status = 422): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'validation_error',
            message: 'The given data was invalid.',
            meta: [
                'errors' => $errors,
            ],
            status: $status,
        );
    }

    /**
     * Return an invalid status error response.
     *
     * Used when an operation cannot be performed due to the resource's current status.
     */
    protected function invalidStatusResponse(string $message): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'invalid_status',
            message: $message,
            status: 422,
        );
    }

    /**
     * Return a provider error response.
     *
     * Used when an external provider operation fails.
     */
    protected function providerErrorResponse(string $message, ?string $provider = null): JsonResponse
    {
        return $this->errorResponse(
            errorCode: 'provider_error',
            message: $message,
            meta: array_filter([
                'provider' => $provider,
            ]),
            status: 400,
        );
    }
}
