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
# stage3/meter/runs/ are taken, plus stage3/progress.json (stage 3's milestone record) while nothing in
# the landing reads that file, so it can't change what a gate tested; the script refuses to push if
# the record changes anything else. A meter run is measured after a landing, so each landing carries
# the newest run there is.
#
# Gate minutes are wall time from the first gate start to the last green piece (the whole run, its
# reruns, input runs and groups), so every row means the same thing (@system_adamic, October 7).
# --correct-minutes <new_main> <minutes> "<note>" rewrites an earlier row that was recorded another
# way, in this landing's velocity commit: the row whose new_main starts with <new_main> gets <minutes>
# and "; <note>" on its branches field. Exactly one row must match.
#
# usage: cloud/integration/push-main.sh [--defer-velocity] [--meter-run <commit>] [--correct-minutes <new_main> <minutes> "<note>"] <full sha> <gate minutes> <pass> <fail> <skip> "<branches landed>"
set -euo pipefail

defer=no
meterRun=""
correctMain=""
correctMinutes=""
correctNote=""
while [ "$#" -gt 0 ]; do
	case "$1" in
	--defer-velocity) defer=yes; shift ;;
	--meter-run) meterRun=$2; shift 2 ;;
	--correct-minutes) correctMain=$2; correctMinutes=$3; correctNote=$4; shift 4 ;;
	*) break ;;
	esac
done
if [ "$defer" = yes ] && [ -n "$meterRun" ]; then
	echo "refused: a meter run is recorded with the velocity row, so it can't ride a deferred push" >&2
	exit 2
fi
if [ -n "$correctMain" ]; then
	if [ "$defer" = yes ]; then
		echo "refused: a corrected row is written with the velocity commit, so it can't ride a deferred push" >&2
		exit 2
	fi
	if ! [[ "$correctMain" =~ ^[0-9a-f]{8,40}$ ]] || ! [[ "$correctMinutes" =~ ^[0-9]+$ ]]; then
		echo "refused: --correct-minutes takes a main sha of 8 or more hex characters and whole minutes" >&2
		exit 2
	fi
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
	# Ruled by @system_adamic, October 7: a green stack lands over a main that moved only by record
	# commits (the velocity table, meter runs, and a progress.json nothing in the landing reads). The
	# landing is a merge whose tree is the gated tree plus exactly those record paths from main, so
	# what reaches main is what the gate tested; any other commit on main means merge and gate again.
	gated=$sha
	# git computes the merge itself (merge-tree, no worktree), so a stack that already holds main's
	# older commits, or two merge bases, is judged by the tree it would really produce.
	if ! tree=$(git merge-tree --write-tree "$old" "$gated" | head -n 1) || [ -z "$tree" ]; then
		echo "refused: ${gated:0:8} doesn't merge cleanly with main ${old:0:8}; merge it in and gate again" >&2
		exit 1
	fi
	recordPaths=$(git diff --name-only "$gated" "$tree")
	notRecords=$(printf '%s\n' "$recordPaths" | grep -v -e '^documentation/velocity/landings\.csv$' -e '^stage3/meter/runs/' -e '^stage3/progress\.json$' | grep . || true)
	if [ -n "$notRecords" ]; then
		echo "refused: main moved to ${old:0:8} under this gate, beyond record commits ($(printf '%s' "$notRecords" | head -n 3 | paste -sd ' ' -)); merge it in and gate again" >&2
		exit 1
	fi
	if printf '%s\n' "$recordPaths" | grep -qx 'stage3/progress.json' && git grep -q 'progress\.json' "$gated" -- '*.go' '*.py' '*.sh' '*.mjs' '*.cjs' '*.js' '*.ts' '*.a'; then
		echo "refused: main's stage3/progress.json differs from ${gated:0:8}'s and ${gated:0:8} reads it; merge it in and gate again" >&2
		exit 1
	fi
	sha=$(git commit-tree "$tree" -p "$old" -p "$gated" -m "Land ${gated:0:8} over main ${old:0:8}, which moved only by record commits

The tree is ${gated:0:8}'s, as gated, plus main's record paths: $(printf '%s\n' "$recordPaths" | grep . | sed 's#^stage3/meter/runs/\([^/]*\)/.*#stage3/meter/runs/\1/#' | sort -u | paste -sd ' ' -).")
	echo "Landing ${gated:0:8} over record-only main ${old:0:8} as ${sha:0:8} (tree = gated tree + record paths)."
fi
if [ -n "$correctMain" ]; then
	matched=$(git show "${sha}:${velocityFile}" 2>/dev/null | awk -F, -v main="$correctMain" 'index($3, main) == 1' | wc -l | tr -d ' ')
	if [ "$matched" != "1" ]; then
		echo "refused: ${matched} velocity rows have new_main ${correctMain}, not one" >&2
		exit 1
	fi
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
	if [ -n "$correctMain" ]; then
		matched=$(printf '%s\n' "$existing" | awk -F, -v main="$correctMain" 'index($3, main) == 1' | wc -l | tr -d ' ')
		if [ "$matched" != "1" ]; then
			echo "main is pushed as ${sha:0:8}, but its velocity row is NOT written: ${matched} rows have new_main ${correctMain}, not one; write the row by hand" >&2
			exit 1
		fi
		existing=$(printf '%s\n' "$existing" | awk -F, -v OFS=, -v main="$correctMain" -v minutes="$correctMinutes" -v note="$(printf '%s' "$correctNote" | tr ',' ';')" 'index($3, main) == 1 { $5 = $5 "; " note; $6 = minutes } { print }')
	fi
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
	added=$(git diff --name-only --diff-filter=AM "$velocity" "$meterCommit" -- stage3/meter/runs/ stage3/progress.json)
	if printf '%s\n' "$added" | grep -qx 'stage3/progress.json' && git grep -q 'progress\.json' "$sha" -- '*.go' '*.py' '*.sh' '*.mjs' '*.cjs' '*.js' '*.ts' '*.a'; then
		echo "refused to record stage3/progress.json as data: something in ${sha:0:8} reads it, so it has to be gated" >&2
		exit 1
	fi
	if [ -z "$added" ]; then
		echo "refused to record the meter run: ${meterCommit:0:8} has nothing under stage3/meter/runs/ or stage3/progress.json that main lacks" >&2
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
	runs=$(printf '%s\n' "$added" | awk -F/ '$2 == "meter" { print $4 } $2 == "progress.json" { print "and progress.json" }' | sort -u | paste -sd ' ' -)
	meter=$(git commit-tree "$tree" -p "$velocity" -m "Record the stage 3 meter run ${runs} from ${meterCommit:0:8}")
	outside=$(git diff --name-only "$velocity" "$meter" | grep -v -e '^stage3/meter/runs/' -e '^stage3/progress\.json$' || true)
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
