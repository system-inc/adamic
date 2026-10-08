# TestCSSNumbers parallel units

Branch: `stage1-split/css-family`, base `origin/main` at
`54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`.

The unchanged full test passed in 146.90 s with Prettier 3.9.6 enabled.
A timing-only overlay of the original test passed in 128.38 s: 16.98 s
building products, 111.40 s test logic, including 80.60 s in the raw-file loop.
The first cold outer Go test compilation took approximately 267 s separately.
An initial process-traced diagnostic failed because LeakSanitizer rejects
ptrace; all reported successful measurements run without tracing.

The complete final package passed in 77.954 s: 360 pass, 0 fail, 0 skip test
events (357 execution units plus the parent and two census/failure proofs).
Longest execution unit: `mutant-quoted-numbers`, 18.86 s. Every unit has an
explicit 30 s failure threshold. Gofmt and `go vet` passed.

Machine: `nproc=5`, cgroup CPU quota 4, `GOMAXPROCS=4`, default Go test
parallelism 4. Clean baseline start load (1/5/15 min): 1.83/3.32/1.89;
final package start: 3.22/2.47/1.88. Unit measurements include subprocess
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
| Go oracle | 1.283 |
| port lowering and backends | 0.071 |
| sanitized | 8.844 |
| native-fast | 5.159 |
| quoted-numbers lowering and backends | 0.068 |
| quoted-numbers-sanitized | 0.546 |
| unknown-unit lowering and backends | 0.074 |
| unknown-unit-sanitized | 0.529 |
| word-prefix lowering and backends | 0.063 |
| word-prefix-sanitized | 0.460 |

## Census and failure evidence

The independent unsplit enumeration and executable shard ranges contain exactly
90,294 batch preferences + 522 raw preferences + 3 full-corpus mutant checks +
4 full-corpus throughput checks = 90,823 unique case IDs. Missing, repeated,
unexpected, and range/ID mismatches fail. Gate selections are checked to form
the same union for 1, 2, 7, and 358 boxes.

`ADAMIC_TEST_SHARD=i/n` uses zero-based i and selects stable unit ordinals modulo
n; unset runs all. The test comment documents it. Unit categories: 89 batch
ranges of at most 512 texts, 261 individual raw files, 3 mutants, 4 throughput
sides. All original side comparisons, ASan/UBSan, leak checks, preferences,
mutants, and three throughput repetitions remain.

The subprocess proof plants a disagreement in `batch-002049-s` and requires
exactly `batch-002048-002559` to fail. Existing mutants are caught on both
native and Node: quoted-numbers at preference case 152 in
`mutant-quoted-numbers`, unknown-unit at case 68 in `mutant-unknown-unit`,
word-prefix at case 4 in `mutant-word-prefix`.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 ADAMIC_CSSNUMBERS_LIBRARY=/tmp/adamic-css-family-library \
  go test -json -count=1 -timeout 3h ./stage1/cohere/cssnumbers
GOMAXPROCS=4 go vet ./stage1/cohere/cssnumbers
```

Install `prettier@3.9.6` in the external library directory. Initialize the pinned
cohere and TypeScript submodules. `split-results.json` records every unit wall
and each product time for this run. Only this package's tracked files changed.
