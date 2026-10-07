# Gate shards

## Provisioned whole-gate timings, October 7

The timing file was regenerated with the tool, rather than estimated from package spans:

```sh
adamic-gate timings -out cmd/adamic-gate/timings.json /workspace/plain-2adf65c/gate-out/test.jsonl
```

Input: `gate-logs/2adf65c2514e/plain:plain.tgz`, `gate-out/test.jsonl`, tested commit
`2adf65c2514eefbbb073bfdab922d00a4e80a3d3`, 4,114 pass, 0 fail, 21 skip,
4,135 distinct terminal tests. The gate duration recorded by that worker is
3,471.920 seconds. Build flags: `nproc=5`, `cpu.max="400000 100000"`,
`go="go version go1.27.1 linux/amd64"`, `clang="clang version 20.1.8 (87f0227)"`,
`node="v24.19.0"`, uncached gate, setup `--gate-inputs`. Gate load before/after
was not recorded; setup recorded 0.25/6.51, which is a different interval.
These fields apply to the source observations behind every prediction below.

| Shards | Maximum predicted seconds | Maximum predicted minutes |
| ---: | ---: | ---: |
| 8 | 2,034.970 | 33.92 |
| 10 | 1,656.950 | 27.62 |
| 12 | 1,404.950 | 23.42 |
| 13 | 1,308.010 | 21.80 |
| 14 | 1,224.940 | 20.42 |
| 15 | 1,152.940 | 19.22 |
| 16 | 1,089.910 | 18.17 |

Fifteen is the smallest passing count; sixteen has more margin. Counts up to twelve
cannot fit the 15,120.390 seconds of known unit work into twenty minutes per shard,
even before parent setup; thirteen and fourteen also fail in the actual plans. Instrument:
`adamic-gate plan -count N`, using the regenerated timings and the current source.
These are sums of recorded elapsed work plus repeated parent setup, not measured
fleet wall times. Missing timings, compilation, discovery, vet and contention
are not included. Ten new integration and gate-tool units have unknown weights;
no production shards or new 58-minute plain reference were run on this box.
The full per-shard predictions, including each shard's largest unit, are in
`cmd/adamic-gate/evidence/replan-provisioned.json`.

| Largest single unit | Seconds | Minutes |
| --- | ---: | ---: |
| `internal/unicodeproperties::TestCanonicalizeUnicodeNode` | 822.510 | 13.71 |
| `stage1/cohere/json::TestPortMatchesGoCohere` | 762.200 | 12.70 |
| `stage1/cohere/markdownblocks::TestMarkdownWhitespaceLayout` | 762.150 | 12.70 |
| `stage1/cohere/markdownblocks::TestMarkdownASTPreprocessing` | 729.800 | 12.16 |
| `stage1/cohere/markdownblocks::TestMarkdownQuoteLayout` | 626.800 | 10.45 |

No planned unit exceeds 15 minutes. The 2,762.341-second Markdown package terminal
is an aggregate span, not one unit: its top-level tests are already selectable.
The Unicode sweep and JSON parity test contain no `t.Run` around their heavy work.
Markdown layout mutants are selectable, but the parent performs corpus generation,
compilation and batch parity outside them; splitting those children would repeat
the heavy work. No test files were changed.

## Compare with a plain reference branch

Local raw logs remain supported. To fetch and compare a reference at the merged
plan's SHA, use:

```sh
adamic-gate compare /workspace/merged git:origin:gate-logs/TESTED_SHA/plain
```

The branch's SHA may be abbreviated to at least seven hex digits. The archive
must contain `gate-out/test.jsonl` and `gate-out/run-notes.txt`, with exactly one
`commit=<full tested SHA>` line matching `merged.json`'s `Plan.Commit`. A plain
reference at 2adf65c supplies timing weights for newer code, but cannot certify
a fleet at the new tip. Run the new reference at that fleet's exact SHA and
provide its corresponding branch to compare.

The command fetches fresh evidence into a temporary disk-backed directory and
removes it afterward. It neither checks out the evidence branch nor reuses a
cached archive. Every Git child has a bounded process-group deadline. Only the
two expected regular archive members are copied; missing, linked or duplicate
members, a wrong tested SHA, and corrupt compression are refused. The ordinary
comparison still checks every test verdict and the terminal event count.

