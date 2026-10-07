The CSS parser and printer now compose natively after merging the cycle proof.
The optimized printer matches all 48,152 Go answers across two option sets.
The composed parser matches 24,076 Go trees and error positions on all backends.
Sanitizer, leak, native output mutant and throughput evidence is recorded below.
The original library boundaries and CSS/SCSS scope remain unchanged.

# Change and provenance

On `codex/stage1-css-printer`, merged requested cycle-proof commit
`32f8106613fb7708ce7192e04f7f1799b35c8434` in merge `7ef0344`.
The accepted printer remains `b902761`, with prior report `27a14ce` and regex
integration `2a3416f`. This unit changes verification and gap documentation,
not parser/printer production semantics, recursive ownership, or compiler/runtime
source. The four protected compiler files have no hand edits.

The former cycle-refusal tests are now executable native regressions. Both
small proving programs still produce the original Node output, now also on
native ASan/UBSan and the JavaScript backend with separate leak checks.
The complete parser and printer are held against their external Go oracle,
not merely tested for successful compilation.

# Setup

Ran `bash cloud/setup.sh`, then sourced `/workspace/adamic-tools/env.sh`.
`nproc` is 5; CPU cgroup is four cores, memory 17.6 GB. Timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (89s)
setup: done in 89s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. The explicit single-branch fetch
populated FETCH_HEAD without creating a remote-tracking ref; merging the
requested full commit SHA then succeeded. Both complete CLI builds exited
successfully with no output. The toolchain setup succeeded.

# Corpus and agreement

Unchanged corpus: 12,038 texts, 24,076 CSS/SCSS pairs, including every collected
cohere test selector/stylesheet constant and concatenation, repository and
submodule CSS/SCSS/Less, pinned fork fixtures, generated cases, comments,
custom properties, and malformed Unicode-boundary truncations.
The file walker finds 158 CSS, 90 SCSS and 43 Less files. External fixtures are
`/tmp/adamic-css-prettier` at `cb4b33fba24a8428d00e54be85fc886288a374ea`.
Corpus construction, seed and scratch checkout commands remain in `REPORT.md`.

The composed parser has 5,006 successes. All 24,076 canonical trees and error
positions agree with Go on native ASan/UBSan, source Node and the JavaScript
backend. A separate LeakSanitizer run is clean. Optimized native parser output
also matches Node on every case, transitively held to the same Go tree oracle.

The printer has 4,966 formats per option set. Default is width 80, tab width 2,
spaces, double quotes, trailing commas all. Narrow is width 24, tab width 4,
tabs, single quotes, trailing commas none. The optimized native printer matches
all 24,076 Go results per set, including exact refusals and their positions.
All 48,152 printer results agree on sanitized native, Node and the JavaScript
backend too. Separate LeakSanitizer runs are clean for both option sets.

Both npm Prettier 3.9.6 and cohere's complete bundled Prettier remain independent
full-format oracles. Each option set/oracle has 4,952 byte-identical formats,
19,080 shared refusals and 44 exactly recorded boundary occurrences: BOM 12,
carriage return 2, nonbreaking space 2, YAML delegation 28. `GAPS.md` explains
these different entry-point behaviors. No new difference is admitted; errors
from Prettier's public code-frame API are not compared to Go's internal API.
The two pre-existing raw PostCSS surrogate discrepancies remain unchanged.

# Mutants and check strength

Three source printer mutants must compile and run with clean stderr and exit 0
before their output can count as caught. They now run under native ASan/UBSan
and Node in both option sets:

| Mutant | First output difference |
|---|---|
| Declaration semicolons omitted | Line 2, byte 69 |
| Rule body indentation omitted | Line 6, byte 7 |
| Document groups ignore remaining width | Line 26, byte 12 |

Existing parser mutants continue to run on raw native and Node, and on composed
Node. The public Range-corruption proof now also runs composed native and the
JavaScript backend and prints `caught`, with a clean leak check.

`TestComposedMemoryChecksCanFail` mutates only temporary generated C artifacts.
All three mutations preserve ordinary optimized printer bytes on two dynamic
file inputs. Only the intended memory check catches each:

| Mutation | Observation |
|---|---|
| Read allocated storage after freeing it | ASan reports heap-use-after-free |
| Overflow a volatile signed integer with argc | UBSan reports signed integer overflow |
| Omit generated releases | Output and ASan/UBSan with leaks off pass; LSan reports 644,428 bytes in 8,903 allocations |

The access and arithmetic probes are instrumented harness checks, not newly
found printer defects. The release mutant affects actual composed printer
allocations. No erroneous artifact or source modification is deployed.

