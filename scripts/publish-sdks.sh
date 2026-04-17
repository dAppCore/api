#!/bin/bash

set -euo pipefail

VERSION="${1:-}"

if [ -z "${VERSION}" ]; then
  echo "Usage: ./scripts/publish-sdks.sh <version>"
  exit 1
fi

echo "PHP: publish via Packagist automation"

cd sdks/typescript
npm version "${VERSION}"
npm publish --access public
cd ../..

cd sdks/python
rm -rf dist
python -m build
twine upload dist/*
cd ../..

cd sdks/go
git tag "v${VERSION}"
git push origin "v${VERSION}"
cd ../..

cd sdks/rust
cargo publish
cd ../..

cd sdks/ruby
gem build core.gemspec
gem push "core-${VERSION}.gem"
cd ../..

cd sdks/java
./gradlew publish
cd ../..

cd sdks/csharp
: "${NUGET_API_KEY:?NUGET_API_KEY must be set}"
dotnet pack -c Release
while IFS= read -r -d '' pkg; do
  dotnet nuget push "$pkg" --source https://api.nuget.org/v3/index.json --api-key "${NUGET_API_KEY}"
done < <(find . -name '*.nupkg' -print0)
cd ../..

cd sdks/dart
dart pub publish --force
cd ../..

echo "All SDKs published at version ${VERSION}"
