Built: one verdict entry point, exact acceptance/tiny checks, and a counted upstream CLI subset.
Base: ef3141e9; TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8; early runnable commit is recorded in branch history.
Observed early: A 301/301 acceptance, 1/1 tiny, 40/40 explicitly limited baselines; 5865 eligible baselines deferred.
Mutants: B changes one diagnostic byte in one case per suite; results 300/301, 0/1, 39/40; C passes zero everywhere.
Not covered early: native compiler, full baseline measurement, virtual hosts, variants, emit, syntax/API differences and excluded directives.

The first runnable proof used:

```sh
source /workspace/adamic-tools/env.sh
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-adapted
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove.py --baseline-limit 40 /tmp/stage3-verdict-proof-early > /tmp/stage3-verdict-proof-early.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/test_verdict.py > /tmp/stage3-verdict-checks.log 2>&1
```

Both exit 0. A, B and C summary JSON and markdown and the proof log are in
`evidence/`. B changes `e` to `E` in the first diagnostic of `argument.ts` for
acceptance/tiny and `ArrowFunctionExpression1.ts` for baselines. Each introduced
failure is stdout only; the proof compared raw A/B captures and required exactly
one differing byte, equal lengths, unchanged stderr and exit. C exits 1 with no
output and fails every case, including clean projects. Acceptance includes tiny.
The focused comparison tests independently change stdout, stderr and exit and
are caught only by the corresponding byte comparison; EOF offsets are checked.

The initial census selects 5905 of 12444 inputs (3434 compiler, 2471 conformance)
and excludes 6539. Every source has one selected/excluded row in selection.json.
The early proof is an explicitly limited smoke measurement, not 5905 passes.
A separate one-command run with no upstream environment override successfully
cloned the cached pin into its new output and rejected C in all 301 acceptance
projects, tiny, and one smoke baseline. Shell syntax, Python AST parsing and
`git diff --check` passed. No Adamic fixtures or counts were added. No full gate
or whole-package test was run.

Setup succeeded with GOPROXY='https://proxy.golang.org|direct'. Timing lines:
Go 0.067s, Node 0.067s, clang 0.461s, markdown dependencies 2.092s,
submodules 23.896s, Go build 227.748s, cache warm 227.885s, done 227.917s.
nproc=5; cpu.max=400000 100000. The environment file is
/workspace/adamic-tools/env.sh. The adapted CLI was built only after apply
finished: npm ci succeeded, and npx hereby tsc --no-typecheck completed in 1s.
This is a runnable Node compiler, not adapted-source typechecking or native
linking evidence. Artifact hashes are in evidence/provenance.json.
