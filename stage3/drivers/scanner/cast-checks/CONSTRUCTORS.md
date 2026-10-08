Built: step 09 String scanner cast now lowers through a checked intrinsic member view; Error has an exact library NotYet fixture.
Commits: mechanism b52fbc43; fixtures, counts and this report follow in the branch tip, without another lane merge.
Commands and outputs: focused lower/oracle PASS, vet and formatting PASS, unit counts PASS; global counts FAIL on 39 fixtures.
Mutants: admission, native producer certificate, JavaScript producer certificate, and whole-read omission all caught by named assertions.
Not covered: native Error facts/stack capture, escaping or mutable constructors, general any/rest callable views, source-level intrinsic mutation liars, full scanner or full gate.

## Exact ownership and stops

| Scanner stop | Blocking file and line before this follow-up | Ownership and delivered result |
| --- | --- | --- |
| `debug.ts:15:14`, `(Error as any).captureStackTrace` | `internal/lower/cast_proof.go:118` rejects the legacy any wrapper. After reducing that wrapper to the actual declared member contract, `internal/lower/library_node.go:93-94` emits `node:globals.ErrorConstructor.captureStackTrace`. The declaration is `stage3/api/node_modules/@types/node/globals.d.ts:51`. | The executable Error operation and constructor representation need library's Error facts, identified by the user as `library/area-on-next`, task `#ddwcejg`. The typed reduction has no blanket any refusal hiding that dependency. No Error host behavior is fabricated here. |
| `scanner.ts:3497:67`, `(String as any).fromCodePoint` | `internal/lower/cast_proof.go:118` rejects any; `internal/lower/library_globals.go:29` rejects the first-class constructor value. Once those are bypassed for a demanded member, `internal/lower/functions.go:80-90` cannot represent the original arrow's inferred any result. | This scanner use needs our checked-view admission and an intrinsic producer adapter, not Error facts. The ordinary one-number Unicode operation already exists at `internal/lower/object.go:774-775` and `:1948`. These pieces are built in b52fbc43. General first-class String values still retain their library frontier. |

Line references above use this delivery tree. The base cast guard was at `cast_proof.go:115` and the base function result refusal at `functions.go:87`. The adapted scanner coordinates are the supplied worker stops; original TypeScript provenance remains in [REPORT.md](REPORT.md).

The exact library branch fetch failed with `fatal: couldn't find remote ref library/area-on-next`; [fetch log](logs/constructor-fetch.log). Error ownership comes from the user's task assignment and the observed library member frontier, not from an inspection or merge of that unavailable branch.

## Error fixture that flips

[error-library-cast.a](error-library-cast.a) keeps a cast on Error but projects `Pick<ErrorConstructor, "captureStackTrace">` from the real Node declaration. Source Node prints `ok`. `TestScannerErrorLibraryCastNotYet` requires the error to be a `*lower.NotYet`, with exact fixture location `:3:1` and exact `What`:

```
node:globals.ErrorConstructor.captureStackTrace
```

The current full diagnostic is:

```
error-library-cast.a:3:1: stage 0 can't lower node:globals.ErrorConstructor.captureStackTrace yet
```

The actual diagnostic prefixes the absolute fixture path. The test fails if lowering succeeds or changes to another failure. It therefore flips when the library member frontier changes, instead of accepting a different NotYet under a loose substring. Its explicit closed-gap message requests both-backend execution and counts. Registering a member name alone is not evidence of native stack behavior. No counts row is invented for a fixture that cannot execute. The original `error-any.a` stays as a distinct legacy-wrapper refusal control; this follow-up does not claim to lower that original wrapper or all of Debug.fail.

## String checked view

The minimum supported scope is the scanner's immediate `fromCodePoint` presence test and an immediate call with one checker-proven numeric argument. It does not expose constructor or callable identity. The private carrier represents just that demanded intrinsic member, with an initialized callable and a certificate tied to the generated Unicode adapter. Every consumed member read goes through the existing field and callable checks. The adapter reuses `ir.StringFromCodes`, so Unicode validation and native string ownership remain library behavior.

Only an arrow whose entire expression body is this supported direct call can recover a string result from the legacy any inference. No arbitrary any-returning function is admitted. A closed-program scan rejects other escaping or mutable String uses. Lower tests retain the two actual consumed checks and refuse detached members, constructor escapes, nonnumeric arguments, multiple/spread arguments, shadowed lookalikes, identity comparisons, aliases, mutable hosts, and unrelated any-returning arrows.

