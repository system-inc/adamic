#!/bin/sh
# build.sh: assemble and link dedication with Apple's toolchain.
set -e
# dedication.s is arm64 assembly for macOS, assembled and linked with Apple's toolchain, so anywhere
# else there's nothing to build. Say so and stop, rather than fail; the dedication's Adamic source is
# still held by the oracle on every platform.
if [ "$(uname -s)" != Darwin ] || [ "$(uname -m)" != arm64 ]; then
	echo "skipped: dedication.s is arm64 macOS assembly, and this is $(uname -s) $(uname -m)" >&2
	exit 0
fi
cd "$(dirname "$0")"
sdk=$(xcrun --show-sdk-path)
ver=$(xcrun --show-sdk-version)
as -arch arm64 -o dedication.o dedication.s
ld -arch arm64 -platform_version macos "$ver" "$ver" -syslibroot "$sdk" \
	-lSystem -e _main -x -o dedication dedication.o
rm dedication.o
codesign -v dedication 2>/dev/null || codesign -s - dedication
