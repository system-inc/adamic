# Stricter checker options: verified audit, incomplete conversion

Built project options/lib loading, supported indexed guards and JSON use guards in both backends, with explain counts.
Commit: recorded in the worker's final report; branch codex/stricter-options-checks only.
Measured 0 project errors and all 173 exact updated-ledger sites, with no missing, extra or wrongly attributed sites.
Six audit, four loader, thirteen indexed (including five sparse), three JSON and one inherited iterator mutant were run and caught.
Catch and optional checks and the individual 99-site witnesses belong to other workers. This branch owns indexed representations; typed arrays and records remain unsupported.

## What is built

`load.AuditProjectOptions(ctx, tsconfig)` parses the project's config through the
existing TypeScript shim, including inherited options, lib and references. It
checks the unchanged project, then the same inputs with noUncheckedIndexedAccess,
exactOptionalPropertyTypes, useUnknownInCatchVariables and strictBindCallApply.
A separate checker run restores each option to its project's value. Site identity
is file, position and diagnostic code, rather than total-count subtraction.
Ordinary project diagnostics remain separate. The audit also collects declaration
diagnostics. Compiler options use the checker's Clone method, rather than copying
its noCopy state.

The audit API remains analysis only. Production Load now uses its site report to retain unconverted option errors and expose them on CheckError.OptionSites. An empty
Options list remains unclassified evidence. Diagnostics introduced jointly by
options retain every responsible option.

The probe and validator reproduced the ledger at
3f0926c0a55a7b5f64f037b1745e0e984e08c8be against the adapted tree built with that
revision's stage3/apply.sh. That revision includes the adaptations based on
234ab1aa. The validator compares every site and the full option-ablation membership
to rows.csv. Normalized measured diagnostics are retained in sites.json.

| Measurement | Observed result |
|---|---:|
| Project errors | 0 |
| Stricter-option sites | 171 |
| Missing sites | 0 |
| Extra sites | 0 |
| Incorrect option attribution | 0 |
| Indexed-read primary attribution | 99 |
| Exact-optional primary attribution | 67 |
| Catch primary attribution | 5 |
| strictBindCallApply primary attribution | 0 |

One indexed site is also removed by the exact-optional ablation. Its two option
names are retained; the primary order follows the ledger. These 171 rows describe diagnostics, not inserted check counts. Current dispositions are recorded below and in disposition.csv.

## Runtime proof gaps observed on current main

The four .a witnesses under gaps are intentionally small. gaps.cjs independently
runs their type-erased original source on Node, then attempts a native build.
Each Node result exits 0. Each native build exits 1 before producing a runnable
artifact, with the diagnostic below. These are compiler gaps, not exit-70 checks.

| Witness | Source Node stdout | Native build observation |
|---|---|---|
| array-hole.a | undefined | Cannot lower new Array<number>(2) |
| typed-array.a | undefined | Cannot lower new Uint8Array(0) |
| record.a | undefined | Cannot lower the record's element access |
| catch-value.a | not Error | Refuses throwing a string |

The existing native array runtime has dense storage and no hole bit. The existing
exception lowering accepts Error construction and caught Errors; its caught
instanceof Error lowering returns true. Inference: using that constant as the
requested catch check would provide no failing native witness. A correct runtime
conversion requires extending these representations and lowering paths first.
This branch does not claim these extensions are impossible; they are not built.

The five catch rows in the actual rows.csv are commandLineParser.ts:2301,
program.ts:406 and 441, and sys.ts:1281 and 1617. They are not all in sys.ts.
The 18:05 ruling is retained: their eventual .ts implementation must keep the
project's type and check the relevant uses, rather than blanket unknown typing.

All 67 optional rows remain errors in this loader slice. The user reports that optional-field-write-2 at 7e7464e6 now matches Node for in, Object.keys and Object.hasOwn across aliases. The latest scope assigns all 67 to codex/stricter-optional-writes, based on b651b305 and that presence branch. This worker does not merge or implement that slice.
The two JSON.stringify declaration-contract rows are outside this flag-only
audit. Their use-site checks and the .a diagnostic naming the fix are not built.
Production .ts loading now preserves the owning project's options and lib, with strictNullChecks and strictFunctionTypes retained as Adamic soundness rules. It keeps the existing prelude. --explain-checks now counts actual lowered guards by kind and reports trusted: 0. Catch/JSON and optional guards have not yet been built. Planning range communicated October 7: October 8 to 9 by 21:30 MDT for supported indexed representations and catch/JSON, conditional on a second catch/JSON worker and a failing caught-value representation witness; it is not a verified completion date.

