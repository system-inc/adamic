#!/usr/bin/env bash
# cloud/setup.sh makes a fresh Linux machine ready to run the gate: Go, a clang whose sanitizers
# work, Node 24 where an unprivileged user can run it, cohere checked out, and Go's build cache warm.
# Every cloud environment runs it (Codex's setup and maintenance scripts, Claude's, a plain VM), it's
# safe to run again, and it prints how long each step took, so time-to-green is a number we watch.
#
#   bash cloud/setup.sh [--warm-tests]   then: source the env.sh path printed below
#
# Go checks its own content-addressed action cache before a warming stamp can skip linking.
# Test binaries are optional because most workers need one package.
#
# Nothing here is needed on a Mac: Xcode's clang and leaks already do this job there.
set -euo pipefail

started=$EPOCHREALTIME
loadBefore=$(cat /proc/loadavg)
warmTests=false
wasiSDK=false
wasmtime=false
workerd=false
for argument in "$@"; do
	case "$argument" in
		--warm-tests) warmTests=true ;;
		--wasi-sdk) wasiSDK=true ;;
		--wasmtime) wasmtime=true ;;
		--workerd) workerd=true ;;
		*) echo "usage: bash cloud/setup.sh [--warm-tests] [--wasi-sdk] [--wasmtime] [--workerd]" >&2; exit 2 ;;
	esac
done
step() {
	local elapsed
	elapsed=$(awk -v start="$started" -v now="$EPOCHREALTIME" 'BEGIN {printf "%.3f", now - start}')
	echo "setup: $1 (${elapsed}s)"
}

