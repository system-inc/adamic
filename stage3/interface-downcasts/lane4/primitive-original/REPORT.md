Certified four original primitive pairs, 69 candidate reads; direct EmitNode destructuring remains open.
Commit: this checkpoint on codex/views-mixed-unions-2, after 6ed3ced0.
Checks: stock declaration preparation passed; original C/JS/Node oracle passed in 80.768s.
Mutants: five read-check bypass cases run in both backends; each violates its named exit-70 pin.
Uncovered: direct EmitNode destructuring, remaining primitive shapes, arrays and dictionary selectors; no exact tsc reachability claim.

| Original pair | Candidate reads | Status |
| --- | ---: | --- |
| EvaluatorResult<string \| number \| undefined>.value | 61 | Certified |
| NodeLinks.isExhaustive | 4 | Certified, including rejection of number 1 where only 0 is allowed |
| EvaluatorResult<string \| number \| undefined> \| undefined, value | 3 | Certified |
| EmitNode \| undefined, constantValue | 1 | Certified |
| EmitNode.constantValue | 1 | Property component passes; original destructuring remains uncertified |

Fixtures import the complete original declarations from pinned TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8. Preparation verifies all 78 declaration
hashes, original source read locations and complete receiver field sets. Required
undefined and optional missing values remain distinct. Each valid member is held
to Node in both backends; positive native runs use sanitizers and leak checks.
Wrong boolean, null, wrong numeric literal and wrong string values have named
refusal pins. Helpers receive viewed values through the shared flow.

Run: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewOriginal(EvaluatorPrimitivePairs|FinitePrimitiveFields)$'
-count=1 -v. Output is logs/oracle.log. No new compiler hook is needed for these
four pairs: the integrated nullable primitive selector supplies the checks.
The five bypass cases include the uncredited direct EmitNode property component;
eight backend kills belong to the four certified groups, two to that component.

Original mixed-primitive column now has 38 / 620 certified and 25 / 70 remaining.
One remaining intersection pair / one read belongs to lane 7. Our nonbrand
remainder is 24 / 69, using candidate counts. __String is delivered early;
mixed primitives retain October 13, 2026, 17:00 MDT as the estimate. Next is the
original destructuring syntax, then the remaining pure primitive unions.
