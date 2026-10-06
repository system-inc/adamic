# YAML port: lexer and CST checkpoints

The Adamic lexer, CST parser, and their comparison drivers are implemented.
All 36 repository files and 8,696 generated/chunked cases match Go byte for byte.
Native sanitizers, Node source, emitted JavaScript and yaml 2.9.0 all agree.
Six parser-layer mutants are caught on native and Node; four stage 0 gaps have proving programs.
Composition, unist conversion, the printer and formatting driver remain unfinished.

## Green CST step

`cstParser.ts`, `cst.ts` and `collectionItem.ts` port the CST parser, token
classification and collection item representation. `cst_main.ts` serializes the
entire tree with field presence, UTF-16 sources and offsets, and line starts.
Children are numeric indexes into a parser-owned token arena. Optional arrays
and keys have separate presence flags; empty separator arrays remain present,
and an explicit null key remains distinct from an absent key.

The same 8,732 cases produced 8,198,809 exact bytes from Go, native with
ASan/UBSan/LeakSanitizer, raw Node source, emitted JavaScript and yaml 2.9.0.
The combined lexer/CST/gap/mutant suite passed in 41.094s; the CST comparison
itself took 13.60s. Three additional mutants compile and exit zero with empty
stderr on native and Node, and the byte comparison catches each:

| Mutation | First differing output byte |
| --- | ---: |
| Key presence lost, including explicit null keys | 8086 |
| Source token offset advanced by one | 15 |
| Recorded line start shifted back by one | 180490 |

[emptyAlternative.ts](gaps/emptyAlternative.ts) proves the additional
`lower.NotYet` observation `an array of never`: an untyped empty array in a
conditional branch does not lower, though Node prints `1`. The workaround is a
separately initialized, explicitly typed array followed by conditional assignment.
The earlier GraphQL-era forward class method limitation is closed on this main:
a scratch forward-method probe compiled successfully. No new refusal is claimed
for that behavior.

Cohere lint/type passed on all four new files, 276 rules, 100% Adamic-ready.
In-place list and token updates have explicit `@mutates` ownership contracts;
the output builder owns its parts. No cohere, compiler or runtime source changed.
Logs: [combined suite](audit/cst-suite.log), [CST lint](audit/cst-lint.log).
The existing throughput section still measures only lexical drivers.
Document composition and printing have not been implemented or measured yet.

## Green lexer step

`lexer.ts` ports yaml 2.9.0's lexer using the same UTF-16 units as Go cohere.
Eager token arrays replace generators. Scalar calls remain eager: a block header
at end of input must emit an empty scalar even after the last source unit has
been consumed. A first rewrite incorrectly made those calls deferred states;
the generated comparison caught the missing token and the final port restores it.

`lex_main.ts` reads an escaped batch and prints token units in hexadecimal.
This is a lexer comparison driver, not the requested file-formatting driver.
Each generated input is also parsed in chunks of one, two and seven UTF-16 units;
these exercise splits through quoted escapes, markers and astral characters.
The deterministic seed is 20261006. The corpus walk excludes only `.git`,
includes submodules, and fails on invalid UTF-8 rather than replacing it.

The final suite passed in 19.029s: 8,732 lexical cases, 4,181,796 output bytes.
Native ran with ASan, UBSan and LeakSanitizer enabled; stderr was empty.
The exact comparisons also passed on raw source Node, emitted JavaScript, and
scratch-installed yaml 2.9.0. The external oracle imports only the pinned
library. Go's lexer driver is built through an overlay; cohere is unmodified.

### Mutants

Each mutant compiles, exits zero, and has empty stderr on native and Node.
Only the wrong token bytes catch it:

| Mutation | First differing output byte | Control reached |
| --- | ---: | --- |
| Keep chomping becomes clip | 859077 | `a: |+` with trailing blank lines |
| Tab is removed from whitespace classification | 864210 | Tab after a mapping colon |
| BOM is no longer split as its own token | 858065 | BOM-only input |

