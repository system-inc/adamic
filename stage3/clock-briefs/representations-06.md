a value of type string | null

Reserved for the hourly contribution clock. The representations worker will not implement this kind tonight. 1 listed root sites. Compiler SHA 3cead3fdfeaf8dca6920d748f016a0b0ef1823f2 includes checked non-null merge c41c0e062e99da37820f822968d4df1b48cdaee7.

Census reproducer

Use the exact adapted source bytes. To reconstruct this session's input:
    git fetch origin codex/stage3-latent-full
    mkdir -p /tmp/representation-clock-input
    git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3 | tar -x -C /tmp/representation-clock-input
    bash /tmp/representation-clock-input/stage3/apply.sh /tmp/notyet-representations-adapted
The original manifest has 81 sources; this worker verified all 81 byte lengths and SHA-256 hashes before replay. Run the guarded replay command from the Adamic repository, with its setup environment sourced.

    go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:52 -kind NotYet -reason 'a value of type string | null' > /tmp/clock-string_null.json 2> /tmp/clock-string_null.log

The equivalent prebuilt guarded replay worker exited 0 on this SHA. Its stderr, verbatim:
replay load/register/lower: 1.896942859s
reproduced NotYet: a value of type string | null at /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:52

Selected stdout lowering finding:
{
  "measurement": "measured on a checker-rejected entry-root program",
  "unit": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:5",
  "phase": "lowering",
  "kind": "NotYet",
  "where": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:52",
  "reason": "a value of type string | null",
  "text": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:52: stage 0 can't lower a value of type string | null yet"
}

Minimal program for internal/oracle/testdata/representation_clock_string_null.a:

function probe(value: string | null): void {
    console.log(`${typeof value}:${value === null}:${value ?? 'fallback'}`);
}
probe('text');
probe('');
probe(null);

Node command:
    node --input-type=module-typescript < internal/oracle/testdata/representation_clock_string_null.a
Node stdout, verbatim (exit 0, empty stderr):
string:false:text
string:false:
object:true:fallback

Production compiler observation on the same minimal program:
    go run ./cmd/adamic c internal/oracle/testdata/representation_clock_string_null.a
Exit 1. Observed stderr:
adamic: /tmp/notyet-representations-clock/string_null.a:1:16: stage 0 can't lower a value of type string | null yet
exit status 1

Where to fix

internal/lower/expression.go:30 representation, nullable-union refusal at 83-93; typeOf at 15. Existing typeOfNull in internal/lower/typeof.go:10 records what a missing pointer means.
Packages: Expected internal/lower and internal/oracle; if existing string/null storage is insufficient, also test the IR/native/JavaScript packages actually changed.

This is a two-state case and has no undefined member. A null reference may already encode null correctly if typeof carries the declared null interpretation. Prove that before introducing a wider tag protocol. Do not enable string | null | undefined accidentally. Keep empty string present, strict null comparison correct, ?? lazy, and stale narrowing checks intact. No unchecked assertion or type erasure is allowed.

Fixture and registration

Create internal/oracle/testdata/representation_clock_string_null.a and its own internal/oracle/representation_clock_string_null_test.go. Register the fixture in that file; do not edit internal/oracle/oracle_test.go. For census-only parameter cases, also add a direct checked-type representation test in a new internal/lower clock-specific _test.go. Existing oracle registration and mutant examples are in internal/oracle/representation_generic_state_test.go and representation_scalar_constraint_test.go.

Done when

The registered fixture agrees with Node in the JavaScript backend and sanitized native backend, including leak checks. Add and run this named mutant: clock-string-null-typeof-undefined: clear the null interpretation on the emitted typeof observation. Node's object:true:fallback line becomes undefined:true:fallback and stdout must kill the mutant. Every new admission guard also needs a mutant that demonstrably fails its targeted test. Replay the exact site again: its representation stop is gone, or list the next named stop; if it is a census echo, retain evidence and cancel the kind explicitly. Run focused tests of every package touched. Refresh counts with:
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/clock-string_null-counts.log 2>&1
Write all test output directly to log files. Name every file outside representation in the commit message, name any new separate runtime helper in the report, commit on your own branch and push.

Ownership

Before editing a lowering helper, check git log --remotes=origin/codex/notyet-* -- <file>. Another worker's internal/lower function stays theirs. Keep this contribution limited to the reserved kind; do not implement the other five clock briefs or the parent worker's remaining list.
