<?php

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
     * Return a no workspace response.
     */
    protected function noWorkspaceResponse(): JsonResponse
    {
        return response()->json([
            'error' => 'no_workspace',
            'message' => 'No workspace found. Please select a workspace first.',
        ], 404);
    }

    /**
     * Return a resource not found response.
     */
    protected function notFoundResponse(string $resource = 'Resource'): JsonResponse
    {
        return response()->json([
            'error' => 'not_found',
            'message' => "{$resource} not found.",
        ], 404);
    }

    /**
     * Return a feature limit reached response.
     */
    protected function limitReachedResponse(string $feature, ?string $message = null): JsonResponse
    {
        return response()->json([
            'error' => 'feature_limit_reached',
            'message' => $message ?? 'You have reached your limit for this feature.',
            'feature' => $feature,
            'upgrade_url' => route('hub.usage'),
        ], 403);
    }

    /**
     * Return an access denied response.
     */
    protected function accessDeniedResponse(string $message = 'Access denied.'): JsonResponse
    {
        return response()->json([
            'error' => 'access_denied',
            'message' => $message,
        ], 403);
    }

    /**
     * Return a success response with message.
     */
    protected function successResponse(string $message, array $data = []): JsonResponse
    {
        return response()->json(array_merge([
            'message' => $message,
        ], $data));
    }

    /**
     * Return a created response.
     */
    protected function createdResponse(mixed $resource, string $message = 'Created successfully.'): JsonResponse
    {
        return response()->json([
            'message' => $message,
            'data' => $resource,
        ], 201);
    }

    /**
     * Return a validation error response.
     */
    protected function validationErrorResponse(array $errors): JsonResponse
    {
        return response()->json([
            'error' => 'validation_failed',
            'message' => 'The given data was invalid.',
            'errors' => $errors,
        ], 422);
    }

    /**
     * Return an invalid status error response.
     *
     * Used when an operation cannot be performed due to the resource's current status.
     */
    protected function invalidStatusResponse(string $message): JsonResponse
    {
        return response()->json([
            'error' => 'invalid_status',
            'message' => $message,
        ], 422);
    }

    /**
     * Return a provider error response.
     *
     * Used when an external provider operation fails.
     */
    protected function providerErrorResponse(string $message, ?string $provider = null): JsonResponse
    {
        $response = [
            'error' => 'provider_error',
            'message' => $message,
        ];

        if ($provider !== null) {
            $response['provider'] = $provider;
        }

        return response()->json($response, 400);
    }
}
