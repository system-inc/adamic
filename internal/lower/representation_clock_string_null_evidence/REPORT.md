# Nullable string contribution

Baseline: origin/area/compiler at 8f89d5f4624feaffdf1797bcb3db8e20a202d79d.
Scope: the reserved `a value of type string | null` representation kind only.

Assumption: the brief authorizes an exact string/null union using the existing string reference
representation. Null is the missing pointer; empty string remains present. There is no undefined
member and no wider tag protocol. `clockNullableString` admits only null plus StringLike members.
The direct checked-parameter test also rejects undefined, numbers, mixed members and objects.

The exact minimal fixture is registered in its own oracle test file. Its Node stdout is:

```text
string:false:text
string:false:
object:true:fallback
```

Additional registered fixtures cover lazy fallback (text and empty string do not evaluate it),
strict equality and inequality in both directions, undefined comparison, observations after a call
invalidates narrowing, and a stale string access that still traps identically to Node.
The existing `typeOfNull` helper supplies typeof's declared null interpretation unchanged.
`clockStringNullObservation` repairs strict null/undefined observations for declared nullable string
slots after stale narrowing. The shared `combine` helper is unchanged.

Ownership history for expression.go and typeof.go was inspected before editing; the attached logs
retain that history. Changes in expression.go are two dispatch hooks into the new clock helpers.
No new separate runtime helper, IR layout, native runtime or JavaScript runtime was introduced.
No unchecked assertion, type erasure, or production checker bypass was added.

## Validation

Setup environment: `source /workspace/adamic-tools/env.sh`.
All test command output was redirected directly to log files.

- `go test ./internal/lower -count=1` passed.
- Relevant oracle fixtures and existing typeof-null controls/mutants passed; registered oracle
  checks compare Node, JavaScript, sanitized native and release native, with leak checks.
- `clock-string-null-typeof-undefined` clears the emitted typeof null interpretation. The native
  mutant finishes cleanly and leak-free; stdout changes to `undefined:true:fallback` and kills it.
- `clock-string-null-null-guard` is killed by the checked `present-literals` admission test.
- `clock-string-null-member-guard` is killed by the checked `three-state` admission test.
- `clock-string-null-stale-null-observation` and `clock-string-null-stale-undefined-observation`
  are killed by the observations fixture's Node stdout comparison.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts` passed. Only the
  three new fixture rows were added in internal/oracle/counts.md.

The two Python mutant scripts run from the repository root with setup sourced, save each test's
output to /tmp, assert its targeted failure, and restore the production helper in a finally block.

## Exact census replay

Reconstructed stage3 from commit 9d534d3a31814f1a192a528e701f6c2ea7c910bc and ran its apply.sh.
All 81 adapted tsc source byte lengths and SHA-256 hashes matched the archived source manifest.
The representations worker's guarded replay tooling was used temporarily and removed afterward;
none of that worker's commits or replay implementation is carried in this contribution.

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:102:52 -kind NotYet -reason 'a value of type string | null'
```

The requested representation stop is gone. The replay exits 1 because the old requested signature
no longer reproduces; the next named lowering stop is `NotYet: an array of never` at
`/tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:105:51`, within the same selected
`setSourceContent` unit at 102:5. Full measurement stdout and selected findings are retained.
This is measured on a checker-rejected entry-root program, not a successful production compilation.
No census echo cancellation or implementation of the next kind is needed.
