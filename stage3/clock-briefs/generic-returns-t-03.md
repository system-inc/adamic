a function returning Declaration & HasModifiers

Reserved hand-off 3 of six. The generic-returns-t worker will not implement this kind. Each brief owns only its named signature case. Keep new helpers in separate files so other clock workers can merge independently. Check the current branch and other clock briefs before editing shared functions.

Measured code revision: ba33e8490905e75588f420c6ccae6cce60947ef4 (contains checked non-null merge c41c0e06).
Root site: src/compiler/checker.ts:6600:18. Raw table: origin/codex/stage3-notyet-table:stage3/notyet-table/roots/raw.csv. This kind has one root site. Deduplicate lowering NotYet observations by kind, where, reason and text, excluding groups whose every observation has blocked_by_kind.

Census replay preparation, from the repository root:
export GOPROXY='https://proxy.golang.org|direct'
bash stage3/apply.sh /tmp/notyet-generic-adapted > /tmp/clock_generic_returns_t_03-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/notyet-clock-replay/overlay > /tmp/clock_generic_returns_t_03-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/notyet-clock-replay/overlay/overlay.json -o /tmp/notyet-clock-replay/worker ./stage3/census/latent/replay/worker > /tmp/clock_generic_returns_t_03-build.log 2>&1

Measured replay command:
/tmp/notyet-clock-replay/worker -project /tmp/notyet-generic-adapted/src/tsc/tsc.ts -where /tmp/notyet-generic-adapted/src/compiler/checker.ts:6600:18 -kind NotYet -reason 'a function returning Declaration & HasModifiers' > /tmp/clock_generic_returns_t_03-replay.json 2> /tmp/clock_generic_returns_t_03-replay.log

Measured output (the matching finding extracted from JSON; replay process exit 0):
kind: NotYet
where: /tmp/notyet-generic-adapted/src/compiler/checker.ts:6600:18
reason: a function returning Declaration & HasModifiers
The same signature stop remains after the checked non-null merge. This is evidence of a lowering stop, not a claim that the complete census unit is otherwise accepted.

Minimal program (write this as internal/oracle/testdata/clock_generic_returns_t_03.a):
interface Declaration { readonly name: string; readonly children: readonly string[]; }
interface HasModifiers { readonly modifiers: readonly string[]; }
function declaration(): Declaration & HasModifiers {
    return { name: ["dec", "laration"].join(""), children: ["child"], modifiers: ["public", "readonly"] };
}
console.log(`${declaration().name}:${declaration().modifiers.join("+")}`);

Node command:
node --input-type=module-typescript < internal/oracle/testdata/clock_generic_returns_t_03.a > /tmp/clock_generic_returns_t_03-node.log 2>&1
Measured Node exit: 0
Measured Node stdout:
declaration:public+readonly

Current Adamic command:
adamic c internal/oracle/testdata/clock_generic_returns_t_03.a > /tmp/clock_generic_returns_t_03-lower.log 2>&1
Measured compiler exit: 1
Measured compiler output (temporary reduction path):
adamic: /tmp/notyet-clock-programs/03.a:3:10: stage 0 can't lower a function returning Declaration & HasModifiers yet

Where to fix on the measured pushed code revision:
internal/lower/functions.go:274, (*lowering).signatureResult; caller at functions.go:80 in (*lowering).lowerFunction. Add a narrow helper in internal/lower/clock_generic_returns_t_03.go and a minimal dispatch from signatureResult. The existing representation entry at internal/lower/expression.go:30 is owned by notyet-representations; coordinate rather than editing its typeOf or representation function. The object-intersection proof in internal/lower/library_object.go is a useful reference, but its scalar-field restriction must not be bypassed without proving the nested fields in this fixture. Use the checker's concrete type before deciding a representation. Preserve every intersection field and readonly property; do not erase type parameters. Reject unsupported callable, indexed or container intersections rather than silently treating them as plain objects.

Expected packages touched: internal/lower and internal/oracle. An optional object result can use the existing object pointer's undefined/null sentinel only after proving this specific intersection is a supported object shape; preserve the distinction from any independently supported null value. Native and JavaScript changes should be unnecessary for the existing object, string-array and undefined representations. If IR, backend or flow changes prove necessary, keep them specific and run focused tests in every touched package. Do not change another worker's lowering function without checking git log origin/codex/notyet-* -- <file>.

Own registration file: internal/oracle/clock_generic_returns_t_03_test.go. Register only internal/oracle/testdata/clock_generic_returns_t_03.a there, using the local oracle fixture-registration pattern. Add a named test TestClockGenericReturnsT03Mutant that proves the mutant below fails its fixture.
Named mutant: clock_generic_returns_t_03_drop_modifiers.
Replace the returned modifiers array with an empty string array.
The mutant must execute cleanly and disagree with Node stdout in native and JavaScript output; a compiler failure or sanitizer crash is not sufficient evidence for this semantic mutant.

Done when:
The fixture matches the recorded Node result in release native, sanitized native (address/undefined/leak checks), and JavaScript backend using the oracle harness.
The named mutant fails the fixture's Node comparison in both backends.
Run focused tests in every package touched, redirecting all test output to log files. Do not run the full package suite or full gate.
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/clock_generic_returns_t_03-counts.log 2>&1 refreshes counts.md successfully.
Replay the original census command again and report whether it now lowers or its next named stop, with the pushed SHA first and the one raw-table root counted conservatively.
Commit with a plain message naming every file outside your owned lowering function, and push only your own branch. New Adamic source files use .a. No existing test should be weakened.
