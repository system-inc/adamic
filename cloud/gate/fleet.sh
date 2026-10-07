#!/usr/bin/env bash
# Runs Adamic's whole gate, uncached, as shards on separate cloud boxes, then merges them on one more
# box and prints the verdict. The shard runner is cmd/adamic-gate (docs/gate-shards.md); this script
# only starts the boxes, waits for them and reads back what they pushed. Every box pushes its raw
# logs alone on an orphan branch, gate-logs/<sha12>/<run>/<name>, so nothing it does can touch main,
# and a run never reads an older run's logs for the same sha.
#
#   cloud/gate/fleet.sh run <full sha> [count]          start count shard boxes (default 8), wait, merge
#   cloud/gate/fleet.sh merge <full sha> <run> [count]  merge a run's pushed logs again
#   cloud/gate/fleet.sh status <full sha> [run]         which logs have arrived
#
# Boxes are started by ADAMIC_GATE_LAUNCH, called as: $ADAMIC_GATE_LAUNCH <label> <brief file>.
# The default starts a Codex cloud task in the adamic environment through `ahra ai` from
# ~/Projects/ahra. A launcher must give the box a clone of this repository with push access.
# Green means: every shard's logs arrived, `adamic-gate merge` is green on them (every planned unit
# has one complete pass, no cache hit, no duplicate, and no required WASI unit skipped), and the
# merge box prints its totals. The merge report lands on gate-logs/<sha12>/<run>/merge, and its
# commit message, which this script prints, starts with GATE GREEN <sha> or GATE RED <sha>.
set -euo pipefail

usage() {
	echo "usage: $0 run <full sha> [count] | merge <full sha> <run> [count] | status <full sha> [run] | briefs <full sha> <plan.json>" >&2
	exit 2
}
[ "$#" -ge 2 ] || usage
verb=$1
sha=$2
case "$verb" in
run)
	run=$(date -u +%Y%m%dT%H%M%S)
	count=${3:-8}
	;;
briefs)
	[ "$#" -eq 3 ] || usage
	run=briefs
	planFile=$3
	count=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["Count"])' "$planFile")
	;;
merge)
	[ "$#" -ge 3 ] || usage
	run=$3
	count=${4:-8}
	;;
*)
	run=${3:-}
	count=8
	;;
esac
case "$sha" in
*[!0-9a-f]* | "") usage ;;
esac
[ "${#sha}" = 40 ] || { echo "fleet: give the full 40-character sha, not $sha" >&2 && exit 2; }
short=${sha:0:12}
prefix=gate-logs/$short${run:+/$run}
work=${ADAMIC_GATE_FLEET_DIRECTORY:-$HOME/.adamic-gate/fleet}/$short${run:+/$run}
mkdir -p "$work"
fleet=gate-${sha:0:8}

launch() {
	if [ -n "${ADAMIC_GATE_LAUNCH:-}" ]; then
		$ADAMIC_GATE_LAUNCH "$1" "$2"
	else
		(cd "$HOME/Projects/ahra" && ahra ai start codex --directory "$HOME/Projects/system/adamic" --environment adamic --label "$1" --fleet "$fleet" --prompt-file "$2")
	fi
}

arrived() {
	git fetch -q origin "+refs/heads/$prefix/*:refs/remotes/origin/$prefix/*" 2> /dev/null || true
	git for-each-ref --format='%(refname)' "refs/remotes/origin/$prefix/" | sed "s#^refs/remotes/origin/$prefix/##"
}

# Wait for every named log branch, up to the deadline: a box that never pushes must not hang this.
await() {
	local deadline=$(($(date +%s) + $1))
	shift
	while :; do
		local missing=()
		local present
		present=$(arrived)
		for name in "$@"; do
			grep -qx "$name" <<< "$present" || missing+=("$name")
		done
		[ "${#missing[@]}" = 0 ] && return 0
		if [ "$(date +%s)" -ge "$deadline" ]; then
			echo "fleet: no logs after the deadline from: ${missing[*]}" >&2
			echo "fleet: rerun one with its brief in $work, or start again" >&2
			return 1
		fi
		sleep "${ADAMIC_GATE_FLEET_POLL:-30}"
	done
}

