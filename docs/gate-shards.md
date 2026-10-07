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
reported rather than treated as a controlled performance comparison. Proof/prediction metadata
and the 8/10/12 plan tables follow below after the clean-commit plans are generated.
