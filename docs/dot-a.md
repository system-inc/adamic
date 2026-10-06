# Adamic source is .a

Every new Adamic source file uses `.a`. `.ts` is TypeScript that has not passed the
gate. The integration worker chooses the rename moment immediately after
integration 16. This preparation branch changes no existing source filename.

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

The inventory includes stage 1 ports and their gap fixtures, the spec's compile
and refusal fixtures, oracle fixtures, benchmarks, dedication and examples.
Refusal and gap fixtures move as language test inputs even though their purpose
is to be rejected. Existing `.a` files stay in place. The TypeScript parser finds
static imports, re-exports, dynamic imports and import types in both extensions;
ordinary source strings and comments are preserved. Go fixture paths, globs and parsed corpus predicates accepting both extensions,
document links and commands, JSON patterns and tooling references are updated
against the rename inventory. No source import is found with a regular expression.
On current main both JSON configurations use directory patterns already covering
`.a`; `tsconfig.json` also already declares `sourceExtensions: [".a"]`.

The files intentionally kept as TypeScript are:

| File | Reason |
| --- | --- |
| `internal/load/prelude.d.ts` | Ambient declarations for the checker, not an executable program |
| `cmd/adamic-meter/testdata/adapt/main.ts` | Pre-gate input for source adaptation |
| `cmd/adamic-meter/testdata/adapt/types.ts` | Types belonging to that adaptation input |
| `cmd/adamic-meter/testdata/corpus/not_yet.ts` | Input used to measure unsupported TypeScript |
| `cmd/adamic-meter/testdata/corpus/refused.ts` | Input used to measure rejected TypeScript |
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

## Recorded validation

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