The real local-remote test compares green, replaces the same branch's archive
with a wrong-SHA mutant and requires refusal, then changes one terminal verdict
and requires failure. A source-overlay mutant removing the tested-SHA check
fails this test. All 42 named gate-package tests pass with `-race`; vet and
Darwin arm64 cross-compilation pass. A transport smoke check fetched the actual
origin archive and compared it with a summary derived from that same log:
4,135 terminal events, empty diff. That validates transport, not a new fleet.
Proof logs are in `cmd/adamic-gate/evidence/replan-proofs.tar.gz`.

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
adamic-gate shard -index 0 -count 8 -jobs 4 -resume -out /workspace/shard-0
# After an interruption, repeat the same command with -resume.
# Other boxes use indices 1 through 7.
adamic-gate merge -out /workspace/merged /workspace/shard-{0,1,2,3,4,5,6,7}
adamic-gate compare /workspace/merged /workspace/whole.jsonl
```

The external launcher supplies the boxes. Indices are zero based. A new output directory starts fresh; `-resume` accepts either a new directory or complete
package checkpoints for the identical plan, environment, toolchain and job count. Merge works from the repository checkout used by the workers, re-enumerates it, and
writes `merged.json` and the concatenated raw `test.jsonl`. A red merge retains its evidence
and exits 1. Compare exits 1 for a different test name, verdict, or distinct terminal-event
count, or for a red merge. Test identity includes both package and the entire Go test name.

## Recursive submodule provenance

Before discovery or reuse of a checkpoint, each shard reads every gitlink recursively
from the tested `HEAD` tree using `git ls-tree -rz`. Nested pins come from the pinned
parent commit, even if its checkout is stale. `summary.json` records `Submodules`, sorted
by repository-relative `Path`, with `Pinned`, `Checked` (the actual `rev-parse HEAD`),
and `Dirty` for each module. In this tree these are `cohere` and `cohere/TypeScript`.
An uninitialized module, a checkout differing from its pin, or tracked/untracked dirt
refuses the shard before any Go discovery or tests. Local submodule ignore settings
cannot hide changes. The audit is repeated after execution and changes make the shard red.

Merge checks every shard against the tested tree's paths and pins and against other
shards' checkout records. Missing, duplicate, unexpected, dirty, or conflicting records
make merge red; conflicts name both shard indices and the submodule. Older summaries
without recursive provenance cannot certify a tree containing submodules.

`TestShardRefusesStaleRecursiveSubmodules` creates two cohere pins with different nested
pins, moves the real checkout, and verifies the shard refuses it by name. It also checks
stale nested checkouts on resume and tracked/untracked dirt. A two-shard miniature gate
in `TestMergeRejectsDifferentCohereCommits` first merges green, then changes one recorded
cohere checkout and requires a red merge naming shards 0 and 1 and cohere. No new cache
is introduced; each invocation, including resume, performs the provenance audit again.

Verified after merging integration at `2adf65c`: 34 named package tests pass with
`go test -race -count=1 -json ./cmd/adamic-gate`; focused vet and Darwin arm64 test
cross-compilation pass. Source-overlay mutants removing the shard preflight and the
merge provenance check each fail their respective real-entry-point regression test.
Raw JSON, separate stderr, and overlay inputs are in
`cmd/adamic-gate/evidence/submodule-provenance.tar.gz`. No production gate or Mac
execution was attempted. Setup reached Go 1.27.1, clang 20.1.8 and Node 24.19.0
on a box with `nproc=5`, then failed downloading `github.com/klauspost/compress@v1.20.0`
because the module proxy's Google Storage redirect returned Forbidden. The existing
toolchain, sourced from `/workspace/adamic-tools/env.sh`, completed these checks.

## Selection and coverage

Discovery uses `go list ./...` and `go test -json -list . ./...`, including examples and fuzz
seed tests. Packages with no tests are build-only units. Ordinary units are top-level tests;
the following audited parents have independently selectable, enumerable children:

- `internal/oracle`: `TestNativeAgreesWithNode`, `TestInputAgreesWithNode`, and
  `TestFreshWriteProbesStayRefused`.
- `stage1/cohere/typeaware`: the 15 literal `changes` rows of `TestVolumeAgreementAndMutants`.
- `internal/native`: `TestNormalizeMatchesNode` and `TestStringIndexMatchesNode`.
- `stage1/cohere/lint`: `TestMutants` and `TestVolumeMutants`.
- `stage1/cohere/css`: the two modes of `TestCSSPrinterAgreesWithGo`, including each mode's nested mutants.

The first two fixture lists come from a Go test overlay that observes the registered slice
after every package `init()` runs, with every named input checked on disk. Globbing all oracle testdata would incorrectly include helpers
and probes. Fresh probes come from the same glob the test uses. Native sweep labels come
from the literal slice iterated by the audited parent. Unsupported enumeration fails loudly.
The selector escapes regex metacharacters and anchors each slash-separated component.
A top-level, unparenthesized `|` joins complete anchored paths. Go testing splits these into
alternative full-path matchers, so no Cartesian product is formed. One invocation per package
is sufficient, except the separate native WASI probe that requires SDK clang on PATH.

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

## Required WASI coverage

The planner reads test source skip conditions. Direct `os.Getenv` comparisons and local aliases
of variables whose names contain `WASI` identify required test parents. Their selected children
inherit the requirement. Unknown predicates, dynamic environment names, `LookupEnv`, globals or
helpers containing WASI environment reads fail planning and require an audit. Nonstandard getters
such as `syscall.Getenv`, custom environment calls and named WASI constants also fail closed. The runtime branch's
`TestWASIAgreesWithNode` and `TestWASIEmission` use the ordinary oracle fixture table, so both are
enumerated before execution. This list is derived rather than maintained by test name.

When present, `Plan.WASI` names the last shard, its source-derived gates and provisioning command.
Every required unit belongs to that shard; ordinary work uses the other shards. Provision it with
`bash cloud/setup.sh --wasi-sdk`, source the printed env file, then export
`ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1`. Setup supplies `WASI_SYSROOT`. This requires the runtime
setup implementation until it lands on main. The runner refuses missing variables, SDK headers,
libc, executable SDK clang or its linker before executing shard tests. `TestWASI` gets a separate invocation
with SDK clang on PATH; native checks retain native clang. Resumed evidence includes the entire SDK
(header, library, compiler, linker and resources) in its input identity.

Merge reads raw events and rejects every skip within a required unit, including a skipped parent
of selected children, regardless of the reason. `TestWASIEmission` contains a skip clause for fixtures
that do not lower. All 363 rows at the audited runtime SHA are marked lowering, so that clause did
not produce skips in the build mutant. Any future use of it prevents a whole-green verdict under
this rule. This unit does not edit existing tests to hide skips. A missing-variable preflight refusal produces
no test evidence. An old or mutated runner that bypasses that check and records skips is rejected
by merge with the required unit names.

### WASI proof

Fetched and merged `origin/area/developer-tools` at `14e8372816b1d3bf5bcc37ca57336e46d41b7a41`
with `git merge`, without rebasing. The source audit covered main
`f8013f0baac41ddc340d76f83bddde38536a8f07` (no WASI gates) and runtime
`4d86c305dda261768b35687d199c1b2188c7ab71` (seven gated parents). Runtime has 731 required units:
363 fixture children under each oracle parent and five other top-level tests. The required shard
is index 7 at count 8. Its new timings are unknown; no healthy complete WASI duration is claimed.

The runtime setup installed SDK 27, with its multiarch `include/wasm32-wasi` header layout.
The source audit uses lexical AST variable identities so a later, shadowed compiler error cannot
turn an earlier tool-availability skip into an unrecognized environment predicate.

A detached runtime checkout held the real compiler controls: the anchored request and throwing
request tests both passed, uncached. The build mutant changed the actual `native.Flags` target from
`--target=wasm32-wasi` to `--target=wasm64-wasi`, committed at
`96686785269691a9a1f34dfe2c21ec7726ed8e0a`. The whole required shard produced **38 pass, 731 fail,
0 skip**. Merge returned red and explicitly named shard 7 and `cmd/adamic::TestWASIRequest`;
the compiler output says the Wasm64 standard headers are missing. The unchanged native `TestWASI`
probe and its 36 children passed using their own Wasm32 flags.

The missing-variable mutant at `285113a7b5b0fb7346a49fb2f97a15602013531d` bypassed only the runner's
preflight, then ran the same required shard with all three WASI variables unset. It produced
**0 pass, 0 fail, 7 skip** by canonical test name, with 19 raw terminal test events because selected
fixture parents repeat across selector invocations. Merge returned red, naming required skipped
units and their actual opt-in skip messages. The final production binary also refused this run
before creating output evidence.

These two real shard merges also reject the seven absent fleet shards. To isolate the skip verdict
from those errors, a complete miniature repository supplied a mandatory environment-gated test,
an ordinary optional skip and placeholder SDK files. Its control merged green (1 pass, 0 fail,
1 skip). Bypassing preflight and removing the variables merged red (0 pass, 0 fail, 2 skip), with
**exactly one error**, naming `gateprobe::TestWASIProbe` as a skipped required unit. This miniature
checks the merger and does not execute wasm.

Thirteen tool tests and focused vet pass. Eleven source-overlay mutants fail their intended tests:
SDK resume key, mandatory skip verdict, preflight, unknown predicate, each opt-in variable, SDK
header, libc, executable compiler, multiarch layout and linker. The SDK key mutant is caught by
changing real sysroot bytes and linker bytes with the environment unchanged. The target and missing
variable mutations change only detached scratch checkouts; repository test files were not edited.

| Loop | Before | After | Instrument |
| --- | --- | --- | --- |
| Required WASI shard, build mutant | Healthy full shard unmeasured | 148.074 s | `ADAMIC_TOOLS= ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 adamic-gate shard -index 7 -count 8 -out <build-mutant>` |
| Required WASI shard, missing variables | Correct runner refuses to start | 34.645 s with mutated preflight | `env -u ADAMIC_TEST_WASI -u ADAMIC_ORACLE_WASI -u WASI_SYSROOT ADAMIC_TOOLS= adamic-gate shard -index 7 -count 8 -out <skip-mutant>` |

These are correctness experiments, not performance comparisons or healthy shard predictions.
Both source the setup env file before invoking the command. The uncached build-flags lines are:

```
commit=96686785269691a9a1f34dfe2c21ec7726ed8e0a nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="2.08 6.03 6.93 2/260 213343" load_after="1.64 4.34 6.16 1/263 221308" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps=""
commit=285113a7b5b0fb7346a49fb2f97a15602013531d nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="1.29 6.63 7.17 1/259 211138" load_after="1.56 6.20 7.01 5/261 212609" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps=""
```

SDK clang identifies itself as `clang version 20.1.8-wasi-sdk` with that same LLVM source revision.
The small `cmd/adamic-gate/evidence/wasi-evidence.tar.gz` preserves raw logs, summaries, manifests,
mutant scripts, the isolated merger proof and a verified Git bundle of the exact mutant commits.
The bundle requires runtime commit `4d86c305` already in the repository; import its two
`refs/gate-proof/*` references locally and use detached worktrees to inspect them. No proof branch
was pushed. No full unsharded gate, eight-shard run, healthy full WASI shard or fleet-wide green
comparison was run on this box.

## Coverage repair after the first fleet comparison

The earlier green verdict was incomplete. At `2e165469ec95`, the user measured 3,146 distinct
terminal tests in the plain run and 3,099 in eight shards. A source census at that exact SHA
observed 275 initial oracle rows and **322 rows after initialization**, a difference of exactly 47.
The old AST reader examined the initial slice declaration in `oracle_test.go` and ignored the
other test files' `init()` registrations. The complete missing path list is checked in at
`cmd/adamic-gate/evidence/missing-47.json`. Registration sources are:

| File in internal/oracle | Missing rows | Registration |
| --- | ---: | --- |
| class_inheritance_test.go | 15 | append class feature, inheritance and identity rows |
| narrowed_test.go | 12 | loop over `047cb0d_n_*` names and construct paths |
| regexp_cycle_test.go | 5 | conditional append after reflecting on `ir.Program.Regexps`; includes fresh regexp_tree.ts |
| narrowed_number_field_test.go | 4 | construct paths and checked fields in a loop |
| iteration_test.go | 2 | append literal paths in a loop |
| library_map_set_iterator_test.go | 2 | concatenate names into keyed struct rows |
| literal_optional_test.go | 2 | append optional and shape rows |
| literal_undefined_test.go | 2 | append discriminated and widened undefined rows |
| devirtualize_test.go | 1 | append a row directly |
| override_representation_test.go | 1 | append a row directly |
| prototype_test.go | 1 | append a row directly |

Every split parent now has a mandatory complement in `Plan.Complements`. At count eight it
belongs to shard 7, or shard 6 for ordinary parents when shard 7 is reserved for WASI. That
shard runs `-run '^Parent$' -skip '<exact children owned by other shards>'`. Its own known
children and every unknown child remain eligible. All skip components are anchored and quoted,
including slash levels. Every other shard runs its exact assigned paths. There is no result cache
or persistent fixture-discovery cache; the listing test executes all initialization each plan.

Merge requires complete package logs, the complement invocation's exact run/skip arguments,
a terminal complement parent event, every assigned unit's terminal event and matching run/terminal
counts. It rejects duplicate children within or between shards. Structural ancestor events can
repeat and are reduced by test name, as before. `merged.json` records `SplitChildren`, the observed
union for every split parent, and `Unplanned`. Unknown children accepted only from the designated
complement are printed as **unplanned, ran anyway**. Their verdicts and required WASI skips affect
green exactly like planned children. An arbitrary future child need not appear in enumeration to
execute and participate in the verdict. Unknown work can still make the complement slower.

### Scheduling and bounded measurement

The previous runner processed packages serially and invoked a package repeatedly for different
literal prefixes. That loses `go test ./...` package concurrency and repeats parent setup,
compilation/linking checks and process startup. The corrected runner uses `-jobs` independent
package workers (default `runtime.GOMAXPROCS(0)`, four on this cgroup), durable per-package logs,
and one complete-path union per package. Scratch is removed when each package finishes; aggregate
logs are assembled in stable package order after the workers finish. Changing `-jobs` invalidates
resume evidence, as do changes to either `-run` or `-skip`. Concurrent package execution can raise
peak disk use; use `-jobs 1` for a sequential diagnostic or tight disk budget.

The user's 80.5-minute shard/count-one observation versus 38-minute plain observation is not a
controlled comparison collected on this box. The scheduling mechanism above is directly visible
in the old code. Its exact contribution to those production durations remains unmeasured.
A fixed miniature corpus measured the change on this box, interleaved before/after, best of three:
16 packages with 300 ms tests plus 32 split children whose parent costs 180 ms per invocation.
Each shard selected two ordinary packages and four children. Old: six serial Go invocations.
New: three package invocations, four package workers. These are **bounded scheduling timings in
seconds, not whole-gate or production-shard timings**.

| Loop | Before | After | Instrument |
| --- | ---: | ---: | --- |
| Corpus shard 0, best of 3 | 3.541 s | 1.695 s | `OLD shard -index 0 -count 8 -out before-R-0` / `NEW shard -index 0 -count 8 -jobs 4 -out after-R-0` |
| Corpus shard 1, best of 3 | 3.443 s | 1.565 s | `OLD shard -index 1 -count 8 -out before-R-1` / `NEW shard -index 1 -count 8 -jobs 4 -out after-R-1` |
| Corpus shard 2, best of 3 | 3.433 s | 1.602 s | `OLD shard -index 2 -count 8 -out before-R-2` / `NEW shard -index 2 -count 8 -jobs 4 -out after-R-2` |
| Corpus shard 3, best of 3 | 3.461 s | 1.643 s | `OLD shard -index 3 -count 8 -out before-R-3` / `NEW shard -index 3 -count 8 -jobs 4 -out after-R-3` |
| Corpus shard 4, best of 3 | 3.386 s | 1.556 s | `OLD shard -index 4 -count 8 -out before-R-4` / `NEW shard -index 4 -count 8 -jobs 4 -out after-R-4` |
| Corpus shard 5, best of 3 | 3.396 s | 1.575 s | `OLD shard -index 5 -count 8 -out before-R-5` / `NEW shard -index 5 -count 8 -jobs 4 -out after-R-5` |
| Corpus shard 6, best of 3 | 3.449 s | 1.581 s | `OLD shard -index 6 -count 8 -out before-R-6` / `NEW shard -index 6 -count 8 -jobs 4 -out after-R-6` |
| Corpus shard 7, best of 3 | 3.425 s | 1.578 s | `OLD shard -index 7 -count 8 -out before-R-7` / `NEW shard -index 7 -count 8 -jobs 4 -out after-R-7` |

`OLD=/workspace/adamic-gate-wasi-final` (72836ef runner);
`NEW=/workspace/adamic-gate-complement` (f7586d1 implementation plus the job-count identity check).
The corpus commit is `798a0623d35eb7e06a170ac26832ba45f27dad5f`, unchanged for all 48 runs.
Every observation's exact argv and full build-flags line, including both loads, are in the
coverage evidence archive's `coverage-bench/results.json`. Toolchain: nproc=5,
cpu.max="400000 100000", Go1.27.1, clang20.1.8 (LLVM87f0227), Node24.19.0, uncached=1.
Source the setup env file first; the experiment sets `ADAMIC_TOOLS=` and unsets `WASI_SYSROOT`,
so hashing a provisioned SDK tree does not dominate a miniature scheduling measurement.
Its complete eight-shard merge is **49 pass, 0 fail, 0 skip**, with 56 raw terminal events because
seven parent events repeat. Its separate plain run has **49 pass, 0 fail, 0 skip**, 49 raw events;
comparison by canonical test name and verdict has an empty diff.

A second real CLI proof intentionally filters an `init()`-registered `planted_init.a` fixture
out of discovery in a scratch repository. The same faulty planner is used by both shards and
merge. The complement executes it; merge is green with three terminal tests and prints the
unplanned notice; the plain run has the same three tests and comparison is empty. Removing
complement invocation provenance from a copy makes merge red. Five source-overlay mutants are
caught by tool tests: omit the skip flag, ignore complement duplicates, omit the complement parent
terminal check, omit skip selectors from checkpoint keys, and omit job count from resume identity.
The last two mutate newly added cache-key components; no persistent cache was added.

The follow-up fetched `gate-logs/2e165469ec95/plain` and extracted `plain.tgz`.
Its `gate-out/test.jsonl` has exactly 3,146 distinct terminal tests and 322 children under
`TestNativeAgreesWithNode`. At that same source SHA, the post-init fixture set equals the plain
child set exactly, and plain children minus the 275 initial rows equals the checked-in 47-name
list exactly. All 47 passed; both set differences are empty. The reproducible result and plain-log
SHA-256 are in `cmd/adamic-gate/evidence/plain-47-check.json`. This verifies the historical coverage
hole against the actual reference, rather than inferring it from counts. It does not substitute
for comparing the new full fleet against a plain run at the new SHA. No long production reference
or sequential eight-shard run was started. Production per-shard before/after timings and the
final whole-gate diff must come from the new fleet at the same clean source SHA.

The planted-init CLI proof was rerun for the follow-up with the reader deliberately filtering
`planted_init.a` out of discovery. The fixture was appended by an `init()` loop constructing its
path. Shard 1 of the two-shard proof ran it despite having no planned children; merge printed
`unplanned, ran anyway: coverageprobe/internal/oracle::TestNativeAgreesWithNode/internal/oracle/testdata/planted_init.a`.
Merge and plain both had three distinct terminal tests, and comparison was empty. Missing
complement invocation provenance still made merge red. The fresh raw proof and 17 passing
race-tested tool tests are in `cmd/adamic-gate/evidence/coverage-followup.tar.gz`.

At count eight, if `Plan.WASI` is present, the required WASI shard is **7**, ordinary split-parent
complements are **6**, and WASI split-parent complements also run on **7**. Thus ordinary
complements and the required WASI shard are separate, while WASI parents share the required
shard. If the checkout has no source-derived WASI gates, `Plan.WASI` is absent and every complement
uses the last shard, **7**. This branch's current source has no WASI gates; runtime integration
activates the 6/7 arrangement automatically. `TestComplementPlacementWithRequiredWASI` checks
both layouts and their exact run/skip selectors.

### Fleet commands for this repair

Use the same clean branch commit on all workers and on the merger. Setup and the selected SDK
environment must remain stable across restarts. Keep evidence outside the checkout. On each box:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh # use the actual path printed by setup
# If Plan.WASI exists, provision its designated worker with setup --wasi-sdk,
# source that env file, and export ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1.
go build -o /workspace/adamic-gate ./cmd/adamic-gate
/workspace/adamic-gate plan -count 8 > /workspace/gate-plan.json
# One index per box; substitute 0 through 7, never loop here.
/workspace/adamic-gate shard -index INDEX -count 8 -jobs 4 -resume \
  -scratch /workspace/adamic-gate-scratch -out /workspace/shard-INDEX
```

Use the Markdown dependency path provisioned by setup, or provision/export the pinned path
shown at the top of this document if the checked-out setup version lacks it. After transferring
all eight output directories to the merger checkout:

```sh
/workspace/adamic-gate merge -out /workspace/gate-merged /workspace/shard-{0,1,2,3,4,5,6,7}
/workspace/adamic-gate compare /workspace/gate-merged /workspace/plain.jsonl
```

The plain reference must be from that same source SHA; if WASI gates exist, enable them and use
the SDK as required by the test implementation. Reusing 2e16546's log after changing the gate tool's
own tests will naturally show those added tests as a diff. Inspect `Plan.Complements` before launch
and `Unplanned` after merge. Predictions are known elapsed work estimates, not a guarantee of
parallel wall time; unknown complement work, missing timings, discovery and concurrency contention
remain unpredicted.


## macOS stderr regression repair

`TestComplementRunsInitRegisteredFixture` previously directed both streams into its `.jsonl`.
On macOS, a Go toolchain diagnostic on stderr made that file malformed. The selector test and
runtime fixture-discovery listing had the same defect. They now use `runJSONCommand`: stdout
alone goes to the JSON file, stderr to its sibling `.stderr`. A failed command reports that
stderr path and diagnostics. The production shard runner already kept separate per-package
`test.jsonl` and `test.stderr` streams. Structured queries made through `output`, including
`go test -json -list` and `go env -json`, now also parse stdout independently of stderr.

The complement regression sets `GOPATH` to a temporary directory containing a `go.mod`, while
running a separate valid fixture module. Go emits the real warning
`go: warning: ignoring go.mod in $GOPATH ...`. The test requires that warning to exist in the
stderr file, then parses the stdout log and checks its exact assigned/complement children.
The fixture-discovery regression repeats that condition for its overlay listing and structured
query. The tiny fixture module requires Go 1.27, avoiding an unnecessary patch-version download.

Both raw-log readers now reject malformed JSON with the filename and one-based line number.
A clean eight-shard miniature merge stays green; appending a `go:` line to a copied shard's raw
log makes merge exit 1 with `bad-shard-0/test.jsonl:41: malformed JSON`. This is a hard refusal;
no green merged artifact is emitted for the malformed stream. Fixture discovery also fails on
malformed stdout instead of ignoring a bad line.

Linux host metadata lives in `host_linux.go`; Darwin uses `host_darwin.go`. Darwin reads logical
CPU count and load through BSD `sysctl -n`, reports cgroup CPU quota as `not-applicable`, and
requires neither `/proc`, `/sys` nor GNU `nproc`. Metadata tests provide POSIX-shell command
fixtures and explicitly detect any Darwin invocation of `nproc`.

All 19 package tests pass with `-race` in the ordinary run and in a Linux run whose overlay
selects the Darwin metadata implementation. The complete package test binary cross-compiles to
Mach-O for both Darwin arm64 and amd64. Focused vet passes. No actual Mac is available, so these
checks exercise Darwin metadata semantics and compilation, rather than claiming execution on
Darwin's kernel or filesystem. Commands, logs and mutants are preserved in
`cmd/adamic-gate/evidence/stderr-portability.tar.gz`:

```sh
go test -race -count=1 -json ./cmd/adamic-gate > tests.jsonl 2> tests.stderr
# The archived overlay replaces host_linux.go with host_darwin.go on Linux.
go test -race -count=1 -json -overlay=darwin-semantics-overlay.json ./cmd/adamic-gate > darwin-semantics.jsonl 2> darwin-semantics.stderr
GOOS=darwin GOARCH=arm64 go test -c -o gate-darwin-arm64.test ./cmd/adamic-gate
GOOS=darwin GOARCH=amd64 go test -c -o gate-darwin-amd64.test ./cmd/adamic-gate
go vet ./cmd/adamic-gate
```

Four source-overlay mutants fail their intended tests: join stderr to JSON stdout, restore
combined output for structured queries, remove file/line diagnostics, and use GNU `nproc` in
Darwin metadata. No cache was added or changed.


## Child-deadline integration and skip-census landing hook

Merged `origin/area/developer-tools` at `7863216ed419151286c34df3896496dd15f4ec11`,
which contains the `97ac3d6` child-deadline audit, using a merge rather than a rebase.
Conflicts in the gate entry point and selector test retain stdout-only JSON logs while using
`internal/boundedrun` commands. Every Go child launch in the gate package now uses the bounded
constructor: metadata/listing probes, fixture overlays, formatting/vet, concurrent package
workers and test probes. Probe queries use the audit's 30-second limit, cold Go listings and
fixture discovery use 10 minutes, and real shard test invocations use 70 minutes around Go's
existing 60-minute test timeout. Contexts are released after reaping; deadline expiration kills
the process group. The helper only accepts a `*boundedrun.Cmd`, preventing a raw `exec.Cmd`
from silently bypassing its contract. Buffered probe stderr uses a locked snapshot so bounded
pipe draining cannot race the returned diagnostic.

The structural complements, required WASI derivation/preflight/skip rejection, resume validation,
and concurrent package scheduling remain in place. Twenty-four gate-tool tests and the three
bounded-run tests pass with `-race`; focused vet and Darwin arm64 test cross-compilation pass.
The two additional deadline proofs exercise the JSON-file helper and actual fixture discovery
against a hung child with a grandchild heartbeat. Actual concurrent shard workers also execute
under the race detector on the bounded scheduling corpus. Logs and the skip-hook mutant are
preserved in `cmd/adamic-gate/evidence/deadline-merge.tar.gz`.

Fetched and read `devtools/skip-census` at `a2bf5a93db623cdd45d4975efb77e48b1cac8231`.
Its `internal/skipcensus` package is absent from both the integration area at `7863216` and
main at `c7991b9`; it has not landed. No census classifications are invented or copied into
this branch. `checkSkipCensus` in `skip_policy.go` is the explicit merge integration hook:

- Before `Green` is computed, its result is recorded as `SkipCensus` in `merged.json`.
- Until the package lands, `Status` is `not-landed`; classification arrays are empty.
- If the census package's source appears before the hook is wired, merge fails with a
  `not-wired` status, rather than silently bypassing the checker.
- When it lands, replace the hook with source `Scan`, declaration `Load`/`Validate`, and
  `CheckLog` over the shard streams. Feed required-input and unclassified errors into
  `Errors`; populate the named `NotApplicable` and `Measurement` summary arrays from the
  checker's report. Keep the census declaration file as the single classification authority.

The hook test plants the package source in a temporary tree and proves the unwired refusal.
A source-overlay mutant allowing that bypass fails the test. Existing required WASI skips
remain red independently. General census-required-input enforcement and classified skip
listing await the census landing; this branch does not claim those checks are active yet.


## Deadline proofs wait for readiness

On a heavily loaded Mac, the 200 ms test deadline could expire before the fake grandchild
created its heartbeat. That proved startup was bounded, but could not prove that a running
grandchild was killed. The shared fixture now waits for the first heartbeat, signals parent
readiness, and only then blocks. Its cleanup records the actual process group, including when
the fixture runs below an isolated runtime builder.

All affected Go proofs install a nonparallel readiness clock through `boundedrun.WithTimeout`.
Startup has a separate 30-second bound; the requested short execution deadline is armed only
after the heartbeat and parent signal exist. The command, process-group cancellation and bounded
wait implementations remain real. `WithTimeout` defaults to ordinary `context.WithTimeout`;
production still counts startup against its original 30s/10m/70m limits and environment cap.
A constructor test verifies that default remains immediate. Cancellation causes preserve the
execution-deadline classification, including the test262 runner's `TimedOut` result.

The audit covered every shared fixture user: gate probes, shard constructor, JSON-file command,
fixture discovery, boundedrun's group/pipe/writer waits, fuzz execution/preparation, test262
execution/runtime builder and persistent compiler worker. The exited-leader test waits for child
readiness before exiting rather than sleeping for 100 ms. The blocked-writer proof waits until
its writer is entered before arming its short deadline. The compiler worker waits for its tree
before starting the first request's own timer. Shutdown observation starts at readiness and
retains a finite five-second allowance for the bounded reap and scheduler contention.

A deterministic test delays the first grandchild heartbeat by 650 ms, longer than its 200 ms
execution deadline. It still passes with the two phases, and rejects a child that exits before
readiness. Mutants restoring the early timer and killing only the parent both fail that test;
the latter is caught because the grandchild heartbeat keeps changing.

Python's deadline proof delays the real `communicate(timeout=...)` call until readiness. Shell
proofs still run real GNU timeout: a test shim supplies its bounded startup watchdog and sends
the execution SIGALRM only after readiness, exercising its actual signal/kill-after path. The
shim preserves the shell's diagnostic descriptor and records the group for cleanup. Neither
Python nor shell production limits were changed. These shell tests retain their existing GNU
timeout dependency. The historical manual `prove.py` benchmark is not a load-proof unit suite
and was not rerun; its already-recorded results were not rewritten.

Verification sends stdout and stderr to separate files:

- Initial load trial: three parallel `go test -race -count=20` invocations with 40 CPU burners,
  720 named deadline passes, zero failures.
- Final code: two parallel invocations with the same burners and count, 560 named passes,
  zero failures; load average reached 39.55 on this four-core cgroup.
- The full four-package race suite: 101 pass, zero fail, one measurement skip,
  `cmd/adamic-test262::TestCompilerStartupMeasurement`.
- Twenty Python/shell suite repetitions: 120 tests pass. Repetitions overlapped the load trial;
  the CPU supervisor stopped when the Go trial completed, so later Python repetitions ran idle.
- Vet passes for all four changed Go packages. Gate tests cross-compile for Darwin arm64;
  no actual Mac is available on this worker.

The Go load command, launched concurrently under the CPU supervisor, was:

```sh
go test -race -count=20 -json -run 'Deadline|ExitedLeader|WaitCannot|CompilerWorkerTimeout' ./cmd/adamic-gate ./internal/boundedrun ./internal/fuzz ./cmd/adamic-test262 > load.jsonl 2> load.stderr
```

`cmd/adamic-gate/evidence/readiness-proofs.tar.gz` preserves the supervisors, raw trial logs,
source hashes, toolchain/CPU metadata, full suite, Python tests, Darwin build, vet and both mutants.
This is a scheduling-contention correctness proof, not a whole-gate performance measurement.


## Concurrency measurement before the provisioned fleet

All four trials tested fixed commit `6f61080201f13b84cdec1bc6078b1813d135ac21`,
uncached, on this box. Shard 8 is the heaviest shard of the 15-shard plan: predicted
1,152.940 seconds, 51 planned units. Required parity corpora and the verified tsgo
C archive were supplied. The three requested settings each ran once because they
take over five minutes. A fourth run repeated the baseline after the alternatives
to detect run-order/compiler-cache effects. Each emitted the identical set of
429 named terminal verdicts: 429 pass, zero fail, zero skip; summary errors empty.

| Loop / setting | Wall seconds | Peak RSS GiB | Sampled cgroup peak GiB | Load 1/5/15 min before → after |
| --- | ---: | ---: | ---: | --- |
| today (`auto`) | 654.116 | 2.230 | 12.592 | 4.99/5.23/3.16 → 1.09/3.88/4.33 |
| four-one (`4x1`) | 510.437 | 2.225 | 12.551 | 1.09/3.88/4.33 → 1.03/2.32/3.55 |
| two-two (`2x2`) | 562.544 | 2.225 | 12.308 | 1.03/2.32/3.55 → 1.04/1.80/2.86 |
| today-repeat (`auto`) | 524.655 | 2.222 | 12.744 | 1.04/1.80/2.86 → 1.09/2.15/2.87 |

The fastest alternative, `4x1`, is only 2.710% faster than the repeated baseline,
below the requested 5% threshold. Retain today's default: `-concurrency auto`.
`2x2` is slower than the repeated baseline. The initial baseline paid additional
compilation work despite plan-discovery warming; treating its 21.97% difference
as a concurrency improvement would confound compilation-cache/run-order effects.
ADAMIC_GATE_UNCACHED bypassed answer caches in every trial; the Go compilation
cache remained enabled. No new answer cache was added.

Here `auto` resolved to five package jobs and Go's unset `-parallel`, whose effective
default was five: a nominal budget of 25, despite a four-core cgroup quota.
`4x1` and `2x2` each have a nominal budget of four. Limits apply to package workers
and Go parallel tests; they do not cap every native subprocess inside one test.
`-concurrency JOBSxPARALLEL` sets both through one flag. The shard summary records
Setting, Jobs, Parallel, EffectiveParallel and Budget; its build-flags line also
records them. Resume keys include both explicit and effective parallelism, and
merge validates package checkpoints against the actual `-parallel` argument.
Two source-overlay mutants dropping those individual key components must fail
`TestConcurrencyBudgetAndResumeIdentity`; the checkpoint test also rejects an
invocation with different parallelism.

Every trial's build flags: commit above; `nproc=5`; `cpu.max="400000 100000"`;
`go version go1.27.1 linux/amd64`;
`clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`;
`node=v24.19.0`; `uncached=1`; GOFLAGS, CGO_ENABLED and GOMAXPROCS unset.
Per-trial loads are in the table and complete build-flags lines are in summaries.
GNU time's maximum RSS is a process peak, not the sum of concurrent process RSS.
The supplemental cgroup peaks were sampled at 100 ms and include page cache and
other cgroup processes, so they are not an isolated application-memory measurement.

Exact instrument, with CASE/SETTING respectively `today/auto`, `four-one/4x1`,
`two-two/2x2`, `today-repeat/auto`, in that order:

```sh
source /workspace/adamic-tools/env.sh
source /workspace/gate-concurrency-inputs.env
export ADAMIC_GATE_UNCACHED=1
/workspace/gate-concurrency-measure/time-tools/usr/bin/time -v \
  -o /workspace/gate-concurrency-measure/CASE/time.txt \
  /workspace/adamic-gate-budget shard -index 8 -count 15 -concurrency SETTING \
  -scratch /workspace/gate-concurrency-scratch \
  -out /workspace/gate-concurrency-measure/CASE/shard
```

Before/after for the scheduling change: repeated auto 524.655 s before,
4x1 510.437 s after, instrument above; retained default because the difference
is under 5%. No whole gate or other fleet shards were run for this measurement.

The timing replan remains: 8 shards 33.92 min; 10 27.62; 12 23.42;
14 20.42; 15 19.22; 16 18.17 predicted maximum. Choose **15** as the smallest
count below twenty minutes; 16 gives more margin. These predictions retain the
plain-log weights, rather than scaling the fleet using this one shard's observed
wall time. Unknown complement work and host contention remain prediction limits.
No single known unit exceeds fifteen minutes; the largest remains
`internal/unicodeproperties::TestCanonicalizeUnicodeNode`, 822.510 s.

For the fleet, build the CLI once, generate `adamic-gate plan -count 15`, then run
one index per box with `adamic-gate shard -index INDEX -count 15 -concurrency auto
-resume -scratch DISK_SCRATCH -out OUTPUT`. Provision gate inputs everywhere and
the WASI SDK/variables on the required WASI shard as documented above.

Setup's module warming hit the already-observed forbidden compress-module ZIP
redirect. The pinned tsgo C archive was built directly, and verified gate-input
environment was generated from setup's helper. GNU time was extracted from the
Ubuntu package because this box had no installed time binary. The evidence archive
contains raw stdout-only JSON logs, separate stderr, summaries, commands, samples,
verification and both key mutants.


## Portable submodule fixtures

The fixture made no chmod changes to its Git object directories and set up no
alternates. It did inherit every Git environment variable, global/system config
and init template. Those inputs can redirect object writes away from the writable
temporary repository; a read-only redirected object store can therefore break a
commit despite a writable TMPDIR. The failing Mac's environment was unavailable,
so its exact redirect is not established. On this Linux worker an inherited
template with an object-directory symlink reproduced fixture setup failure.

Fixture Git children now discard inherited GIT_* variables, disable global and
system config, templates, hooks and commit signing, and retain bounded process
group deadlines. Each initialized repository owns its object directory. Production
Git commands retain their caller environment. The regression supplies hostile
object/index paths, an alternates path, a template symlink and global template config
without changing process-wide environment or using Linux paths or GNU flags.
It checks a private object directory and absence of alternates. The original
uninitialized-submodule test first proves clean recursive provenance, then removes
the nested gitfile and requires the name cohere/typescript-go in the refusal.

The race suite passes 45 named tests; twenty repetitions of both focused proofs
pass with a hostile inherited template. Removing fixture environment isolation is
caught by the new test. Vet and Darwin arm64 cross-compilation pass. No actual
Mac was available. Raw proof logs are in evidence/submodule-portability.tar.gz.


## Required environment gates and const names

The source audit resolves package-level string const literals, aliases and string
concatenations across files in a package. Mutable names, local shadows, unresolved
expressions and ambiguous declarations still fail closed. Skip predicates retain
the existing restricted comparison/alias audit; helper-routed gates are not assumed
safe. Only variables feeding recognized skip conditions become required inputs,
so the optional ADAMIC_WASI_RUNTIME override is not made mandatory.

The plan now records RequiredVariables and Environment, and each selected unit
records RequiredEnvironment. The required list is ADAMIC_TEST_WASI,
ADAMIC_ORACLE_WASI and ADAMIC_GATE_COHERE. internal/skipcensus has not landed in
the merged developer-tools area (f13e632e); the existing fail-closed census hook
remains, and this list must be replaced by its required-input declarations when
it lands. WASI_SYSROOT retains the SDK-specific readiness validation.

Every required-gate unit goes to the last shard. The cohere baseline is its own
unit, cloud::TestRepositoryPassesCohereBaseline, on the same shard as WASI.
Its reported roughly two-minute cost has not been remeasured or inserted as a
fabricated timing. The plan's printed command enables every required variable;
shard refuses startup if any is absent or not 1, and additionally checks the WASI
SDK when needed. With fifteen shards this is index 14; ordinary complements
remain on index 13, and required-parent complements stay on index 14.

```sh
bash cloud/setup.sh --wasi-sdk
source "${ADAMIC_TOOLS:-/opt/adamic-tools}/env.sh"
ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1 \
  adamic-gate shard -index 14 -count 15 -resume -scratch DISK_SCRATCH -out OUTPUT
```

Merge checks raw terminal events and refuses every skip of a required unit,
including descendants and previously unplanned complement children, regardless
of the skip reason. Its error names the shard, unit, skipped test and input names.
The integration proof first merges a real const-gated fixture green with its
input, then verifies preflight refusal without it. It runs Go with the variable
absent and refreshes the package/aggregate log checksums and verdict summary;
merge goes red for exactly one error: the named required-input skip. Thus malformed
evidence or unrelated validations do not explain the red verdict.

Mutants removing constant resolution, required-skip merge enforcement and
required-input startup enforcement each fail their corresponding test. The
package race suite, vet and Darwin arm64 cross-compilation are the verification
scope; neither the whole gate nor the fleet is rerun.


## Lint registry parents do not create backend children

At proof tree 6edcd86de24b1b5cfca609e96171880e66b5c8ed, TestMutants reads each
registered rule's mutant.json and registers change.Name. Its inner Node, emitted
JavaScript and native rows run those execution modes inside a child; they are
not t.Run registrations. The old literal reader mistook those three rows for
children, including the nonexistent TestMutants/emitted_JavaScript.

The literal reader now requires an unconditional direct t.Run bound to the loop
value and the enclosing test receiver. For struct rows it selects the field that
t.Run actually consumes, rather than assuming the first field names the child.
Conditional calls and a name changed before t.Run are not statically enumerated.
Registry-driven TestMutants uses a whole-parent unit: the literal reader cannot
prove its metadata-driven names, so the planner does not invent selected children.
All of its actual registry children run through the anchored ^TestMutants$ selector.
Existing complement coverage remains for parents that are still split.

The regression fixture contains dynamic registered names and an inner literal
backend loop. It requires the literal audit to reject that loop, then runs the
planner and a real Go test log and verifies every planned unit exists. Restoring
the old reader is caught before a phantom child can be planned; restoring the
old split fails the planner proof. A separate test covers a real name in the
second struct field, conditional registrations and reassigned names.

The fixed planner binary is run against a detached, clean worktree of the exact
proof SHA, with both submodules pinned. Its fifteen-shard plan and a separate
go test -json -list listing are retained in the evidence archive and compared
for every stage1/cohere/lint unit. Gate-package race tests, vet and Darwin arm64
cross-compilation are the verification scope. No lint compiler parity run,
whole gate or new runtime measurement is claimed; the registry parent is whole
and can affect shard wall time relative to splitting it.

Merge now prints a labelled wall-seconds line for every shard, sorted by index,
on both green and red verdicts. merged.json also records ShardWallTimes with
explicit indices, wall seconds and the original build-flags line; the older
WallSeconds array is retained for compatibility. Input-directory order therefore
does not obscure which shard took each time.


## Skip census integration

The gate is based on area/developer-tools 540fa7f0. Merge now calls skipcensus.Scan,
Load and Validate before CheckLog over the concatenated shard streams, which are
exactly the merged test.jsonl content. The checked-in table is the single authority
for required-input, not-applicable, measurement and opt-in-lane classification.
A stale table, unknown skip or required-input skip makes the verdict red. Required
skip diagnostics include the named test and the table's Provides text, including
the missing input and setup instructions.

SkipCensus in merged.json contains named arrays for all four classes and unknown
skips. The printed GATE line carries the not-applicable, measurement and opt-in-lane
names, rather than reducing them to counts. Legacy WASI/required-environment skip
checks remain only for trees without a landed census; landed trees use its checker.

required_environment.go has no checked-in variable list. Boolean opt-in gates
are derived from census required-input conditions of the form Getenv(key) != "1";
package-level const keys use the existing source resolver. This yields the cohere
and two WASI switches from the current table. Path inputs are supplied by setup
and checked by CheckLog if a test skips. The table's Provides field is prose and
does not encode SDK paths or a machine-readable provisioning command, so the
minimum WASI setup command and SDK-readiness logic remain in wasi.go. No result
cache was introduced.

The regression checks a validated table, all three allowed named skip classes,
a missing required input and an added source skip. Mutants bypassing table
validation or discarding CheckLog's report/errors both fail. The actual shard
proof compares the parser shard with ADAMIC_TYPESCRIPT_SOURCE absent and present,
with other setup gate inputs supplied. Only that shard is supplied to merge: its
census status is checked independently of the expected missing-shard errors.
This is not a claim that a one-shard partial run is a whole green gate.

The cloud proof at 773d5fc79122ec9e98ad2368f4cadd5c43869182 ran shard 1 of 15
uncached with setup --gate-inputs otherwise intact. With ADAMIC_TYPESCRIPT_SOURCE
unset it recorded 377 pass, 0 fail, 1 skip in 620.299 seconds; merge reported:

```text
required-input skip github.com/system-inc/adamic/stage1/typescript/parser::TestWholeCompilerAgrees; input: ADAMIC_TYPESCRIPT_SOURCE: TypeScript v6.0.3 source checkout at 050880ce59e30b356b686bd3144efe24f875ebc8, including src/compiler/*.ts; see docs/gate-inputs.md.
```

Restoring the source recorded 378 pass, 0 fail, 0 skip in 607.991 seconds, with
SkipCensus.Status=checked and empty RequiredInput and Unknown arrays. These were
partial merges, correctly red for the other fourteen absent shards; only the
census check is claimed to pass. Package race tests recorded 50 pass; vet and
Darwin arm64 cross-compilation passed. Both policy mutants failed the intended
regression test. Raw JSON logs, summaries, build flags, setup log and mutant
outputs are in cmd/adamic-gate/evidence/skip-census-773d5fc.tgz. Setup took
281.655 seconds on nproc=5, cpu.max=400000 100000; its full timing lines are
included in that archive. No whole gate was run.

One policy mismatch remains in the authoritative table: TestStage3FixtureHook
is described in its source as dormant outside the stage3 fixture runner, but
the table classifies its missing runner input as required-input. Normal whole
merges therefore refuse that skip. The gate does not override that classification.

## Shared fixture affinity (October 7, 2026)

`TestVolumeAgreementAndMutants` is a whole unit. Its stage0 build, checker
archives, controls, released-handle probes, and TypeScript-source comparison
must run once, before its fifteen rows. Splitting those rows repeats that work.

The plan records an `Affinity` list. Its members keep their individual coverage
identities, but assignment and prediction use one packing unit. The audited
fixtures are:

- `markdownblocks.layoutOnce`: all nine `testBlockLayout` tests, including
  whitespace policy work. The original logs contain eight separate full-library
  fixture markers, with spans from 231.7 to 941.6 seconds.
- `markdownblocks.formatterOnce`: source decoding, text splitting, and Unicode
  widths. The regeneration checks appear on original shards 7, 12, and 4.
- `oracle.gateOnce`: conservatively the entire oracle package, including split
  selectors and their parent complements. Its cache miss summaries on different
  shards confirm that identity initialization is repeated even in uncached mode.

`cmd/adamic-test262` holds its `sync.Once` in an instance, rather than a shared
package fixture. No other package-level `sync.Once` was found by the source scan.
The oracle runtime-library disk cache is unchanged. This change adds no cache.

Run `adamic-gate timings -out timings.json <shard logs>...` to collect observations
from shard boxes. A concatenated merged JSON log also works: package start/end
records preserve separate invocations rather than letting the last parent win.
The command writes `timings.json.audit.json` with raw observations, fixture
spans, log SHA-256 digests, and the original commit/toolchain/load metadata.
Failed or malformed logs are refused. Timings missing from these logs stay
unknown; they are not filled from a plain run.

For a split sequential parent, calibration adds the union of measured child
medians to one median parent residual. Parallel parent spans are retained, not
replaced by a sum of overlapping children. Layout calibration subtracts each
observed cont-to-oracle-marker span and adds one median span. This marker is
before the final release check, so the remaining work is conservatively retained.
Formatter and oracle setup have no separate duration markers; their group
estimates conservatively retain the measured work rather than invent a discount.
These reconstructed estimates are not joint-run measurements.

The calibration source is all fifteen logs at
`gate-logs/47fbaf174d16/20261007T153523/shard-N`. Shards 0 and 13 appeared during
this unit and were fetched before the rerun. All fifteen volume rows and all
nine layout members are now observed. A plain/sharded equivalence certificate
is not claimed. The `Unknown` counts in the plan remain visible.

Use `plan -timings /absolute/path/timings.json` and the same option to `shard`
when measuring an older clean source checkout with the new runner. The plan
contains the effective unit/group costs in its digest. Before the skip census
landed, the historical runner's literal `requiredGateVariables` declaration is
read as that checkout's authority; dynamic declarations are refused.

The regression suite rejects a split fixture, split complement, repeated group
charge, split volume parent, lost parent setup, overwritten merged invocation,
and repeated layout setup. Deliberate overlay mutants prove those checks fail.

### Affinity measurement and 15-shard table

Source: `47fbaf174d168f14cd50dedcaabfaa09c51cdefd`. Runner: `1e9192b65a41bcd108ff04c09b732864b28a5bbe`, binary SHA-256 `9238514115d652e0cc7d456b8853ae703154d64348dc1b40ef2fc6a61a417601`. Measured runner base: `0081e98d25a4e9d560da789a2ad389d38dc43a26`. The final branch is rebased onto the requested `976ce7fa3f4e157353947fcdd60d24ae60ae4f9a`; the historical rerun retains the exact frozen runner binary above.

All fifteen original logs merged green with the original runner: 6,016 pass, 0 fail, 25 skip. Calibration from the merged stream is byte-identical to calibration from the fifteen individual streams. The new plan is preserved in `cmd/adamic-gate/evidence/affinity-47fb-plan.json`.

| Shard | Predicted test seconds | Actual test seconds |
| --- | ---: | ---: |
| 0 | 3456.456 | 2421.176 |
| 1 | 1613.870 | not rerun |
| 2 | 1613.870 | not rerun |
| 3 | 1613.870 | not rerun |
| 4 | 1613.865 | not rerun |
| 5 | 1613.870 | not rerun |
| 6 | 1613.870 | not rerun |
| 7 | 1613.870 | not rerun |
| 8 | 1613.870 | not rerun |
| 9 | 1613.870 | not rerun |
| 10 | 1613.870 | not rerun |
| 11 | 1613.870 | not rerun |
| 12 | 1613.870 | not rerun |
| 13 | 1613.870 | not rerun |
| 14 | 1731.075 | not rerun |

Actual test seconds are the joint Go package test span. Parent elapsed times overlap and include waits on the shared fixture and layout slots; summing them does not measure box time. Only the heaviest resulting shard was rerun.

The rerun reported 39 pass, 0 fail, 0 skip; checks: {'gofmt': 'pass', 'vet': 'pass'}; errors: None. The 1,200-second target was missed. The span is also above 1,800 seconds. This shard already isolates the layout package, plus two packages without tests.

| Loop or phase | Before (original shard observations) | After (this rerun) | Instrument |
| --- | ---: | ---: | --- |
| Layout fixture through original-oracle agreement | 4,742.720 s summed over nine boxes | 448.889 s, one marker | Original JSON logs; new shard command below, `cont` to `lists_test.go:252` |
| Layout mutant checks | See original per-shard audit | 3736.740 s summed across 30 checks; peak 2 concurrent | New shard command, child `run`/`pass` records |
| Joint layout test span | Layout tests were spread across nine boxes | 2421.176 s | Go package terminal `Elapsed` |
| Runner wall | Different original shard allocation | 2527.140 s | Runner `WallSeconds` |
| Runner preflight, checks, and outside-package overhead | Not isolated in original logs | 104.148 s | Runner wall minus layout invocation wall |
| Toolchain and gate-input setup | Original shard 2: 601 s | 84.939 s, warm | `bash cloud/setup.sh --gate-inputs` |

The fixture reached its first mutant after 466.611 seconds. The remaining cost is the full-corpus mutation and whitespace-policy work, including native compilation and execution, with the two-slot layout channel limiting concurrency. Grouping removes repeated fixture construction; it does not remove those checks. The logs and measurement JSON show the individual spans.

The original observations came from other boxes with different allocations. This is not a controlled, interleaved before/after speedup experiment. The requested rerun exceeds five minutes and was performed once. No new cache was added. Answer checks ran with `ADAMIC_GATE_UNCACHED=1`; existing Go build and runtime-library caches were warm.

Exact rerun command, from the clean historical worktree after sourcing the setup environment:

```sh
/workspace/adamic-gate-affinity shard -count 15 -index 0 -concurrency auto -timings /workspace/gate-affinity-evidence/calibration-timings.json -scratch /workspace/gate-affinity-scratch -out /workspace/gate-affinity-evidence/rerun-shard-0
```

Rerun build-flags line:

```text
concurrency="auto" package_jobs=5 test_parallel=0 effective_parallel=5 budget=25 commit=47fbaf174d168f14cd50dedcaabfaa09c51cdefd nproc=5 cpu.max="400000 100000" go="go version go1.27.1 linux/amd64" clang="clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)" node="v24.19.0" load_before="0.50 2.15 3.88 2/261 28575" load_after="1.80 1.97 2.06 1/262 31786" uncached=1 GOFLAGS="" CGO_ENABLED="" GOMAXPROCS="" width_deps="/workspace/adamic-tools/markdown-width"
```

Setup build-flags lines (initial branch setup and warm historical checkout):

```text
setup: build-flags commit=0081e98d25a4e9d560da789a2ad389d38dc43a26 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gate-inputs=true load-before=7.73 3.61 1.36 1/147 2154 load-after=10.10 7.29 3.95 4/189 10007
setup: build-flags commit=47fbaf174d168f14cd50dedcaabfaa09c51cdefd nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gate-inputs=true load-before=2.34 7.14 5.82 1/220 19462 load-after=5.25 6.84 5.83 2/256 22093
```

The source observations carry their own build-flags lines in `timings.json.audit.json` (`Sources`), including commit, nproc, cpu.max, Go, Clang, Node, cache mode, and load before/after.

### Heaviest packing units for the next unit

| Unit | Planning seconds | Measurement basis |
| --- | ---: | --- |
| `stage1/cohere/markdownblocks::@layoutOnce` | 3456.456 | joint rerun 2421.176 s |
| `internal/oracle::@gateOnce` | 1368.845 | conservative estimate from measured shard observations |
| `stage1/cohere/markdownblocks::@formatterOnce` | 1342.480 | conservative estimate from measured shard observations |
| `stage1/cohere/lint::TestMutants` | 1069.120 | measured unit on original shard box |
| `stage1/cohere/markdownblocks::TestMarkdownASTPreprocessing` | 1037.670 | measured unit on original shard box |
| `stage1/cohere/typeaware::TestVolumeAgreementAndMutants` | 731.755 | reconstructed parent: 15 measured rows plus one median residual |
| `stage1/cohere/json::TestPortMatchesGoCohere` | 686.000 | measured unit on original shard box |
| `stage1/cohere/typeaware::TestTypeAwareAgreementAndMutants` | 615.050 | measured unit on original shard box |
| `stage1/cohere/markdowninline::TestMarkdownInline` | 588.250 | measured unit on original shard box |
| `internal/unicodeproperties::TestCanonicalizeUnicodeNode` | 571.080 | measured unit on original shard box |
| `stage1/cohere/lint::TestEmittedJavaScriptMismatch` | 507.020 | measured unit on original shard box |
| `stage1/cohere/css::TestCSSPrinterAgreesWithGo/default` | 466.250 | measured unit on original shard box |

The volume parent ran twelve times for 4,987.88 seconds across the complete original evidence. Its fifteen measured rows sum to 363.13 seconds; one median residual is 368.625 seconds, giving 731.755 seconds for the whole-unit estimate. The layout members sum to 7,722.53 seconds before the explicit shared-fixture adjustment. Oracle and formatter joint spans have not been rerun.

### Largest changed timing observations

| Unit | Previous timing seconds | Shard-box timing seconds |
| --- | ---: | ---: |
| `stage1/cohere/markdownblocks::TestMarkdownWhitespaceLayout` | 762.150 | 1754.000 |
| `stage1/cohere/lint::TestMutants` | 291.970 | 1069.120 |
| `stage1/cohere/markdownblocks::TestMarkdownCodeBlockLayout` | 607.970 | 1136.670 |
| `stage1/cohere/typeaware::TestTypeAwareAgreementAndMutants` | 98.460 | 615.050 |
| `stage1/cohere/markdownblocks::TestMarkdownStructureLayout` | 444.430 | 897.440 |
| `stage1/cohere/markdownblocks::TestMarkdownUnicodeWidths` | 157.160 | 565.620 |
| `stage1/cohere/typeaware::TestVolumeAgreementAndMutants` | 371.760 | 731.755 |
| `stage1/cohere/lint::TestCountGuardMutant` | 34.330 | 352.480 |

Layout member observations include their own fixture construction; the plan charges the affinity estimate rather than summing those raw member times. These comparisons describe a change in calibration source, not a matched benchmark.

### Validation and deliberate mutants

The full runner race suite and vet passed. The final focused race tests also passed. The following overlay mutants all exited nonzero and named the expected failed regression test. Exact commands, mutated sources, overlays, and outputs are in the evidence archive.

| Mutant | Regression that caught it |
| --- | --- |
| `split-affinity` | `TestKnownAffinitiesStayWhole` |
| `repeat-group-price` | `TestKnownAffinitiesStayWhole` |
| `split-volume` | `TestVolumeParentStaysWhole` |
| `drop-parent-setup` | `TestTimingsRetainRepeatedParentSetupOnce` |
| `overwrite-invocations` | `TestTimingsMergedLogPreservesInvocations` |
| `repeat-layout-setup` | `TestTimingsPriceLayoutSetupOnce` |
| `split-complement` | `TestWholePackageAffinityOwnsComplements` |
| `drop-historical-authority` | `TestHistoricalRequiredInputAuthority` |
| `unordered-timing-sums` | `TestTimingsReorderingIsByteIdentical` |

Evidence: `cmd/adamic-gate/evidence/affinity-47fb-shard-0.tgz` and `affinity-47fb-measurement.json`. Fourteen new shards and a complete plain/sharded equivalence comparison were not run. The historical merge was validated under its original runner, before the newer skip census.

### Plain-log provenance

`compare` requires the plain log's full tested commit SHA, Go version, and Node
version to match the merged plan. New plans record `GoVersion` and `NodeVersion`.
Older plans use every shard's recorded `BuildFlags`; missing versions or
inconsistent shard toolchains refuse comparison. Versions come from recorded
run evidence, never the machine running `compare`.

A local `plain.jsonl` must have `plain.jsonl.provenance.json`:

```json
{"commit":"47fbaf174d168f14cd50dedcaabfaa09c51cdefd","go":"go version go1.27.1 linux/amd64","node":"v24.19.0"}
```

Alternatively, sibling `run-notes.txt` must contain exactly one of each:

```text
commit=47fbaf174d168f14cd50dedcaabfaa09c51cdefd
go=go version go1.27.1 linux/amd64
node=v24.19.0
```

Quoted version values and the keys `go_version`/`node_version` also work.
An existing sidecar takes precedence; malformed sidecars cannot fall back to
notes. Git-reference archives preserve their notes and use the same three
checks after the independent branch/archive commit check. A missing field or
mismatch refuses comparison and names the field. This is recorded provenance,
not a cryptographic attestation of a local file's author.

`TestCompareLocalProvenance` uses identical test names and passing verdicts for
all cases: another commit, another Node, another Go, missing Go or Node, missing
metadata, matching sidecar, matching notes, matching legacy build flags, and
inconsistent shard versions. Overlay mutants that independently remove the
commit, Go, and Node comparisons all fail this test. Sources, overlays and
outputs are in `cmd/adamic-gate/evidence/provenance-checks.tgz`.

The filtered oracle command on the requested final base fails to compile:
`internal/oracle/gcc_lane_test.go:84:15: undefined: leakSanitizer`. The missing
helper is inherited from `976ce7fa3`; this unit does not alter oracle tests.
The exact historical `47fbaf17` checkout is checked separately below. No full
15-shard gate or plain/sharded equivalence certificate is claimed.

The filtered oracle passed on the clean measured source `47fbaf174d16`, including
`TestNativeAgreesWithNode/internal/oracle/testdata/typeof_null.a`, with zero
native/Node cache hits. Commands (all output saved to files):

```sh
go test -race ./cmd/adamic-gate
go vet ./cmd/adamic-gate
go test -race ./cmd/adamic-gate -run '^TestCompareLocalProvenance$' -count=1
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^typeof_null[.]a$'
go test -overlay <commit|go|node>.json ./cmd/adamic-gate -run '^TestCompareLocalProvenance$' -count=1
```

The three provenance mutants fail specifically at `other-commit`, `other-go`,
and `other-node` (the Node mutant also accepts missing Node incorrectly).

Final validation: the full runner race suite passed, vet passed, and the final
focused provenance race suite passed. The three provenance mutants were killed
by the expected behavioral assertions, with no compilation failure. The final
branch contains the requested base's setup and stage1 changes; measurements
above remain attributed to their exact historical source and frozen runner.

### Four-slot follow-up: memory guard stops the experiment

The requested kill test was checked at 18:31 UTC on October 7 before changing
`layoutSlots`. The current execution cgroup reports `memory.peak=17179873280`
bytes against `memory.max=17179869184` bytes: approximately 16.00 GiB peak and
16.00 GiB limit. `memory.events` records `max=6618`, `oom=0`, and `oom_kill=0`.
The CPU quota remains `400000 100000` (four CPUs).

This is the cgroup's lifetime peak, not an isolated peak of the layout rerun.
The previous rerun did not reset/read the peak at its boundaries or measure max
RSS, so attributing this peak to layout would be unsupported. Assumption: the
instruction to stop near the box limit takes precedence over the subsequent
four-slot/split experiments; insufficient attributable headroom is not evidence
that increasing the memory guard is safe. The lifetime peak is already at the
limit, so the experiment stops here. No stage1 test-file change was made and no
separate stage1 commit is needed.

| Loop | Before | After | Instrument |
| --- | ---: | --- | --- |
| Joint layout group, two slots | 2421.176 s | Four slots not run: memory guard | Existing exact shard command and build-flags line above |
| Current cgroup lifetime memory peak | Not captured at rerun boundaries | 16.00 GiB / 16.00 GiB limit | Read `/sys/fs/cgroup/memory.peak`, `memory.max`, `memory.events` |

The 15-shard table above remains unchanged: layout shard 0 predicts 3456.456
seconds and measured 2421.176 seconds. It does not establish a sub-20-minute
gate. No four-slot wall drop or two-half budget is claimed. The four-slot claim
remains untested; a fresh isolated cgroup or attributable process-tree RSS
measurement is needed to establish headroom. Even the other shard predictions
are above 1200 seconds, so fitting layout alone would not certify the whole
gate under that budget. Raw readings are committed in
`cmd/adamic-gate/evidence/layout-memory-guard.json`.

### Final scope after the markdown mutant backend change

The requested two-half layout split is withdrawn. The final planner retains
one whole, pinned `layoutOnce` affinity containing all nine layout tests,
`formatterOnce`, and the whole oracle package's `gateOnce` affinity. The volume
agreement parent remains whole. `timings.json` remains derived from all fifteen
original 47fbaf17 shard logs, with the original audit and toolchain metadata.
The compare-provenance implementation is retained.

The measured pre-backend-change whole layout group at exact source 47fbaf17
was 2421.176 seconds. Its historical shard-derived planning estimate is
3456.456 seconds. These are explicitly pre-change values, not forecasts for
the new semantic-mutant backend. Once `stage1-format/markdown-mutants` at
`13f16800` reaches main, re-derive the markdownblocks timing units from logs of
that merged source. Do not transplant its timings into a plan for 47fbaf17.

Integration evidence supplied by the user, not measured by this worker:
`main eac1fb47` exceeded the 3600-second package timeout in both the whole gate
and a package-only retry on a five-core Linux box, recorded at
`gate-logs/eac1fb47cc64/manual/whole`. The user's measured `13f16800` change
moves semantic source mutants off native with no native-only survivor;
reported layoutOnce is 377 -> 89 seconds and the whole markdownblocks package
is 523 seconds with no skips. That package result is under 3600 seconds. Our
pre-change whole layout-group rerun is also under 3600 seconds; we did not
measure the rest of that package jointly with it. Neither its span nor the
failed main package run establishes a package-wide timeout result for 47fbaf17.

The abandoned split's half A passed fifteen mutant checks and four parents in
2022.108 test seconds, below 3600 but above 1200; setup was separately 43.848
seconds. Its end-of-run fresh cgroup peak was 3.383 GiB, with zero memory limit
or OOM events. Half B was cancelled on the user's instruction, so it has no
completed wall or end-of-run peak. No stage1 source file changed. The historical
split evidence remains for audit, but its plans are superseded and are not the
final planner behavior. The final report uses the whole-group plan above.

### Heaviest planned shards excluding markdownblocks

For the whole-affinity 15-shard plan, subtracting markdownblocks contributions
identifies the heaviest remaining work; this is a filtered view, not a new
packing allocation or a measured rerun. The full table retains those historical
contributions pending the backend change's merge and recalibration.

| Shard | Predicted seconds excluding markdownblocks |
| --- | ---: |
| 14 | 1731.075 |
| 11 | 1613.870 |
| 2 | 1613.870 |
| 8 | 1613.870 |
| 10 | 1613.870 |
| 5 | 1613.870 |
| 6 | 1613.870 |
| 9 | 1613.870 |
| 12 | 1613.870 |
| 4 | 1440.255 |
| 7 | 1421.250 |
| 13 | 1215.470 |
| 3 | 541.050 |
| 1 | 271.390 |
| 0 | 0.000 |

The remaining heaviest individual units are oracle `gateOnce` (1368.845 s,
conservative group price), lint `TestMutants` (1069.120 s), typeaware
`TestVolumeAgreementAndMutants` (731.755 s, reconstructed parent plus setup),
json `TestPortMatchesGoCohere` (686.000 s), typeaware
`TestTypeAwareAgreementAndMutants` (615.050 s), unicode
`TestCanonicalizeUnicodeNode` (571.080 s), and lint
`TestEmittedJavaScriptMismatch` (507.020 s). These measured/calibrated source47
units are inputs to a later isolated-heavy-package unit, not new measurements
of the backend change. Setup remains separate: the whole-group rerun used
84.939 s of warm setup; original shard 2 recorded 601 s. Setup was not rerun
on the other fourteen planned shards, so per-box setup there is unmeasured.

### Oracle Once measurement at 33935158

The provenance unit is already included in 33935158, implemented by
be217bbddd8bf627c94263ef89c771734d9b09af. Reverification command:
`go test -race ./cmd/adamic-gate -run '^TestCompareLocalProvenance$' -count=1`.
It passed (1.102 s), including identical green names/verdicts from another
commit, a different Node version, different/missing Go, missing metadata,
matching sidecar and notes, and legacy shard toolchain records. Existing
commit/Go/Node guard mutants and their caught tests remain in the committed
provenance proof archive. No duplicate implementation was added.

Oracle measurement is blocked at exact source
33935158c8126e795d90c044f04b35238afcd4ed. A compile-only probe and two full
package attempts all failed before running tests:
`internal/oracle/gcc_lane_test.go:84:15: undefined: leakSanitizer`.
The compile probe was `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
-run '^$' -count=1`; each full attempt was `ADAMIC_GATE_UNCACHED=1 go test
-count=1 -json -timeout 60m -parallel 5 ./internal/oracle`. Output was sent
directly to files. Therefore there are zero executed fixture/waiter samples
and no three-run median, not zero seconds. No oracle file was changed and
no compile workaround, omitted GCC test file, or substituted source was
used to claim a measurement of this commit.

| Step inside identity's gateOnce | Median seconds | Dependency from source inspection |
| --- | --- | --- |
| Cache-directory lookup | unavailable | Independent lookup |
| Node version query | unavailable | Independent subprocess |
| Sanitized runtime-library preparation | unavailable | Independent flag variant |
| Release runtime-library preparation | unavailable | Independent flag variant |
| Counted runtime-library preparation | unavailable | Independent flag variant |
| Harness/runner discovery and reads | unavailable | Independent of runtime variants |
| Repository path, stack limit, environment capture | unavailable | Independent of runtime variants |
| Context digest | unavailable | Depends on harness bytes and captured context |
| Each identity caller's wait | unavailable | All callers depend on Once completion |

These are dependency inferences, not measured concurrency results. The three
runtime preparations have distinct flag-keyed cache entries; they can be
logically independent, but concurrent compilation's memory/CPU cost remains
unmeasured. Context hashing depends on its captured inputs and is sequential
with respect to collecting them. Waiters receive the combined identity only
after all initialization finishes. The 1368.845-second historical gateOnce
planning price is a conservative whole-package grouping of measured test
observations; it is not an observed duration of the Once callback itself.
The requested profile is needed before attributing that floor to setup.

The available box reports nproc=5 and cpu.max="400000 100000" (four effective
CPUs). Toolchain: Go 1.27.1 linux/amd64, Clang 20.1.8, Node v24.19.0;
uncached oracle result mode. No fixture run completed, so no fixture load
before/after or before/after speedup is reported. Setup and failure logs are
preserved in the accompanying oracle measurement evidence archive.

Setup completed successfully in 88.205 s: Go ready 0.199 s, submodules
0.325 s, Clang 0.792 s, Node 1.014 s, module dependencies 27.890 s,
Go build 87.206 s (setup cumulative marks). Full setup build flags and
loads are in the evidence; these are setup intervals, not fixture timings.

### Completed oracle Once profile after the area repair

Measured source and pushed clean merge tip: `47b8b8d56a959540008574641b17b1c925a5c2fe`,
including area `08e0db2036acd710df716f4f6856191da3adb0be`. No oracle conflict
or source edit was resolved. Instrumentation used a Go overlay outside the
repository, archived with its generator, raw JSON logs and analysis.

Assumption: three complete sequential package runs on the same available box,
uncached oracle results, existing fully keyed warm native runtime archives;
normal opt-in defaults remain unchanged. This is not a cold archive build
benchmark or a new whole gate. All three runs passed with identical terminal
test names/verdicts: 1467 pass events (1466 tests plus package) and nine skips.
Each run made 3082 identity calls. All twelve callback steps have three samples.

| Callback step | Run 1 s | Run 2 s | Run 3 s | Median s |
| --- | ---: | ---: | ---: | ---: |
| cache-directory | 0.000001825 | 0.000002325 | 0.000002104 | 0.000002104 |
| node-version | 0.007769706 | 0.008578494 | 0.009520263 | 0.008578494 |
| runtime-0 | 0.087610034 | 0.024628872 | 0.022576758 | 0.024628872 |
| runtime-1 | 0.018757449 | 0.038917184 | 0.035150695 | 0.035150695 |
| runtime-2 | 0.020123608 | 0.018144502 | 0.030602013 | 0.020123608 |
| go-harness-glob | 0.000660947 | 0.000153187 | 0.000149978 | 0.000153187 |
| runner-glob | 0.000029483 | 0.000071184 | 0.000061326 | 0.000061326 |
| repository-context | 0.000006987 | 0.000008971 | 0.000008026 | 0.000008026 |
| stack-limit | 0.000001343 | 0.000001767 | 0.000001582 | 0.000001582 |
| environment-capture | 0.000012906 | 0.000032827 | 0.000011538 | 0.000012906 |
| harness-reads | 0.001079789 | 0.000630551 | 0.000818452 | 0.000818452 |
| context-digest | 0.000314435 | 0.000756495 | 0.000348848 | 0.000348848 |
| callback | 0.136552088 | 0.092104684 | 0.099624676 | 0.099624676 |

| Loop | Historical before s | Measured after median s | Instrument |
| --- | ---: | ---: | --- |
| Whole oracle group/package | 1368.845 planning price | 328.577567 command wall | `ADAMIC_GATE_UNCACHED=1 go test -overlay /workspace/gate-affinity-evidence/oracle-once-47b8b8d5/overlay.json -count=1 -json -timeout 60m -parallel 5 ./internal/oracle` |
| Once callback only | Not measured in historical logs | 0.099624676 | `time.Now` / `time.Since` around the original callback |

This before column is a different-source historical planning price, not a
controlled speedup. Package command walls include Go build and TestMain.
Callback medians do not add exactly to the sum of per-step medians. Logging
is excluded from individual step intervals but included in callback wall.
The callback median is about 0.0073% of the historical group price. It does
not explain a 1369-second package floor under these warm-build conditions.

| Longest completed top-level tests | Median recorded elapsed s |
| --- | ---: |
| TestCountsAreRecorded | 311.240 |
| TestLoopCountersAgreeWithNode | 79.310 |
| TestGCCLaneComparisonCatchesMutants | 9.550 |
| TestLibraryMapSetIteratorCopiesRefused | 2.510 |
| TestOct6InheritanceMutants | 2.290 |

Top-level elapsed values can overlap and include subtest scheduling; they
are not additive CPU time. Every identity caller has its individual per-run
call count, total/maximum interval, callback ownership and medians in
`identity-callers.csv` in the evidence archive. The maximum non-owner call
intervals were 0.000473279 s, 0.034029889 s, 0.000380643 s.
These measure time through sync.Once including scheduling, not a separately
instrumented mutex wait. Only one non-owner interval exceeded 1 ms across
all three runs.

The node query, cache-directory lookup and three flag-distinct runtime
preparations are logically independent. File globs precede their reads;
context digest depends on captured harness/environment/path/stack inputs.
All users wait for the combined identity. Concurrent cold-build CPU/memory
cost was not measured. No concurrency or cache change was made. The longest
observed parent was TestCountsAreRecorded; profiling its fixture work is a
separate decision from parallelizing this sub-second callback.

Build flags for every run: commit above, nproc=5, cpu.max="400000 100000",
Go="go version go1.27.1 linux/amd64", Clang="clang version 20.1.8
(https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)",
Node="v24.19.0", oracle results uncached, native runtime build cache warm.

| Run | Command wall s | Load before | Load after |
| --- | ---: | --- | --- |
| 1 | 328.577567 | 0.02 0.54 0.77 1/422 51457 | 5.97 3.86 2.16 1/440 71372 |
| 2 | 327.125017 | 5.97 3.86 2.16 1/440 71381 | 4.88 4.73 3.05 1/443 91244 |
| 3 | 361.592167 | 4.88 4.73 3.05 1/443 91252 | 5.29 5.11 3.77 1/448 111080 |

Raw evidence: `cmd/adamic-gate/evidence/oracle-once-47b8b8d5.tgz`. The prior
compile blocker section remains as historical evidence and is now resolved
by the area repair. No new cache or production check was introduced.

### Provision the checker archive on its owner only

Merged setup branch 77fc4632 via merge 7cf1320d, then implemented the runner
and fleet side in 9310d552. Plan.Archive identifies the sole consumer,
`github.com/system-inc/adamic/internal/native::TestSplitTSGoAgrees`, its shard,
and ADAMIC_CLANG_TSGO_ARCHIVE. Ordinary packing defers this consumer; after
packing, its owner is the lightest predicted shard including repeated parent
residuals (lowest index wins exact ties). For the measured-source 9310d552
22-shard plan, the owner is shard 10: 1027.150 s before the consumer,
1096.550 s with its 69.400 s historical unit price. This is a plan prediction,
not a new test timing.

Fleet obtains a declared plan from a plan box before launching workers. `fleet.sh briefs <full SHA>
<plan.json>` produces reviewable briefs without launching anything. The real
22-shard plan generated exactly one archive phase, in shard-10.md; all 22
briefs use --gate-inputs-no-archive. A plan box publishes plan.tgz on the
run's plan log branch. This unit still lets shard/merge enumerate for themselves;
the queued frozen-plan unit will consume that artifact directly.

Assumption resolving the setup API conflict: 77fc4632 deliberately rejects
--gate-inputs-no-archive and --gate-archive in one invocation. Furthermore,
archive-only setup unsets the other gate inputs. The owner therefore copies
the non-archive env.sh outside the checkout, runs --gate-archive separately,
saves its archive path, restores the copied env.sh, and exports that archive
path. Every other shard runs only the non-archive phase. WASI setup flags and
required boolean opt-ins remain provisioned. Record both setup phases
separately and sum them for the owner; no combined-flag command is required.

Before execution identity hashing or running tests, the runner inspects both
coverage units and Plan.Archive. An undeclared/misplaced consumer, an unset
archive variable, or a missing/nonregular file refuses by shard, test, and
variable name. A real shard-10 CLI invocation with the variable unset refused
before creating test evidence: `TestSplitTSGoAgrees refuses:
ADAMIC_CLANG_TSGO_ARCHIVE missing`. Merge's existing skip census remains the
independent post-run authority.

Proofs: focused race tests passed, full `go test -race ./cmd/adamic-gate
-count=1` passed (111.500 s), go vet passed, and Bash syntax/diff checks passed.
Three mutants were caught: drop archive preflight by
TestArchiveLightestAndMissingInput; drop parent residual by
TestArchiveCountsRepeatedParentSetup; archive every worker by
TestFleetTwentyTwoArchiveBriefs. The first test also mutates the plan's owner
onto a worker without an archive and requires the refusal to name the test.
No compiler or stage1 test was changed. No new cache was introduced. We did
not launch 22 boxes or remeasure cold setup costs; the setup branch's measured
costs and proofs remain its evidence, not this worker's measurements.
Raw plan, 22 briefs, CLI refusal, tests and mutant logs are in
cmd/adamic-gate/evidence/archive-shard.tgz.

### Freeze discovery once and feed heavy packages first

Implementation measurement commit: 39862fb6f7f09a6582b9fee38d8dbc71375b1493.
`plan` now writes a frozen identity along with coverage and package prices.
`shard -plan PATH` and `merge -plan PATH` consume it without repository-wide
Go discovery. Omitting -plan preserves self-planned execution. This is a
validated run input, not a test-result cache: ADAMIC_GATE_UNCACHED=1 still runs
all tests and all identity checks.

The identity includes the source commit, individual tracked file bytes and
modes (including nested pinned submodules), Go/Clang/Node versions and binary
identities, Clang resource contents, relevant Go environment, and answer-bearing
ADAMIC/Node/compiler/WASI inputs. External files/directories are hashed freshly;
external Git HEAD and dirty worktree bytes count, Git config/pack housekeeping
clearly does not. The checker archive identity is required on the owner and
merge, and ignored on nonowners. A changed file is refused by name before test
execution. Digest, coverage, affinity and price consistency are checked too.

Fleet always publishes plan.tgz on the run's plan log branch before launching
workers. Workers and merge fetch that exact artifact and validate against their
own checkout and inputs. Plan and merge boxes provision all inputs including
the checker archive for identity validation; among the 22 shard boxes only the
single declared archive owner builds it. These validation boxes' setup remains
separate cost. Local ADAMIC_GATE_PLAN bypass of plan publication was removed.

Package workers now receive descending planned package seconds; ties break by
name. Prices reuse affinity and repeated-parent accounting. Output aggregation
retains deterministic alphabetical order independently of execution order.
TestPackageFeedDescending catches an ascending-feed mutant. The suggested
1,996 to 1,776 s and 834 to 677 s are estimates from the historical logs, not
measurements here; a multi-package proving fleet remains necessary.

Six valid guard/order mutants failed their intended tests: omitted named source
check, tool identity, input bytes, Go environment, artifact digest, and descending
feed. Full runner race tests passed (86.639 s), followed by focused tests on the
final publication/mode changes (23.854 s), vet, Bash syntax and diff checks.
The actual repository source mutant appends a valid comment to main.go after
planning, leaving test names unchanged, and must refuse before a summary exists.

Cold-container discovery baseline at cb4ee1c4: preparation `go build ./...`
277.437 s; makePlan alone via `/workspace/adamic-gate-archive plan -count 512`
260.747 s. A toolexec tracer counted 50 link steps, 46 unique targets. The
chosen proof shard 380 runs only internal/flow; 49 links covered 45 unrelated
targets. This is a fresh 4-CPU/16-GiB container on the existing physical box,
not a fresh independent VM; module cache and installed tools already existed.
The Go cache started empty and was warmed by the preparation command. This
older-commit cold baseline is not a controlled speedup comparison.

Same-commit controlled proof: three sequential interleaved pairs on this cloud
box, nproc 5, cpu.max 400000 100000, Go 1.27.1, Clang 20.1.8 (LLVM
87f0227cb60147a26a1eeb4fb06e3b505e9c7261), Node v24.19.0. Results are
uncached (-count=1 and ADAMIC_GATE_UNCACHED=1); Go compile and keyed native build
caches are warm. GOFLAGS=-toolexec=/workspace/gate-affinity-evidence/plan-once/tool-trace.py.
GOCACHE=/workspace/gate-affinity-evidence/plan-once/fresh-go-cache. All source
and runner inputs are at 39862fb6; no competing build ran during a pair.

| Loop | Self-planned wall s | Frozen wall s | Saved s |
| --- | ---: | ---: | ---: |
| 1 | 128.680 | 115.447 | 13.233 |
| 2 | 126.408 | 116.710 | 9.698 |
| 3 | 116.477 | 68.938 | 47.539 |

Best of three: 116.477 vs 68.938 s. Median walls: 126.408 vs 115.447 s; median paired savings 13.233 s. Variance is substantial, so no 15/22-shard extrapolation is claimed.

Exact instrument: /workspace/adamic-gate-plan-once shard -count 512 -index 380
-scratch /workspace/gate-affinity-evidence/plan-once/scratch -out
/workspace/gate-affinity-evidence/plan-once/self-N; after command replaces
self-N with frozen-N and adds -plan
/workspace/gate-affinity-evidence/plan-once/frozen-plan.json. The 512-shard
proof isolates a small package and is not the production 15/22-shard layout.
The precomputed plan command took 67.456 s on this warm measurement checkout;
it runs once. Every measured shard passed 428 tests with zero skips/failures.
Each pair has byte-identical sorted projections of all 1,713 run, terminal,
pause and cont events; timestamps and diagnostic output are not test answers.
The actual changed-main.go mutant refused by filename before creating a summary.
The environment restart interrupted self-2; that partial attempt was excluded
and preserved under self-2-interrupted. Setup was not rerun during these pairs.

| Invocation | Load before | Load after |
| --- | --- | --- |
| self-1 | 1.41 0.88 1.27 1/532 133516 | 2.92 1.76 1.55 1/541 137812 |
| frozen-1 | 2.92 1.76 1.55 1/541 137812 | 1.74 1.64 1.53 1/545 141282 |
| self-2 | 2.14 2.09 1.73 1/556 145507 | 4.05 2.79 2.03 2/551 149736 |
| frozen-2 | 4.05 2.79 2.03 1/551 149736 | 2.24 2.57 2.04 2/555 153131 |
| self-3 | 2.24 2.57 2.04 2/555 153131 | 2.26 2.60 2.12 1/554 157350 |
| frozen-3 | 2.26 2.60 2.12 1/554 157350 | 4.05 2.96 2.27 2/551 160733 |

Raw evidence and reproduction scripts: cmd/adamic-gate/evidence/plan-once.tgz.

### Remove the cheap oracle pin after a partition proof

Candidate and all five oracle measurement checkouts:
179756a84dfcfa0a677692697ce470640a040724. No internal/oracle or stage1
source was edited. The measured 0.099625-s warm gateOnce callback is 0.0303%
of the prior 328.577567-s package command wall, so its WholePackage affinity
was removed conditionally. Both repeats of both partitions passed the condition.

Retained affinity declarations now include measured SetupSeconds and
SetupMeasurement next to the fixture: layoutOnce 448.889 s (47fbaf17,
runner 1e9192b6, cont through original-oracle agreement, fixture plus controls);
formatterOnce 206.546468 s (1abec4c7, first Go formatter command build on a
prepared box, formatter dependencies cold). The latter is a real fresh-box
cost beyond root-package warming, not its 1,342.480-s group price. The two
subsequent warm command builds took 7.789418 and 5.700614 s. No cost is
inferred for an unmeasured fixture.

Oracle pricing uses the three whole uncached logs at 47b8b8d5: median observed
unit elapsed proportions, scaled so the 515 observed planned units total the
measured 328.577567-s command wall. These prices are shares of a concurrent
whole wall, not measurements of individual isolated units. The 851 skipped or
unobserved planned units retain their original shard-log prices: this includes
844 WASI rows that the nine-skip profiling run did not execute. Repeating
parent residuals remains separately charged by the planner. The old @gateOnce
price of 1,368.845 s is removed; the old calibration and new override both
remain in timings.json.audit.json. A fully enabled WASI remeasurement is still
needed; assigning those unobserved rows zero would invent a saving.

The proof partitions use assignUnits and selections from the actual candidate
planner. They balance the observed input mode at predicted 164.287 and 164.290 s;
unexecuted WASI/other skip rows have zero proof weights only. This is separate
from production prices, which retain those rows. Each partition runs a single
Go test process with the planner's exact -run/-skip selectors. Complements
cover the parents' remaining children. Whole/partition execution is uncached,
-count=1 -parallel 5 -timeout 60m, with identical source, tools and inputs.

Fresh-box interpretation: no independent VM provisioning tool is available.
Five fresh Docker containers ran sequentially on one physical cloud box, each
nproc 5, cpu.max="400000 100000", memory.max=16 GiB. Each has an independent
cgroup, fresh /tmp and XDG fixture/runtime-library cache, and the same
preinstalled tool/input/Go-action-cache image. This proves the result for that
image, not uncached download/bootstrap costs on independent VMs. Go compile
cache is warm; oracle result caching is bypassed. Flags for every case:
Go="go version go1.27.1 linux/amd64", Clang="clang version 20.1.8
(https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)",
Node="v24.19.0", GOFLAGS empty, commit above; actual GOCACHE, XDG paths,
loads and argv are preserved per case.

| Case | Setup s | Tests s | Setup + tests s | Improvement vs whole | End-read fresh cgroup peak GiB |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole | 43.410 | 284.520 | 327.930 | reference | 3.271 |
| half-0-1 | 64.686 | 200.469 | 265.155 | 19.14% | 2.439 |
| half-1-1 | 65.059 | 126.912 | 191.971 | 41.46% | 3.266 |
| half-0-2 | 68.346 | 196.819 | 265.165 | 19.14% | 2.627 |
| half-1-2 | 92.132 | 149.484 | 241.616 | 26.32% | 3.430 |

Every one of the four boxes improves at least 2%, including its measured setup.
In each repeat the union exactly equals the whole's 1,466 pass and nine skip
names/verdicts, including duplicate parent verdict agreement. No fail, OOM,
memory limit event or dirty checkout occurred. Each partition is below both
1,200 s of tests and the 3,600-s package timeout. The whole also fits both.
Half 0's remaining longest parent is TestCountsAreRecorded, 175.320 / 175.180 s;
half 1's TestLoopCountersAgreeWithNode is 58.770 / 105.750 s. They are later
isolated-unit inputs; the normalized plan prices are not those elapsed spans.

Exact instrument: python3 cmd/adamic-gate/evidence/oracle-partitions/boxes.py
(after extracting the evidence archive). worker.py invokes bash cloud/setup.sh,
then either go test -count=1 -json -timeout 60m -parallel 5
 github.com/system-inc/adamic/internal/oracle, or the exact Go argv in
halves.json[index].Commands and each case's metadata.json. The full generated
selectors are retained there instead of shortening them into an inaccurate
command. Whole, A1, B1, A2, B2 run sequentially. The user-requested two repeats
are used; the whole setup-plus-test loop exceeds five minutes.

Initial setup failed before tests twice: offline curl exit 7 reaching the
package proxy, then curl exit 5 because Docker could not resolve proxy.
The host's configured proxy installed the checksum-verified stage3 dependency
image once; that preparation took 68.386 s, recorded separately in
image-preparation.log. The final five cases use the same installed image and
work offline. Failed attempts are excluded, preserved as failure evidence.

| Case | Setup load before | Setup load after | Test load before | Test load after |
| --- | --- | --- | --- | --- |
| whole | 7.15 2.25 1.41 2/563 6 | 7.77 3.14 1.75 1/565 947 | 7.77 3.14 1.75 1/565 947 | 5.13 4.48 2.70 1/570 20898 |
| half-0-1 | 4.72 4.41 2.68 1/566 6 | 7.30 5.18 3.06 1/568 1212 | 7.30 5.18 3.06 1/568 1212 | 5.01 5.03 3.41 1/567 14812 |
| half-1-1 | 4.69 4.97 3.39 1/564 6 | 5.54 5.13 3.56 1/566 1214 | 5.54 5.13 3.56 1/566 1214 | 4.76 4.94 3.68 1/573 7825 |
| half-0-2 | 4.46 4.87 3.66 1/570 6 | 4.88 4.91 3.76 1/574 1225 | 4.88 4.91 3.76 1/574 1225 | 5.74 5.04 4.02 1/570 14827 |
| half-1-2 | 5.36 4.97 4.00 1/569 6 | 7.43 5.56 4.29 1/569 1223 | 7.43 5.56 4.29 1/569 1223 | 4.67 4.97 4.25 1/575 7833 |

The setup encounter also exposed an implicit input of frozen plans: installed
stage3/api/node_modules is outside tracked source. Frozen identities now hash
that tree alongside its pinned tracked lock. TestFrozenPlanRefusalAndEquivalentEvents
changes an ignored installed API file and requires a named refusal. A mutant
omitting that input fails this test. Restoring the oracle pin fails
TestOracleMeasuredTestsMaySpread. Full runner race tests passed (73.261 s),
go vet passed, Bash syntax and diff checks passed. Earlier six frozen/order
mutants remain independently caught. No new result cache was introduced.

### Updated 47fbaf17 fifteen-shard prediction

Reassigned the preserved 47fbaf174d168f14cd50dedcaabfaa09c51cdefd discovery
inventory with the actual candidate planner, new prices, two expensive groups,
and descending package feed. This is a prediction artifact, not a runnable
frozen plan for a different checkout. The archive consumer is on shard 14,
the lightest predicted shard before adding its measured historical cost.

| Shard | Predicted tests s | Actual tests s | Setup s |
| ---: | ---: | ---: | ---: |
| 0 | 3456.456 | 2421.176 historical | 84.939 historical |
| 1 | 1633.810 | unmeasured | unmeasured |
| 2 | 1633.809 | unmeasured | unmeasured |
| 3 | 1633.805 | unmeasured | unmeasured |
| 4 | 1633.803 | unmeasured | unmeasured |
| 5 | 1633.808 | unmeasured | unmeasured |
| 6 | 1633.810 | unmeasured | unmeasured |
| 7 | 1633.804 | unmeasured | unmeasured |
| 8 | 1633.801 | unmeasured | unmeasured |
| 9 | 1633.803 | unmeasured | unmeasured |
| 10 | 1633.808 | unmeasured | unmeasured |
| 11 | 1633.805 | unmeasured | unmeasured |
| 12 | 1633.809 | unmeasured | unmeasured |
| 13 | 1633.808 | unmeasured | unmeasured |
| 14 | 761.770 | unmeasured | unmeasured |

Shard 0's complete 11-unit selection equals the earlier measurement; its
actual/setup columns are historical, not a new proving run. Its build flags:
source 47fbaf17, runner 1e9192b6, nproc 5, cpu.max 400000 100000, Go 1.27.1,
Clang 20.1.8, Node v24.19.0, uncached tests, warm Go/runtime build caches,
GOFLAGS empty; load before "0.50 2.15 3.88 2/261 28575", after
"1.80 1.97 2.06 1/262 31786". Original per-unit observation flags are in the
calibration audit, new oracle override flags above and in that audit.

The 1,200-s full-shard test target is not established: predicted shards 1-13
remain about 1,633.8 s, and layout's historical 2,421.176-s joint span remains.
The markdown semantic-mutant owner change at 13f16800 must reach the target
before those markdown prices can be re-derived. Excluding markdownblocks,
shards 2, 4-13 remain about 1,633.8 s by prediction; no actual proving fleet
has run with longest-first feed. Heaviest non-markdown historical measured
units remain lint/TestMutants 1069.120, typeaware/TestVolumeAgreementAndMutants
731.755, json/TestPortMatchesGoCohere 686.000, typeaware/TestTypeAwareAgreementAndMutants
615.050, markdowninline/TestMarkdownInline 588.250, unicodeproperties/TestCanonicalizeUnicodeNode
571.080, and lint/TestEmittedJavaScriptMismatch 507.020 s. Those are the inputs
to later heavy-package box measurements; their exact source/build flags are
preserved in the original audit, not replaced by new oracle share prices.

Raw evidence: cmd/adamic-gate/evidence/oracle-partitions.tgz; replan JSON,
selectors, whole and union verdict files, all setup/test logs, metadata,
calibration overrides and both guard mutants are included. No internal/oracle,
protected compiler or stage1 test file was edited.
