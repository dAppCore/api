#!/bin/bash

set -euo pipefail

VERSION="${1:-}"

if [ -z "${VERSION}" ]; then
  echo "Usage: ./scripts/publish-sdks.sh <version>"
  exit 1
fi

echo "PHP: publish via Packagist automation"

cd sdks/typescript
npm version --no-git-tag-version "${VERSION}"
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
nupkg_dir="./nupkgs"
rm -rf "$nupkg_dir"
mkdir -p "$nupkg_dir"
dotnet pack -c Release -o "$nupkg_dir"
pkg_count=0
while IFS= read -r -d '' pkg; do
  pkg_count=$((pkg_count + 1))
  dotnet nuget push "$pkg" --source https://api.nuget.org/v3/index.json --api-key "${NUGET_API_KEY}"
done < <(find "$nupkg_dir" -name '*.nupkg' -print0)

if [ "${pkg_count}" -eq 0 ]; then
  echo "Error: no .nupkg files found after dotnet pack" >&2
  exit 1
fi
cd ../..

cd sdks/dart
dart pub publish --force
cd ../..

echo "All SDKs published at version ${VERSION}"