The merged cycle fixture's measured counts row is recorded: 97 allocations,
97 frees, 65 retains, 81 releases, peak 35, regions 0. The full counts gate was
not run. The whole `internal/fresh` package and filtered native regex/cycle
oracles pass; future unproved regex effects stay refused.

# Commands and logs

All test output goes directly to files, never through a pipe. Commands source
the environment file above and use these scratch variables:

```sh
export ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier
export ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library
export ADAMIC_CSS_PRINTER_LIBRARY=/tmp/adamic-css-printer-library

go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^(TestCompositionMatchesGo|TestCSSPrinterAgreesWithGo)$'
go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^(TestClosed.*RegexGap|TestTheCanonicalRangeChecksCanFail|TestEachGapStandsWhereGapsMdSaysItDoes)$'
go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^(TestThePortParsesAsGoCohereDoes|TestCSSPrinterBoundaryProofs)$'
go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestComposedMemoryChecksCanFail$'
go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSPrinterOptimizedMatchesGo$'
go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSParserOptimizedMatchesNode$'
go test -count=1 ./internal/fresh
go test -count=1 -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp'
go test -v -count=1 -timeout=30m ./internal/fresh ./internal/oracle -run 'TestRegex|^TestNativeAgreesWithNode$/internal/fresh/testdata/regexp_tree.ts$'
go run ./cmd/adamic build internal/fresh/testdata/regexp_tree.ts -o /tmp/css-cycle-count --count
/tmp/css-cycle-count
go vet ./...
gofmt -l stage1/cohere/css
git diff --check
```

Final native agreement PASS in 1,415.352s: parser test 114.44s, printer test
1,300.90s (default 579.86s, narrow 716.15s). All three native printer mutants
are caught in both option sets. Other package outputs: optimized printer PASS in 43.786s,
optimized parser PASS in 14.583s, memory-check mutants PASS in 51.444s,
closed-gap/range/workaround proofs PASS in 21.291s, raw-parser and boundary
regressions PASS in 77.571s, full fresh package PASS in 14.124s, regex oracle
PASS in 1.778s, merged cycle fixture oracle PASS in 0.678s.
`go vet ./...`, gofmt and diff checks produce no diagnostics.

Logs are in `verification/native-*.log`. The complete repository gate was not
run; the touched CSS and cycle-proof packages plus filtered oracles were run.

# Throughput protocol

Native is the complete optimized parse-and-print binary (`-O2`, sanitizers off).
Compare Go cohere, source on Node, npm Prettier and bundled Prettier on the same 4,952
successful exact-output CSS/SCSS cases. Use three isolated rounds, one pass
per runtime/round, 510,298 UTF-16 output units. Each checksum must equal the
Go oracle; the optimized artifact's full output is independently held above.
One-pass timing bounds the slow native run while retaining the complete shared
successful corpus. The default benchmark mode still supports ten repetitions.

Native and Node timings include process startup and case decoding. Go timing
measures in-process parse and print after decoding, with its command startup
excluded. No test or compilation runs concurrently with measured rounds.
This measures entire stylesheet formatting, not printing an already built tree.

```sh
ADAMIC_CSS_PRINTER_BENCH=1 ADAMIC_CSS_PRINTER_BENCH_ONCE=1 go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSPrinterThroughput$'
```

Final throughput PASS in 77.697s. Every measured round formats 4,952
stylesheets and produces the expected 510,298 output units.

| Runtime | Round 1 / s | Round 2 / s | Round 3 / s | Median / s |
|---|---:|---:|---:|---:|
| native | 790 | 758 | 741 | 758 |
| Go | 5870 | 5875 | 5799 | 5870 |
| Adamic source on Node | 1430 | 1406 | 872 | 1406 |
| Prettier fork on Node | 1291 | 1424 | 1293 | 1293 |
| Prettier npm on Node | 1313 | 1362 | 1279 | 1313 |

Native is slower on this corpus: approximately 7.7 times below Go's
measured throughput. This is an observation, not a diagnosis of the cause;
no performance profile was collected. The one-pass comparison includes
startup for native and Node but excludes it for Go, as described above.

Native verification commit: `a45ffb7`; cycle-proof merge: `7ef0344`.
This throughput report and log are committed separately. Logs have trailing
whitespace trimmed for review. The working tree is pushed without a pull request.

# Scope

CSS/SCSS only: Less files are exercised through those supported grammars, not
claimed as Less semantics. YAML delegation, HTML style attributes, CSS-in-JS
placeholder orchestration, private files, incremental edits, arbitrary invalid
UTF-8 output and public Prettier code frames remain outside this slice.
Prior reports describe the historical Node-only state; this report supersedes
those native-composition limitations. No assertion of native speed superiority
is made without the measured results below.
