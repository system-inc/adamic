# Uniform field slots

Implementation branch `codex/perf-field-access`, from main
`fe3b9f236e0672e968bcb7ad53c1882cdde18f29`. The optimization is in
`internal/native/fields.go`. The existing emitter has only a three-line hook,
a cached-analysis member with its comment, and one clarified comment.
No lowering, runtime, oracle driver or ownership logic changed. The earlier
UTF-16 optimization is not merged into this branch.

## Choice from measured costs

Both ports identify generated class/object accesses as a compiler proposal.
Their reports do not isolate all checks within method bodies. This unit
attributes disjoint self instructions by source line in fresh baseline
Callgrind profiles, using the saved generated C and adamic.h:

| Measured source-line group | Scanner Ir / share | JSON Ir / share | Mean share |
| --- | ---: | ---: | ---: |
| Generated class-field lines (`->shape ==`) | 157,932,733 / 6.500% | 17,619,567 / 1.937% | 4.219% |
| Inline object-field cache, header 170-176 | 939 / 0.00004% | 21,002,301 / 2.309% | 1.154% |
| Generated entry-stack-check lines | 40,665,286 / 1.674% | 4,860,292 / 0.534% | 1.104% |

Class-field lines are the largest measured component among those groups.
They include the field loads, address calculation and possibly surrounding
work assigned to the same line; **they are not an isolated count of redundant
shape-check instructions**. ABI handling, optional offsets and generic APIs
are not separately isolated here. Whole Scanner.code or Documents.get self
cost is not attributed to checks. Inclusive rows overlap and are not added.

## Optimization and proof boundary

For each field name, scan every ObjectLiteral layout in every function and
the top level. If all layouts containing that field agree on its slot,
fieldSlot emits the slot address directly. A conflicting offset permanently
disables specialization for that name; otherwise the old class-shape fast
path and generic field cache are kept. The field's existence remains the
checker's existing proof, the same assumption used by object_find. This is
a closed-program layout proof, not an assumption of nominal class identity.

Structural literals, extra fields, generic instantiations, functions, callbacks
and nested control flow participate in the scan. Spreads, copies, reuse and
regions preserve their source layout. The undefined-spread branch is different:
its synthesized emptyFields layout is scanned explicitly.

Runtime objects are included conservatively even if the program does not call
those APIs: Map entries, Error objects, and file/directory success/error
results. The new audit reads all seven embedded static runtime shapes and
checks every declared name and offset against the proof's initial map. A new static runtime shape or an unrecognized declaration in the existing
form fails this audit rather than silently passing.
Map entry fields 0 and 1 are also seeded, because shapeOf synthesizes them.
Future producers of layouts in the compiler must extend this analysis too.

The patch changes no bounds, stack, undefined/narrowing or ownership checks,
no optional-number encoding, number semantics, string API, evaluation order,
field snapshots or retain/release insertion. It specializes reads and writes
through the existing fieldSlot hook, including plain objects. Other lookup
sites, such as spread writes and cast checks, retain their current behavior.

The scanner's static class-shape comparisons fall from 388 to 134; JSON's
from 193 to 64. Scanner.text cannot be specialized: readTextFile's runtime
result places text at slot 1, while the class puts it at slot 0. This
conservative whole-program approach leaves that conflict checked.

## Before and after instructions

| Port | Main baseline Ir | Final Ir | Reduction |
| --- | ---: | ---: | ---: |
| Scanner | 2,429,650,952 | 2,342,851,635 | **3.5725%** |
| JSON bounded sample | 909,698,064 | 883,391,243 | **2.8918%** |

Both totals are whole-process instructions. Field-cache lines fall from
21.00M to 0.862M in JSON, but that disappearance is not itself the gain:
specialized loads move to generated-C lines, and clang changes surrounding
register allocation and code. No wall-time or isolated-check saving is claimed.

