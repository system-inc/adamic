#!/usr/bin/env bash
# cloud/block-report/environment.sh lays, once and with network, everything a sealed block job needs to build and test
# one Adamic commit with none: the toolchain, every Go module, every npm tarball and the pinned TypeScript source, under
# one fixed prefix, read-only, with a manifest of digests. It runs on the image builder, as root or a builder user, at
# the block's commit. The job (cloud/block-report/run.sh) runs under PrivateNetwork=yes and sources the
# environment.sh this writes, so anything missing here fails there by name, never by reaching out.
#
#   bash cloud/block-report/environment.sh <adamic checkout at the block's sha> [<prefix>, default /opt/adamic-block]
#
# Under the prefix:
#
#   tools/              cloud/setup.sh --wasi-sdk's tool directory (Go, clang, Node 24, tsc, WASI SDK) and its env.sh
#   gomodcache/         every module the workspace's build and tests read, and the Go toolchain if go.work's had to be
#                       fetched
#   npmcache/           every tarball each tracked package-lock.json names
#   npm/                npm from its pin in cloud/markdown-width/npm-bootstrap.json, when the toolchain has none
#   typescript-source/  stage3/source.json's repository at its tag, for ADAMIC_TYPESCRIPT_SOURCE
#   logs/               each step's whole output, kept on failure
#   manifest.json       each piece's path and digest, the commit, and the tools' versions
#   environment.sh      what the job sources
#
# The checkout's submodules must be laid at their gitlinks already; this checks them and never fetches them.
set -euo pipefail

[ "$(uname)" = Linux ] || { echo "block-env: Linux image builders only, never a Mac" >&2; exit 2; }
case $# in
	1 | 2) ;;
	*) echo "usage: bash cloud/block-report/environment.sh <adamic checkout> [<prefix>]" >&2; exit 2 ;;
esac
checkout=$(cd "$1" && pwd)
prefix=${2:-/opt/adamic-block}
case "$prefix" in
	/?*) prefix=${prefix%/} ;;
	*) echo "block-env: the prefix must be an absolute path, not '$prefix'" >&2; exit 2 ;;
esac

# What the builder's own shell holds must not reach the lay: each of these is set below, or by setup.sh's env.sh.
unset GOFLAGS GOWORK GOTOOLCHAIN GOPROXY GOSUMDB GOMODCACHE GOCACHEPROG NPM_CONFIG_CACHE npm_config_cache \
	npm_config_offline ADAMIC_TOOLS ADAMIC_TYPESCRIPT_SOURCE WASI_SYSROOT
# setup.sh keeps any Go, Node or clang it finds on PATH, so a builder whose shell already carries a toolchain would lay
# paths outside the prefix into the job's environment, and the job would find them missing. Only the system's own
# directories stay (a trial on Home, Nov 12: clang came from ~/.adamic-tools and the prefix's tsc check failed).
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

started=$EPOCHREALTIME
stepStarted=$started
stepName=start
log=""
scratch=""

# begin <name>: the step a failure is named by, and the log its commands write to.
begin() {
	stepName=$1
	stepStarted=$EPOCHREALTIME
	log="$prefix/logs/$1.log"
	: > "$log"
}

# step <what it did>: one line per step, with the step's own time and the time since the start.
step() {
	local elapsed total
	elapsed=$(awk -v start="$stepStarted" -v now="$EPOCHREALTIME" 'BEGIN {printf "%.3f", now - start}')
	total=$(awk -v start="$started" -v now="$EPOCHREALTIME" 'BEGIN {printf "%.3f", now - start}')
	echo "block-env: $stepName: $1 (${elapsed}s, ${total}s in all)"
}

# Only a directory under this run's own mktemp scratch is ever removed whole.
removeScratch() {
	case "$1" in
		/tmp/block-env.?*) rm -rf -- "$1" ;;
		*) echo "block-env: refusing to remove '$1', which isn't this run's scratch" >&2; return 1 ;;
	esac
}