repository=$(cd "$(dirname "$0")/.." && pwd)
tools=${ADAMIC_TOOLS:-/opt/adamic-tools}
mkdir -p "$tools/bin" 2> /dev/null || { tools=$HOME/.adamic-tools && mkdir -p "$tools/bin"; }
gate=/tmp/adamic-gate
# The input tests drop to uid 65534 when run as root, so everything they touch must be traversable
# by that user: the temporary directory, and the Node they run.
install -d -m 1777 "$gate"
# Serialize installers sharing this tool directory. Children inherit the lock descriptor.
exec 9> "$tools/setup.lock"
flock 9
run=$(mktemp -d "$gate/setup.XXXXXX")
markdownDependencies="$tools/markdown-width"
case $(realpath -m "$markdownDependencies") in
	/root | /root/*) markdownDependencies="$gate/markdown-width-$(printf '%s' "$tools" | sha256sum | cut -d' ' -f1)" ;;
esac
case $(uname -m) in
	x86_64) goArchitecture=amd64 nodeArchitecture=x64 llvmArchitecture=X64 ;;
	aarch64 | arm64) goArchitecture=arm64 nodeArchitecture=arm64 llvmArchitecture=ARM64 ;;
	*) echo "setup: no tools for $(uname -m)" >&2 && exit 1 ;;
esac

# Go. Any Go from 1.21 on fetches the version go.mod names by itself (GOTOOLCHAIN=auto), so an
# installed one is enough; otherwise the newest stable release goes in $tools/go.
export GOTOOLCHAIN=auto
# Some cloud boxes reach proxy.golang.org but not storage.googleapis.com, where it redirects module
# downloads, and answer 403. The default proxy list falls back to direct only on 404 and 410, so a
# module the box hasn't cached fails setup. The pipe falls back on any error, fetching from the
# module's own source instead; go.sum still checks every module's hash either way.
export GOPROXY="https://proxy.golang.org|direct"
prepareGo() {
# Only a go that answers `go version` counts: some images ship an unrelated /usr/bin/go.
realGo() { "$1" version 2> /dev/null | grep -q '^go version go1\.'; }
if ! { command -v go > /dev/null 2>&1 && realGo go; } && ! realGo "$tools/go/bin/go"; then
	goVersion=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1)
	curl -fsSL "https://dl.google.com/go/$goVersion.linux-$goArchitecture.tar.gz" | tar --no-same-owner -xz -C "$tools"
fi
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"
(cd "$repository" && go version)
step "go ready"
}

# clang. A candidate counts only if a program that reads past an array, built with AddressSanitizer
# and UndefinedBehaviorSanitizer, is stopped by the sanitizer: a clang without its sanitizer runtime
# compiles and links that program, then runs it silently.
prepareClang() {
probe=$run
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
}

# Node 24, outside /root.
prepareNode() {
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
# Stock tsc, TypeScript 6.0.3 as cohere's checks pin it: library's test262 runner starts it as an oracle
# (cmd/adamic-test262/typescript.go, exec.LookPath("tsc")), so a missing or other tsc fails setup, never a skipped test.
typescriptVersion=6.0.3
if ! PATH="$tools/bin:$PATH" tsc --version 2> /dev/null | grep -qx "Version $typescriptVersion"; then
	# Downloaded whole and checked against its pinned digest before anything unpacks, then staged and renamed in, so a
	# cut download or a half-unpacked tree never becomes $tools/typescript.
	curl -fsSL "https://registry.npmjs.org/typescript/-/typescript-$typescriptVersion.tgz" -o "$run/typescript.tgz"
	echo "33cd0ee1beaa8c9e9d15a9da836c62ddea4c34a42d7c2d349dbc80d94165d22a  $run/typescript.tgz" | sha256sum -c --quiet - ||
		{ echo "setup: typescript-$typescriptVersion.tgz doesn't match its pinned sha256" >&2; return 1; }
	mkdir "$run/typescript"
	tar --no-same-owner -xz -C "$run/typescript" --strip-components 1 -f "$run/typescript.tgz"
	rm -f "$run/typescript.tgz"
	[ ! -e "$tools/typescript" ] || mv "$tools/typescript" "$run/typescript.old"
	mv "$run/typescript" "$tools/typescript"
fi
# A link to the package's own bin/tsc, which runs on the node in $tools/bin: cloud/fast-gate/tools.txt proves tsc by
# lib/typescript.js beside the file the link resolves to.
ln -sfn ../typescript/bin/tsc "$tools/bin/tsc.partial"
mv -f "$tools/bin/tsc.partial" "$tools/bin/tsc"
PATH="$tools/bin:$PATH" tsc --version | grep -qx "Version $typescriptVersion" && test -f "$(dirname "$(readlink -f "$tools/bin/tsc")")/../lib/typescript.js" ||
	{ echo "setup: $tools/bin/tsc isn't TypeScript $typescriptVersion" >&2; return 1; }
step "tsc ready (TypeScript $typescriptVersion)"
python3 "$repository/cloud/setup-markdown-width.py" "$repository/cloud/markdown-width" "$markdownDependencies" "$tools/bin/node" > "$run/markdown.log" 2>&1 || { cat "$run/markdown.log"; return 1; }
cat "$run/markdown.log"
step "markdown dependencies ready"
}

prepareSubmodules() {
# cohere, and typescript-go inside it, over HTTPS (the recorded URL is SSH, which clouds can't use).
cd "$repository"
git config submodule.cohere.url https://github.com/system-inc/cohere.git
# A fresh worker box restores cohere from the gate box's cache in R2 (keyed by the gitlink, every part
# hash-checked) instead of cloning 81,500 files; with no cache, or any doubt, the clone below does it.
bash "$repository/cloud/submodule-cache.sh" restore "$repository" || true
git submodule update --init --recursive --depth 1 --filter=blob:none
step "submodules ready"

}

# Downloads and the sanitizer check do not need each other's results. Build only after all
# have succeeded; wait for every child even when one fails, so no installer outlives setup.
prepareGo & goProcess=$!
prepareClang & clangProcess=$!
prepareNode & nodeProcess=$!
prepareSubmodules & submoduleProcess=$!
failed=0
for process in "$goProcess" "$clangProcess" "$nodeProcess" "$submoduleProcess"; do
	wait "$process" || failed=1
done
[ "$failed" = 0 ] || exit 1
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"

# Optional WASI SDK 27: native clang remains the default in PATH.
if "$wasiSDK"; then
 wasiVersion=27
 case $(uname -m) in
  x86_64)
   wasiArchitecture=x86_64
   wasiSHA256=b7d4d944c88503e4f21d84af07ac293e3440b1b6210bfd7fe78e0afd92c23bc2 ;;
  *)
   wasiArchitecture=arm64
   wasiSHA256=4cf4c553c4640e63e780442146f87d83fdff5737f988c06a6e3b2f0228e37665 ;;
 esac
 wasiDirectory="$tools/wasi-sdk"
 # Loom keys products and units built with the SDK by its archive's sha256, read from archive-sha256 beside it, as
 # wasmtime's is: an install counts only when its record is the pin, and anything else (an older install with no
 # record, another release, an interrupted lay) is laid again. The archive is checked whole before anything unpacks,
 # and unpacked beside the SDK and renamed in whole, since units on a live box compile against it and an extract in
 # place could hand one a half-written file.
 if [ ! -x "$wasiDirectory/bin/clang" ] || [ "$(cat "$wasiDirectory/archive-sha256" 2> /dev/null)" != "$wasiSHA256" ]; then
  wasiArchive="$run/wasi-sdk-$wasiVersion.0-$wasiArchitecture-linux.tar.gz"
  curl -fsSL "https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-$wasiVersion/wasi-sdk-$wasiVersion.0-$wasiArchitecture-linux.tar.gz" -o "$wasiArchive"
  printf '%s  %s\n' "$wasiSHA256" "$wasiArchive" | sha256sum -c - ||
   { echo "setup: the wasi-sdk $wasiVersion archive isn't its pin, $wasiSHA256; the installed SDK is left as it was" >&2 && exit 1; }
  wasiStaged="$tools/.wasi-sdk.partial"
  rm -rf "$wasiStaged"
  mkdir -p "$wasiStaged"
  tar --no-same-owner -xzf "$wasiArchive" -C "$wasiStaged" --strip-components 1
  rm -f "$wasiArchive"
  printf '%s\n' "$wasiSHA256" > "$wasiStaged/archive-sha256"
  rm -rf "$tools/.wasi-sdk.replaced"
  if [ -e "$wasiDirectory" ]; then
   mv "$wasiDirectory" "$tools/.wasi-sdk.replaced"
  fi
  mv "$wasiStaged" "$wasiDirectory"
  rm -rf "$tools/.wasi-sdk.replaced"
 fi
 "$wasiDirectory/bin/clang" --version | head -n 1
 # The WASI tests link with the clang on PATH ($tools/bin/clang), whose resource directory holds builtins for the
 # native target only, so wasm-ld can't open wasm32-unknown-wasi/libclang_rt.builtins.a and every WASI test skips
 # (Oct 10: the 36 internal/native wasm units on each box-strict box). The SDK carries the wasm32 builtins for its
 # own clang; linked into the PATH clang's resource directory when both are the same clang major, they serve it.
 resource=$("$tools/bin/clang" -print-resource-dir)
 sdkBuiltins="$wasiDirectory/lib/clang/$(basename "$resource")/lib"
 [ -d "$sdkBuiltins" ] || { echo "setup: wasi-sdk $wasiVersion has no builtins for clang $(basename "$resource") ($(ls "$wasiDirectory/lib/clang"))" >&2 && exit 1; }
 for builtins in "$sdkBuiltins"/wasm32-*; do
  [ -e "$resource/lib/${builtins##*/}" ] && [ ! -L "$resource/lib/${builtins##*/}" ] || ln -sfn "$builtins" "$resource/lib/${builtins##*/}"
 done
 # A claimed WASI toolchain links: the WASI tests' own probe, with the clang they run.
 printf '#include <stdlib.h>\nint main(void) { void *p = malloc(16); free(p); return 0; }\n' > "$run/wasi.c"
 "$tools/bin/clang" --target=wasm32-wasi --sysroot="$wasiDirectory/share/wasi-sysroot" -O2 "$run/wasi.c" -o "$run/wasi.wasm" ||
  { echo "setup: $tools/bin/clang can't link for wasm32-wasi with wasi-sdk $wasiVersion" >&2 && exit 1; }
 step "wasi sdk ready ($wasiDirectory, archive sha256 $wasiSHA256, linking with $tools/bin/clang)"
fi

# Optional independent WASI engine, pinned with archive checksums.
if "$wasmtime"; then
 wasmtimeVersion=38.0.3
 case $(uname -m) in
  x86_64)
   wasmtimeArchitecture=x86_64
   wasmtimeSHA256=101d79dff495b0392d583d11c3c78dd50941c3ae28e80cf1604ca43acdf05af7 ;;
  *)
   wasmtimeArchitecture=aarch64
   wasmtimeSHA256=10f8dd0f4789075321a439a1fb4a3d1888e3a45c0620dc9c562e198095120120 ;;
 esac
 wasmtimeDirectory="$tools/wasmtime-$wasmtimeVersion"
 # Loom's runner reads archive-sha256 in the directory of the wasmtime a unit's PATH resolves to, and voids a unit
 # whose key names another digest, so a pass is never recorded under bytes that didn't run. An install counts only
 # when its record is the pin; anything else (an older install with no record, an interrupted lay) is laid again.
 if [ ! -x "$wasmtimeDirectory/wasmtime" ] || [ "$(cat "$wasmtimeDirectory/archive-sha256" 2> /dev/null)" != "$wasmtimeSHA256" ]; then
  wasmtimeArchive="$tools/wasmtime-$wasmtimeVersion-$wasmtimeArchitecture.tar.xz"
  curl -fsSL "https://github.com/bytecodealliance/wasmtime/releases/download/v$wasmtimeVersion/wasmtime-v$wasmtimeVersion-$wasmtimeArchitecture-linux.tar.xz" -o "$wasmtimeArchive"
  printf '%s  %s\n' "$wasmtimeSHA256" "$wasmtimeArchive" | sha256sum -c -
  mkdir -p "$wasmtimeDirectory"
  rm -f "$wasmtimeDirectory/archive-sha256"
  tar --no-same-owner -xJf "$wasmtimeArchive" -C "$wasmtimeDirectory" --strip-components 1
  printf '%s\n' "$wasmtimeSHA256" > "$wasmtimeDirectory/archive-sha256"
 fi
 "$wasmtimeDirectory/wasmtime" --version
 ln -sf "$wasmtimeDirectory/wasmtime" "$tools/bin/wasmtime"
 step "wasmtime ready ($wasmtimeDirectory, archive sha256 $wasmtimeSHA256)"
