#!/usr/bin/env bash
# Merges one finished branch into an area branch and pushes the area if the area's fast tests pass.
# The area's owner runs it, on their Mac or in a cloud box. It never touches main.
#
# Fast tests, all uncached (ADAMIC_GATE_UNCACHED=1, -count=1):
#   - gofmt -l on the Go files the merge changed, go vet and go test on every package it changed;
#   - internal/oracle whole when the merge changes the compiler, the runtime or the oracle itself,
#     or only the fixtures it changed when it changes nothing but fixtures;
#   - for an area whose oracle is test262 (areas.tsv), cmd/adamic-test262 on the filters you name,
#     which must report zero disagreements (fail) in every filter. Set ADAMIC_TEST262 to a test262
#     checkout at the pinned commit.
# A branch that passes against its outside oracle merges with no reader. Branches without one
# (language, memory, concurrency) get a reader before they come here; that's the owner's call.
#
# Each area keeps one worktree of its own, under ADAMIC_AREA_WORKTREES (default ~/.adamic-areas),
# reused across runs and locked while one runs, on this machine and as locks/area-<area> on origin. The script refuses a worktree with local changes.
#
# usage: cloud/integration/area-merge.sh <area> <branch> <sha> [--no-push] [test262 filter...]
#   e.g. cloud/integration/area-merge.sh library codex/library-string-2 e911e46 built-ins/String
set -euo pipefail

if [ "$#" -lt 3 ]; then
	echo "usage: $0 <area> <branch> <sha> [--no-push] [test262 filter...]" >&2
	exit 2
fi
area=$1
branch=$2
requested=$3
shift 3
push=yes
filters=()
for argument in "$@"; do
	if [ "$argument" = "--no-push" ]; then
		push=no
	else
		filters+=("$argument")
	fi
done

directory=$(cd "$(dirname "$0")" && pwd)
oracle=$(awk -F'\t' -v area="$area" '$1 == area { print $3 }' "$directory/areas.tsv")
if [ -z "$oracle" ]; then
	echo "refused: no area named $area in areas.tsv" >&2
	exit 2
fi

# A fresh cloud clone fetches only main, so name the two refs this needs.
git fetch -q origin "+refs/heads/$branch:refs/remotes/origin/$branch" "+refs/heads/area/$area:refs/remotes/origin/area/$area"
sha=$(git rev-parse --verify "${requested}^{commit}")
branchTip=$(git rev-parse --verify "refs/remotes/origin/${branch}")
if [ "$branchTip" != "$sha" ]; then
	echo "refused: origin/$branch is ${branchTip:0:8}, not ${sha:0:8}; name the tip you mean" >&2
	exit 1
fi
areaTip=$(git rev-parse --verify "refs/remotes/origin/area/$area")
if git merge-base --is-ancestor "$sha" "$areaTip"; then
	echo "area/$area ${areaTip:0:8} already holds $branch ${sha:0:8}"
	exit 0
fi

worktrees=${ADAMIC_AREA_WORKTREES:-$HOME/.adamic-areas}
worktree="$worktrees/$area"
lock="$worktrees/$area.lock"
mkdir -p "$worktrees"
if ! mkdir "$lock" 2>/dev/null; then
	echo "refused: another merge into area/$area is running ($lock); wait for it" >&2
	exit 1
fi
# The same lock on origin, so merge-back run from any machine sees this merge and leaves the area alone.
remoteLock="refs/heads/locks/area-$area"
# The lease with an empty expected value creates the ref only if no other machine holds it. The lock
# is a commit of its own naming this run, because pushing a sha a ref already holds succeeds.
lockCommit=$(git commit-tree "${areaTip}^{tree}" -p "$areaTip" -m "Lock area/$area for a merge: $(hostname) pid $$ at $(date -u +%Y-%m-%dT%H:%M:%SZ)")
if ! git push -q --force-with-lease="${remoteLock}:" origin "${lockCommit}:${remoteLock}" 2>/dev/null; then
	rmdir "$lock"
	echo "refused: another machine is merging into area/$area (${remoteLock#refs/heads/} on origin); wait for it" >&2
	exit 1
fi
trap 'git push -q --force-with-lease="${remoteLock}:${lockCommit}" origin ":${remoteLock}" 2>/dev/null; rmdir "$lock"' EXIT