The implementation branch contains no port merge. Only the scratch worktree
`/tmp/adamic-field-ports`, branch `codex/scratch-field-ports`, merges the port
branches onto the same main; merge commit
`40cca9c8e4f12190676110d45f04ed1267eb6c2c`. Pins: JSON
`67796a5a8419d5052e1cb7532d6b39bb5777da6c`, scanner
`6636e8f16a932951458030c902cc3c502a6933e9`. The latter also carries later stage1
ports, so its complete-parity corpus is larger than the earlier scanner-only
report. Cohere is `715ba94f3608a6500086b1076ce5cb7e51b836db`, its TypeScript
submodule `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.

A saved main compiler and the patched compiler emit C from the same scratch
port sources. Unlike the runtime-only unit, generated C intentionally differs.
Both builds use the identical saved main runtime, clang 20.1.8, release flags
`-O2 -ffp-contract=off -fno-optimize-sibling-calls` plus `-g`, no sanitizers or
count instrumentation. Callgrind 3.24.0 records Ir. The ports' unchanged
profile.py requires every self record to sum exactly to summary. Raw profiles,
top-twenty inclusive/self lists and diagnostics are adjacent to this report.

Scanner measurement: all 77 compiler files from TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`, **434,790 tokens**. Count mode still
computes values and errors. Complete protocol parity passed for both measured binaries against Go and Node:
77 compiler files, 88 stage1 files, 18,236 generated inputs, **18,401 inputs**
and **25,075,613 identical answer bytes**. The complete scanner suite passed
in **83.071s**, including sanitizers/leaks, gaps and all three port mutants.
The measured -O2 -g binary is supplied in both scanner/profiled snapshot slots;
those two filenames are not independently compiled variants.

JSON measurement: the same fixed 97-text, 340,876-source-byte sample as the
previous two units; all generated cases included, largest multi-megabyte
inputs excluded. Input SHA256
`48cc2ac1454143609af5bf87a5f2b1d9a8b9c65957636dcc095d312fe0b11d18`.
Both profile outputs equal the independent Go answers. The 2.89% is a sample
instruction result, not an extrapolation to the complete corpus. JSON's
nonfatal Valgrind brk-segment diagnostic remains; both processes finish normally.

## Validation and counts

The complete final `go test -v -count=1 -timeout 30m ./...` **passed, exit 0**.
All native, flow, fresh, fuzz, load, lower, oracle and main's stage1 packages
passed. Native: **376.396s**; complete oracle: **966.159s**. The Node/native/JS
fixture comparisons, input fixtures, release builds, ASan/UBSan and leak checks
all passed. `TestCountsAreRecorded` passed in **943.31s**: **all 173 rows are
identical to main; no row gets worse and no exception is needed**.

The new Node-backed uniform-field test passed in both release and sanitized
builds, including writes, extra/reordered fields, generic numeric/string
representations, optional access and undefined spreads. Both new native tests
passed separately in **6.202s** and again in the complete native gate. The
runtime-layout audit checks seven C layouts. Vet, gofmt and whitespace checks
passed with empty diagnostics.

The complete JSON suite passed in **474.992s**. Both measured binaries passed
full-corpus snapshots; all **1,076** texts remain byte-identical on Go, Node,
native and the JS backend. The 72 boundaries, single-file driver, sanitizers,
leaks, three port mutants, cached-width mutant, gap refusals and external
printer mutants passed. Prettier 3.9.6 still lists **exactly the nine known
upstream differences**, in its separate comparison. No new difference appeared.

Baseline counts are saved as counts-before.md, identical to main's
internal/oracle/counts.md. Its 173 data rows have SHA256
`18925c2153195e928190ff5323b5ad9a9e61c5db5f60648aeee0c5bde5517ee4`.
The table records allocations, frees, retains, releases, peak live and regions,
not CPU instructions. It is not updated by this unit. Existing exclusions and
refusals remain; no new exclusion is introduced.

An initial implementation omitted runtime-generated layouts. The scanner
comparison caught it with panic `missing source`, because it read runtime
text from the class's slot. Those after profiles and the interrupted initial
gate/port run are rejected and are not the final evidence above. All affected
measurements and gates were restarted after including runtime layouts. The
new test source initially needed corrections to fit the prelude's single-string
console.log and sound readonly widening rules; only its passing final run is used.

## Mutants

| Mutant | Check and observed failure |
| --- | --- |
| Accept a conflicting offset by replacing the permanent -1 with the latest slot | class_layouts.a compiled successfully; Node exits 0, sanitized native exits 1 under UBSan reading a null array; release build crashes. The ordinary byte/exit oracle catches it. |
| Disable specialization entirely | New semantic/code-generation test rejects remaining lookup calls: uniform fields still use shape lookup. This proves parity alone cannot mask an optimization that never runs. |
| Omit runtime text from the initial layout map | Runtime-layout audit rejects the missing field. This proves the new audit can fail. |

