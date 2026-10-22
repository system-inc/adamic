# 91: discard patcher callback results

Conditions 1 through 5 pass; only the two arrow bodies change.
The void-widening source scan falls from 5 rows to 1, with no new finding.
Control and with-91 each have 106,366 passing, one sanctioned API failure, zero pending.
The one-site reversion restores two scan rows and leaves one transpile difference.
The emitted public API and the complete API baseline diff are byte-identical.

Branch: `typescript/stage3-adapt91-patcher-discard`, cut from fetched main
`f9d7464755c8de94d581b6b3fda9ac39af6e7c62`. Task: #wks9251.
The upstream source is TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. The parser and language service are
stock TypeScript 6.0.3 from stage3/api's locked install.

**91 intentionally changes emitted JavaScript.** A patcher's truthy return no
longer stops the loop; each callback result is discarded. This follows the
language lead's October 20 ruling on #wks9251. No declaration or type changes.
No public API line changes. The ruling's reason is that a void-typed value is
not a value an accepted Adamic program can read.

## Exact scope and public API

[scope.json](proofs/scope.json), [source.diff](proofs/source.diff), and
[transpile.diff](proofs/transpile.diff) prove the complete change across all
746 source files. Both comments-retained and comments-removed transpilation
have only these two arrow-body differences:

| Site | Before | After |
| --- | --- | --- |
| factory/nodeFactory.ts:1164 | `forEach(nodeFactoryPatchers, fn => fn(factory));` | `forEach(nodeFactoryPatchers, fn => { fn(factory); });` |
| utilities.ts:8582 | `forEach(objectAllocatorPatchers, fn => fn(objectAllocator));` | `forEach(objectAllocatorPatchers, fn => { fn(objectAllocator); });` |

The utilities callback takes `ObjectAllocator`, rather than `NodeFactory`.
The adapter resolves both callback signatures and requires exactly `void`,
checks their parameter types and the compiler's `forEach` declaration, and
validates both targets before writing either file. Complete before/after source
hashes reject moved or otherwise changed targets. Rerunning makes zero edits.
The actual built and tested tree passes the same source-scope and idempotence
checks after the suites. Full `stage3/apply.sh` passed; its incremental row is
2 files, 2 lines added, 2 removed. The repository's patch table is untouched.

[api.json](proofs/api.json) and [oracle-comparison.json](proofs/oracle-comparison.json)
record byte-identical `built/local/typescript.d.ts` with and without 91:
590,861 bytes, SHA-256
`477a56fb7b35a5f268be6131441c67d237b4646a9f8d2e1f50dc163879450051`.
A real scratch declaration artifact with an added declaration fails the same
byte checker. Appending an unrelated source comment fails the scope checker;
appending a newline to a target fails the adapter's movement guard.

## Registration references

[references.cjs](proofs/references.cjs) parses all 39,349 TypeScript and JavaScript
program files under src/, tests/, and scripts/, including harness, test inputs,
baselines, .cts and .mts. It uses an iterative AST inventory, then stock
`LanguageService.findReferences` with 711 source roots and any test candidates.
This is not a text search. [references.json](proofs/references.json) records
only each definition, with no caller or value alias:

- `addNodeFactoryPatcher`: factory/nodeFactory.ts:481.
- `addObjectAllocatorPatcher`: utilities.ts:8574.

A scratch test-runner module importing and calling `addNodeFactoryPatcher`
fails the reference check. Its call reference is retained in
[registrant-mutant.json](proofs/registrant-mutant.json). This is a claim about
the adapted tree; these internal registration functions remain callable by the
separate JavaScript witnesses.

The initial recursive inventory exhausted Node's stack on a deeply nested test
input. It was replaced with an iterative AST walk. The complete inventory and
the registrant mutant both passed their expected checks on the final tool.
The initial failure log is preserved.

## JavaScript witnesses

[hooks.cjs](proofs/hooks.cjs) is copied byte-for-byte from
`typescript/stage3-void-widening`, commit
`175f0de29e9b2334023c8b5550a6eeb3df8e4eee`. Its SHA-256 is
`a2df41c33e68c4d9e5361dc6ddbdc851d0eb50f25cfb8b9697c7965eea70fa57`.
It registers a factory callback returning 42 and an allocator callback returning
17. On the control's actual built compiler, each first callback runs once and
each second callback is skipped. Its original early-stop assertion fails on
the with-91 compiler, as expected.

[hooks-check.cjs](proofs/hooks-check.cjs) separately checks both actual built
compilers: `[first, second]` is `[1, 0]` before and `[1, 1]` after for each
registry. Allocator counters are reset after registration's immediate calls.
[hooks-before.json](proofs/hooks-before.json) and
[hooks-after.json](proofs/hooks-after.json) retain the Node observations.

## Refusal scan and required mutant

