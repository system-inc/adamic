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
# After an interruption, repeat the same command with -resume.
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
- `stage1/cohere/typeaware`: the 15 literal `changes` rows of `TestVolumeAgreementAndMutants`.
- `internal/native`: `TestNormalizeMatchesNode` and `TestStringIndexMatchesNode`.
- `cmd/adamic-test262`: `TestLargeCompilerOutputIsComplete` (`large.js`) and
  `TestParallelCachedMatchesSerial` (`cold`, `warm`, `limit`). Compiler setup is shared;
  each cache comparison child owns fresh work and observations, and `warm` primes itself.
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
Unknown names use SHA-256 modulo shard count. Predictions sum known elapsed weights and
add positive parent elapsed time minus its enumerated child spans once per selected parent
invocation. This accounts for repeated setup, including 44.98 seconds for the type-aware volume
parent. `Unknown` counts weights missing from the file. Discovery, vet, compilation overhead,
and unknown durations are not predicted. There is no automatic result or discovery cache. A plan
contains its full universe, commit, shard count, predicted totals and largest units in `Shards`,
a `Digest` of the complete plan, and a digest of tracked bytes, symlink targets,
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
with `-run`. Getting every shard below twelve minutes requires every indivisible parent to finish below
twelve minutes, enough aggregate four-core capacity, and time for compilation, discovery and vet.
Batch partitioning needs a separate test-harness change, which this unit does not make.

## Resume

`-resume` is explicit reuse of completed raw evidence, while every newly executed package still
exports `ADAMIC_GATE_UNCACHED=1`. Without `-resume`, use a fresh output directory to execute
all selected tests again. `run.json` binds the output to the plan digest, shard index and execution
context. Context hashes the environment, this tool and Go/clang/Node binaries, effective Go
settings and compiler tools, clang resources, and existing external paths named by `ADAMIC_*`
variables. External input bytes, permissions, ownership and symlink targets are included.
Private environment values are hashed rather than written into evidence. Go's automatically
random `go-buildNNN` remapping in `GOGCCFLAGS` is normalized; actual compiler flags and
`TMPDIR` remain keyed. Keep the same environment and immutable system image when resuming;
changed inputs fail closed and require a fresh output directory.

Each package writes `packages/<package-hash>/test.jsonl` and `test.stderr`, then atomically
publishes `complete.json` after syncing its logs. Resume verifies the key, checksums, exact
invocations and all planned terminal events before copying those bytes into the aggregate.
An interrupted package without `complete.json` runs again. Corrupt complete evidence is refused.
A complete failed package stays failed on resume. An output lock prevents two live writers;
recorded package scratch is removed safely on restart. Shard 0 reruns formatting and vet.
Summary wall time covers the current attempt, while retained invocation times remain original.

Merge checks package checkpoints and verifies that their concatenation equals the aggregate.
It rejects missing package terminals even if test verdicts exist. A failing verdict prints the
shard index, full package/test name and captured output. It never trusts summary totals to decide
a raw test's verdict.

## Audit limits and validation

The type-aware package's 918.003-second entry, lint's 738.926-second entry and CSS's
678.296-second entry are package totals in the timing file, not planned units. These packages
already distribute their top-level tests. Type-aware volume now also distributes its 15 genuine
Go subtests. The other type-aware heavy loops, including `TestSixRuleAgreementAndMutants`,
`TestTypeAwareAgreementAndMutants`, `TestVolumeProfileAgreementAndMutants` and
`TestInspectRequestRefusals`, do not call `t.Run`; their loop labels cannot be selected by Go.
`TestBridge` likewise has no subtests. No existing test files were changed.

`internal/unicodeproperties/canonicalize_test.go:205`, `TestCanonicalizeUnicodeNode`, has no
subtests and its checked-in span is 742.71 seconds. It alone exceeds the 720-second target.
More workers cannot divide it. Meeting twelve minutes needs a separately authorized test-harness
partition or implementation optimization, then fresh isolated measurements and rebalanced weights.
The fleet launcher and distributed full-gate comparison remain outside this unit.

The full reference and sequential shard proof were stopped at the user's revised instruction.
The stopped reference at `b790232` has 2724 pass, 0 fail, 21 skip and is incomplete. Its compressed
raw log is `cmd/adamic-gate/evidence/stopped-b790232.jsonl.gz` (353594 bytes). Two earlier
interrupted references had 2839/0/30 and 2870/0/30. No full comparison or sequential shard times
are claimed. The initial calibration had 2923/1/30: its width failure was missing pinned npm
inputs; the corrected focused width run passed all four test verdicts. The timing file is a
scheduling seed from that calibration, not a clean performance proof.

