Sanitized native build: 15.690 -> 7.136 s; release: 4.208 -> 2.979 s. Full compiler samples: 26.136 -> 17.720 s sanitized; 15.012 -> 13.569 s release.
Built: stable per-module initialization wrappers and flat initializer chunks, preserving initialization, readiness and cleanup order.
Commits: implementation f4d39ac6 and a3391884; current main ce0750f2 merged as fda4bf50; stable names aca41891 plus main 48c05d09 merged as c4e13365; stable splitter 99814057 merged as 742d841d; branch codex/outline-module-main.
Validation: full uncached native/lower/oracle and split oracle gates, fixture parity, sanitizer location canaries and eight independent mutants; results below.
Limits: structured source statements remain atomic; native medians exclude lowering/emission; full compiler samples include them; cold runtime preparation excluded; default full -g retained.

Same-machine medians of three interleaved runs, ADAMIC_NATIVE_SPLIT=1,
ADAMIC_NATIVE_JOBS=5, ADAMIC_GATE_UNCACHED=1:

| Mode | main.c before | main.c after | Slowest unit before | Slowest unit after | C-to-native before | C-to-native after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sanitized, full -g | 0.039 s | 0.039 s | 14.213 s | 3.525 s | 15.690 s | 7.136 s |
| Release | 0.037 s | 0.023 s | 3.123 s | 0.306 s | 4.208 s | 2.979 s |
| Sanitized, line tables | 0.046 s | 0.044 s | 13.695 s | 3.304 s | 15.215 s | 5.845 s |

The stable splitter already extracts module initialization from main. Its before
bottleneck is now widthTables.ts's module unit, which is also the slowest after
unit. The main.c figure therefore stays small; the improvement comes from smaller
functions and distributing large initialization across module-local buckets.

Full frontend-to-native builds, one isolated sample per version/mode:

| Mode | Before | After |
| --- | ---: | ---: |
| Sanitized, full -g | 26.136 s | 17.720 s |
| Release | 15.012 s | 13.569 s |
| Sanitized, line tables | 25.910 s | 16.580 s |

These timers include load.Load, lower.Lower, native.C and native.Build, excluding
Go test startup and separately warmed runtime archives. Each emitted input is
checked byte-for-byte against the saved C used for the native-only measurements.
These single observations are not medians.

The input is stage1/cohere/markdownblocks/testdata/list_probe.ts. The before
emitter is the untouched merge c4e13365. Both versions use the merged stable splitter
and identical clang options, except the explicit scratch line-table variant.
There are 58 program translation units before and 80 after. Runtime-source
compiles logged during untimed warming are excluded from these counts. Runtime archives
are warmed before timing. The build timer covers native.Build, including splitting,
preprocessing, unit compilation and linking; it excludes load/lower/C emission.
main.c times cover the real clang -c process, excluding preprocessing and wrapper
copying. No other tests ran concurrently with these measurements. Raw per-run
values, compiler arguments, generated C and the runner are in outline_evidence.
The historical developer-tools floor (27.835 s sanitized main, 6.882 s release)
used a different compiler baseline and splitter. These new before values are
measured here, not inferred from that floor. Positional-splitter measurements
preceding the requested merge are archived under outline_evidence/legacy-splitter
and are not the final acceptance numbers.

Before splitting, main's emitted body falls from 19,607 lines to 133. There are 154 initializer
functions. Each flat chunk contains at most 128 generated C statements. Large
literal initializers are divided inside a single source statement. The largest
remaining helper is 1,050 lines of structured list_probe.ts control flow: loops,
conditionals, exception handlers and lexical blocks remain atomic. This bounds
the measured table initializer bottleneck; it does not impose a hard size limit
on an arbitrarily large structured source statement.

The named assembly hook is emitter.moduleMain in outline.go, called by C in
emit.go. It emits the same IR statements and scope releases, then
emitter.outlineInitializers moves their generated text into stable module-qualified
wrappers and chunks. Prototypes carry noinline so clang cannot rebuild the huge
function in an unsplit build. Module markers now accompany helper definitions,
not their prototypes. main emits direct calls without markers, so the splitter
cannot re-extract those calls and move surviving locals into a different frame.
No lower/IR or native.go changes.

