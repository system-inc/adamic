#!/usr/bin/env bash
# cloud/setup.sh makes a fresh Linux machine ready to run the gate: Go, a clang whose sanitizers
# work, Node 24 where an unprivileged user can run it, cohere checked out, and Go's build cache warm.
# Every cloud environment runs it (Codex's setup and maintenance scripts, Claude's, a plain VM), it's
# safe to run again, and it prints how long each step took, so time-to-green is a number we watch.
#
#   bash cloud/setup.sh [--warm-tests] [--gate-inputs]   then: source the env.sh path printed below
#
# Go checks its own content-addressed action cache before a warming stamp can skip linking.
# Test binaries are optional because most workers need one package.
#
# Nothing here is needed on a Mac: Xcode's clang and leaks already do this job there.
set -euo pipefail

scriptDirectory=${BASH_SOURCE[0]%/*}
[ "$scriptDirectory" != "${BASH_SOURCE[0]}" ] || scriptDirectory=.
repository=$(cd -- "$scriptDirectory/.." && pwd)
cloudSource="$repository/cloud"
# shellcheck source=../internal/boundedrun/shell.sh
source "$repository/internal/boundedrun/shell.sh"
# A clean main worktree can use this installer without copying source into it.
repository=${ADAMIC_SETUP_REPOSITORY:-$repository}
# Even local utility children get a bound; a network filesystem can stall them.
for boundedTool in cat awk mkdir install mktemp realpath uname ls sort dirname ln mv grep nproc sha256sum cut head; do
 eval "$boundedTool() { bounded 30 $boundedTool \"\$@\"; }"
done

started=$EPOCHREALTIME
loadBefore=$(cat /proc/loadavg)
warmTests=false
gateInputs=false
for argument in "$@"; do
	case "$argument" in
		--warm-tests) warmTests=true ;;
		--gate-inputs) gateInputs=true ;;
		*) echo "usage: bash cloud/setup.sh [--warm-tests] [--gate-inputs]" >&2; exit 2 ;;
	esac
done
step() {
	local elapsed
	elapsed=$(awk -v start="$started" -v now="$EPOCHREALTIME" 'BEGIN {printf "%.3f", now - start}')
	echo "setup: $1 (${elapsed}s)"
}

tools=${ADAMIC_TOOLS:-/opt/adamic-tools}
mkdir -p "$tools/bin" 2> /dev/null || { tools=$HOME/.adamic-tools && mkdir -p "$tools/bin"; }
gate=/tmp/adamic-gate
# The input tests drop to uid 65534 when run as root, so everything they touch must be traversable
# by that user: the temporary directory, and the Node they run.
install -d -m 1777 "$gate"
# Serialize installers sharing this tool directory. Children inherit the lock descriptor.
exec 9> "$tools/setup.lock"
bounded 600 flock 9
run=$(mktemp -d "$gate/setup.XXXXXX")
markdownDependencies="$tools/markdown-width"
case $(realpath -m "$markdownDependencies") in
	/root | /root/*) markdownDependencies="$gate/markdown-width-$(printf '%s' "$tools" | sha256sum | cut -d' ' -f1)" ;;
esac
gateInputsRoot="$tools/gate-inputs"
case $(realpath -m "$gateInputsRoot") in
	/root | /root/*) gateInputsRoot="$gate/gate-inputs-$(printf '%s' "$tools" | sha256sum | cut -d' ' -f1)" ;;
esac
case $(uname -m) in
	x86_64) goArchitecture=amd64 nodeArchitecture=x64 llvmArchitecture=X64 ;;
	aarch64 | arm64) goArchitecture=arm64 nodeArchitecture=arm64 llvmArchitecture=ARM64 ;;
	*) echo "setup: no tools for $(uname -m)" >&2 && exit 1 ;;
esac

# Go. Any Go from 1.21 on fetches the version go.mod names by itself (GOTOOLCHAIN=auto), so an
# installed one is enough; otherwise the newest stable release goes in $tools/go.
export GOTOOLCHAIN=auto
prepareGo() {
# Only a go that answers `go version` counts: some images ship an unrelated /usr/bin/go.
realGo() { bounded 30 "$1" version 2> /dev/null | grep -q '^go version go1\.'; }
if ! { command -v go > /dev/null 2>&1 && realGo go; } && ! realGo "$tools/go/bin/go"; then
	goVersion=$(bounded 600 curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1)
	bounded 600 curl -fsSL "https://dl.google.com/go/$goVersion.linux-$goArchitecture.tar.gz" | bounded 600 tar --no-same-owner -xz -C "$tools"
fi
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"
(cd "$repository" && bounded 30 go version)
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
	bounded 120 "$1" -fsanitize=address,undefined -g "$probe/overflow.c" -o "$probe/overflow" > /dev/null 2>&1 || return 1
	! bounded 15 "$probe/overflow" > /dev/null 2> "$probe/report" && grep -q 'heap-buffer-overflow' "$probe/report"
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
	bounded 600 curl -fsSL "https://github.com/llvm/llvm-project/releases/download/llvmorg-$llvmVersion/LLVM-$llvmVersion-Linux-$llvmArchitecture.tar.xz" | bounded 600 tar --no-same-owner -xJ -C "$tools/llvm" --strip-components 1
	saneClang "$tools/llvm/bin/clang" || { echo "setup: LLVM $llvmVersion's sanitizers don't work here" >&2 && exit 1; }
	clang=$tools/llvm/bin/clang
fi
# llvm-symbolizer turns a sanitizer's addresses into file and line.
for tool in clang clang++ llvm-symbolizer; do
	[ -x "$(dirname "$clang")/$tool" ] && ln -sf "$(dirname "$clang")/$tool" "$tools/bin/$tool"
done
bounded 30 "$clang" --version | head -n 1
step "clang ready ($clang)"
}

# Node 24, outside /root.
prepareNode() {
if ! bounded 30 "$tools/bin/node" --version 2> /dev/null | grep -q '^v24\.'; then
	existing=""
	for candidate in $(command -v node || true) /root/.nvm/versions/node/v24*/bin/node; do
		if [ -x "$candidate" ] && bounded 30 "$candidate" --version | grep -q '^v24\.'; then
			existing=$candidate
			break
		fi
	done
	if [ -n "$existing" ]; then
		install -m 755 "$existing" "$tools/bin/node"
	else
		nodeVersion=$(bounded 600 curl -fsSL https://nodejs.org/dist/index.json | bounded 30 python3 -c 'import json, sys; print(next(r["version"] for r in json.load(sys.stdin) if r["version"].startswith("v24.")))')
		bounded 600 curl -fsSL "https://nodejs.org/dist/$nodeVersion/node-$nodeVersion-linux-$nodeArchitecture.tar.xz" | bounded 600 tar --no-same-owner -xJ -C "$tools" --strip-components 2 --wildcards '*/bin/node'
		mv "$tools/node" "$tools/bin/node"
	fi
fi
bounded 30 "$tools/bin/node" --version
step "node ready"
bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-markdown-width.py" "$cloudSource/markdown-width" "$markdownDependencies" "$tools/bin/node" > "$run/markdown.log" 2>&1 || { cat "$run/markdown.log"; return 1; }
cat "$run/markdown.log"
step "markdown dependencies ready"
bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-stage3-api.py" "$repository" "$tools" "$tools/bin/node" > "$run/stage3.log" 2>&1 || { cat "$run/stage3.log"; return 1; }
cat "$run/stage3.log"
step "stage3 API dependencies checked"
if "$gateInputs"; then
	bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-gate-inputs.py" npm "$repository" "$gateInputsRoot" "$tools/bin/node" > "$run/gate-npm.log" 2>&1 || { cat "$run/gate-npm.log"; return 1; }
	cat "$run/gate-npm.log"
	step "gate npm inputs ready"
fi
}

prepareGoAndSignal() {
	trap ': > "$run/go-failed"' EXIT
	prepareGo
	: > "$run/go-ready"
	trap - EXIT

}

prepareSubmodules() {
# cohere, and typescript-go inside it, over HTTPS (the recorded URL is SSH, which clouds can't use).
cd "$repository"
bounded 30 git config submodule.cohere.url https://github.com/system-inc/cohere.git
bounded 600 git submodule update --init --recursive --depth 1 --filter=blob:none
step "submodules ready"
# Go and submodules must exist before warming the nested workspace graphs.
# The ready file keeps this in parallel preparation without racing either installer.
while [ ! -f "$run/go-ready" ]; do
	[ ! -f "$run/go-failed" ] || return 1
	sleep 0.05
done
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"
bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-modules.py" "$repository" "$tools" > "$run/modules.log" 2>&1 || { cat "$run/modules.log"; return 1; }
cat "$run/modules.log"
step "module dependencies ready"
}

prepareGateCorpora() {
	bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-gate-inputs.py" corpora "$repository" "$gateInputsRoot" "$tools/bin/node" > "$run/gate-corpora.log" 2>&1 || { cat "$run/gate-corpora.log"; return 1; }
	cat "$run/gate-corpora.log"
	step "gate corpora ready"
}

# Downloads and the sanitizer check do not need each other's results. Build only after all
# have succeeded; wait for every child even when one fails, so no installer outlives setup.
# Bound the preparation shells as well as their individual children. Observed
# complete setup was 32.8s; 30m also allows cold downloads and toolchain work.
export started repository cloudSource tools gate run gateInputs gateInputsRoot markdownDependencies goArchitecture nodeArchitecture llvmArchitecture ADAMIC_BOUNDED_REPORT
export -f bounded step prepareGo prepareGoAndSignal prepareClang prepareNode prepareSubmodules prepareGateCorpora
for boundedTool in cat awk mkdir install mktemp realpath uname ls sort dirname ln mv grep nproc sha256sum cut head; do
 export -f "$boundedTool"
done
bounded 1800 bash -c "set -euo pipefail; prepareGoAndSignal" & goProcess=$!
bounded 1800 bash -c "set -euo pipefail; prepareClang" & clangProcess=$!
bounded 1800 bash -c "set -euo pipefail; prepareNode" & nodeProcess=$!
bounded 1800 bash -c "set -euo pipefail; prepareSubmodules" & submoduleProcess=$!
prepareProcesses=("$goProcess" "$clangProcess" "$nodeProcess" "$submoduleProcess")
if "$gateInputs"; then
	bounded 1800 bash -c "set -euo pipefail; prepareGateCorpora" & prepareProcesses+=("$!")
fi
failed=0
for process in "${prepareProcesses[@]}"; do
	wait "$process" || failed=1
done
[ "$failed" = 0 ] || exit 1
[ -x "$tools/go/bin/go" ] && export PATH="$tools/go/bin:$PATH"

# One file every shell sources: the agent's shell in Codex is a different session from this one.
cat > "$tools/env.sh" << ENV
export PATH="$tools/bin:$([ -x "$tools/go/bin/go" ] && echo "$tools/go/bin:")\$PATH"
export GOTOOLCHAIN=auto
export TMPDIR=$gate
export ADAMIC_MARKDOWNWIDTH_DEPS="$markdownDependencies"
ENV
if "$gateInputs"; then
	# Node, Go and submodules are now ready; build this seat's checker archive.
	export PATH="$tools/bin:$PATH"
	bounded 1800 python3 "$cloudSource/../internal/boundedrun/python.py" "$cloudSource/setup-gate-inputs.py" archive "$repository" "$gateInputsRoot" "$tools/bin/node" > "$run/gate-archive.log" 2>&1 || { cat "$run/gate-archive.log"; exit 1; }
	cat "$run/gate-archive.log"
	bounded 30 python3 "$cloudSource/setup-gate-inputs.py" env "$repository" "$gateInputsRoot" "$tools/bin/node" >> "$tools/env.sh"
else
	# A later ordinary setup must not inherit a previous opt-in gate seat.
	echo "unset ADAMIC_ESTREE_LIBRARY ADAMIC_YAML_LIBRARY ADAMIC_TS_PRETTIER ADAMIC_TYPESCRIPT_SOURCE ADAMIC_CSS_LIBRARY ADAMIC_GRAPHQL_LIBRARY ADAMIC_MEDIA_QUERY_LIBRARY ADAMIC_SELECTOR_LIBRARY ADAMIC_VALUES_LIBRARY ADAMIC_JSON_PRETTIER ADAMIC_CSS_PRINTER_LIBRARY ADAMIC_GITIGNORE_LARGEST ADAMIC_CLANG_TSGO_ARCHIVE" >> "$tools/env.sh"
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
	bounded 600 go list -deps -export "${listArguments[@]}" -json ./... > "$run/packages.json" 2> "$run/list.log" || { cat "$run/list.log"; exit 1; }
	key=$(bounded 30 python3 "$cloudSource/setup-key.py" "$repository" "$run/packages.json" "$warmTests") || key=""
fi
if [ -n "$key" ] && [ "$(cat "$stamp" 2> /dev/null || true)" = "$key" ]; then
	step "go build skipped (validated warming stamp)"
	"$warmTests" && step "test binaries skipped (validated warming stamp)"
else
	bounded 600 go build "${buildArguments[@]}" ./... > "$run/build.log" 2>&1 || { cat "$run/build.log"; exit 1; }
	step "go build ready"
	if "$warmTests"; then
		bounded 600 go test "${buildArguments[@]}" -count=1 -run '^$' ./... > "$run/tests.log" 2>&1 || { cat "$run/tests.log"; exit 1; }
		step "test binaries warm"
	fi
	if [ -n "$key" ]; then
		# Build/link adds cache entries. Publish only the key of the completed cache, atomically.
		if bounded 30 python3 "$cloudSource/setup-key.py" "$repository" "$run/packages.json" "$warmTests" > "$run/stamp"; then
			mv "$run/stamp" "$stamp"
		fi
	fi
fi
"$warmTests" || step "test binaries deferred (use --warm-tests)"
step "build cache warm"

cpuQuota=$(cat /sys/fs/cgroup/cpu.max 2> /dev/null || echo unknown)
memory=$(awk '/MemTotal/ {printf "%.1f GB", $2 / 1048576}' /proc/meminfo)
echo "setup: build-flags commit=$(bounded 30 git -C "$repository" rev-parse HEAD) nproc=$(nproc) cpu.max=$cpuQuota go=$(bounded 30 go version) clang=$(bounded 30 clang --version | head -n 1) node=$(bounded 30 node --version) cached=$([ "${ADAMIC_GATE_UNCACHED:-0}" = 1 ] && echo no || echo yes) warm-tests=$warmTests gate-inputs=$gateInputs load-before=$loadBefore load-after=$(cat /proc/loadavg)"
step "done on $(nproc) processors (cgroup cpu.max: $cpuQuota), $memory"
echo "setup: source $tools/env.sh"
echo "setup: logs $run"
