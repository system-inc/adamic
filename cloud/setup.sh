#!/usr/bin/env bash
# cloud/setup.sh makes a fresh Linux machine ready to run the gate: Go, a clang whose sanitizers
# work, Node 24 where an unprivileged user can run it, cohere checked out, and Go's build cache warm.
# Every cloud environment runs it (Codex's setup and maintenance scripts, Claude's, a plain VM), it's
# safe to run again, and it prints how long each step took, so time-to-green is a number we watch.
#
#   bash cloud/setup.sh            then: source /opt/adamic-tools/env.sh
#
# Nothing here is needed on a Mac: Xcode's clang and leaks already do this job there.
set -euo pipefail

wasiSDK=false
case ${1:-} in
 "") ;;
 --wasi-sdk) wasiSDK=true ;;
 *) echo "usage: bash cloud/setup.sh [--wasi-sdk]" >&2; exit 2 ;;
esac

started=$(date +%s)
step() { echo "setup: $1 ($(($(date +%s) - started))s)"; }

repository=$(cd "$(dirname "$0")/.." && pwd)
tools=${ADAMIC_TOOLS:-/opt/adamic-tools}
mkdir -p "$tools/bin" 2> /dev/null || { tools=$HOME/.adamic-tools && mkdir -p "$tools/bin"; }
gate=/tmp/adamic-gate
# The input tests drop to uid 65534 when run as root, so everything they touch must be traversable
# by that user: the temporary directory, and the Node they run.
install -d -m 1777 "$gate"
case $(uname -m) in
	x86_64) goArchitecture=amd64 nodeArchitecture=x64 llvmArchitecture=X64 ;;
	aarch64 | arm64) goArchitecture=arm64 nodeArchitecture=arm64 llvmArchitecture=ARM64 ;;
	*) echo "setup: no tools for $(uname -m)" >&2 && exit 1 ;;
esac

# Go. Any Go from 1.21 on fetches the version go.mod names by itself (GOTOOLCHAIN=auto), so an
# installed one is enough; otherwise the newest stable release goes in $tools/go.
export GOTOOLCHAIN=auto
# Only a go that answers `go version` counts: some images ship an unrelated /usr/bin/go.
realGo() { "$1" version 2> /dev/null | grep -q '^go version go1\.'; }
if ! { command -v go > /dev/null 2>&1 && realGo go; } && [ ! -x "$tools/go/bin/go" ]; then
	goVersion=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1)
	curl -fsSL "https://dl.google.com/go/$goVersion.linux-$goArchitecture.tar.gz" | tar --no-same-owner -xz -C "$tools"
fi
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"
(cd "$repository" && go version)
step "go ready"

# clang. A candidate counts only if a program that reads past an array, built with AddressSanitizer
# and UndefinedBehaviorSanitizer, is stopped by the sanitizer: a clang without its sanitizer runtime
# compiles and links that program, then runs it silently.
probe=$(mktemp -d)
cat > "$probe/overflow.c" << 'C'
#include <stdlib.h>
int main(void) { int *values = malloc(4 * sizeof(int)); int read = values[4]; free(values); return read; }
C
saneClang() {
	"$1" -fsanitize=address,undefined -g "$probe/overflow.c" -o "$probe/overflow" > /dev/null 2>&1 || return 1
	! "$probe/overflow" > /dev/null 2> "$probe/report" && grep -q 'heap-buffer-overflow' "$probe/report"
}
clang=""
for candidate in "$tools/llvm/bin/clang" $(ls -d /usr/lib/llvm-*/bin/clang 2> /dev/null | sort -t- -k2 -V -r) $(command -v clang || true); do
	if [ -x "$candidate" ] && saneClang "$candidate"; then
		clang=$candidate
		break
	fi
done
if [ -z "$clang" ]; then
	llvmVersion=${ADAMIC_LLVM_VERSION:-20.1.8}
	mkdir -p "$tools/llvm"
	curl -fsSL "https://github.com/llvm/llvm-project/releases/download/llvmorg-$llvmVersion/LLVM-$llvmVersion-Linux-$llvmArchitecture.tar.xz" | tar --no-same-owner -xJ -C "$tools/llvm" --strip-components 1
	saneClang "$tools/llvm/bin/clang" || { echo "setup: LLVM $llvmVersion's sanitizers don't work here" >&2 && exit 1; }
	clang=$tools/llvm/bin/clang
