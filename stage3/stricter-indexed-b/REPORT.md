Built 27 isolated project-.ts fixture templates and a Node/native/JavaScript witness harness; 5 sites proven.
Base 390af985; latest preceding commit 390af985caf2a81edb239cdf5e3238add50a7fcd; branch codex/stricter-indexed-b only.
Logged filtered go tests: 5 proven, 0 blocked, 22 remaining; final gate recorded below.
Every proven site has an emitted-C erase-panic mutant that builds under sanitizers and loses the pinned exit-70 observation.
Whole-program compilation, sparse holes, typed arrays and records are not covered; refused shapes are never counted as proven.

## Scope and assumptions

The exact 27 rows D195 through D221 are retained in sites.json from a1a16427
stage3/ledger/checker-259/rows.csv. All receivers are arrays. Read shapes were
traced through the referenced cohere/TypeScript submodule source at
tsc/testdata/fixtures/compiler/transformers/generators.ts; no implementation
was copied from cohere. The fixtures are independently written minimal shapes.
Statement and declaration arrays use small object elements; clause and CodeBlock
arrays preserve discriminated object unions. readonly array receivers are preserved.
Label is number; labelNumbers is number[][]; blockOffsets is number[].

D196/D197 share variables[i]. D200 uses one caseBlock.clauses[i] read;
D201-D205 and D207/D208 share another such read. D213/D214 share blockStack[i].
These are ledger diagnostic rows, not a claim of 27 unique upstream reads.
Each row has an isolated witness so a prior trap cannot hide its observation.
Minimal witnesses stop at the read, before downstream property uses, allowing
source Node to observe undefined rather than throw on a later dereference.

The specific .ts witness requirement supersedes the general new-.a convention.
Following the base worker, tracked .ts.txt templates materialize as main.ts in
a temporary project with its own tsconfig. noUncheckedIndexedAccess is omitted
(project default false); strictNullChecks and strictFunctionTypes stay enabled.
Each source is run directly through Node 24 type stripping, outside Adamic.

Present input prints present; an empty receiver with the same numeric index prints
undefined on source Node. Both have empty stderr and exit 0. Native release and
ASan/UBSan and the JavaScript backend must match present input. For absent input
all backends must produce empty stdout, exit 70, and exactly
`adamic: panic: indexed read is absent: <main.ts>:<read line>:<read column>\n`.
The expected location is calculated independently from source text, and the
same location must appear in the inserted guard. The actual --explain-checks CLI
must list exactly one checked indexed-presence site, count one, and trust zero.

## Site results

| Row | Upstream line | Source expression | Receiver | Result |
|---|---:|---|---|---|
| D195 | 1283 | `statements[i]` | object array | proven, erase-panic caught |
| D196 | 1370 | `variable.initializer` | object array | proven, erase-panic caught |
| D197 | 1374 | `transformInitializedVariable(variable)` | object array | proven, erase-panic caught |
| D198 | 1684 | `initializer.declarations` | object array | proven, erase-panic caught |
| D199 | 1729 | `initializer.declarations` | object array | proven, erase-panic caught |
| D200 | 1866 | `clause.kind` | object union array | remaining |
| D201 | 1880 | `clause.kind` | object union array | remaining |
| D202 | 1881 | `clause.expression` | object union array | remaining |
| D203 | 1881 | `clause.expression` | object union array | remaining |
| D204 | 1887 | `clause.expression` | object union array | remaining |
| D205 | 1887 | `clause.expression` | object union array | remaining |
| D206 | 1889 | `clauseLabels[i]` | number[] | remaining |
| D207 | 1889 | `clause.expression` | object union array | remaining |
| D208 | 1889 | `clause.expression` | object union array | remaining |
| D209 | 1911 | `clauseLabels[defaultClauseIndex]` | number[] | remaining |
| D210 | 1918 | `clauseLabels[i]` | number[] | remaining |
| D211 | 1919 | `caseBlock.clauses` | object union array | remaining |
| D212 | 2446 | `supportsLabeledBreakOrContinue(containingBlock)` | object union array | remaining |
| D213 | 2469 | `supportsLabeledBreakOrContinue(block)` | object union array | remaining |
| D214 | 2472 | `supportsUnlabeledBreak(block)` | object union array | remaining |
| D215 | 2480 | `supportsUnlabeledBreak(block)` | object union array | remaining |
| D216 | 2499 | `supportsUnlabeledContinue(block)` | object union array | remaining |
| D217 | 2507 | `supportsUnlabeledContinue(block)` | object union array | remaining |
| D218 | 2880 | `withBlock.expression` | object array | remaining |
| D219 | 2951 | `labelNumbers[labelNumber]` | number[][] | remaining |
| D220 | 2983 | `blockOffsets![blockIndex]` | number[] | remaining |
| D221 | 2984 | `block` | object union array | remaining |

## Commands, setup and limits

All test output is written to log files, never piped. Current proof logs:

- `/tmp/stricter-indexed-b-group1-final.log`

Setup: `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh`
with `/workspace/adamic-tools/env.sh` sourced for builds and tests. nproc=5.

```text
setup: go ready (0.086s)
setup: node ready (0.090s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.465s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.691s
setup: markdown dependencies ready (0.895s)
setup: submodules ready (158.967s)
setup: go build ready (412.702s)
setup: test binaries deferred (use --warm-tests) (412.817s)
setup: build cache warm (412.818s)
setup: build-flags commit=390af985caf2a81edb239cdf5e3238add50a7fcd nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.08 0.02 0.01 1/137 712 load-after=14.24 8.47 3.55 3/195 3896
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (412.842s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.G56QUY
```

The first test attempt preceded submodule checkout and failed with missing
cohere/TypeScript/tsc/go.mod. It was rerun after checkout completed.
No compiler implementation files are edited. Unsupported receivers are
reported with the observed refusal; this unit does not extend them.
The full repository gate is not claimed. The final focused gate and vet
results are appended after the last group.