Validation uses seven tool tests, full `go vet ./...`, one exactly selected type-aware child
(parent plus `assignable-types`, both pass), and a tiny scratch repository with real Go JSON logs.
Two scratch shards merge GREEN (1 pass, 0 fail, 1 skip) and compare with a separate scratch
unsharded log with an empty diff. Resume and recovery after deleting the aggregate and summary
preserve the raw answer bytes. The real-log failing mutant changes only `TestPass` to fail;
merge says RED and names `shard 0 test gateprobe::TestPass failed`. Package checkpoint digests
are updated in that mutant so its failure is the verdict guard, not a checksum mismatch.
Missing package terminals, an actually widened selector across two shards, false uncached
provenance, and a real cache-hit line also make merge red. An environment change refuses resume.

Source-overlay mutants separately omit plan digest, execution context, shard index, package,
selectors, environment, tools, external inputs, compiler flags, external bytes, external permissions,
raw-log checksum, stderr checksum, package terminal validation and scratch ownership. Each
fails its intended tool test. These add to the original regex-escaping and duplicate-guard mutants.
Small CLI and mutant evidence is archived under `cmd/adamic-gate/evidence/`; complete large logs
remain outside the repository. The evidence manifest identifies each artifact and its totals.

Toolchain setup reported `go ready (0s)`, `clang ready (0s)`, `node ready (0s)`,
`submodules ready (0s)`, `build cache warm (94s)`, and `done in 94s` on `nproc=5`,
with `cpu.max=400000 100000`. Setup load averages were not recorded; its raw step lines are
reported rather than treated as a controlled performance comparison. The clean-commit prediction metadata and tables follow below.

## Predicted plans at 64c9144

These are scheduling predictions from known elapsed weights, not measured shard wall times.
There are 725 units. Each table includes positive repeated parent setup; unknown durations,
compilation, discovery and vet can add time. The eight-shard maximum rises with the volume
subdivision because the 44.98-second parent setup repeats on multiple workers. No speedup
or distributed equivalence is claimed by these predictions.

| Loop | Before | After (predicted maximum) | Instrument (exact command, from `/workspace/adamic`) |
| --- | --- | ---: | --- |
| 8 shards | Full measurement stopped by request | 1356.21 s | `ADAMIC_GATE_UNCACHED=1 ADAMIC_MARKDOWNWIDTH_DEPS=/workspace/adamic-markdown-width TMPDIR=/workspace/gate-evidence/typeaware-select-scratch /workspace/adamic-gate plan -count 8` |
| 10 shards | Full measurement stopped by request | 1097.44 s | `ADAMIC_GATE_UNCACHED=1 ADAMIC_MARKDOWNWIDTH_DEPS=/workspace/adamic-markdown-width TMPDIR=/workspace/gate-evidence/typeaware-select-scratch /workspace/adamic-gate plan -count 10` |
| 12 shards | Full measurement stopped by request | 921.60 s | `ADAMIC_GATE_UNCACHED=1 ADAMIC_MARKDOWNWIDTH_DEPS=/workspace/adamic-markdown-width TMPDIR=/workspace/gate-evidence/typeaware-select-scratch /workspace/adamic-gate plan -count 12` |

Build-flags lines for the three plan calculations, all uncached discovery on the same box:

```text
count=8 commit=64c914495d3ddba632f229400b869348e4eb2909 nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="0.16 0.47 1.74 1/229 193406" load_after="0.13 0.45 1.72 1/228 193925" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps="/workspace/adamic-markdown-width"
count=10 commit=64c914495d3ddba632f229400b869348e4eb2909 nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="0.13 0.45 1.72 1/228 193925" load_after="1.75 0.79 1.81 1/228 194444" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps="/workspace/adamic-markdown-width"
count=12 commit=64c914495d3ddba632f229400b869348e4eb2909 nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="1.75 0.79 1.81 1/228 194444" load_after="2.62 1.01 1.87 1/228 195049" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps="/workspace/adamic-markdown-width"
```

### 8 shards

