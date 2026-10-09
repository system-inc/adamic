#!/usr/bin/env bash
# Moves main to a sha whose gate ran green and uncached, in one landing commit, and merges main back
# into every area. Only integration runs it. It checks what it can (a full sha, main still its
# ancestor, no failures) and refuses otherwise.
#
# Each landing is one commit on main: first parent the old main, second the landed sha, its tree
# exactly what was gated (or, over a main that moved by records and tests only, that plus main's), and
# its numbers as trailers: Old-main, Landed-commits, Branches, Gate-minutes, Pass, Fail, Skip, Backlog
# (@system_adamic, Oct 9, from Kirk: "whats the point of all these record the landings?"). Nothing else
# is committed, so main moves once per landing. `git log --first-parent` over the landing commits is
# the velocity table, and landings.py beside this script prints it as CSV.
#
# Gate minutes are wall time from the first gate start to the last green piece (the whole run, its
# reruns, input runs and groups), so every landing means the same thing (@system_adamic, October 7).
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
# usage: cloud/integration/push-main.sh [--revert | --fix-forward <red log ref>] <full sha> <gate minutes> <pass> <fail> <skip> "<branches landed>"
#        cloud/integration/push-main.sh [same options] (--fast-gate | --full-gate) <gate-logs ref> [--smoke-list-reviewed] <full sha> "<branches landed>"
#        cloud/integration/push-main.sh --test-only <full sha> "<branches landed>"
#        cloud/integration/push-main.sh --deletion [--not-a-reader <file>]... <full sha> "<branches landed>"
#        cloud/integration/push-main.sh --test-only --markdown-corpus "<its corpus verdict>" <full sha> "<branches landed>"
set -euo pipefail

fastGate=""
alsoGates=()
mainReds=""
gateKind=fast
smokeReviewed=no
pauseException=""
testOnly=no
deletion=no
extraTrailers=""
notReaders=()
markdownCorpus=""
ruledReds=()
while [ "$#" -gt 0 ]; do
	case "$1" in
	--fast-gate) fastGate=${2#origin/}; shift 2 ;;
	# --full-gate <gate-logs ref>/full-main of this exact sha: the whole gate as the landing's verdict
	# and main's confirmation in one (@system_adamic, October 8). It is a superset of the fast gate,
	# deferred tests included, so the landed tree is exactly the tree that passed everything.
	--full-gate) fastGate=${2#origin/}; gateKind=full; shift 2 ;;
	# --also-gate <gate-logs ref>, repeatable: another record of the same sha whose stages complete the
	# gate's (@system_adamic, Oct 9 04:34Z: a pool record of Go tests only lands beside a record of the
	# build, vet, smoke and census stages; a stage that ran on another runner counts when it ran on the
	# same sha). Every record must be green and finished, and together they must cover every stage.
	--also-gate) alsoGates+=("${2#origin/}"); shift 2 ;;
	# --main-reds <gate-logs ref of a finished whole gate of main>: the candidate lands on 'no new reds against
	# main' (@system_adamic, Oct 9 05:28Z): every top-level test its records fail must fail in that record of
	# main too. Those reds are named in the landing and stay owned on main's red list; one red main doesn't
	# have refuses it, by name. It never lands a red main doesn't already carry.
	--main-reds) mainReds=${2#origin/}; shift 2 ;;
	# --infra-red "<package> <Test>=<ruling>", repeatable: a red @system_adamic has ruled infrastructure on every
	# candidate (Oct 9 03:34: typeaware TestVolumeAgreementAndMutants, green on main run alone, killed by its own 90 s
	# per-command limit under pool load; #t4b9j71 splits it). By that exact name only, it isn't the candidate's red, alone
	# or beside --main-reds, and the landing names the ruling. Check the candidate's log shows the ruled cause (the kill
	# line) before passing it: the same test failing another way is the candidate's.
	--infra-red) ruledReds+=("$2"); shift 2 ;;
	--smoke-list-reviewed) smokeReviewed=yes; shift ;;
	# --test-only <sha> "<branches>": a change that touches only tests goes to main with no gate in front of it;
	# Loom's next whole-suite run of main is its check (Kirk, Oct 8). The merged diff must be tests only.
	--test-only) testOnly=yes; shift ;;
	# --deletion <sha> "<branches>": a change that only deletes files no code reads goes to main with no gate,
	# since a gate of it would test main's tree unchanged (@system_adamic, Oct 9 07:53Z: "a deletion no code
	# reads lands ungated", first for the velocity table's CSV). It rides the test-only lane's checks.
	--deletion) testOnly=yes; deletion=yes; shift ;;
	# --not-a-reader <file>, repeatable: a file that names a deleted path without reading it (a list of record
	# paths, a test writing its own copy), checked by hand; the landing names each one.
	--not-a-reader) notReaders+=("$2"); shift 2 ;;
	# --markdown-corpus "<verdict>", with --test-only: Markdown files whose only reader is the markdown corpus land on
	# that corpus's units alone (@system_adamic, Oct 9 02:16: "a doc no code reads except the markdown corpora is
	# gated by the markdown corpora alone"; internal/corpusfiles takes every tracked .md). The verdict names where
	# those units ran green on this sha merged onto main; the landing carries it. Not a docs class: the run is required.
	--markdown-corpus) markdownCorpus=$2; shift 2 ;;
	--revert) pauseException=revert; shift ;;
	--fix-forward) pauseException="fix-forward ${2#origin/}"; shift 2 ;;
	*) break ;;
	esac