fi

# Optional Workers engine, pinned with archive checksums.
if "$workerd"; then
 workerdVersion=1.20261007.1
 case $(uname -m) in
  x86_64)
   workerdArchitecture=64
   workerdSHA256=7b5fc5ba52c0db4b39d70cef061b2eea6b3d05ce590765e11d953516dec4a83f ;;
  *)
   workerdArchitecture=arm64
   workerdSHA256=7306969ef569b45d171ea7ddb24210f12fad1a34fda3d6b740ee8472ae092e03 ;;
 esac
 workerdDirectory="$tools/workerd-$workerdVersion"
 # Loom's runner reads archive-sha256 in the directory of the workerd a unit's PATH resolves to, and voids a unit
 # whose key names another digest, so a pass is never recorded under bytes that didn't run. An install counts only
 # when its record is the pin; anything else (an older install with no record, an interrupted lay) is laid again.
 if [ ! -x "$workerdDirectory/workerd" ] || [ "$(cat "$workerdDirectory/archive-sha256" 2> /dev/null)" != "$workerdSHA256" ]; then
  workerdArchive="$tools/workerd-$workerdVersion-$workerdArchitecture.tgz"
  curl -fsSL "https://registry.npmjs.org/@cloudflare/workerd-linux-$workerdArchitecture/-/workerd-linux-$workerdArchitecture-$workerdVersion.tgz" -o "$workerdArchive"
  printf '%s  %s\n' "$workerdSHA256" "$workerdArchive" | sha256sum -c -
  mkdir -p "$workerdDirectory"
  rm -f "$workerdDirectory/archive-sha256"
  tar --no-same-owner -xzf "$workerdArchive" -C "$workerdDirectory" --strip-components 2 package/bin/workerd
  printf '%s\n' "$workerdSHA256" > "$workerdDirectory/archive-sha256"
 fi
 "$workerdDirectory/workerd" --version
 ln -sf "$workerdDirectory/workerd" "$tools/bin/workerd"
 step "workerd ready ($workerdDirectory, archive sha256 $workerdSHA256)"
