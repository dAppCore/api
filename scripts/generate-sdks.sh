#!/bin/bash

set -euo pipefail

OPENAPI_FILE="storage/app/openapi.json"
OUTPUT_DIR="./sdks"

# Export a fresh OpenAPI document before generating SDKs.
php artisan scramble:export

SDKS=(
  "php"
  "typescript"
  "python"
  "go"
  "rust"
  "ruby"
  "java"
  "csharp"
  "kotlin"
  "swift"
  "dart"
)

for sdk in "${SDKS[@]}"; do
  echo "Generating ${sdk} SDK..."
  openapi-generator-cli generate \
    -i "${OPENAPI_FILE}" \
    -c "sdk-config/${sdk}.yaml" \
    -o "${OUTPUT_DIR}/${sdk}"
done

echo "All SDKs generated in ${OUTPUT_DIR}"
