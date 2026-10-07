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

## Resumption audit

An all-heads fetch on the next continuation inspected 417 origin refs and 117 distinct stage1 trees. Main is `e011f8f60899586d6373a5ccb07335ad82cfbf3c`, the bridge branch is `eb6df00e91b07c9fc81b2ec396a6f118be55d172`, and the harness branch is `f4d98cab50048692781da3599131317dc569d466`. All three retain parser blob `bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d`, with no JSX production. Four distinct parser.ts blobs occur across origin branches; three contain no JSX references.

There IS now a published JSX implementation on `origin/codex/stage1-jsx-lint` at `a8a62d62ca49db7415e14c3887dd305022b17309`, parser blob `62ab6514da477dfc54f477b6a968735e24901b58`. Source inspection confirms its `Jsx` descent and extension-based TSX mode. Its JSX_REPORT.md reports parity tests; those upstream tests were not executed by this worker. This corrects the implication that JSX support is absent on every upstream branch: it is available for integration, but has not been integrated into this branch, main, the bridge branch or the named harness branch. No shared parser files were imported or modified.

A native-source search across all 117 distinct stage1 trees found no occurrences of `UnconditionalBlocks`, `ControlDominators`, `AsCompilationUnit`, `ForFunctionWithoutManualMemoization`, post-dominator spellings, or Phi/HIR class declarations, and no native HIR/SSA/post-dominator shelf paths. This is a source inventory observation, not an executed proof that no differently named equivalent could exist. The inspected JSX change supplies syntax trees, not the required native SSA/closure/control-flow substrate. All three ports therefore remain blocked and unported; none was silently registered as producing zero findings.

The exact owned blocker command above was repeated with artifacts at `/workspace/wave-28-fourth-resume-validation` and log `/tmp/wave-28-fourth-resume-blocker.log`. It passed in 15.659s: each Go control again emits one finding; each native parser run exits 70 at the same byte; the non-JSX control passes; the parser-bypass mutant exits normally with empty stderr on all three controls and is distinguished by the rejection check. The retained log is `validation/resume-blocker.log`. No native rule mutants, corpus parity, sanitizer/released-handle gates or native rule timings are claimed for this blocked batch. No further rules were claimed. Work stops under the instruction to report other blockers rather than edit shared files.

## Numeric listener resumption audit

The explicit all-heads fetch still places main at `e8ba3d5d81de4d3773c723914fccd4c76248b965` and the own remote branch at `f6cb0d06388799c23d67dc81877786834c58faf1`. Main is already an ancestor of this branch. The previous full nine-rule oracle gates therefore remain against current main; no rebase is necessary. Only this own branch is a push destination.

The newly required numeric dispatch contract is not available from the integrated parser. `stage1/typescript/parser/nodes.ts` declares `readonly kind: string` and accepts a string in its constructor, with no numeric kind field. Its main blob is `1e964ea7a66e8d6a09ce7d772ecd673caa2cc68c`; parser.ts remains `bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d`. The named dot-a harness branch also exposes the same string-only node API. Searching the integrated stage1 TypeScript/typeaware sources found no SyntaxKind, kindId or kindCode API. This is a source inventory observation.

The nine existing ports still use string kinds and per-rule walks. Their prior correctness results do not establish compliance with the new speed rule. No numeric listener declarations or performance improvement are claimed. A local string-to-number adapter would still read a node kind as a string, contrary to the instruction. Adding numeric fields requires changing shared parser files, outside the explicit own-directory boundary. Work stops at that dependency rather than claiming a compliant listener.

The three React rules remain blocked by the same JSX and native SSA/control-flow requirements documented above. Published JSX support is still on its separate branch. No further rules were claimed, no placeholder verdicts were registered, and no shared files were edited.

The fresh `TestSharedParserBlocker` run passed in 61.462s. Go again reports one finding per rule, and native JSX parsing exits 70 at bytes 101, 75 and 62. The parser-bypass mutant exits 0 with empty stderr on all three controls and is caught by the expected-rejection comparison; the ordinary non-JSX control passes. This is one probe mutant exercised three times, not three native rule mutants. `go vet ./stage1/cohere/typeaware/wave28_fourth` and `git diff --check` pass. The exact logs are retained as `validation/numeric-resume-*.log`. Setup completed successfully; its timing lines and nproc=5 are retained in the setup log.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_STAGE0=/workspace/wave-28-latest-third/adamic ADAMIC_WAVE28_FOURTH_ARTIFACTS=/workspace/wave-28-speed-blocker go test ./stage1/cohere/typeaware/wave28_fourth -count=1 -timeout=10m -v > /tmp/wave28-speed-blocker.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_fourth > /tmp/wave28-speed-vet.log 2>&1
```

No completed rule code changed, so the nine full gates on this same main were not repeated. Full repository tests, numeric listeners, new performance measurements and completed React ports remain uncovered.
