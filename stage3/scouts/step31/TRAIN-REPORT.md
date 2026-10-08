Built: merged the named views train and rebuilt all three Node observers/comparators plus train and main Adamic compilers.
Commits: scout parent 2de9fc1b; train 1e81b051; merge 49183f57; delivery SHA is in the final dispatch.
Observed: binder/checker/emitter each match stock Node over 301 projects and retain their pre-train golden bytes.
Mutants: three binder bodies, checker assignment, emitter initializer, three process streams and four evidence fields are caught.
Not covered: native runtime parity, ownership/leaks, step32 linking, 6262-case rerun or a new own-file meter; train-only successes remain pending.

The requested origin/cloud/land-train-4-views-v3-787cea7a-6c28ada8 was fetched
explicitly and merged into codex/step31-scout without conflicts. The checkout's
normal fetch refspec tracks only main, so the first ordinary fetch did not
populate the train ref. No branch was substituted. The remote-head snapshot
contains the views train families through train 4, including the named branch.

Exact pins:

- Main: 031a1259bc7973934792dc6cb1bd4074fc2204b9.
- Named train: 1e81b051c52eee40d47c44f4a5f68b821e606708.
- Merge used for build: 49183f57f8ac273665be143c4a525bff0896a54a.
- Scout parent: 2de9fc1b01d3da3beb35c554e70d98cd6d9fe536.
- TypeScript: 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
- Cohere on both main and train: 7945d102a6c18dd36adf9114a758ce646e8b2359.

Main is an ancestor of the train. The build tree includes all queued changes
from the explicitly named tip. No new compiler implementation or adaptation
was introduced by this scout; changes outside scout territory came through the
authorized merge. The Node drivers themselves did not need repair. All scout
JS/Python/shell files were syntax-checked and loaded afresh. apply.sh produced
a fresh adapted tree, not a reused prior output. The adaptation/API inputs are
identical between fetched main and the named train, as verified by git diff.

Node confirmation, exactly 300 selected acceptance projects plus the three-file
tiny project (301 projects total):

| Piece | Projects compared | Golden bytes | SHA256 | Stock driver seconds |
|---|---:|---:|---|---:|
| binder | 301 | 3475070 | 3f83fbf1a9c1587e2923484f7a22911f21c3ac47382ba80dbca597af04d846aa | not timed |
| checker | 301 | 333703 | f5b3ccb280ac67ed516240b52b476a9748dd8a60a516cec86f76be371265828a | 4.609707645 |
| emitter | 301 | 222338 | c25c363bc358f8b5a5ba4d06172eb89b4b1cb8128a9af2157df2c9370bd6eebd | 4.785898885000001 |

Each real adapted-source observer matches the independent stock bundle with
empty stderr and exit zero. Checker and emitter were compared through the
component-compare.py request/hash validation path; binder used compare.py.
Their comparison processes took 7.616s and 7.851s, respectively, including the
source import/transpilation and comparator overhead. The two stock driver times
above include request loading, observations and cache writes, but not process
startup. Binder was not separately timed in this rerun.

Checker/emitter acceptance goldens are byte-identical to the first 301 projects
of the pre-train full goldens committed in 2de9fc1b. Binder retains the initial
scout's acceptance hash. These are baseline-compatible Node confirmations,
not a new native pass or a train-only success. No new fixture was added, so
counts.md remains unchanged; all existing binder/component fixture assertions
passed on the fresh tree.

Native component build admission used freshly gathered, byte-audited closures:

| Component entries | Code declarations | Code files | Copied spans | Evaluation modules |
|---|---:|---:|---:|---:|
| binder.bindSourceFile | 1407 | 21 | 1492 | 79 |
| checker.createTypeChecker | 3390 | 39 | 3485 | 79 |
| emitter.emitFiles and createPrinter | 2573 | 35 | 2664 | 79 |

--no-adapt retains each gathered declaration's exact adapted-source bytes.
Original evaluation imports and compiler barrel initialization are preserved.
The native roots import these gathered closures. They are component admission
roots, not implemented native JSONL drivers; no component can yet reach that
runtime proof. Both raw slice-entry.a and a separate native-entry.a were built.
The latter adds only `import type {} from "node:util"` before the slice import,
following the scanner driver's documented declaration-loader activation.
The exact locked @types/node 25.3.3 package was installed under stage3/api.
It supplies the real ErrorConstructor declarations, not a synthesized shim.

First native stops after loading Node declarations, mapped through audited
slice spans back to original adapted source:

| Piece | Train first stop | Main first stop | Result |
|---|---|---|---|
| binder | corePublic.ts:14:5, MapLike index signature refused; use a Map | same | blocked |
| checker | checker.ts:12034:30, TS18048, file possibly undefined | same | blocked |
| emitter | moduleNameResolver.ts:2772:21, TS2322, SearchResult object does not fit SearchResult<Resolved> | same | blocked |

Slice locations are corePublic.ts:9:5, checker.ts:10894:30 and
moduleNameResolver.ts:1549:21. The raw binder/emitter roots first stop at
Debug.fail's Error.captureStackTrace, slice debug.ts:113:19 (original :200:19),
TS2339 because the Node declaration loader is not activated. The raw checker
already has an earlier optional-file error and also contains later diagnostics.
Every main/train/profile attempt exits 1; no binary was produced. Full captured
stderr is retained, not reduced to the headline. Main and train have identical
first-stop strings and exits on every component/profile pair. Binder with Node
declarations reaches lowering; checker/emitter remain blocked by the checker.
No clang or native differential run is claimed.

