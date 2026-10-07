# Adamic source is .a

Every new Adamic source file uses `.a`. `.ts` is TypeScript that has not passed the
gate. The integration worker runs the rename on fresh main after the stage 3 stack,
with the cohere reformat as two commits under one uncached gate. This preparation branch changes no existing source filename.

From the repository root, preview the complete plan:

```sh
go run ./cmd/adamic-rename-dot-a .
```

At the integration worker's chosen moment, apply the rename with one command:

```sh
go run ./cmd/adamic-rename-dot-a --apply .
```

Stage and commit the resulting renames and reference edits together. A second
invocation is empty even before staging. The command reads tracked and unignored
untracked files, classifies their purpose, and refuses an unknown `.ts` file, a
destination collision or unparseable source before writing anything. It skips
missing tracked paths after a rename. Review the printed plan before applying.
The apply is not a repository-wide atomic transaction; an I/O failure can leave
partial changes, so use a clean checkout and inspect any error before retrying.

The inventory includes stage 3 authored programs and fixtures, stage 1 ports and their gap fixtures, the spec's compile
and refusal fixtures, oracle fixtures, benchmarks, dedication and examples.
Refusal and gap fixtures move as language test inputs even though their purpose
is to be rejected. Existing `.a` files stay in place. The TypeScript parser finds
static imports, re-exports, dynamic imports and import types in both extensions;
ordinary source strings and comments are preserved. Go fixture paths, globs and parsed corpus predicates accepting both extensions,
document links and commands, JSON patterns, fixture lists and tooling references are updated
against the rename inventory. Status fixture paths and diagnostic paths are rewritten;
recorded Node stdout/stderr and upstream `tsc` provenance are preserved as literal
data. Bare test filenames are restricted to their suite. The upstream tsc driver
corpus is not parsed or rewritten: its documented selection prohibits relative
imports and deliberately includes diagnostic inputs outside Adamic.
No source import is found with a regular expression.
On current main the root JSON configurations already cover `.a`. The mixed
stage 1 type-aware fixture config retains TypeScript includes, adds `.a` includes
and source-extension metadata, and seeds stock tsc config validation with the
unchanged prelude declaration; its callers then supply manifest roots.

The files intentionally kept as TypeScript are:

| File | Reason |
| --- | --- |
| `internal/load/prelude.d.ts` | Ambient declarations for the checker, not an executable program |
| `cmd/adamic-meter/testdata/adapt/main.ts` | Pre-gate input for source adaptation |
| `cmd/adamic-meter/testdata/adapt/types.ts` | Types belonging to that adaptation input |
| `cmd/adamic-meter/testdata/corpus/not_yet.ts` | Input used to measure unsupported TypeScript |
| `cmd/adamic-meter/testdata/corpus/refused.ts` | Input used to measure rejected TypeScript |
| `cmd/adamic-meter/testdata/optional/main.ts` | Pre-gate input for optional-return adaptation |
| `bridge/tsgo/testdata/sample.ts` | Upstream tsc checker input for the bridge |
| `stage3/drivers/tsc/corpus/**` | Upstream single-file diagnostic inputs, often intentionally invalid, with no relative imports; existing `.a` names stay unchanged |
| `cohere/**`, including `cohere/TypeScript/**` | Upstream submodules, not this repository's Adamic source |

Node tooling stays `.mjs`. `.gitattributes` marks `.a` as text with normal diffs
and GitHub's TypeScript language highlighting. The malformed UTF-8 input fixtures
retain their existing binary-byte attributes.

## Running source on Node 24

Node 24 supports synchronous `registerHooks`. Run a source entry directly with:

```sh
node --disable-warning=ExperimentalWarning --import ./oracle/register-dot-a.mjs main.a argument
```

The same registration module is imported by the existing wrapper:

```sh
node --disable-warning=ExperimentalWarning ./oracle/node.mjs main.a argument
```

Both paths load `.a` and transitional `.ts` imports as ES modules, strip types with
Node's `stripTypeScriptTypes` in transform mode, and resolve `adamic` to the
oracle runtime. The runtime is loaded even when a program has no `adamic` import,
so uncaught errors and stream failures keep the oracle's exit contract. Direct
loading already has the right argv; the wrapper removes its own argv entry.
Stage 1's existing parity and speed runs call `node.mjs`, so they now share this
loader without changing their timing protocol. This executes original source,
independently of Adamic lowering. Node execution does not certify the types.

## In-flight branches

Branches adding new `.ts` Adamic files must rename them at merge with
`go run ./cmd/adamic-rename-dot-a --apply .` and commit the generated reference edits.

## Stage 3 stack landing refresh