if [ -d "$worktree" ]; then
	if [ -n "$(git -C "$worktree" status --porcelain --untracked-files=no)" ]; then
		echo "refused: $worktree has local changes; look at them before anything moves" >&2
		exit 1
	fi
	git -C "$worktree" switch -q --detach "$areaTip"
else
	git worktree add -q --detach "$worktree" "$areaTip"
fi
git -C "$worktree" submodule update -q --init --recursive --depth 1

if ! git -C "$worktree" merge -q --no-ff -m "Merge $branch at ${sha:0:8} into area/$area" "$sha"; then
	echo "conflict merging $branch ${sha:0:8} into area/$area ${areaTip:0:8}:"
	git -C "$worktree" diff --name-only --diff-filter=U
	git -C "$worktree" merge --abort
	exit 3
fi
# The merge can move a submodule pin (a cohere bump), so the checkout has to follow it before
# anything builds.
git -C "$worktree" submodule update -q --init --recursive --depth 1

logs=$(mktemp -d)
changed=$(git -C "$worktree" diff --name-only "$areaTip" HEAD)
packages=()
compiler=no
fixtures=()
while IFS= read -r file; do
	[ -n "$file" ] || continue
	case "$file" in
	internal/oracle/testdata/*) fixtures+=("$(basename "$file")") ;;
	internal/lower/* | internal/native/* | internal/ir/* | internal/flow/* | internal/javascript/* | internal/load/* | internal/oracle/*) compiler=yes ;;
	esac
	package=$(dirname "$file")
	# Go never builds a testdata directory as a package; its files belong to the package above it.
	case "$package" in
	testdata | testdata/*) package=. ;;
	*/testdata | */testdata/*) package=${package%%/testdata*} ;;
	esac
	while [ "$package" != "." ] && ! ls "$worktree/$package"/*.go >/dev/null 2>&1; do
		package=$(dirname "$package")
	done
	if [ "$package" != "." ] && [ "$package" != "internal/oracle" ]; then
		packages+=("./$package")
	fi
done <<<"$changed"
# Other packages walk every oracle program (flow, fresh, lower, native and more), so a changed
# fixture runs them too; they're found by what their tests read, not listed by hand.
if [ "${#fixtures[@]}" -gt 0 ]; then
	while IFS= read -r reader; do
		[ -n "$reader" ] && [ "$reader" != "internal/oracle" ] && packages+=("./$reader")
	done < <(git -C "$worktree" grep -l "oracle/testdata" -- '*_test.go' | xargs -n1 dirname | sort -u)
fi
if [ "${#packages[@]}" -gt 0 ]; then
	packages=($(printf '%s\n' "${packages[@]}" | sort -u))
fi
if [ "${#packages[@]}" -gt 0 ]; then
	# A package whose every file sits behind a build tag isn't built by go test ./..., and naming it
	# fails, so leave those out the way the whole gate does.
	packages=($(cd "$worktree" && go list -e -f '{{.ImportPath}}{{"\t"}}{{if .Error}}{{.Error.Err}}{{end}}' "${packages[@]}" 2>/dev/null | grep -v 'build constraints exclude all Go files' | cut -f1 | sed 's#^github.com/system-inc/adamic/#./#' || true))
fi

status=0
cd "$worktree"
export ADAMIC_GATE_UNCACHED=1
goFiles=$(printf '%s\n' "$changed" | grep '\.go$' | while IFS= read -r file; do [ -f "$file" ] && echo "$file"; done || true)
if [ -n "$goFiles" ]; then
	unformatted=$(gofmt -l $goFiles)
	if [ -n "$unformatted" ]; then
		echo "gofmt: $unformatted"
		status=1
	fi
fi
if [ "${#packages[@]}" -gt 0 ]; then
	go vet "${packages[@]}" >"$logs/vet.log" 2>&1 || { echo "go vet failed: $logs/vet.log"; status=1; }
fi
oraclePattern=""
if [ "$compiler" = no ] && [ "${#fixtures[@]}" -gt 0 ]; then
	oraclePattern="^(TestNativeAgreesWithNode|TestCountsAreRecorded)\$/internal/oracle/testdata/($(printf '%s\n' "${fixtures[@]}" | sort -u | sed 's/[.]/[.]/g' | paste -sd '|' -))\$"
fi

# runTests runs the area's fast tests on whatever is checked out and writes the failing tests, one
# "package test" per line, to $logs/<phase>-failures.txt.
runTests() {
	local phase=$1
	if [ "${#packages[@]}" -gt 0 ]; then
		go test -json -count=1 -timeout 60m "${packages[@]}" >"$logs/$phase-packages.json" 2>&1 || true
	fi
	if [ "$compiler" = yes ]; then
		go test -json -count=1 -timeout 60m ./internal/oracle >"$logs/$phase-oracle.json" 2>&1 || true
	elif [ -n "$oraclePattern" ]; then
		go test -json -count=1 -timeout 60m ./internal/oracle -run "$oraclePattern" >"$logs/$phase-oracle.json" 2>&1 || true
	fi
	python3 -c 'import json,sys
for path in sys.argv[1:]:
    for line in open(path, errors="replace"):
        try:
            event = json.loads(line)
        except ValueError:
            continue
        if event.get("Action") == "fail":
            print(event.get("Package", "?"), event.get("Test", "(the package itself)"))' "$logs/$phase-"*.json 2>/dev/null | sort -u >"$logs/$phase-failures.txt"
}

# A failure the area tip already has on this machine (macOS's last-bit Math, no detect_leaks) is
# named but doesn't hold the merge. Only a failure the merge brings does.
merged=$(git rev-parse HEAD)
if [ "${#packages[@]}" -gt 0 ] || [ "$compiler" = yes ] || [ -n "$oraclePattern" ]; then
	echo "ran: go test on ${packages[*]:-no packages}$([ "$compiler" = yes ] && echo ', internal/oracle whole')$([ -n "$oraclePattern" ] && echo ', internal/oracle on the changed fixtures')"
	runTests merged
	if [ -s "$logs/merged-failures.txt" ]; then
		git switch -q --detach "$areaTip"
		git submodule update -q --init --recursive --depth 1
		runTests base
		git switch -q --detach "$merged"
		git submodule update -q --init --recursive --depth 1
		comm -23 "$logs/merged-failures.txt" "$logs/base-failures.txt" >"$logs/new-failures.txt"
		echo "already failing on area/$area ${areaTip:0:8} here, not held against the merge: $(comm -12 "$logs/merged-failures.txt" "$logs/base-failures.txt" | wc -l | tr -d ' ') ($logs/base-failures.txt)"
		if [ -s "$logs/new-failures.txt" ]; then
			echo "new failures from the merge:"
			sed 's/^/  /' "$logs/new-failures.txt"
			status=1
		fi
	fi
fi
if [ "$oracle" = test262 ] && [ "${#filters[@]}" -gt 0 ]; then
	if [ -z "${ADAMIC_TEST262:-}" ]; then
		echo "test262 filters named but ADAMIC_TEST262 is unset"
		status=1
	elif go run ./cmd/adamic-test262 -test262 "$ADAMIC_TEST262" -json "${filters[@]}" >"$logs/test262.json" 2>"$logs/test262.log"; then
		python3 -c 'import json,sys
document = json.load(open(sys.argv[1]))
bad = [f for f in document["filters"] if f["fail"] != 0]
for f in document["filters"]:
    print("test262 %s: %d pass, %d fail, %d refused, %d crashed, %d skipped" % (f["path"], f["pass"], f["fail"], f["refused"], f["crashed"], f["skipped"]))
sys.exit(1 if bad else 0)' "$logs/test262.json" || status=1
	else
		echo "test262 runner failed: $logs/test262.log"
		status=1
	fi
fi

if [ "$status" -ne 0 ]; then
	echo "not merged: area/$area stays ${areaTip:0:8}; logs in $logs"
	exit 1
fi
if [ "$push" = no ]; then
	echo "green, not pushed (--no-push): $branch ${sha:0:8} on area/$area ${areaTip:0:8} is ${merged}; logs in $logs"
	exit 0
fi
if git push -q origin "${merged}:refs/heads/area/$area"; then
	echo "merged: area/$area ${areaTip:0:8}..${merged:0:8} takes $branch ${sha:0:8}; logs in $logs"
else
	echo "area/$area moved while this ran; run again"
	exit 1
fi
