Built an isolated exact-byte comparator and an undecided Object.entries design.
Base: origin/main 45487a809f89885a3fc651cd590e7dabf31362dc; branch codex/step12-scout.
Focused checks pass: 12 Node goldens/mutants, two native controls, 168 source-span audits.
Six output mutants and two comparator implementation mutants are caught.
No native scanner claim; enum follow-up pins the expression but not the historical refusal.

# Step 12 scout

The [enum-slot follow-up](enum-followup/REPORT.md) pins upstream scanner.ts:1823,
tests both distinct enum-tag-narrowing tips, and records four new witnesses.
The exact assignment expression remains NotYet; split statements match Node.
It supersedes the original enum candidate as the expression witness, while
keeping the historical full-closure refusal explicitly unconfirmed.

Read [RESEARCH.md](RESEARCH.md) for upstream locations, counts, Node behavior,
Go alternatives, spec references and hard cases. [ENTRIES.md](ENTRIES.md) keeps
proof, runtime-check and refusal options open and ends with questions for
@system_adamic. Nothing here changes the language or production compiler.

The report is bounded by the fetched main. The requested land-area-next report
is absent from that main's scanner evidence. Three unchecked-cast families are
identified in raw source, but the claim of three individual remaining casts and
the unproven-enum stopping shape cannot be reconstructed from absent evidence.
The candidate enum program is explicitly NOT that reproduction. Generated diag's
original explicit result type makes its reduced assertion control green on main.
Uint16's exact reduced compound writes encounter TS2532 before constructor
lowering. Those are limitations, not retired feature stops.

## Comparator package

```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/step12-compare ./stage3/scouts/step12/compare
/tmp/step12-compare NODE_STDOUT NATIVE_STDOUT > comparison.json
```

Exit 0 means every byte and both EOFs match. Exit 1 means unequal output, with
zero-based first byte offset, one-based line/byte column, each byte's integer
value (-1 for EOF), both SHA-256 values, byte counts and newline counts. Exit 2
means invocation, input-read or output-write failure. Hash equality alone never
passes. It continues through both streams after the first mismatch to hash and
count complete inputs. Memory use for input blocks is bounded; arbitrarily long
JSON token rows have no scanner line limit. It works on the historical six-field
509,014-token stream or the expanded coverage stream, without interpreting either.
No speedup over diff is claimed. This is a standalone package, not wired into the
shared scanner driver or progress.json.

## Reproduce the scout

Obtain the pinned TypeScript source in stage3/source.json, generate diagnostics
with stage3/adapt/00-setup/adapt.cjs, and install stage3/api's locked dependencies
in scratch. All new Adamic witness sources are .a.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step12-setup.log 2>&1
source /workspace/adamic-tools/env.sh
SLICE_TYPESCRIPT=API/lib/typescript.js bash stage3/slice/run.sh TREE NEW_SLICE \
  src/compiler/scanner.ts:createScanner src/compiler/types.ts:ScriptTarget \
  src/compiler/types.ts:SyntaxKind --no-adapt > /tmp/step12-slice.log 2>&1
node stage3/slice/verify.cjs NEW_SLICE > /tmp/step12-verify.log 2>&1
SCOUT_TYPESCRIPT=API/lib/typescript.js node stage3/scouts/step12/census.cjs \
  TREE NEW_SLICE > /tmp/step12-sites.json

go build -buildvcs=false -o /tmp/step12-adamic ./cmd/adamic > /tmp/step12-adamic-build.log 2>&1
SCOUT_TYPESCRIPT=API/lib/typescript.js node stage3/scouts/step12/run-fixtures.cjs \
  NEW_OUTPUT /tmp/step12-adamic > /tmp/step12-fixtures.log 2>&1

go test ./stage3/scouts/step12/compare \
  -run 'TestCompare|Test509014RowsAndMutants|TestReadError' -count=1 -v \
  > /tmp/step12-compare.log 2>&1
python3 stage3/scouts/step12/run-compare-mutants.py NEW_MUTANT_LOG_DIRECTORY \
  > /tmp/step12-overlay-mutants.log 2>&1
```

API/lib/typescript.js means the installed typescript package's lib/typescript.js.
The actual run used /workspace/scratch/step12-api/node_modules/typescript/lib/typescript.js,
/workspace/scratch/step12-typescript, and /workspace/scratch/step12-closure.
Fixture outputs use new directories, never overwritten. The final run directory
was /workspace/scratch/step12-fixtures-goldens. Results and raw logs are retained
in evidence; no generated JS or native binaries are committed.

The runner checks pinned Node stdout, empty control stderr, each actual Adamic
exit/diagnostic, and byte equality of every successful native control. It also
requires the local a-check marker. `refused` and `type error TS2532` use the
repository's expected-error convention. `pass` and `NotYet` are informational
local markers, not invented expected-error permissions for the repository Gate.
NotYet is not a language refusal. This runner does not claim a cohere Gate run.

## Checks and mutants actually run

- Comparator named tests passed (0.637s), and its executable built successfully.
  Scale control used 509,014 SYNTHETIC token rows, 11,707,322 bytes. This is
  comparator sensitivity, not a new scanner corpus measurement.
- Four single-byte output mutants at offsets 0, 65535, 65536 and 11707320 were
  caught at their exact offsets. Truncating one byte and appending one byte were
  both caught. NUL/high-bit bytes and both EOF directions are included.
- Two independent Go overlay mutants changed only comparator semantics:
  `l != r` -> `false`, and `l != r` -> `l != r && l != -1 && r != -1`.
  TestCompare's equality assertion failed, exit 1, for each. Neither failure was
  a build warning, syntax error or compiler refusal. Logs are in evidence.
- Twelve fixture controls and twelve input mutants passed the runner. The
  namespace differing-argument mutant throws (Node exit 1); the other eleven
  exit 0 with changed stdout. [counts.md](counts.md) lists every mutation.
- Shared slicer byte/order audit passed: 168 spans, 78 module import lists.
- git diff --check passed before commit. No whole compiler package, full gate,
  full upstream baseline, native scanner lane, sanitizer or ownership check ran.

Setup timings (cumulative): Go 0.092s, Node 0.096s, clang 0.807s, markdown
1.693s (install step 1.373s), submodules 19.536s, Go build 231.655s, deferred
test binaries 231.757s, warm cache 231.758s, done 231.784s. nproc=5;
CPU quota=400000/100000. Setup succeeded; no workaround beyond requested GOPROXY.

No internal/oracle fixture was registered; internal/oracle/counts.md is unchanged.
The scout's own witness counts are refreshed in counts.md. All tracked changes
are confined to stage3/scouts/step12. No upstream adaptation, compiler patch,
progress declaration, native scanner equality, native performance, parser-directed
rescan or feature-owner closure is claimed. The missing snapshot questions remain
for @system_adamic; no ruling or private message was sent.
