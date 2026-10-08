# Stricter checker options: verified audit, incomplete conversion

Built an in-process option audit in internal/load, preserving project options and lib inputs.
Commit: recorded in the worker's final report; branch codex/stricter-options-checks only.
Measured 0 project errors and all 171 exact stricter-option sites, with no missing or extra sites.
Six analysis mutants failed their intended assertions; no runtime-check mutants were run.
Production loading now uses one TypeScript project's options and lib and records sites; mixed ownership and runtime conversion remain unfinished.

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
names are retained; the primary order follows the ledger. No one of these 171
sites has been converted. The report describes diagnostics, not inserted checks.

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

The five catch rows in the actual rows.csv are commandLineParser.ts:2301 and
2952, program.ts:406 and 441, and sys.ts:1281. They are not all in sys.ts.
The 18:05 ruling is retained: their eventual .ts implementation must keep the
project's type and check the relevant uses, rather than blanket unknown typing.

All 67 optional rows remain errors in this loader slice. The user reports that optional-field-write-2 at 7e7464e6 now matches Node for in, Object.keys and Object.hasOwn across aliases. Merge that branch after the current steps and convert all 67 rows; this update authorizes the next implementation, not suppression in the loader slice.
The two JSON.stringify declaration-contract rows are outside this flag-only
audit. Their use-site checks and the .a diagnostic naming the fix are not built.
Production .ts loading now preserves the owning project's options and lib, with strictNullChecks and strictFunctionTypes retained as Adamic soundness rules. It keeps the existing prelude. No --explain-checks interface was added. There are zero
new converted checks of any kind. Planning range communicated October 7: October 8 to 9 by 21:30 MDT for supported indexed representations and catch/JSON, conditional on a second catch/JSON worker and a failing caught-value representation witness; it is not a verified completion date.

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

The full repository gate, native/oracle package gates, sanitizer runtime proofs,
and erase-runtime-check mutants were not run. No runtime or emitter is changed.
The production loader acceptance boundary is preserved conservatively until
runtime conversion is implemented and proven. This is an incomplete unit.

## Production loader slice

The production probe reproduces all 171 ledger sites and all ablation memberships:
99 indexed, 67 optional, 5 catch. All remain errors. The production checker also
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
Project overlays and explicit extra roots outside the config are refused until
the audit can attest to exactly those inputs. These conservative refusals are
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