Multi-chunk initializers use the adamic_initialize_chunk_ prefix. The splitter
places them in eight fixed buckets within their real source module, selected by
SHA-256 of the stable helper name. A module's chunk index is local to that module;
there is no program-wide ordinal or positional group. A single-chunk initializer
and its wrapper stay in the ordinary module unit. Global storage and its ready
flag remain owned by the original module even when the ready write is in a bucket.
The first merged-splitter prototype kept every chunk in its module unit and did
not remove the serial compile bottleneck; the final numbers include the placement
hook. This is an observed need for the hook, not a claim about splitter internals.

Only units_stable.go has splitter edits. Every changed hook line is identified:

- Line 15: initializationOwner records the original module's ownership.
- Lines 115, 116: explain the initializer-only noinline declaration exception.
- Line 117: limit the exception to initializer function definitions.
- Line 118: remove only the exact noinline suffix for the comparison.
- Lines 119, 120, 121: continue rejecting every other declaration mismatch.
- Line 145: recognize large-initializer chunk symbols.
- Lines 146, 147: explain fixed placement and retained storage ownership.
- Line 148: save the original module owner.
- Line 149: hash the stable helper name.
- Line 150: choose one of eight module-local initializer unit filenames.
- Line 151: close that chunk-only placement hook.
- Line 217: explain already outlined readiness ownership.
- Line 218: allocate the readiness scan set.
- Lines 219, 220, 221: retain the splitter's existing extracted helpers.
- Lines 222, 223, 224, 225, 226: include source-marked emitted initializers.
- Line 227: scan the combined set with the existing ready-write logic.
- Lines 237, 238, 239: use the original module for bucket-owned readiness writes.

No edits to units.go, units_stable_test.go or other splitter files. Existing
ABI guards, selective dependencies, object keys and ownership rules remain in use.

Locals spanning chunks travel as typed address parameters. Module-owned storage
uses typed backing arrays to keep wrapper declarations small. Locals needed by
another module or final scope cleanup stay in main. No new retain/release,
readiness write or initialization reordering is introduced. String/comment bytes
are preserved during token-based substitution. Structured source statements keep
their labels and handler frames together. Unsupported flat declarators fail closed.

The merge needed two test adaptations: native's imported-specialization fixture
now gives its two Box classes different private fields, preserving nominal
identity under main's newer structural canonicalization; the enum cleanup mutant
locates the emitted stable symbol instead of an obsolete whole-program ordinal.
The module-body stability assertion now compares outlined initializer definitions
instead of expecting markers inside main. None changes production lowering or
runtime behavior.

Sanitizer runs with full -g and -gline-tables-only both catch a deliberately early
release of an allocated array. Both reports retain named initializer frames and
generated C file/line locations. This is generated-source readability, not a new
Adamic source map. Sanitizer instrumentation and recovery flags are unchanged.
Line tables reduce this outlined build median by 18.1%; the before improvement is small
(15.690 -> 15.215 s). Recommendation: offer line tables as an
explicit sanitizer/CI option; retain full -g by default for debugger local/type
information. The implementation changes no debug defaults. Current main later added an
inherited-signal reset in adamic_start; it is merged before the final re-green.
The isolated timings above were taken before that runtime-only refresh, on the
requested 48c05d09 baseline. Loading/lowering/emitted C and splitter code are
unchanged by that refresh; runtime archives are separately warmed in the timings.

Reproduce measurements and location checks with the setup toolchain sourced:

```
python3 internal/native/outline_evidence/reproduce.py --end-to-end
python3 internal/native/outline_evidence/summarize.py <printed scratch directory>
python3 internal/native/outline_evidence/mutants.py
```

The first script writes scratch overlays, generated units, timings and reports.
It uses saved before/after compiler inputs, not copied cohere source files. The
second overlays only outline.go and requires each intended failure diagnostic.

Final commands and outputs are recorded below and in outline_evidence logs.

Mutants (each applied independently to outline.go through a Go overlay):

| Mutation | Check that caught it | Observed failure |
| --- | --- | --- |
| Reverse module wrapper calls | import_cycles/order/main.a vs Node | stdout differs; both exit 0 |
| Raise flat chunk bound to 100000 | TestInitializerChunksKeepStorageAndOrder | large initializer was not outlined |
| Remove noinline | TestInitializerChunksKeepStorageAndOrder | large initializer was not outlined |
| Omit final root scope cleanup | TestInitializerChunksKeepStorageAndOrder under ASan | LeakSanitizer: detected memory leaks |
| Disable chunk bucket placement | TestInitializerUnitOwnership | large initializer stayed in its module unit |
| Move global storage into its chunk bucket | TestInitializerUnitOwnership | initializer global table owned by module bucket |
| Ignore outlined readiness writes | TestInitializerUnitOwnership | initializer global table owned by main.c |
| Discard retained noinline prototype | TestInitializerUnitOwnership | module initializer lost noinline prototype |

