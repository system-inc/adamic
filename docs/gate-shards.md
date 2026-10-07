# Gate shards

Build `go build -o /workspace/adamic-gate ./cmd/adamic-gate` after `bash cloud/setup.sh`
and source the environment file setup prints. The mandatory Markdown width oracle also needs
its three pinned dependencies, which cloud/setup.sh does not install:

```sh
npm install --prefix /workspace/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
export ADAMIC_MARKDOWNWIDTH_DEPS=/workspace/adamic-markdown-width
```

Provision that directory and variable on the fleet before executing shard commands. From the
same clean commit on every worker:

```sh
adamic-gate plan -count 8 > /workspace/plan.json
adamic-gate shard -index 0 -count 8 -out /workspace/shard-0
# Other boxes use indices 1 through 7.
adamic-gate merge -out /workspace/merged /workspace/shard-{0,1,2,3,4,5,6,7}
adamic-gate compare /workspace/merged /workspace/whole.jsonl
```

The external launcher supplies the boxes. Indices are zero based. Each output directory must
be new. Merge works from the repository checkout used by the workers, re-enumerates it, and
writes `merged.json` and the concatenated raw `test.jsonl`. A red merge retains its evidence
and exits 1. Compare exits 1 for a different test name, verdict, or distinct terminal-event
count, or for a red merge. Test identity includes both package and the entire Go test name.

## Selection and coverage

Discovery uses `go list ./...` and `go test -json -list . ./...`, including examples and fuzz
seed tests. Packages with no tests are build-only units. Ordinary units are top-level tests;
the following audited parents have independently selectable, enumerable children:

- `internal/oracle`: `TestNativeAgreesWithNode`, `TestInputAgreesWithNode`, and
  `TestFreshWriteProbesStayRefused`.
- `internal/native`: `TestNormalizeMatchesNode` and `TestStringIndexMatchesNode`.
- `stage1/cohere/lint`: `TestMutants` and `TestVolumeMutants`.
- `stage1/cohere/css`: the two modes of `TestCSSPrinterAgreesWithGo`, including each mode's nested mutants.

The first two fixture lists come from literal rows in the checked-in test AST, with every
named input checked on disk. Globbing all oracle testdata would incorrectly include helpers
and probes. Fresh probes come from the same glob the test uses. Native sweep labels come
from the literal slice iterated by the audited parent. Unsupported enumeration fails loudly.
The selector escapes regex metacharacters and anchors each slash-separated component.
Only siblings with an identical literal prefix share one pattern, avoiding the Cartesian
product that Go creates when alternatives at different slash levels are combined.

`cmd/adamic-gate/timings.json` contains terminal elapsed seconds keyed by `package::test`.
Generate it with `adamic-gate timings -out cmd/adamic-gate/timings.json whole.jsonl` and commit
it. Known units are assigned by greedy longest-first packing, ties by name and shard index.
Unknown names use SHA-256 modulo shard count. There is no result or discovery cache. A plan
contains its full universe, commit, shard count, and a digest of tracked bytes, symlink targets,
and submodule revisions. Identical inputs give identical plans without consulting machine
speed, load, wall clock, map iteration order, or previous local test runs.

Merge uses raw logs for verdicts and cache lines. It rejects missing shard indices, inconsistent
plans or source identities, failing commands or shard-0 checks, unplanned tests, duplicate work,
missing terminal evidence for any planned unit, and any nonzero or malformed gate-cache hit
line. It also requires each recorded invocation to export `ADAMIC_GATE_UNCACHED=1`; a cold
cache with no hits does not excuse a cached invocation. Skip records contain their diagnostic
reason, and failure records contain their output. The complete output is always in the log.

Go necessarily runs a selected subtest's parent on each worker. These declared ancestor
bookkeeping events repeat and are combined once, with failure taking precedence, then pass,
then skip. All other repeated tests fail merge, including repeated descendants of an unsplit
top-level unit. Compare counts one terminal verdict per package/test name, including parents,
and requires the unsharded log to have exactly one terminal event per test. The merged report
also exposes the raw terminal count. Literal equality of raw counts and a prohibition on all
repeated parent executions are impossible with Go subtest sharding without changing the tests:
Go emits those parents automatically. This is an explicit limitation, not discarded evidence.

## Scratch and measurement

Shard runs use `go test -count=1 -json -timeout 60m` with the uncached variable explicitly set.
Shard 0 additionally checks `gofmt -l cmd internal` and runs `go vet ./...`. Both write logs;
a nonempty formatting list fails the gate. The test timeout follows this unit's explicit 60m
requirement. The ordinary whole-gate command in CLAUDE.md still specifies 30m.

Scratch defaults to `/workspace/adamic-gate-scratch`; `-scratch` can name another disk-backed
root. The root is world traversable, each package gets its own directory, and cleanup happens
after that package's selected invocations finish. The runner rejects tmpfs and scratch inside
the repository, including symlinks into it: corpus tests walk the repository. It reports free
bytes before and after the shard and after each package. On this worker `/tmp` is tmpfs, so
placing scratch there would consume memory and risk the failures seen in whole gates.

Summary wall time includes discovery and shard-0 checks. Build-flags lines include commit,
`nproc`, CPU quota, Go, clang and Node versions, load before and after, uncached mode, and
`GOFLAGS`, `CGO_ENABLED`, `GOMAXPROCS`, and the width-oracle dependency path. Weights collected during a parallel whole gate are
elapsed times under contention, not independent CPU costs or predictions of an eight-box run.
Sequential shard measurements give an observed largest-shard floor on this worker; the fleet
must measure again on its own hardware.

## Parents kept whole

`TestCountsAreRecorded` must see all fixture rows: its parent compares the complete reconstructed
table with counts.md. Selecting some children makes the final check fail even when every selected
row is correct. `TestEveryMutationIsInItsRange` checks an aggregate nonzero mutation count;
individual fixtures without mutations would fail the parent. These remain top-level units.
`TestBridge` performs a complete bridge integration sequence without selectable subtests, and
`TestPortMatchesGoCohere` in the JSON port compares and measures whole batches without subtests.
Their inner builds and corpus loops cannot be selected by `-run`. The exhaustive
`TestCanonicalizeUnicodeNode` sweep likewise has no selectable inner subtests.

Markdown block layout, preprocessing, and inline tests build corpora and compare complete batches
outside their mutant subtests. Selecting a mutant still repeats the heavy parent work, and selecting
only its children cannot distribute that work. In particular `TestMarkdownInline`,
`TestWholeDocumentOraclePreflight`, and the block layout tests cannot partition their batch comparisons
with `-run`. Getting every shard below five minutes requires every indivisible parent to finish below
five minutes, enough aggregate four-core capacity, and time for compilation, discovery and vet.
Batch partitioning needs a separate test-harness change, which this unit does not make.

## Proof records

The fixed revision, full-gate and sequential shard totals, times, and mutant outcomes are recorded
here after execution. Raw evidence is kept outside the repository so it does not enter corpus tests.
