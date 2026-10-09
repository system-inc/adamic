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
# The fast gate (Kirk, October 7: "we need tests to run in <1 minute"): --fast-gate <gate-logs ref>
# lands on the fast gate developer tools runs on the Threadripper (build and vet everywhere, the
# touched packages' tests, a pinned oracle smoke set, the census), read from its own fast.json on that
# gate-logs branch rather than from numbers passed here. The script refuses unless that log gated this
# exact sha, from a base this sha holds, green, with no failure, no build or vet error and no
# unclassified or required-input skip. The landing note names what it ran: the packages, the smoke
# list's blob, the wall time and the machine. The whole uncached gate runs after landing, on every
# new main.
#
# The pause rule: if the newest finished whole gate on main (gate-logs/<sha12>/<stamp>/full-main) is
# red, every landing is refused until a later main's whole gate is green or that red is explained in
# main-reds.tsv beside this script, except a revert (--revert) or a fix-forward naming that red log
# (--fix-forward <gate-logs ref>). Main never carries two unexplained reds.
#
# usage: cloud/integration/push-main.sh [--defer-velocity] [--meter-run <commit>] [--correct-minutes <new_main> <minutes> "<note>"] [--revert | --fix-forward <red log ref>] <full sha> <gate minutes> <pass> <fail> <skip> "<branches landed>"
#        cloud/integration/push-main.sh [same options] --fast-gate <gate-logs ref> [--smoke-list-reviewed] <full sha> "<branches landed>"
set -euo pipefail

defer=no
meterRun=""
correctMain=""
correctMinutes=""
correctNote=""
fastGate=""
gateKind=fast
smokeReviewed=no
pauseException=""
testOnly=no
while [ "$#" -gt 0 ]; do
	case "$1" in
	--defer-velocity) defer=yes; shift ;;
	--meter-run) meterRun=$2; shift 2 ;;
	--correct-minutes) correctMain=$2; correctMinutes=$3; correctNote=$4; shift 4 ;;
	--fast-gate) fastGate=${2#origin/}; shift 2 ;;
	# --full-gate <gate-logs ref>/full-main of this exact sha: the whole gate as the landing's verdict
	# and main's confirmation in one (@system_adamic, October 8). It is a superset of the fast gate,
	# deferred tests included, so the landed tree is exactly the tree that passed everything.
	--full-gate) fastGate=${2#origin/}; gateKind=full; shift 2 ;;
	--smoke-list-reviewed) smokeReviewed=yes; shift ;;
	# --test-only <sha> "<branches>": a change that touches only tests goes to main with no gate in front of it;
	# Loom's next whole-suite run of main is its check (Kirk, Oct 8). The merged diff must be tests only.
	--test-only) testOnly=yes; shift ;;
	--revert) pauseException=revert; shift ;;
	--fix-forward) pauseException="fix-forward ${2#origin/}"; shift 2 ;;
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
if [ "$testOnly" = yes ]; then
	if [ "$#" -ne 2 ]; then
		echo "usage: $0 --test-only <full sha> \"<branches landed>\"" >&2
		exit 2
	fi
	sha=$1
	branches="$2; test-only lane, no gate (Kirk, Oct 8)"
	gateMinutes=0
	pass=0
	fail=0
	skip=0
elif [ -n "$fastGate" ]; then
	if [ "$#" -ne 2 ]; then
		echo "usage: $0 [options] --fast-gate <gate-logs ref> <full sha> \"<branches landed>\"" >&2
		exit 2
	fi
	sha=$1
	branches=$2
else
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
fi
velocityFile=documentation/velocity/landings.csv
velocityHeader=pushed_at_utc,old_main,new_main,commits_landed,branches_landed,gate_minutes,pass,fail,skip,backlog_commits
directory=$(cd "$(dirname "$0")" && pwd)
heldRows="$(git rev-parse --git-common-dir)/velocity-held-rows.csv"

if ! [[ "$sha" =~ ^[0-9a-f]{40}$ ]]; then
	echo "refused: pass the full 40-character sha, not $sha" >&2
	exit 1
