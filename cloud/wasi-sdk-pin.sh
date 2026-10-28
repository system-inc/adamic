#!/usr/bin/env bash
# cloud/wasi-sdk-pin.sh <clang> prints the WASI SDK archive setup.sh lays, "<architecture> <sha256>", chosen by the
# architecture the toolchain's clang targets, never the kernel's. The Mac guests run the house's amd64 toolchain under
# Rosetta on an arm64 kernel, so `uname -m` there names the other archive, and every wasi-keyed unit placed on a guest
# would void missingTools against a record of arm64's SDK (#9nqg4ex, Computers on #b9v3c7n).
set -euo pipefail

clang=${1:?usage: bash cloud/wasi-sdk-pin.sh <clang>}
machine=$("$clang" -dumpmachine) || { echo "wasi-sdk-pin: $clang -dumpmachine failed" >&2 && exit 1; }
case ${machine%%-*} in
	x86_64) echo "x86_64 b7d4d944c88503e4f21d84af07ac293e3440b1b6210bfd7fe78e0afd92c23bc2" ;;
	aarch64 | arm64) echo "arm64 4cf4c553c4640e63e780442146f87d83fdff5737f988c06a6e3b2f0228e37665" ;;
	*) echo "wasi-sdk-pin: no WASI SDK for $clang's target, $machine" >&2 && exit 1 ;;
esac
