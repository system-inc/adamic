#!/bin/bash
# Packages the VS Code extension as a .vsix, versioned with the cohere release it ships beside.
#
#     package-vscode-extension.sh <version> <output-directory>
#     package-vscode-extension.sh 1.0.0 dist
#
# The extension is versioned in lockstep with cohere, 1.0.0 with 1.0.0, so the version is stamped here
# from the release's own rather than kept in step by hand in editors/vscode/package.json. It packages a
# copy, so the checkout is never written to.
#
# cohere is MIT OR Apache-2.0, and the extension ships both texts from the repository root. vsce looks
# only for a LICENSE, LICENSE.md or LICENSE.txt to show on the Marketplace, so the stage gets a LICENSE.md
# naming the two. vsce does not enforce it: with no terminal on stdout it answers its own "continue
# without a license?" with yes (vsce 4.0.0, out/util.js), and packaged without one in a test. So the
# listing check below is what requires all three files.
set -euo pipefail

version=$1
output=$(mkdir -p "$2" && cd "$2" && pwd)
source=$(cd "$(dirname "$0")/../../editors/vscode" && pwd)
root=$(cd "$(dirname "$0")/../.." && pwd)
vsix="$output/cohere-$version.vsix"

stage=$(mktemp -d)
cp -R "$source/." "$stage/"
cp "$root/LICENSE-MIT" "$root/LICENSE-APACHE" "$root/NOTICE" "$root/THIRD_PARTY_NOTICES.md" "$stage/"
cat > "$stage/LICENSE.md" << 'EOF'
# License

cohere is licensed under either of

- the Apache License, Version 2.0, in [LICENSE-APACHE](LICENSE-APACHE), or
- the MIT license, in [LICENSE-MIT](LICENSE-MIT),

at your option. What it is built on, and those projects' licenses, are in [NOTICE](NOTICE) and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
EOF
cd "$stage"
npm pkg set version="$version"
npx --yes @vscode/vsce@4.0.0 package --no-dependencies --out "$vsix"

# Proved from the package itself rather than from the stage: what ships is what the .vsix holds.
packaged=$(unzip -p "$vsix" extension/package.json | node -e 'process.stdout.write(JSON.parse(require("fs").readFileSync(0, "utf8")).version)')
if [ "$packaged" != "$version" ]; then
    echo "the .vsix says version $packaged, and this release is $version"
    exit 1
fi
listing=$(unzip -l "$vsix")
for required in extension/extension.js extension/cohere-format.js extension/icon.png extension/readme.md extension/license.md \
    extension/LICENSE-MIT extension/LICENSE-APACHE extension/NOTICE extension/THIRD_PARTY_NOTICES.md; do
    if ! grep -qi " $required\$" <<< "$listing"; then
        echo "the .vsix holds no $required"
        exit 1
    fi
done
if grep -q "cohere-format.test.js" <<< "$listing"; then
    echo "the .vsix ships the tests, which .vscodeignore leaves out"
    exit 1
fi
echo "packaged $vsix, version $packaged"
