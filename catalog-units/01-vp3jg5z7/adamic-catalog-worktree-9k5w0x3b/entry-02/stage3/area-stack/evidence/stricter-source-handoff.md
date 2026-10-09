Built a non-merge stricter-options transplant on compiler/area-next 337aa466, preserving area representations and refusing missing optional presence.
Commit provenance is recorded in TRANSPLANT.csv; the final branch SHA is supplied in the handoff.
The exact 79-root production run reports 173 IDs: 106 scheduled, 67 named remaining errors, and zero ordinary errors.
Six attribution mutants, four loader mutants, and supported indexed, holes, typed-array, record, JSON and caught-value runtime mutants are caught.
173 scheduled / 0 errors and the D069 runtime witness are blocked by dependencies absent from the area; whole-compiler emission and the repository-wide gate are not claimed.

Base: 337aa466bf7b519ffdc50c6f16e976a330a4a4f5. Branch: codex/stricter-options-next.
No merge commits were imported. The three protected emitter/lowerer files and internal/oracle/oracle_test.go are unchanged from the area base.

The selected work includes project options and lib, stricter-site attribution, indexed guards, holes, six additional numeric typed-array kinds, readonly records, nullable strings, JSON use checks, caught values, and all four indexed witness slices. Existing area Uint8Array/Int32Array/Float64Array storage, methods and checked reads remain intact. Their presence guards reuse the existing single lookup. Area predicate explanations, namespace behavior, overload adapters, regexp callback facts and uninitialized fields remain intact.

The loader retains a project's Node declarations rather than adding its standalone pinned Node package. Project files receive their selected lib without standalone Set extensions. Global console discovery preserves a host console and supplies the Adamic declaration only when the host has none. Attribution checks the same source overlay and explicit roots as production. Ordinary diagnostics continue to reject compilation. Collection iterator facts from isolated library commit d5b7660f were required by the exact ledger and included without importing that branch's history.

| Contract | Scheduled | Remaining errors |
| --- | ---: | ---: |
| Indexed presence | 99 | 0 |
| Caught use | 5 | 0 |
| JSON.stringify result | 2 | 0 |
| Optional presence / implements relation | 0 | 67 |
| Total | 106 | 67 |

Every optional row names the missing own-presence representation: 5bb775ca plus alias/presence fixes through 7e7464e6. These prerequisites are absent from the area. The area Object.hasOwn implementation recognizes shape fields and has no separate absent/present-and-undefined state. In accordance with the instruction to name missing dependencies instead of importing their branches, optional guard commits 58e743bf, 9af86993, 59ca7f85, 6db50e57, 4b1922bd, a022b1e1 and 8c81fd11 were not admitted. Their expression cleanup dependency a593d04e also needs reconciliation when those guards are transplanted.

D069 likewise writes into an initially absent optional field. On this area, its valid present input otherwise reaches a missing-slot panic. Lowering now refuses that destination with the same named dependency. Its Node present and absent observations and the CLI refusal are verified; it is not counted as an emitted check. The other 98 indexed ledger IDs have runtime witnesses and guard-erasure proofs. The census's 99 scheduled indexed contracts are obligations, not a claim of 99 emitted whole-program guards.

production.json.gz and states.csv replace the old branch's historical census with this verified result. The original ledger IDs, coordinates and diagnostic codes are unchanged. All 79 source-hashes.json entries match the adapted tree. The ledger is a1a16427, rows-2026-10-08.csv, retained as rows.csv. Missing, extra, duplicate or changed identities reject the census; there are no diagnostic exclusions.

Mutant coverage: six option-attribution overlays (index flag, optional flag, catch flag, ordinary-error membership, project lib, and site position); four production overlays (project options, recorded sites, optional waiver, and mixed-file refusal); per-witness indexed panic erasure; record and array guard erasure; prototype guard erasure; JSON guard erasure for undefined, function and held results; caught-value carrier/string/member/null-tag mutants; nullable sentinel and Map-key runtime review mutants. Every runtime mutant must compile before its differing exit, named message or Node observation catches it. D037's indexed mutant targets its exact message constant and retains the area's additional overload guard.

The combined explain fixture reports indexed-presence=1, catch-error=0, json-stringify-defined=1, optional-write=0, caught-type=1, trusted=0. Predicate-only area output remains unchanged. Per-site witness logs also assert actual guard counts rather than scheduling totals.

Validation logs are saved under evidence/area-next. The final evidence manifest records commands and their package outcomes. The full internal/load, internal/lower and cmd/adamic packages pass; all four indexed slices, stricter-options and stricter-records suites pass, with D069 explicitly blocked. Filtered TestCaught oracle tests pass. The flow and fresh packages, indexed-all routing, go vet, the host24 Node oracle, and the host Error-brand erasure mutants pass. The affected native closure-owner, record, and Map-hash tests pass. The full native package retains TestDecodeASCII/native's linker failure: missing adamic_write_raw, adamic_share, adamic_stack_thread_start and adamic_heap_thread_end. Exporting the exact pinned area runtime and running that test reproduces the failure; those owner implementations were not imported. The full repository oracle/counts gate and optional guard suites were not run because their presence prerequisite was not imported.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Timing seconds: Node 0.041, Go 0.050, Markdown 0.136, submodules 0.140, clang 0.311, Go build 65.470, deferred test binaries 65.665, cache warm 65.667, total 65.711. nproc=5; CPU quota=4; memory=17.6 GB. Node=24.19, Go=1.27.1, clang=20.1.8.

Host error producers now use a known code-bearing Error shape and the heap exception carrier. Node witnesses and compiled brand-erasure mutants verify this without changing object layout. Parallel owner changes only update four opaque pointer types to the carrier ABI; scheduler and ownership branches are unchanged.

Automatic review rejected a combined cross-layer edit, a broader parallel pending-state rewrite, and a shared object-layout change. They were replaced by explicit tested hunks, ABI-only pointer changes, and the existing shape/factory pattern, respectively. No rejected action was executed, and no approval is outstanding.
