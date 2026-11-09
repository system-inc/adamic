# Deferred refusal census

Overstatement of unit first-error admission: **0 units** move from no lowering/deferred first stop to a deferred NotYet.

This is the body/refusal census view. The baseline pooled-checker overflow conservatively marks all attempts at run end, so these counts are not proof that any unit successfully compiles.

| Unit | Location | Deferred reason |
| --- | --- | --- |
| None | | |

## New findings by exact reason and file

| Reason | File | Raw | Distinct |
| --- | --- | ---: | ---: |
| a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/classFields.ts | 1 | 1 |
| a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/es2017.ts | 1 | 1 |
| a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/es2018.ts | 1 | 1 |
| a function taking (hint: EmitHint, node: Node) => void seen as one taking (hint: EmitHint, node: Node) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/ts.ts | 1 | 1 |
| a function taking NamedExportBindings seen as one taking NamedExports (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/ts.ts | 1 | 1 |
| a function taking NamedImportBindings seen as one taking NonNullable<NamedImportBindings \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/ts.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<BigIntLiteral \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/esDecorators.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<BigIntLiteral \| Identifier \| NumericLiteral \| PrivateIdentifier \| StringLiteralLike> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/esDecorators.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<BindingOrAssignmentElement> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/destructuring.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<Expression & AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: ThisExpression; }> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/classFields.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<Expression & CallExpression & { readonly expression: Identifier; }> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/classFields.ts | 1 | 1 |
| a function taking Node seen as one taking NonNullable<MethodDeclaration \| ShorthandPropertyAssignment \| AccessorDeclaration> (tsc relates a method's parameters both ways), so it can be handed what it can't take | src/compiler/transformers/es2018.ts | 1 | 1 |
| a value of type (ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 9 more ... \| VariableDeclaration)[] \| undefined seen as BindingOrAssignmentElement[] \| undefined, which can write BindingOrAssignmentElement where ArrayLiteralExpression \| AssignmentExpression<EqualsToken> \| BindingElement \| ElementAccessExpression \| ... 9 more ... \| VariableDeclaration is read | src/compiler/transformers/destructuring.ts | 1 | 1 |
| optional property modifiers in ArrowFunction absent from structural source never, which can hide fields | src/compiler/transformers/ts.ts | 1 | 1 |
| overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes | src/compiler/debug.ts | 1 | 1 |
| overload 1 of assertNode parameter node cannot be served by implementation parameter node | src/compiler/debug.ts | 1 | 1 |
| overload 1 of visitEachChild parameter node cannot be served by implementation parameter node | src/compiler/visitorPublic.ts | 1 | 1 |

## Other movements

Existing unit first stops changed: 10. Other new unit stops: 0. Unit first stops cleared: 0.

| View | New findings | New sites | Cleared findings | Moved findings | Panics before | Panics after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| raw_findings | 17 | 17 | 73 | 2 | 0 | 0 |
| distinct_reads | 17 | 17 | 73 | 2 | 0 | 0 |

### Changed existing first stops

| Unit | Before | After |
| --- | --- | --- |
| src/compiler/transformers/destructuring.ts:244:1 | src/compiler/debug.ts:273:82: Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test | src/compiler/debug.ts:273:61: Refused: overload 1 of assertNode parameter node cannot be served by implementation parameter node |
| src/compiler/transformers/destructuring.ts:600:1 | src/compiler/debug.ts:257:86: Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | src/compiler/debug.ts:257:65: Refused: overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes |
| src/compiler/transformers/destructuring.ts:605:1 | src/compiler/debug.ts:257:86: Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | src/compiler/debug.ts:257:65: Refused: overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes |
| src/compiler/transformers/destructuring.ts:610:1 | src/compiler/debug.ts:257:86: Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | src/compiler/debug.ts:257:65: Refused: overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes |
| src/compiler/transformers/destructuring.ts:615:1 | src/compiler/debug.ts:257:86: Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | src/compiler/debug.ts:257:65: Refused: overload 1 of assertEachNode parameter nodes cannot be served by implementation parameter nodes |
| src/compiler/transformers/es2016.ts:21:1 | src/compiler/visitorPublic.ts:587:119: Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | src/compiler/visitorPublic.ts:585:48: Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node |
| src/compiler/transformers/es2019.ts:17:1 | src/compiler/visitorPublic.ts:587:119: Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | src/compiler/visitorPublic.ts:585:48: Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node |
| src/compiler/transformers/es2020.ts:42:1 | src/compiler/visitorPublic.ts:587:119: Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | src/compiler/visitorPublic.ts:585:48: Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node |
| src/compiler/transformers/es2021.ts:25:1 | src/compiler/visitorPublic.ts:587:119: Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | src/compiler/visitorPublic.ts:585:48: Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node |
| src/compiler/transformers/legacyDecorators.ts:88:1 | src/compiler/visitorPublic.ts:587:119: Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | src/compiler/visitorPublic.ts:585:48: Refused: overload 1 of visitEachChild parameter node cannot be served by implementation parameter node |

### Other newly stopped units

| Unit | Before | After |
| --- | --- | --- |
| None | | |

### Cleared unit first stops

| Unit | Before | After |
| --- | --- | --- |
| None | | |

## Pooled-checker overflow and production roots

Baseline TS2859 (`" "` against `{}`) was first observed by the run-end relation probe after the last timed unit, `src/tsc/_namespaces/ts.ts:3:1`. The duration-only journal has no per-unit overflow answers; the onset unit is unknown. This does not establish that the last unit caused the overflow.

The complete relation trailers and conservative all-unit attribution are retained in the streams. Program findings use the <program> identity; deferred_scan module events are not unit first stops.

Production first stops: 0 of 73 move from `core.ts:2599:15`. Production lowering and refusal sources are byte-identical to 018b7c21. The 73 production roots were not recompiled.

The paired JSON retains all finding removals, reason changes, new findings, unit movements and panics. Counts here use the pair tool identities without re-normalizing them.

## Eligibility and other finding changes

Both runs have the same 4,059 attempted units and 504 deferred generic declarations, including 112 instantiated witnesses. Every unit-ledger entry and the complete relation trailers are identical. Neither run has a deferred or deferred_scan finding on this entry.

All 17 new findings are Refused: 14 refusal_scan and 3 lowering events. The 73 cleared findings are 61 Refused and 12 NotYet, comprising 70 refusal_scan and 3 lowering events. These are finding removals, not units newly proven to lower.

The two moved findings are the written-any assertion at parser.ts:1253:17 and the corresponding visitorPublic.ts:603:17 assertion. Both retain the ordinary unchecked-cast refusal. The continuing baseline scan had emitted deferred fallbacks despite ordinary findings; the corrected scan suppresses them.

Ten existing unit first stops change among ordinary overload refusals: five destructuring units change assertNode/assertEachNode parameter choice, and five transformer units change visitEachChild overload/parameter choice. The exact before/after rows are above; none becomes newly admitted or newly stopped.

This paired entry therefore measures zero deferred admission overstatement. It does not negate the direct and instantiated written-any regression fixtures, whose dropped-read mutant fails as required. See metadata-comparison.json and paired.json.gz for complete records.
