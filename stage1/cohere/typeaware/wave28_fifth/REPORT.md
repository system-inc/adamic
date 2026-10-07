Published the parked React status, claimed the next three AST/checker rules, and added named listener declarations plus JSX blocker evidence.
Commits: parking f7b44295d; fifth-batch claim 0eba71324; completed nine-rule landing 786d33d3c on main f8013f0b.
Commands/output: owned named-kind JSX/listener test PASS 19.846s; three Go positive findings and native parser exits 70; vet PASS.
Mutants: parser bypass exits zero and is caught on three controls; three wrong-kind metadata mutants are caught by production Go listener comparison.
Not covered: all three new native ports, rule parity/mutants, corpus/sanitizer/released-handle gates and native/Go rule timings; stopped at shared JSX parsing.

## Selection and listener contract

The explicit all-heads audit covers 529 origin refs, 33 unique claim blobs naming 154 ranked rules, 25 ranked base/main ports, and 34 unique typeaware trees. The first three of 18 remaining candidates are react/jsx-fragments, react/jsx-no-constructed-context-values and react/jsx-no-undef; all have zero findings on both volume populations. No matching implementation was found in those trees. selection.json retains the exact candidate/count inventory. Their Go rules use JSX node listeners and syntax/checker walks, without the parked HIR/SSA/capture lowering. The constructed-context rule has its own AST/checker dependency-stability walk; this must be ported too, including cohere's memo divergence, when the parser dependency is available.

Each owned rule directory contains rule.json with typescript-go ast.Kind names, as clarified by the user's corrected contract. The independent oracle calls unmodified production subject.Run with a real source file/checker, reads its actual listener keys, strips the Kind prefix from their names and compares sorted lists: fragments [JsxElement, JsxFragment, JsxSelfClosingElement]; context and undef [JsxOpeningElement, JsxSelfClosingElement]. Replacing the first kind with Identifier is rejected for each declaration. There is no numeric-node API dependency: the earlier numeric declarations and logs are historical and superseded. No handler is registered yet; JSX parsing remains the actual blocker, and metadata is not completed native rule logic.

## Executed JSX blocker

All three controls are valid TSX and each emits exactly one Go finding: fragment preferFragment, context defaultMsg, undef jsxIdentifierNotDefined. The current native parser rejects them before any listener executes. Native stderr records GreaterThanToken/SlashToken at byte 63 for fragments, GreaterThanToken/Identifier at byte 68 for the context value attribute, and GreaterThanToken/SlashToken at byte 34 for undef. Non-JSX parsing succeeds. Skipping parser.file makes the mutant print parsed, exit zero and leave stderr empty; the expected-rejection comparison catches it on each control. This is a parser-probe mutant, not three native rule mutants.

The first fragment fixture placed React's declaration in a separate seed file, making Go's default jsx-no-undef also report React. The independent oracle correctly rejected that fixture's one-finding assertion. React/Ctx declarations were moved into the corresponding source files; no rule was changed or finding ignored. The initial failed log is preserved. Source controls committed as .a are copied to temporary .tsx files only for Go's JSX parser selection.

Main f8013f0b still has parser.ts blob bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d, without JSX descent. JSX support exists on codex/stage1-jsx-lint and is landing through area/stage1-lint; it has not landed in current main. The explicit territory instruction prohibits shared parser/harness edits. No shared code was modified, no other branch was pushed, and no placeholder rule returning clean findings was registered. These three remain claimed, blocked and unported; they are not the three HIR/SSA claims parked under #dnv6f2c.

## Reproduction and limits

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_STAGE0=/workspace/wave28-f801-third/adamic ADAMIC_WAVE28_FIFTH_ARTIFACTS=/workspace/wave28-fifth-probe go test ./stage1/cohere/typeaware/wave28_fifth -count=1 -timeout=10m -v > /tmp/wave28-fifth-final.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_fifth > /tmp/wave28-fifth-vet.log 2>&1
```

Without ADAMIC_WAVE28_STAGE0 the blocker test explicitly skips. Toolchain setup remains the recorded 165-second successful setup, nproc 5. The complete prior nine-rule gates remain green against the same main; no completed implementation changed here. No full repository gate or new native-rule performance/corpus/parity claim is made. Work stops at the shared parser dependency rather than editing outside owned directories.

## Named-kind contract correction

The explicit user correction supersedes all earlier numeric-kind metadata and numeric-API blocker descriptions. Current rule.json kinds are names as required by the registry; no numeric parser field is needed. The updated oracle compares those names to the actual unmodified Go listener maps, and an Identifier substitution is caught for each declaration. The fresh blocker/listener test passes in 19.846s; vet is clean. The JSX findings/rejections and parser-bypass mutant observations are unchanged. Evidence is validation/named-kinds.log, named-kinds-vet.log and named-listener-kinds.stdout.

Main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, already an ancestor of own pushed landing 27e2ba2d3. No completed native rule implementation changed, so its nine full gates remain applicable. The named harness ab70f38d4 still has not landed on main or area/stage1-lint; shared JSX parsing is the remaining active-claim blocker. The separate-history rebase previously rejected by automatic review was not retried. No new rules were claimed and no shared source changed.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_STAGE0=/workspace/wave28-c019-third/adamic ADAMIC_WAVE28_FIFTH_ARTIFACTS=/workspace/wave28-named-probe go test ./stage1/cohere/typeaware/wave28_fifth -count=1 -timeout=10m -v > /tmp/wave28-named-test.log 2>&1
go vet ./stage1/cohere/typeaware/wave28_fifth > /tmp/wave28-named-vet.log 2>&1
```
