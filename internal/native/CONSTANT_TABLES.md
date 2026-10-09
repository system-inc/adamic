Sanitized full build: **18.512 s -> 14.039 s (24%)**; C-to-native: **6.989 s -> 2.769 s (60%)**, split on, three interleaved runs.
Built static const numeric and record templates with bounded copy helpers; widthTables module: **3.483 s -> 0.077 s**, slowest initializer: **2.990 s -> 0.228 s**.
Commits: f70f3c85 implementation, d878e816 merges main f4efdd23, 397fa125 bounds record specialization; branch codex/outline-module-main, baseline a7ddba49.
Validation: native/lower/oracle green, 497 oracle fixtures in each split mode (494 original plus 3 from main), 13,029,128 markdown output bytes identical; five new mutants and eight inherited mutants caught.
Limits: Linux clang 20.1.8 only; no full repository gate or new release timing; runtime archive warmed outside timings; runtime operands and large record shapes keep ordinary emission.

The expensive inputs were executable initialization. The fixture emitted 1,515
adamic_object_new sites, 1,611 adamic_array_new sites and 6,572 adamic_array_push
sites. References surviving across chunks also gave the widthTables wrapper a
large backing array of live pointers. Afterward there are 124, 142 and 368 sites.
These are **code sites, not allocation counts**: copy loops still allocate fresh
objects and arrays. C shrinks from 4,960,224 bytes / 29,508 lines to 1,496,073
bytes / 15,087 lines, with 1,392 record rows and 50 numeric tables as static data.

Clang -ftime-report corroborates the backend cost. For the widthTables module,
backend CPU time falls from 2.4808 s to 0.0325 s; greedy register allocation
falls from 0.7272 s to 0.0053 s. Initializer bucket 04's backend CPU time falls
from 2.4821 s to 0.1762 s. Profiles ran alongside tests, so their contended wall
times are **not** the performance measurements below. Preprocessed inputs,
flags, reports, measurements and hashes are in
[constant_table_evidence](constant_table_evidence/metadata.json).

Numeric templates accept number literals and unary plus/negation of such literals,
using the existing exact hexadecimal cNumber representation. A shared noinline
helper allocates an ordinary array of the same capacity and copies values.
Single-element literals retain ordinary emission so edits such as [14].length
stay within their source unit.

Module-level object arrays can batch consecutive closed records with the same
layout into static const adamic_value rows. Numeric array fields use flattened
const numbers and offsets. The helper reconstructs every nested array and record
with the normal allocator and shape, then transfers each fresh record to the
result array. String fields use existing immortal literals. Objects and nested
arrays keep their independent identities and remain mutable.

Runs require eight records, within arrays of at least sixteen elements. Records
have at most sixteen fields so their specialized constructor cannot become huge.
Larger shapes retain ordinary stores and statement outlining. Classes, methods,
private fields, tuples, spreads, function bodies and unknown expressions retain
the normal record path. Imported range reads break a run and execute in their
original position with their existing readiness checks. The fresh outer array is
allocated earlier but stays unexposed until the original assignment. Runtime
operands retain their evaluation order and statement cleanup.

The production hook is internal/native/emit_expressions.go:523-528:
constantRecords and constantArray precede the existing array emission.
internal/native/constant_tables.go contains eligibility, data emission and helpers.
Helper names use the existing adamic_initialize_ noinline-prototype convention
and content-addressed declaration identities. No new hook was needed in emit.go,
IR, lower or the splitter. Existing shared-data buckets, helper buckets and
dependency declarations handle the templates. Field stores, switch lowering
and check elision were not restructured.

The following are median wall times of three actual clang unit compilations.
The full basename prefix is
module__2e__2e__2f_widthTables_2e_ts_9a9983a619e7fdf13ddd9ca4f1c25ab84372ffced184ad862375bca49e0149a1.
A removed bucket has no compile invocation; remaining bucket contents have changed.

| Unit suffix | Before s | After s |
| --- | ---: | ---: |
| .c | 3.483 | 0.077 |
| _initialize_00.c | 1.365 | 0.123 |
| _initialize_01.c | 2.806 | removed |
| _initialize_02.c | 2.039 | 0.228 |
| _initialize_03.c | 1.826 | removed |
| _initialize_04.c | 2.990 | 0.226 |
| _initialize_05.c | 0.834 | removed |
| _initialize_06.c | 1.859 | 0.061 |
| _initialize_07.c | 1.862 | 0.138 |

