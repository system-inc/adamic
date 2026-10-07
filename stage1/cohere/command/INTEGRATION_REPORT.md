Integration repair on codex/stage1-command

Reproduced both reported failures after fetching and merging origin/main ef3d907
into the existing branch (2573b3b). The clean merge commit is 821e40a. The cohere
submodule did not change. No TypeScript behavior or expected Go answers changed.

Observed before the fix:

- TestOriginalCommandBoundaries failed: native printed solutions `.` and
  `tsconfig.app.json`; Go kept `tsconfig.app.json` as a running TypeScript project
  and classified only `.` as a solution. Node source and the JavaScript backend
  agreed with Go.
- TestLoadersMatchGoCohere failed at canonical line 895: native printed a project
  reference instead of TypeScript's `test/async/api.bench.ts` file. Node source and
  the JavaScript backend agreed with Go.
- TestRuntimeFieldLayoutsAreIncluded failed: directory.c's `path` at slot 1 was
  absent from the whole-program field-offset proof.

Cause: main's uniform field-access optimization counted generated objects but
omitted the realPath runtime result introduced on this branch. Source objects can
put `path` in slot 0. The optimizer then read the realPath result's `kind` instead
of its `path`; visited-directory deduplication suppressed subsequent walks.
Registering the runtime layout makes this conflict retain checked field lookup.
Also registered runProcess's runtime layout, which the same audit requires.
The audit now checks 25 runtime layouts. None of the four protected compiler
files was edited.

Added TestRealPathFieldConflictMatchesNode: a source object puts `path` in slot 0,
and an actual realPath call returns its path in slot 1. Source Node supplies the
expected bytes through the existing Node oracle runtime. Native normal and
ASan/UBSan builds must agree; Linux LeakSanitizer is enabled on sanitized runs.
No oracle catalog fixture was added, so recorded fixture counts are unchanged.

Mutation evidence, using Go overlays without changing the production working tree:

- Omit the realPath layout: the original Go ownership comparison fails with the
  exact reported extra solution. The compiled mutant executes cleanly, with
  empty stderr; only the canonical output comparison rejects it.
- The same omission is independently caught by the new Node fixture: normal
  native exits successfully and prints `literal\nOk\n`, whereas Node prints
  `literal\n<resolved directory>\n`. No compiler failure counts as a kill.
- Omit the process-result layout: TestRuntimeFieldLayoutsAreIncluded rejects
  process.c's unregistered `output` at slot 0. This is a layout audit mutant,
  not an executable process-output mutant.

Reproduction commands, each after sourcing /workspace/adamic-tools/env.sh:

```sh
bash cloud/setup.sh > /tmp/stage1-integration-setup.log 2>&1
go test -count=1 -timeout 15m -v ./stage1/cohere/command -run '^TestOriginalCommandBoundaries$' > /tmp/stage1-integration-ownership-before.log 2>&1
go test -count=1 -timeout 15m -v ./stage1/cohere/config -run '^TestLoadersMatchGoCohere$' > /tmp/stage1-integration-loader-before.log 2>&1
go test -count=1 ./internal/native -run '^TestRuntimeFieldLayoutsAreIncluded$' > /tmp/integration-layout-before.log 2>&1
```

The two parity tests failed in 12.477 s and 96.004 s respectively. Setup succeeded:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (116s)
setup: done in 116s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc = 5.

Validation and mutation commands:

```sh
go test -count=1 -timeout 20m -v ./stage1/cohere/config > /tmp/stage1-integration-config-after.log 2>&1
go test -count=1 -timeout 20m -v ./stage1/cohere/command > /tmp/stage1-integration-command-after.log 2>&1
go test -count=1 -timeout 15m ./internal/native > /tmp/stage1-integration-native.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m ./internal/oracle -run '^TestCountsAreRecorded$' > /tmp/stage1-integration-oracle.log 2>&1
go vet ./... > /tmp/stage1-integration-vet.log 2>&1
go test -count=1 -v ./internal/native -run '^TestRealPathFieldConflictMatchesNode$' > /tmp/stage1-integration-node-after.log 2>&1
go test -count=1 -timeout 15m -overlay=/tmp/stage1-integration-realpath-overlay.json -v ./stage1/cohere/command -run '^TestOriginalCommandBoundaries$' > /tmp/stage1-integration-realpath-mutant.log 2>&1
go test -count=1 -overlay=/tmp/stage1-integration-realpath-overlay.json -v ./internal/native -run '^TestRealPathFieldConflictMatchesNode$' > /tmp/stage1-integration-node-mutant.log 2>&1
go test -count=1 -overlay=/tmp/stage1-integration-process-overlay.json -v ./internal/native -run '^TestRuntimeFieldLayoutsAreIncluded$' > /tmp/stage1-integration-process-layout-mutant.log 2>&1
```

A full repository test gate was not run. The command layer's existing engine,
resolver, scheduler and signal-forwarding gaps remain as recorded in GAPS.md.
This fix restores the two reported integration comparisons and accounts for the
branch's runtime layouts in main's optimizer; it does not close those gaps.

Results observed on the merged tree:

- Full config package PASS 142.231 s: 143 settings, 130 tsconfigs, 1,368 source
  files; discovery census 132 roots, 637 projects, 288 glob queries. All six
  existing mutants were caught: directory pruning, double-star zero segments,
  nested repository boundary, strict settings JSON comments, inherited rule
  options, tsconfig excludes. Each executes cleanly before comparison fails.
- Native package PASS 87.170 s. Focused Node regression PASS 0.359 s on its final
  source; normal and sanitized native output agrees with source Node.
- Uncached oracle/counts filter PASS 16.226 s.
- Repository go vet, changed-file gofmt and git diff --check report no findings.
- Full command package PASS 236.526 s: original ownership/command comparisons,
  36 actual CLI cases, 74 argument cases, graph/process Go boundaries and eight
  independent Node process cases. All nine existing executable mutants were
  caught: root marker precedence, dot from subdirectory, scope value copy,
  nearest owner tie, bare boolean flag, reverse import direction, closure limit,
  child verdict, and native child status replacement. Their checks require clean
  execution before an output disagreement counts.
