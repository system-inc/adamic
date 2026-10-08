Built: repaired graph adoption relocation and preservation of checked-view object metadata; pinned all eight reported stops.
Commits: starts at 348ba8665bffa1f5d16718a59f572650c8980423; integration ba59427ccc7afecae29a305c41e6e9c7867e5610 merged in ff2ebffcbc2400eb0412aec86371e0fb5520e51d; repair commit follows this report.
Commands and outputs: native adoption regression PASS; eight-location Node/native/sanitizer/JS frontier regression PASS; original 43-row graph guard has 40 matches and three named array-write blockers.
Mutants: stale-inline-map, overwrite-external-map, truncate-object-metadata all compile and fail TestShapeGraphMapAdoption's semantic assertion.
Not covered: the shared array source-contract blocker, the complete gate, checker ledger work, or a claim that Unknown proves nonconformance.

## What the previous report actually observed

GENERIC-FLOW-SHARE-REPORT.md points back to ADAPTED-STAGE3-REPORT.md. The eight stops in its integration baseline log are **runtime graph-region fixture aborts**, not eight unnamed allocation-to-cast graph edges. Each ended with `free(): invalid pointer` (exit -1). The same eight were reported on views-integration's older f1c91970 baseline. Assumption: step 1 authorizes the narrow graph-adoption runtime repair needed to make these named program locations reviewable. No executable flow edge was invented from a runtime abort.

All locations below are under `internal/oracle/testdata/`. The line identifies the implicated Map/Set allocation, except the final current frontier which identifies the checked array write.

| Program location | Earlier message | Current result |
|---|---|---|
| graph_regions_cache.a:5 | free(): invalid pointer | agrees with source Node |
| graph_regions_entries.a:7 | free(): invalid pointer | agrees with source Node |
| graph_regions_literals.a:11 | free(): invalid pointer | agrees with source Node |
| graph_regions_regression_06.a:4 | free(): invalid pointer | agrees with source Node |
| graph_regions_regression_07.a:5 | free(): invalid pointer | agrees with source Node |
| graph_regions_regression_08.a:5 | free(): invalid pointer | agrees with source Node |
| graph_regions_regression_09.a:5 | free(): invalid pointer | agrees with source Node |
| graph_regions_symbols.a:10; current frontier :4 | free(): invalid pointer | exit 70: `adamic: panic: element read failed: <array write> expected object, found uncertified source element contract` |

`TestShapeGraphReportedFrontiers` pins seven successes against source Node in release, sanitized native and JavaScript, with leak checks on successful native runs. For symbols it pins the specific refusal in all three execution modes and explicitly does not call that agreement with Node. Source Node succeeds there. The original oracle and count guard remain intact, so the outstanding discrepancy remains visible.

## Cause and repair

Graph adoption copies an allocation to storage with an ownership prefix, then frees the old allocation. Small Map/Set entries point inside that allocation. Preserve the fact that entries were inline before adoption and rebase only that pointer afterward. External buffers must retain their original address.

The current checked-view object layout also includes aligned unsigned field contracts. Older graph-adoption producers passed the earlier, shorter object size. After repairing the Map pointer, ASan exposed contract writes past the shortened copy. Normalize object adoption and counted size through `adamic_object_size`, preserving values, initialized flags, representations and contracts. This avoids edits to the shared emit/lower files.

The native harness checks counts 0 through 8 (both inline and external maps), entry contents, and all object metadata after adoption with a deliberately stale caller size. Three valid C mutants respectively remove rebasing, overwrite external pointers, and truncate the object copy; all are caught by semantic assertions, not compilation errors.

## Validation and remaining guard failures

```
go test ./internal/native -run '^TestShapeGraphMapAdoption$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestShapeGraphReportedFrontiers$' -count=1 -v -timeout 15m
python3 stage3/shape-conformance/graph-adoption-mutants.py
```

All pass. Logs are `logs/graph-adoption.log`, `logs/graph-frontiers.log`, `logs/graph-mutants.log`, and each `logs/graph-mutant-*.log`.

The broader uncached original run of `TestShapeGraphCountSnapshot` plus the eight original Node oracles and flow/parse shows no invalid-free abort or metadata ASan overflow. Forty of 43 count rows match. `graph_regions_flow.a`, `graph_regions_parse.a`, and `graph_regions_symbols.a` stop at the named array-write source-contract refusal above and consequently have fewer count events. Do not update the guard's expected counts to hide those failures. The original run is retained in `logs/graph-original-guard.log`.

## Toolchain

`export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`.

Timing lines: node ready 0.061s; Go ready 0.079s; clang ready 0.506s; markdown install step 0.990s, ready 1.161s; submodules ready 216.195s; Go build ready 509.965s; build cache warm 510.131s; done 510.188s. `nproc` = 5; cgroup cpu.max = 400000 100000, 17.6 GB. An early test before submodules completed failed for the missing cohere/TypeScript/tsc/go.mod; setup resolved it. Exact cohere and nested TypeScript pins were retained; no cohere file was copied.