# One brief per shard. The last shard carries any required WASI units (Plan.WASI), so every box
# provisions the WASI SDK when this tree's setup offers it; the runner refuses to start a shard whose
# requirements are missing, and merge refuses any skipped required unit. Every box also provides the
# gate's required inputs (--gate-inputs: the pinned TypeScript source, reference libraries and
# oracles) whenever this tree's setup offers them.
brief() {
	local name=$1 command=$2 index=${3:--1}
	local archiveStep=""
	if [ "$index" = "$archiveShard" ]; then
		archiveStep=' This shard owns TestSplitTSGoAgrees: copy the sourced env.sh to /workspace/gate-inputs-env.sh, run `bash cloud/setup.sh --gate-archive > /workspace/archive-setup.log 2>&1`, source its printed env.sh, then `gateArchivePath=$ADAMIC_CLANG_TSGO_ARCHIVE; source /workspace/gate-inputs-env.sh; export ADAMIC_CLANG_TSGO_ARCHIVE="$gateArchivePath"`. The setup flags are deliberately separate invocations because their combination is rejected; restoring the first env preserves the other required inputs.'
	fi
	cat << BRIEF
Unit: run one part of Adamic's test gate at a fixed commit and return the raw logs. Branch for the logs: $prefix/$name. This is a measured run; do not change any code.

1. In the repository: \`git fetch origin && git checkout --detach $sha\` and confirm \`git rev-parse HEAD\` prints $sha. Fetch the frozen plan: \`git fetch origin refs/heads/$prefix/plan && git show FETCH_HEAD:plan.tgz > /workspace/gate-plan.tgz && mkdir -p /workspace/gate-plan && tar xzf /workspace/gate-plan.tgz -C /workspace/gate-plan\`.
2. Set up with every gate input this tree's setup offers: \`flags=""; grep -q -- --wasi-sdk cloud/setup.sh && flags="\$flags --wasi-sdk"; grep -q -- --gate-inputs-no-archive cloud/setup.sh && flags="\$flags --gate-inputs-no-archive"; bash cloud/setup.sh \$flags\`. Source the env file it prints, then \`export ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1\`. Record setup's timing lines, the flags it ran with, and \`nproc\`. A test that skips for a missing input is a gate failure, not a pass.$archiveStep
3. \`go build -o /workspace/adamic-gate ./cmd/adamic-gate\`, read docs/gate-shards.md, then run, with output to a log file (never piped):
   \`date -u; $command > /workspace/gate-run.log 2>&1; echo exit=\$?; date -u\`
   If the box restarts or the command is interrupted, rerun the same command with \`-resume\` added (same output directory) until it completes. Record each start, end and interruption.
4. Only if the shard command actually ran (\`/workspace/gate-out/summary.json\` exists) save and push its logs. If setup or the box failed before the shard ran (no network to a package server, a dead tool), do not push anything: say exactly what failed and stop, and the driver will start this shard again on a fresh box. Save the results without touching the repository's working tree: \`tar czf /workspace/$name.tgz -C /workspace gate-out gate-run.log gate-plan\`, then from a fresh scratch clone so nothing else is committed:
   \`git clone --no-checkout --depth 1 "\$(git remote get-url origin)" /workspace/logs-repo && cd /workspace/logs-repo && git checkout --orphan $prefix/$name && cp /workspace/$name.tgz . && git add $name.tgz && git commit -m "Gate logs for $name at $short" && git push origin $prefix/$name\`
   Retry the push on failure. Never push main or any other branch.
5. Reply with a five-line summary first: exit code, pass/fail/skip counts from the summary the command wrote, wall time of the shard command and of setup, interruptions and resumes, and the pushed branch tip. Then every failing test with its first error line.
BRIEF
}

mergeBrief() {
	local shards=$1
	cat << BRIEF
Unit: merge Adamic's gate shard logs at a fixed commit and report the verdict. No code changes. Results go on branch $prefix/merge.

1. \`git fetch origin && git checkout --detach $sha\`, confirm HEAD. \`bash cloud/setup.sh --gate-inputs --wasi-sdk\`, source the env file it prints and export ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1. \`go build -o /workspace/adamic-gate ./cmd/adamic-gate\`. Fetch/extract $prefix/plan to /workspace/gate-plan as the shard briefs do. Read docs/gate-shards.md.
2. For each of $shards: \`git fetch origin $prefix/<name>\` and extract <name>.tgz from that branch's tip into /workspace/logs/<name>/ (it holds gate-out/ and gate-run.log).
3. \`/workspace/adamic-gate merge -plan /workspace/gate-plan/plan.json -out /workspace/merged $(for name in $shards; do printf '/workspace/logs/%s/gate-out ' "$name"; done)> /workspace/merge.log 2>&1; echo exit=\$?\`
4. Save merge.log and merged/merged.json in merge.tgz and push it alone on an orphan branch $prefix/merge from a fresh scratch clone (\`git clone --no-checkout --depth 1\`, \`git checkout --orphan\`, add only merge.tgz, push). The commit message's first line is exactly \`GATE GREEN $sha\` or \`GATE RED $sha\`, as merge decided, and its body is the totals line and every failure or refusal merge printed, one per line. Never push anything else.
5. Reply, first line exactly \`GATE GREEN $sha\` or \`GATE RED $sha\` as merge decided, then: pass, fail, skip, distinct and raw terminal counts; every failure with its shard and test; every refusal merge printed; per-shard wall times and the slowest; the pushed tip.
BRIEF
}

shards() {
	local names=""
	for index in $(seq 0 $((count - 1))); do names="$names shard-$index"; done
	echo "$names"
}

# Read the archive owner from the plan; do not duplicate the packing algorithm in Bash.
readPlan() {
	archiveShard=$(python3 - "$planFile" "$sha" "$count" <<'PYPLAN'
import json,sys
p=json.load(open(sys.argv[1]))
if p['Commit'] != sys.argv[2] or p['Count'] != int(sys.argv[3]):
    raise SystemExit('fleet: plan commit/count differs from requested run')
unit='github.com/system-inc/adamic/internal/native::TestSplitTSGoAgrees'
owners=[u['Shard'] for u in p['Units'] if u['Package']+'::'+u['Test']==unit]
a=p.get('Archive')
if owners:
    if len(owners)!=1 or not a or a['Unit']!=unit or a['Variable']!='ADAMIC_CLANG_TSGO_ARCHIVE' or a['Shard']!=owners[0]:
        raise SystemExit('fleet: TestSplitTSGoAgrees archive declaration differs')
    print(owners[0])
else:
    if a: raise SystemExit('fleet: unexpected archive declaration')
    print(-1)
PYPLAN
)
}

planBrief() {
	cat << BRIEF
Unit: compute Adamic's shard plan at $sha. No code changes. Publish only gate-logs/$short/$run/plan.
1. Fetch origin and checkout --detach $sha; confirm HEAD.
2. Run bash cloud/setup.sh --gate-inputs --wasi-sdk > /workspace/setup.log 2>&1, source its printed env.sh, and export ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1. The plan box provisions all inputs, including the archive, to freeze their byte identities. Record setup separately.
3. Build /workspace/adamic-gate from ./cmd/adamic-gate. Run /workspace/adamic-gate plan -count $count > /workspace/plan.json 2> /workspace/plan.log. Refuse to publish if it fails.
4. Archive plan.json, plan.log and setup.log as plan.tgz. From a fresh scratch clone, checkout --orphan $prefix/plan, add only plan.tgz, commit and push $prefix/plan. Never push main or any other branch.
5. Report the plan digest, archive owner shard, setup and planning durations, and log branch tip.
BRIEF
}

preparePlan() {
	planBrief > "$work/plan.md"
	launch "$fleet-plan" "$work/plan.md"
	await "${ADAMIC_GATE_PLAN_WAIT:-3600}" plan
	git show "origin/$prefix/plan:plan.tgz" > "$work/plan.tgz"
	tar xzf "$work/plan.tgz" -C "$work"
	planFile=$work/plan.json
	readPlan
}

runMerge() {
	local names
	names=$(shards)
	mergeBrief "$names" > "$work/merge.md"
	launch "$fleet-merge" "$work/merge.md"
	await "${ADAMIC_GATE_FLEET_MERGE_WAIT:-3600}" merge || exit 1
	git log -1 --format=%B "refs/remotes/origin/$prefix/merge"
	echo "fleet: merge report on $prefix/merge"
	git log -1 --format=%s "refs/remotes/origin/$prefix/merge" | grep -qx "GATE GREEN $sha"
}

case "$verb" in
run)
	git cat-file -e "$sha^{commit}" 2> /dev/null || git fetch -q origin
	git cat-file -e "$sha^{commit}" || { echo "fleet: $sha is not a commit on origin" >&2 && exit 1; }
	started=$(date +%s)
	preparePlan
	for index in $(seq 0 $((count - 1))); do
		brief "shard-$index" "/workspace/adamic-gate shard -plan /workspace/gate-plan/plan.json -index $index -count $count -out /workspace/gate-out" "$index" > "$work/shard-$index.md"
		launch "$fleet-shard-$index" "$work/shard-$index.md"
	done
	echo "fleet: run $run, logs on $prefix/"

	# A shard takes about 20 minutes; two hours covers setup, restarts and resumes. A box can fail
	# in ways its brief can't fix (no push credentials, a dead container), so each shard whose logs
	# never arrive is started once more on a fresh box before the run gives up.
	# shellcheck disable=SC2046
	if ! await "${ADAMIC_GATE_FLEET_SHARD_WAIT:-7200}" $(shards); then
		present=$(arrived)
		retried=()
		for name in $(shards); do
			grep -qx "$name" <<< "$present" && continue
			launch "$fleet-$name-retry" "$work/$name.md"
			retried+=("$name")
		done
		echo "fleet: started again on fresh boxes: ${retried[*]}"
		await "${ADAMIC_GATE_FLEET_SHARD_WAIT:-7200}" "${retried[@]}" || exit 1
	fi
	echo "fleet: all $count shard logs arrived after $((($(date +%s) - started) / 60)) minutes"
	runMerge
	;;
briefs)
	readPlan
	for index in $(seq 0 $((count - 1))); do
		brief "shard-$index" "/workspace/adamic-gate shard -plan /workspace/gate-plan/plan.json -index $index -count $count -out /workspace/gate-out" "$index" > "$work/shard-$index.md"
	done
	echo "fleet: $count briefs, archive shard $archiveShard, in $work"
	;;
merge) runMerge ;;
status) arrived ;;
*) usage ;;
esac
