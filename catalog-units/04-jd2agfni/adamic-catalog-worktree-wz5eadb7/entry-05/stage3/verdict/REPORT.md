Built: stage3/verdict/run.sh runs acceptance, tiny and a faithful counted upstream CLI subset.
Commits: base ef3141e9; first runnable 60722e70 pushed at 03:47 MDT; expanded evidence is on codex/stage3-verdict-harness.
Observed full proof: A 301/301 acceptance, 1/1 tiny and 6262/6262 baselines; zero deferred.
Mutants: B 300/301, 0/1, 6261/6262; C 0/301, 0/1, 0/6262; all proof assertions pass.
Not covered: native linking and 6182 explicitly excluded upstream input files; virtual hosts, variants, emit and unsupported directives.

The first runnable proof used:

```sh
source /workspace/adamic-tools/env.sh
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-adapted
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove.py --baseline-limit 40 /tmp/stage3-verdict-proof-early > /tmp/stage3-verdict-proof-early.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/test_verdict.py > /tmp/stage3-verdict-checks.log 2>&1
```

Both exit 0. A, B and C summary JSON and markdown and the proof log are in
`evidence/early/`. B changes `e` to `E` in the first diagnostic of `argument.ts` for
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


## Expanded coverage and the faithful CLI boundary

The size bound was removed. A complete 6264-case coverage probe reported
acceptance 301/301, tiny 1/1, and baselines 6262/6264. Its JSON and markdown are
retained in `evidence/coverage-probe/`. Observed: diagnostic codes and messages
agree, while library paths, locations and ordering differ. Inference: these are
harness-formatting scope differences, not compiler behavior differences.
Upstream `harnessIO.ts:iterateErrorBaseline` rewrites library diagnostic
locations as `(--,--)`. The census's existing outside-source rule recognized
only numeric locations and missed these two library diagnostics. The corrected
rule recognizes both forms and excludes exactly:

- `tests/cases/conformance/types/members/duplicateNumericIndexers.ts`
- `tests/cases/conformance/types/members/objectTypeHidingMembersOfExtendedObject.ts`

Both have errors in the default library, whose upstream hidden positions and
virtual paths the current CLI subset cannot reproduce. Every diagnostic outside
the single source is now counted under the same explicit scope rule. These are
input-based exclusions using reference structure, not matching the tested CLI
against expected text to decide which cases to keep. The updated census has
6262 cases (3582 compiler, 2680 conformance) and 6182 exclusions, totaling all
12444 input files. The standalone files have 2967 clean and 3295 diagnostic
expectations. Option variants are counted as excluded input files, not expanded
configurations. The complete census was independently regenerated from a
separate pristine pinned checkout and matched byte for byte.

The baseline summary also removes its exact case scratch-directory prefix,
reproducing upstream `util.ts:removeTestPathPrefixes` for the virtual `/.src/`
root. Raw stdout remains untouched in each capture. A filename-byte mutant and
an unrelated-root mutant prove this rule does not hide other path differences.
The raw observation that led to this formatting rule is retained in
`evidence/root-format-observation.txt.gz`.

## Focused guards and their mutants

`test_verdict.py` runs five focused tests. Every mutation is observed to fail
its designated assertion or comparison:

| Mutant | Catch |
|---|---|
| stdout `error` becomes `Error` | stdout bytes only |
| empty stderr gains a byte | stderr bytes only |
| exit 2 becomes 1 | exit bytes only |
| quoted filename changes one byte | baseline diagnostic bytes after exact root mapping |
| unrelated root replaces the case root | baseline diagnostic bytes; no broad path stripping |
| census pin changes | invalid baseline census |
| selected count changes | invalid baseline census |
| excluded count changes | invalid baseline census |
| excluded row duplicates a selected source | invalid baseline census |
| outside-source regex only recognizes numeric locations | library-placeholder assertion fails |
| source gains a newline | source SHA256 rejects before compiler execution |
| baseline gains a newline | reference SHA256 rejects before that compiler execution |

The source and reference mutations ran through `baseline_suite` on the disposable
cached checkout and were restored in finally blocks. Logs are
`evidence/comparison-checks.log.gz` and `evidence/pin-mutants.log.gz`.
No oracle fixtures or count rows were added. New-output refusal was checked:
a second command at an existing output path exits 2 and preserves its evidence.

## Complete stand-in proof

The final census was measured without a baseline limit:

```sh
source /workspace/adamic-tools/env.sh
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-adapted
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove.py /tmp/stage3-verdict-proof-ready > /tmp/stage3-verdict-proof-ready.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/test_verdict.py > /tmp/stage3-verdict-checks-final.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/audit_mutants.py /tmp/stage3-verdict-auto-upstream/upstream > /tmp/stage3-verdict-pin-mutants.log 2>&1
```

All three proof/check commands exit 0. The verdict commands themselves exit
0 for A and 1 for B and C. No harness errors occurred.

| Stand-in | Acceptance pass/total | Tiny pass/total | Baselines pass/total |
|---|---:|---:|---:|
| A: stock adapted Node CLI | 301/301 | 1/1 | 6262/6262 |
| B: one diagnostic byte per suite | 300/301 | 0/1 | 6261/6262 |
| C: exit 1, empty output | 0/301 | 0/1 | 0/6262 |

Every baseline run counts 6182 exclusions and zero deferred cases. Acceptance
includes tiny. Final JSON, markdown and compressed logs are in `evidence/`;
earlier bounded evidence remains in `evidence/early/`. Raw A/B stdout, stderr
and exit witnesses are in `evidence/mutants/`: exactly one stdout byte differs
for each suite, with identical stderr and exit. The five focused tests pass,
and both disposable artifact hash mutants are caught and restored. Shell
syntax, Python AST parsing and git diff whitespace checks pass. No full gate
or whole Go package tests were run.
