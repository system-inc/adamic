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
for argument in "$@"; do
	case "$argument" in
		--warm-tests) warmTests=true ;;
		--wasi-sdk) wasiSDK=true ;;
		*) echo "usage: bash cloud/setup.sh [--warm-tests] [--wasi-sdk]" >&2; exit 2 ;;
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

# Every tarball is downloaded whole to a file, checked against its published sha256, extracted into a staging
# directory and renamed into place (Loom, Oct 9: a download cut short, piped straight into tar, left a Go without
# its standard library on 8 of 15 Codex instances, and every later run accepted it). A tool dir already there, whole
# or not, is moved aside, never deleted.
verifiedDownload() {
	local url=$1 expected=$2 archive actual
	archive=$(mktemp "$run/download.XXXXXX")
	curl -fsSL --retry 3 -o "$archive" "$url"
	actual=$(sha256sum "$archive" | cut -d' ' -f1)
	if [ -z "$expected" ] || [ "$actual" != "$expected" ]; then
		echo "setup: $url has sha256 $actual, not '$expected' (cut off or altered): refused" >&2
		rm -f "$archive"
		return 1
	fi
	echo "$archive"
}
installDirectory() {
	local archive=$1 destination=$2 staging
	shift 2
	staging=$(mktemp -d "$(dirname "$destination")/.staging.XXXXXX")
	tar --no-same-owner -C "$staging" "$@" -f "$archive"
	rm -f "$archive"
	[ ! -e "$destination" ] || mv "$destination" "$destination.aside.$(date +%s).$$"
	chmod 755 "$staging"
	mv "$staging" "$destination"
}
# Pinned releases' digests, from each GitHub release's asset digest.
llvmDigest() {
	case $1 in
		20.1.8-X64) echo 1ead36b3dfcb774b57be530df42bec70ab2d239fbce9889447c7a29a4ddc1ae6 ;;
		20.1.8-ARM64) echo b855cc17d935fdd83da82206b7a7cfc680095efd1e9e8182c4a05e761958bef8 ;;
		*) echo "${ADAMIC_LLVM_SHA256:-}" ;;
	esac
}
wasiDigest() {
	case $1 in
		27-x86_64) echo b7d4d944c88503e4f21d84af07ac293e3440b1b6210bfd7fe78e0afd92c23bc2 ;;
		27-arm64) echo 4cf4c553c4640e63e780442146f87d83fdff5737f988c06a6e3b2f0228e37665 ;;
	esac
}

# Go. Any Go from 1.21 on fetches the version go.mod names by itself (GOTOOLCHAIN=auto), so an
# installed one is enough; otherwise the newest stable release goes in $tools/go.
export GOTOOLCHAIN=auto
# Some cloud boxes reach proxy.golang.org but not storage.googleapis.com, where it redirects module
# downloads, and answer 403. The default proxy list falls back to direct only on 404 and 410, so a
# module the box hasn't cached fails setup. The pipe falls back on any error, fetching from the
# module's own source instead; go.sum still checks every module's hash either way.
export GOPROXY="https://proxy.golang.org|direct"
prepareGo() {
# Only a go that answers `go version` and has its standard library counts: some images ship an unrelated
# /usr/bin/go, and a cut-off install answers `go version` with no std.
realGo() {
	local root
	"$1" version 2> /dev/null | grep -q '^go version go1\.' && root=$("$1" env GOROOT 2> /dev/null) && [ -f "$root/src/runtime/runtime.go" ]
}
if ! { command -v go > /dev/null 2>&1 && realGo go; } && ! realGo "$tools/go/bin/go"; then
	goVersion=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1)
	goArchive=$(verifiedDownload "https://dl.google.com/go/$goVersion.linux-$goArchitecture.tar.gz" "$(curl -fsSL "https://dl.google.com/go/$goVersion.linux-$goArchitecture.tar.gz.sha256")")
	installDirectory "$goArchive" "$tools/go" -xz --strip-components 1
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
	llvmArchive=$(verifiedDownload "https://github.com/llvm/llvm-project/releases/download/llvmorg-$llvmVersion/LLVM-$llvmVersion-Linux-$llvmArchitecture.tar.xz" "$(llvmDigest "$llvmVersion-$llvmArchitecture")")
	installDirectory "$llvmArchive" "$tools/llvm" -xJ --strip-components 1
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
		nodeFile=node-$nodeVersion-linux-$nodeArchitecture.tar.xz
		nodeArchive=$(verifiedDownload "https://nodejs.org/dist/$nodeVersion/$nodeFile" "$(curl -fsSL "https://nodejs.org/dist/$nodeVersion/SHASUMS256.txt" | awk -v file="$nodeFile" '$2 == file {print $1}')")
		installDirectory "$nodeArchive" "$run/node" -xJ --strip-components 2 --wildcards '*/bin/node'
		mv "$run/node/node" "$tools/bin/node.partial"
		mv "$tools/bin/node.partial" "$tools/bin/node"
	fi
fi
"$tools/bin/node" --version
step "node ready"
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
 wasiDirectory="$tools/wasi-sdk"
 if [ ! -x "$wasiDirectory/bin/clang" ]; then
  case $(uname -m) in
   x86_64) wasiArchitecture=x86_64 ;;
   *) wasiArchitecture=arm64 ;;
  esac
  wasiArchive=$(verifiedDownload "https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-$wasiVersion/wasi-sdk-$wasiVersion.0-$wasiArchitecture-linux.tar.gz" "$(wasiDigest "$wasiVersion-$wasiArchitecture")")
  installDirectory "$wasiArchive" "$wasiDirectory" -xz --strip-components 1
 fi
 "$wasiDirectory/bin/clang" --version | head -n 1
 step "wasi sdk ready ($wasiDirectory)"
fi

# One file every shell sources: the agent's shell in Codex is a different session from this one.
cat > "$tools/env.sh" << ENV
export PATH="$tools/bin:$([ -x "$tools/go/bin/go" ] && echo "$tools/go/bin:")\$PATH"
export GOTOOLCHAIN=auto
export GOPROXY="https://proxy.golang.org|direct"
export TMPDIR=$gate
export ADAMIC_MARKDOWNWIDTH_DEPS="$markdownDependencies"
ENV
if "$wasiSDK"; then
 printf 'export WASI_SYSROOT=%q\n' "$wasiDirectory/share/wasi-sysroot" >> "$tools/env.sh"
fi
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
echo "setup: build-flags commit=$(git -C "$repository" rev-parse HEAD) nproc=$(nproc) cpu.max=$cpuQuota go=$(go version) clang=$(clang --version | head -n 1) node=$(node --version) cached=$([ "${ADAMIC_GATE_UNCACHED:-0}" = 1 ] && echo no || echo yes) warm-tests=$warmTests load-before=$loadBefore load-after=$(cat /proc/loadavg)"
step "done on $(nproc) processors (cgroup cpu.max: $cpuQuota), $memory"
echo "setup: source $tools/env.sh"
echo "setup: logs $run"
