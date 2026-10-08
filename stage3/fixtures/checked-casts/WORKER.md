Current correction: see [PROVENANCE.md](PROVENANCE.md) for separate read-stop diagnostics, cast-site provenance, the authorized verifier update and complete observer controls. Earlier delivery below is historical.

Current delivery: **all 18 formerly blocked contracts pass; all 20 acceptance fixtures pass** on committed compiler 1af953e6. The unchanged strict verifier is green. Error stays pinned to library #ddwcejg. See [DELIVERY.md](DELIVERY.md) for every contract, command, mutant and limitation.

The following sections retain earlier measurements as historical milestones.

# Scanner cast acceptance measurement

Acceptance inputs copied from ea1b2359 without merging its branch. Compiler base 561166d1, checked-views integration 432d4913. The ledger was fetched as an immutable object. Error remains pinned to library #ddwcejg; no library facts are fabricated.

Fresh baseline: all 20 Node goldens pass. Five of 20 runtime contracts pass: 01 positive and failing, 02 positive and failing, and 03 positive. Of the 18 contracts blocked on the acceptance-author branch, three pass here. 03 failing stops 70 at the specified read with the right stdout but uses the existing `field read failed:` diagnostic instead of the required `cast failed:` prefix. 04 positive and failing refuse the backing string/number field conversion during the view scan. 05 through 10 positive and failing refuse unchecked casts.

The supplied observer aborts at 01 failing because it only recognizes `adamicCast`; this compiler emits `adamicCheckedViewCast`. Its strict verifier was run unchanged and exits 1, but reads the historical observations, so its two-pass summary is not our baseline. worker-evidence/baseline-observations.json is a fresh independent audit with compiler/backend/source receipts; audit execution disables implementation mutants, not runtime checks. Its fixture hashes refer to the input fixtures before removing the now-obsolete 01 refusal headers.

Tagged contract group: TestStep09AcceptanceTagged holds all four tagged fixtures to Node, sanitized native, native release, and JavaScript, pins exact stop diagnostics, and checks leaks for successful executions. Its real IR mutants remove one cast-point check in each failing twin. Both backends still stop on a later checked read, but with a different diagnostic; the exact cast-point pin catches each omission. No unchecked-success mutant is claimed. No compiler change was needed for this group.

Commands (all output saved to logs):
- GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh: pass; setup 49.190s; nproc 5, CPU quota 4.
- go build -o /tmp/step09-worker-adamic ./cmd/adamic: pass.
- NODE_PATH=stage3/api/node_modules node stage3/fixtures/checked-casts/observe.cjs /tmp/step09-worker-adamic /tmp/step09-worker-baseline: exits 1, missing emitted-check mutant at 01 failing.
- node stage3/fixtures/checked-casts/verify.cjs --require-runtime: exits 1 on historical 18 blockers.
- node /tmp/step09-baseline-audit.cjs /tmp/step09-worker-adamic /tmp/step09-worker-audit: pass; fresh five contracts pass, fourteen lowering blockers, one runtime diagnostic mismatch.
- go test ./internal/oracle -run 'TestStep09AcceptanceTagged|TestScannerCastCounts' -v -args -update-counts: pass, 3.630s; four new canonical counts rows.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: fails on inherited fixtures, 50.949s; the focused refresh records this group's rows.

No full package suite or full gate was run. Untagged, finite-domain, generic and primitive contracts remain unfinished; the strict acceptance requirement is not green. The historical a-check and native-mutant receipts are not fresh worker validation.

## Owned scalar and finite-enum contract group

Mechanism commit: 675951e7. Admission is limited to checkable owned primitive values and finite primitive target domains, and evaluates the operand once as a helper argument. Both backends receive ordinary If/Panic/Narrow IR. No any/unknown, double assertions, intersections or nonphantom brand assertions are admitted. Whole numeric enums remain open number domains.

Fresh committed-compiler observation: all 20 source Node goldens pass; 11 runtime contracts pass, eight casts refuse lowering, and one view liar has a diagnostic mismatch. Therefore nine of the original 18 contracts now pass. The supplied verifier remains unchanged and exits 1 on the first nonpassing record; it does not accept the view diagnostic mismatch as a runtime pass.

| Originally blocked contract | Current result |
| --- | --- |
| 01_enum_declaration.a | pass |
| 01_enum_declaration_fails.a | pass |
| 03_chain_info.a | pass |
| 03_chain_info_fails.a | stops at correct read; wrong diagnostic prefix |
| 04_literal_payload.a | lowering refusal: backing primitive-union field |
| 04_literal_payload_fails.a | same lowering refusal |
| 05_extension.a | pass |
| 05_extension_fails.a | pass |
| 06_parser_keyword.a | pass |
| 06_parser_keyword_fails.a | pass |
| 07_generic_next.a | lowering refusal: generic cast preflight |
| 07_generic_next_fails.a | same lowering refusal |
| 08_generic_node.a | lowering refusal: generic cast preflight |
| 08_generic_node_fails.a | same lowering refusal |
| 09_primitive_string.a | pass |
| 09_primitive_string_fails.a | pass |
| 10_generator_label.a | lowering refusal: number/Expression/undefined array representation |
| 10_generator_label_fails.a | same lowering refusal |

TestStep09AcceptanceScalars verifies 05, 06 and 09 against Node, sanitized native, native release and JavaScript, including leak checks on success. Native/JS finite-domain omission mutants continue with `.wrong` and token 80 and fail the required stop. The primitive type-check omission is caught by the exact message: the surviving representation conversion stops with `a union value does not match its narrowed type`, lacking the cast expression, expected type and found type. The observer recognizes both tag-check call forms and owned-scalar panic calls; it removes the actual emitted cast-point checks. A later read may still stop, but cannot satisfy the pinned cast-point diagnostic. All original runtime goldens and verify.cjs are preserved.

Verification commands and outputs:
- go test ./internal/lower -run TestScalarCast -v: pass, 0.159s; includes uncheckable-type refusals and open numeric-enum semantics.
- go test ./internal/oracle -run 'TestStep09Acceptance|TestScannerCastCounts|TestScannerErrorLibraryCastNotYet' -v -args -update-counts: pass, 8.760s; six additional canonical counts rows.
- Final same focused oracle selection without update: pass, 6.489s.
- go vet ./internal/lower ./internal/oracle: pass.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: inherited failure, 64.325s; focused counts refresh passes.
- observe.cjs /tmp/step09-worker-adamic /tmp/step09-scalars-committed-observation: 20 Node goldens, 11 runtime passes, eight blockers, one mismatch, 25 observer controls.
- verify.cjs --require-runtime: exit 1, `exact compiler blocker and header` at the recorded view mismatch. No verifier requirement was weakened.

Setup timing lines: node 0.030s, go 0.030s, markdown dependencies 0.091s, submodules 0.099s, clang 0.306s, Go build 49.034s, deferred test binaries 49.161s, cache warm 49.163s, done 49.190s; nproc 5, CPU quota 4. These are observed setup durations, not estimates.

Pending view diagnostic decision: verify.cjs:30 requires `cast failed:` for all liars, while object.c:248 and delivered scanner pins use `field read failed:` for lazy reads. Changing the existing diagnostics changes existing oracle contracts. The backing primitive-union, generic and mixed-array mechanisms are still unfinished and are not presented as delivered acceptance contracts. Error's independent pinned NotYet test remains green for library #ddwcejg. No full gate or whole-package suite was run.
