#!/usr/bin/env bash
# Proof for runtime-performance/setblocking-receiver (#0bxc4jh): three heads on Library's 83ea288d.
# 1, Lower's: a narrowed call through an optional field, if (o.m) o.m(), keeps the method road
# (ir.MethodCallee), with the Defined's own check. 2, Runtime's and Library's: the stdout handle's
# setBlocking slot holds a static closure that throws TypeError: Illegal invocation, and the method
# road's call site blocks off the handle alone (node_process.c, emit_functions.go, javascript). 3, the
# fixtures. Nothing of this ran on the Mac but gofmt, clang -fsyntax-only on node_process.c, Node on
# the fixtures and the JavaScript runtime pieces (adamic's CLAUDE.md: no Go builds or tests there),
# so this script is the whole check. Run from the worktree root on a box. Every run logs to a file,
# then the log is grepped. The mutants are applied and undone by exact text replacement (python),
# never git, and each is checked byte for byte against a copy after undo.
#
# DRY=1 checks only that every mutant's anchor is in its file exactly once, and runs nothing else.
set -u
source /workspace/adamic-tools/env.sh 2>/dev/null || true
dry=${DRY:-}

# replace FILE OLD NEW: exactly one OLD in FILE becomes NEW, or it stops.
replace() {
  python3 - "$1" "$2" "$3" <<'PY'
import sys
path, old, new = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(path).read()
if text.count(old) != 1:
    sys.exit("expected one %r in %s, found %d" % (old, path, text.count(old)))
open(path, "w").write(text.replace(old, new))
PY
}

# anchored FILE OLD: whether OLD is in FILE exactly once (DRY's check).
anchored() {
  python3 - "$1" "$2" <<'PY'
import sys
path, old = sys.argv[1], sys.argv[2]
count = open(path).read().count(old)
print("anchor in %s: %d %s" % (path, count, "ok" if count == 1 else "WRONG"))
sys.exit(0 if count == 1 else 1)
PY
}