The current-main compiler was independently built in a detached worktree at
031a1259. Both pins use the same cohere submodule; the worktree shares that
source read-only through a symlink. Go VCS stamping cannot inspect that shared
submodule path, so this baseline build uses -buildvcs=false and records the
source SHA and binary hash separately. The train binary's build info records
49183f57 and vcs.modified=false. Compiler hashes are in summary.json.

Train-only build success must remain pending. native-stops.py labels any exit-0
build pending and native_differential_run=false; exit-1 builds are blocked.
This run found no native build success on either pin. A component cannot become
a native pass from Node observations, a compiled import root or queued views
coverage alone. The first stops did not advance past main's boundaries.

Commands actually run, all test output sent to logs:

```sh
git fetch origin
git fetch origin refs/heads/cloud/land-train-4-views-v3-787cea7a-6c28ada8:refs/remotes/origin/cloud/land-train-4-views-v3-787cea7a-6c28ada8
git merge --no-edit origin/cloud/land-train-4-views-v3-787cea7a-6c28ada8
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-train-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/step31-train-adapted > /tmp/step31-train-apply.log 2>&1
export STEP31_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
python3 stage3/scouts/step31/prepare-corpus.py /tmp/step31-train-node/projects.json
node stage3/scouts/step31/binder-dump.cjs /tmp/step31-train-node/projects.json
node stage3/scouts/step31/checker-dump.cjs /tmp/step31-train-node/projects.json /tmp/step31-train-node/checker
node stage3/scouts/step31/emitter-dump.cjs /tmp/step31-train-node/projects.json /tmp/step31-train-node/emitter
go build -o /tmp/step31-train-adamic ./cmd/adamic
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
python3 stage3/scouts/step31/native-stops.py /tmp/step31-train-adamic train /tmp step31-train /tmp/step31-train-native-final
python3 stage3/scouts/step31/native-stops.py /tmp/step31-main-adamic main /tmp step31-train /tmp/step31-train-main-stops
python3 stage3/scouts/step31/audit-train.py stage3/scouts/step31/evidence/train-tip
```

The recorded Node comparator command.json files give the exact source invocation,
and the two native report.json files give all twelve build commands. Closure
commands were `SLICE_TYPESCRIPT=$STEP31_TYPESCRIPT bash stage3/slice/run.sh TREE
NEW_SLICE <entries above> --no-adapt`, followed by verify.cjs for each slice.
Existing focused fixtures/mutants were run with test.cjs, test-components.cjs,
check-compare.py, mutants.py and component-mutants.py on the fresh train tree.
No whole-package tests, full gate, 6262-case rerun or meter/census run was used
as confirmation. No PR was opened.

Setup succeeded: Node ready 0.013s, Go 0.023s, submodules 0.046s, markdown
0.053s, clang 0.111s, Go build 41.374s, cache warm 41.452s, done 41.476s.
nproc=5, cpu.max=400000 100000, 17.6 GB; Go 1.27.1, clang 20.1.8, Node
24.19.0. The environment is /workspace/adamic-tools/env.sh.

The main build/native capture initially exhausted the separate 8.8 GB /tmp
filesystem. Verbose Go output proves `no space left on device`; this was an
infrastructure failure, not a compiler first stop. Removed only 722,014,842
bytes of named, regeneratable previous scout dumps whose compressed evidence
and hashes were already committed. Main Go temporaries moved to
/workspace/scratch/step31-go-build via GOTMPDIR; its final build succeeded.
Native measurements were completed afterward. The interrupted attempt is not
counted as a component result.

Every mutant rerun and catcher:

| Mutant | Observed catcher |
|---|---|
| binder function flags become block-scoped flags | stock golden bytes, aliases/scopes differ, Node exit 0/empty stderr |
| binder block variable flags become function-scoped flags | scopes bytes differ, Node exit 0/empty stderr |
| binder declaration flags cleared | all fixture files differ, Node exit 0/empty stderr |
| checker assignment elaboration returns true | TS2322 disappears, stock diagnostic bytes differ, Node exit 0/empty stderr |
| emitter omits variable initializer | emitted text differs, Node exit 0/empty stderr |
| comparator stdout changes one byte | only stdout comparison fails, comparison exit 1 |
| comparator stderr gains one byte | only stderr comparison fails, comparison exit 1 |
| comparator exit changes 0 to 1 | only exit comparison fails, comparison exit 1 |
| recorded native first-stop string changes | captured-stderr recount rejects it, audit exit 1 |
| recorded native exit changes 1 to 0 | captured-exit recount rejects it, audit exit 1 |
| native status is falsely promoted to pass | pending/blocked status audit rejects it, audit exit 1 |
| recorded Node population rises from 301 to 302 | request population recount rejects it, audit exit 1 |

Native harness ports, linking, ownership, leaks and lower failures beyond these
first stops remain unobserved. Existing language questions for @system_adamic
remain undecided in REPORT.md.