| Shard | Predicted seconds | Largest unit | Unit seconds |
| ---: | ---: | --- | ---: |
| 0 | 1342.15 | `internal/unicodeproperties::TestCanonicalizeUnicodeNode` | 742.71 |
| 1 | 1356.20 | `bridge/tsgo::TestBridge` | 507.44 |
| 2 | 1345.05 | `stage1/cohere/markdownblocks::TestMarkdownASTPreprocessing` | 473.38 |
| 3 | 1341.73 | `stage1/cohere/json::TestPortMatchesGoCohere` | 432.18 |
| 4 | 1356.21 | `stage1/cohere/markdownblocks::TestMarkdownWhitespaceLayout` | 416.56 |
| 5 | 1345.04 | `stage1/cohere/markdowninline::TestMarkdownInline` | 394.16 |
| 6 | 1345.04 | `stage1/cohere/markdownblocks::TestMarkdownRootLayout` | 367.03 |
| 7 | 1300.06 | `stage1/cohere/markdownblocks::TestMarkdownQuoteLayout` | 346.11 |

### 10 shards

| Shard | Predicted seconds | Largest unit | Unit seconds |
| ---: | ---: | --- | ---: |
| 0 | 1086.28 | `internal/unicodeproperties::TestCanonicalizeUnicodeNode` | 742.71 |
| 1 | 1083.38 | `bridge/tsgo::TestBridge` | 507.44 |
| 2 | 1083.39 | `stage1/cohere/markdownblocks::TestMarkdownASTPreprocessing` | 473.38 |
| 3 | 1086.28 | `stage1/cohere/json::TestPortMatchesGoCohere` | 432.18 |
| 4 | 1094.11 | `stage1/cohere/markdownblocks::TestMarkdownWhitespaceLayout` | 416.56 |
| 5 | 1082.94 | `stage1/cohere/markdowninline::TestMarkdownInline` | 394.16 |
| 6 | 1086.27 | `stage1/cohere/markdownblocks::TestMarkdownRootLayout` | 367.03 |
| 7 | 1041.30 | `stage1/cohere/markdownblocks::TestMarkdownQuoteLayout` | 346.11 |
| 8 | 1097.44 | `stage1/cohere/markdownblocks::TestMarkdownStructureLayout` | 330.65 |
| 9 | 1082.95 | `stage1/cohere/markdownblocks::TestMarkdownListLayout` | 288.38 |

### 12 shards

| Shard | Predicted seconds | Largest unit | Unit seconds |
| ---: | ---: | --- | ---: |
| 0 | 910.86 | `internal/unicodeproperties::TestCanonicalizeUnicodeNode` | 742.71 |
| 1 | 910.86 | `bridge/tsgo::TestBridge` | 507.44 |
| 2 | 921.59 | `stage1/cohere/markdownblocks::TestMarkdownASTPreprocessing` | 473.38 |
| 3 | 910.46 | `stage1/cohere/json::TestPortMatchesGoCohere` | 432.18 |
| 4 | 913.78 | `stage1/cohere/markdownblocks::TestMarkdownWhitespaceLayout` | 416.56 |
| 5 | 910.43 | `stage1/cohere/markdowninline::TestMarkdownInline` | 394.16 |
| 6 | 913.77 | `stage1/cohere/markdownblocks::TestMarkdownRootLayout` | 367.03 |
| 7 | 910.44 | `stage1/cohere/markdownblocks::TestMarkdownQuoteLayout` | 346.11 |
| 8 | 868.78 | `stage1/cohere/markdownblocks::TestMarkdownStructureLayout` | 330.65 |
| 9 | 913.76 | `stage1/cohere/markdownblocks::TestMarkdownListLayout` | 288.38 |
| 10 | 921.60 | `stage1/cohere/typeaware::TestSixRuleAgreementAndMutants` | 271.06 |
| 11 | 910.44 | `stage1/cohere/markdownblocks::TestMarkdownTextSplitting` | 260.02 |

The largest units overall are Unicode canonicalization (742.71 s), bridge integration
(507.44 s), Markdown AST preprocessing (473.38 s), the JSON batch oracle (432.18 s),
and Markdown whitespace layout (416.56 s). The new type-aware largest unit is
`TestSixRuleAgreementAndMutants` (271.06 s); its 24 units include the 15 volume children.
Lint has 38 units with `TestRulesAgree` largest (192.82 s), and CSS has 17 with
`TestComposedMemoryChecksCanFail` largest (130.16 s). Bridge has two top-level units;
`TestBridge` has no enumerable subtests.
