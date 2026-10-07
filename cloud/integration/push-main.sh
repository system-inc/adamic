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
# --meter-run <commit> also records a stage 3 meter run (stage3/meter/twice-daily.sh, run on a box
# against main) in a commit of its own after the velocity row. Only the files that commit adds under
# stage3/meter/runs/ are taken, and the script refuses to push if the record changes anything else.
# A meter run is measured after a landing, so each landing carries the newest run there is.
#
# usage: cloud/integration/push-main.sh [--defer-velocity] [--meter-run <commit>] <full sha> <gate minutes> <pass> <fail> <skip> "<branches landed>"
set -euo pipefail

defer=no
meterRun=""
while [ "$#" -gt 0 ]; do
	case "$1" in
	--defer-velocity) defer=yes; shift ;;
	--meter-run) meterRun=$2; shift 2 ;;
	*) break ;;
	esac
done
if [ "$defer" = yes ] && [ -n "$meterRun" ]; then
	echo "refused: a meter run is recorded with the velocity row, so it can't ride a deferred push" >&2
	exit 2
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
record=$velocity
meterNote=""
if [ -n "$meterRun" ]; then
	git fetch -q origin "$meterRun" 2>/dev/null || true
	meterCommit=$(git rev-parse --verify "${meterRun}^{commit}")
	# The runs this commit has that the landing doesn't, so a run main already holds is never recorded twice.
	added=$(git diff --name-only --diff-filter=A "$velocity" "$meterCommit" -- stage3/meter/runs/)
	if [ -z "$added" ]; then
		echo "refused to record the meter run: ${meterCommit:0:8} has nothing under stage3/meter/runs/ that main lacks" >&2
		exit 1
	fi
	index=$(mktemp)
	GIT_INDEX_FILE=$index git read-tree "$velocity"
	while IFS= read -r file; do
		entry=$(git ls-tree "$meterCommit" -- "$file")
		GIT_INDEX_FILE=$index git update-index --add --cacheinfo "$(printf '%s' "$entry" | awk '{print $1}'),$(printf '%s' "$entry" | awk '{print $3}'),${file}"
	done <<<"$added"
	tree=$(GIT_INDEX_FILE=$index git write-tree)
	rm -f "${index:?}"
	runs=$(printf '%s\n' "$added" | awk -F/ '{print $4}' | sort -u | paste -sd ' ' -)
	meter=$(git commit-tree "$tree" -p "$velocity" -m "Record the stage 3 meter run ${runs} from ${meterCommit:0:8}")
	outside=$(git diff --name-only "$velocity" "$meter" | grep -v '^stage3/meter/runs/' || true)
	if [ -n "$outside" ]; then
		echo "refused to push the meter record: it changes $outside" >&2
		exit 1
	fi
	record=$meter
	meterNote="; meter run ${runs} in ${meter:0:8}"
fi
git push origin "${record}:refs/heads/main"
rm -f "${heldRows:?}"

echo "Pushed main ${old:0:8}..${sha:0:8}, ${commitsLanded} commits (${branches}), gate ${pass} pass / ${fail} fail / ${skip} skip, ${gateMinutes} minutes uncached; velocity row ${velocity:0:8}${meterNote}; backlog ${backlog} commits."
"$directory/merge-back.sh"