fi

if [ -n "$fastGate" ]; then
	# The fast gate's verdict, read from its own log: what it gated, from which base, and what it ran.
	if ! git fetch -q origin "+refs/heads/${fastGate}:refs/remotes/origin/${fastGate}" 2>/dev/null; then
		echo "refused: no gate log ${fastGate} on origin" >&2
		exit 1
	fi
	statusLine=$(git show "origin/${fastGate}:status.txt" 2>/dev/null | head -n 1)
	# The log goes to python as a file: a landing that touches thousands of paths lists them all in
	# fast.json, past what one argument can carry (stage 3 batch 4, October 8).
	fastJSON=$(mktemp)
	trap 'rm -f "$fastJSON"' EXIT
	if ! git show "origin/${fastGate}:${gateKind}.json" >"$fastJSON" 2>/dev/null; then
		echo "refused: ${fastGate} has no ${gateKind}.json" >&2
		exit 1
	fi
	if ! verdict=$(GATE_KIND="$gateKind" python3 - "$sha" "$statusLine" "$fastGate" "$fastJSON" <<'VERDICT'
import json, os, subprocess, sys
sha, status, log = sys.argv[1], sys.argv[2], sys.argv[3]
with open(sys.argv[4]) as fastFile:
    fast = json.load(fastFile)
problems = []
if fast.get("sha") != sha:
    problems.append("it gated %s, not %s" % (fast.get("sha"), sha))
if not status.startswith("green"):
    problems.append("its status is %r" % status)
if fast.get("fail") != 0:
    problems.append("%s failures" % fast.get("fail"))
if not fast.get("build_ok") or not fast.get("vet_ok"):
    problems.append("go build or go vet failed")
for kind in ("unclassified_skips", "required_input_skips"):
    if fast.get(kind):
        problems.append("%s: %s" % (kind.replace("_", " "), " ".join(fast[kind])))
if fast.get("uncached_tests") is not True:
    problems.append("its tests weren't run uncached")
# A green line alone isn't the verdict: an exception inside a stage once left failure unset and
# published green (EMFILE under load, round 65's angel). Every planned stage must have run, and where
# the log records each stage's exit, every one must be 0.
kind = os.environ.get("GATE_KIND", "fast")
planned = ("build", "vet", "tests", "census", "stage3") if kind == "full" else ("build", "vet", "tests", "smoke", "census")
if kind == "full" and fast.get("packages") != "all":
    problems.append("a full gate must run every package, this one ran %r" % fast.get("packages"))
ran = fast.get("steps_seconds") or {}
missing = [stage for stage in planned if stage not in ran]
if missing:
    problems.append("stages with no completion recorded: %s" % " ".join(missing))
exits = fast.get("stages_exit") or fast.get("exit_codes") or {}
nonzero = ["%s=%s" % (stage, code) for stage, code in exits.items() if code != 0]
if nonzero:
    problems.append("stages that exited nonzero: %s" % " ".join(nonzero))
# Since developer tools' fail-closed fix (62a5fcf1) the log names its planned stages, and each must
# have an exit of 0 on record.
unrecorded = [stage for stage in fast.get("planned_stages") or [] if exits.get(stage) != 0]
if unrecorded:
    problems.append("planned stages without a recorded exit of 0: %s" % " ".join(unrecorded))
if fast.get("finished") is not True:
    problems.append("the gate didn't record that it finished")
# Loom's pool as the landing gate (@system_adamic, Oct 9 02:33Z): a pool record lands once the parent rules the phase
# units' parity proven (the switch ~/.adamic-full-gate/pool-promoted, the same file the whole-gate loops read). It must
# cover every stage a box's whole gate runs, the same checks above apply to it, and a fifth of shas, chosen by the sha
# so every reader agrees, also need a green box record of the same sha (the spot check).
if kind == "full" and fast.get("runner") == "pool":
    promoted = os.environ.get("PUSH_MAIN_POOL_PROMOTED", os.path.expanduser("~/.adamic-full-gate/pool-promoted"))
    if not os.path.exists(promoted):
        problems.append("it's a pool record, and the pool isn't promoted to land yet (%s)" % promoted)
    whole = {"coverage", "tools", "build", "vet", "tests", "wasi", "stage3", "catalog", "determinism", "census"}
    uncovered = sorted(whole - set(fast.get("planned_stages") or []))
    if uncovered or fast.get("covers") not in (None, "all"):
        problems.append("the pool record doesn't cover every stage of a whole gate (missing %s, covers %r)" % (" ".join(uncovered) or "none", fast.get("covers")))
    if int(sha[:8], 16) % 5 == 0:
        refs = subprocess.run(["git", "ls-remote", "origin", "refs/heads/gate-logs/%s/*" % sha[:12]], capture_output=True, text=True).stdout.split()
        boxGreen = False
        for ref in [ref for ref in refs if ref.endswith("/full-main")]:
            if subprocess.run(["git", "fetch", "-q", "origin", ref], capture_output=True).returncode != 0:
                continue
            record = subprocess.run(["git", "show", "FETCH_HEAD:full.json"], capture_output=True, text=True).stdout
            line = subprocess.run(["git", "show", "FETCH_HEAD:status.txt"], capture_output=True, text=True).stdout
            try:
                boxRecord = json.loads(record)
            except ValueError:
                continue
            if boxRecord.get("sha") == sha and boxRecord.get("runner", "box") != "pool" and boxRecord.get("finished") is True and line.startswith("green"):
                boxGreen = True
                break
        if not boxGreen:
            problems.append("its sha is one the boxes spot-check (a fifth, by the sha), and no green box whole gate of it is on origin yet")
# A scoped run (ADAMIC_LINT_RULES, cohere's rule-only batches) skips lint's corpus-wide tests by name,
# so it never lands anything (cohere, #60hxabf). Since developer tools' 796e9810 the gate refuses to
# start with it set and records "scoped_env"; a log from before that records nothing, and no gate
# before it ever set the variable.
if fast.get("scoped_env"):
    problems.append("it ran scoped (%s set)" % " ".join(fast["scoped_env"]))
# The fast gate defers the full gate's slow tests, and 04's split-build red reached main through
# that hole (@system_adamic, October 8): a landing that changes emitted C or the runtime, or that
# asks for them (a Gate-runs: deferred trailer), proves its deferred tests before it pushes.
import subprocess
message = subprocess.run(["git", "log", "-1", "--format=%B", sha], capture_output=True, text=True).stdout
changed = subprocess.run(["git", "diff", "--name-only", fast.get("base", sha), sha], capture_output=True, text=True).stdout.split()
# Tests and their fixtures don't change emitted C, so a test-only change there isn't covered.
covered = "Gate-runs: deferred" in message or any(
    path.startswith(("internal/native/", "internal/lower/", "internal/ir/", "internal/javascript/"))
    and not path.endswith("_test.go") and "/testdata/" not in path for path in changed)
if covered and kind != "full":
    results = fast.get("deferred_run_results") or {}
    if fast.get("deferred_all_requested") is not True:
        problems.append("it changes emitted C or the runtime but didn't run its deferred tests (put Gate-runs: deferred on the candidate)")
    unproven = ["%s=%s" % (test, verdict) for test, verdict in sorted(results.items()) if verdict != "pass"]
    if unproven:
        problems.append("deferred tests not passing: %s" % " ".join(unproven))
if problems:
    print("; ".join(problems))
    sys.exit(1)
seconds = float(fast["wall_seconds"])
packages = (fast.get("package_list") if kind == "full" else fast.get("packages")) or []
machine = fast.get("machine") or {}
# A full gate tested this exact tree whole; its base for the moved-paths check below is where it left main.
print(subprocess.run(["git", "merge-base", sha, "origin/main"], capture_output=True, text=True).stdout.strip() if kind == "full" else fast["base"])
print("%.2f" % (seconds / 60))
print(fast["pass"], fast["fail"], fast["skip"])
print(kind + " gate %s: %.0f s on %s (%s threads), tools %s, base %s, %d packages (%s), smoke list %s blob %s" % (
    log, seconds, machine.get("hostname", "?"), machine.get("nproc", "?"), str(fast.get("tools_sha", "?"))[:8],
    fast["base"][:8], len(packages), " ".join(packages), fast.get("smoke_list", "?"), str(fast.get("smoke_list_blob", "?"))[:8]))
VERDICT
	); then
		echo "refused: the fast gate ${fastGate} doesn't pass ${sha:0:8}: ${verdict}" >&2
		exit 1
	fi
	fastBase=$(printf '%s\n' "$verdict" | sed -n 1p)
	gateMinutes=$(printf '%s\n' "$verdict" | sed -n 2p)
	read -r pass fail skip <<<"$(printf '%s\n' "$verdict" | sed -n 3p)"
	fastNote=$(printf '%s\n' "$verdict" | sed -n 4p)
	# The fast gate chose its packages by the gated tree's difference from its base, a main. That
	# choice is this landing's only if every path main has changed since, beyond record commits, is in
	# that difference too: main moved by records only, or by commits this landing already holds (a
	# stack whose lower candidate landed first), which the gate then tested along with the rest. The
	# landing itself may be a record-only merge over main (below), whose tree differs from the gated
	# one only in record paths, which no package reads; that check refuses a stack that doesn't hold
	# main's other commits.
	git fetch -q origin main
	if ! git merge-base --is-ancestor "$fastBase" origin/main; then
		echo "refused: the fast gate diffed against ${fastBase:0:8}, which isn't on main's line" >&2
		exit 1
	fi
	movedSince=$(git diff --name-only "$fastBase" origin/main | grep -v -e '^documentation/velocity/landings\.csv$' -e '^stage3/meter/runs/' -e '^stage3/progress\.json$' | grep . || true)
	untested=$(comm -23 <(printf '%s\n' "$movedSince" | grep . | sort) <(git diff --name-only "$fastBase" "$sha" | sort) || true)
	if [ -n "$untested" ]; then
		echo "refused: main moved past the fast gate's base ${fastBase:0:8} in paths the gate didn't see change ($(printf '%s' "$untested" | head -n 3 | paste -sd ' ' -)); merge main in and fast-gate again" >&2
		exit 1
	fi
	# The fast gate reads its smoke list from the gated tree when the tree has one, so a landing could
	# shrink the set that judges it. The list changes only by a reviewed commit: a landing whose list
	# differs from main's says so (--smoke-list-reviewed), and the note names it.
	smokeList=cloud/fast-gate/smoke.txt
	landingList=$(git rev-parse -q --verify "${sha}:${smokeList}" 2>/dev/null || echo none)
	mainList=$(git rev-parse -q --verify "origin/main:${smokeList}" 2>/dev/null || echo none)
	if [ "$landingList" != "$mainList" ]; then
		if [ "$smokeReviewed" != yes ]; then
			echo "refused: ${sha:0:8} changes the fast gate's smoke list (${smokeList}: main ${mainList:0:8}, here ${landingList:0:8}), which judges this landing; review the change, then pass --smoke-list-reviewed" >&2
			exit 1
		fi
		fastNote="${fastNote}; smoke list changed by this landing (reviewed), ${mainList:0:8} to ${landingList:0:8}"
	fi
	branches="${branches}; ${fastNote}"
