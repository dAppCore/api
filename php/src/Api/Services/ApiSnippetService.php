<?php

declare(strict_types=1);

namespace Core\Api\Services;

/**
 * API Code Snippet Generator - generates code snippets in multiple languages.
 *
 * Used to enhance API documentation with copy-paste ready examples.
 */
class ApiSnippetService
{
    /**
     * Supported languages with their display names.
     */
    public const LANGUAGES = [
        'curl' => 'cURL',
        'php' => 'PHP',
        'javascript' => 'JavaScript',
        'python' => 'Python',
        'ruby' => 'Ruby',
        'go' => 'Go',
        'java' => 'Java',
        'csharp' => 'C#',
        'swift' => 'Swift',
        'kotlin' => 'Kotlin',
        'rust' => 'Rust',
    ];

    /**
     * Generate snippets for all supported languages.
     */
    public function generateAll(
        string $method,
        string $endpoint,
        array $headers = [],
        ?array $body = null,
        string $baseUrl = 'https://api.host.uk.com'
    ): array {
        $snippets = [];

        foreach (array_keys(self::LANGUAGES) as $language) {
            $snippets[$language] = $this->generate($language, $method, $endpoint, $headers, $body, $baseUrl);
        }

        return $snippets;
    }

    /**
     * Generate a snippet for a specific language.
     */
    public function generate(
        string $language,
        string $method,
        string $endpoint,
        array $headers = [],
        ?array $body = null,
        string $baseUrl = 'https://api.host.uk.com'
    ): string {
        $url = rtrim($baseUrl, '/').'/'.ltrim($endpoint, '/');

        // Add default headers
        $headers = array_merge([
            'Authorization' => 'Bearer YOUR_API_KEY',
            'Content-Type' => 'application/json',
            'Accept' => 'application/json',
        ], $headers);

        return match ($language) {
            'curl' => $this->generateCurl($method, $url, $headers, $body),
            'php' => $this->generatePhp($method, $url, $headers, $body),
            'javascript' => $this->generateJavaScript($method, $url, $headers, $body),
            'python' => $this->generatePython($method, $url, $headers, $body),
            'ruby' => $this->generateRuby($method, $url, $headers, $body),
            'go' => $this->generateGo($method, $url, $headers, $body),
            'java' => $this->generateJava($method, $url, $headers, $body),
            'csharp' => $this->generateCSharp($method, $url, $headers, $body),
            'swift' => $this->generateSwift($method, $url, $headers, $body),
            'kotlin' => $this->generateKotlin($method, $url, $headers, $body),
            'rust' => $this->generateRust($method, $url, $headers, $body),
            default => "# Language '{$language}' not supported",
        };
    }

    protected function generateCurl(string $method, string $url, array $headers, ?array $body): string
    {
        $lines = ["curl -X {$method} '{$url}' \\"];

        foreach ($headers as $key => $value) {
            $lines[] = "  -H '{$key}: {$value}' \\";
        }

        if ($body) {
            $json = json_encode($body, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
            $lines[] = "  -d '{$json}'";
        } else {
            // Remove trailing backslash from last header
            $lastIndex = count($lines) - 1;
            $lines[$lastIndex] = rtrim($lines[$lastIndex], ' \\');
        }

        return implode("\n", $lines);
    }

    protected function generatePhp(string $method, string $url, array $headers, ?array $body): string
    {
        $headerStr = '';
        foreach ($headers as $key => $value) {
            $headerStr .= "    '{$key}' => '{$value}',\n";
        }

        $bodyStr = $body ? json_encode($body, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) : 'null';

        return <<<PHP
\$response = Http::{$this->phpMethod($method)}('{$url}', [
    'headers' => [
{$headerStr}    ],
    'json' => {$bodyStr},
]);

\$data = \$response->json();
PHP;
    }

    protected function generateJavaScript(string $method, string $url, array $headers, ?array $body): string
    {
        $headerJson = json_encode($headers, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
        $bodyJson = $body ? json_encode($body, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) : 'null';

        return <<<JS
const response = await fetch('{$url}', {
  method: '{$method}',
  headers: {$headerJson},
  body: {$bodyJson}
});

const data = await response.json();
JS;
    }

    protected function generatePython(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "    \"{$key}\": \"{$value}\"";
        }
        $headerStr = implode(",\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) : 'None';

        return <<<PYTHON
import requests

response = requests.{$this->pythonMethod($method)}(
    "{$url}",
    headers={
{$headerStr}
    },
    json={$bodyStr}
)

data = response.json()
PYTHON;
    }

    protected function generateRuby(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "  \"{$key}\" => \"{$value}\"";
        }
        $headerStr = implode(",\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : 'nil';

        return <<<RUBY
require 'httparty'

response = HTTParty.{$this->rubyMethod($method)}(
  "{$url}",
  headers: {
{$headerStr}
  },
  body: {$bodyStr}
)

data = JSON.parse(response.body)
RUBY;
    }

    protected function generateGo(string $method, string $url, array $headers, ?array $body): string
    {
        $bodySetup = $body
            ? 'jsonData, _ := json.Marshal(map[string]interface{}{'.$this->goMapEntries($body)."})\\n\\treq, _ := http.NewRequest(\"{$method}\", \"{$url}\", bytes.NewBuffer(jsonData))"
            : "req, _ := http.NewRequest(\"{$method}\", \"{$url}\", nil)";

        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "\treq.Header.Set(\"{$key}\", \"{$value}\")";
        }
        $headerStr = implode("\n", $headerLines);

        return <<<GO
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "io/ioutil"
)

func main() {
    {$bodySetup}
{$headerStr}

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    var data map[string]interface{}
    json.Unmarshal(body, &data)
}
GO;
    }

