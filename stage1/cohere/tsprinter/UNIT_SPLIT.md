# Expression test units

Branch: `stage1-split/printers`. Baseline: `origin/main` at
`54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8`, fetched for this session.
Only `stage1/cohere/tsprinter` changes. This completes the first requested test;
the other six remain for subsequent sessions.

| Test | Untouched before wall | After longest unit | Units |
| --- | ---: | ---: | ---: |
| TestExpressionsAgainstGoAndPrettier | 178.10 s | 9.68 s (`unit-009`) | 65 |

The before command used `go test -json -run
'^TestExpressionsAgainstGoAndPrettier$' -count=1 -timeout 3h` with every required
input enabled. Command wall was 184.528 s. The independent Go oracle enumerated
170,221 expressions from 925 files and retained the existing one full-file parse
refusal. No corpus sampling or new skips were introduced.

A timing-only replay of the original unsplit source passed in 155.89 s. Its Go
command took 5.937 s, including 5.04 s in corpus logic and approximately 0.897 s
in build/launch overhead. Lowering took 2.161 s; sanitized compilation including
C emission took 33.242 s; release compilation including another C emission took
24.894 s. These build-related steps total approximately 61.194 s; the remaining
94.696 s is logic and harness work. Node/native/backend/leak logic took
3.741/9.684/6.657/10.068 s; npm/embedded Prettier comparisons took
29.283/27.433 s. The exact replay sources and logs are preserved alongside the
initial untouched measurement; these are different runs, not an attribution
retroactively assigned to the first measurement.

## Build products

`internal/buildcache` is absent on this base. Each build callback writes its
products into its supplied directory, with Name, Files, Flags and Toolchain
inputs beside it. The fallback invokes each build once per test invocation and
shares its products across all parallel children. It creates no package cache.
Transitive inputs include the imported TypeScript parser and module configuration.
C and backend JavaScript are emitted once; both native builds reuse that C.

| Fresh product step in the complete-package measurement | Cold wall |
| --- | ---: |
| Go expression oracle binary | 1.281 s |
| Lowered expression program, C and backend JavaScript | 14.218 s |
| Sanitized native binary | 21.162 s |
| Release native binary | 10.088 s |

These steps prepare fresh output products; Go dependency archives and existing
native runtime archives remain hash-keyed inputs. An empty toolchain archive
cache was not forced. No external checker archive is built by this test.
After subtracting these product steps, the parent's own logic and harness time
was 5.961 s. Parallel child elapsed times are reported separately by Go.

## Partition and checks

There are 64 deterministic expression units, assigned by descending protocol
bytes, with original enumeration position and unit number breaking ties. This
count is fixed independently of CPU count. Every expression unit runs source
Node, sanitized native, backend JavaScript, leak checks, release native, npm
Prettier and embedded Prettier. The separate gap unit retains all original loud
NotYet and Go/Prettier gap proofs. macOS retains `leaks --atExit` on the release
binary; measurements here are Linux only.

`ADAMIC_TEST_SHARD=i/n` selects zero-based units whose index modulo n equals i.
Unset runs all 65. Tests prove exactly one assignment across 1, 2, 7, 65 and 100
boxes and reject malformed selectors. Each unit checks its own measured time and
fails above 30 s, starting after `t.Parallel()` resumes.

The union check requires exactly the unsplit enumeration's case-ID set, rejecting
missing, repeated and out-of-range IDs, and reconstructs the complete original
answer bytes. Position forms part of each ID because generated cases can share
labels. It verifies 170,221 expression IDs and 15 gap IDs: **170,236 unique IDs**
across 65 units. The parent also requires every pinned upstream record to occur;
each unit compares all applicable exact upstream texts and parser errors.

`TestExpressionUnitPlantedDisagreement` runs real Node children over a small
fixture, planting one wrong answer at `case-2`. Exactly **unit-000** rejects it
through the same comparison used in production units. Missing/repeated union
controls and distributed-selection controls pass as well.

## Validation

The complete package ran with TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8` and Prettier 3.9.6, using
`go test -json -count=1 -parallel=4 -timeout 3h ./stage1/cohere/tsprinter`.
It passed **119 tests/subtests (18 top-level tests, 101 subtests), 0 failures, 0 skips**.
All 29 existing mutants were caught. Package wall was 1,002.476 s; command wall
was 1,007.807 s, including the unchanged tests and their builds.

`nproc=5`; `cpu.max=400000 100000` gives a 4-CPU quota. Whole-package load averages
before/after were `4.34 4.46 3.00` / `5.45 7.50 5.83`. Tools were Go 1.27.1,
clang 20.1.8 and Node 24.19.0. All 65 expression units passed; their longest was
9.68 s. The earlier focused split run also passed all 65 units, longest 9.13 s.

`go vet`, package-wide `gofmt -l`, and `git diff --check` are clean. The final
three control tests were rerun after input-description, diagnostic and exported
file-mode cleanup and passed. These cleanup edits preserve the measured unit
logic and build callbacks. A final actual invocation with
`ADAMIC_TEST_SHARD=64/65` passed, running exactly `unit-064-gaps` (0.36 s),
while checking the full 170,236-ID union and executing the expanded input
collection. Artifact export also succeeded: C mode 600 and executable mode 700
under the inherited restrictive umask. The selector log and fresh product times
are preserved in the evidence directory and timing JSON.

Raw JSON logs, exact timing-only overlays, time/load records and the per-unit
census are in [results/unit-split-expressions](results/unit-split-expressions/),
including [timing.json](results/unit-split-expressions/timing.json).

## Remaining work, in requested order

1. `graphql/printer TestPrinterAsGoCohere`
2. `tsprinter TestTSCCorpusAgreement`
3. `graphql/printer TestPrinterUpstreamPreflight`
4. `tsprinter TestStatementsAgainstGoAndPrettier`
5. `graphql/printer TestPrinterWhitespaceGap`
6. `graphql/printer TestPrinterMutants`

These need their own focused baselines and splits. `graphql/printer` was not
modified or run in this session. The TSC and statement tests passed as part of
the complete tsprinter package; their splits remain unfinished.
