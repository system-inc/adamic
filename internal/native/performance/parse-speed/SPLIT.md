# Parse instruction split, October 7, 2026

Measured before any performance source change, on `codex/parse-speed` from
`origin/main` e8ba3d5d81de4d3773c723914fccd4c76248b965. This report records a
reconstructed measurement, not an exact reproduction of #93z4yv7: those notes
and its parse-only harness were not available in the repository. Batch 8 at
4189abd provides the driver and Go oracle, but no parse-only profiling harness.

## Numbers and dated targets

| Workload | Callgrind Ir |
| --- | ---: |
| Native parse-only reconstruction, whole process | 9,259,261,285 |
| Native `collect` entry edge, including line table and parent map | 8,667,357,378 |
| Native input read edges for the 77 sources | 236,943,645 |
| Native collect plus source reads, excluding caller cleanup | 8,904,301,023 |
| Go parse alone, reads and forced line table included | 1,684,637,174 |
| Batch 8 whole Go count-mode run | 3,036,940,751 |

The reported historical 8.97G native and 2.96G Go were **not reproduced
exactly**. The measured native process is 3.23% above 8.97G. Its reads-plus-collect
boundary is 0.73% below it. Go whole-run is 2.60% above 2.96G. A boundary difference
is a hypothesis, not an established explanation. Raw profiles are committed.

On this measured Go parse baseline the requested limits are 2,526,955,761 Ir
by October 8, 2026 at 12:00 MDT (18:00 UTC), and 1,684,637,174 Ir by October 9
at 12:00 MDT. Native whole-process parse is currently 5.496 times Go parse.
Neither dated performance claim is established by these observations.

Native uses batch 8's exact source snapshot, converted to `.a` in scratch,
with absolute imports into current main's parser/scanner. The only measurement
change is deleting `visit(context, root)` in the scratch driver. The Context,
its eager line table, Parser, second Scanner, parent-array map callback, recursive
parent traversal and empty findings sort remain. Count mode prints `0`. The
Go parse harness uses cohere's unmodified typescript-go `ParseSourceFile`, which
sets parents as it constructs nodes, and forces its lazy `ECMALineMap`. It reads
the identical manifest and prints `0`. Go process startup, allocation and GC
are included; `GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1` applies only to profiling.

TypeScript v6.0.3 is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8,
cohere to 715ba94f3608a6500086b1076ce5cb7e51b836db. There are 77 files and
`corpus.sha256` records each file's bytes. Full, unchanged batch 8 release
output, including diagnostics, suggestions, edits and fixed sources, compares
byte for byte with its Go oracle: **11,442,907 bytes**, `cmp` exits 0.

## Disjoint self costs

Each self instruction in the native whole-process profile is in exactly one
row below. These are mechanism buckets: runtime children remain in their own
mechanism bucket rather than being counted again under the phase that called
them. In particular, the file-read row is only the entry function's own work;
its 236.94M inclusive cost includes input decoding, allocation and string work
in other rows. The parent and line-table rows likewise exclude shared runtime
costs. Anonymous libc records, byte-comparator bodies and startup are explicitly
in the remainder. Inclusive costs below are context, never added to this table.

| Bucket | Instructions | Share |
| --- | ---: | ---: |
| Remainder: parser control, arrays, driver, libc and startup | 2,033,417,640 | 21.961% |
| Scanner generated control | 1,903,592,683 | 20.559% |
| Releases and child destruction | 1,542,615,786 | 16.660% |
| Object field and call plumbing | 1,317,090,232 | 14.225% |
| Character reads and UTF-16 indexing | 946,291,026 | 10.220% |
| Retains | 363,315,534 | 3.924% |
| Kind equality headers and dispatch | 315,627,876 | 3.409% |
| Node construction generated control | 286,305,012 | 3.092% |
| Allocation and freeing | 217,432,706 | 2.348% |
| Other string operations, including token values | 114,153,290 | 1.233% |
| Substrings | 85,357,989 | 0.922% |
| Line table generated control | 83,445,905 | 0.901% |
| Parent map generated control | 50,564,366 | 0.546% |
| File reading and input decode | 51,240 | 0.001% |
| Total | 9,259,261,285 | 100.000% |

The largest named mechanism is release/destruction plus retains: 1.906G,
20.58%. Within release, self is 1,344,432,179; `let_go` and child destruction
add 198,183,607. Caller cleanup's release edge is 354,105,637 inclusive.
Object writes/checks alone cost 725,047,843 plus 523,645,837 self. These are
runtime calls on ordinary field updates, not UTF-16 reads. Parser/scanner
control and emitted field/ownership plumbing need separate investigations.

