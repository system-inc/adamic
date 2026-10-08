# TestCSSStrings parallel units

Branch: `stage1-split/css-family`, base `origin/main` at
`54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`. The preceding commit splits
TestCSSNumbers; this commit changes only the cssstrings package.

The unchanged full test passed in 95.72 s with Prettier 3.9.6 enabled.
A timing-only overlay of the original test passed in 106.71 s: 17.54 s
building products, 89.17 s test logic, including 83.21 s in the raw-file loop.

The complete final package passed in 50.099 s: 291 pass, 0 fail, 0 skip test
events (287 execution units, the parent, two census/failure proofs, and the
existing TestMultiPushGap). Longest execution unit: `mutant-preserve-original-escape`,
1.67 s. Every execution unit has an explicit 30 s failure threshold.
Gofmt and `go vet` passed.

Machine: `nproc=5`, cgroup CPU quota 4, `GOMAXPROCS=4`, default Go test
parallelism 4. Clean baseline start load (1/5/15 min): 0.75/2.16/2.02;
final package start: 0.83/1.56/1.81. Unit measurements include subprocess
startup, temp input creation, sanitizer/leak runs, and comparisons.

## Products

There is no `internal/buildcache` on the base. Builders have adjacent inputs
and write products into a supplied directory. Each is called once in the parent
and shared across units; there is no package cache. Fresh product directories
and a fresh native runtime cache were used for the final measurement. Existing
Go dependency build products were retained. Sanitized and native-fast cold
build times include their respective runtime archive builds; mutants reuse those
existing native runtime archives by the native helper's content hash.

| Product | Cold seconds |
| --- | ---: |
| Go oracle | 1.195 |
| port lowering and backends | 0.094 |
| sanitized | 9.303 |
| native-fast | 5.002 |
| quote-tie lowering and backends | 0.103 |
| quote-tie-sanitized | 0.338 |
| escaped-closing-quote lowering and backends | 0.048 |
| escaped-closing-quote-sanitized | 0.338 |
| preserve-original-escape lowering and backends | 0.054 |
| preserve-original-escape-sanitized | 0.305 |

## Census and failure evidence

The independent unsplit enumeration and executable shard ranges contain exactly
19,182 batch preferences + 522 raw preferences + 3 full-corpus mutant checks +
4 full-corpus throughput checks = 19,711 unique case IDs. Missing, repeated,
unexpected, and range/ID mismatches fail. Gate selections are checked to form
the same union for 1, 2, 7, and 288 boxes.

`ADAMIC_TEST_SHARD=i/n` uses zero-based i and selects stable unit ordinals modulo
n; unset runs all. The test comment documents it. Unit categories: 19 batch
ranges of at most 512 texts, 261 individual raw files, 3 mutants, 4 throughput
sides. All original side comparisons, ASan/UBSan, leak checks, preferences,
mutants, and three throughput repetitions remain.

The subprocess proof plants a disagreement in `batch-002049-s` and requires
exactly `batch-002048-002559` to fail. A real `ADAMIC_TEST_SHARD=4/287` run
executed only that batch unit and passed in 0.41 s (3.603 s package wall).
Existing mutants are caught on both native and Node: quote-tie at preference
case 4 in `mutant-quote-tie`, escaped-closing-quote at case 72 in
`mutant-escaped-closing-quote`, preserve-original-escape at case 196 in
`mutant-preserve-original-escape`.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 ADAMIC_CSSSTRINGS_LIBRARY=/tmp/adamic-css-family-library \
  go test -json -count=1 -timeout 3h ./stage1/cohere/cssstrings
GOMAXPROCS=4 go vet ./stage1/cohere/cssstrings
```

Install `prettier@3.9.6` in the external library directory. Initialize the pinned
cohere and TypeScript submodules. `split-results.json` records every unit wall
and each product time for this run. Nine requested tests remain unmodified:
css parsing, composition, composed memory checks, both printer comparisons,
optimized parser comparison, canonical range checks; selector parsing;
gitignore answers. They need a follow-up session, beginning with
`css.TestThePortParsesAsGoCohereDoes`.