[make-scan.py](proofs/make-scan.py) and [extract.py](proofs/extract.py) are copied
unchanged from the same census commit. The original README's guarded latent
overlay is rebuilt against this branch's main compiler. The production
`refuseVoidWidening` is unchanged. Each run retains the no-output loader and IR
guards, examines 79 compiler files, exits 0, and has no panic.

This is explicitly a **void-widening-only-source-scan**, measured on a
checker-rejected program. It performs no lowering or concrete generic
instantiation and makes no full latent census claim.

[scan-delta.json](proofs/scan-delta.json) records 5 to 1, exactly four removed
rows: nodeFactory.ts:1164 columns 34 and 40, utilities.ts:8582 columns 38 and 44.
They are the callback argument and concise returned call at each site. All
other findings are unchanged, and no finding of any kind is added.
The remaining row is outside this adaptation's scope.

In scratch, only nodeFactory's body is reverted. The scan returns to 3 rows,
restoring exactly its two original findings. The transpile comparison against
the control then has exactly one changed body, utilities.ts; see
[mutant-scope.json](proofs/mutant-scope.json) and
[mutant-transpile.diff](proofs/mutant-transpile.diff).

## Sequential oracle runs

Setup finished before either tree was built. Both upstream builds finished
before either suite. The control oracle then ran alone to completion, passed
its exact counts and sanctioned-failure checks, and only then did the with-91
oracle start. Each used `--workers 4 --limit-seconds 1800`, all runners, no
filter. `free -m` was logged immediately before each run. All cgroup memory
counters, including OOM and OOM-kill, stayed zero. There was no concurrent
setup, adaptation, scanner build, or other test workload during either suite.

| Oracle | Passing | Failing | Pending | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Control without 91 | 106,366 | 1 | 0 | 360.135 |
| With 91 | 106,366 | 1 | 0 | 359.893 |

The sole failure in each run is exactly
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
Both install/build phases exit 0; both suite commands exit 1 for that sanctioned
failure. The existing API normalizer accepts all 223 sanctioned declarations,
with no missing, additional, or changed declaration. No sanction or expected
count is edited.

The full [baseline.diff](proofs/baseline.diff) is identical in both runs:
51,093 bytes, SHA-256
`5333568a661aea24fc8af3149e9f4d2dfca70731842461c677f385b092adc4f0`.
An extra-line mutant fails the existing actual baseline-byte checker, and the
diff is restored. Raw suite logs and reports are retained under proofs/logs/.

## Commands and environment

Commands run from /workspace/adamic. Test output goes directly to files.
[commands.json](proofs/commands.json) records every serial preparation command,
working directory, limit, exit, and timing. [oracles.py](proofs/oracles.py)
records and enforces the suite order and stopping gates.
The control was made with the same apply module and complete adaptation list,
excluding only 91; the with-91 tree used the normal full `stage3/apply.sh`.
The original control used an inline Python command; build-control.py preserves
that clone-and-apply operation for reproduction. Use fresh scratch directories
when reproducing the preparation script.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 900 bash cloud/setup.sh > /tmp/adapt91-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
timeout 180 npm ci --prefix stage3/api --no-audit --no-fund > /tmp/adapt91-api-install.log 2>&1
timeout 300 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/adapt91-upstream > /tmp/adapt91-source.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/adapt91-upstream timeout 900 python3 stage3/adapt/91-patcher-discard/proofs/build-control.py > /tmp/adapt91-control-apply.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/adapt91-upstream timeout 900 bash stage3/apply.sh /workspace/scratch/adapt91-after > /tmp/adapt91-after-apply.log 2>&1
timeout 1800 python3 stage3/adapt/91-patcher-discard/proofs/prepare.py > /tmp/adapt91-prepare-retry.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 5400 python3 stage3/adapt/91-patcher-discard/proofs/oracles.py > /tmp/adapt91-oracles.log 2>&1
```

The actual oracle commands were, serially:

```sh
bash stage3/oracle/run.sh /workspace/scratch/adapt91-control /workspace/scratch/adapt91-oracle-control --workers 4 --limit-seconds 1800
bash stage3/oracle/run.sh /workspace/scratch/adapt91-after /workspace/scratch/adapt91-oracle-after --workers 4 --limit-seconds 1800
```

Their wrapper imposes a 2,400-second hard limit per command and logs output.
Setup timing lines: Node ready 0.071s; Go ready 0.085s; clang ready 0.550s;
markdown ready 1.259s; submodules ready 246.962s; shared cache off 246.964s;
Go build ready 401.292s; test binaries deferred 401.397s; build cache warm
401.398s; total 401.429s. `nproc` is 5; CPU quota is 4. Node 24.19.0,
Go 1.27.1, clang 20.1.8. The environment file is
`/workspace/adamic-tools/env.sh`; [setup.log](proofs/setup.log) retains all lines.

No new Adamic program or production compiler change is introduced. All committed
paths are under this adaptation. No pull request is opened.
