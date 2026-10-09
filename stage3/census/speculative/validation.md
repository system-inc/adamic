Current delivery is an audited partial directory census, not a completed whole-compiler speculative census. README.md and RESULT.json name the four covered directories and every unexamined file. The historical whole-corpus runs below are provenance only; their unverified depths do not contribute to these tables.

The final measurement base is `a5630a90d05abe85670ba1e00bff3643a2742e53`, merging user-requested checker mapper fix `69501280a81259fb512edbb8dd0e52c6eb0d88c8` into the dispatched `ed6e2975` candidate. `origin/main` did not contain the dispatched candidate. TypeScript 6.0.3 is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8`; `bash stage3/apply.sh /tmp/speculative-adapted` supplied its adaptations. RESULT.json inventories the measured source hashes. This unit's own changes are entirely within stage3/census/latent/ and stage3/census/speculative/. Earlier pre-fix census totals are superseded and excluded from the final evidence.

Before setup: `export GOPROXY='https://proxy.golang.org|direct'`. An initial setup overlapped correction of the selected base and failed with undefined `usesNodeModules`, `nodeTypesIndex`, `starCollision` and `nodePrelude`. Retrying `bash cloud/setup.sh` on the stable candidate passed. Its timing lines were Go ready 0.021s, Node ready 0.023s, submodules ready 0.055s, markdown ready 0.071s, clang ready 0.157s, Go build ready 48.000s, cache warm 48.100s, total 48.132s. `nproc` was **5**, with cgroup quota **4 CPUs**. Tools: Go 1.27.1, Node 24.19.0, clang 20.1.8; sourced `/workspace/adamic-tools/env.sh`. Both setup logs are preserved in evidence/.

The speculative and no-stubs measurements use `/tmp/speculative-census-final`, rebuilt after the mapper-fix merge. The running full baseline uses the preceding immutable-copy binary from the same base. Later changes recover binding positions only when speculation is enabled; whole-corpus equality with the final binary's no-stubs mutant proves the full path stayed unchanged. Executable hashes are in evidence/manifest.json. The managed environment restarted during the prior measurements, terminating the running processes without an OOM event. The full baseline preserves its 67 completed file records and resumes its remaining 12 using a scratch-only file filter over the same complete checker project. The assembled header matches exactly, and the inventory is sorted, unique and complete. Whole-corpus no-stubs equality checks every retained and resumed record. The final speculative and no-stubs runs are each fresh detached single-process walks. Speculation starts with `GOMEMLIMIT=3GiB`; baseline and no-stubs use `2GiB`. At 18:38 UTC, a live-heap sample showed 1,973,138,352 bytes under the 3 GiB cap. The census process's Go runtime memoryLimit was raised to 6 GiB while stopped under Delve, then it was detached and resumed; the setting and memory-event logs are preserved. This changes runtime collection pressure, not AST or IR data. One earlier run lost its execution-service session, and a later pre-fix run was explicitly stopped to incorporate the mapper fix. Neither incomplete run supplies these results. Immutable scalar and struct-field snapshot copying avoids repeated reflective field walks while retaining fresh slice storage; full and no-stubs keep the original copy semantics. Mode checks are cached once per snapshot invocation in every mode; the controls toggle modes between invocations and compare values and ownership.

The requested deadline, 2026-10-08 13:00 MDT (19:00 UTC), passed before the speculative corpus walk finished. No partial unit was pushed at the deadline.

The completed no-stubs run is **byte-identical** to the complete baseline across all 79 files: **19,168 distinct sites, 10,482 NotYet and 8,686 Refused**. Its equality validates both the 67 retained and 12 resumed baseline file records. The initial whole-project depth audit failed; the corrected whole-project speculative run was interrupted before buffered records were emitted. Only the new directory projects supply verified tables.

| Check or mutant | Observation and check that caught it |
|---|---|
| Nested with/with/debugger | NotYet and Refused each have outer with depth 0, inner with depth 1, inner debugger depth 2, later sibling debugger depth 0. Eight sites, no unvisited nodes. Stock TypeScript matches all eight depths. |
| Checker-clean nested signature | Outer generic signature 0, failing call 1, arrow child 2, captured value 3. Five NotYet, including the hidden parameter at depth 1, versus one full, no checker diagnostics. Stock TypeScript matches every depth. |
| Imported generic-body recovery | Two findings discovered while attempting main.a retain depths 1 and 2 after target.a fails its own signature later. Stock TypeScript matches all nine raw records. main.a's visited set includes three dependency nodes beyond its 17 own nodes; no own nodes are unvisited. |
| No-stubs mutant | Each nesting control rejects restored full output; whole-corpus comparison requires byte-identical full JSON, unchanged checker diagnostics and spans, and a lower site count. |
| Depth-zero mutant | Each of the three nesting controls rejects incorrect inner depths. |
| Shared-token identity mutant | The async fixture deliberately has an invalid async return annotation so its generic signature retains a failing boundary after finer local recovery. Async root incorrectly moves from depth 0 to 1 while await stays at depth 1. Exact AST identity assertion rejects it. |
| Bypass binding lowerer mutant | Checker-clean signature control rejects loss of its hidden parameter failure. |
| Omit scalar struct values mutant | Legacy-value comparison rejects the differing snapshot. |
| Shared scalar-slice mutant | Snapshot ownership check rejects mutations leaking back into source locals, parameters or environments. |
| Unrebound function-cursor mutant | Snapshot test rejects its cursor still pointing into the original program. |
| Trust-private-scalar mutant | Private-field guard witness rejects silently copying a private field. |
| Fresh checker-mapper mutant | Snapshot test rejects replacement of the checker-owned mapper identity. The merged mapper-alias regression test also passes. |
| Non-nil measurement IR mutant | Explicit guard panics with `measurement returned usable IR`. |
| Production loader exposes measurement input mutant | Explicit guard panics with `measurement loader exposed an output program`. |
| Overlay off and flag-on production | Clean merged-base and working-tree normal builds have byte-identical C and JS stdout, stderr and exit codes on unchanged functions.a. Output hashes are in manifest.json. |
| Leaked-stub-output artifacts | Appending a comment to emitted C or JS fails each byte comparator. These are comparator mutants, not production edits. |
| Existing planted-site and attribution mutants | Synthetic NotYet adds exactly one to one.a and none to two.a; incorrect attribution is rejected. |
| Existing signature/body-range mutant | Legacy audit rejects classifying signature diagnostics as body diagnostics. |
| Existing first-error and retained-state mutants | Actual rollback witness `reading poison` disappears. Compatibility verifier rejects both and requires all observations to equal the untouched merged-base overlay. |
| Stock ancestry depth-zero artifact mutant | Independent stock AST ancestry rejects the changed inner depths. |
| Stock version and source-hash artifact mutants | Pinned-version and source-hash assertions reject both artifacts. |
| Checker diagnostic and span artifact mutants | Whole-mode comparator rejects each changed checker observation. |
| Headline depth-count artifact mutant | Independent result recount rejects the added site. |
| Top-twenty depth-count artifact mutant | Independent ranking/depth recount rejects the altered count. |
| Omitted-source artifact mutant | Independent directory inventory rejects the missing file. |
| Examined-byte artifact mutant | Independent denominator/share check rejects the changed byte count. |

The inherited `full_audit.py` fails on both untouched merged-base overlay and this overlay: it expects string and number conditions to be refused, while this candidate accepts them. Its three real observations are preserved. `verify_full_compatibility.py` compares all three against the unchanged base and checks the actual candidate rollback witness against both mutants. No production change was made to satisfy the stale assertion.

Commands run; every stdout/stderr stream was directed to a named log:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/speculative-final-overlay
go build -buildvcs=false -overlay=/tmp/speculative-final-overlay/overlay.json -o /tmp/speculative-census-final ./stage3/census/latent/tool
python3 stage3/census/speculative/audit_binding.py "$PWD" /tmp/speculative-final-overlay /tmp/speculative-census-final /tmp/speculative-binding-proof
python3 stage3/census/speculative/audit_snapshot.py "$PWD" /tmp/speculative-final-overlay /tmp/speculative-binding-snapshot-final
python3 stage3/census/speculative/audit.py /tmp/speculative-census-final /tmp/speculative-binding-control
python3 stage3/census/speculative/audit_signature.py /tmp/speculative-census-final /tmp/speculative-binding-signature
python3 stage3/census/speculative/audit_dependency.py /tmp/speculative-census-final /tmp/speculative-binding-dependency
python3 stage3/census/speculative/prove_identity.py "$PWD" /tmp/speculative-final-overlay /tmp/speculative-census-final /tmp/speculative-binding-identity-final
python3 stage3/census/speculative/isolation.py "$PWD" /tmp/speculative-fixed-isolation
python3 stage3/census/latent/audit_output_guards.py "$PWD" /tmp/speculative-final-overlay /tmp/speculative-binding-guards
go test ./stage3/census/latent/statecopy -run '^TestCheckerMapperAliasKeepsIdentity$' -count=1
python3 stage3/census/latent/audit.py /tmp/speculative-census-final
python3 stage3/census/latent/full_audit.py /tmp/speculative-census-final /tmp/speculative-binding-full-control
# Original latent templates archived from a5630a90, with original module files.
python3 /tmp/speculative-fixed-original/stage3/census/latent/make_overlay.py "$PWD" /tmp/speculative-fixed-original-overlay
go build -buildvcs=false -overlay=/tmp/speculative-fixed-original-overlay/overlay.json -o /tmp/speculative-fixed-original-census ./stage3/census/latent/tool
python3 stage3/census/latent/full_audit.py /tmp/speculative-fixed-original-census /tmp/speculative-fixed-original-full-control
python3 stage3/census/speculative/verify_full_compatibility.py /tmp/speculative-binding-full-control /tmp/speculative-fixed-original-full-control
# Baseline: 67 completed file records plus the scratch resumed suffix.
GOMEMLIMIT=2GiB LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_RESUME_AFTER=/tmp/speculative-adapted/src/compiler/transformers/taggedTemplate.ts /tmp/speculative-census-resume /tmp/speculative-adapted/src/compiler /tmp/speculative-full-suffix.jsonl
GOMEMLIMIT=2GiB LATENT_SPECULATIVE=1 LATENT_MUTANT_NO_STUBS=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/speculative-census-final /tmp/speculative-adapted/src/compiler /tmp/speculative-complete-no-stubs.jsonl
GOMEMLIMIT=3GiB LATENT_SPECULATIVE=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/speculative-census-final /tmp/speculative-adapted/src/compiler /tmp/speculative-complete.jsonl
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/census/speculative/stock.cjs /tmp/speculative-adapted/src/compiler /tmp/speculative-complete.jsonl /tmp/speculative-complete-stock.json
python3 stage3/census/speculative/report.py /tmp/speculative-adapted/src/compiler /tmp/speculative-complete.jsonl /tmp/speculative-complete-stock.json stage3/census/speculative
python3 stage3/census/speculative/verify.py /tmp/speculative-complete.jsonl stage3/census/speculative/RESULT.json /tmp/speculative-adapted/src/compiler
python3 stage3/census/speculative/audit_report_controls.py /tmp/speculative-binding-control/source /tmp/speculative-binding-control/speculative.jsonl /tmp/speculative-binding-control/stock.json /tmp/speculative-binding-control/full.jsonl /tmp/speculative-binding-control/no-stubs-mutant.jsonl /tmp/speculative-binding-report-guards
python3 stage3/census/speculative/compare.py /tmp/speculative-complete-full.jsonl /tmp/speculative-complete-no-stubs.jsonl /tmp/speculative-complete.jsonl
```