The order mutant produces main/inline/a/b/c/shared instead of Node's
shared/c/b/a/inline/main. None of these eight was caught by a compiler error.
The sanitizer readability canary separately injects an early release of an
allocated array; ASan catches heap-use-after-free with both debug variants.

Setup: bash cloud/setup.sh succeeded; source /workspace/adamic-tools/env.sh.
Timing lines: go 0.025 s, Node 0.025 s, clang 0.189 s, markdown dependency
installation step 0.812 s (ready 0.890 s), submodules 3.366 s, go build
183.331 s, cache warm 183.517 s, done 183.548 s. nproc=5, cpu.max=400000
100000; Go 1.27.1, clang 20.1.8, Node v24.19.0. Raw setup.log retained.

No default debug flag changes, source-map implementation,
recursive outlining of structured control flow, or full-repository test gate
are claimed. The complete oracle package is run in each native mode; the
markdownblocks coverage is the filtered list layout fixture, not its full suite.

Final gates on fda4bf50 (current main plus stable splitter and both implementation commits):

```
ADAMIC_NATIVE_SPLIT=0 ADAMIC_GATE_UNCACHED=1 go test -json ./internal/native ./internal/lower ./internal/oracle -count=1 -timeout=30m
  PASS native 602.002 s; lower 66.316 s; oracle 541.444 s; exit 0
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test -json ./internal/oracle -count=1 -timeout=30m
  PASS oracle 671.243 s; exit 0
go vet ./...
  exit 0, no diagnostics
python3 internal/native/outline_evidence/mutants.py
  all eight intended failures caught; runner exit 0
OUTLINE_MEASURE=/tmp/outline-module-main/stable-final go test -overlay=/tmp/outline-module-main/debug.json ./internal/native -run '^TestOutlinedSanitizerLocations$' -v -count=1
  PASS 93.948 s; both modes retain generated module file and line
```

The full native package includes the unchanged-body stability regression,
split/unsplit and sanitized/release multi-chunk object/number literal checks,
closure capture and a nonglobal reference surviving chunks/modules into final
cleanup. The complete oracle package includes import order, readiness, cycles,
exception behavior, inherited signals and recorded allocation counts.

The filtered port gate also passed:

```
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownListLayout$' -v -count=1 -timeout=30m
  PASS 1346.486 s; 5166 source documents; Go/source Node/backend/native/original bytes identical
  unordered_marker, task_box and ordered_cap output-only mutants all caught
```

That port gate began before the runtime-only main refresh. Its generated native
and canonical input files were preserved and hashed. After the refresh, the
fixture_parity.go.txt overlay rebuilt the saved after C with the current runtime
in sanitized, release and line-table modes, and compared each against fresh Go,
source Node and original Node library observations. All six observations agree
on 13,029,128 output bytes and the same SHA-256; all stderr is empty. PASS 146.888 s.
See fixture-current-runtime.log and fixture-inputs.json. These are oracle inputs
and observations, not new source copied from cohere. The Go/original-Node oracles
reference the pinned cohere submodule. Regenerate corpus parity with the filtered
port command above; the scratch parity overlay requires its saved generated
native.txt, canonical.txt and Go oracle binary.

The port gate's three existing layout mutants are additional coverage, separate
from the eight compiler mutants listed above. The early-release array canary is
also separate; ASan catches heap-use-after-free in both debug variants.

Gate times include concurrent validation workloads and are not performance
measurements. Isolated timing runs had no concurrent tests. The generated corpus
includes repository Markdown, so adding this report changes the corpus count;
input hashes identify the exact observed snapshot.


Final formatting: gofmt -l on the six changed Go files produced no output;
git diff --check exited 0. Evidence scripts pass Python syntax parsing.
Current origin/main ce0750f2 is an ancestor of the branch at the final fetch.
The source commits are f4d39ac6 and a3391884; dependency merges are c4e13365
(main 48c05d09), 742d841d (splitter 99814057), and fda4bf50 (current main).
