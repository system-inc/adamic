Built: empty-children counts on all 77 files and a direct empty-array allocation probe; production unchanged.
Commits: measured base 21afae52681c65b4ee15b846890e324d083a9dfc on codex/parse-speed; this is a stop-condition report, not a runtime optimization.
Commands and outputs: 463,817 of 889,146 allocated nodes keep empty/no-buffer children; empty array uses one fresh 56-byte header and no element buffer.
Mutants: forced empty element-buffer allocation fails the probe; changed destroyed-node count fails reconciliation.
Not covered: a shared-header mutant, a new full oracle gate, a complete array-consumer audit, or before/after Ir, user time and cache misses, because the requested stop condition applies.

# Empty children allocation check, October 7, 2026

**Stop: adamic_array_new(0) already makes only one allocation.** array.c creates
one header through adamic_allocate, initializes elements to NULL, and calls
malloc for an element buffer only when capacity > 0. adamic_array_push grows a
zero-capacity array to four slots with realloc(NULL, ...). Each call constructs
its own header, never a shared empty object. The proposed representation is
already the runtime's current behavior; there is no second allocation to remove.

## Observed node counts

On the pinned 77 TypeScript compiler files, 050880ce59e30b356b686bd3144efe24f875ebc8,
using the accepted batch8 parse-only driver and runtime-parse.c from 3cbc640:

| Observation | Count |
| --- | ---: |
| ParseNode constructor calls | 889,146 |
| Nodes destroyed | 889,146 |
| Empty children at construction | 463,817 |
| Empty children at destruction | 463,817 |
| Empty children at destruction with any element buffer | 0 |
| Reachable whole-tree nodes | 887,803 |
| Reachable leaves | 463,796 |

Empty children are 52.16% of allocated nodes and 52.24% of reachable nodes,
above the proposed 30% threshold. The other stop condition nevertheless applies.
The lead's 557,010-node total is not substituted for this measured snapshot's
counts. Allocated nodes include 1,343 nodes not retained in the reachable tree;
whole-tree traversal and constructor counts are distinct observations.

Scratch generated-C probes count both ordinary and region ParseNode constructor
entries and empty arguments. Scratch heap.c counts the ParseNode shape at
free_one, inspecting children length and whether elements is NULL before
releasing fields. Identification checks the 13-field shape's kind/children/
multiLine names. Born/free counts reconcile exactly. No production source is
edited. The runtime retains the buffer after pushes/removals until array
release, so empty-at-free with no buffer supplies evidence that these arrays
never grew; equality of initial/final empty totals alone would not prove that.

Reachable counts come from the separate whole-tree driver previously checked
against Go in BRANCHES.md. The 77-file run completed with empty stderr and
44,766,682 output bytes. Each record is a node; a node is a leaf when the next
record has equal/lower depth, or its file ends. File/case markers are metadata,
not nodes. That traversal does not count discarded speculative nodes.

## Direct allocation control and mutant

The retained allocation.c probe wraps malloc and realloc, observes just
adamic_array_new(0, false), then makes a second live array and checks header
identity. Under ASan the slab allocator is disabled, making allocation events
unambiguous: one 56-byte header malloc, zero reallocs, one counted heap value,
capacity zero and elements NULL. First push makes exactly one realloc and reads
back its value. Both arrays are released: allocations 2, frees 2, live zero,
no ASan/UBSan/LeakSanitizer report.

ADAMIC_COUNT alone is insufficient to detect the extra buffer because its
allocation hook counts heap headers, not element buffers. The malloc wrapper
and NULL-buffer assertion independently check the representation.

The scratch mutant changes array.c's `if (capacity > 0)` to
`if (capacity > 0 || references == false)`, forcing a malloc even for this empty
numeric array. It compiles under all warning flags and exits 1 with
`empty array allocation check failed`. The probe releases that array on the
failure path, so allocations/frees are 1/1 and no sanitizer or leak error kills
this mutant. The allocation assertion is the catcher. A destroyed-node count
increased by one independently fails born/free reconciliation.

## Exact probe configurations

The counter build uses clang 20.1.8, -O2 -g, sanitizers off,
-ffp-contract=off, -fno-optimize-sibling-calls, no LTO and no ADAMIC_COUNT:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/empty-array-check/runtime /workspace/scratch/empty-array-check/count.c /workspace/scratch/empty-array-check/runtime/*.c -lm -o /workspace/scratch/empty-array-check/count > /workspace/scratch/empty-array-check/count-build.log 2>&1
/workspace/scratch/empty-array-check/count --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/empty-array-check/count.stdout 2> /workspace/scratch/empty-array-check/count.stderr
```

The allocation control/mutant use -O1 -g with ASan, UBSan, LeakSanitizer and
counted headers, rather than a release instruction/time measurement:

```sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all -DADAMIC_COUNT -I internal/native/runtime /workspace/scratch/empty-array-check/allocation.c internal/native/runtime/*.c -lm -Wl,--wrap=malloc -Wl,--wrap=realloc -o /workspace/scratch/empty-array-check/allocation > /workspace/scratch/empty-array-check/allocation-build.log 2>&1
/workspace/scratch/empty-array-check/allocation > /workspace/scratch/empty-array-check/allocation.stdout 2> /workspace/scratch/empty-array-check/allocation.stderr
```

The mutant replaces the include/runtime paths with allocation-mutant-runtime and
writes separate build/stdout/stderr logs. The only runtime source difference is
the stated capacity condition. Production array.c is unchanged.

Setup: Go/clang/Node/submodules ready at 0s; cache warm and done at 25s; nproc=5,
cpu.max=400000 100000, memory 17.6 GB. Evidence, probe source and input/C hashes
are in empty_array_evidence. Scratch instrumented C/runtime remain under
/workspace/scratch/empty-array-check. No compiler/emission/parser/scanner edits,
header sharing, new optimization claim or unrequested continuation past the
stop condition is made.
