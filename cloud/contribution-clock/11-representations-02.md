Clock-base: origin/area/compiler
Clock-owner: system_adamic_compiler
a value of type string | null | undefined

Reserved for the hourly contribution clock. The representations worker will not implement this kind tonight. 6 listed root sites. Compiler SHA 3cead3fdfeaf8dca6920d748f016a0b0ef1823f2 includes checked non-null merge c41c0e062e99da37820f822968d4df1b48cdaee7.

Census reproducer

Use the exact adapted source bytes. To reconstruct this session's input:
    git fetch origin codex/stage3-latent-full
    mkdir -p /tmp/representation-clock-input
    git archive 9d534d3a31814f1a192a528e701f6c2ea7c910bc stage3 | tar -x -C /tmp/representation-clock-input
    bash /tmp/representation-clock-input/stage3/apply.sh /tmp/notyet-representations-adapted
The original manifest has 81 sources; this worker verified all 81 byte lengths and SHA-256 hashes before replay. Run the guarded replay command from the Adamic repository, with its setup environment sourced.

    go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:701:24 -kind NotYet -reason 'a value of type string | null | undefined' > /tmp/clock-string_null_undefined.json 2> /tmp/clock-string_null_undefined.log

The equivalent prebuilt guarded replay worker exited 0 on this SHA. Its stderr, verbatim:
replay load/register/lower: 1.837817799s
reproduced NotYet: a value of type string | null | undefined at /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:701:24

Selected stdout lowering finding:
{
  "measurement": "measured on a checker-rejected entry-root program",
  "unit": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:699:1",
  "phase": "lowering",
  "kind": "NotYet",
  "where": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:701:24",
  "reason": "a value of type string | null | undefined",
  "text": "/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:701:24: stage 0 can't lower a value of type string | null | undefined yet"
}

Minimal program for internal/oracle/testdata/representation_clock_string_null_undefined.a:

function probe(value: string | null | undefined): void {
    console.log(`${typeof value}:${value === null}:${value === undefined}:${value ?? 'fallback'}`);
}
probe('text');
probe('');
probe(null);
probe(undefined);

Node command:
    node --input-type=module-typescript < internal/oracle/testdata/representation_clock_string_null_undefined.a
Node stdout, verbatim (exit 0, empty stderr):
string:false:false:text
string:false:false:
object:true:false:fallback
undefined:false:true:fallback

Production compiler observation on the same minimal program:
    go run ./cmd/adamic c internal/oracle/testdata/representation_clock_string_null_undefined.a
Exit 1. Observed stderr:
adamic: /tmp/notyet-representations-clock/string_null_undefined.a:1:16: stage 0 can't lower a value of type string | null | undefined yet
exit status 1

Where to fix

internal/lower/expression.go:30 representation, nullable-union refusal at 83-93; typeOf at 15. Related observations are internal/lower/typeof.go:10 typeOfNull and the existing conversion helper fit in expression.go. Do not edit the binary worker's combine function; preserve or coordinate its null and undefined comparison contract.
Packages: internal/lower, internal/oracle; if tags are added, internal/ir, internal/native, internal/javascript, and flow walkers that carry that IR.

This is one three-state string representation: present string, null, undefined. Keep empty string distinct from both absent states. Merely returning ir.String or ir.Union collapses information or mislabels typeof. Hold typeof, strict null/undefined equality, and ?? to Node. Keep conversions, retain/release, and stale narrowing checks consistent. Any runtime C change must be a new, separate helper reviewed by the runtime owner.

Fixture and registration

Create internal/oracle/testdata/representation_clock_string_null_undefined.a and its own internal/oracle/representation_clock_string_null_undefined_test.go. Register the fixture in that file; do not edit internal/oracle/oracle_test.go. For census-only parameter cases, also add a direct checked-type representation test in a new internal/lower clock-specific _test.go. Existing oracle registration and mutant examples are in internal/oracle/representation_generic_state_test.go and representation_scalar_constraint_test.go.

Done when

The registered fixture agrees with Node in the JavaScript backend and sanitized native backend, including leak checks. Add and run this named mutant: clock-string-null-undefined-collapse-null: emit the undefined encoding for a null input. Node's object:true:false:fallback line must differ from the mutant's undefined result. Every new admission guard also needs a mutant that demonstrably fails its targeted test. Replay the exact site again: its representation stop is gone, or list the next named stop; if it is a census echo, retain evidence and cancel the kind explicitly. Run focused tests of every package touched. Refresh counts with:
    go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/clock-string_null_undefined-counts.log 2>&1
Write all test output directly to log files. Name every file outside representation in the commit message, name any new separate runtime helper in the report, commit on your own branch and push.

Ownership

Before editing a lowering helper, check git log --remotes=origin/codex/notyet-* -- <file>. Another worker's internal/lower function stays theirs. Keep this contribution limited to the reserved kind; do not implement the other five clock briefs or the parent worker's remaining list.