These use Go overlays of fields.go and leave real source unchanged. The
unsafe conflict mutant reaches a native executable; it is not a -Werror or
compiler rejection. Full command outputs are archived beside this report.

## Commands

All test output is redirected, never piped. Actual commands:

```sh
source /workspace/adamic-tools/env.sh
go test -v -count=1 -timeout 30m ./... > /tmp/adamic-field-perf/full-gate.log 2>&1
go test -v -count=1 ./internal/native -run '^Test(UniformFieldsMatchNode|RuntimeFieldLayoutsAreIncluded)$' > /tmp/adamic-field-perf/fields-test.log 2>&1
go test -overlay=/tmp/adamic-field-perf/conflict-mutant-overlay.json -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_layouts\.a$' > /tmp/adamic-field-perf/conflict-mutant.log 2>&1
go test -overlay=/tmp/adamic-field-perf/disabled-mutant.json -v -count=1 ./internal/native -run '^TestUniformFieldsMatchNode$' > /tmp/adamic-field-perf/disabled-mutant.log 2>&1
go test -overlay=/tmp/adamic-field-perf/missing-runtime-mutant.json -v -count=1 ./internal/native -run '^TestRuntimeFieldLayoutsAreIncluded$' > /tmp/adamic-field-perf/missing-runtime-mutant.log 2>&1
go vet ./... > /tmp/adamic-field-perf/vet.log 2>&1
gofmt -l cmd internal > /tmp/adamic-field-perf/gofmt.log 2>&1
```

In the physically initialized scratch port merge, copy only emit.go and fields.go
from the implementation branch. The complete port commands:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-utf16-perf/typescript ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/tmp/adamic-field-perf/scanner-before-snapshot:/tmp/adamic-field-perf/scanner-after-snapshot go -C /tmp/adamic-field-ports test -v -count=1 -timeout 30m ./stage1/typescript/scanner > /tmp/adamic-field-perf/scanner-parity.log 2>&1
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier ADAMIC_JSON_PROFILE_BINARIES=/tmp/adamic-field-perf/json-before:/tmp/adamic-field-perf/json-after go -C /tmp/adamic-field-ports test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-field-perf/json-parity.log 2>&1
```

Save main's compiler before modifying the emitter. For each port and side:

```sh
/tmp/adamic-field-perf/adamic-SIDE c /tmp/adamic-field-ports/stage1/PORT/main.ts > /tmp/adamic-field-perf/PORT-SIDE.c 2>/tmp/adamic-field-perf/PORT-SIDE-build.log
clang -std=c11 -O2 -g -ffp-contract=off -fno-optimize-sibling-calls -I/tmp/adamic-field-perf/runtime-before /tmp/adamic-field-perf/PORT-SIDE.c /tmp/adamic-field-perf/runtime-before/*.c -lm -o /tmp/adamic-field-perf/PORT-SIDE > /tmp/adamic-field-perf/PORT-SIDE-clang.log 2>&1
VALGRIND_LIB=/tmp/adamic-json-perf/valgrind/usr/libexec/valgrind /tmp/adamic-json-perf/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/tmp/adamic-field-perf/PORT-SIDE.callgrind /tmp/adamic-field-perf/PORT-SIDE ARGS > /tmp/adamic-field-perf/PORT-SIDE.stdout 2>/tmp/adamic-field-perf/PORT-SIDE.stderr
python3 /tmp/adamic-field-ports/stage1/cohere/json/profile.py /tmp/adamic-field-perf/PORT-SIDE.callgrind > /tmp/adamic-field-perf/PORT-SIDE-top.txt
```

Here PORT in source paths is typescript/scanner or cohere/json; in artifact
names it is scanner or json. SIDE is before or after. Scanner ARGS is
`--manifest /tmp/adamic-utf16-perf/scanner-before/compiler.txt --count`; JSON
ARGS is `--cases /tmp/adamic-json-perf/profile-cases.txt`. The two scanner
snapshot directories contain each corresponding measured binary as scanner
and profiled, and identical port TS files for the Node side. Compare JSON
stdout to `/tmp/adamic-json-perf/profile-expected.txt` with cmp.

Setup: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm **8s**, total **8s**.
`nproc`: **5**; CPU quota 4, RAM 17.6 GB. Go 1.27.1, clang 20.1.8, Node
24.19.0. The previous unit's extracted Valgrind was reused. No Go/Node
instruction or allocation attribution and no wall-time improvement is claimed.
