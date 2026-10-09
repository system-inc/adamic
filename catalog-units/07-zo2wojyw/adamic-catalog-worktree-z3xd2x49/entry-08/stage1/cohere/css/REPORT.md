Native-composition follow-up: see [NATIVE_REPORT.md](NATIVE_REPORT.md). The
composition refusal below describes the original parser commit.

CSS and SCSS tokenizer/parser run natively; full slice composition runs on Node.
24,076 raw and composed answers match Go; raw native sanitizers and leaks pass.
PostCSS has 24,074 exact agreements and two occurrences of one proved surrogate gap.
Three semantic mutants are caught; native composition is refused by the regexp cycle-proof gap.
Branch: `codex/stage1-css`; the final commit SHA and push result are in the handoff.

## What is built

`input.ts`, `tokenize.ts`, `parser.ts` and `nodes.ts` port cohere's PostCSS 8.5.16
and postcss-scss 4.0.9 tokenizer/parser machinery. They preserve own-property
creation order, raw whitespace and comment maps, important flags, nested SCSS
declarations, interpolation, errors and error positions. Tokenization walks
UTF-16; public offsets and source ranges follow Go's byte conventions, including
BOM removal, surrogate cuts and positions beyond the end of the input.

`compose.ts`, `parse_value.ts`, `convert.ts`, `tree.ts` and `location.ts` port the
CSS/SCSS composition. They reuse the existing selector, value and media-query
slices, group value arguments, preserve raw url arguments, parse selector()
arguments, handle SCSS directives, front matter and custom-property blocks,
and compute nested byte locations. A single arena owns objects; all graph links
are integer handles, with no strong parent cycles.

`main.ts` drives the working native raw parser. `compose_main.ts` drives the
composed parser on Node. The complete composed program is refused natively
before clang; it is not presented as a working native formatter parser. The
smallest independent reproduction and all other workarounds are in
[GAPS.md](GAPS.md).

No compiler-owned file, existing slice or submodule source was edited. Go
oracles run with `go test -overlay` in the unmodified cohere submodule.

## Branch and environment

Fetched `origin/main` was `fe3b9f236e0672e968bcb7ad53c1882cdde18f29`, which did not
yet contain the selector/native-regexp work. The new branch was cut from that
main and fast-forwarded to the already pushed selector branch,
`c717c5d2b7a426757e04b724b4bc4bb5ca1c400e`, to carry the required dependencies.
The CSS commit changes only `stage1/cohere/css/`.

`nproc` printed `5`. `bash cloud/setup.sh` printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (18s)
setup: done in 18s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Tools: Go 1.27.1, Node 24.19.0, clang 20.1.8. Environment:
`source /workspace/adamic-tools/env.sh`; TMPDIR is the world-traversable
`/tmp/adamic-gate`. Cohere is at `715ba94f3608a6500086b1076ce5cb7e51b836db`.

## Oracles and corpus

Install the original libraries outside the repository:

```sh
npm install --prefix /tmp/adamic-css-library --no-audit --no-fund postcss@8.5.16 postcss-scss@4.0.9
export ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library
```

The library runner checks both package versions. The Go raw canonical print
includes every own field, property order in `keys`, byte source positions,
and the existing Go assertion that Range equals source offsets. The Go composed
canonical print includes every field and cohere's Range assertion, with byte
positions retained rather than approximated as UTF-16. The raw port is compared
against Go on native, Node source and the JavaScript backend; native runs ASan,
UBSan and a separate LeakSanitizer pass. The composed source is compared against
Go on Node only.

The corpus takes every string constant and constant concatenation from cohere's
CSS tests, its explicit CSS/SCSS parser fixtures, every `.css`, `.scss` and `.less`
file found recursively in the repository and submodules, and the corresponding
Prettier fork fixtures. It adds 6,000 generated inputs with seed 20261006,
custom-property/comment combinations, and every Unicode-boundary truncation of
the CSS parser fixtures. Both CSS and SCSS parse every input, including malformed
ones. There are 12,038 texts, 24,076 dialect/input pairs, 5,021 successful raw
parses, and 5,006 successful composed parses. The fixture walker found:

| Extension | Files |
|---|---:|
| CSS | 158 |
| SCSS | 90 |
| Less | 43 |

157 CSS files, 90 SCSS files and 43 Less files come from the actual
`system-inc/prettier` fork at `cb4b33fba24a8428d00e54be85fc886288a374ea`; one CSS
file comes from this repository/submodules. The fixture checkout is external:

```sh
git clone --depth 1 --filter=blob:none --sparse https://github.com/system-inc/prettier.git /tmp/adamic-css-prettier
git -C /tmp/adamic-css-prettier sparse-checkout set tests/format/css tests/format/scss tests/format/less tests/format/js/multiparser-css tests/format/js/template-literals
export ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier
```

The named commit is the observed fixture version; a later moving HEAD can add
cases. Missing fixture configuration is explicitly logged, not presented as
complete fixture coverage. An explicitly supplied missing path fails the test.

Original PostCSS agrees on 24,074 inputs. The two disagreements are the exact
Go/JavaScript surrogate-cut gap proved and described in GAPS.md. The full
composition was checked against Go, not against Prettier's bundled composition.

## Three mutants

Each mutant compiles and terminates; compiler diagnostics are not counted as a
catch. The raw comparison catches every mutant on both native and Node, and
the composed comparison catches it on Node too.