fi

# The pause rule, read from the whole gates on main.
pause=$(python3 - "$directory/main-reds.tsv" <<'PAUSE'
import os, subprocess, sys
explained = set()
if os.path.exists(sys.argv[1]):
    for line in open(sys.argv[1]):
        if line.strip() and not line.startswith("#"):
            # A row that starts OPEN names the red's owner and its fix-forward (for the star-idle
            # alarm) without explaining it: landings stay paused until the row is lifted or fixed.
            reference, _, note = line.partition("\t")
            if note.startswith("OPEN"):
                continue
            explained.add(reference.removeprefix("origin/"))
references = subprocess.run(["git", "ls-remote", "origin", "refs/heads/gate-logs/*"], capture_output=True, text=True, check=True).stdout.split("\n")
logs = {}
for line in references:
    if line.endswith("/full-main"):
        name = line.split("\t")[1].removeprefix("refs/heads/")
        logs.setdefault(name.split("/")[1], []).append(name)
# Every commit main holds, newest first. Not --first-parent: a landing that fast-forwards main to an
# area's tip puts the area's own history on main's first-parent line, and a walk along it never
# reaches the mains before (the first fast-gate landing hid 8b388310's full-main this way).
mains = subprocess.run(["git", "rev-list", "--topo-order", "origin/main"], capture_output=True, text=True, check=True).stdout.split()
for main in mains:
    if main[:12] not in logs:
        continue
    for name in sorted(logs.get(main[:12], []), reverse=True):
        subprocess.run(["git", "fetch", "-q", "origin", f"+refs/heads/{name}:refs/remotes/origin/{name}"], check=True)
        status = subprocess.run(["git", "show", f"origin/{name}:status.txt"], capture_output=True, text=True).stdout.split("\n")[0]
        if status.startswith("running") or not status:
            continue
        if status.startswith("green"):
            print(f"green {name}")
        elif name in explained:
            print(f"explained {name}")
        else:
            print(f"red {name} {status}")
        sys.exit(0)
print("none")
PAUSE
)
case "$pause" in
red\ *)
	if [ "$testOnly" = yes ]; then
		echo "Landing test-only files while main's whole gate is red (${pause#red }); they can't change it."
	else
	redLog=$(printf '%s' "$pause" | awk '{print $2}')
	if [ "$pauseException" = revert ]; then
		echo "Landing a revert while main's whole gate is red (${redLog})."
		branches="${branches}; revert while main is red (${redLog})"
	elif [ "$pauseException" = "fix-forward ${redLog}" ]; then
		echo "Landing a fix-forward for main's red whole gate (${redLog})."
		branches="${branches}; fix-forward for ${redLog}"
	else
		echo "refused: landings are paused, main's newest finished whole gate is red: ${pause#red }. Land a revert (--revert) or a fix-forward (--fix-forward ${redLog}), or explain the red in cloud/integration/main-reds.tsv" >&2
		exit 1
	fi
	fi
	;;