`audit_snapshot.py` runs only `^TestLatentSpecSnapshotScalarCopy$` under a scratch overlay, not the whole lower package. Stock/report commands were additionally run on all three small nesting controls. `reproduce.sh` builds one current binary and runs the controls and three corpus modes into a new output directory; source the tool environment and set NODE_PATH first. Rendering there preserves this checked-in README.

No whole-package test, full gate, final module ordering, ownership/freshness pass, native execution or semantic oracle was run. The six `.a` census witnesses are outside the native oracle corpus; this unit adds no native fixtures. Its local fixture ledger is [counts.md](counts.md). [Method and limits](method.md) explains context-sensitive continuation findings and byte coverage.

Partial delivery observation: all 79 speculative files completed at 22:27 UTC, with 52,596 raw distinct NotYet and 13,650 Refused sites. The stock AST audit failed on binder.ts:926:17 (reported depth 1, stock depth 0). No final depth table is verified. No traversal files are unfinished; every file in depth-audit-pending.txt remains pending for the independent depth audit. The automatic publisher failed before committing. Follow-up authorization requests this partial commit and a second final-results commit.

The independent audit diagnosed missing boundary evidence after finding deduplication, span-only collisions between parameters and their identifiers, and inconsistent source attribution across attempts. The corrected mode keeps a complete typed boundary catalogue. Speculative finding identity includes the exact source AST kind and byte span, distinguishing constructs with the same printed error location and reason; full/no-stubs retain the original identity. audit_topology.py adds a checker-clean eight-site witness: seven NotYet, one Refused. Call, array and element-access failures sharing the same printed reason/location remain three sites at depths 2, 4 and 1; the equal-span identifier is at depth 4. Omitted-boundaries, legacy-site-key and span-only matching mutants are all caught. The existing nested, signature, imported-body, exact-token identity and binding controls retain their expected counts and depths.