The missing producer enforcement was in `internal/native/view_unions_untagged.go:149` and `internal/javascript/view_unions_untagged.go:141`: both only applied `ProducerCertified` to callable union roots. The smallest sound extension now also applies it to a single callable root, at native `:157` and JavaScript `:149`. This is the lane 1 admission / callable producer integration piece, using lane 5's existing immutable code registry, rather than a new callable ABI or a type assertion treated as proof.

[string-any.a](string-any.a) is the unchanged original reduction and prints `A`. [string-checked.a](string-checked.a) prints a supplementary-plane character and then `1`, demonstrating the numeric argument's side effect is evaluated once. Source Node, ASan/UBSan native, release native, and emitted JavaScript agree. Both finishing native controls pass leak checks.

## Liars and mutants

The producer liar substitutes an independently generated closure into the lowered intrinsic carrier. It has the same numeric-argument/string-result ABI but lacks the intrinsic's certified code identity. This is an IR host-producer fault injection, not a source program accepted despite mutation of the builtin. Both backends, including sanitized native, stop with exit 70, empty stdout, and exactly:

```
adamic: panic: field read failed: (String as any).fromCodePoint expected (codePoint: number) => string, found function with unknown signature
```

Removing the whole consumed read check makes valid release native and JavaScript exit 0 and print `liar` followed by `1`. This mismatch catches the omission without relying on clang or a sanitizer.

[run-constructor-mutants.py](run-constructor-mutants.py) also runs three independent Go overlays, never mutating the checkout:

- Remove scoped cast admission: `TestScannerStringConstructorView` fails with the original any cast refusal.
- Remove only native single-producer certification: `TestScannerStringConstructorProducerMutant` fails because native executes the liar with exit 0 instead of 70.
- Remove only JavaScript single-producer certification: the same test fails because JavaScript executes the liar with exit 0 instead of the exact pinned stop.

[Mutant summary](logs/constructor-mutants.log) and the three individual logs record named test exits 1. Existing callable union and ordinary checked-callable controls pass after the extension.

## Commands and results

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-constructor-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/drivers/scanner/cast-checks/run-constructor-mutants.py > /tmp/scanner-constructor-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/scanner-constructor-counts-global.log 2>&1
go test ./internal/oracle -run '^TestScannerCastCounts$' -count=1 -v -timeout 10m -args -update-counts > /tmp/scanner-constructor-counts-unit.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestScanner|^TestOptionalPresenceView|^TestViewPreflightPreservesOtherRefusals$|^TestLibraryStringRefusals$|^TestCheckedViewUntaggedCallableUnion$|^TestCheckedViewCallables$' -count=1 -v -timeout 10m > /tmp/scanner-constructor-final.log 2>&1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/scanner-constructor-vet.log 2>&1
```

Final focused suite PASS: lower 0.945s, oracle 7.603s. [Full final log](logs/constructor-final.log). Targeted vet exits 0, changed Go files produce no `gofmt -l` output, and source/document `git diff --check` is clean.

The required global counts command fails on the same number of inherited fixture failures, 39; it cannot write the global table. [Failures](logs/constructor-counts-global.log). The prior unit's report records two representative baseline reproductions; this follow-up does not claim to have re-proved every failure at the base. The dedicated count update PASS records two new String rows and preserves all earlier rows: original scanner reduction allocations/frees 6/6, retain/releases 4/9, peak 4; one-evaluation control 6/6, retain/releases 1/6, peak 4. [Count log](logs/constructor-counts-unit.log). An earlier combined run encountered the new rows before this update finished; the final suite above was rerun after the rows were recorded and passes.

Setup PASS, `nproc` 5, cgroup CPU quota 4. Timing lines: Node ready 0.022s, Go ready 0.023s, submodules ready 0.062s, markdown dependencies skipped 0.006s, dependencies ready 0.077s, clang ready 0.181s, Go build ready 42.092s, test binaries deferred 42.202s, build cache warm 42.203s, done 42.230s. [Setup log](logs/constructor-setup.log).

No protected compiler orchestration files, cohere files, or library Error implementation were edited. No peer branch was merged, no full package test or full gate was run, and no PR was opened. Delivery stays on `codex/scanner-cast-checks` based on the requested checked-views integration commit.
