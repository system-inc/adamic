Built a reproducible blocker probe for the three fourth-batch claims; the nine earlier ports remain completed and pushed.
Commits: earlier ports `431ab666`; fourth-batch claim `abe2cf3d`; this report and probe are committed separately.
Commands: owned blocker test PASS in 10.692s; three Go positive findings, three native exits 70; go vet PASS.
Mutant: skipping parser.file exits 0 with empty stderr on all three inputs and fails the expected-rejection comparison; this is a probe mutant, not three native rule mutants.
Not covered: fourth-batch native ports, corpus parity, rule mutants, sanitizer/released-handle gates and native/Go timing; work stopped at the shared parser blocker.

## Claimed rules and status

| Rule | Go positive control | Native entry |
| --- | --- | --- |
| react-hooks/set-state-in-effect | 1 finding, span 72..76, no fixes/suggestions | Parser rejects JSX at byte 101 |
| react-hooks/set-state-in-render | 1 finding, span 54..58, no fixes/suggestions | Parser rejects JSX at byte 75 |
| react-hooks/static-components | 1 finding, span 60..61, no fixes/suggestions | Parser rejects JSX at byte 62 |

All three remain **claimed, blocked, and unported**. No further rules were claimed. The claim was pushed before this reproduction was written. The all-heads scan inspected 389 origin refs, 33 distinct claim blobs naming 142 ranked rules, the base/main implementations, and 34 distinct origin typeaware trees. The selected rules each have zero compiler and repository findings in the frozen volume ranking.

## Observations

The native probe imports the existing `stage1/typescript/parser/parser.ts` and calls `Parser.file()` directly. It is compiled from `parser_probe.a` with the same stage-0 binary used for the preceding completed batch. Each syntax-valid TSX control fails before a native rule could run, with exit 70 and:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken
```

The three byte offsets and paths are preserved in `validation/*-native.stderr`. A non-JSX `function useThing() { return 1; }` control parses successfully, exits 0 and prints `parsed`. The parser-bypass mutant also exits 0 and prints `parsed`, with empty stderr, on each JSX control. The expected-rejection checks distinguish all three mutant results.

The independent Go oracle is an overlay inside cohere, calling its unmodified registry rules with a real checker, shared per-file cache, and independent program loader. It imports no bridge. Seed declarations give useState a real Dispatch alias and useEffect its declared name; the controls have no parse diagnostics. Full finding streams, including zero fixes and suggestions, are retained. JSX input copies have `.a` suffixes in the evidence directory; the Go test writes temporary `.tsx` files to select Go's TSX parser. These are probe fixtures, not rule implementations.

## Required substrate still absent

Reading the three production Go sources establishes additional dependencies. `set-state-in-render` consumes reverse-postorder SSA blocks, translated closure captures and `UnconditionalBlocks` (post-dominance), plus Dispatch alias identity. `set-state-in-effect` consumes the same representation after manual-memoization erasure/inlining, with ref value taint and `ControlDominators` exemptions, and reads Dispatch aliases and RefObject/effect-hook type symbols. `static-components` consumes SSA phi merges and dynamic component taint, compilation-unit boundaries and JSX tag values.

No corresponding native IR shelf is present under stage1/cohere/typeaware or stage1/typescript at this branch. This is a source inspection result, distinct from the executed JSX rejection. A type-checker bridge question can supply raw type identities; it does not supply native control flow, SSA or JSX parsing. Exporting cohere's rule verdicts or running its lowering as native rule logic would not establish an Adamic port.

The existing parser treats a leading `<` as a type assertion in `Parser.unary()` and has no JSX production. Resolving this requires shared parser support and a native IR/lowering substrate. The user's instruction says to stay inside own rule directories and stop on other blockers rather than edit shared files. No shared parser, harness, registration generator or protected compiler file was changed. No rule skeleton returning clean findings was registered. Nothing here claims byte parity or completed ports.

## Reproduction

Toolchain setup from the completed wave remains applicable: Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node 24.19.0 ready 1s, submodules 1s, cache/setup 114s; nproc 5. Environment: `source /workspace/adamic-tools/env.sh`.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_STAGE0=/workspace/wave-28-third-validation/adamic ADAMIC_WAVE28_FOURTH_ARTIFACTS=/workspace/wave-28-fourth-validation go test ./stage1/cohere/typeaware/wave28_fourth -count=1 -timeout=10m -v >/tmp/wave-28-fourth-blocker.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_fourth >/tmp/wave-28-fourth-vet.log 2>&1
```

On another checkout, build stage 0 with `go build -o /tmp/wave28-adamic ./cmd/adamic` (redirect output to a log) and set `ADAMIC_WAVE28_STAGE0` to that executable. Without that variable the reproduction explicitly skips. The first assertion expected an unsupported-primary panic; execution showed the more precise GreaterThanToken/SlashToken failure, and the assertion was corrected before the reported passing run.

No full repository test gate, compiler/repository corpus comparison, sanitizers or released-handle checks ran for this blocked batch. Native rule timing is unavailable. The prior nine rules' complete evidence remains in their existing reports. The claims remain owned for resumption once the substrate gap is resolved.