| Mutant | First raw difference | First composed difference |
|---|---|---|
| `custom = false`, so custom properties lose their special block/token handling | line 162, byte 396: custom-property whitespace/raw value differs | line 162, byte 1001: custom block becomes a separate rule |
| Drop an embedded comment which raw() must keep | line 3126, byte 302: `c/* x */d` becomes `cd` | line 3126, byte 77: raw value loses the embedded comment |
| Do not include the closing brace byte in a block's source end | line 6, byte 659: end offset 28 becomes 27 | line 6, byte 2264: source/endOffset is one byte short |

The checks also hold all five proving programs and the complete native
composition refusal. Once that refusal closes, the gap test intentionally fails
until full native composition agreement and memory checks are enabled.

## Validation commands

All test output was written to log files, not piped. Exact final commands and
results are recorded below and in the checked-in verification logs.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier go test -v -count=1 -timeout 30m ./stage1/cohere/css > /tmp/css-final-tests.log 2>&1
go vet ./... > /tmp/css-vet.log 2>&1
go test -count=1 -timeout 30m ./stage1/cohere/values ./stage1/cohere/mediaquery > /tmp/css-reused-slices.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp' > /tmp/css-regexp-oracle.log 2>&1
node stage1/cohere/css/gaps/upstream_surrogate.mjs /tmp/adamic-css-library > /tmp/css-surrogate-proof.log 2>&1
```

Final correctness results:

```text
ok github.com/system-inc/adamic/stage1/cohere/css 96.189s
ok github.com/system-inc/adamic/stage1/cohere/values 112.973s
ok github.com/system-inc/adamic/stage1/cohere/mediaquery 93.983s
ok github.com/system-inc/adamic/internal/oracle 40.045s
```

`go vet ./...`, `gofmt -l cmd internal stage1/cohere/css` and `git diff --check`
produced no output and exited 0. The raw native and JavaScript backend,
ASan/UBSan/LeakSanitizer checks, all three raw and composed mutants, five gap
proofs and the full native composition refusal passed. After the explicit Range
invariants were added, the final slice run passed in 96.189s.

The public Range corruption probes additionally leave all source properties
intact and change only the Range. They prove that the raw canonical guard catches
it on native and Node, and the composed guard catches it on Node. The raw probe
is leak-clean. Command/result:

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/css -run '^TestTheCanonicalRangeChecksCanFail$' > /tmp/css-range-probes.log 2>&1
```

```text
public Range corruption caught on raw native, raw Node and composition Node; raw leak clean
ok github.com/system-inc/adamic/stage1/cohere/css 7.055s
```

The complete correctness run predates only these added Range test drivers, which
were checked separately; production source is the same in both runs.

## Coverage limits

Native composition, its JavaScript backend, its sanitizers/leaks and its native
throughput are blocked by the documented compiler refusal. Raw parser throughput
is not formatter throughput. Less-specific syntax/composition is not ported in
Go or here. Invalid UTF-8 cannot be carried through the UTF-8 text/JSON oracle
transport and is rejected by the corpus generator. The private 26-file
application corpus named in cohere's tests was unavailable; the public fork,
repository/submodule files and inline tests were included instead. CSS-in-JS
JavaScript files were fetched but no additional template extraction is claimed.
The full repository test gate was not rerun; the CSS slice, reused slices and a
filtered native regexp oracle were run, plus repository-wide vet.

## Isolated raw-parser throughput

After all other tests and compilation finished, the final source was measured
alone. Command:

```sh
ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier ADAMIC_CSS_BENCH=1 go test -v -count=1 -timeout 30m ./stage1/cohere/css -run '^TestCSSThroughput$' > /tmp/css-bench-final.log 2>&1
```

Each measurement parses the same 24,076 dialect/input pairs ten times. Every
side reports `50210 of 240760 stylesheets parsed, 95850 nodes`. Rates are
stylesheets per second, including rejected inputs:

| Parser | Round 1 | Round 2 | Round 3 | Median |
|---|---:|---:|---:|---:|
| Native Adamic | 45,425 | 50,825 | 46,269 | 46,269 |
| Node running the port source | 78,785 | 80,345 | 72,750 | 78,785 |
| Original PostCSS / postcss-scss on Node | 56,404 | 49,560 | 53,951 | 53,951 |
| Go cohere, one in-process sample | 111,694 | | | 111,694 |

The benchmark test passed in 45.484s. Go's loop took 2.155527868s. Native is
slower than both Go and original Node on this corpus. Native/Node timings include
process startup and reading/decoding the transport; Go's timer surrounds only
the parse loop. The native failure path constructs the canonical error message
as JSON, while Go returns an error object. These are observed implementation
costs, not a claim of equal work or formatter performance. About 79% of this
edge-case corpus is refused; these rates should not be extrapolated to a
successful-only real-file workload. No composed-native rate is reported.

## Retained evidence

- [Complete CSS correctness log](verification/css.log)
- [Range corruption probes](verification/range-probes.log)
- [Isolated throughput log](verification/throughput.log)
- [Reused-slice regression log](verification/reused-slices.log)
- [Filtered regexp oracle log](verification/regexp-oracle.log)
- [Original-library surrogate proof](verification/upstream-surrogate.log)
- [Setup timings](verification/setup.log)