# A failure names its step and shows the end of the step's log; the whole log stays under the prefix.
finish() {
	local code=$?
	if [ -n "$scratch" ]; then
		removeScratch "$scratch" || true
	fi
	[ "$code" = 0 ] && return 0
	echo "block-env: failed in $stepName (exit $code)" >&2
	if [ -n "$log" ] && [ -s "$log" ]; then
		echo "block-env: the last lines of $log:" >&2
		tail -n 20 "$log" >&2
	fi
}
trap finish EXIT

# The tree's digest is over its files' contents and paths relative to it, sorted bytewise, so it reads the same
# wherever the tree is and whatever the locale. Symbolic links are not followed.
treeDigest() {
	(cd "$1" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 -r sha256sum | sha256sum | cut -d' ' -f1)
}

mkdir -p "$prefix/logs"
scratch=$(mktemp -d /tmp/block-env.XXXXXX)
gomodcache="$prefix/gomodcache"
npmcache="$prefix/npmcache"
npmHome="$prefix/npm"
typescriptSource="$prefix/typescript-source"
# A second run writes into what the first left read-only.
for laid in "$gomodcache" "$npmcache" "$npmHome" "$typescriptSource"; do
	[ ! -e "$laid" ] || chmod -R u+w "$laid"
done

begin checkout
commit=$(git -C "$checkout" rev-parse --verify 'HEAD^{commit}')
dirty=$(git -C "$checkout" status --porcelain --untracked-files=no --ignore-submodules=all)
[ -z "$dirty" ] || { echo "block-env: $checkout has changes to tracked files, so it isn't the commit's tree: $(printf '%s' "$dirty" | head -n 3 | tr '\n' ' ')" >&2; exit 1; }
step "$checkout at $commit, clean"

# Submodules, checked before setup.sh runs, since setup.sh would fetch any the caller didn't lay and hide that.
begin submodules
git -C "$checkout" submodule status --recursive > "$log" 2>&1
submodules=""
while IFS= read -r line; do
	[ -n "$line" ] || continue
	state=${line:0:1}
	read -r sha path _ <<< "${line:1}"
	[ "$state" = " " ] || { echo "block-env: submodule $path isn't laid at its gitlink (git submodule status says '$state')" >&2; exit 1; }
	submodules="$submodules$path"$'\t'"$sha"$'\n'
done < "$log"
[ -n "$submodules" ] || { echo "block-env: $checkout has no submodules laid; cohere and cohere/TypeScript are needed" >&2; exit 1; }
step "$(printf '%s' "$submodules" | awk -F'\t' '{printf "%s%s at %s", (NR > 1 ? ", " : ""), $1, $2}')"

# The toolchain is setup.sh's own: it puts its tools in ADAMIC_TOOLS when that directory can be made. GOMODCACHE
# points into the prefix already, so a Go toolchain switch to go.work's version, and the modules setup.sh's warming
# build fetches, land where the job reads them.
begin toolchain
tools="$prefix/tools"
environmentScript="$tools/env.sh"
mkdir -p "$tools"
ADAMIC_TOOLS="$tools" GOMODCACHE="$gomodcache" bash "$checkout/cloud/setup.sh" --wasi-sdk > "$log" 2>&1
# setup.sh falls back to ~/.adamic-tools when ADAMIC_TOOLS can't be made, and the job would never find that one.
grep -qxF "setup: source $environmentScript" "$log" || { echo "block-env: setup.sh didn't lay its tools in $tools" >&2; exit 1; }
[ -r "$environmentScript" ] || { echo "block-env: setup.sh left no env.sh at $environmentScript" >&2; exit 1; }
# shellcheck disable=SC1090
source "$environmentScript"
export GOMODCACHE="$gomodcache"
# The go that answers for go.work may be one GOTOOLCHAIN=auto fetched into the module cache, not the one on PATH; from
# here on it's the only go, and GOTOOLCHAIN=local proves it's new enough for go.work with no fetch.
goRoot=$(cd "$checkout" && GOTOOLCHAIN=auto go env GOROOT)
export PATH="$goRoot/bin:$PATH" GOTOOLCHAIN=local
(cd "$checkout" && go list -m) >> "$log" 2>&1 || { echo "block-env: $goRoot/bin/go can't load go.work at GOTOOLCHAIN=local" >&2; exit 1; }
goVersion=$(go version)
nodeVersion=$(node --version)
command -v clang > /dev/null || { echo "block-env: setup.sh's env.sh puts no clang on PATH" >&2; exit 1; }
clangPath=$(readlink -f "$(command -v clang)")
clangVersion=$(clang --version | head -n 1)
step "env.sh $environmentScript; $goVersion at $goRoot; node $nodeVersion; $clangVersion"