### Stage 0 gap programs

All three programs print the recorded behavior on Node and refuse with
`lower.NotYet` before clang. `TestLexerGaps` holds these observations so a
closed gap requires updating the workaround.

| Program | Node stdout | Observed refusal | Port workaround |
| --- | --- | --- | --- |
| [prefixIncrement.ts](gaps/prefixIncrement.ts) | `1` | `a PrefixUnaryExpression on a number` | Increment in a separate statement |
| [assignmentValue.ts](gaps/assignmentValue.ts) | `1` | `a BinaryExpression with a number and a number` | Assign before reading the result |
| [stringPresence.ts](gaps/stringPresence.ts) | `false` | `a PrefixUnaryExpression on a string` | Compare the character sentinel with the empty string |

Labels and logical assignment operators are explicit 0.1 refusals, not newly
observed language gaps. The port uses ordinary loop control and assignments.
The npm library's ISC license accompanies its adapted lexer in `LICENSE-yaml`.

### Lexer throughput

Five interleaved fresh-process rounds, after tests and builds had finished.
Each reads the same 1,189,637-byte input and writes 4,181,796 token bytes.
Each measured answer is compared to Go; exit zero and empty stderr are required.
Startup, file input, chunk handling and output are included; compilation is excluded.
Chunked cases count as texts. This does not measure the unfinished formatter.

| Driver | Median texts/s |
| --- | ---: |
| Native release | 8184.87 |
| Port source on Node | 14155.37 |
| Go cohere | 40714.98 |
| Original yaml 2.9.0 on Node | 11382.30 |

The native driver is 4.97 times slower than Go and 1.73 times slower than Node
on these inputs. This includes different output implementations: buffered Go
`fmt.Fprintf` versus token-by-token `console.log` on native and Node.

### Commands and evidence

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_YAML_LIBRARY=/tmp/stage1-yaml-library ADAMIC_YAML_ARTIFACTS=/tmp/stage1-yaml-artifacts go test -v -count=1 -timeout 30m ./stage1/cohere/yaml > /tmp/stage1-yaml-lexer-final.log 2>&1
go run ./cmd/adamic build stage1/cohere/yaml/lex_main.ts -o /tmp/stage1-yaml-artifacts/native-lexer > /tmp/stage1-yaml-lexer-build.log 2>&1
python3 stage1/cohere/yaml/benchmark_lexer.py /tmp/stage1-yaml-artifacts /tmp/stage1-yaml-library > /tmp/stage1-yaml-lexer-timing.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$' > /tmp/stage1-yaml-filtered-oracle.log 2>&1
/tmp/stage1-yaml-cohere --no-fix --format-only --format-all stage1/cohere/yaml/lexer.ts stage1/cohere/yaml/lex_main.ts > /tmp/stage1-yaml-format-check.log 2>&1
/tmp/stage1-yaml-cohere --no-fix --no-format stage1/cohere/yaml/lexer.ts stage1/cohere/yaml/lex_main.ts > /tmp/stage1-yaml-lint.log 2>&1
```

The filtered oracle passed in 4.097s. Cohere reported 276 rules, two files
checked, 100% Adamic-ready. No compiler or runtime implementation changed.
Logs: [suite](audit/lexer-suite.log), [throughput](audit/lexer-timing.log),
[filtered oracle](audit/lexer-oracle.log), [lint](audit/lexer-lint.log).
The complete repository gate was not run. Formatter parity and formatter
throughput remain pending; this checkpoint must not be treated as a finished YAML slice.

# Baseline audit before the lexer port

No Adamic YAML parser, printer, or stdout driver was built.
This directory records baseline evidence, not a completed stage 1 slice.
Go matches its formatter oracle on all 36 YAML files in this checkout.
Pinned JavaScript parser comparisons passed fixtures and generated cases.
Native parity, performance measurements, and three port mutants remain unrun.

## Revisions and setup

Branch `codex/stage1-yaml` starts at main
`1a4e29984b0b1727b063e76fe01c2aa11dfa56c7`.
Cohere is pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`.
The checkout's 36 YAML files contain 200,885 source bytes: 35 under
TypeScript and one under cohere. The inventory excludes `.git` only.