`Parser.file` costs 7,439,580,376 inclusive; initial scanner work in construction
is outside that entry. `Scanner.scan` is 3,516,786,347 inclusive. Recursive clang
clones have overlapping inclusive costs and even exceed the process total;
they are not disjoint phases. Parent mapping is 333,773,523 inclusive.
The line-table body is inlined into `collect`: its generated C lines 8626..8676
account for 83,445,905 self, while its direct character-read edge is 211,749,230.

There are **889,146 ParseNode constructor calls**, including speculative work,
not necessarily 889,146 nodes in the finished AST. The constructor's object-new
edge is 72,986,425 inclusive. `ParseNode.new` plus `Parser.make` self is
286,305,012. Child arrays also allocate separately. The general allocation bucket
includes all allocation/freeing, not just nodes; no per-node allocator amount
is inferred by dividing the total.

### Character reads

Reads preserve JavaScript UTF-16 over UTF-8/WTF-8 bytes. They are not uniformly
walks from byte zero. ASCII uses direct byte indexing; long non-ASCII strings
have checkpoints and a cursor; BMP-only indexed strings have a compact UTF-16
view. Bounds/ToIntegerOrInfinity checks remain. Scanner.code also checks high
surrogates before requesting codePointAt. Generated Scanner.code self is
624,410,321, Scanner.advance self 282,630,568. Runtime character/index work is
946,291,026 self: char_code 376,933,671, decode 225,738,054, usable 106,392,664,
units 86,408,542, bmp_view 80,161,060, locate 68,688,350, plus smaller entries.
Shape checks, stack checks and method/default-argument plumbing are in generated
scanner bodies, not charged a second time to runtime character work.

### Kind strings and literal interning

Literal interning already exists: `internal/lower/locals.go:174` uses a map keyed
by string contents; `internal/native/emit.go` emits one immortal header for each
program.Strings entry. The C snapshot has 564 directly spelled literal headers,
293 named kinds, and no duplicate literal contents. Kind values are taken from
these literals; scanner keyword/punctuation map results also reference them.

A scratch-only runtime replay classifies equality operands by pointer membership
in the program's named-kind headers, derived from typescript-go's Kind list.
It counts every equality, including generic Map/Array runtime comparisons. This
is a pointer provenance check, not guessing from a string's textual spelling.
It observed the same kind headers throughout; no dynamically allocated kind
string was classified. Counts include parser speculation and rescans.

| Equality path | Kind names | Other strings |
| --- | ---: | ---: |
| Calls | 24,049,704 | 1,039,883 |
| Identical headers, subset of calls | 1,247,243 | 643,398 |
| NULL path | 0 | 0 |
| Different lengths, header-only return | 21,046,212 | 203,722 |
| Equal empty strings, header-only return | 0 | 2,548 |
| Byte comparison | 3,003,492 | 833,613 |

87.511% of kind comparisons stop at unequal lengths. 12.489% reach byte
comparison, including **all 1,247,243 same-header kind comparisons** in the
baseline. Of those byte comparisons, 1,756,249 compare distinct kind headers.
Only 5.186% of all kind comparisons can use the proposed identical-header path.
Distinct immortal headers do not by themselves establish inequality for arbitrary
strings, so a generic equality function must retain the content fallback.

Kind equality's **315,627,876 self instructions** are isolated by replaying the
unchanged equality body in a separately named function. Its machine instructions
match the baseline, apart from addresses. Kind and other body self costs sum
exactly to baseline equality self: 315,627,876 + 19,986,213 = 335,614,089. A direct
instruction-path tally independently agrees: 12 instructions per different-length
return and 21 per byte path (including the PLT jump), plus resolver instructions
on the other path. The bucket table charges only these exact self costs.

The classified kind equality body is **374,802,438 inclusive**, including
59,174,562 byte-comparator instructions; the other body is 29,921,325 inclusive.
These are supplementary replay costs. The two inclusive bodies together differ
from baseline equality inclusive (404,753,771) by 30,008 instructions, with
relocated literal bytes. The exact source of this small libc difference is not
isolated. That difference is not
silently allocated: byte-comparator self remains in the original profile's
remainder. Classifier/counter overhead and scan instrumentation are excluded
from the performance baseline. The replay's 15.104G total is not a parse result.

`4ffed41` was not in the fetched remote branch. At the time of inspection
`codex/lint-runtime-fixes` ends at 8c171be and contains the identical-header
optimization as **59a0850**. Its equality-only change will be measured separately
after this split is committed and pushed. No numeric-kind source change or new
interning implementation is justified by the existing literal representation.

