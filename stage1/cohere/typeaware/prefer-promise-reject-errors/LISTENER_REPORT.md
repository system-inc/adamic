Added numeric listener declarations for all twelve wave 20 claims; nine are implemented rules.
Implementation commit: `9b6c2b1734ed6151feaf438956fe71f5a09b4361`, based on origin/main `e8ba3d5d`.
Checks: production-registration bytes and sanitizers PASS; all three diagnostic gates PASS; thirteen Adamic files lint clean, format idempotently, and vet passes.
Mutants: twelve declaration mutations, nine rule mutations, six checker-fact mutations, and five registry-retention mutations caught.
Not covered: shared numeric-node handoff, per-node handler conversion, three full React ports, own-rule emitted JavaScript comparison, full repository gate.

The new `syntaxKinds` export is a readonly numeric array. Each implemented
rule also exposes it as a readonly instance property. Numeric values are held
to the pinned typescript-go AST kinds, rather than guessed from stock
TypeScript enum values:

| Rule | Numeric listener kinds |
| --- | --- |
| no-floating-promises | ExpressionStatement 245 |
| no-implied-eval | CallExpression 214, NewExpression 215 |
| no-meaningless-void-operator | VoidExpression 223 |
| no-process-exit-after-output | SourceFile 307 |
| no-uncleared-race-timeout | CallExpression 214 |
| require-blocking-standard-streams | SourceFile 307 |
| prefer-promise-reject-errors | CallExpression 214, NewExpression 215 |
| prefer-regex-literals | SourceFile 307 |
| prefer-rest-params | Identifier 79 |
| react-hooks/set-state-in-effect | SourceFile 307, declaration only |
| react-hooks/set-state-in-render | SourceFile 307, declaration only |
| react-hooks/static-components | SourceFile 307, declaration only |

`testdata/listener_oracle.go` uses Go's parser to read the unchanged production
`rule.Listeners` literals, and obtains their numbers from the pinned Go AST
constants. It imports no native implementation and executes no copied lint
predicate. Native `listeners.a` imports the declarations and prints the same
registration records. Complete output is 136 identical bytes in normal and
ASan/UBSan/LSan native runs, with empty native stderr.

Each declaration mutant changes that module's first numeric kind by one.
Every mutant compiles and exits 0 with empty stderr. Only comparison with the
independent production-registration bytes rejects it. All twelve results and
first differing bytes are committed in `validation/listeners/declaration-mutants.json`.
These are declaration checks; the three React declaration mutants do not
establish that their absent source analyses are correct.

## Remaining speed-contract gap

This completes the requested declaration step, not the full speed conversion.
`stage1/typescript/parser/nodes.ts` still defines `readonly kind: string` and
has no numeric syntax-kind field. The existing parser returns that node type.
Legacy `run`, `visit`, and helper methods still scan nodes, read string kinds,
and refetch entry nodes. They have not been represented as satisfying the new
per-node handoff contract, and no dispatch performance win is claimed.

A shared numeric-node interface and kind-indexed driver are needed before
replacing those paths while retaining parser/bridge identity. This worker's
scope forbids changing shared parser, driver, harness or registration-generator
files. The new declarations are ready for that shared integration; current
runners ignore them. The numeric dialect must remain aligned with the pinned
parser, and the declaration oracle will reject a changed Go registration or
renumbered kind rather than silently accepting a mismatched dispatch table.

The three React source analyses also remain blocked by native JSX integration
and native React HIR/SSA/capture/control analysis. Their claims overlap waves
06 and 29, as documented in each BLOCKED.md. Their new `.a` files contain
metadata only. No replacement or further rules were claimed.

## Diagnostic regression evidence

All nine implemented rules were re-greened after the declarations were added.
The first batch passed in 160.306s, with 48 control findings and 12,953 identical
bytes normal and sanitized. The second batch passed all 121 projects and 160
findings. The third batch passed all 682 projects (662 valid upstream and 20
additional), 502 findings and 365 suggestions. Its known invalid TSX exclusion
remains documented in REPORT.md.

Every gate compares complete messages, UTF-8 spans, fixes, suggestion messages
and every repair, preserving duplicates and fix order. Each separately checks
the frozen 287-root repository and TypeScript v6.0.3's 77 compiler roots,
normal and ASan/UBSan/LSan: zero findings, 18,485 and 5,241 identical bytes.
Controls contain positive findings, so zero-finding corpora are not a vacuous
pass. All nine existing semantic rule mutants exit normally and are rejected
by Go finding bytes. Six raw checker-fact mutants are also caught, including
both accessed-property and callback-parameters in the first batch. Released
questions panic 70; all five registry-retention probes exit 0 and fail that
required-panic expectation. New listener metadata requires no bridge question
and changes no shared file.

Commands source `/workspace/adamic-tools/env.sh`. The existing setup remains
Go 1.27.1, clang 20.1.8, Node 24.19.0; last setup timing lines were Go 0s,
clang 1s, Node 1s, submodules 1s, cache warm 164s, done 164s. Toolchain is
reused; this is not a new setup run. `nproc` was rechecked and reports 5,
with the previously measured four-core quota.

```sh
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/listener_verify.py \
 --repository /workspace/adamic --artifacts /workspace/wave20-validation/listener-guards-twelve \
 --compiler /workspace/wave20-validation/landing-current-third/adamic \
 --checker /workspace/wave20-validation/landing-current-third/checker.a \
 > /workspace/wave20-validation/listener-guards-twelve.log 2>&1

ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/listener-first \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave20-validation/listener-first.log 2>&1

python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py \
 --repository /workspace/adamic --artifacts /workspace/wave20-validation/listener-second \
 --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored \
 > /workspace/wave20-validation/listener-second.log 2>&1

python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py \
 --repository /workspace/adamic --artifacts /workspace/wave20-validation/listener-third \
 --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases \
 > /workspace/wave20-validation/listener-third.log 2>&1

go vet ./... > /workspace/wave20-validation/listener-vet.log 2>&1
```

The isolated existing virtual-TypeScript formatter and configured Go lint
adapter checked thirteen changed/new Adamic source files. Formatting is
idempotent. An initial pass reported three comment-style findings on an
unquoted filename; quoting it fixed those findings, and final lint prints
`findings 0`. No physical Adamic `.ts` file or shared-harness workaround was
created. All diagnostic/test process output went to log files. The full
repository gate, old 26-rule comprehensive gate, and a fresh compiler Node
gate were not rerun for this metadata change; prior unchanged compiler Node
checks remain documented in LANDING_REPORT.md.

## Whole-process time

Fresh three-round alternating native/Go measurements after all builds,
mutants, sanitizer runs, formatting and lint completed. Every timed pair
produces complete matching diagnostic bytes; output hashes and each round
are in `validation/listeners/bench.json`. Native remains slower. These are
observations, not evidence of a dispatch optimization.

| Batch and corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| first-repository | 1.017448 | 0.418208 | 2.43 |
| first-compiler | 4.983445 | 1.425671 | 3.50 |
| second-repository | 0.365707 | 0.216989 | 1.69 |
| second-compiler | 2.123277 | 0.417635 | 5.08 |
| third-repository | 0.368543 | 0.215259 | 1.71 |
| third-compiler | 2.226982 | 0.519172 | 4.29 |

The branch was already based on unchanged current origin/main `e8ba3d5d`.
This worker publishes only `codex/typeaware-wave-20`, never main or an area
branch. The evidence commit following the implementation commit changes
reports and compressed logs only.
