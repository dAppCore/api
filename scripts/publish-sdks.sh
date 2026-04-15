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
dotnet pack -c Release
dotnet nuget push **/*.nupkg --source https://api.nuget.org/v3/index.json
cd ../..

cd sdks/dart
dart pub publish
cd ../..

echo "All SDKs published at version ${VERSION}"
