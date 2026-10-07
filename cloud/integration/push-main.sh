#!/usr/bin/env bash
# Moves main to a sha whose whole gate ran green and uncached, then records the landing and merges
# main back into every area. Only integration runs it. It checks what it can (a full sha, main still
# its ancestor, no failures) and refuses otherwise; the gate itself is the caller's evidence, passed
# as numbers, and must have run with ADAMIC_GATE_UNCACHED=1 and "gate cache: ... hits=0".
#
# The landing's row goes to documentation/velocity/landings.csv in its own commit on top of the green
# sha. That commit changes no other file, which the script checks before it pushes, so the gated tree
# and the pushed tree differ by that one row (ruled by @system_adamic, October 6).
#
# --defer-velocity holds this landing's row in a local file instead of committing it now, so a
# speculative stack built on this sha stays a fast-forward; the next push writes every held row
# in its own velocity commit.
#
# usage: cloud/integration/push-main.sh [--defer-velocity] <full sha> <gate minutes> <pass> <fail> <skip> "<branches landed>"
set -euo pipefail

defer=no
if [ "${1:-}" = "--defer-velocity" ]; then
	defer=yes
	shift
fi
if [ "$#" -ne 6 ]; then
	echo "usage: $0 <full sha> <gate minutes> <pass> <fail> <skip> \"<branches landed>\"" >&2
	exit 2
fi
sha=$1
gateMinutes=$2
pass=$3
fail=$4
skip=$5
branches=$6
velocityFile=documentation/velocity/landings.csv
velocityHeader=pushed_at_utc,old_main,new_main,commits_landed,branches_landed,gate_minutes,pass,fail,skip,backlog_commits
directory=$(cd "$(dirname "$0")" && pwd)
heldRows="$(git rev-parse --git-common-dir)/velocity-held-rows.csv"

if ! [[ "$sha" =~ ^[0-9a-f]{40}$ ]]; then
	echo "refused: pass the full 40-character sha, not $sha" >&2
	exit 1
fi
if [ "$fail" != "0" ]; then
	echo "refused: the gate has $fail failures" >&2
	exit 1
fi

git fetch -q origin
old=$(git rev-parse origin/main)
if [ "$old" = "$sha" ]; then
	echo "refused: main is already $sha" >&2
	exit 1
fi
if ! git merge-base --is-ancestor "$old" "$sha"; then
	echo "refused: main moved to ${old:0:8} under this gate; merge it in and gate again" >&2
	exit 1
fi
git push origin "${sha}:refs/heads/main"
commitsLanded=$(git rev-list --count "${old}..${sha}")

# The row. Branch names are joined with semicolons so the field needs no quoting.
backlog=$(git rev-list --count --no-merges --remotes=origin "^${sha}")
row="$(date -u +%Y-%m-%dT%H:%M:%SZ),${old},${sha},${commitsLanded},$(printf '%s' "$branches" | tr ',' ';'),${gateMinutes},${pass},${fail},${skip},${backlog}"
if [ "$defer" = yes ]; then
	printf '%s\n' "$row" >>"$heldRows"
	echo "Pushed main ${old:0:8}..${sha:0:8}, ${commitsLanded} commits (${branches}), gate ${pass} pass / ${fail} fail / ${skip} skip, ${gateMinutes} minutes uncached; velocity row held for the next push; backlog ${backlog} commits."
	"$directory/merge-back.sh"
	exit 0
fi
rows=$row
if [ -s "$heldRows" ]; then
	rows="$(cat "$heldRows")
$row"
fi
if existing=$(git show "${sha}:${velocityFile}" 2>/dev/null); then
	blob=$(printf '%s\n%s\n' "$existing" "$rows" | git hash-object -w --stdin)
else
	blob=$(printf '%s\n%s\n' "$velocityHeader" "$rows" | git hash-object -w --stdin)
fi
index=$(mktemp)
GIT_INDEX_FILE=$index git read-tree "$sha"
GIT_INDEX_FILE=$index git update-index --add --cacheinfo "100644,${blob},${velocityFile}"
tree=$(GIT_INDEX_FILE=$index git write-tree)
rm -f "${index:?}"
velocity=$(git commit-tree "$tree" -p "$sha" -m "Record the landing ${old:0:8}..${sha:0:8} in the velocity table")
changed=$(git diff --name-only "$sha" "$velocity")
if [ "$changed" != "$velocityFile" ]; then
	echo "refused to push the velocity commit: it changes $changed" >&2
	exit 1
fi
git push origin "${velocity}:refs/heads/main"
rm -f "${heldRows:?}"

echo "Pushed main ${old:0:8}..${sha:0:8}, ${commitsLanded} commits (${branches}), gate ${pass} pass / ${fail} fail / ${skip} skip, ${gateMinutes} minutes uncached; velocity row ${velocity:0:8}; backlog ${backlog} commits."
"$directory/merge-back.sh"