done
if [ "$testOnly" = yes ]; then
	if [ "$#" -ne 2 ]; then
		echo "usage: $0 --test-only <full sha> \"<branches landed>\"" >&2
		exit 2
	fi
	sha=$1
	branches="$2; test-only lane, no gate (Kirk, Oct 8)"
	[ "$deletion" = no ] || branches="$2; deletion no code reads, no gate (@system_adamic, Oct 9 07:53Z)"
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
# A candidate that doesn't hold main's tip is gated as a merge commit onto it (developer tools #11ymb02, tools
# a79ee19), kept at refs/gate-merges/<merge sha> rather than on a branch; that merge is what lands, so fetch it.
git cat-file -e "${sha}^{commit}" 2>/dev/null || git fetch -q origin "refs/gate-merges/${sha}" 2>/dev/null || true
# The paths the test-only lane takes with no gate (Kirk, Oct 8). One list, each piece by ruling: Go tests (_test.go) and
# testdata; Python tests (_test.py, test_*.py, and -test.py since @system_adamic Oct 9 04:20, refused when a non-test
# file names one); review evidence; shard tables; stage3 fixtures and meter (Oct 8); the root README; internal/oracle/
# counts.md rows added for fixtures the same change adds (Oct 9 05:27, below); and with --markdown-corpus, Markdown
# whose corpus units ran green (Oct 9 02:16). Main moving by them doesn't spend a
# candidate's gate either: the lane already lands them ungated, and Loom's whole run of main is their
# check (@system_adamic, Oct 9). star-train.py's testOnlyPaths is the same pattern.
# One guard: a test-only change in a package where the candidate changes code still counts, since a
# test asserting the old behavior meets the new code there.
testOnlyPattern='(_test\.go$|_test\.py$|-test\.py$|(^|/)test_[^/]*\.py$|/testdata/|^review/|(^|/)shards\.json$|^stage3/fixtures/|^stage3/meter/|^README\.md$)'
# The packages (directories) where <base>..<tip> changes anything but tests and records.
codePackages() {
	local path
	git diff --name-only "$1" "$2" | while IFS= read -r path; do
		case "$path" in documentation/velocity/landings.csv | stage3/meter/runs/* | stage3/progress.json) continue ;; esac
		[[ "$path" =~ $testOnlyPattern ]] && continue
		printf '%s\n' "${path%/*}"
	done | sort -u
}
# Reads paths main changed past a gate and prints the ones that spend it: not records, and not
# test-only unless in one of the candidate's code packages ($1, one per line).
countingPaths() {
	local path package
	while IFS= read -r path; do
		[ -n "$path" ] || continue
		case "$path" in documentation/velocity/landings.csv | stage3/meter/runs/* | stage3/progress.json) continue ;; esac
		if [[ "$path" =~ $testOnlyPattern ]]; then
			package=${path%%/testdata/*}
			[ "$package" != "$path" ] || package=${path%/*}
			printf '%s\n' "$1" | grep -qxF -- "$package" || continue
		fi
		printf '%s\n' "$path"
	done
}
directory=$(cd "$(dirname "$0")" && pwd)
# The oracle's counts ledger in the test-only lane (@system_adamic, Oct 9 05:27): added rows only, each naming a fixture
# the same change adds under testdata. A changed or removed row is a counts change and gets audited through a gate.
countsFile=internal/oracle/counts.md
countsRowsAdded() {
	python3 - "$1" "$2" "$countsFile" <<'COUNTS'
import re, subprocess, sys
old, tree, counts = sys.argv[1:4]
diff = subprocess.run(["git", "diff", "-U0", old, tree, "--", counts], capture_output=True, text=True).stdout.splitlines()
added = {path for path in subprocess.run(["git", "diff", "--name-only", "--diff-filter=A", old, tree], capture_output=True, text=True).stdout.split()}
for line in diff:
    if line.startswith("-") and not line.startswith("---"):
        sys.exit("it changes or removes a row: " + line[1:80])
for line in diff:
    if not line.startswith("+") or line.startswith("+++"):
        continue
    row = re.match(r"\+\| *([^ |]+) *\|", line)
    if not row:
        sys.exit("it adds a line that isn't a fixture's row: " + line[1:80])
    if row.group(1) not in added or "/testdata/" not in row.group(1):
        sys.exit("it adds a row for %s, which this change doesn't add under testdata" % row.group(1))
COUNTS
}
# Prints why a test-only change (its paths on stdin's argument, one per line) waits for the star, or
# nothing: the board's star (@system_adamic, Oct 9 04:56Z: the hold protects only the star, never a train
# below it), named by its landing branch in ~/.adamic-integration/star, while a gate of it is running (a
# fast gate in the watcher's running set, or a whole gate on origin still reading running, under 90
# minutes old) and not on main, and the change touches a package where it changes code.
starHolds() {
	local candidate name running= package
	name=$(head -n 1 "${ADAMIC_STAR_FILE:-$HOME/.adamic-integration/star}" 2>/dev/null)
	[ -n "$name" ] || return 0
	candidate=$(git rev-parse -q --verify "origin/${name}" 2>/dev/null) || return 0
	candidate=$(git rev-parse -q --verify "${candidate}^{commit}" 2>/dev/null) || return 0
	git merge-base --is-ancestor "$candidate" origin/main && return 0
	if grep -qs -- " ${candidate} " "${ADAMIC_FAST_GATE_STATE:-$HOME/.adamic-fast-gate-watch}"/running/*; then
		running=fast
	else
		for ref in $(git ls-remote origin "refs/heads/gate-logs/${candidate:0:12}/*/full-main" | awk '{print $2}'); do
			git fetch -q origin "+${ref}:refs/remotes/origin/${ref#refs/heads/}" 2>/dev/null || continue
			# A record still reading running after 90 minutes is a gate that died, not one to wait for.
			[ $(($(date +%s) - $(git log -1 --format=%ct "origin/${ref#refs/heads/}"))) -lt 5400 ] || continue
			git show "origin/${ref#refs/heads/}:status.txt" 2>/dev/null | head -n 1 | grep -q '^running' && running=whole
		done
	fi
	[ -n "$running" ] || return 0
	local code
	code=$(codePackages "$(git merge-base "$candidate" origin/main)" "$candidate")
	while IFS= read -r path; do
		[ -n "$path" ] || continue
		package=${path%%/testdata/*}
		[ "$package" != "$path" ] || package=${path%/*}
		if printf '%s\n' "$code" | grep -qxF -- "$package"; then
			echo "${name} ${candidate:0:8} has its ${running} gate running and changes code in ${package}"
			return 0
		fi
	done <<<"$1"
}

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
	# A unit that named tests and ran none of them read as passed (@system_adamic, Oct 9 11:01Z: a void is never a pass).
	# Loom's zerorun.py is the one check (no second copy here to drift): every record that carries its job (job.json) is
	# checked out and run through it, the gate's and each --also-gate's; one zero-run spec refuses, by unit and package.
	zerorun=${ADAMIC_ZERORUN:-$HOME/.loom/bin/zerorun.py}
	for record in "$fastGate" ${alsoGates[@]+"${alsoGates[@]}"}; do
		git fetch -q origin "+refs/heads/${record}:refs/remotes/origin/${record}" 2>/dev/null || continue
		git cat-file -e "origin/${record}:job.json" 2>/dev/null || continue
		if [ ! -f "$zerorun" ]; then
			echo "refused: ${record} carries its job but ${zerorun} isn't here to check it for zero-run units" >&2
			exit 1
		fi
		checkout=$(mktemp -d)
		git archive "origin/${record}" | tar -x -C "$checkout"
		# Under set -e a failing assignment would end the script silently, so the exit status is taken explicitly.
		code=0
		lost=$(python3 "$zerorun" "$checkout" 2>&1) || code=$?
		rm -rf -- "$checkout"
		if [ "$code" -ne 0 ]; then
			echo "refused: ${record} has units that ran none of the tests they named (zerorun.py exit ${code}): $(printf '%s\n' "$lost" | cut -f2- | head -n 3 | paste -sd ';' -)" >&2
			exit 1
		fi
	done
	for also in ${alsoGates[@]+"${alsoGates[@]}"}; do
		if ! git fetch -q origin "+refs/heads/${also}:refs/remotes/origin/${also}" 2>/dev/null; then
			echo "refused: no gate log ${also} on origin" >&2
			exit 1
		fi
		alsoStatus=$(git show "origin/${also}:status.txt" 2>/dev/null | head -n 1)
		alsoJSON=$(git show "origin/${also}:fast.json" 2>/dev/null || git show "origin/${also}:full.json" 2>/dev/null || true)
		if ! printf '%s' "$alsoStatus" | grep -q '^green' || [ -z "$alsoJSON" ]; then
			echo "refused: ${also} isn't a green record (${alsoStatus:-no status})" >&2
			exit 1
		fi
		# The union: the stages either ran, failures summed, wall time added, the first record's base and
		# packages kept (its diff chose them). The verdict below then judges the union as one record.
		if ! printf '%s' "$alsoJSON" | python3 -c '
import json, sys
path, sha = sys.argv[1], sys.argv[2]
first, other = json.load(open(path)), json.load(sys.stdin)
if other.get("sha") != sha or other.get("finished") is not True:
    sys.exit("it gated %s (finished %r), not %s" % (other.get("sha"), other.get("finished"), sha))
for key in ("steps_seconds", "stages_exit"):
    first[key] = dict(first.get(key) or {}, **(other.get(key) or {}))
first["planned_stages"] = sorted(set(first.get("planned_stages") or []) | set(other.get("planned_stages") or []))
first["fail"] = (first.get("fail") or 0) + (other.get("fail") or 0)
first["wall_seconds"] = float(first.get("wall_seconds") or 0) + float(other.get("wall_seconds") or 0)
for key in ("build_ok", "vet_ok", "uncached_tests"):
    first[key] = bool(first.get(key) or other.get(key))
for key in ("unclassified_skips", "required_input_skips", "undeclared_tools"):
    first[key] = (first.get(key) or []) + (other.get(key) or [])
for key in ("smoke_list", "smoke_list_blob", "machine", "tools_sha"):
    first.setdefault(key, other.get(key))
json.dump(first, open(path, "w"))
' "$fastJSON" "$sha"; then
			echo "refused: ${also} doesn't complete ${fastGate} for ${sha:0:8}" >&2
			exit 1
		fi
		fastGate="${fastGate} + ${also}"
	done
	if [ -n "$mainReds" ] || [ "${#ruledReds[@]}" -gt 0 ]; then
		if [ -n "$mainReds" ] && ! git fetch -q origin "+refs/heads/${mainReds}:refs/remotes/origin/${mainReds}" 2>/dev/null; then
			echo "refused: no gate log ${mainReds} on origin" >&2
			exit 1
		fi
		# The candidate's failing top-level tests, from every record it lands on, against main's.
		known=$(RULED_REDS="$(printf '%s\n' ${ruledReds[@]+"${ruledReds[@]}"} | sed 's/=.*//')" python3 - "$fastJSON" "origin/${fastGate%% + *}" "origin/${mainReds}" ${alsoGates[@]+"${alsoGates[@]/#/origin/}"} 2>&1 <<'MAINREDS'
import gzip, json, subprocess, sys
path, mainRef, records = sys.argv[1], sys.argv[3], [sys.argv[2]] + sys.argv[4:]
def failing(ref):
    raw = subprocess.run(["git", "show", ref + ":test.jsonl.gz"], capture_output=True).stdout
    if not raw:
        # A box's fast record carries no test.jsonl.gz; its json names the failing tests instead.
        for name in ("fast.json", "full.json"):
            record = subprocess.run(["git", "show", ref + ":" + name], capture_output=True, text=True).stdout
            if record and "failed_tests" in json.loads(record):
                return {entry.split("/adamic/")[-1] for entry in json.loads(record)["failed_tests"] or [] if "/" not in entry.split()[-1]}
        return None
    names = set()
    for line in gzip.decompress(raw).decode(errors="replace").splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        if event.get("Action") == "fail" and event.get("Test") and "/" not in event["Test"]:
            names.add(event["Package"].split("/adamic/")[-1] + " " + event["Test"])
    return names
import os
ruled = {line.strip() for line in os.environ.get("RULED_REDS", "").splitlines() if line.strip()}
if mainRef == "origin/":
    # --infra-red alone: no main record; only the ruled names are excused.
    onMain = set()
else:
    mainStatus = subprocess.run(["git", "show", mainRef + ":status.txt"], capture_output=True, text=True).stdout.split("\n")[0]
    mainFull = json.loads(subprocess.run(["git", "show", mainRef + ":full.json"], capture_output=True, text=True).stdout or "{}")
    onMain = failing(mainRef)
    if not mainStatus.startswith(("green", "red")) or mainFull.get("finished") is not True or onMain is None:
        sys.exit("main's record %s isn't a finished whole gate with a test record (%s)" % (mainRef, mainStatus[:60]))
    if subprocess.run(["git", "merge-base", "--is-ancestor", mainFull.get("sha", ""), "origin/main"]).returncode != 0:
        sys.exit("main's record gated %s, which isn't on main" % mainFull.get("sha"))
onMain |= ruled
ours = set()
for ref in records:
    names = failing(ref)
    if names is None and ref == records[0]:
        sys.exit("%s has no test record to compare" % ref)
    ours |= names or set()
# The standing class (@system_adamic, Oct 9 05:38): a red is infra by exact name only when its output is a test-owned
# time limit (a per-command kill, a package watchdog) with no assertion text at all. --infra-red names it and its
# evidence; this checks the output so a ruled name can't hide a real failure of the same test.
import re
timeLimit = re.compile(r"signal: killed|exceeded|watchdog|deadline|timed out|cooked")
assertion = re.compile(r"\b(want|got|expected|mismatch|differs?|disagree)\b", re.I)
def outputOf(ref, package, test):
    raw = subprocess.run(["git", "show", ref + ":test.jsonl.gz"], capture_output=True).stdout
    if not raw:
        return None
    lines = []
    for line in gzip.decompress(raw).decode(errors="replace").splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        name = event.get("Test") or ""
        if event.get("Action") == "output" and event.get("Package", "").endswith("/" + package) and (name == test or name.startswith(test + "/")):
            lines.append(event.get("Output", ""))
    return lines
for name in sorted(ruled & ours):
    package, test = name.split(" ", 1)
    lines = None
    for ref in records:
        lines = outputOf(ref, package, test)
        if lines:
            break
    if not lines:
        sys.exit("%s is named infra but its output isn't in the record to check" % name)
    if not any(timeLimit.search(line) for line in lines):
        sys.exit("%s is named infra but its output shows no test-owned time limit" % name)
    asserted = [line.strip() for line in lines if assertion.search(line)]
    if asserted:
        sys.exit("%s is named infra but its output asserts: %s" % (name, asserted[0][:120]))
new = sorted(ours - onMain)
if new:
    sys.exit(("new reds against main: " if mainRef != "origin/" else "reds not ruled infra: ") + "; ".join(new[:8]))
fast = json.load(open(path))
expected = len(ours)
if (fast.get("fail") or 0) > expected:
    sys.exit("%s failures recorded but only %d failing top-level tests named" % (fast.get("fail"), expected))
# Every red is main's own: the verdict judges the rest.
fast["fail"] = 0
fast["main_reds"] = sorted(ours)
fast["stages_exit"] = {stage: (0 if stage == "tests" else code) for stage, code in (fast.get("stages_exit") or {}).items()}
json.dump(fast, open(path, "w"))
print(len(ours))
MAINREDS
) || { echo "refused: $([ -n "$mainReds" ] && echo "--main-reds ${mainReds}" || echo "--infra-red"): ${known}" >&2; exit 1; }
		statusLine="green: ${sha} with ${known} reds main already has or ruled infra (${mainReds:-no main record})"
		[ -z "$mainReds" ] || branches="${branches}; lands with ${known} reds main already has or ruled infra (${mainReds}), owned on main's red list"
		[ "${#ruledReds[@]}" -eq 0 ] || branches="${branches}; ruled infra, not the candidate's: $(printf '%s; ' "${ruledReds[@]}" | sed 's/; $//')"
	fi
	if ! verdict=$(GATE_KIND="$gateKind" RERUN_MERGE="$(dirname "${BASH_SOURCE[0]}")/rerun_merge.py" python3 - "$sha" "$statusLine" "$fastGate" "$fastJSON" <<'VERDICT'
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
# A rerun on a run's kept verdicts (Kirk, Oct 8 21:17 MDT; #hpjftdj): after a test killed at 90 s is fixed, only it
# reruns, plus the units whose input hash its change moved, and the base run's other verdicts stand. rerun_merge.py
# holds the rule; the base record is the gate-logs ref the rerun names.
if fast.get("rerun_of"):
    import importlib.util
    merge = importlib.util.spec_from_file_location("rerun_merge", os.environ.get("RERUN_MERGE", "cloud/integration/rerun_merge.py"))
    rerunMerge = importlib.util.module_from_spec(merge)
    merge.loader.exec_module(rerunMerge)
    baseRecord = None
    if subprocess.run(["git", "fetch", "-q", "origin", fast["rerun_of"]], capture_output=True).returncode == 0:
        try:
            baseRecord = json.loads(subprocess.run(["git", "show", "FETCH_HEAD:full.json"], capture_output=True, text=True).stdout)
        except ValueError:
            pass
    if baseRecord is None:
        problems.append("its base run %s has no readable full.json on origin" % fast["rerun_of"])
    else:
        problems.extend(rerunMerge.problems(baseRecord, fast, rerunMerge.descends(baseRecord.get("sha", ""), sha)))
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
	# A rerun on kept verdicts names the units it ran, and a B (#mbexftz: the moved main merged with the gated sha,
	# its subject "Moved main: ...") names the main the gate saw and the one it lands over.
	rerunUnits=$(python3 -c 'import json, sys
record = json.load(open(sys.argv[1]))
if record.get("rerun_of"):
    print(" ".join(sorted(unit["id"] for unit in record.get("units") or [] if unit.get("verdict") != "kept")) or "none")' "$fastJSON")
	[ -z "$rerunUnits" ] || extraTrailers="Rerun-units: ${rerunUnits}"
	if [ "$(git log -1 --format=%s "$sha" | cut -c1-11)" = "Moved main:" ]; then
		extraTrailers="${extraTrailers:+${extraTrailers}
}Moved-main: $(git merge-base "${sha}^2" "${sha}^1")..$(git rev-parse "${sha}^1")"
	fi
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
	movedSince=$(git diff --name-only "$fastBase" origin/main | countingPaths "$(codePackages "$fastBase" "$sha")")
	untested=$(comm -23 <(printf '%s\n' "$movedSince" | grep . | sort) <(git diff --name-only "$fastBase" "$sha" | sort) || true)
	if [ -n "$untested" ]; then
		why="main moved past the fast gate's base ${fastBase:0:8} in paths the gate didn't see change ($(printf '%s' "$untested" | head -n 3 | paste -sd ' ' -))"
		# B, main merged with the gated sha, descends from the gate (#mbexftz). On a whole pool record, Loom's rerun.sh
		# reruns only the units whose input hash the move changed (inputs.py rerun-plan) and publishes a rerun_of record
		# keeping the rest's verdicts; B lands on it with --full-gate. Held (exit 3) until then. A conflict means recut.
		if movedTree=$(git merge-tree --write-tree origin/main "$sha" | head -n 1) && [ -n "$movedTree" ]; then
			moved=$(git commit-tree "$movedTree" -p "$(git rev-parse origin/main)" -p "$sha" -m "Moved main: ${sha:0:8} merged onto main $(git rev-parse --short=8 origin/main), for a rerun of the units the move changed")
			rerun=${LOOM_RERUN:-$HOME/.loom/bin/rerun.sh}
			if [[ "$fastGate" =~ ^gate-logs/[0-9a-f]{12}/[0-9TZ]+/full-main$ ]] && [ -x "$rerun" ]; then
				git push -q origin "${moved}:refs/gate-merges/${moved}"
				logs=${ADAMIC_RERUN_LOGS:-$HOME/.adamic-integration/reruns}
				mkdir -p "$logs"
				nohup "$rerun" "$fastGate" "$moved" >"${logs}/${moved}.log" 2>&1 &
				echo "held: ${why}; rerunning the units the move changed on B ${moved} (${rerun} ${fastGate}, log ${logs}/${moved}.log), then land B with --full-gate <the record it publishes>" >&2
				exit 3
			fi
			why="${why}; rerun the units the move changed on B ${moved}"
		fi
		echo "refused: ${why}; merge main in and fast-gate again" >&2
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
		echo "Landing $([ "$deletion" = yes ] && echo 'a deletion no code reads' || echo 'test-only files') while main's whole gate is red (${pause#red }); they can't change it."
	else
	redLog=$(printf '%s' "$pause" | awk '{print $2}')
	if [ "$pauseException" = revert ]; then
		echo "Landing a revert while main's whole gate is red (${redLog})."
		branches="${branches}; revert while main is red (${redLog})"
	elif [ -n "$mainReds" ] && [ "$mainReds" = "$redLog" ]; then
		# Compared against this very red record and adding none of its own (--main-reds), it can't make main
		# worse; the reds stay main's, owned on its red list (@system_adamic, Oct 9 05:28Z).
		echo "Landing over main's red whole gate (${redLog}) with no new reds against it."
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
if [ "$old" = "$sha" ] || git merge-base --is-ancestor "$sha" "$old"; then
	# Already on main: the train or another hand landed it first. Landing it again would push an empty
	# landing commit counting the same work twice (Oct 9: e3f2be21 landed twice, 5bc33818..b620d51b).
	echo "refused: main ${old:0:8} already holds ${sha:0:8}" >&2
	exit 1
fi
gated=$sha
if [ "$testOnly" = yes ]; then
	if ! tree=$(git merge-tree --write-tree "$old" "$gated" | head -n 1) || [ -z "$tree" ]; then
		echo "refused: ${gated:0:8} doesn't merge cleanly with main ${old:0:8}" >&2
		exit 1
	fi
	# Tests only: Go test files, testdata, review evidence and shard tables. Anything else needs a gate.
	# Harness directories read only by tests and gates, never by the compiler, runtime, library or a shipped
	# tool, are allowed too, each added by ruling (@system_adamic, Oct 8: stage3/fixtures and stage3/meter).
	changed=$(git diff --name-only "$old" "$tree")
	if [ "$deletion" = yes ]; then
		# Every path a deletion, in its history too, and none named by code on main or by developer tools'
		# gate tools, which read main's tree: then nothing that runs can see the change.
		kept=$(git diff --name-only --diff-filter=d "$old" "$tree")
		if [ -z "$changed" ] || [ -n "$kept" ]; then
			echo "refused: not deletions only against main ${old:0:8}: $(printf '%s' "$kept" | head -n 5 | paste -sd ' ' -)" >&2
			exit 1
		fi
		for commit in $(git rev-list --reverse --topo-order --no-merges "${old}..${gated}"); do
			outside=$(git diff-tree --no-commit-id --name-only --diff-filter=d -r "$commit")
			if [ -n "$outside" ]; then
				echo "refused: carries more than deletions: ${commit:0:8} changes $(printf '%s' "$outside" | head -n 3 | paste -sd ' ' -)" >&2
				exit 1
			fi
		done
		tools=$(git rev-parse -q --verify origin/devtools/fast-gate 2>/dev/null || true)
		for path in $changed; do
			for where in "$old" $tools; do
				readers=$(git grep -l -F -e "$path" -e "${path##*/}" "$where" -- '*.go' '*.py' '*.sh' '*.mjs' '*.cjs' '*.js' '*.ts' '*.json' '*.yml' '*.yaml' '*.toml' ':!*/testdata/*' | sed 's/^[^:]*://' | grep -v -x -F -e '' "${notReaders[@]/#/-e}" || true)
				if [ -n "$readers" ]; then
					echo "refused: code reads ${path}: $(printf '%s\n' "$readers" | head -n 3 | paste -sd ' ' -); if one only names it, check by hand and pass --not-a-reader <file>" >&2
					exit 1
				fi
			done
		done
		landingSubject="Land deletion ${gated:0:8} over main ${old:0:8}"
		landingBody="It only deletes $(printf '%s' "$changed" | paste -sd ' ' -), which no code on main or in developer tools'
gate tools reads, so a gate would test main's tree unchanged; it lands with no gate (@system_adamic, Oct 9
07:53Z: a deletion no code reads lands ungated)."
		[ "${#notReaders[@]}" -eq 0 ] || landingBody="${landingBody}
Named but not read, checked by hand: ${notReaders[*]}."
		echo "Landing deletion ${gated:0:8} over main ${old:0:8}."
	else
	nonTest=$(printf '%s\n' "$changed" | grep -v -E "$testOnlyPattern" | grep . || true)
	countsAdded=no
	if printf '%s\n' "$nonTest" | grep -qxF "$countsFile"; then
		if ! countsWhy=$(countsRowsAdded "$old" "$tree" 2>&1); then
			echo "refused: ${countsFile} in the test-only lane takes added fixture rows only; ${countsWhy}" >&2
			exit 1
		fi
		countsAdded=yes
		nonTest=$(printf '%s\n' "$nonTest" | grep -v -x -F -e "$countsFile" | grep . || true)
	fi
	if [ -n "$markdownCorpus" ]; then
		nonTest=$(printf '%s\n' "$nonTest" | grep -v -E '\.md$' | grep . || true)
		branches="${branches}; Markdown gated by its corpus alone (@system_adamic, Oct 9 02:16): ${markdownCorpus}"
	fi
	if [ -n "$nonTest" ]; then
		echo "refused: not test-only against main ${old:0:8}: $(printf '%s' "$nonTest" | head -n 5 | paste -sd ' ' -)" >&2
		exit 1
	fi
	# A Python test named by a non-test file (a gate script, a .sh that runs it) is gate logic and gets a gate
	# (@system_adamic, Oct 9 04:20, ruling *-test.py into the lane). Go's _test.go can't be imported; a Python test can
	# be run or read by name, so the lane looks for its name outside tests. A document naming it (.md, .txt) runs nothing.
	for path in $(printf '%s\n' "$changed" | grep -E '(_test\.py$|-test\.py$|(^|/)test_[^/]*\.py$)' || true); do
		users=$(git grep -l -F -e "${path##*/}" "$tree" -- . 2>/dev/null | sed 's/^[^:]*://' | grep -v -x -F -e "$path" | grep -v -E "$testOnlyPattern" | grep -v -E '\.(md|txt)$' || true)
		if [ -n "$users" ]; then
			echo "refused: ${path} is named by $(printf '%s' "$users" | head -n 3 | paste -sd ' ' -), which isn't a test, so it's gate logic; gate it" >&2
			exit 1
		fi
	done
	# Its history, not only its tree (@system_adamic, Oct 9 04:59Z): a commit beyond main that touches anything
	# but tests would be recorded as merged even where the merge drops its changes, and a later plain merge of
	# its branch then silently deletes them (cohere's estree split carried buildcache-shared 2afbfa75 this way).
	# The lane never filters files out of a merge; the worker cherry-picks its test commits onto main instead.
	for commit in $(git rev-list --reverse --topo-order --no-merges "${old}..${gated}"); do
		outside=$(git diff-tree --no-commit-id --name-only -r "$commit" | grep -v -E "$testOnlyPattern" | { if [ -n "$markdownCorpus" ]; then grep -v -E '\.md$'; else cat; fi; } | { if [ "$countsAdded" = yes ]; then grep -v -x -F -e "$countsFile"; else cat; fi; } | grep . || true)
		if [ -n "$outside" ]; then
			echo "refused: carries non-test history: ${commit:0:8} ($(git log -1 --format=%s "$commit" | cut -c1-60)) changes $(printf '%s' "$outside" | head -n 3 | paste -sd ' ' -); cherry-pick the test commits onto main instead" >&2
			exit 1
		fi
	done
	# A harness change can hide a check that can't fail, so it lands only with its mutant evidence.
	harness=$(printf '%s\n' "$changed" | grep -E '^stage3/(fixtures|meter)/' | grep -v -E '(_test\.go$|/testdata/)' | grep . || true)
	if [ -n "$harness" ] && ! printf '%s\n' "$changed" | grep -q -i 'mutant'; then
		echo "refused: it changes test harness ($(printf '%s' "$harness" | head -n 3 | paste -sd ' ' -)) with no mutant evidence among its files" >&2
		exit 1
	fi
	# The star's hold (@system_adamic, Oct 9 03:48Z, interim until Loom's rerun by hash): while a star
	# slice's gate is running, a test-only change touching that slice's code packages waits, since landing
	# it would make the train re-cut the slice and void the gate. It lands the minute the gate finishes,
	# green or red. Exit 3 means held; pr-lane.py tries again every minute.
	held=$(starHolds "$changed")
	if [ -n "$held" ]; then
		echo "held for the star: ${held}" >&2
		exit 3
	fi
	landingSubject="Land test-only ${gated:0:8} over main ${old:0:8}"
	landingBody="Every path it changes against main is a test, testdata, review evidence, a shard table or ruled harness,
so it lands with no gate (Kirk, Oct 8); Loom's next whole-suite run of main is its check."
	echo "Landing test-only ${gated:0:8} over main ${old:0:8}."
	fi
elif git merge-base --is-ancestor "$old" "$gated"; then
	tree=$(git rev-parse "${gated}^{tree}")
	landingSubject="Land ${gated:0:8} over main ${old:0:8}"
	landingBody="The tree is ${gated:0:8}'s, as gated."
	echo "Landing ${gated:0:8} over main ${old:0:8}."
else
	# Ruled by @system_adamic, October 7 and October 9: a green candidate lands over a main that moved only
	# by record and test-only commits (the old velocity table, meter runs, a progress.json nothing in the
	# landing reads, and tests outside the candidate's code packages). The landing's tree is the gated
	# tree plus exactly those paths from main; any other commit on main means merge and gate again. A
	# candidate built ahead on a gated sha lands this way over that sha's landing commit, whose tree is
	# the same as the sha's.
	# git computes the merge itself (merge-tree, no worktree), so a stack that already holds main's
	# older commits, or two merge bases, is judged by the tree it would really produce.
	if ! tree=$(git merge-tree --write-tree "$old" "$gated" | head -n 1) || [ -z "$tree" ]; then
		echo "refused: ${gated:0:8} doesn't merge cleanly with main ${old:0:8}; merge it in and gate again" >&2
		exit 1
	fi
	recordPaths=$(git diff --name-only "$gated" "$tree")
	notRecords=$(printf '%s\n' "$recordPaths" | countingPaths "$(codePackages "$(git merge-base "$old" "$gated")" "$gated")")
	if [ -n "$notRecords" ]; then
		# B, the moved main merged with the gated sha (@system_adamic, Oct 9 02:52, #mbexftz): it descends from the
		# gate, so a rerun of only the units whose input hash the move changed can land it on the gate's kept
		# verdicts (rerun_merge.py). Until that rerun path is wired, B is named and the landing refused.
		moved=$(git commit-tree "$tree" -p "$old" -p "$gated" -m "Moved main: ${gated:0:8} merged onto main ${old:0:8}, for a rerun of the units the move changed")
		echo "refused: main moved to ${old:0:8} under this gate, beyond record and test-only commits ($(printf '%s' "$notRecords" | head -n 3 | paste -sd ' ' -)); merge it in and gate again, or rerun the units the move changed on B ${moved}" >&2
		exit 1
	fi
	if printf '%s\n' "$recordPaths" | grep -qx 'stage3/progress.json' && git grep -q 'progress\.json' "$gated" -- '*.go' '*.py' '*.sh' '*.mjs' '*.cjs' '*.js' '*.ts' '*.a'; then
		echo "refused: main's stage3/progress.json differs from ${gated:0:8}'s and ${gated:0:8} reads it; merge it in and gate again" >&2
		exit 1
	fi
	landingSubject="Land ${gated:0:8} over main ${old:0:8}, which moved only by record and test-only commits"
	landingBody="The tree is ${gated:0:8}'s, as gated, plus $(printf '%s\n' "$recordPaths" | grep -c . || true) record and test-only paths from main."
	echo "Landing ${gated:0:8} over main ${old:0:8}, which moved only by record and test-only commits."
fi

# The checks that take seconds (@system_adamic, Oct 9 03:51Z), on every landing (developer tools: the fast
# gate runs cmd/adamic-gate's t.Parallel analyzer only when a change reaches that package): gofmt, declared
# tools, the analyzer and a bounded go vet, so a violator is refused here instead of reddening main.
laneCommit=$(git commit-tree "$tree" -p "$old" -p "$gated" -m "lane checks for ${gated:0:8}")
if ! laneChecks=$(python3 "$directory/lane-checks.py" "$old" "$tree" "$laneCommit" 2>&1); then
	echo "refused: the lane's checks: $(printf '%s' "$laneChecks" | head -n 5 | paste -sd ' ' -)" >&2
	exit 1
fi
branches="${branches}; ${laneChecks}"

# One commit per landing, first parent the old main and second the landed sha, its tree exactly the
# landing's, its numbers as trailers (@system_adamic, Oct 9, from Kirk: the velocity table left main).
# `git log --first-parent` over these is the velocity table; landings.py prints it as CSV.
commitsLanded=$(($(git rev-list --count "${old}..${gated}") + 1))
backlog=$(git rev-list --count --no-merges --remotes=origin "^${gated}")
landing=$(git commit-tree "$tree" -p "$old" -p "$gated" -F - <<MESSAGE
${landingSubject}

${landingBody}

Old-main: ${old}
Landed-commits: ${commitsLanded}
Branches: $(printf '%s' "$branches" | tr '\n' ' ')
Gate-minutes: ${gateMinutes}
Pass: ${pass}
Fail: ${fail}
Skip: ${skip}
Backlog: ${backlog}${extraTrailers:+
${extraTrailers}}
MESSAGE
)
if [ "$(git rev-parse "${landing}^{tree}")" != "$tree" ]; then
	echo "refused: the landing commit's tree isn't the landing's" >&2
	exit 1
fi
git push origin "${landing}:refs/heads/main"

echo "Pushed main ${old:0:8}..${landing:0:8}, ${commitsLanded} commits (${branches}), gate ${pass} pass / ${fail} fail / ${skip} skip, ${gateMinutes} minutes; its numbers are ${landing:0:8}'s trailers; backlog ${backlog} commits."
"$directory/merge-back.sh"
