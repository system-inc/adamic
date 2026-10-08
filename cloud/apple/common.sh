#!/usr/bin/env bash
# Shared preparation for the two Apple bundle builds.
set -Eeuo pipefail
PS4='+ ${BASH_SOURCE##*/}:${LINENO}: '
set -x
trap 'status=$?; echo "apple: failed at $BASH_SOURCE:$LINENO (exit $status)" >&2; exit "$status"' ERR

repository=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repository"
if [[ $(uname -s) != Darwin ]]; then
    echo 'apple: Xcode on macOS is required; no Apple build was performed' >&2
    exit 1
fi
for tool in go node xcrun codesign cmp; do
    command -v "$tool"
done
# Always use a fresh output directory. Failed or stale bundles cannot pass a later comparison.
out=$(mktemp -d "${TMPDIR:-/tmp}/adamic-${platform}-XXXXXX")
echo "apple: artifacts and output remain in $out" >&2
runtime="$repository/internal/native/runtime"
go build -o "$out/adamic" ./cmd/adamic
"$out/adamic" c dedication/dedication.a > "$out/main.c"
node oracle/node.mjs dedication/dedication.a > "$out/node.stdout"
# Match native.Flags: floating operations round independently and recursion keeps its frames.
flags=(-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable
    -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter
    -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2)
sources=()
for source in "$runtime"/*.c; do
    # The external Go checker bridge is not used by this dedication program.
    [[ ${source##*/} == tsgo.c ]] && continue
    sources+=("$source")
done
bundle_id=org.system.adamic.dedication