none) echo "No finished whole gate on main yet (gate-logs/*/full-main); the pause rule has nothing to read." ;;
*) echo "Main's newest finished whole gate: ${pause}." ;;
esac
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
if [ "$testOnly" = yes ]; then
	if ! tree=$(git merge-tree --write-tree "$old" "$sha" | head -n 1) || [ -z "$tree" ]; then
		echo "refused: ${sha:0:8} doesn't merge cleanly with main ${old:0:8}" >&2
		exit 1
	fi
	# Tests only: Go test files, testdata, review evidence and shard tables. Anything else needs a gate.
	nonTest=$(git diff --name-only "$old" "$tree" | grep -v -E '(_test\.go$|/testdata/|^review/|(^|/)shards\.json$)' | grep . || true)
	if [ -n "$nonTest" ]; then
		echo "refused: not test-only against main ${old:0:8}: $(printf '%s' "$nonTest" | head -n 5 | paste -sd ' ' -)" >&2
		exit 1
	fi
	if ! git merge-base --is-ancestor "$old" "$sha" || [ "$(git rev-parse "${sha}^{tree}")" != "$tree" ]; then
		sha=$(git commit-tree "$tree" -p "$old" -p "$sha" -m "Land test-only ${sha:0:8} over main ${old:0:8}

Every path it changes against main is a test, testdata, review evidence or a shard table, so it lands with no
gate (Kirk, Oct 8); Loom's next whole-suite run of main is its check.")
	fi
	echo "Landing test-only ${sha:0:8} over main ${old:0:8}."
elif ! git merge-base --is-ancestor "$old" "$sha"; then
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
