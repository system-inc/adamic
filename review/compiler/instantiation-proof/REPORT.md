Delivered census deferral and five contextual oracle controls toward Outcome 16; the requested production proof unit remains incomplete.
Implementation a97e7b32 on base 55347dce; the final evidence commit is identified in the final response.
Six top-level oracle tests pass in 0.881s; counts refresh passes in 49.867s; TestCallTargetReaders passes in 36.110s; selected real-core reasons change 23 to 0.
Three mutants are caught: premature generic lowering by the status assertion, missing inventory by the presence assertion, and erased present return by Node disagreements in both backends.
Not delivered: instantiation-named union refusals, the two requested instantiation-proof mutants, an 81-file closure rerun, or a completed optional-array caller replay.

The change is a measurement correction, not a new compiler admission. No production compiler, backend or representation changed. Compiler admission delta against main over internal/oracle/testdata is exactly zero because those compiler sources are unchanged. All five new fixtures were accepted by the compiler built before the census edits and agree with source Node. There is no newly admitted disagreement to put on the first line.

Observed cause: the census bypasses generic registration and monomorphization when it attempts a standalone generic unit. Its old lower.go.txt comment explicitly said “Attempt generic declarations as written, without invented instantiations.” Both first-error and full mode then called lowerFunction directly. Production modules.go instead registers generics, and generic.go instantiates them at calls. inferTypes already passes the caller's concrete mapping through another generic and recovers optional binders and callback types. The pinned census predates the current optional-binder recovery in 6b33ec61.

The largest established cause is fixed in the measurement paths. They keep standalone generic units inventoried with status deferred_generic_instantiation and reach their executable bodies through concrete callers. Full mode also defers standalone nested units whose lexical types depend on a generic ancestor. Deferred means unmeasured without an instantiation witness; it never means admitted. Concrete dependencies still use the existing instantiateFunction path. The caller-control census has five inventoried units, two deferred generics, an attempted nongeneric caller and no findings. This correction also suppresses other diagnostics formerly measured inside those independent bodies. It must not be interpreted as hundreds of production gaps closing.

The real-source evidence uses the pinned TypeScript v6.0.3 tree prepared by stage3/apply.sh. No implementation was copied from cohere. The five runtime fixtures were independently written from inspected contracts and retain optional parameters, optional locals, callbacks, generic forwarding or constraints. The scout source-sites.json was fetched from 8fbfddb8, and the census scripts and results from 807d65d9 were inspected.

| Selected site | Observation on current main census | Runtime control |
| --- | --- | --- |
| core.ts:1083 firstOrUndefined | standalone signature says returning T \| undefined without a mapper | another generic forwards its T; object and scalar instantiations, present, empty and undefined arrays |
| core.ts:1116 lastOrUndefined | same standalone uninstantiated signature failure | generic forwarding under a reference interface constraint; present, empty and undefined arrays |
| core.ts:2412 findBestPatternMatch | same standalone uninstantiated signature failure | T reached through a callback, an optional generic local, present and absent matches |
| core.ts:33 forEach | standalone signature says returning U \| undefined without a mapper | U inferred through callback results, reference results present or missing |
| core.ts:22 length | its standalone unit has no finding on current main | readonly T[] \| undefined forwarded through another generic; an unused generic contributes no body lowering |

The fifth scout occurrence is contextual: core.ts:22:27 under checker.ts:5454:5. A scratch selection overlay attempted that real caller with a hard 90s limit, but did not complete a file record. Consequently the readonly-array cause in that real caller is not established here. The positive reduction rules out generic forwarding alone as a sufficient explanation; it does not rule out missing inference evidence in the full checker context. The interface-constrained runtime control succeeds, so a reference interface constraint alone is likewise not a cause of refusal in that control.

The paired census loads the same current adapted compiler-directory roots and measures only real core.ts in full mode. It uses the current descendant of the fetched census driver through stage3/census/latent/make_overlay.py, not the old compiler hooks transplanted unchanged. This is a core-file replay, not a rerun of the 81-file closure or a direct delta from the old RESULT.json. Both runs report 321 checker diagnostics, 267 units and the measurement-only output guard. Before has 266 attempted and one checker-body split unit; after has 63 attempted, 203 deferred and one split. Total findings change 775 to 147.

| Raw core NotYet reason | Before | After |
| --- | ---: | ---: |
| function returning T \| undefined | 10 | 0 |
| function returning U \| undefined | 4 | 0 |
| value of type T \| undefined | 6 | 0 |
| value of type U \| undefined | 2 | 0 |
| value of type readonly T[] \| undefined | 0 | 0 |
| value of type T \| T[] \| readonly T[] \| undefined | 1 | 0 |