### Token text

Ordinary punctuation already uses immortal spellings from Scanner.punctuation;
it does not mint a source slice for every punctuation token. Identifier scanning
does mint text, then keyword lookup uses that text. Literals evaluate/normalize
values and can build multiple strings per token. Greater-than rescans still make
speculative short slices. Runtime string_share copies a slice below 64 bytes or
below one quarter of its ultimate owner's bytes, so typical token spellings copy.

Scratch-only counters bracket Scanner.scan and direct rescan/value methods,
fold nested helper calls into their outer scan, count every heap string allocation
inside that interval, and group by the final emitted kind. These are string
objects, not allocated bytes. They include failed speculative spellings and
intermediate literal construction; scans include parser rewinds and rescans.

| Final token class | Scan calls | String allocations |
| --- | ---: | ---: |
| Punctuation | 890,192 | 4,822 |
| Keywords | 128,102 | 132,970 |
| Identifiers, including private | 380,828 | 383,187 |
| Literals, including template pieces | 18,193 | 139,960 |
| Other, including EOF | 78 | 0 |
| Total | 1,417,393 | 660,939 |

Observed baseline instruction edges for text creation:

| Caller and operation | Calls | Inclusive instructions |
| --- | ---: | ---: |
| Scanner.identifier -> string_slice, identifiers and keywords together | 504,799 | 175,907,871 |
| Scanner.scan -> string_slice, numeric prefix and trivia probes | 14,853 | 5,047,663 |
| Scanner.number -> string_slice | 10,234 | 3,709,059 |
| numberValue -> string_slice | 11,161 | 2,843,148 |
| Scanner.string -> string_slice | 6,566 | 2,036,496 |
| Scanner.rescanGreater -> string_slice | 4,819 | 1,509,414 |
| Scanner.template -> string_slice | 1,289 | 398,312 |
| Scanner.rescanSlash -> string_slice | 92 | 27,285 |

These are disjoint caller edges for the same runtime entry, not extra self
buckets. Literal construction also uses fragments, concat and code-point builders.
**Exact instruction separation of keyword versus identifier copies, and of all
intermediate allocations by final token class, is not measured here.** Giving
both classes 175.91M or scaling by allocation counts would be false precision.
The combined identifier/keyword slice edge is only 1.90% of whole-process parse.
Even removing that entire edge cannot close the measured 5.5x gap. A span
prototype is deferred pending this evidence and the runtime equality experiment;
no scanner source is edited. A borrowed slice would require a proven lifetime
and immutable-source aliasing contract, not merely removal of a retain.

## Tools, exact commands and limitations

Setup passed: Go ready 0s; clang, Node, submodules ready 1s; cache warm and done
103s. `nproc` is 5, cgroup quota 400000/100000, 17.6 GB memory. Go 1.27.1,
clang 20.1.8, Node 24.19.0. Source `/workspace/adamic-tools/env.sh` in each shell.
Valgrind 3.24.0 was extracted into scratch from the Ubuntu package without a
system install. The preserved brk-segment warning is nonfatal; profiles finish
successfully. No sanitizer or ADAMIC_COUNT is enabled for instruction counts.

The exact baseline clang line is:

```sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/parse-speed/parse.c internal/native/runtime/*.c -lm -o /workspace/scratch/parse-speed/native-parse > /tmp/parse-speed-clang.log 2>&1
```

The glob expands to all tracked runtime C translation units in lexical order.
Generated C is preserved as parse.c.gz. Build inputs use the pinned driver,
current main compiler/runtime, and corpus hashes above. Reproduction helpers
and command history are recorded next to this report; scratch instrumentation
has fixed paths for this workspace and is not a compiler change.

```sh
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/parse-speed/native-before.callgrind /workspace/scratch/parse-speed/native-parse --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-speed/native-before.stdout 2> /workspace/scratch/parse-speed/native-before.stderr
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/parse-speed/go-parse.callgrind /workspace/scratch/parse-speed/go-parse --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-speed/go-parse.stdout 2> /workspace/scratch/parse-speed/go-parse.stderr
```

All program and test output goes to files. Self reconciliation is checked by
stage1/typescript/scanner/profile.py. A mutant changing only `summary:` by +1 is
rejected with `callgrind self costs do not sum to summary`; accounting-mutant.log
records it. No performance source change has been made, so no fix mutant or
native/oracle package gate is yet claimed. Full repository gate, exact historical
boundary reproduction, and exact per-final-token-class allocation instructions
remain uncovered. Observations do not establish either dated speed claim.
