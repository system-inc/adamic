Clock-base: origin/area/compiler
Clock-owner: system_adamic_compiler
a value of type Child

Reserved for the hourly contribution clock. The representations worker will not implement this kind tonight. 1 listed root sites. Compiler SHA 3cead3fdfeaf8dca6920d748f016a0b0ef1823f2 includes checked non-null merge c41c0e062e99da37820f822968d4df1b48cdaee7.

Census reproducer

Use the exact adapted source bytes. To reconstruct this session's input:
    git fetch origin codex/stage3-latent-full
    mkdir -p /tmp/representation-clock-input
    git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3 | tar -x -C /tmp/representation-clock-input
    bash /tmp/representation-clock-input/stage3/apply.sh /tmp/notyet-representations-adapted
The original manifest has 81 sources; this worker verified all 81 byte lengths and SHA-256 hashes before replay. Run the guarded replay command from the Adamic repository, with its setup environment sourced.

    go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/emitter.ts:4754:19 -kind NotYet -reason 'a value of type Child' > /tmp/clock-child.json 2> /tmp/clock-child.log

The equivalent prebuilt guarded replay worker exited 0 on this SHA. Its stderr, verbatim:
replay load/register/lower: 2.010598255s
reproduced NotYet: a value of type Child at /tmp/notyet-representations-adapted/src/compiler/emitter.ts:4754:19

Selected stdout lowering finding:
{
  "measurement": "measured on a checker-rejected entry-root program",
  "unit": "/tmp/notyet-representations-adapted/src/compiler/emitter.ts:4729:5",
  "phase": "lowering",
  "kind": "NotYet",
  "where": "/tmp/notyet-representations-adapted/src/compiler/emitter.ts:4754:19",
  "reason": "a value of type Child",
  "text": "/tmp/notyet-representations-adapted/src/compiler/emitter.ts:4754:19: stage 0 can't lower a value of type Child yet"
}

Minimal program for internal/oracle/testdata/representation_clock_child.a:

interface NodeView { readonly kind: number; readonly text: string; }
function render<Child extends NodeView>(child: Child): string { return `${child.kind}:${child.text}`; }
console.log(render({ kind: 1, text: 'leaf' }));

Node command:
    node --input-type=module-typescript < internal/oracle/testdata/representation_clock_child.a
Node stdout, verbatim (exit 0, empty stderr):
1:leaf

Production compiler observation on the same minimal program:
    go run ./cmd/adamic c internal/oracle/testdata/representation_clock_child.a
Exit 0. C generation succeeds for the concrete instantiations; this is not proof that the isolated generic census case is fixed.

Where to fix

internal/lower/expression.go:30 representation, unsubstituted type parameter branch at 36; typeOf at 15. Child extends Node in emitNodeListItems; do not modify that other worker's lowering function or the compiler source.
Packages: internal/lower and internal/oracle; IR/backend tests only if an additional representation is genuinely needed.

The concrete reduction already compiles on the pushed SHA. The exact census stop is Child without a concrete substitution. A structural constraint alone does not always select a header: length-bearing views can hide arrays or strings and callable views can hide closures. Prove the Node-shaped case has one admissible ABI, or keep the refusal and explain a census-context ruling. Add a checked-type probe that exercises this unresolved constraint. Keep the generic identity/dataflow contract.

Fixture and registration

Create internal/oracle/testdata/representation_clock_child.a and its own internal/oracle/representation_clock_child_test.go. Register the fixture in that file; do not edit internal/oracle/oracle_test.go. For census-only parameter cases, also add a direct checked-type representation test in a new internal/lower clock-specific _test.go. Existing oracle registration and mutant examples are in internal/oracle/representation_generic_state_test.go and representation_scalar_constraint_test.go.

Done when

The registered fixture agrees with Node in the JavaScript backend and sanitized native backend, including leak checks. Add and run this named mutant: clock-child-ignore-argument: replace the render specialization result with the string mutant. Node prints 1:leaf; the clean sanitized mutant must disagree in stdout. Every new admission guard also needs a mutant that demonstrably fails its targeted test. Replay the exact site again: its representation stop is gone, or list the next named stop; if it is a census echo, retain evidence and cancel the kind explicitly. Run focused tests of every package touched. Refresh counts with:
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/clock-child-counts.log 2>&1
Write all test output directly to log files. Name every file outside representation in the commit message, name any new separate runtime helper in the report, commit on your own branch and push.

Ownership

Before editing a lowering helper, check git log --remotes=origin/codex/notyet-* -- <file>. Another worker's internal/lower function stays theirs. Keep this contribution limited to the reserved kind; do not implement the other five clock briefs or the parent worker's remaining list.
