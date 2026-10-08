Built: roadmap step 09 diagnostic optional-presence checked view; constructor casts remain blocked.
Commits: mechanism b6cf696e; fixture/evidence commit is the branch tip following it, based on 432d4913.
Commands and outputs: focused lower/oracle PASS, admission mutant caught, dedicated counts PASS; required global counts FAIL with 39 fixture failures.
Mutants: field-check omission exits 0 on native release and JavaScript instead of pinned 70; admission omission fails TestOptionalPresenceViewCast.
Not covered: Error/String constructor lowering or their runtime liars, full scanner, full gate, unrestricted any contracts, new initialization semantics.

The String frontier recorded below is superseded by [CONSTRUCTORS.md](CONSTRUCTORS.md). Error now also has an exact library cast NotYet pin in that follow-up.

## Reductions and outside evidence

The three supplied coordinates refer to the adapted scanner scratch tree. Read-only evidence at scanner worker commit `6a2d222c` contains `03-error-cast.a`, `04-diagnostic-cast.a`, and `07-string-cast.a`; no worker branch was merged. Source declarations are from TypeScript commit `050880ce59e30b356b686bd3144efe24f875ebc8`:

- [debug.ts](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/debug.ts#L200): `(Error as any).captureStackTrace`. `error-any.a` preserves the presence test and invocation on the constructor. Source Node prints `ok`.
- [processDiagnosticMessages.mjs](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/scripts/processDiagnosticMessages.mjs#L91) emits the `diag(...) as DiagnosticMessage` casts. [types.ts](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/types.ts#L7218) declares `elidedInCompatabilityPyramid?: boolean`. The smallest decisive projection is the original optional boolean parameter returned as an always-present object property. `diagnostic-good.a` and `diagnostic-undefined.a` keep precisely that presence mismatch. This projection does not claim the other original fields or the entire generated map lower.
- [scanner.ts](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/scanner.ts#L4063): `(String as any).fromCodePoint`. `string-any.a` retains the conditional callable selection; source Node prints `A`.

All three are untagged: none has a declared discriminant proving a variant. They need checked views, not tag checks. For the diagnostic projection, exact optional assignability rejects required `boolean | undefined` becoming optional `boolean`, although reads have the same type. Lane 1's optional admission is missing this route. The separate mechanism commit admits only unchanged source property read types and qualifiers, requires all source properties survive, and rejects nominal/callable/indexed/container roots. It uses the existing view builder and runtime checks. It changes no backend, readiness analysis, protected orchestration file, or cohere source.

`diagnostic-missing.a` demonstrates a genuinely absent optional field returning `undefined`; `diagnostic-wrong.a` supplies a number through an erased structural origin. Source Node prints `42`, while sanitized native, release native, and JavaScript stop with exit 70 and this exact stderr:

```
adamic: panic: field read failed: value.elidedInCompatabilityPyramid matches no member of boolean | undefined; expected boolean | undefined, found number
```

Finishing native controls also pass leak checks. Existing lazy unread/uninitialized and missing-name mutant tests were included in the focused run; no new readiness mechanism is claimed.

## Constructor dependencies remain open

`as any` has no concrete member contract in the current cast proof. Lane 1/common admission needs a demanded member contract and a host constructor value representation. Callable contracts are lane 2 territory in the plan; intrinsic producer certification needs the producer work described by the integration, rather than merely checking `typeof function`. These are missing for both constructor uses. The blocker census leaves general any-field contracts unassigned.

`error-typed.a` uses the installed Node `ErrorConstructor.captureStackTrace(targetObject: object, constructorOpt?: Function): void` declaration and still receives the named `node:globals.ErrorConstructor.captureStackTrace` library frontier. Native stack capture cannot soundly be replaced with a no-op. `string-typed.a` uses `Pick<StringConstructor, "fromCodePoint">`, retaining the actual rest-call signature; it stops at the missing `String` constructor value representation. All four frontiers are pinned by Node-valid tests. No claim is made that these two original stops lower; their backend liar/mutant matrices remain unfinished. Building generic host constructor storage and callable certification would exceed the small optional-presence mechanism delivered here.

## Commands and evidence

All test output was redirected to logs. Environment commands used `export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Setup passed with Go 1.27.1, Node 24.19.0, clang 20.1.8; `nproc` was 5 (cgroup CPU quota 4). Timing lines: Go ready 0.067s; Node ready 0.068s; clang ready 0.583s; markdown install 0.885s; dependencies ready 1.078s; submodules ready 199.405s; go build ready 480.924s; test binaries deferred 481.017s; build cache warm 481.018s; done 481.047s. Full setup output: [logs/setup.log](logs/setup.log).

Final focused command:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestOptionalPresenceView|^TestViewPreflightPreservesOtherRefusals$|^TestOptionalWidening(Views|Allowed)$|^TestScannerCast|^TestCheckedViewLazy(Unread|MissingNameMutant)$' -count=1 -v -timeout 10m > /tmp/scanner-cast-final-tests.log 2>&1
```

PASS: lower 0.947s, oracle 7.431s. [Complete output](logs/final-tests.log).

```
python3 stage3/drivers/scanner/cast-checks/run-admission-mutant.py > /tmp/scanner-cast-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/scanner-cast-counts-global.log 2>&1
go test ./internal/oracle -run '^TestScannerCastCounts$' -count=1 -v -timeout 10m -args -update-counts > /tmp/scanner-cast-counts-unit.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/scanner-cast-vet.log 2>&1
```

The admission mutant uses a temporary Go overlay to return false from the new route, and catches the named positive lower test failing; [mutant output](logs/mutants.log), [failed test](logs/admission-mutant.log). The field-check mutant removes exactly one consumed field's view check from IR. Valid release native prints `false`; JavaScript prints `42`; both exit 0, violating the original pinned stop. The focused log records both catches.

The required global counts refresh fails on 39 fixtures and cannot write the global table. Failures include graph frees, existing lowering frontiers, FS representation checks, and C signatures. [Full failures](logs/counts-global.log). Restoring the base preflight file with a Go overlay reproduces the `process_exit_default` and `census_small_boolean` failures; this verifies two representative baseline failures, not all 39. [Baseline output](logs/counts-baseline.log). The dedicated scanner counts update passes and records four measured rows while preserving every other row. The ordinary counts hook includes those four fixtures for the next successful global refresh. [Unit counts output](logs/counts-unit.log).

Formatting and `git diff --check` are clean; targeted vet exits 0. No whole-package test, full gate, peer-lane merge, main merge, or PR was run. The requested integration base takes precedence over the ordinary main landing instructions.
