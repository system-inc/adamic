#!/usr/bin/env bash
# cloud/agreement-daily/job.sh <k/n> runs one shard of the admission delta's agreement mode (cmd/adamic-admission-delta,
# "Agreement mode") as a sealed Loom job, so the run that judges every program head accepts runs every day and
# diff.py can hold it against the day before (#wfnmhe9). Four jobs make the day, shards 1/4 to 4/4:
#
#   source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && bash cloud/agreement-daily/job.sh 2/4
#
# It prints, in order: the tool's stderr lines that aren't progress (its timeout line, every disagreeing, pending,
# compiler failure and stale run-time stop line, and its summary); one line
# `AGREEMENT_DAILY shard=<k/n> head=<sha> exit=<the tool's exit>`; then the compact report, diff.py compact's JSON
# gzipped and base64'd on one line, between `AGREEMENT_REPORT_BEGIN` and `AGREEMENT_REPORT_END`, which
# `diff.py extract` reads back out of the job's log.
#
# Exit: the tool's own, 0 for a pass or a partial shard, 1 for red, 3 for void. Before the tool, 64 is a shard that
# isn't k/n, 65 a preamble step that failed (the tree's commit, its submodules as repositories, npm) and 66 the tool's
# build or the manifest; after it, 67 is a report.json that didn't parse, so no report is printed.
set -u -o pipefail

shard=${1:-}
if ! [[ $shard =~ ^([1-9][0-9]{0,5})/([1-9][0-9]{0,5})$ ]] || ((10#${BASH_REMATCH[1]} > 10#${BASH_REMATCH[2]})); then
	echo "agreement-daily: the shard is '$shard', not k/n with 1 <= k <= n, so nothing ran"
	exit 64
fi

# The preamble's functions live in the shell that sourced it, not in this one, so they're read again here when absent.
# Sourcing it twice only remakes what it made.
if ! declare -F sandboxed_repositories > /dev/null || ! declare -F sandboxed_npm > /dev/null; then
	. cloud/sandboxed-job.sh || { echo "agreement-daily: cloud/sandboxed-job.sh didn't source, so nothing ran"; exit 65; }
fi
: "${TMPDIR:?the preamble names TMPDIR}"
head=$(git rev-parse --verify 'HEAD^{commit}') || { echo "agreement-daily: $PWD holds no commit, so nothing ran"; exit 65; }
# The tool runs `git -C cohere worktree add` at the pin, so every submodule is a repository before anything else
sandboxed_repositories || { echo "agreement-daily: sandboxed_repositories failed, so nothing ran"; exit 65; }
sandboxed_npm stage3/api || { echo "agreement-daily: sandboxed_npm stage3/api failed, so nothing ran"; exit 65; }
go build -o "$TMPDIR/agreement-tool" ./cmd/adamic-admission-delta ||
	{ echo "agreement-daily: go build ./cmd/adamic-admission-delta failed, so nothing ran"; exit 66; }
python3 -I cloud/admission-corpus/manifest.py --sha "$head" > "$TMPDIR/manifest.json" ||
	{ echo "agreement-daily: cloud/admission-corpus/manifest.py failed at $head, so nothing ran"; exit 66; }

# Half the box: a shard run on every thread timed out under its own load (Nov 21)
threads=$(nproc) || threads=2
workers=$((threads / 2))
[ "$workers" -ge 1 ] || workers=1
echo "agreement-daily: shard $shard at $head on $workers workers of $threads threads"

# Plain files only. A `2> >(tee ...)` once wrote stderr lines into the JSON.
toolExit=0
"$TMPDIR/agreement-tool" --agreement --head "$head" --manifest "$TMPDIR/manifest.json" --shard "$shard" --workers "$workers" \
	--json > "$TMPDIR/report.json" 2> "$TMPDIR/report.err" || toolExit=$?

grep -Ev '^agreement: [0-9]+ s: ' "$TMPDIR/report.err"
echo "AGREEMENT_DAILY shard=$shard head=$head exit=$toolExit"
python3 -I cloud/agreement-daily/diff.py compact "$TMPDIR/report.json" > "$TMPDIR/compact.json" ||
	{ echo "agreement-daily: the tool's report.json didn't parse, so there is no report"; exit 67; }
report=$(gzip -nc < "$TMPDIR/compact.json" | base64 -w0) && [ -n "$report" ] ||
	{ echo "agreement-daily: the compact report didn't gzip and base64, so there is no report"; exit 67; }
echo AGREEMENT_REPORT_BEGIN
echo "$report"
echo AGREEMENT_REPORT_END
exit "$toolExit"