Current main also admits the scalar-union witness union.a using its existing boxed Union path. Source Node, emitted JavaScript and native with ASan, UBSan and leak detection all print 0, ss, missing on separate lines. A proposed number \| string refusal is therefore an intentional restriction of current admission, not evidence of a missing representation. This delivery conservatively preserves that existing behavior and does not claim the requested refusal or skip-proof miscompile witness. The present-return mutant is different from either requested instantiation-proof mutant.

Exact final checks, each redirected to its retained log:

```text
source /workspace/adamic-tools/env.sh
unset GOCACHEPROG
timeout 90 go test -p 2 ./internal/oracle -run '^TestInstantiation' -count=1 -v -timeout 90s
ok github.com/system-inc/adamic/internal/oracle 0.881s
timeout 120 go test -p 2 ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 49.867s
timeout 120 go test -p 2 ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
ok github.com/system-inc/adamic/internal/ir 36.110s
```

Final leaves in seconds: Find 0.61, First 0.55, Last 0.56, Callback 0.57, Array 0.38, PresenceMutant 0.49. Their first uncached fixture run was 11.83 to 11.88s per leaf. Each ordinary fixture compares source Node with emitted JavaScript, sanitized native and release native, and checks leaks. Counts gains exactly five rows, with no changes to existing rows. No package-wide run or full gate was run.

The census commands were:

```text
python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/instantiation-overlay-before
go build -p 2 -overlay=/tmp/instantiation-overlay-before/overlay.json -o /tmp/instantiation-census-before ./stage3/census/latent/tool
# Repeat after edits with overlay-after and census-after.
LATENT_FULL=1 LATENT_ONLY_FILE=/tmp/instantiation-adapted/src/compiler/core.ts LATENT_ASSERT_NO_OUTPUT=1 timeout 180 /tmp/instantiation-census-before /tmp/instantiation-adapted/src/compiler /tmp/instantiation-core-before.jsonl
# Repeat with census-after and core-after.jsonl.
python3 review/compiler/instantiation-proof/census-check.py /tmp/instantiation-core-after.jsonl
# exits 0: generic inventory retained; independent bodies deferred
python3 review/compiler/instantiation-proof/census-check.py /tmp/instantiation-core-before.jsonl
# exits 1: premature-lowering mutant, status is attempted
python3 review/compiler/instantiation-proof/census-check.py /tmp/instantiation-inventory-mutant.jsonl
# exits 1: firstOrUndefined missing from inventory
```

The inventory data mutant deletes only the firstOrUndefined unit from the after JSON. The reversed census source patch restores the premature-lowering behavior. PresenceMutant changes exactly one firstOrUndefined object specialization's return to undefined in real lowered IR; both backends execute cleanly with stdout differing from Node and no leak. This mutant is not killed by a build warning or a sanitizer failure.

Setup: GOPROXY was set to https://proxy.golang.org|direct before bash cloud/setup.sh. The setup command hit its 240s limit while warming the build cache. Printed readiness lines: Go 0.031s; Node 0.031s; clang 0.257s; Markdown dependencies 0.984s; submodules 17.146s; shared cache 23.379s. There was no warm-cache-ready or done line. nproc is 5. /opt/adamic-tools/env.sh does not exist; the generated /workspace/adamic-tools/env.sh was sourced instead. Cold compiler, census and oracle build commands reached their hard limits; local-cache retries with GOCACHEPROG unset and -p 2 succeeded. Initial counts refresh failed on missing pinned @types/node. npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund installed the locked dependencies, and counts refresh then passed. These failures are retained rather than called green.

Remaining pieces, each a separate unit: implement the requested conservative production instantiation proof and named refusals, explicitly accounting for current boxed-union admission, with the two exact requested mutants; complete the real readonly-array caller replay and repair any remaining concrete-binder inference failure; then refresh the 81-file closure census with deferred generic coverage reported separately from admitted executable instantiations. This delivery lands evidence and a measurement correction toward Outcome 16, not completion of its production proof requirement.

Integration lane checks on the committed branch exit 0: lane checks 6.4 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; a-check 2 .a files; vet 1 packages. The repository fetch configuration initially left the integration branches only in FETCH_HEAD; explicit remote refspecs populated origin/devtools/fast-gate and origin/cloud/merge-tree before the prescribed checker command. Origin/main remains 55347dce, the branch base. A production-source diff against main for internal/lower, internal/native, internal/javascript and internal/load is empty.
