# Take-day on next merge rehearsal, October 8, 2026

Status: stopped at an unresolved semantic conflict. This is a report-only branch,
not a completed integration and not a landing candidate.

Base: runtime/area-on-next 48ea5354267465c02416d7a023c0a25f425ecc07.
Incoming: runtime/area-take-day 03ed6ebfd4d1e1187a75ca15dd1eb1cdbb67387d.
Both fetched remote tips matched the requested pins. The rehearsal used
`git merge --no-commit --no-ff origin/runtime/area-take-day`, producing exactly
33 conflicted files. No conflict resolution was written or staged. The attempt
was aborted after preserving the evidence in the scratch directory below.

## Named blocker

Area-on-next's e78878b6 introduces `ir.Labeled` and named `Break.Label` and
`Continue.Label` targets. Take-day's 26136b17 and its async completion work
use source/emitted breakable stacks and numeric depths.
`internal/flow/async_completion.go:statements` handles Break/Continue by Depth,
does not inspect Label, and has no Labeled case. Its `abrupt` method reconstructs
jumps with Depth alone. Merely retaining both fields in ir.go does not reconcile
this: labeled bodies would not be traversed by the completion pass, and named
jumps would not identify the correct target or intervening finally. This is a
source inspection finding, not an executed failing witness.

The messages of ab827d7c and 48ea5354 settle object metadata and closure owner
representation, but do not settle this control-flow representation. The source
commits say respectively "Compile labels through nested control flow and finally
cleanup" and "runtime: support labeled blocks and continue targets in async
functions". Neither specifies the integration boundary. Per the instruction
"stop and name any conflict you can't decide from the commit messages", the
rehearsal stops here rather than silently discard either behavior.

Decision needed: translate named labels to the completion pass's target/depth
representation before async normalization, or extend that pass to traverse
Labeled bodies and resolve named targets through finally routing. Both require
a reviewed integration change beyond choosing conflict text.

## Complete conflict inventory

All rows remain unresolved. The descriptions below are provisional requirements
from inspection, not implemented or tested resolutions.

| File | Requirement or blocker |
| --- | --- |
| `docs/nested-functions.md` | Both histories append independent evidence. Retain both sections. |
| `docs/runtime-statics.md` | Retain the closure/Node-host audit, deterministic normalization and exit evidence, deterministic signal evidence, and typed-array sort audit. |
| `internal/flow/build.go` | BLOCKER: named continue targets overlap depth-based targets. Completion normalization must agree with this representation. |
| `internal/flow/infer.go` | Combine Node-buffer mutation classification with MapClear mutation classification. |
| `internal/ir/call_targets_guard_test.go` | Keep both audited call-target consumer registrations. |
| `internal/ir/ir.go` | BLOCKER: named Break/Continue targets overlap depth targets. Preserve readiness and source identity fields independently. |
| `internal/javascript/javascript.go` | BLOCKER: named versus depth continue targets. Also preserve iteration-cell resets and readiness semantics. |
| `internal/lower/async_test.go` | Reconcile obsolete refusals only after the merged async capability is established; awaited void is supported by area-on-next. |
| `internal/lower/borrow.go` | Keep the EnvironmentCell ownership exclusion; compare captured-local exclusion against current frame lowering. |
| `internal/lower/cycles.go` | Retain cycleTypes extraction while running resolveCountTypes after graphTypes on successful paths. |
| `internal/lower/expression.go` | Preserve closure argument conventions and source identities; reconcile fitted boxed slots with the newer callable boolean representation. |
| `internal/lower/functions.go` | Preserve rest-element lowering, optional-parameter metadata and source identity; reconcile callable slot representation. |
| `internal/lower/generic.go` | Combine lexical closure owner metadata with source identity. |
| `internal/lower/library_map_set.go` | Preserve incoming fitted union slot support without restoring superseded callable representation refusals. |
| `internal/lower/lower.go` | Keep readiness analysis and the optional ownership-query gate. |
| `internal/lower/lower_test.go` | Do not resurrect refusals superseded by area-on-next representation support; verify each disputed refusal. |
| `internal/lower/object.go` | Combine contextual array-element selection with fitted boxed union storage checks. |
| `internal/lower/statements.go` | BLOCKER: ir.Labeled/name-based lowering versus synthetic breakable/depth-based lowering. |
| `internal/native/class_accessors.go` | Preserve packed optional argument/count ABI while using stable accessor storage names. |
| `internal/native/class_inheritance.go` | Preserve typed numeric virtual target identities, with stable class-derived names and declaration deduplication. |
| `internal/native/emit.go` | Combine readiness/declaration flags with stable names; preserve typed closure prototypes and deduplicate declarations. |
| `internal/native/emit_locals.go` | Keep readiness checks and stable names; use integerCounter for counter reads. |
| `internal/native/emit_objects.go` | Keep dynamic shape registration and stable class names; shape identity includes incoming field kinds. |
| `internal/native/emit_statements.go` | BLOCKER: named versus depth continue targets. Preserve CopyCell, readiness and stable names independently. |
| `internal/native/fields.go` | Union the host field-name sets and string-iterator fields. |
| `internal/native/library.go` | Combine source feature defines, profile validation, and split-build jobs in the runtime cache build. |
| `internal/native/runtime/adamic.h` | Combine canonical closure cache with externally supplied environment cells; keep shape validation and typed virtual dispatch. |
| `internal/native/runtime/closure.c` | Initialize the canonical cache on every environment initialization path; retain owner-kind cache lookup and cell owners. |
| `internal/native/runtime/object.c` | Keep shape checks and adamic_object_size; preserve counted method-entry ABI in the outlined cache miss path. |
| `internal/native/runtime/region.c` | Keep per-slot readiness/representation bytes and adamic_object_size; avoid counting the same regional object twice. |
| `internal/native/runtime_statics_parallel_test.go` | Combine deterministic signal-handler proof with deterministic exit-flush proof. |
| `internal/oracle/counts.md` | Keep fixture union and regenerate once after a resolved, buildable merge. Not regenerated in this stopped rehearsal. |
| `internal/oracle/oracle_test.go` | Retain both fixture registration sets. |

## Validation and failures

No merged-tree build, vet, lower/native/oracle package gate, cycle corpus, or
count regeneration was run: no resolved merged tree exists. No new checks or
mutants were introduced. The reported 6,630-pass take-day result is the user's
prior result, not independently reproduced here.

Merge-caused tooling failure: the first setup cache build overlapped the merge
and failed on unresolved conflict-marker syntax in lower/borrow.go, cycles.go
and expression.go. This scheduling error does not diagnose a resolved merge.
Toolchain readiness completed: node 0.022s, Go 0.027s, markdown dependencies
0.068s, submodules 0.079s, clang 0.167s. `nproc` reported 5.

Pre-existing failures: not assessed. No baseline failure is classified without
a baseline reproduction. Setup was rerun after aborting the rehearsal and passed on the exact base:
Go ready 0.019s, Node 0.022s, submodules 0.064s, markdown dependencies
0.072s, clang 0.177s, go build ready 30.420s, build cache warm 30.616s,
total 30.654s. Test binary warming was deferred. The environment file is
`/workspace/adamic-tools/env.sh`; CPU quota is 4, visible processors 5.

## Evidence

Scratch files in `/workspace/scratch/take-day-on-next/`:
`merge.log`, `conflicts.txt`, `unresolved.diff`, `unmerged-index.txt`,
`setup.log`, and `setup-base.log`. Parents and the exact merge command above
allow the conflict state to be reproduced. Only this report is committed.
No runtime, compiler, fixture, or count changes are included.