The interrupted corrected whole-compiler foreground census used GOMEMLIMIT=6GiB, GOGC=200, GOMAXPROCS=4; its completed no-stubs mutant used 4GiB, GOGC=200, GOMAXPROCS=2. Both ran /tmp/speculative-census-fast, with unchanged production source base a5630a90. Mode metadata changes do not touch the full lowerer path. The six .a witnesses form five projects, counted in counts.md. Earlier raw depths and the earlier six-site equal-span prototype are superseded. The gate fix 20ebb633 added exact expected-error headers to the four checker-error fixtures; their bodies are unchanged and the nested line assertions account for the inserted first line. Header validation is recorded separately in header-validation.md.

Corrected binary and focused validation commands (stdout/stderr saved to matching named /tmp/speculative-fast-* logs). The whole-compiler speculative invocation was interrupted. The subsequent whole-compiler stock/report/verify/compare commands were planned but never completed, and are omitted here:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/speculative-fast-overlay
go build -buildvcs=false -overlay=/tmp/speculative-fast-overlay/overlay.json -o /tmp/speculative-census-fast ./stage3/census/latent/tool
python3 stage3/census/speculative/audit_snapshot.py "$PWD" /tmp/speculative-fast-overlay /tmp/speculative-census-fast-snapshot
go test -overlay=/tmp/speculative-census-fast-snapshot/overlay.json ./internal/lower -run '^$' -bench '^BenchmarkLatentSpecFunctionSlice$' -benchtime=1x -count=1
python3 stage3/census/speculative/audit.py /tmp/speculative-census-fast /tmp/speculative-fast-control
python3 stage3/census/speculative/audit_signature.py /tmp/speculative-census-fast /tmp/speculative-fast-signature
python3 stage3/census/speculative/audit_dependency.py /tmp/speculative-census-fast /tmp/speculative-fast-dependency
python3 stage3/census/speculative/prove_identity.py "$PWD" /tmp/speculative-fast-overlay /tmp/speculative-census-fast /tmp/speculative-fast-identity
python3 stage3/census/speculative/audit_topology.py "$PWD" /tmp/speculative-fast-overlay /tmp/speculative-census-fast /tmp/speculative-fast-topology
GOMEMLIMIT=6GiB GOGC=200 GOMAXPROCS=4 LATENT_SPECULATIVE=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/speculative-census-fast /tmp/speculative-adapted/src/compiler /tmp/speculative-fast-complete.jsonl
GOMEMLIMIT=4GiB GOGC=200 GOMAXPROCS=2 LATENT_SPECULATIVE=1 LATENT_MUTANT_NO_STUBS=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/speculative-census-fast /tmp/speculative-adapted/src/compiler /tmp/speculative-fast-no-stubs.jsonl

