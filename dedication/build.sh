#!/bin/sh
# build.sh: assemble and link dedication with Apple's toolchain.
set -e
cd "$(dirname "$0")"
sdk=$(xcrun --show-sdk-path)
ver=$(xcrun --show-sdk-version)
as -arch arm64 -o dedication.o dedication.s
ld -arch arm64 -platform_version macos "$ver" "$ver" -syslibroot "$sdk" \
	-lSystem -e _main -x -o dedication dedication.o
rm dedication.o
codesign -v dedication 2>/dev/null || codesign -s - dedication