## Mutants actually run

`mutants.py` uses Go overlays, leaving the working tree untouched. Each independent
mutant exits 1 with a TestProjectOption assertion failure, never a build failure.
These prove checker analysis, not emitted runtime behavior.

| Mutant | Catcher |
|---|---|
| erase-index-option | TestProjectOptionAttribution loses the indexed row |
| erase-optional-option | TestProjectOptionAttribution loses the optional row |
| erase-catch-option | TestProjectOptionAttribution loses the catch row |
| erase-ordinary-membership | TestProjectOptionAttribution reclassifies an ordinary type error |
| overwrite-project-lib | TestProjectOptionsPreserveInheritedLibAndStrictness reports document missing |
| erase-site-position | TestProjectOptionAttribution conflates different TS2322 sites |

Final per-mutant logs are named in /tmp/stricter-options-mutants-final.log.
The compiler source remains unchanged by all six runs.

## Commands and outputs

Every test and probe wrote its complete output to a log file.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-options-setup-final.log 2>&1
source /workspace/adamic-tools/env.sh
bash /tmp/stricter-options-stage3/stage3/apply.sh /tmp/stricter-options-adapted > /tmp/stricter-options-apply.log 2>&1
npm ci --prefix /tmp/stricter-options-adapted --ignore-scripts --no-audit --no-fund > /tmp/stricter-options-npm.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node /tmp/stricter-options-stage3/stage3/ledger/checker-259/measure.cjs /tmp/stricter-options-adapted /tmp/stricter-options-ledger-final > /tmp/stricter-options-ledger-final.log 2>&1
go run ./stage3/stricter-options/probe.go /tmp/stricter-options-adapted/src/compiler/tsconfig.json > /tmp/stricter-options-project-report-final.json 2> /tmp/stricter-options-project-report-final.log
python3 stage3/stricter-options/validate.py /tmp/stricter-options-project-report-final.json /tmp/stricter-options-adapted --save stage3/stricter-options/sites.json > /tmp/stricter-options-validation-final.log 2>&1
python3 stage3/stricter-options/mutants.py > /tmp/stricter-options-mutants-final.log 2>&1
go build -o /tmp/stricter-options-adamic ./cmd/adamic > /tmp/stricter-options-build.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/stricter-options/gaps.cjs /tmp/stricter-options-adamic > /tmp/stricter-options-gaps.log 2>&1
go test ./internal/load ./stage3/stricter-options -count=1 -timeout 10m > /tmp/stricter-options-final-packages.log 2>&1
go vet ./internal/load ./stage3/stricter-options > /tmp/stricter-options-vet.log 2>&1
```

The final loader package passes in 3.104s; the probe package has no test files.
Vet exits 0 without diagnostics. Six mutants are caught. Four gap witnesses
agree with the stated Node outputs and native rejection. The validator reports
171 sites, 0 project errors, 0 missing, 0 extra and 0 wrong attributions.

The independent stock measurement prints: own 0, census-inputs 258, effective
233, adamic 413, project-stricter 171; ablations leave 159 without unchecked
indexes, 192 without exact optionals, 254 without unknown catches, and 258
without strictBindCallApply. It also matches every one of the 171 site keys.

Setup's first run completed Go readiness at 0.095s, Node at 0.249s, clang at
0.709s, markdown dependencies at 1.384s and submodules at 228.616s. Its cache
warming encountered the new audit's int32-to-int diagnostic conversion error.
That implementation error was corrected and setup rerun successfully. The retry
prints Go 0.045s, Node 0.046s, submodules 0.111s, markdown dependencies 0.121s,
clang 0.368s, Go build 37.762s, cache warm 37.910s and done 37.999s. nproc=5;
cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.

The initial audit did not run runtime proofs. The indexed proof below now covers sanitizer builds, erase-check mutants and a filtered existing oracle. The full repository gate remains unrun. This is an incomplete unit.

## Production loader slice

The production probe reproduces all 171 ledger sites and all ablation memberships:
99 indexed, 67 optional, 5 catch. At that loader-only commit all remained errors. The production checker also
retains 90 diagnostics from the existing prelude/library contracts, separately
from the 171 sites. These are not converted or trusted. The stock-project audit
still has 0 ordinary errors. `--allow-project-errors` is an explicit validation
mode for this comparison; it does not suppress loader errors.

A single project's .ts roots use inherited project options and lib. Catch bindings
keep the project's type. Composite programs retain their complete root list and
references, avoiding manufactured TS6307 errors. .a roots keep Adamic options.
Standalone .ts files without a tsconfig keep the existing Adamic fallback.

Mixed .a/project-.ts ownership in either import direction and separate project
ownership are explicitly refused until distinct checker/AST ownership is built.
At b651b305 project overlays and explicit extra roots were refused. The library integration now audits overlay inputs and explicit ambient roots; extra implementation roots remain refused. These conservative refusals are
limitations, not completion of general per-file options. The compiler host caches
filesystem reads but parses fresh ASTs for each program; reusing it does not solve
cross-program symbol identity.

Production tests cover ordinary errors, retained sites, project catch types,
old-versus-new lib facts, .a strictness, mixed imports, composite roots and overlay
refusal. Four additional Go-overlay mutants actually failed their intended
TestProduction assertions: replacing project options, erasing recorded sites,
waiving optional sites and erasing the mixed-file refusal. Logs:
/tmp/stricter-options-production-mutants.log. These prove loader behavior only.

Commands:

```sh
go test ./internal/load ./stage3/stricter-options -count=1 -timeout 10m > /tmp/stricter-options-production-loader-final-3.log 2>&1
go vet ./internal/load ./stage3/stricter-options > /tmp/stricter-options-production-vet-final.log 2>&1
python3 stage3/stricter-options/production_mutants.py > /tmp/stricter-options-production-mutants.log 2>&1
# production probe takes the adapted compiler root source paths after --production
python3 stage3/stricter-options/validate.py /tmp/stricter-options-production-report.json /tmp/stricter-options-adapted --allow-project-errors > /tmp/stricter-options-production-validation.log 2>&1
```

Revised planning range after optional admission: October 9 to 10 by 21:30 MDT, assuming clean presence-branch integration and a second catch/JSON worker; October 10 to 12 without that split. Runtime witnesses and erase-check mutants remain unverified. This supersedes the earlier range.

## Supported indexed guards and actual check counts

Project array/string reads relied on as present now reuse the existing lookup and
Coalesce panic IR. Both backends carry the same site message:
`indexed read is absent: file:line:column`. The native lookup already decides
bounds/presence, so the guard does not repeat bounds or evaluate the receiver/index
twice. Null elements and tagged/weak element representations explicitly refuse
until presence can be retained independently of their payload. Tuple positions
remain statically checked. Existing strict-project guards retain their behavior.
Direct observations of undefined keep undefined, including string interpolation.

The joint indexed/exact-optional row is handled at its read. Its recorded ancestor
forces the guard despite the optional receiving context; this is independently
proved by the joint-optional fixture. All 99 primary indexed ledger rows now defer
to guarded lowering or an explicit representation refusal. They are not blanket
trusted facts. All 171 site keys and memberships still exactly match the ledger.
72 rows remain checker errors: optional 67 and catch 5. The existing prelude adds
90 other retained diagnostics; those are outside the flag ledger. Full adapted-tree
native lowering and its total emitted-check count are not claimed.

`--explain-checks <source>`, or build/c/js with `--explain-checks`, counts panic guards
actually present in lowered IR. A one-read fixture prints indexed-presence=1,
catch-error=0, json-stringify-defined=0, optional-write=0 and trusted: 0. Audited
rows and refused witnesses are never counted as inserted guards.

Twelve runtime fixtures pass against handwritten type-erased source on Node and
both backends. Eight fail at runtime with the named check and exit 70: number,
boolean, string array, string, negative index, fractional index, joint optional
context and receiver/index evaluation order. Each has an emitted-C erase-panic
mutant that successfully builds under sanitizers. Number/boolean/negative/
fractional/joint/order mutants exit 0; string-array/string mutants exit 1 under the
sanitizers after the erased guard permits a null dereference. All lose the required
named exit-70 stop and are caught. Present number/string and direct array/string
undefined observations pass. Each fixture asserts exactly one native lookup and
checks its actual IR guard count (one, or zero for direct observations).

The existing differential oracle passes indexing.a, string_index.a,
narrowed_reads.a and narrowed_numbers.a, including sanitized and release native
builds. The full gate was not run. Holes still refuse new Array<number>(2), typed
arrays refuse new Uint8Array(0), and records refuse their element access before
native emission. These existing lowering gaps remain outside this guard slice.

Logs and commands:

```sh
go test ./internal/load ./internal/lower ./internal/ir ./cmd/adamic ./stage3/stricter-options -count=1 -timeout 10m > /tmp/stricter-options-indexed-final-packages-2.log 2>&1
go vet ./internal/load ./internal/lower ./internal/ir ./cmd/adamic ./stage3/stricter-options > /tmp/stricter-options-indexed-final-vet-2.log 2>&1
go test ./stage3/stricter-options -run TestIndexedPresenceRuntime -count=1 -timeout 10m -v > /tmp/stricter-options-indexed-joint.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(indexing|narrowed_reads|narrowed_numbers|string_index)\.a$' -count=1 -timeout 10m -v > /tmp/stricter-options-indexed-oracle.log 2>&1
python3 stage3/stricter-options/validate.py /tmp/stricter-options-indexed-report-final-2.json /tmp/stricter-options-adapted --allow-project-errors --disposition stage3/stricter-options/disposition.csv > /tmp/stricter-options-indexed-validation-final.log 2>&1
```

CLI smoke log: /tmp/stricter-options-indexed-cli-smoke.log. Its generated .ts fixture
builds, reports one checked guard and zero trusted, then exits 70. The same source
loaded as .a remains a checker error. Loader-mutant rerun log:
/tmp/stricter-options-indexed-production-mutants.log. Ledger summary:
/tmp/stricter-options-indexed-ledger-summary.log. The optional branch is authorized
for a later merge at 7e7464e6 after the current steps, not yet merged here.

## Latest scope and date

This worker owns 106 sites: indexed 99, catch 5 and JSON.stringify 2. The optional worker owns the other 67. The updated ledger at a1a16427 records 173 after the library iterator fix; this indexed commit still reports the earlier 171 comparison, and has not yet integrated that fix or reproduced 173. Planning estimate: October 8 to 9 by 21:30 MDT; first 50 take 4 to 8 hours assuming their representations already lower. Catch representation support, JSON use checks, per-site witnesses and the updated ledger remain work, not measured completion. The previous optional-merge estimates are superseded.

Final generic-read verification: internal/lower passes in 18.045s and stage3/stricter-options in 11.040s; log /tmp/stricter-options-indexed-generic-final.log.

## Current verified result after library and JSON integration

The library dependency 24d980a7 is merged at 5d074f12. This worker has not merged
optional-field-write-2: the latest scope assigns its 67 rows to
codex/stricter-optional-writes. The indexed slice was pushed at 5aef91ae.

The fresh adapted tree comes from area/stage3 3b255125, applied with its apply.sh.
Its patch-set total is 78 files, 5128 lines added and 5102 removed. The isolated
ledger merge 8c50a0a9 could not be fetched (remote: not our ref); the published
adaptation parent was fetched and applied instead. The library changes affect
Adamic, not those upstream adaptations. The production probe on this fresh tree
matches every site in a1a16427's rows-2026-10-08.csv:

| Measurement | Result |
|---|---:|
| Ordinary project errors | 0 |
| All recorded sites | 173 |
| Indexed | 99 |
| Optional | 67 |
| Catch | 5 |
| JSON.stringify | 2 |
| Missing / extra / wrong attribution | 0 / 0 / 0 |
| Remaining checker errors | 72 |
| Deferred to guarded lowering or explicit refusal | 101 |

The 90 previous prelude rows are gone. The two JSON errors are now recorded
contract sites, not ordinary errors. These 101 deferred rows are not a count of
101 guards emitted from the whole adapted compiler. The compiler's entire native
build and one native witness per ledger row remain unverified. sites-173.json and
disposition.csv retain the comparison and disposition.

JSON.stringify retains string | undefined in the production checker. A separate
checker changes only that declaration to identify diagnostics caused solely by
the truthful result type. Diagnostics at the same file/position/code that survive
remain errors. In .a, the diagnostic names narrowing or ?? as the fix. In project
.ts, a string-requiring use gets a Coalesce panic with the message
`JSON.stringify result is undefined: file:line:column`. The original result may
still be stored or observed as string | undefined.

Five JSON fixtures pass: undefined, function, held result, present number and
observed undefined. The first three give undefined on Node; both backends stop
at the use with the same named message and exit 70. Each emitted-C erase-panic
mutant successfully builds under sanitizers, then exits 1 after the missing value
reaches the unguarded string use. The named exit-70 assertion catches each mutant.
The present fixture has one actual JSON guard; observation has none. Together
with the twelve indexed fixtures, seventeen runtime fixtures pass.

The CLI fixture with one indexed read and one JSON use prints:

```
checked: indexed-presence=1 catch-error=0 json-stringify-defined=1 optional-write=0
trusted: 0
```

The final touched packages pass: load 24.310s, lower 51.632s, ir 40.306s,
cmd/adamic 4.337s, runtime fixtures 31.475s. Vet exits 0 without diagnostics.
The filtered existing oracle passes eight fixtures (three JSON, indexing,
string index, exceptions, class-inheritance exceptions, fallthrough exceptions).
The integrated iterator oracle also passes six project fixtures, six standalone
fixtures and TestLibraryIteratorDoneMutant. That mutant replaces omitted done's
false fallback with true; it exits 0 cleanly and Node's stdout comparison catches
it. WASI and the full repository gate were not run.

The four loader mutants were rerun and caught by their intended assertions, not
build failures: replace-project-options, erase-recorded-sites, waive-optional-sites
and erase-mixed-file-refusal. The six audit mutants listed earlier are rerun after
the filesystem-aware attribution refactor. The runtime mutants are the eight
indexed fixtures listed earlier, three JSON fixtures above and the inherited
iterator mutant. No catch mutant is claimed.

Current commands and logs:

```sh
bash /tmp/stricter-options-stage3-current/stage3/apply.sh /tmp/stricter-options-adapted-current > /tmp/stricter-options-apply-current.log 2>&1
npm ci --prefix /tmp/stricter-options-adapted-current --ignore-scripts --no-audit --no-fund > /tmp/stricter-options-npm-current.log 2>&1
go run ./stage3/stricter-options/probe.go --production /tmp/stricter-options-adapted-current/src/compiler/checker.ts > /tmp/stricter-options-173-current.json 2> /tmp/stricter-options-173-current-probe.log
python3 stage3/stricter-options/validate.py /tmp/stricter-options-173-current.json /tmp/stricter-options-adapted-current --ledger-ref a1a16427 --ledger-file rows-2026-10-08.csv --include-json --save stage3/stricter-options/sites-173.json --disposition stage3/stricter-options/disposition.csv > /tmp/stricter-options-173-current-validation.log 2>&1
go test ./internal/load ./internal/lower ./internal/ir ./cmd/adamic ./stage3/stricter-options -count=1 -timeout 10m > /tmp/stricter-options-json-final-packages.log 2>&1
go vet ./internal/load ./internal/lower ./internal/ir ./cmd/adamic ./stage3/stricter-options > /tmp/stricter-options-json-final-vet.log 2>&1
python3 stage3/stricter-options/production_mutants.py > /tmp/stricter-options-json-production-mutants.log 2>&1
python3 stage3/stricter-options/mutants.py > /tmp/stricter-options-json-audit-mutants.log 2>&1
```

Additional proof logs: /tmp/stricter-options-json-runtime-final.log (first four
JSON fixtures), /tmp/stricter-options-json-cli.log, /tmp/stricter-options-json-filtered-oracle.log
and /tmp/stricter-options-library-oracle-final.log. The final package run includes
the fifth JSON observation fixture.

Catch bindings keep the project's declared type, as ruled. All five catch sites
remain errors. Lowering currently refuses non-Error throws and treats caught
instanceof Error as constant true. Converting these sites before a runtime
thrown-value representation can distinguish a non-Error would have no honest
failing witness. Recommended split: a caught-value representation and Error-use
check worker; this worker can continue indexed representations and site witnesses.

The October 8 to 9 estimate is conditional on that representation work and
supported indexed representations. It is not a verified delivery of all 106.
The first 50 ledger rows include record reads (for example D155/D156), so the
4 to 8 hour estimate covers the census and initial supported guards; completing
all 50 needs those representation gaps closed. Unsupported witnesses are still
explicit refusals, not checked counts. A firm whole-106 date remains dependent on
that work; no completion claim is made.

Latest whole-106 planning range communicated after the fresh reproduction: October 9 to 12 MDT, assuming caught-value representation work is split out and indexed representation gaps close. This supersedes the October 8 to 9 estimate. First 50: 4 to 8 hours for census and supported-case proof, with record reads still gating all-50 completion. All six audit mutants and all four loader mutants completed their reruns and were caught by assertions.


## Sparse array representation batch, October 7 at 21:35 MDT

The worker scope changed: four indexed workers own the 99 per-site witnesses;
this branch only supplies the representations they need. D155/D156 were
previously misclassified above as records. Inspecting the adapted transformer
shows `enabledSyntaxKindFeatures = new Array<SyntaxKindFeatureFlags>(SyntaxKind.Count)`:
these sites are holes, so sparse arrays were built first.

`new Array<T>(length)` now lowers when length is a statically proven valid
integer, with presence stored separately from zeroed payload. Presence travels
with the array across aliases. Ordinary and optimized integer indexed reads use
the existing bounds lookup plus its presence test; they do not add another
bounds check. Indexed writes within the length, push and fill create presence.
Reference payloads and the presence table use existing counted cleanup.
Both backends share the site guard. Compound numeric indexed updates now use
that same guard, after evaluating receiver and index once.

Nine representation fixtures cover missing number/string slots, an optimized
integer loop, a compound read, a partial fill through an alias, successful alias
writes and fills, allocated reference writes and observation of undefined.
Five erased-guard mutants all build and are caught by loss of the required
site-named exit 70. Number/loop/compound/partial-fill mutants exit 0; the string
mutant exits 1 under ASan. Each required guard reports indexed-presence=1 and
trusted=0; observing undefined reports zero guards.

Conservative assumption: a program containing sparse construction refuses
operations whose implementation still assumes dense arrays, even when a
particular operand is dense. This includes other array methods, array iterators,
for...of, spread, array-based library iteration and collection construction,
and JSON.stringify arrays. Eight boundary fixtures prove explicit refusals.
Dynamic/invalid lengths and omitted array-literal elements remain unsupported.
The existing immediate `new Array(length).fill(value)` dense construction is
preserved. Typed arrays and record reads are the next representation groups.

The production ledger rerun remains exact: 0 ordinary project errors; 173 sites,
99 indexed + 67 optional + 5 catch + 2 JSON; missing=0, extra=0, attribution=0.
72 sites remain checker errors on this branch. Deferred rows are not emitted
check counts and do not claim the whole compiler lowers.

Validation logs:

- `/tmp/stricter-options-sparse-witness.log`: all 21 indexed fixtures pass, including thirteen guard mutants.
- `/tmp/stricter-options-sparse-packages.log`: load, lower, ir and native pass; the new Set boundary fixture initially hit an unrelated statement refusal.
- `/tmp/stricter-options-sparse-boundaries.log`: all eight corrected boundaries pass.
- `/tmp/stricter-options-sparse-final-fixtures.log`: the complete representation/JSON fixture package passes after correction.
- `/tmp/stricter-options-sparse-vet.log`: vet passes for all six touched packages.
- `/tmp/stricter-options-sparse-ledger-validation.log`: exact 173-site reproduction above.
- `/tmp/stricter-options-sparse-filtered-oracle.log`: indexing, string indexing and RegExp result arrays.

Current origin/main 855d114e is merged as 736410fc. Main's compiler files did
not change in that merge. The complete repository gate was not run.