Main advanced during verification to `71d7e49`, the pinned stage 3 stack merge.
This branch merges it with `b06f366`. The current committed
[dry run](../cloud/dot-a-dry-run.txt) contains **337 files, 1,875 reference rewrites
in 346 files**. The additional 88 renamed files are stage 1 lint registry code.
The exact integration command remains `go run ./cmd/adamic-rename-dot-a --apply .`.
The merge was conflict-free and retains the shared loader's transform mode.
The registry's deliberately stale `.ts` ambiguity witness is kept as TypeScript;
a parsed Go call distinguishes that negative input from the real module path.
Its generated, ignored `.generated/registry.ts` is pre-gate TypeScript emitted by
Go, and is intentionally still generated with that suffix; the shared loader
supports it. The real registry-owned rule modules move to `.a`.

The fresh current-main scratch checkout is `/tmp/dot-a-stack-main`. Its apply,
idempotence, build and vet completed with exit 0. The literal command printed
`TOTAL files=337 references=1876 reference_files=346` before protecting the new
stale witness, and its immediate repeat was empty. The final stale-witness fix
removes one reference edit, giving 1,875; a final fresh apply is recorded below.
The whole uncached oracle, stage 3 and all stage 1 tests are running.
Results below for `b6b1538` are earlier evidence, not a claim that this new
main revision has completed its gate.

## October 7 refresh

The branch merges current main `b6b1538` with merge commit `877621e`.
The refreshed [dry run](../cloud/dot-a-dry-run.txt) records 249 renames and
1,464 reference rewrites in 263 files. This is the inventory at that main commit;
run the same command on fresh main to include any later in-flight additions.
Stage 3's current authored sources and oracle fixtures already use `.a`.
Of the renamed files, 226 are stage 1 ports/fixtures, 16 are spec fixtures,
six are benchmarks and one is the fresh-analysis regex fixture.