fi

# Bootstrap without the hook: the cache program cannot cache its own build. Older branches and
# bootstrap failures still warm Go's local cache, so a shared-cache outage never blocks setup.
# The shared cache is opt-in (ADAMIC_GOCACHE=1): nothing has written refs/gocache since the gate's publish path
# went (08bc0930), so every go command in a shell with it on asked the store for a ref no one can hit, about
# 94.5K 404s a day on Oct 14 (#6dg2cwd).
unset GOCACHEPROG
gocacheprog=off
if [ "${ADAMIC_GOCACHE:-0}" != 1 ]; then
	step "shared cache off (opt in with ADAMIC_GOCACHE=1)"
elif [ ! -d "$repository/cmd/adamic-gocacheprog" ]; then
	step "shared cache off (cmd/adamic-gocacheprog absent)"
elif (cd "$repository" && env -u GOCACHEPROG go build -o "$tools/bin/adamic-gocacheprog" ./cmd/adamic-gocacheprog) > "$run/gocacheprog.log" 2>&1; then
	export GOCACHEPROG="$tools/bin/adamic-gocacheprog"
	gocacheprog=on
	step "shared cache ready"
else
	step "shared cache off (adamic-gocacheprog build failed; see $run/gocacheprog.log)"
fi

# One file every shell sources; write-env.sh keeps the lines other steps wrote, such as a box's gate inputs.
bash "$repository/cloud/write-env.sh" "$tools" "$gate" "$markdownDependencies" "$repository" "$gocacheprog" "$("$wasiSDK" && echo "$wasiDirectory" || true)"
grep -qs "$tools/env.sh" ~/.bashrc || echo "source $tools/env.sh" >> ~/.bashrc
# shellcheck disable=SC1091
source "$tools/env.sh"
cd "$repository"