    protected function generateJava(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "    .header(\"{$key}\", \"{$value}\")";
        }
        $headerStr = implode("\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : '""';

        return <<<JAVA
HttpClient client = HttpClient.newHttpClient();
HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("{$url}"))
    .method("{$method}", HttpRequest.BodyPublishers.ofString("{$bodyStr}"))
{$headerStr}
    .build();

HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
String body = response.body();
JAVA;
    }

    protected function generateCSharp(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            if ($key === 'Content-Type') {
                continue;
            }
            $headerLines[] = "client.DefaultRequestHeaders.Add(\"{$key}\", \"{$value}\");";
        }
        $headerStr = implode("\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : '""';

        return <<<CSHARP
using var client = new HttpClient();
{$headerStr}

var content = new StringContent("{$bodyStr}", Encoding.UTF8, "application/json");
var response = await client.{$this->csharpMethod($method)}Async("{$url}", content);

var body = await response.Content.ReadAsStringAsync();
CSHARP;
    }

    protected function generateSwift(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "request.setValue(\"{$value}\", forHTTPHeaderField: \"{$key}\")";
        }
        $headerStr = implode("\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : 'nil';

        return <<<SWIFT
var request = URLRequest(url: URL(string: "{$url}")!)
request.httpMethod = "{$method}"
{$headerStr}
request.httpBody = "{$bodyStr}".data(using: .utf8)

let task = URLSession.shared.dataTask(with: request) { data, response, error in
    if let data = data {
        let json = try? JSONSerialization.jsonObject(with: data)
    }
}
task.resume()
SWIFT;
    }

    protected function generateKotlin(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "    .addHeader(\"{$key}\", \"{$value}\")";
        }
        $headerStr = implode("\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : '""';

        return <<<KOTLIN
val client = OkHttpClient()
val mediaType = "application/json".toMediaType()
val body = "{$bodyStr}".toRequestBody(mediaType)

val request = Request.Builder()
    .url("{$url}")
    .method("{$method}", body)
{$headerStr}
    .build()

val response = client.newCall(request).execute()
val json = response.body?.string()
KOTLIN;
    }

    protected function generateRust(string $method, string $url, array $headers, ?array $body): string
    {
        $headerLines = [];
        foreach ($headers as $key => $value) {
            $headerLines[] = "    .header(\"{$key}\", \"{$value}\")";
        }
        $headerStr = implode("\n", $headerLines);

        $bodyStr = $body ? json_encode($body, JSON_UNESCAPED_SLASHES) : '""';

        return <<<RUST
use reqwest::blocking::Client;

let client = Client::new();
let response = client
    .{$this->rustMethod($method)}("{$url}")
{$headerStr}
    .body("{$bodyStr}")
    .send()?;

let json: serde_json::Value = response.json()?;
RUST;
    }

    // Helper methods for language-specific syntax
    protected function phpMethod(string $method): string
    {
        return strtolower($method);
    }

    protected function pythonMethod(string $method): string
    {
        return strtolower($method);
    }

    protected function rubyMethod(string $method): string
    {
        return strtolower($method);
    }

    protected function csharpMethod(string $method): string
    {
        return match (strtoupper($method)) {
            'GET' => 'Get',
            'POST' => 'Post',
            'PUT' => 'Put',
            'PATCH' => 'Patch',
            'DELETE' => 'Delete',
            default => 'Send',
        };
    }

    protected function rustMethod(string $method): string
    {
        return strtolower($method);
    }

    protected function goMapEntries(array $data): string
    {
        $entries = [];
        foreach ($data as $key => $value) {
            $val = is_string($value) ? "\"{$value}\"" : json_encode($value);
            $entries[] = "\"{$key}\": {$val}";
        }

        return implode(', ', $entries);
    }

    /**
     * Get language metadata for UI display.
     */
    public static function getLanguages(): array
    {
        return collect(self::LANGUAGES)->map(fn ($name, $code) => [
            'code' => $code,
            'name' => $name,
            'icon' => self::getLanguageIcon($code),
        ])->values()->all();
    }

    /**
     * Get icon class for a language.
     */
    public static function getLanguageIcon(string $code): string
    {
        return match ($code) {
            'curl' => 'terminal',
            'php' => 'code-bracket',
            'javascript' => 'code-bracket-square',
            'python' => 'code-bracket',
            'ruby' => 'sparkles',
            'go' => 'cube',
            'java' => 'fire',
            'csharp' => 'window',
            'swift' => 'bolt',
            'kotlin' => 'beaker',
            'rust' => 'cog',
            default => 'code-bracket',
        };
    }
}
