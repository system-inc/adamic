# Validation and delivery

| Group | Added topics | Delivery commit |
|---|---|---|
| 1 | checked non-null; four namespace reads/receiver/class/callable fixes; ruled runtime drop | `1d60ab525` |
| 2 | conditions, 78 historical sites and 241 samples | `d40016bd9` |
| 3 | overloads, statements-small, element-access, syntax-kinds; refreshed counts | `e0e234ea2` |
| 4 | devirtualize-review coverage plus this conflict/validation report | commit containing this report |

Only compiler/area-stack was pushed. Each group is pushed once after its checks finish. No pull request, area/ push, main push, or merge of an unapproved topic branch. No changes were copied from cohere; checker dependencies use its pinned submodule.

## Toolchain

First setup used GOPROXY='https://proxy.golang.org|direct'. It failed with undefined usesNodeModules, nodeTypesIndex, starCollision and nodePrelude; the exact output is retained. Retrying setup on the stable stack succeeded. No load/compiler workaround was committed.

Successful timing lines (elapsed): Go .024s, Node .027s, submodules .061s, markdown .076s (step .008s), clang .180s, build 49.385s, test binaries deferred 49.521s, cache warm 49.522s, total 49.555s. nproc=5; CPU quota 4; memory 17.6 GB. Go1.27.1, Node24.19.0, clang20.1.8. Source /workspace/adamic-tools/env.sh. Node typings installed with npm ci --prefix stage3/api.

Initial parallel test/catalog attempts exhausted RAM and the /tmp tmpfs. Final runs use TMPDIR=/workspace/adamic-scratch, GOMAXPROCS=2, GOFLAGS=-p=1; catalog jobs=1 for group1 and jobs=2 thereafter. Earlier resource failures are not counted as successful mutant detection.

## Commands actually run

All test output was redirected directly to logs. Required package command after every group:

```sh
go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript
```

Counts after every group:

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

Group1 final counts ran alongside TestParserNamespace, TestParserCallableNamespace, TestModuleNamespace, and TestCheckedNonNull. Additional exact fixture/CLI/migration selectors are retained in the evidence. TestDecodeASCII -count=1 passed native (146.034s); its former link blocker is fixed by the ruled runtime-drop base. WASI was not enabled.

Group2:

```sh
go test ./internal/oracle -run 'TestTypeScriptConditionsAgreeWithNode|TestCondition|TestNativeAgreesWithNode/.*condition' -count=1 -timeout 15m
```

TestTypeScriptConditionsAgreeWithNode explicitly covers all 14 representations including the complete ledger; all three condition mutants are separate selected tests. The supplementary generic fixture pattern was not relied on for their coverage.

Final group3:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(overload_|census_overload_binders|element_access_|statements_small_|syntax_)|TestOverload|TestElementAccess|TestStatementsSmall|TestSyntax' -count=1 -v -timeout 20m
python internal/lower/testdata/overload_binders/run-mutants.py
python stage3/notyet-element-access/read_mutants.py
python stage3/notyet-element-access/read_mutants.py tuple
python stage3/notyet-element-access/read_mutants.py presence
python internal/oracle/testdata/run-statements-small-mutants.py
python /workspace/adamic-scratch/optional-index-mutants.py
```

The source mutants run in a detached checkout of the exact final compiler sources. Optional-index mutants use Go overlays; the overlay script is retained in evidence. Every source edit is restored between mutants. Inline syntax mutants run through the final oracle selector.

Group4 commands and its exact fresh/borrowing overlays are retained in evidence: the four required packages; go test ./internal/fresh -run '^TestMethodKeepsArgument$' -count=1 -v; go test ./internal/native -run '^TestDevirtualizeBorrowDocClaim$' -count=1 -v; ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/devirt_borrow_doc_claim' -count=1 -v; and the filtered fresh-refusal oracle. Catalog after groups1-3:

```sh
bash verify/catalog/check.sh 1d60ab525 --jobs 1
bash verify/catalog/check.sh d40016bd9 --jobs 2
bash verify/catalog/check.sh e0e234ea2 --jobs 2
```

The final group-four catalog uses the commit containing this report. Push output and final group-four catalog output are reported at completion; they cannot be embedded in their own commit.

## Limits

No full repository gate or full oracle package ran. No claim that native tsc/parser/scanner passed core.ts:11 or debug.ts:7. Enum-init and all other judgment/dependency topics stay held. See REPORT.md for every skipped topic, every observed conflict file and both sides, and the proposal for a ruling. Initial new-expression/for-of/representation trial controls were red and were removed from the delivered stack, rather than weakening their checks. No predicate, views-lane or map-keys topic was stacked. Paired integration merge/revert bookkeeping was excluded; only own topic commits were replayed.