if [ -z "$dry" ]; then
source ~/.adamic-tools/env.sh 2>/dev/null || source ~/adamic-tools/env.sh 2>/dev/null || true
export TMPDIR=$HOME/scratch/system_adamic_runtime_performance/tmp; mkdir -p "$TMPDIR"
git log --oneline -1 | cut -c1-12
want=$(git ls-tree HEAD cohere | awk '{print $3}')
found=""
for d in ~/scratch/*/cohere ~/scratch/*/*/cohere ~/*/cohere; do
  [ -f "$d/TypeScript/tsc/go.mod" ] || continue
  have=$(git -C "$d" rev-parse HEAD 2>/dev/null)
  [ "$have" = "$want" ] || continue
  tswant=$(git -C "$d" ls-tree HEAD TypeScript | awk '{print $3}')
  tshave=$(git -C "$d/TypeScript" rev-parse HEAD 2>/dev/null)
  [ "$tswant" = "$tshave" ] || continue
  [ -z "$(git -C "$d" status --porcelain 2>/dev/null | head -1)" ] || continue
  found=$d; break
done
[ -z "$found" ] && [ -f cohere/TypeScript/tsc/go.mod ] && found=$(pwd)/cohere
if [ -z "$found" ]; then echo "cohere: none at $want"; else rmdir cohere 2>/dev/null; ln -sfn "$found" cohere; echo "cohere: $found at $want"; fi
[ -f cohere/TypeScript/tsc/go.mod ] || { echo "no cohere, stopping"; echo "== end"; exit 0; }
go version
node --version
fi

logs=$(mktemp -d "${TMPDIR:-/tmp}/setblocking-proof.XXXXXX")
echo "logs: $logs"
summary() { grep -E "(^(    )?--- FAIL|^FAIL|^ok|^panic|--- PASS: TestNodeProcess[A-Za-z]*(/[a-z_]+\.a)?( |$)|--- PASS: TestEvery[A-Za-z]*|--- PASS: TestCallTargetReaders[A-Za-z]*|--- PASS: TestMethodCallee[A-Za-z]*|--- PASS: TestNarrowedMethod[A-Za-z]*|--- PASS: TestRuntime(Statics|FieldLayouts)[A-Za-z]*( |$)|--- PASS: TestHostShape[A-Za-z]*( |$)|--- PASS: TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_|--- PASS: TestReflectGetStopsAlikeOnBothBackends/stop_host_handle)" "$1" | head -80; }

if [ -z "$dry" ]; then
gofmt -l cmd internal stage3/census > "$logs/gofmt.log" 2>&1; echo "gofmt=$? files:"; cat "$logs/gofmt.log"
go build -buildvcs=false -o "$logs/adamic" ./cmd/adamic > "$logs/build.log" 2>&1; echo "build=$?"; tail -5 "$logs/build.log"
go vet ./internal/ir/ ./internal/javascript/ ./internal/native/ ./internal/oracle/ ./internal/lower/ ./internal/flow/ ./internal/fresh/ ./stage3/census/latent/statecopy/ ./cmd/adamic-admission-delta/ > "$logs/vet.log" 2>&1; echo "vet=$?"; tail -5 "$logs/vet.log"
# ir whole: MethodCallee's own tests, TestEveryCalleeFieldReadGoesThroughMethodCallee (the scan that goes
# red on a pass asserting a callee to a Property without MethodCallee) and TestCallTargetReaders* (no new
# CallClosure.Closure read outside the approved functions).
go test -count=1 -timeout 30m -v ./internal/ir/ > "$logs/ir.log" 2>&1; echo "ir=$?"; summary "$logs/ir.log"
grep -E -- "--- (PASS|FAIL): (TestEveryCalleeFieldReadGoesThroughMethodCallee|TestMethodCalleeSeesTheNarrowedField|TestNarrowedMethodCallKeepsTheMethodArm|TestCallTargetReaders[A-Za-z]+)( |$)" "$logs/ir.log"
go test -count=1 -timeout 30m ./internal/javascript/ > "$logs/javascript.log" 2>&1; echo "javascript=$?"; summary "$logs/javascript.log"
# lower whole: callClosure, checkedFunctionTypeCall, nodeProcessStorageABI, the view pass, and
# TestEveryOperandIsClassedForViews.
go test -count=1 -timeout 30m -v ./internal/lower/ > "$logs/lower.log" 2>&1; echo "lower=$?"; summary "$logs/lower.log"
# flow and fresh whole, with the IR classification tests (TestEveryExpressionIsClassified,
# TestEveryStatementIsClassified) and their corpora, which lower every program at the top of testdata,
# the seven new handle fixtures among them.
go test -count=1 -timeout 30m -v ./internal/flow/ ./internal/fresh/ > "$logs/flow.log" 2>&1; echo "flow+fresh=$?"; summary "$logs/flow.log"
grep -E -- "--- (PASS|FAIL): TestEvery[A-Za-z]+( |$)" "$logs/flow.log" "$logs/lower.log"
go test -count=1 -timeout 20m ./stage3/census/latent/statecopy/ > "$logs/statecopy.log" 2>&1; echo "statecopy=$?"; summary "$logs/statecopy.log"
# native: the statics row for set_blocking_closure, layouts, host shapes and the product tests.
go test -count=1 -timeout 30m -v -run 'TestRuntimeStatics|TestRuntimeFieldLayouts|TestHostShape|TestHostGuards|TestProduct_' ./internal/native/ > "$logs/native.log" 2>&1; echo "native=$?"; summary "$logs/native.log"
# wasm32's runtime products build under -Werror: node_process.c's new closure and its two functions on
# the WASI side of the #ifdef.
sr=$(ls -d ~/.adamic-tools/wasi-sdk/share/wasi-sysroot ~/adamic-tools/wasi-sdk/share/wasi-sysroot 2>/dev/null | head -1); echo "wasi sysroot: ${sr:-none}"
ADAMIC_TEST_WASI=1 WASI_SYSROOT=$sr go test -count=1 -timeout 30m -v -run 'TestProduct_.*WASI' ./internal/native/ > "$logs/wasi-products.log" 2>&1; echo "wasi-products=$?"; summary "$logs/wasi-products.log"; grep -c -- "--- PASS: TestProduct_" "$logs/wasi-products.log"; grep -m3 "error:" "$logs/wasi-products.log"
# The oracle: every TestNodeProcess test (the handle view table with the seven new fixtures and tsc's
# guard, the terminal, the blocking probe, the process mutants, the detached stop), stop_host_handle.a,
# the review programs of head 1, and the censuses that read every fixture.
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -v -run '^TestReviewProgramsAgreeWithNode_[0-9]+$/^narrowed_method_' ./internal/oracle/ > "$logs/review.log" 2>&1; echo "review=$?"; summary "$logs/review.log"
grep -E -- "--- (PASS|FAIL|SKIP): TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_[a-z_]+\.a" "$logs/review.log"
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -v -run 'TestRuntimeSetsAreEachUsed|TestFamiliesOwnEveryCaseOnce|TestReviewProgramsNoLooseFiles|TestCountsFilesCensus' ./internal/oracle/ > "$logs/census.log" 2>&1; echo "census=$?"; summary "$logs/census.log"
# node_process_terminal.a's counts row: its guarded call took the method road in head 1, so the row may
# move; a red here is a row owed, measured with -update-counts on a box.
go test -count=1 -timeout 20m ./cmd/adamic-admission-delta/ > "$logs/delta.log" 2>&1; echo "delta=$?"; summary "$logs/delta.log"

# Runtime's question, on the built compiler: a program's own method called through a narrowed optional
# field. Each probe prints 3 on Node. The guard-read forms should stop (refused or not yet); the in
# forms should lower and print 3 natively and on JavaScript. class_in.a is the class instance's in
# form, kept out of the fixtures until its native in test is seen to answer for a class's method.
probes=$logs/probes; mkdir -p "$probes"
cat > "$probes/literal_signature.a" <<'EOF'
interface I {
	x: number;
	m?(): number;
}
const o: I = { x: 3, m() { return this.x; } };
if (o.m) console.log(String(o.m()));
EOF
cat > "$probes/literal_field.a" <<'EOF'
interface I {
	x: number;
	m?: () => number;
}
const o: I = { x: 3, m() { return this.x; } };
if (o.m) console.log(String(o.m()));
EOF
cat > "$probes/class_signature.a" <<'EOF'
interface I {
	x: number;
	m?(): number;
}
class C {
	x = 3;
	m(): number {
		return this.x;
	}
}
const o: I = new C();
if (o.m) console.log(String(o.m()));
EOF
cat > "$probes/class_field.a" <<'EOF'
interface I {
	x: number;
	m?: () => number;
}
class C {
	x = 3;
	m(): number {
		return this.x;
	}
}
const o: I = new C();
if (o.m) console.log(String(o.m()));
EOF
cat > "$probes/compared_field.a" <<'EOF'
interface I {
	x: number;
	m?: () => number;
}
function show(o: I): void {
	if (o.m !== undefined) console.log(String(o.m()));
}
show({ x: 3, m() { return this.x; } });
EOF
cat > "$probes/literal_in.a" <<'EOF'
interface I {
	x: number;
	m?(): number;
}
function show(o: I): void {
	if ('m' in o) console.log(String(o.m()));
}
show({ x: 3, m() { return this.x; } });
EOF
cat > "$probes/class_in.a" <<'EOF'
interface I {
	x: number;
	m?(): number;
}
class C {
	x = 3;
	m(): number {
		return this.x;
	}
}
function show(o: I): void {
	if ('m' in o) console.log(String(o.m()));
}
show(new C());
EOF
for probe in "$probes"/*.a; do
  name=$(basename "$probe" .a)
  node --disable-warning=ExperimentalWarning oracle/node.mjs "$probe" > "$probes/$name.node" 2>&1
  if "$logs/adamic" c "$probe" > "$probes/$name.c" 2> "$probes/$name.stop"; then
    "$logs/adamic" build "$probe" -o "$probes/$name.bin" --sanitize > "$probes/$name.build" 2>&1
    "$probes/$name.bin" > "$probes/$name.native" 2>&1; native=$?
    "$logs/adamic" js "$probe" > "$probes/$name.mjs" 2> "$probes/$name.jsstop"
    node --disable-warning=ExperimentalWarning oracle/node.mjs "$probes/$name.mjs" > "$probes/$name.js" 2>&1; js=$?
    echo "probe $name: lowers; Node $(tr '\n' '|' < "$probes/$name.node"); native exit $native $(tr '\n' '|' < "$probes/$name.native" | cut -c1-200); JavaScript exit $js $(tr '\n' '|' < "$probes/$name.js" | cut -c1-200)"
  else
    echo "probe $name: stops; Node $(tr '\n' '|' < "$probes/$name.node"); $(head -2 "$probes/$name.stop" | tr '\n' ' ' | cut -c1-300)"
  fi
done
fi

# verdict MUTANT LOG PATTERN...: each test (an extended regular expression, as a family's shard number
# isn't known ahead) the mutant must turn red, CAUGHT by it or SURVIVED it.
verdict() {
  name=$1; log=$2; shift 2
  # A panic ends the whole test binary, so a test it cut short never prints its FAIL.
  grep -q "^panic: " "$log" && echo "mutant $name: the run panicked, so these verdicts mean nothing: $(grep -m1 '^panic: ' "$log")"
  for test in "$@"; do
    grep -Eq -- "--- FAIL: $test( |\$)" "$log" && echo "mutant $name: CAUGHT by $test" || echo "mutant $name: SURVIVED $test"
  done
}

# mutate NAME FILE OLD NEW PACKAGE RUN TEST...: apply, run PACKAGE's RUN tests, undo, and say what
# caught it. Under DRY, only that OLD is in FILE once.
mutate() {
  name=$1; file=$2; original=$3; mutant=$4; package=$5; run=$6; shift 6
  if [ -n "$dry" ]; then anchored "$file" "$original"; return; fi
  cp "$file" "$logs/mutant-$name.before"
  if replace "$file" "$original" "$mutant"; then
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m -v -run "$run" "$package" > "$logs/mutant-$name.log" 2>&1; code=$?
    replace "$file" "$mutant" "$original"
    cmp -s "$file" "$logs/mutant-$name.before" && echo "mutant $name undone" || { cp "$logs/mutant-$name.before" "$file"; echo "mutant $name: undo mismatch, restored from the copy"; }
    echo "mutant $name exit=$code"; summary "$logs/mutant-$name.log"
    grep -E "^\s+[a-z_]+_test.go:[0-9]+: " "$logs/mutant-$name.log" | head -6 | cut -c1-600
    verdict "$name" "$logs/mutant-$name.log" "$@"
  fi
}

oracle=./internal/oracle/
# (a) Runtime's named mutant: the slot's closure blocks, as before this head. The let fixture's
# detached call blocks and prints blocked where Node throws Illegal invocation, on a pipe and a terminal.
# The closure's code itself calls set_blocking, so nothing goes unused and -Werror has nothing to say.
# (b) the receiver test dropped: the handle's closure blocks off any object, so the copy fixture's call
# off another object blocks where Node throws.
# (c) head 1 undone in callClosure: the narrowed call takes the plain road again. tsc's guarded call
# (node_process_handle_view.a) reaches the throwing closure with no receiver and throws where Node
# blocks (c1), and the literal methods called through an in test's narrowing lose this (c2).
mutate c2 internal/lower/expression.go \
  $'\t\t\tproperty.Method = true\n\t\t\tcallee.Value = property\n\t\t\tclosure = callee\n' \
  $'\t\t\t_ = property // mutant c2: head 1 undone\n' \
  $oracle '^TestReviewProgramsAgreeWithNode_[0-9]+$/^narrowed_method_in_' 'TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_in_signature\.a' 'TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_in_field\.a'
# (d) the Defined check's words replaced by the method road's natively: narrowed_method_absent.a's
# native and release legs panic as a compiler bug, not in the pinned narrowing's words.
mutate d internal/native/emit_functions.go \
  ', e.methodPresent(method)), defined.Message)' \
  ', e.methodPresent(method)), "compiler bug: a method the checker proved is there is missing") // mutant d' \
  $oracle '^TestReviewProgramsAgreeWithNode_[0-9]+$/^narrowed_method_absent\.a$' 'TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_absent\.a'
# (e) its JavaScript twin: the narrowed lookup checks nothing, so adamicCallee meets the field's
# undefined and throws a TypeError where the pin is the narrowing's panic.
mutate e internal/javascript/javascript.go \
  '", (value) => " + e.defined(*defined, "value") + "), ["' \
  '", (value) => value), ["' \
  $oracle '^TestReviewProgramsAgreeWithNode_[0-9]+$/^narrowed_method_absent\.a$' 'TestReviewProgramsAgreeWithNode_[0-9]+/narrowed_method_absent\.a'
[ -n "$dry" ] || git status --short --ignore-submodules=all -- internal cmd cloud docs stage3
# The narrowing probes (#6na4x1b): each shape in .a and .ts, natively, sanitized and on the
# JavaScript backend, at this head and at its parent, main, against Node (7 for every one).
P=$TMPDIR/narrowprobes.$$; mkdir -p $P
cat > $P/bang.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	console.log(String(counter.add!(4)));
}
run(literal);
PROBE
cat > $P/bang.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	console.log(String(counter.add!(4)));
}
run(literal);
PROBE
cat > $P/class_instance.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
class Tally {
	total = 3;
	add(amount: number): number {
		return this.total + amount;
	}
}
function run(counter: Counter): void {
	if ('add' in counter) console.log(String(counter.add(4)));
}
run(new Tally());
PROBE
cat > $P/class_instance.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
class Tally {
	total = 3;
	add(amount: number): number {
		return this.total + amount;
	}
}
function run(counter: Counter): void {
	if ('add' in counter) console.log(String(counter.add(4)));
}
run(new Tally());
PROBE
cat > $P/early_return.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (!counter.add) return;
	console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/early_return.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (!counter.add) return;
	console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/in_test.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if ('add' in counter) console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/in_test.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if ('add' in counter) console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/not_undefined.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (counter.add !== undefined) console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/not_undefined.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (counter.add !== undefined) console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/optional_call.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	console.log(String(counter.add?.(4)));
}
run(literal);
PROBE
cat > $P/optional_call.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	console.log(String(counter.add?.(4)));
}
run(literal);
PROBE
cat > $P/typeof_test.a <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (typeof counter.add === 'function') console.log(String(counter.add(4)));
}
run(literal);
PROBE
cat > $P/typeof_test.ts <<'PROBE'
interface Counter {
	total: number;
	add?(amount: number): number;
}
const literal: Counter = {
	total: 3,
	add(amount: number): number {
		return this.total + amount;
	},
};
function run(counter: Counter): void {
	if (typeof counter.add === 'function') console.log(String(counter.add(4)));
}
run(literal);
PROBE
here=$(pwd)
git worktree add --detach $P/base 39953a7ed4fdcb31a090b03138fc9e466f09c9a1~1 > /dev/null 2>&1 && (cd $P/base && rmdir cohere 2>/dev/null; ln -sfn "$found" cohere)
for side in head base; do
  dir=$here; [ $side = base ] && dir=$P/base
  (cd $dir && go build -buildvcs=false -o $P/adamic-$side ./cmd/adamic) > $P/build-$side.log 2>&1 || { echo "$side build failed"; tail -3 $P/build-$side.log; continue; }
  for p in $P/*.a $P/*.ts; do
    n=$(basename $p); line="$side $n:"
    truth=$(cd $here && node --disable-warning=ExperimentalWarning oracle/node.mjs $p 2>&1 | tr '\n' ' ')
    if $P/adamic-$side build $p -o $P/x > $P/b 2>&1; then line="$line native[$($P/x 2>&1 | head -2 | tr '\n' ' ')exit $?]"; else line="$line native-refused[$(grep -o "refuses[^;]*\|NotYet[^;]*\|error TS[0-9]*" $P/b | head -1)]"; fi
    if $P/adamic-$side build $p -o $P/s --sanitize > $P/b 2>&1; then $P/s > $P/so 2>&1; e=$?; line="$line sanitized[$(grep -m1 -o "AddressSanitizer: [a-z-]*\|runtime error: [^:]*" $P/so || head -1 $P/so | tr '\n' ' ')exit $e]"; fi
    if $P/adamic-$side js $p > $P/j.mjs 2> $P/b; then line="$line javascript[$(cd $here && node --disable-warning=ExperimentalWarning oracle/node.mjs $P/j.mjs 2>&1 | head -2 | tr '\n' ' ')]"; fi
    echo "$line node[$truth]"
  done
done
git worktree remove --force $P/base > /dev/null 2>&1
git status --short --ignore-submodules=all -- internal cmd cloud docs stage3
echo "== end"