```

The benchmark executes no tests (`-run '^$'`). The snapshot auditor executes only `^TestLatentSpecSnapshotScalarCopy$`, including its subtests and six mutants. Header checking invokes the unchanged Gate.aCheck method; the final in-place scan includes all six .a controls, with four matching type-error headers and two NotYet outcomes, zero mismatches and eight rejected header predicate mutations. All speculative test binaries assert unusable measurement IR and loader output. The fast copy has byte-identical complete observations to the preceding physical-site binary on the combined six-witness project.

Current foreground partial delivery commands, with stdout/stderr captured in named logs:

```sh
source /workspace/adamic-tools/env.sh
# Run independently for each of _namespaces, factory, transformers/declarations, transformers/module.
GOMEMLIMIT=2GiB GOGC=200 GOMAXPROCS=1 LATENT_SPECULATIVE=1 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/speculative-census-fast /tmp/speculative-adapted/src/compiler/DIRECTORY /tmp/speculative-directory-results/DIRECTORY/speculative.jsonl > /tmp/speculative-directory-results/DIRECTORY/speculative.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" GOMEMLIMIT=2GiB GOGC=200 GOMAXPROCS=2 python3 stage3/census/speculative/partial.py /tmp/speculative-adapted/src/compiler /tmp/speculative-directory-results stage3/census/speculative /tmp/speculative-census-fast > /tmp/speculative-partial-report.log 2>&1
python3 /tmp/speculative-a-check/run.py > /tmp/speculative-a-check/partial-final-run.log 2>&1
```

partial.py runs stock.cjs, report.py and verify.py for each completed directory. It invokes full and no-stubs modes on exactly the same directory project, requires byte-identical JSON and unchanged checker diagnostics/spans, and independently recounts the combined table, top twenty, source hashes, covered inventory and whole-compiler byte denominator. _namespaces has no findings, so its top-twenty mutant is explicitly inapplicable; its other three artifact mutants pass. Other directories exercise all four artifact mutants. Six additional partial-report mutants corrupt a headline count, a top-twenty depth count, the covered file inventory, byte coverage, an unexamined file and the claimed directory scope. All must be rejected. An attempt to render before factory completed was rejected as an unfinished run.

The census-only fast function-slice copier copies scalar fields by value, deeply copies all six mutable fields and rejects unknown/private mutable schema fields. Its focused snapshot test covers nested bodies, frame environment, reference parents, optional parameters, function cursor and mapper identity. The schema-skip mutant is rejected. Complete observations on the combined six-witness project match byte-for-byte before and after the optimization. The one-iteration 10,000-function copy benchmark reports legacy 7,510,382 ns and speculative 545,853 ns; this is not a whole-corpus throughput result. Full/no-stubs retain the original copier.

Evidence logs are losslessly gzip-compressed; their original byte lengths and SHA-256 hashes are in evidence/manifest.json. No detached census is relied on for this delivery. The remaining whole-compiler sources are explicitly unexamined in RESULT.json, rather than extrapolated from the measured subset.