The stage 3 runner guard now accepts the shared loader after checking transform
mode, accepts an inline runner already in transform mode, and retains the older
inline-runner conversion. The merged branch's `go test -count=1 -timeout 30m
./stage3/fixtures` passed, with output in `/tmp/dot-a-refresh-branch-stage3.log`.

The final scratch checkout is `/tmp/dot-a-final-proof`, cloned at main
`b6b1538`. Preparation assets were copied in: the command, parser module entry,
shared loader/wrapper, oracle cache identity and stage 3 runner guard. No renamed
source from scratch is committed. Applying the final binary printed
`TOTAL files=249 references=1464 reference_files=263`; its immediate repeat printed
`TOTAL files=0 references=0 reference_files=0` before staging.

Scratch build, vet, load and lower passed with exit 0. The complete uncached oracle and
stage 3 fixtures passed without a fixture filter. The first build retry hit
temporary storage exhaustion; obsolete scratch copies were removed and build
was rerun successfully. The first combined stage 1 run passed 15 test packages, but type-aware
config validation failed and Markdown had regeneration/dependency failures and
a package timeout. The mixed config retains `.ts`, includes `.a`, and seeds config
validation with the unchanged prelude declaration before callers replace roots.
Formatter scratch inputs remain pre-gate TypeScript `.ts`; generated outputs
use `.a`. Type-aware and all 20 Markdown tests are being rerun, the latter in
three disjoint groups, each with the 30-minute limit. Final results will be added
when those runs finish. The final scratch commands are:

```sh
go build ./... > /tmp/dot-a-final-proof-build-rerun.log 2>&1
go vet ./... > /tmp/dot-a-final-proof-vet-rerun.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -p 2 ./internal/oracle ./stage3/fixtures ./internal/load ./internal/lower ./stage1/cohere/typeaware ./cmd/adamic-rename-dot-a > /tmp/dot-a-final-proof-core.log 2>&1
go test -count=1 -timeout 30m ./stage1/cohere/typeaware > /tmp/dot-a-final-proof-typeaware-rerun.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks -run '^(TestMarkdown(ASTPreprocessing|ParserPrefixes|UnicodeWidths|SourceDecoding|TextSplitting)|TestMicromarkInputChunks|TestFrontMatterStage|TestOptionalStringInitializationWitness|TestParserRepresentationProbes|TestNativeBuildModesAreDistinct|TestWholeDocumentOraclePreflight)$' > /tmp/dot-a-final-proof-markdown-core-rerun.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks -run '^TestMarkdown(ListLayout|QuoteLayout|TableLayout|CodeBlockLayout|HTMLBlockLayout)$' > /tmp/dot-a-final-proof-markdown-layout-rerun.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks -run '^TestMarkdown(WhitespaceLayout|LeafComposition|RootLayout|StructureLayout)$' > /tmp/dot-a-final-proof-markdown-composition.log 2>&1
```

The other 15 tested stage 1 packages passed in the combined scratch run, logged
in `/tmp/dot-a-release-gate.log`; three generator packages have no test files.
That run used `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -p 2
./internal/oracle ./stage3/fixtures ./stage1/... ./cmd/adamic-rename-dot-a`.
The final config seed was added after the core command completed, so its log
also contains the superseded type-aware failure; the dedicated rerun above is
the validation of the final config.

The standalone codemod tests and both Node loader tests passed. CSS numbers again
ran end to end after rename: direct `--import`, the wrapper and native each exited
0, printed `0.1e2kHz "1.000px"` plus newline and had empty stderr. Both output
comparisons passed. Build/run evidence is in `/tmp/dot-a-refresh-cssnumbers-build.log`
and `/tmp/dot-a-refresh-{direct,wrapper,native}.{out,err}`.

Toolchain setup completed in 118 seconds and `nproc` printed 5:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (118s)
setup: done in 118s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Sixteen new mutants were caught, each by a test failure with exit 1:

| Mutant | What caught it |
| --- | --- |
| Omit stage 3 classification | `TestStage3PurposeAndStatus`: authored source is unclassified |
| Omit status references | Same test: fixture and diagnostic paths remain stale |
| Parse the upstream diagnostic corpus | Same test: intentionally invalid upstream input is parsed |
| Rename the bridge's upstream input | Same test: extra rename is rejected |
| Rewrite recorded Node observations | Same test: extra golden edits are rejected |
| Allow cross-suite bare filenames | Same test: negative path case is rewritten |
| Rewrite upstream tsc citations | Same test: upstream document acquires an unwanted edit |
| Omit parent-relative and snapshot references | `TestPlanAndRerun`: relocation/profile references remain stale |
| Keep generator output extensions `.ts` | `TestGeneratedAdamicNames`: parsed output names remain stale |
| Accept overlapping edits | `TestOverlappingEditsRefuseBeforeWriting`: unsafe plan is accepted |
| Remove the shared loader's transform mode | Stage 3 `TestFixtures`: runner guard rejects the loader |
| Rename pre-gate formatter scratch inputs | `TestGeneratedAdamicNames`: temporary TypeScript input is changed |
| Omit mixed-config handling | `TestMixedTypeScriptConfig`: `.a` include is absent |
| Omit the custom `.a` source extension | Same test: source-extension metadata is absent |
| Omit the prelude validation seed | Same test: stock tsc config has no seed input |
| Rewrite the deliberately stale TypeScript module | `TestRegistryRenameWitness`: ambiguity witness acquires an unwanted edit |

Mutant logs are `/tmp/dot-a-refresh-mutant-*.log`. An earlier attempt that removed
only stage 3's suffix fallback survived: that fixture's local path already
resolved without the fallback. The targeted status-omission mutant above was
then caught. Earlier scratch runs exposed accidental changes to Node goldens
and a negative path case. Those were fixed, their regressions and mutants were
added, and superseded scratch runs were replaced by the final run.

The full `./...` test gate, opt-in throughput/profile and external-library tests,
cohere's upcoming reformat, non-Linux platforms and
I/O-failure injection are outside this verification. The whole oracle and all
ordinary stage 1 tests are included across the commands above.

Pinned ordinary width-test dependencies were installed under
`/tmp/adamic-markdown-width`: emoji-regex 10.6.0, get-east-asian-width 1.6.0
and narrow-emojis 0.0.3. Install output is `/tmp/dot-a-width-dependencies.log`.

## Original preparation validation

The full dry run against main `5d4c801` is in
[cloud/dot-a-dry-run.txt](../cloud/dot-a-dry-run.txt): 126 files, 592 reference
edits in 119 files. The scratch apply is in `/tmp/adamic-dot-a-scratch`; none of
its renamed files is included in this preparation commit. Its immediate second
run reports `TOTAL files=0 references=0 reference_files=0`.

Toolchain setup completed successfully:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (75s)
setup: done in 75s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.
Setup's sanitizer overflow probe failed as required.

The loader is included in the oracle cache identity: a changed registration
module invalidates prior observations even when the wrapper stays the same.

The scratch checks ran with output sent directly to logs:

```sh
go vet ./... > /tmp/dot-a-scratch-vet.log 2>&1
go test -count=1 -timeout 30m -p 2 ./cmd/adamic-rename-dot-a ./internal/load ./internal/lower ./stage1/... > /tmp/dot-a-scratch-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestSharedLoaderInvalidatesOracleEvidence|TestNativeAgreesWithNode/internal/load/testdata/0.1/compile' > /tmp/dot-a-scratch-oracle.log 2>&1
node --test oracle/register-dot-a.test.mjs > /tmp/dot-a-scratch-loader.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/dot-a-typescript go test -v -count=1 -timeout 30m -p 2 ./cmd/adamic-rename-dot-a ./stage1/cohere/lint ./stage1/typescript/parser ./stage1/typescript/scanner > /tmp/dot-a-scratch-filter-packages.log 2>&1
```

Vet exited 0 on both the preparation branch and the scratch apply. The oracle exited 0: all ten renamed spec entries passed against
Node, both compiler backends, sanitized native and release native, with the leak
check on every entry that finishes. The cache regression also passed. That run
was uncached: native hits 0, misses 29; Node hits 0, misses 20. The loader tests
passed 2, failed 0. The standalone codemod tests passed 3, failed 0.

The broad package run exited 1 on the scanner's empty corpus after the rename.
Load, lower, all eleven cohere packages and the TypeScript parser passed. After
fixing the Go corpus predicates, the affected-package rerun above exited 0:
codemod, lint, parser and scanner all passed. Thus every one of the thirteen
stage 1 packages passed against the scratch apply. Lint's compiler/stage1 corpus
check also passed with the pinned corpus enabled. The parser compared 77 whole
compiler files (28,836,875 identical expression-tree bytes and 44,766,682
identical whole-tree bytes); the scanner checked 77 compiler files, 104 stage 1
files and 18,236 generated inputs. Optional throughput/profile tests requiring
environment flags were skipped.

In scratch, CSS numbers was also compiled with
`go run ./cmd/adamic build stage1/cohere/cssnumbers/main.a -o /tmp/dot-a-cssnumbers`.
Direct Node `--import`, the wrapper and the native binary each exited 0, had empty
stderr, and printed identical output for `/tmp/dot-a-example.css`:

```text
0.1e2kHz "1.000px"
```

The input was `.1000E+002KHZ "1.000px"` followed by a newline. Both `cmp` checks
passed. The package parity test additionally checks the Go implementation,
source on Node and both Adamic backends, sanitizers, leaks and speed runs.

All fifteen preparation mutants were isolated from the preparation branch:

| Mutant | What caught it |
| --- | --- |
| Skip static import edits | `TestPlanAndRerun`: source differs from expected parsed-import rename |
| Accept a destination collision | `TestRefuseCollisionAndUnknownPurpose`: unsafe plan accepted |
| Accept an unclassified TypeScript file | Same test: unsafe plan accepted |
| Ignore syntax diagnostics | Same test: malformed source accepted |
| Omit non-source reference edits | `TestPlanAndRerun`: stale document link |
| Refuse missing tracked files after apply | `TestPlanAndRerun`: repeat run cannot read the old path |
| Treat `stage1/cohere` references as upstream | `TestPlanAndRerun`: stale port document reference |
| Change a plain TypeScript test-data glob | `TestPlanAndRerun`: TypeScript data glob changed |
| Keep corpus filters limited to `.ts` | `TestSourceCorpusFilters`: executed predicate rejects `.a` |
| Remove `.a` handling | Loader source test: exit/output assertion |
| Stop resolving `adamic` | Loader source test: exit/output assertion |
| Remove transform mode | Loader source test: enum input cannot run |
| Keep the wrapper's argv entry | Loader source test: argument output differs |
| Omit runtime preload | Loader error test: uncaught error exits 1 instead of 70 |
| Omit the loader from the cache hash | `TestSharedLoaderInvalidatesOracleEvidence`: identity does not change |

Each mutant run exited 1 from a test assertion, not a compiler diagnostic.
Go logs are `/tmp/dot-a-mutant-*.log`; Node logs are
`/tmp/dot-a-loader-mutant-*.log`; the cache mutant log is
`/tmp/dot-a-cache-mutant.log`. The cache mutant used a Go overlay, leaving the
working source untouched. Codemod tests also preserve source comments and
ordinary strings, handle escaped import specifiers, check already-existing `.a`
imports and JSON/glob references, and keep TypeScript data globs unchanged.

The full repository gate, external Prettier integration and non-Linux platforms
were not run. No timing improvement is claimed. The command's handling of I/O
failure midway through apply was not fault-injected. The first scratch package
run was stopped after correcting the reference classifier and replaced by the
final run recorded here. The first complete package run then caught the scanner's
`.ts`-only corpus filter (`empty corpus: ../../../stage1`). The codemod now widens
these parsed Go predicates, keeping `.ts` for upstream inputs and adding `.a` for
Adamic. Its additional four edits touch scanner, parser and lint corpus walkers.
Those packages were rerun with the pinned v6.0.3 compiler corpus enabled; their
separate log is `/tmp/dot-a-scratch-filter-packages.log`.