`bash cloud/setup.sh` exited zero. Its reported timings were Go 0s,
clang 1s, Node 1s, submodules 1s, build cache 78s, total 78s.
`nproc` reported 5; CPU quota was four CPUs (`cpu.max: 400000 100000`).
Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0.
The environment file is `/workspace/adamic-tools/env.sh`.
The complete setup output is [audit/setup.log](audit/setup.log).

## External baseline

Scratch npm dependencies were installed with exact versions:
Prettier 3.9.6, yaml 2.9.0, yaml-unist-parser 3.2.0.
The formatter tests use cohere's vendored fork bundles, not the npm
Prettier package, as their printer oracle. Lexer, composer, and unist
comparisons use the pinned npm packages under Node.
No direct npm Prettier formatting comparison is claimed.

The first Go test run passed but skipped external comparisons because
no Prettier fork path existed. After npm installation, a second run
failed: setting `COHERE_PRETTIER_FORK` also overrides the formatter's
bundle path, and npm does not provide `dist/prettier/standalone.js`.
A scratch symlink to cohere's vendored bundles supplied that path.
The corrected run passed all four packages:

| Package | Seconds | Comparisons |
| --- | ---: | --- |
| yaml | 11.202 | Formatting fixtures; 36/36 YAML files, full stack and tree loader |
| compose | 5.278 | 735 fixtures, 30,000 generated texts, 36 corpus files |
| cst | 2.191 | 586 fixtures, 2,930 chunked parses, 20,000 generated texts, 36 corpus files |
| unist | 5.110 | 444 fixtures, 20,000 generated texts, 42 corpus inputs |

The extra six corpus inputs are Markdown front matter selected by
cohere's existing tests. Their Go comparisons passed; this does not
constitute an Adamic front matter implementation.

`TestFormatOracleCanFail`, the three parser-layer `TestOracleCanFail`
tests, and `TestComparisonSeesEveryField` passed. These are existing
Go oracle checks. They are not the requested three Adamic port mutants.

YAML test-suite comparisons and the broader prose-wrap oracle test
skipped because the upstream test-suite fixture directory was absent.
All test output was redirected to log files. The successful external
run is [audit/go-upstream.log](audit/go-upstream.log).

## Reproduction

Run from the repository root, substituting its absolute path for
`/workspace/adamic` if necessary:

```sh
bash cloud/setup.sh > /tmp/stage1-yaml-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/stage1-yaml-library --save-exact prettier@3.9.6 yaml@2.9.0 yaml-unist-parser@3.2.0 > /tmp/stage1-yaml-npm.log 2>&1
mkdir -p /tmp/stage1-yaml-library/dist
ln -s /workspace/adamic/cohere/internal/format/prettier/bundles /tmp/stage1-yaml-library/dist/prettier
cd cohere
COHERE_PRETTIER_FORK=/tmp/stage1-yaml-library COHERE_YAML_CORPUS=/workspace/adamic go test -v -count=1 -timeout 30m ./internal/format/yaml/... > /tmp/stage1-yaml-external-fixed.log 2>&1
```

## Unfinished work

The formatter depends on the CST lexer/parser, document composer,
unist transformation and comment attachment, then the YAML printer and
document layout engine. The survey covered representative source from
these layers; it did not finish reading every implementation file.
The GraphQL reference was read. The JSON reference reading remains
incomplete, including portions of its generated tables and auxiliary
performance tooling. No implementation was edited before those reads.

No language blocker is established by this audit. There are no new
Adamic gap claims or proving programs. No file-formatting driver,
Adamic corpus comparison, native sanitizers, Go/Node/native texts-per-second
measurements, three port mutants, filtered Adamic oracle or full repository
gate was run. This audit must not be treated as a merge-ready YAML port.
