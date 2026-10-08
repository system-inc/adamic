# Three-state string representation

Base: `origin/area/compiler` at `cf308b2c121814ebb6d89cd76be0742d599a1965`.
Scope: the reserved `a value of type string | null | undefined` kind only.

`ir.NullishString` preserves present strings (including empty strings), null, and undefined.
Native storage is a counted string reference, an immortal generated null tag, or NULL respectively.
The null tag uses an existing object header with zero reference count, so the existing typeof
classifier reports `object` and the existing retain/release rules treat it as immortal.
There are **no runtime C source changes and no new runtime helper**.
JavaScript uses its native three states through the existing Box emission and flow walkers.

Clock-owned helpers admit only string members plus both absent states, preserve declared field
storage, and compare the two absent tags independently. The binary worker's `combine` function
is unchanged. Coalescing rejects both absent tags; string equality uses the existing value comparator.
Calls, returns, assignments, property reads, truthiness, and stale narrowing have focused witnesses.
Narrowed reads used as strings check the held member; observations preserve the actual state.

The fixture is registered in its own test file; `oracle_test.go` is unchanged. It agrees with Node
in JavaScript and sanitized native runs, including leak checks. Node's four stdout lines are:

```text
string:false:false:text
string:false:false:
object:true:false:fallback
undefined:false:true:fallback
```

The named `clock-string-null-undefined-collapse-null` mutant changes exactly the null argument to
undefined. It finishes cleanly with leak checks and changes the third line to
`undefined:false:true:fallback`; the Node stdout comparison kills it.

Admission mutants and their targeted negative witnesses:

| Mutant | Witness |
|---|---|
| clock-string-null-undefined-admit-without-null | string \| undefined keeps its existing String representation |
| clock-string-null-undefined-admit-without-undefined | string \| null remains unadmitted |
| clock-string-null-undefined-admit-without-string | null \| undefined remains unadmitted |
| clock-string-null-undefined-admit-mixed | string \| number \| null \| undefined remains unadmitted |
| clock-string-null-undefined-typeof-admission | the registered fixture must lower typeof its carrier |

Each mutant fails its targeted test without a build failure. Repeat after sourcing the setup environment:

```sh
python3 internal/lower/representation_clock_string_null_undefined_mutants.py --logs /tmp/clock-admission-mutants
```

Validation output is retained in `representation_clock_string_null_undefined_evidence/`:

- Full `internal/lower` package.
- `internal/ir`, `internal/flow`, `internal/fresh`, and `internal/javascript` (the latter has no test files).
- Native borrowing/reference tests including `TestPassThroughsAreNotConsumers`.
- Clock fixture, named mutant, storage and local/field stale narrowing tests.
- Existing typeof oracle mutants together with the clock tests.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`.
- Production `adamic c` succeeds on the fixture; exact Node stdout is retained.

Counts changed only by the new fixture row: allocations 4, frees 4, retains 4, releases 8,
peak live 1, values released in regions 0.

## Exact-source replay

Input reconstruction uses `stage3/apply.sh` from
`9d534d3a31814f1a192a528e701f6c2ea7c910bc`. All 81 source byte lengths and SHA-256 hashes agree
with that commit's tsc source manifest. The guarded replay tool was temporarily read from
`origin/codex/notyet-representations` at `37ff860f` and restored after measurement; none of its
commits or files are carried by this branch.

```sh
go run ./stage3/census/latent/replay \
  -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts \
  -where /tmp/notyet-representations-adapted/src/compiler/sourcemap.ts:701:24 \
  -kind NotYet -reason 'a value of type string | null | undefined'
```

The requested representation signature no longer reproduces. The replay exits 1 because its
old-signature reproduction assertion now fails; this is the expected result after removing the stop.
The next same-statement stop is `NotYet: reading mapDirectory` at `sourcemap.ts:701:83`.
The unit also records `Refused: overload 1 of getDirectoryPath result Path cannot be served by
implementation result string` at `path.ts:278:1`, followed by other independent stops.
Full measurement JSON and stderr are retained. These are measurements of a checker-rejected
entry-root program, not a claim that the whole entry root compiles. This kind was implemented,
not canceled as a census echo.

Assumption: admission is limited to the complete three-state string union and string-literal
variants, rather than extending other nullable representations or importing the parent worker's changes.

## Fast-gate correction

The uncached gate exposed an invalid leak assertion in the intentional stale-narrowing panic
tests. `natively` correctly returns exit 70; `leakSanitizer` requires exit 0 and consequently
reported that expected panic as a failure. The cached native path skips leak checks for nonzero
exits, masking the mistake in the initial verification.

Remove the leak assertion only from the panic cases. They still run sanitized native code and
require the expected panic and agreement with JavaScript for both absent states in local and
field storage. Normal fixture, storage, and collapse-null mutant runs retain their leak checks.
No compiler, runtime, fixture registration, or counts changes are needed.
