Rebuilt the project-aware loader and types union on main for step 32 and #acxgkff.
The five original commits 16c22268 through 88fe8de4 were replayed on c7090c58; delivery SHA is in the handoff.
Focused loader tests and Node/native/JavaScript reference oracles pass; executable fixtures print 7.
Eleven original mutants and two main Node-loader interaction mutants fail their intended assertions.
Whole native tsc acceptance remains for the combined stricter-options branch; no full gate was run.

The load.go conflict retains main's Node package pin and star-export diagnostic
formatting alongside project-owned options, transitive source roots, per-edge
compatibility checks and ambient types union. Standalone Node loading keeps its
complete checking roots, including the embedded prelude, instead of reverting
to only execution roots. Project Node imports retain their project's roots and
declarations. source_fs.go retains both Node and project console modes. The
counts conflict keeps all current main rows and adds the two reference fixtures;
counts are then regenerated. No protected lowering/native file was edited.

A stable-tree setup rerun was needed. The first attempt overlapped unresolved
cherry-pick markers and failed with syntax errors in load.go/source_fs.go. The
next found an incorrectly scoped host variable left by the resolution. After
fixing it, setup passed: Go 0.022s, Node 0.025s, submodules 0.064s, markdown
0.080s, clang 0.187s, build 43.780s, cache 43.913s, done 43.940s.
The printed environment is /workspace/adamic-tools/env.sh. nproc is 5, quota 4.

Counts checking then exposed missing prelude declarations in standalone Node
programs. TestProjectLoaderStandaloneNodeKeepsPrelude covers imports from both
node:fs and adamic. A mutant rebuilding from execution roots produces TS2307
for adamic and fails that witness. TestProjectLoaderNodeImportsKeepReferenceRoots
covers the opposite boundary: a mutant forcing project imports through the
standalone Node path discards an unimported dependency and fails the root check.
The types-union clean, duplicate and declaration-skipping witnesses now have
separate top-level tests. Their seconds and all imported loader leaf timings
are in evidence/loader-results.json; every leaf is below 60 seconds.

Commands run with output redirected:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/projects-main-setup-pass.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
npm install --prefix /tmp/projects-stock --ignore-scripts --no-audit --no-fund typescript@6.0.3
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences|TestNodeLibrary' -count=1 -timeout 10m -json > /tmp/projects-main-loader.jsonl 2> /tmp/projects-main-loader.stderr
PROJECT_LOADER_TSC=/tmp/projects-stock/node_modules/typescript/lib/tsc.js python3 stage3/project-references-source/verify_types.py > /tmp/projects-main-types.log 2>&1
PROJECT_LOADER_TSC=/tmp/projects-stock/node_modules/typescript/lib/tsc.js python3 stage3/project-references-source/verify.py > /tmp/projects-main-references.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/project-references-source/' -count=1 -timeout 30m > /tmp/projects-main-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/projects-main-counts.log 2>&1
go test -overlay /tmp/projects-main-node-overlay.json ./internal/load -run '^TestProjectLoaderNodeImportsKeepReferenceRoots$' -count=1 > /tmp/projects-main-node-mutant.log 2>&1
go test -overlay /tmp/projects-main-prelude-overlay.json ./internal/load -run '^TestProjectLoaderStandaloneNodeKeepsPrelude$' -count=1 > /tmp/projects-main-prelude-mutant.log 2>&1
```

The two overlay commands exit 1 with test assertions, not build failures. The
verifiers independently run and catch lost transitive references, declaration
redirects (TS6305), allowed option conflicts, dropped reference audit findings,
duplicate roots, physical prelude substitution, cycles, one-sided types in
production and audit, and skipped union declaration validation in production
and audit. The clean ambient fixture and two-/three-project executable profiles
match stock TypeScript 6.0.3 on Node and both Adamic backends. The duplicate union
matches both TS2451 diagnostics. Authored .a fixtures all pass standalone checks.

Before each candidate push, the committed branch runs integration's exact lane
command from the repository root. Its result is retained in evidence/lane.log.
This candidate contains only the rebuilt loader topic and its main interaction
repairs. It does not import the stricter-options topic.

Counts refresh passes in 53.766s. It only restores row order around the two new
fixtures; no numerical value changes. New Node interaction leaves take 0.77s
and 0.76s. The split clean, duplicate, and skipped-declaration union leaves take
4.76s, 4.72s, and 4.87s. Main movement after the initial branch point is merged
before delivery; it contains only other workers' test changes.
