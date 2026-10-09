# Stage 1 selector report

Branch: `codex/stage1-css-selector`. Base: origin/codex/regex-matcher,
`59d82c5de172e4e53d1224553b319d076f22424a`. Cohere checkout:
`715ba94f3608a6500086b1076ce5cb7e51b836db`; TypeScript submodule:
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Changes are confined to this directory.
No compiler changes, submodule commits or pull request.

The tokenizer, recursive parser and lossless AST shape are ported. Capturing
regexp splits use native Adamic regexp support. Flat node ownership avoids a
strong parent/child cycle. Error text and every canonical tree property are
compared against external oracles. See GAPS.md for six compiler refusals with
proving programs and the upstream nontermination evidence.

## Setup

`bash cloud/setup.sh > /tmp/selector-setup.log 2>&1` succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (69s)
setup: done in 69s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc`: 5. `source /workspace/adamic-tools/env.sh` supplies Go 1.27.1,
Node 24.19.0 and clang 20.1.8. The regex branch needed an explicit ref fetch
because the configured fetch refspec only fetched main.

## Reproduce

Install the external oracle in scratch, not the repository:

```
mkdir -p /tmp/selector-library
npm install --prefix /tmp/selector-library postcss-selector-parser@2.2.3 postcss@8.5.16 > /tmp/selector-npm.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SELECTOR_LIBRARY=/tmp/selector-library ADAMIC_SELECTOR_KEEP=/tmp/selector-cases.txt go test -v -count=1 -timeout 30m ./stage1/cohere/selector > /tmp/selector-test.log 2>&1
ADAMIC_SELECTOR_LIBRARY=/tmp/selector-library ADAMIC_SELECTOR_BENCH=1 go test -v -count=1 -run '^TestSelectorThroughput$' -timeout 30m ./stage1/cohere/selector > /tmp/selector-complete-bench.log 2>&1
ADAMIC_SELECTOR_LIBRARY=/tmp/selector-library go test -count=1 -timeout 30m ./... > /tmp/selector-gate.log 2>&1
```

`COHERE_SELECTOR_SEED` and `COHERE_SELECTOR_GENERATED` override generation.
The library version is checked exactly. Without its environment variable the
original JS comparison and CSS-file/snippet collection skip explicitly; the Go
and Adamic comparisons still run. The library hang proof also requires it.

From the cohere submodule, check only this slice:

```
go run ./command/cohere --directory /workspace/adamic --no-fix stage1/cohere/selector > /tmp/selector-lint.log 2>&1
```

Observed output: `276 rules`, `11 checked`, `100% Adamic-ready (5 of 5)`, success.
There is one documented lint exception for the parser taking ownership of fresh
node arguments, as the original implementation does.

## Validation observations

The final focused command was:

```
ADAMIC_SELECTOR_LIBRARY=/tmp/selector-library go test -v -count=1 -timeout 30m ./stage1/cohere/selector > /tmp/selector-complete-test.log 2>&1
```

It exited zero: 27,160 answers agree on source Node, sanitized native and the
JavaScript backend; upstream JS agrees on 26,759 terminating cases. All three
mutants were caught on native and Node, all six gap programs passed, all four
upstream timeout programs passed, and the baseline leak check passed.
`go vet ./...` exited zero (`/tmp/selector-vet.log`). Repository `gofmt -l`
produced no filenames (`/tmp/selector-gofmt.log`).

The full gate was launched during final lint and harness fixes. Its inherited
`internal/fresh/TestEveryWriteIsRecordedAndKnown` fails because the cycle finder
does not recognize regexp IR nodes (`RegExpNew`, `RegExpCall`, `RegExpProperty`
and `IsNull`). The fresh check's globs only include existing compiler fixtures,
not this slice, and those compiler files are unchanged from the regex base.
An independent focused fresh rerun is logged in `/tmp/selector-fresh-baseline.log`.
The full invocation exited 1. Its compiler oracle passed in 1295.077s, native
package passed in 718.380s, and every stage-1 package passed, including selector
in 110.206s (the earlier corpus). The Unicode properties package also failed:
`TestCanonicalizeUnicodeNode` reported `node: signal: killed` after 403.90s.
Observation: its Node child was killed. The test enforces a four-minute deadline
per Node batch, but the failure output does not establish whether a deadline or
another cause killed it. No Unicode properties files were changed. The full
gate is not claimed green. The independent fresh rerun exited 1 in 51.380s.
The final expanded-corpus selector run exited 0 in 81.995s.

Before porting, `go test -count=1 ./internal/format/css/selector` inside cohere
also passed in 0.099s (`/tmp/selector-go-baseline.log`).

## Mutants

Each runs native and Node, must exit zero, and must produce a different answer.

1. Remove `|` from the attribute operator regexp. `[a|=b]` becomes namespace `a`
   with operator `=` instead of attribute `a` with operator `|=`. Canonical AST
   comparison catches it on both paths.
2. Ignore quote escape parity. Error text changes from `Unexpected "string"
found.` to `Unclosed quote`. Byte comparison catches it on both paths.
3. Use the first token for merged-word location. Source end positions change.
   Canonical AST comparison catches it on both paths.

## Throughput

Measured after the full gate finished, with source-index conversion inside
Parse and only root-child counting, over ten passes of all 27,160 inputs:

| Implementation                                      |                 Selectors/s |
| --------------------------------------------------- | --------------------------: |
| Adamic native, three rounds                         | 232,528 / 232,606 / 235,700 |
| Adamic source on Node, three rounds                 | 185,097 / 184,607 / 175,830 |
| Original postcss-selector-parser Node, three rounds | 105,210 / 105,799 / 113,561 |
| Go cohere, one parser loop                          |                      85,673 |

All paths counted 99,100 successful parses and 105,840 root children over
271,600 attempts. Upstream known hangs are refused without invocation. Native
and Node wall times include startup, file reading and input decoding; Go times
only its parse loop. No sanitizer in throughput builds. Original JS does not
convert source indexes to bytes in this benchmark, so its work is somewhat
smaller. These are corpus measurements under the cloud CPU quota, not a general
speed claim. Raw observations are in `/tmp/selector-complete-bench.log`.
Benchmark command exited zero in 28.537s.

## Coverage limits

See GAPS.md. In particular: upstream hangs cannot be byte-compared, TypeScript
has no CSS files in this checkout, nondefault parser options and public mutation
or printer APIs are not ported, and Unicode comparisons use Go's normalization.
