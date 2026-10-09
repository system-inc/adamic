#!/bin/sh
# check.sh: the dedication two ways, assembly and Adamic, one answer.
set -e
# dedication.s is arm64 assembly for macOS, assembled and linked with Apple's toolchain, so anywhere
# else there's nothing to build. Say so and stop, rather than fail; the dedication's Adamic source is
# still held by the oracle on every platform.
if [ "$(uname -s)" != Darwin ] || [ "$(uname -m)" != arm64 ]; then
	echo "skipped: dedication.s is arm64 macOS assembly, and this is $(uname -s) $(uname -m)" >&2
	exit 0
fi
cd "$(dirname "$0")"
./build.sh
./dedication > native.out
node dedication.a > adamic.out
cmp native.out adamic.out
rm native.out adamic.out
echo "ok: $(./dedication | wc -c | tr -d ' ') bytes, both ways"