# Validate actual dependency actions, including dirty sources, embeds, local replacements and
# compiler flags, before consulting the stamp. HEAD alone cannot safely replace Go's checks.
buildArguments=()
key=""
stamp="$tools/warm-$(printf '%s' "$repository" | sha256sum | cut -d' ' -f1)-$warmTests"
if [ "${ADAMIC_GATE_UNCACHED:-0}" = 1 ]; then
	buildArguments=(-a)
else
	listArguments=()
	"$warmTests" && listArguments=(-test)
	go list -deps -export "${listArguments[@]}" -json ./... > "$run/packages.json" 2> "$run/list.log" || { cat "$run/list.log"; exit 1; }
	key=$(python3 "$repository/cloud/setup-key.py" "$repository" "$run/packages.json" "$warmTests") || key=""
fi
if [ -n "$key" ] && [ "$(cat "$stamp" 2> /dev/null || true)" = "$key" ]; then
	step "go build skipped (validated warming stamp)"
	"$warmTests" && step "test binaries skipped (validated warming stamp)"
else
	go build "${buildArguments[@]}" ./... > "$run/build.log" 2>&1 || { cat "$run/build.log"; exit 1; }
	step "go build ready"
	if "$warmTests"; then
		go test "${buildArguments[@]}" -count=1 -run '^$' ./... > "$run/tests.log" 2>&1 || { cat "$run/tests.log"; exit 1; }
		step "test binaries warm"
	fi
	if [ -n "$key" ]; then
		# Build/link adds cache entries. Publish only the key of the completed cache, atomically.
		if python3 "$repository/cloud/setup-key.py" "$repository" "$run/packages.json" "$warmTests" > "$run/stamp"; then
			mv "$run/stamp" "$stamp"
		fi
	fi
fi
"$warmTests" || step "test binaries deferred (use --warm-tests)"
step "build cache warm"

cpuQuota=$(cat /sys/fs/cgroup/cpu.max 2> /dev/null || echo unknown)
memory=$(awk '/MemTotal/ {printf "%.1f GB", $2 / 1048576}' /proc/meminfo)
echo "setup: build-flags commit=$(git -C "$repository" rev-parse HEAD) nproc=$(nproc) cpu.max=$cpuQuota go=$(go version) clang=$(clang --version | head -n 1) node=$(node --version) cached=$([ "${ADAMIC_GATE_UNCACHED:-0}" = 1 ] && echo no || echo yes) warm-tests=$warmTests gocacheprog=$gocacheprog load-before=$loadBefore load-after=$(cat /proc/loadavg)"
step "done on $(nproc) processors (cgroup cpu.max: $cpuQuota), $memory"
echo "setup: source $tools/env.sh"
echo "setup: logs $run"
