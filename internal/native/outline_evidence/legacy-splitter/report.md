Sanitized: main.c 16.919 -> 0.065 s; C-to-native build 17.947 -> 7.656 s. Release: main.c 3.804 -> 0.043 s; build 4.495 -> 1.697 s.
Built: stable per-module initialization wrappers and flat initializer chunks, preserving initialization, readiness and cleanup order.
Commits: implementation f4d39ac659b739e85cc17ac626ce7c0704cbff7b; stable names aca41891 plus main 48c05d09 merged as c4e13365; branch codex/outline-module-main.
Validation: full uncached native/lower/oracle and split oracle gates, fixture parity, sanitizer location canaries and four independent mutants; results below.
Limits: structured source statements remain atomic; timings exclude lowering/emission and cold runtime preparation; default full -g retained.

Same-machine medians of three interleaved runs, ADAMIC_NATIVE_SPLIT=1,
ADAMIC_NATIVE_JOBS=5, ADAMIC_GATE_UNCACHED=1:

| Mode | main.c before | main.c after | C-to-native before | C-to-native after |
| --- | ---: | ---: | ---: | ---: |
| Sanitized, full -g | 16.919 s | 0.065 s | 17.947 s | 7.656 s |
| Release | 3.804 s | 0.043 s | 4.495 s | 1.697 s |
| Sanitized, line tables | 17.444 s | 0.047 s | 18.543 s | 6.529 s |

The input is stage1/cohere/markdownblocks/testdata/list_probe.ts. The before
compiler is the untouched merge c4e13365. Both versions use the existing splitter
and identical clang options, except the explicit scratch line-table variant.
There are ten program translation units before and nineteen after. Runtime archives
are warmed before timing. The build timer covers native.Build, including splitting,
preprocessing, unit compilation and linking; it excludes load/lower/C emission.
main.c times cover the real clang -c process, excluding preprocessing and wrapper
copying. No other tests ran concurrently with these measurements. Raw per-run
values, compiler arguments, generated C and the runner are in outline_evidence.
The historical developer-tools floor (27.835 s sanitized main, 6.882 s release)
used a different compiler baseline. These new before values are measured here,
not inferred from that floor.

main's generated body falls from 19,607 lines to 133. There are 154 initializer
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
function in an unsplit build. The existing splitter packs sixteen consecutive
functions per unit, so it distributes the new helpers without a splitter change.
Module markers accompany the helpers for future source-owned grouping; they do
not change current grouping. No lower/IR or native.go changes.

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
Neither changes production lowering or runtime behavior.

Sanitizer runs with full -g and -gline-tables-only both catch a deliberately early
release of an allocated array. Both reports retain named initializer frames and
generated C file/line locations. This is generated-source readability, not a new
Adamic source map. Sanitizer instrumentation and recovery flags are unchanged.
Line tables reduce this outlined build median by 14.7%, but do not improve the
large unoutlined main in these runs. Recommendation: offer line tables as an
explicit sanitizer/CI option; retain full -g by default for debugger local/type
information. The implementation changes no debug defaults.

Reproduce measurements and location checks with the setup toolchain sourced:

```
python3 internal/native/outline_evidence/reproduce.py
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

The order mutant produces main/inline/a/b/c/shared instead of Node's
shared/c/b/a/inline/main. None of these four was caught by a compiler error.
The sanitizer readability canary separately injects an early release of an
allocated array; ASan catches heap-use-after-free with both debug variants.

Setup: bash cloud/setup.sh succeeded; source /workspace/adamic-tools/env.sh.
Timing lines: go 0.025 s, Node 0.025 s, clang 0.189 s, markdown dependency
installation step 0.812 s (ready 0.890 s), submodules 3.366 s, go build
183.331 s, cache warm 183.517 s, done 183.548 s. nproc=5, cpu.max=400000
100000; Go 1.27.1, clang 20.1.8, Node v24.19.0. Raw setup.log retained.

No default debug flag changes, splitter changes, source-map implementation,
recursive outlining of structured control flow, or full-repository test gate
are claimed. The complete oracle package is run in each native mode; the
markdownblocks coverage is the filtered list layout fixture, not its full suite.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/lower ./internal/oracle -count=1 -timeout=30m
  PASS native 452.938 s; lower 57.023 s; oracle 431.376 s; exit 0
ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 go test ./internal/oracle -count=1 -timeout=30m
  PASS oracle 450.646 s; exit 0
go vet ./...
  exit 0, no diagnostics
gofmt -l <five changed Go files>
  no output
git diff --check
  exit 0, no output
python3 internal/native/outline_evidence/mutants.py
  four intended failures caught; runner exit 0
OUTLINE_MEASURE=/tmp/outline-module-main go test -overlay=/tmp/outline-module-main/debug.json ./internal/native -run '^TestOutlinedSanitizerLocations$' -v -count=1
  PASS 4.16 s; full and line variants report generated file and line
```

Gate wall times above include concurrent validation workloads and are not
performance measurements. Measurement runs were isolated. The native gate
includes split/unsplit and sanitized/release checks for multi-chunk object and
number literals, closure capture, and a nonglobal reference surviving across
chunks/module boundaries into final cleanup. Full oracles include module order,
readiness, cycles, exception behavior and recorded allocation counts.