fi
# llvm-symbolizer turns a sanitizer's addresses into file and line.
for tool in clang clang++ llvm-symbolizer; do
	[ -x "$(dirname "$clang")/$tool" ] && ln -sf "$(dirname "$clang")/$tool" "$tools/bin/$tool"
done
"$clang" --version | head -n 1
step "clang ready ($clang)"

# Node 24, outside /root.
if ! "$tools/bin/node" --version 2> /dev/null | grep -q '^v24\.'; then
	existing=""
	for candidate in $(command -v node || true) /root/.nvm/versions/node/v24*/bin/node; do
		if [ -x "$candidate" ] && "$candidate" --version | grep -q '^v24\.'; then
			existing=$candidate
			break
		fi
	done
	if [ -n "$existing" ]; then
		install -m 755 "$existing" "$tools/bin/node"
	else
		nodeVersion=$(curl -fsSL https://nodejs.org/dist/index.json | python3 -c 'import json, sys; print(next(r["version"] for r in json.load(sys.stdin) if r["version"].startswith("v24.")))')
		curl -fsSL "https://nodejs.org/dist/$nodeVersion/node-$nodeVersion-linux-$nodeArchitecture.tar.xz" | tar --no-same-owner -xJ -C "$tools" --strip-components 2 --wildcards '*/bin/node'
		mv "$tools/node" "$tools/bin/node"
	fi
fi
"$tools/bin/node" --version
step "node ready"

# Optional WASI SDK 27: native clang remains the default in PATH.
if "$wasiSDK"; then
 wasiVersion=27
 wasiDirectory="$tools/wasi-sdk"
 if [ ! -x "$wasiDirectory/bin/clang" ]; then
  case $(uname -m) in
   x86_64) wasiArchitecture=x86_64 ;;
   *) wasiArchitecture=arm64 ;;
  esac
  mkdir -p "$wasiDirectory"
  curl -fsSL "https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-$wasiVersion/wasi-sdk-$wasiVersion.0-$wasiArchitecture-linux.tar.gz" | tar --no-same-owner -xz -C "$wasiDirectory" --strip-components 1
 fi
 "$wasiDirectory/bin/clang" --version | head -n 1
 step "wasi sdk ready ($wasiDirectory)"
fi

# One file every shell sources: the agent's shell in Codex is a different session from this one.
cat > "$tools/env.sh" << ENV
export PATH="$tools/bin:$([ -x "$tools/go/bin/go" ] && echo "$tools/go/bin:")\$PATH"
export GOTOOLCHAIN=auto
export TMPDIR=$gate
ENV
if "$wasiSDK"; then
 printf 'export WASI_SYSROOT=%q\n' "$wasiDirectory/share/wasi-sysroot" >> "$tools/env.sh"
fi
grep -qs "$tools/env.sh" ~/.bashrc || echo "source $tools/env.sh" >> ~/.bashrc
# shellcheck disable=SC1091
source "$tools/env.sh"

# cohere, and typescript-go inside it, over HTTPS (the recorded URL is SSH, which clouds can't use).
cd "$repository"
git config submodule.cohere.url https://github.com/system-inc/cohere.git
git submodule update --init --recursive --depth 1
step "submodules ready"

# A warm build cache: every package and every test binary compiled once, nothing run.
go build ./...
go test -count=1 -run '^$' ./... > "$gate/setup-warm.log" 2>&1 || { cat "$gate/setup-warm.log" && exit 1; }
step "build cache warm"

cpuQuota=$(cat /sys/fs/cgroup/cpu.max 2> /dev/null || echo unknown)
memory=$(awk '/MemTotal/ {printf "%.1f GB", $2 / 1048576}' /proc/meminfo)
echo "setup: done in $(($(date +%s) - started))s on $(nproc) processors (cgroup cpu.max: $cpuQuota), $memory"
