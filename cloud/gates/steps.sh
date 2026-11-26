# cloud/gates/steps.sh: the steps every gate recipe (cloud/gates/<gate>.sh) is made of, sourced after the image's
# environment and cloud/sandboxed-job.sh. A recipe is the whole command of a Loom sealed job, run at the head's own sha;
# cloud/submit-gates.py matches it by its sha256 against the job's commandSha256 and reads its last line, so each step
# prints PASS or FAIL, and gate_end always ends the job on one line:
#
#   VERDICT PASS <gate> head <sha> base <sha|none>
#   VERDICT FAIL <gate> head <sha> base <sha|none> failed: <step> ...
#
# The base is HEAD^, read from the tree's own history: a head is one commit on main, so its parent is the base its paired
# gates compare against, and submit-gates.py holds that base to main's tip by the map's fresh rule. Logs go to the job's
# scratch, outside the tree; only summaries reach stdout, so the kept tail holds the verdict.

gateFailed=()
gateBase=none

# gate_begin <gate>: names the gate, reads the head, and makes the gate's scratch.
gate_begin() {
	gateName=$1
	gateHead=$(git rev-parse --verify -q HEAD) || {
		gateHead=unknown
		gate_abort head "the laid tree has no HEAD to read"
	}
	gateScratch=$(mktemp -d "${LOOM_JOB_SCRATCH:-$TMPDIR}/gate-$gateName.XXXXXX") && gateScratch=$(cd "$gateScratch" && pwd -P) ||
		gate_abort scratch "couldn't make the gate's scratch"
	case "$gateScratch/" in "$(pwd -P)"/*) gate_abort scratch "$gateScratch is inside the tree" ;; esac
	printf 'gate %s head %s scratch %s\n' "$gateName" "$gateHead" "$gateScratch"
}

# gate_base: the head's parent, which must be in the laid history.
gate_base() {
	gateBase=$(git rev-parse --verify -q 'HEAD^') || {
		gateBase=none
		gate_abort base "the laid history holds no HEAD^, so the head's base can't be read"
	}
	printf 'base %s\n' "$gateBase"
}

gate_pass() { printf 'PASS %s\n' "$*"; }

# gate_fail <step> <words>: one failed step; the gate goes on to its next step and fails at its end.
gate_fail() {
	local step=$1
	shift
	printf 'FAIL %s: %s\n' "$step" "$*"
	gateFailed+=("$step")
}

# gate_abort <step> <words>: a failed step nothing after can stand without.
gate_abort() {
	gate_fail "$@"
	gate_end
}

gate_end() {
	if [ ${#gateFailed[@]} -eq 0 ]; then
		printf 'VERDICT PASS %s head %s base %s\n' "$gateName" "$gateHead" "$gateBase"
		exit 0
	fi
	printf 'VERDICT FAIL %s head %s base %s failed: %s\n' "$gateName" "$gateHead" "$gateBase" "${gateFailed[*]}"
	exit 1
}

# gate_preamble <step> <command...>: a sandboxed-job.sh step whose output goes to a log, its own lines shown.
gate_preamble() {
	local step=$1 began=$SECONDS
	shift
	if "$@" > "$gateScratch/$step.log" 2>&1; then
		grep '^sandboxed-job:' "$gateScratch/$step.log"
		gate_pass "$step ($((SECONDS - began)) s)"
	else
		tail -40 "$gateScratch/$step.log"
		gate_abort "$step" "$*"
	fi
}

# gate_products <package...>: the packages' products settled ahead, as a unit test reads them.
gate_products() {
	gate_preamble products sandboxed_products "$@"
}

# gate_go_test <step> <package...>: go test of the packages whole, its events to a log, and a line per package and per
# failing test. Environment a gate needs (ADAMIC_TEST_WASI=1) is set by its caller, in front of the call.
gate_go_test() {
	local step=$1 began=$SECONDS rc=0 summary=0
	shift
	go test -count=1 -timeout 60m -json "$@" > "$gateScratch/$step.jsonl" 2> "$gateScratch/$step.stderr" || rc=$?
	printf '%s: go test %s exit=%s wall_seconds=%s\n' "$step" "$*" "$rc" "$((SECONDS - began))"
	python3 -I - "$gateScratch/$step.jsonl" << 'PYTHON' || summary=$?
import json, sys
results, outputs, packages = {}, {}, {}
for line in open(sys.argv[1], errors='replace'):
    try:
        event = json.loads(line)
    except ValueError:
        continue
    package, test, action = event.get('Package', ''), event.get('Test'), event.get('Action')
    if action == 'output' and test:
        outputs.setdefault((package, test), []).append(event.get('Output', ''))
    if action in ('pass', 'fail', 'skip'):
        if test is None:
            packages[package] = action
        elif '/' not in test:
            results[(package, test)] = action
passed = bool(packages)
for package, action in sorted(packages.items()):
    counted = [a for (p, t), a in results.items() if p == package]
    passed = passed and (action == 'pass' or (action == 'skip' and not counted))
    print(f'  {action} {package}: {counted.count("pass")} pass, {counted.count("fail")} fail, {counted.count("skip")} skip')
failed = sorted(key for key, action in results.items() if action == 'fail')
for package, test in failed[:10]:
    print(f'  --- FAIL {package} {test}, last 30 lines:')
    print(''.join(outputs.get((package, test), [])[-30:]).rstrip())
sys.exit(0 if passed and not failed else 1)
PYTHON
	if [ "$rc" -eq 0 ] && [ "$summary" -eq 0 ]; then
		gate_pass "$step"
	else
		tail -20 "$gateScratch/$step.stderr"
		gate_fail "$step" "go test $* (exit $rc, summary $summary)"
	fi
}
