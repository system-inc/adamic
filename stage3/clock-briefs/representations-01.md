a value of type NonNullable<T>

Reserved for the hourly contribution clock. The representations worker will not implement this kind tonight. 7 listed root sites. Compiler SHA 3cead3fdfeaf8dca6920d748f016a0b0ef1823f2 includes checked non-null merge c41c0e062e99da37820f822968d4df1b48cdaee7.

Census reproducer

Use the exact adapted source bytes. To reconstruct this session's input:
    git fetch origin codex/stage3-latent-full
    mkdir -p /tmp/representation-clock-input
    git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3 | tar -x -C /tmp/representation-clock-input
    bash /tmp/representation-clock-input/stage3/apply.sh /tmp/notyet-representations-adapted
The original manifest has 81 sources; this worker verified all 81 byte lengths and SHA-256 hashes before replay. Run the guarded replay command from the Adamic repository, with its setup environment sourced.

    go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/core.ts:246:19 -kind NotYet -reason 'a value of type NonNullable<T>' > /tmp/clock-nonnullable_t.json 2> /tmp/clock-nonnullable_t.log

The equivalent prebuilt guarded replay worker exited 0 on this SHA. Its stderr, verbatim:
replay load/register/lower: 1.814830191s
reproduced NotYet: a value of type NonNullable<T> at /tmp/notyet-representations-adapted/src/compiler/core.ts:246:19

Selected stdout lowering finding:
{
  "measurement": "measured on a checker-rejected entry-root program",
  "unit": "/tmp/notyet-representations-adapted/src/compiler/core.ts:242:1",
  "phase": "lowering",
  "kind": "NotYet",
  "where": "/tmp/notyet-representations-adapted/src/compiler/core.ts:246:19",
  "reason": "a value of type NonNullable<T>",
  "text": "/tmp/notyet-representations-adapted/src/compiler/core.ts:246:19: stage 0 can't lower a value of type NonNullable<T> yet"
}

Minimal program for internal/oracle/testdata/representation_clock_nonnullable_t.a:

function keep<T>(value: NonNullable<T>): NonNullable<T> { return value; }
const node = { name: 'node' };
console.log(keep<string | undefined>('text'));
console.log(`${keep<{ readonly name: string } | undefined>(node) === node}`);

Node command:
    node --input-type=module-typescript < internal/oracle/testdata/representation_clock_nonnullable_t.a
Node stdout, verbatim (exit 0, empty stderr):
text
true

Production compiler observation on the same minimal program:
    go run ./cmd/adamic c internal/oracle/testdata/representation_clock_nonnullable_t.a
Exit 1. Observed stderr:
adamic: /tmp/notyet-representations-clock/nonnullable_t.a:1:10: stage 0 can't lower a function returning NonNullable<T> yet
exit status 1

Where to fix

internal/lower/expression.go:30 representation, intersection branch at 48; typeOf at 15. The minimal program also stops at the generic return representation requested by internal/lower/functions.go:80. Fix the NonNullable<T> intersection case in representation and its own helper; functions.go is owned by the generic-return worker, so do not edit it without coordination.
Packages: internal/lower and internal/oracle; IR or backend packages only if the chosen representation needs them.

NonNullable<T> is T intersected with the empty non-nullish structural type. A concrete substitution must determine storage; empty structural views can still contain primitive values. Do not treat every T & {} as a plain object or erase an unsubstituted T. Test concrete string and object cases and a direct checked-type representation probe. If only an isolated generic context remains, document a census-echo cancellation rather than inventing an ABI.

Fixture and registration

Create internal/oracle/testdata/representation_clock_nonnullable_t.a and its own internal/oracle/representation_clock_nonnullable_t_test.go. Register the fixture in that file; do not edit internal/oracle/oracle_test.go. For census-only parameter cases, also add a direct checked-type representation test in a new internal/lower clock-specific _test.go. Existing oracle registration and mutant examples are in internal/oracle/representation_generic_state_test.go and representation_scalar_constraint_test.go.

Done when

The registered fixture agrees with Node in the JavaScript backend and sanitized native backend, including leak checks. Add and run this named mutant: clock-nonnullable-t-drop-object-return: replace only the object specialization's return with an undefined object. The original Node identity line is true; the sanitized mutant must finish cleanly and print false. Every new admission guard also needs a mutant that demonstrably fails its targeted test. Replay the exact site again: its representation stop is gone, or list the next named stop; if it is a census echo, retain evidence and cancel the kind explicitly. Run focused tests of every package touched. Refresh counts with:
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/clock-nonnullable_t-counts.log 2>&1
Write all test output directly to log files. Name every file outside representation in the commit message, name any new separate runtime helper in the report, commit on your own branch and push.

Ownership

Before editing a lowering helper, check git log --remotes=origin/codex/notyet-* -- <file>. Another worker's internal/lower function stays theirs. Keep this contribution limited to the reserved kind; do not implement the other five clock briefs or the parent worker's remaining list.
