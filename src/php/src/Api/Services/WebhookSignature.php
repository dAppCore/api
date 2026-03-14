<?php

declare(strict_types=1);

namespace Core\Api\Services;

use Illuminate\Support\Str;

/**
 * Webhook Signature Service - handles HMAC signing and verification for outbound webhooks.
 *
 * This service provides cryptographic signing for webhook payloads to ensure:
 * 1. **Authenticity**: Recipients can verify the request came from our platform
 * 2. **Integrity**: Recipients can verify the payload wasn't tampered with
 * 3. **Replay Protection**: Timestamps prevent replay attacks
 *
 * ## Signature Algorithm
 *
 * The signature is computed as:
 * ```
 * signature = HMAC-SHA256(timestamp + "." + payload, secret)
 * ```
 *
 * Including the timestamp in the signed data prevents replay attacks where an
 * attacker could capture a valid webhook and resend it later.
 *
 * ## Verification Example (for webhook recipients)
 *
 * ```php
 * // Get headers and body from the request
 * $signature = $request->header('X-Webhook-Signature');
 * $timestamp = $request->header('X-Webhook-Timestamp');
 * $payload = $request->getContent();
 *
 * // Compute expected signature
 * $expectedSignature = hash_hmac('sha256', $timestamp . '.' . $payload, $webhookSecret);
 *
 * // Verify signature using timing-safe comparison
 * if (!hash_equals($expectedSignature, $signature)) {
 *     abort(401, 'Invalid webhook signature');
 * }
 *
 * // Verify timestamp is within tolerance (e.g., 5 minutes)
 * $tolerance = 300; // seconds
 * if (abs(time() - (int)$timestamp) > $tolerance) {
 *     abort(401, 'Webhook timestamp too old');
 * }
 * ```
 */
class WebhookSignature
{
    /**
     * Default secret length in bytes (64 characters when hex-encoded).
     */
    private const SECRET_LENGTH = 32;

    /**
     * Default tolerance for timestamp verification in seconds.
     * 5 minutes allows for reasonable clock skew and network delays.
     */
    public const DEFAULT_TOLERANCE = 300;

    /**
     * The hashing algorithm used for HMAC.
     */
    private const ALGORITHM = 'sha256';

    /**
     * Generate a cryptographically secure webhook signing secret.
     *
     * The secret is a 64-character random string suitable for HMAC-SHA256 signing.
     * This should be stored securely and shared with the webhook recipient out-of-band.
     *
     * @return string A 64-character random string
     */
    public function generateSecret(): string
    {
        return Str::random(64);
    }

    /**
     * Sign a webhook payload with the given secret and timestamp.
     *
     * The signature format is:
     * HMAC-SHA256(timestamp + "." + payload, secret)
     *
     * This format ensures the timestamp cannot be changed without invalidating
     * the signature, providing replay attack protection.
     *
     * @param  string  $payload  The JSON-encoded webhook payload
     * @param  string  $secret  The endpoint's signing secret
     * @param  int  $timestamp  Unix timestamp of when the webhook was sent
     * @return string The HMAC-SHA256 signature (hex-encoded, 64 characters)
     */
    public function sign(string $payload, string $secret, int $timestamp): string
    {
        $signedPayload = $this->buildSignedPayload($timestamp, $payload);

        return hash_hmac(self::ALGORITHM, $signedPayload, $secret);
    }

    /**
     * Verify a webhook signature.
     *
     * Performs a timing-safe comparison to prevent timing attacks, and optionally
     * validates that the timestamp is within the specified tolerance.
     *
     * @param  string  $payload  The raw request body (JSON string)
     * @param  string  $signature  The signature from X-Webhook-Signature header
     * @param  string  $secret  The webhook endpoint's secret
     * @param  int  $timestamp  The timestamp from X-Webhook-Timestamp header
     * @param  int  $tolerance  Maximum age of the timestamp in seconds (default: 300)
     * @return bool True if the signature is valid and timestamp is within tolerance
     */
    public function verify(
        string $payload,
        string $signature,
        string $secret,
        int $timestamp,
        int $tolerance = self::DEFAULT_TOLERANCE
    ): bool {
        // Check timestamp is within tolerance
        if (! $this->isTimestampValid($timestamp, $tolerance)) {
            return false;
        }

        // Compute expected signature
        $expectedSignature = $this->sign($payload, $secret, $timestamp);

        // Use timing-safe comparison to prevent timing attacks
        return hash_equals($expectedSignature, $signature);
    }

    /**
     * Verify signature without timestamp validation.
     *
     * Use this method when you need to verify the signature but handle
     * timestamp validation separately (e.g., for testing or special cases).
     *
     * @param  string  $payload  The raw request body
     * @param  string  $signature  The signature from the header
     * @param  string  $secret  The webhook secret
     * @param  int  $timestamp  The timestamp from the header
     * @return bool True if the signature is valid
     */
    public function verifySignatureOnly(
        string $payload,
        string $signature,
        string $secret,
        int $timestamp
    ): bool {
        $expectedSignature = $this->sign($payload, $secret, $timestamp);

        return hash_equals($expectedSignature, $signature);
    }

    /**
     * Check if a timestamp is within the allowed tolerance.
     *
     * @param  int  $timestamp  The Unix timestamp to check
     * @param  int  $tolerance  Maximum age in seconds
     * @return bool True if the timestamp is within tolerance
     */
    public function isTimestampValid(int $timestamp, int $tolerance = self::DEFAULT_TOLERANCE): bool
    {
        $now = time();

        return abs($now - $timestamp) <= $tolerance;
    }

    /**
     * Build the signed payload string.
     *
     * Format: "{timestamp}.{payload}"
     *
     * @param  int  $timestamp  Unix timestamp
     * @param  string  $payload  The JSON payload
     * @return string The combined string to be signed
     */
    private function buildSignedPayload(int $timestamp, string $payload): string
    {
        return $timestamp.'.'.$payload;
    }

    /**
     * Get the headers to include with a webhook request.
     *
     * Returns an array of headers ready to be used with HTTP client:
     * - X-Webhook-Signature: The HMAC signature
     * - X-Webhook-Timestamp: Unix timestamp
     *
     * @param  string  $payload  The JSON-encoded payload
     * @param  string  $secret  The signing secret
     * @param  int|null  $timestamp  Unix timestamp (defaults to current time)
     * @return array<string, string|int> Headers array
     */
    public function getHeaders(string $payload, string $secret, ?int $timestamp = null): array
    {
        $timestamp ??= time();

        return [
            'X-Webhook-Signature' => $this->sign($payload, $secret, $timestamp),
            'X-Webhook-Timestamp' => $timestamp,
        ];
    }
}
