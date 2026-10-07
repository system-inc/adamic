Published the parked React status, claimed the next three AST/checker rules, and added numeric listener declarations plus JSX blocker evidence.
Commits: parking f7b44295d; fifth-batch claim 0eba71324; completed nine-rule landing 786d33d3c on main f8013f0b.
Commands/output: owned JSX/listener test PASS 16.156s; three Go positive findings and native parser exits 70; vet PASS.
Mutants: parser bypass exits zero and is caught on three controls; three wrong-kind metadata mutants are caught by production Go listener comparison.
Not covered: all three new native ports, rule parity/mutants, corpus/sanitizer/released-handle gates and native/Go rule timings; stopped at shared JSX parsing.

## Selection and listener contract

The explicit all-heads audit covers 529 origin refs, 33 unique claim blobs naming 154 ranked rules, 25 ranked base/main ports, and 34 unique typeaware trees. The first three of 18 remaining candidates are react/jsx-fragments, react/jsx-no-constructed-context-values and react/jsx-no-undef; all have zero findings on both volume populations. No matching implementation was found in those trees. selection.json retains the exact candidate/count inventory. Their Go rules use JSX node listeners and syntax/checker walks, without the parked HIR/SSA/capture lowering. The constructed-context rule has its own AST/checker dependency-stability walk; this must be ported too, including cohere's memo divergence, when the parser dependency is available.

Each owned rule directory now contains rule.json with numeric kinds. The private independent oracle calls the unmodified production subject.Run with a real source file and checker, reads its actual listener keys and compares them with these declarations: fragments [285,286,289] (JsxElement, JsxSelfClosingElement, JsxFragment); context and undef [286,287] (JsxSelfClosingElement, JsxOpeningElement). These are pinned typescript-go AST enum values, not invented rule IDs or strings compared inside a rule. Replacing the first numeric kind with zero is rejected for each declaration. No handler is registered yet: metadata is preparation, not completed native rule logic. The current native ParseNode still has only a string kind, so the declarations await its integrated numeric API. This also corrects the earlier implication that declarations themselves must wait for a numeric node field: declarations can be written and verified now.

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
