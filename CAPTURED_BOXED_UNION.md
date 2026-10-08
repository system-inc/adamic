Allow the existing one-word boxed Union in captured cells.
Commits: codex/generic-function-value, 2026-10-08 UTC; host scratch integration follows.
Checks: uncached Node/native/JavaScript positive oracle 0.429s; full lower 21.758s; vet clean; measured counts 13/13 allocations/frees.
Mutant: remove the captured union cell write; both backends exit normally with repeated owned! and fail against Node stdout, /tmp/captured-union-write-mutant.log.
Uncovered: other slotless captures and closure Union argument/result support remain unchanged; full fixture 25 next fails native emission for narrowed optional depth--.

The cell representation already stores and retains strong reference pointers, including boxed scalar unions and undefined as NULL. The guard incorrectly grouped this one-word Union with the unsupported boolean pair. Exempt only Union, preserving the other checks and cycle analysis. No new runtime representation or null sentinel is introduced.

captured_boxed_union.a covers mutable live capture, owned string replacement, boxing numbers, storing undefined, and snapshots of the previous value across a write; generic captures of string | number also return the correct type. Node prints owned!, 11, missing, tail!, string, number. Counts retain/release 44/55, peak 9, regions 0. /tmp/captured-union-{oracle,write-mutant,counts,lower,vet}.log records evidence.

The pinned host fixture now lowers to C but fails clang at depth--: it subtracts 1 directly from an adamic_maybe_number. The JavaScript/complete-host oracle is not claimed green. /tmp/captured-union-host25-oracle.log records the clean refusal being replaced by this next emitter defect; repair will use the checker-narrowed read and the existing optional storage fit.