# Go modules. Each workspace module's own graph, read whole from its go.mod with the workspace off, then the
# workspace's build and tests as the job loads them, so a module only a test imports is here too. The proxy is
# setup.sh's, falling back to the module's own source on any error; go.sum and go.work.sum check every hash either way.
begin gomodules
export GOPROXY="https://proxy.golang.org|direct"
modules=$(cd "$checkout" && go work edit -json | python3 -I -c 'import json, sys; print("\n".join(use["DiskPath"] for use in json.load(sys.stdin).get("Use") or []))')
[ -n "$modules" ] || { echo "block-env: $checkout/go.work uses no modules" >&2; exit 1; }
while IFS= read -r module; do
	echo "== go mod download all in $module" >> "$log"
	(cd "$checkout/$module" && GOWORK=off go mod download all) >> "$log" 2>&1 ||
		{ echo "block-env: go mod download all failed in $module" >&2; exit 1; }
done <<< "$modules"
echo "== go list -deps -test ./..." >> "$log"
(cd "$checkout" && GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly" go list -deps -test ./... > /dev/null) >> "$log" 2>&1
# The same load with the network refused, as the job will meet it: a module still missing fails here, by name.
echo "== go list -deps -test ./... with GOPROXY=off" >> "$log"
(cd "$checkout" && GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly" GOPROXY=off GOSUMDB=off go list -deps -test ./... > /dev/null) >> "$log" 2>&1 ||
	{ echo "block-env: the workspace doesn't load from $gomodcache alone" >&2; exit 1; }
step "$(printf '%s' "$modules" | tr '\n' ' ')downloaded into $gomodcache, and the workspace's tests load from it offline"

# cohere's own workspace, which its Go-oracle tests load from inside cohere (go list and go test -c in cohere/), is a
# second module graph adamic's go.work never reads: on azure-block-1 they failed on golang.org/x/sys@v0.33.0 (J-z5p8ce2).
# The same download, then the same load online and with the network refused.
begin cohere-gomodules
cohereModules=$(cd "$checkout/cohere" && go work edit -json | python3 -I -c 'import json, sys; print("\n".join(use["DiskPath"] for use in json.load(sys.stdin).get("Use") or []))')
[ -n "$cohereModules" ] || { echo "block-env: $checkout/cohere/go.work uses no modules" >&2; exit 1; }
while IFS= read -r module; do
	echo "== go mod download all in cohere/$module" >> "$log"
	(cd "$checkout/cohere/$module" && GOWORK=off go mod download all) >> "$log" 2>&1 ||
		{ echo "block-env: go mod download all failed in cohere/$module" >&2; exit 1; }
done <<< "$cohereModules"
echo "== go list -deps -test ./... in cohere" >> "$log"
(cd "$checkout/cohere" && GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly" go list -deps -test ./... > /dev/null) >> "$log" 2>&1
echo "== go list -deps -test ./... in cohere with GOPROXY=off" >> "$log"
(cd "$checkout/cohere" && GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly" GOPROXY=off GOSUMDB=off go list -deps -test ./... > /dev/null) >> "$log" 2>&1 ||
	{ echo "block-env: cohere's workspace doesn't load from $gomodcache alone" >&2; exit 1; }
step "cohere's $(printf '%s' "$cohereModules" | wc -l | tr -d ' ') modules downloaded, and its workspace's tests load offline"

# npm. setup.sh lays node alone, so npm comes from its pin, checked against its integrity before it unpacks, as
# cloud/meter-daily/run.sh does it. npm's own logs go beside ours, so the cache holds tarballs and nothing else.
begin npm
export NPM_CONFIG_CACHE="$npmcache" npm_config_logs_dir="$prefix/logs/npm" npm_config_update_notifier=false
npmBin=""
if [ -x "$tools/bin/npm" ]; then
	npmFrom="$tools/bin/npm, the toolchain's"
else
	bootstrap="$checkout/cloud/markdown-width/npm-bootstrap.json"
	npmVersion=$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["version"])' "$bootstrap")
	if [ ! -x "$npmHome/bin/npm" ] || [ "$(cat "$npmHome/version" 2> /dev/null)" != "$npmVersion" ]; then
		[ ! -e "$npmHome" ] || { echo "block-env: $npmHome holds an npm that isn't $npmVersion; remove it by hand" >&2; exit 1; }
		staged="$scratch/npm"
		mkdir -p "$staged/bin"
		curl -fsSL "$(python3 -I -c 'import json, sys; print(json.load(open(sys.argv[1]))["url"])' "$bootstrap")" -o "$scratch/npm.tgz" 2>> "$log"
		python3 -I - "$bootstrap" "$scratch/npm.tgz" << 'PYTHON'
import base64, hashlib, json, sys
algorithm, expected = json.load(open(sys.argv[1]))['integrity'].split('-', 1)
actual = base64.b64encode(hashlib.sha512(open(sys.argv[2], 'rb').read()).digest()).decode()
if algorithm != 'sha512' or actual != expected:
    sys.exit('npm bootstrap integrity mismatch')
PYTHON
		tar --no-same-owner -xzf "$scratch/npm.tgz" -C "$staged"
		printf '#!/bin/sh\nexec node %q "$@"\n' "$npmHome/package/bin/npm-cli.js" > "$staged/bin/npm"
		chmod 755 "$staged/bin/npm"
		printf '%s\n' "$npmVersion" > "$staged/version"
		mv "$staged" "$npmHome"
	fi
	npmBin="$npmHome/bin"
	export PATH="$npmBin:$PATH"
	npmFrom="$npmHome, npm $npmVersion from its pin in cloud/markdown-width/npm-bootstrap.json"
fi
npmVersion=$(npm --version)
step "$npmFrom"

# A lockfile counts when it's named package-lock.json with a tracked package.json beside it; a fixture that only looks
# like one (an audit's lock, an oracle's renamed one) is no install. Each install runs in a copy of the two files, so
# the checkout never gets a node_modules, then again offline in a fresh copy to prove the cache holds every tarball.
begin npmcache
locks=$(git -C "$checkout" ls-files '*package-lock.json')
installed="" skipped=""
index=0
while IFS= read -r lock; do
	[ -n "$lock" ] || continue
	directory=$(dirname "$lock")
	case "/$lock/" in */node_modules/*) continue ;; esac
	if [ "$(basename "$lock")" != package-lock.json ] || ! git -C "$checkout" ls-files --error-unmatch "$directory/package.json" > /dev/null 2>&1; then
		skipped="$skipped $lock"
		continue
	fi
	for mode in online offline; do
		index=$((index + 1))
		copy="$scratch/npm-$index"
		mkdir "$copy"
		cp "$checkout/$directory/package.json" "$checkout/$directory/package-lock.json" "$copy/"
		if git -C "$checkout" ls-files --error-unmatch "$directory/.npmrc" > /dev/null 2>&1; then
			cp "$checkout/$directory/.npmrc" "$copy/"
		fi
		offline=()
		[ "$mode" = online ] || offline=(--offline)
		echo "== npm ci ${offline[*]} for $lock" >> "$log"
		(cd "$copy" && npm ci --ignore-scripts --no-audit --no-fund "${offline[@]}") >> "$log" 2>&1 ||
			{ echo "block-env: npm ci ${offline[*]} failed for $lock" >&2; exit 1; }
		removeScratch "$copy"
	done
	installed="$installed $lock"
done <<< "$locks"
[ -n "$installed" ] || { echo "block-env: no tracked package-lock.json to cache" >&2; exit 1; }
step "cached and installed offline:$installed; skipped, not npm projects:${skipped:- none}"

# The gate's TypeScript corpus. setup.sh lays no gate inputs (write-env.sh only keeps a box's own), so the tag
# stage3/source.json pins is cloned here, as stage3/meter/twice-daily.sh does when none is given, and held to the
# commit the pin records.
begin typescript
sourcePin="$checkout/stage3/source.json"
pin=$(python3 -I -c 'import json, sys; pin = json.load(open(sys.argv[1])); print(pin["repository"], pin["tag"], pin["commit"])' "$sourcePin")
read -r typescriptRepository typescriptTag typescriptCommit <<< "$pin"
if [ -e "$typescriptSource" ]; then
	laidCommit=$(git -C "$typescriptSource" rev-parse --verify 'HEAD^{commit}' 2> /dev/null || true)
	[ "$laidCommit" = "$typescriptCommit" ] || { echo "block-env: $typescriptSource is at '$laidCommit', not $typescriptTag's $typescriptCommit; remove it by hand" >&2; exit 1; }
	typescriptFrom="already laid"
else
	git clone --depth 1 --branch "$typescriptTag" "$typescriptRepository" "$scratch/typescript-source" >> "$log" 2>&1
	clonedCommit=$(git -C "$scratch/typescript-source" rev-parse --verify 'HEAD^{commit}')
	[ "$clonedCommit" = "$typescriptCommit" ] || { echo "block-env: $typescriptRepository's $typescriptTag is $clonedCommit, but stage3/source.json pins $typescriptCommit" >&2; exit 1; }
	mv "$scratch/typescript-source" "$typescriptSource"
	typescriptFrom="cloned"
fi
step "$typescriptSource $typescriptFrom, $typescriptTag at $typescriptCommit"

# The environment the job sources, and the manifest. It checks every piece is here before the job runs a thing, and
# keeps env.sh's GOFLAGS (-buildvcs=false -trimpath, the house's) with -mod=readonly added once.
begin manifest
markdownDependencies=${ADAMIC_MARKDOWNWIDTH_DEPS:-}
wasiSysroot=${WASI_SYSROOT:-}
{
	echo "# Written by cloud/block-report/environment.sh at $commit. Sourced by the sealed block job: everything it reads"
	echo "# is under $prefix, and nothing is fetched."
	for required in "$environmentScript" "$goRoot/bin/go" "$gomodcache" "$npmcache" "$typescriptSource"; do
		printf '[ -e %q ] || { echo %q >&2; return 1 2> /dev/null || exit 1; }\n' "$required" "block-env: $required is missing from the image"
	done
	printf 'source %q\n' "$environmentScript"
	printf 'export PATH=%q:"$PATH"\n' "$goRoot/bin${npmBin:+:$npmBin}"
	printf 'export GOMODCACHE=%q\n' "$gomodcache"
	printf '%s\n' 'case " ${GOFLAGS:-} " in *" -mod=readonly "*) ;; *) export GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly" ;; esac'
	printf 'export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local\n'
	printf 'export NPM_CONFIG_CACHE=%q npm_config_offline=true\n' "$npmcache"
	printf 'export ADAMIC_TYPESCRIPT_SOURCE=%q\n' "$typescriptSource"
	# run.sh installs exactly these from the cache, offline: each a directory holding a tracked lockfile.
	printf 'export BLOCK_REPORT_NPM_PROJECTS=%q\n' "$(for lock in $installed; do dirname "$lock"; done | tr '\n' ' ' | sed 's/ $//')"
} > "$prefix/environment.sh.writing"
mv -f "$prefix/environment.sh.writing" "$prefix/environment.sh"

# The submodules are checked again: nothing between the first look and here may have moved them.
[ "$(git -C "$checkout" submodule status --recursive | awk '{sub(/^[ +U-]/, ""); print $2 "\t" $1}')" = "$(printf '%s' "$submodules")" ] ||
	{ echo "block-env: a submodule moved while the environment was laid" >&2; exit 1; }
pieces="$scratch/pieces.tsv"
: > "$pieces"
# piece <name> <path> <what> <Tree|File>: a digest is taken by assignment, so one that fails stops the step.
piece() {
	local digest
	if [ "$4" = Tree ]; then
		digest=$(treeDigest "$2")
	else
		digest=$(sha256sum "$2" | cut -d' ' -f1)
	fi
	printf '%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$digest" >> "$pieces"
}
piece tools "$tools" "cloud/setup.sh --wasi-sdk's tool directory: Go unless the image had one, clang links, Node 24, tsc, WASI SDK, markdown-width's dependencies" Tree
piece env.sh "$environmentScript" "setup.sh's env.sh, which environment.sh sources" File
case "$goRoot/" in
	"$tools"/*) ;;
	*) piece go "$goRoot" "the Go toolchain go.work runs on, outside tools" Tree ;;
esac
piece clang "$clangPath" "the clang tools/bin/clang resolves to; this digest is of the binary alone" File
case "$markdownDependencies/" in
	/ | "$tools"/*) ;;
	*) piece markdown-width "$markdownDependencies" "cloud/markdown-width's installed dependencies, outside tools" Tree ;;
esac
piece gomodcache "$gomodcache" "the Go module cache for the workspace's build and tests" Tree
piece npmcache "$npmcache" "the npm cache holding every tracked package-lock.json's tarballs" Tree
[ -z "$npmBin" ] || piece npm "$npmHome" "npm $npmVersion from its pin" Tree
piece typescript-source "$typescriptSource" "$typescriptRepository at $typescriptTag ($typescriptCommit), for ADAMIC_TYPESCRIPT_SOURCE" Tree
piece environment.sh "$prefix/environment.sh" "what the sealed job sources" File
printf '%s' "$submodules" > "$scratch/submodules.tsv"
python3 -I - "$prefix/manifest.json.writing" "$pieces" "$commit" "$checkout" "$prefix" "$environmentScript" "$prefix/environment.sh" \
	"$goVersion" "$goRoot" "$nodeVersion" "$clangVersion" "$npmVersion" "$wasiSysroot" "$(date -u +%FT%TZ)" "$scratch/submodules.tsv" << 'PYTHON'
import json, sys
(output, pieces, commit, checkout, prefix, env_sh, environment, go, go_root, node, clang, npm, wasi_sysroot,
 laid, submodule_list) = sys.argv[1:]
submodules = [dict(zip(('path', 'commit'), line.split('\t'))) for line in open(submodule_list).read().splitlines() if line]
rows = [dict(zip(('name', 'path', 'what', 'sha256'), line.split('\t'))) for line in open(pieces).read().splitlines()]
with open(output, 'w') as manifest:
    json.dump({'commit': commit, 'checkout': checkout, 'prefix': prefix, 'laid': laid, 'envSh': env_sh,
               'environment': environment, 'goVersion': go, 'goRoot': go_root, 'nodeVersion': node,
               'clangVersion': clang, 'npmVersion': npm, 'wasiSysroot': wasi_sysroot, 'submodules': submodules,
               'pieces': rows,
               'digest': 'sha256 of `find . -type f | LC_ALL=C sort | xargs sha256sum` run inside a directory; a file\'s own sha256'},
              manifest, indent=2)
    manifest.write('\n')
PYTHON
mv -f "$prefix/manifest.json.writing" "$prefix/manifest.json"
step "$prefix/manifest.json and $prefix/environment.sh, $(wc -l < "$pieces" | tr -d ' ') pieces"

# Read-only for everyone, the owner too, so the job can't change what the manifest digested. Go's extracted modules
# are read-only already; its download cache and the rest aren't. tools stays as setup.sh made it.
begin readonly
for laid in "$gomodcache" "$npmcache" "$typescriptSource"; do
	chmod -R a-w "$laid"
done
[ -z "$npmBin" ] || chmod -R a-w "$npmHome"
chmod a-w "$prefix/manifest.json" "$prefix/environment.sh"
chmod go-w "$prefix"
step "gomodcache, npmcache, typescript-source${npmBin:+, npm} read-only; source $prefix/environment.sh"