| Measurement | Before s | After s |
| --- | ---: | ---: |
| C-to-native runs | 6.938, 6.989, 7.003 | 2.703, 2.852, 2.769 |
| C-to-native median | 6.989 | 2.769 |
| Full frontend + native runs | 18.512, 18.843, 18.095 | 13.798, 14.277, 14.039 |
| Full frontend + native median | 18.512 | 14.039 |
| Program compilation units | 80 | 66 |

Both versions use the same current splitter and runtime. Before emission is
a7ddba49; the frontend comparison overlays only its emit_expressions.go onto
the current branch. Every timed frontend run verifies exact equality with saved
C. Final emission after the record-size guard matches timed C byte for byte.
Options are Split:true, Jobs:5, Sanitize:true, full -g, -O1 and ASAN/UBSAN.
Three interleaved before/after rounds ran on an otherwise idle box, with uncached
program objects and runtime archives warmed separately. These are this workspace's
measurements, distinct from developer tools' fresh-cache measurements on their box.

Reproduce after sourcing the toolchain:

    python3 internal/native/constant_table_evidence/reproduce.py

Setup ran bash cloud/setup.sh, then sourced /workspace/adamic-tools/env.sh.
nproc is 5, cpu.max is 400000 100000. Timing lines: Go 0.024 s, Node 0.023 s,
markdown dependencies 0.073 s, submodules 0.073 s, clang 0.171 s,
Go build 23.832 s, cache warm 24.026 s, done 24.054 s. Full setup log is saved.

Validation commands and complete logs are retained:

    ADAMIC_NATIVE_SPLIT=0 go test -json ./internal/native ./internal/lower ./internal/oracle -count=1 -timeout=30m
    ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 go test -json ./internal/oracle -count=1 -timeout=30m
    go test ./internal/native -run '^TestConstant' -v -count=1 -timeout=10m
    ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -v -count=1 -timeout=30m
    go vet ./...

Both merged oracle gates passed all 497 cases: the original 494 plus three added by main. The complete native/lower gate
preceded the final sixteen-field specialization guard; focused tests then verified
its conservative fallback, Node output and mutant. Pre-main uncached native/oracle
gates also passed. The exhaustive width fixture validated 1,284,947 texts,
every Unicode scalar, generated sequences and its three semantic mutants.
It ran before the latest main merge; subsequent merged frontend emission and
corpus parity verify the timed fixture on current main.

The saved 5,166-document markdown corpus from the preceding unit is unchanged.
All 13,029,128 output bytes match Go, source Node and original cohere Node in
sanitized and release split builds, with empty stderr. Input/output hashes are saved.
The constant-table fixture compares Node against both split and sanitizer modes,
including negative zero, infinities, subnormals, side effects, spreads, empty arrays,
independent mutation and identities. The large-record fallback fixture compares
sanitized split output to Node.

The final lint probe retains the earlier acceptance: add-function, add-temporary
and 14 to [14].length each change **1/140 units**, with no shared-header change.

| New mutant | Check that caught it |
| --- | --- |
| Swap first two numeric template elements | constant_tables.a stdout differs from Node |
| Swap record template rows 2 and 3 | constant_tables.a stdout differs from Node |
| Treat runtime numeric expressions as constants | constant_tables.a side-effect/output comparison |
| Normalize zero after unary negation | constant_tables.a negative-zero comparison |
| Raise record-field cap to 10,000 | 256-field fixture rejects oversized constructor |

The four behavioral mutants compile and run; failures are output differences.
The field-cap mutant fails the intended emitted-function size check.
Eight inherited mutants were rerun: reverse module order (Node output), remove
chunk bound / noinline (initializer guards), omit root cleanup (LSAN), disable
chunk placement / storage ownership / ready-write ownership / retained prototype
(ownership and ABI checks). The width fixture caught East Asian width,
narrow-emoji and DEL shortcut mutants too.

Two initial variants were refined; their logs are retained. A blanket
everything-is-numeric-zero mutation corrupted string fields and was caught by a
runtime segmentation fault. A mutation normalizing only NumberConstant.Value
survived this source fixture because -0 lowers as unary negation of positive zero.
The corrected post-negation mutation is caught by Node output comparison.
No compiler error was counted as a behavioral mutant catch.

No full go test ./..., macOS/WASI validation or release speed measurement was run
for this unit. Debug flags and defaults remain unchanged.
