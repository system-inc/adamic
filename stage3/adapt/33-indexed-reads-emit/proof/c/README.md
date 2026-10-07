# Wave C

184 reviewed indexed-expression occurrences: 13 required assertions, 2 numeric bitwise defaults, 169 declines.

Census before this wave is the tree with 10 and 30; previous waves do not edit these files. Options remain Adamic strictness. All 78 compiler roots are loaded.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 |
| --- | --- | --- | --- | --- | --- |
| transformer.ts | 0 → 0 | 0 → 0 | 4 → 2 | 5 → 4 | 0 → 0 |
| visitorPublic.ts | 1 → 1 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/classFields.ts | 4 → 4 | 2 → 2 | 1 → 1 | 0 → 0 | 1 → 1 |
| transformers/classThis.ts | 1 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/declarations/diagnostics.ts | 0 → 0 | 0 → 0 | 0 → 0 | 1 → 1 | 0 → 0 |
| transformers/destructuring.ts | 11 → 0 | 2 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/es2016.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/es2017.ts | 0 → 0 | 6 → 6 | 1 → 1 | 1 → 1 | 0 → 0 |
| transformers/es2018.ts | 1 → 1 | 0 → 0 | 1 → 1 | 2 → 2 | 0 → 0 |
| transformers/es2019.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/es2020.ts | 1 → 0 | 6 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/es2021.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/esDecorators.ts | 0 → 0 | 2 → 2 | 0 → 0 | 2 → 2 | 1 → 1 |
| transformers/esnext.ts | 6 → 6 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/generators.ts | 11 → 11 | 7 → 7 | 5 → 5 | 1 → 1 | 0 → 0 |
| transformers/jsx.ts | 2 → 2 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/legacyDecorators.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/module/esnextAnd2015.ts | 4 → 4 | 0 → 0 | 1 → 1 | 0 → 0 | 0 → 0 |
| transformers/module/impliedNodeFormatDependent.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/module/module.ts | 1 → 1 | 0 → 0 | 1 → 1 | 2 → 2 | 0 → 0 |
| transformers/module/system.ts | 0 → 0 | 0 → 0 | 1 → 1 | 3 → 3 | 0 → 0 |
| transformers/namedEvaluation.ts | 1 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/taggedTemplate.ts | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/typeSerializer.ts | 1 → 0 | 4 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |
| transformers/utilities.ts | 5 → 0 | 0 → 0 | 0 → 0 | 0 → 0 | 0 → 0 |

Total: 113 → 78. Every remaining finding follows.

- `transformer.ts:367:9 TS2532`: This sparse numeric bitwise compound assignment also reads a slot, but ?? 0 cannot be inserted on its assignment target; a separate assignment-shape adaptation is deferred.
- `transformer.ts:395:9 TS2532`: This sparse numeric bitwise compound assignment also reads a slot, but ?? 0 cannot be inserted on its assignment target; a separate assignment-shape adaptation is deferred.
- `transformer.ts:563:9 TS2322`: The stack slot is populated before restoration but its saved declaration-list payload intentionally can be undefined when no declarations were accumulated.
- `transformer.ts:564:9 TS2322`: The stack slot is populated before restoration but its saved declaration-list payload intentionally can be undefined when no declarations were accumulated.
- `transformer.ts:565:9 TS2322`: The stack slot is populated before restoration but its saved declaration-list payload intentionally can be undefined when no declarations were accumulated.
- `transformer.ts:614:9 TS2322`: The stack slot is populated before restoration but its saved declaration-list payload intentionally can be undefined when no declarations were accumulated.
- `transformers/classFields.ts:1835:21 TS2532`: Presence of privateInstanceMethodsAndAccessors[0] in visitInNewClassLexicalEnvironment has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/classFields.ts:2251:45 TS2538`: Presence of statementsIn[superStatementIndex] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/classFields.ts:2252:97 TS18048`: Presence of statementsIn[superStatementIndex] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/classFields.ts:2253:27 TS18048`: Presence of statementsIn[superStatementIndex] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/classFields.ts:2297:52 TS2345`: Presence of statementsIn[statementOffset] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/classFields.ts:2376:56 TS2345`: Presence of constructor.body.statements[statementOffset] in transformConstructorBody has not been established; its producer and index invariant review remains unfinished.
- `transformers/classFields.ts:2864:52 TS2345`: The private-accessor object explicitly supplies undefined setterName to an exact optional property; not an indexed read.
- `transformers/classFields.ts:2896:52 TS2345`: The private-accessor object explicitly supplies undefined getterName to an exact optional property; not an indexed read.
- `transformers/declarations/diagnostics.ts:166:9 TS2322`: The returned diagnostic object supplies optional typeName with an undefined union; an exact optional property return-shape obligation, not an indexed read.
- `transformers/es2017.ts:325:21 TS2322`: The modeled Set constructor widens iterable elements to include undefined; this allocation is not an indexed read.
- `transformers/es2017.ts:327:17 TS18048`: catchClauseUnshadowedNames remains optional after the preceding Set assignment fails its element-type check; not an indexed read.
- `transformers/es2017.ts:599:86 TS2532`: Presence of node.declarations[0] in visitVariableDeclarationListWithCollidingNames has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/es2017.ts:764:38 TS18048`: The outer-parameter list must remain aligned with the original parameters across the transformation; its parallel length and population invariant review remains unfinished.
- `transformers/es2017.ts:765:25 TS18048`: The outer-parameter list must remain aligned with the original parameters across the transformation; its parallel length and population invariant review remains unfinished.
- `transformers/es2017.ts:765:58 TS18048`: The outer-parameter list must remain aligned with the original parameters across the transformation; its parallel length and population invariant review remains unfinished.
- `transformers/es2017.ts:767:76 TS18048`: The outer-parameter list must remain aligned with the original parameters across the transformation; its parallel length and population invariant review remains unfinished.
- `transformers/es2017.ts:770:44 TS18048`: The outer-parameter list must remain aligned with the original parameters across the transformation; its parallel length and population invariant review remains unfinished.
- `transformers/es2018.ts:534:35 TS2532`: Presence of objects[0] in visitObjectLiteralExpression has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/es2018.ts:537:17 TS2322`: Presence of objects[0] in visitObjectLiteralExpression has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/es2018.ts:540:80 TS2322`: Presence of objects[i] in visitObjectLiteralExpression has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/es2018.ts:636:29 TS2345`: Presence of node.elements[i] in visitCommaListExpression has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/esDecorators.ts:1164:45 TS2538`: Presence of statementsIn[superStatementIndex] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/esDecorators.ts:1165:97 TS18048`: Presence of statementsIn[superStatementIndex] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/esDecorators.ts:1192:80 TS18048`: Presence of superPath[superPathDepth] in transformConstructorBodyWorker has not been established; its producer and index invariant review remains unfinished.
- `transformers/esDecorators.ts:351:9 TS2322`: The lexical-environment object supplies undefined savedPendingExpressions to an exact optional property; not an indexed read.
- `transformers/esDecorators.ts:398:13 TS2322`: The lexical-environment object supplies undefined savedPendingExpressions to an exact optional property; not an indexed read.
- `transformers/esnext.ts:183:34 TS2345`: Presence of node.statements[pos] in visitSourceFile has not been established; its producer and index invariant review remains unfinished.
- `transformers/esnext.ts:205:52 TS2345`: arrayFrom(exportBindings.values()) has optional iterable elements in the built-in model; not an indexed read.
- `transformers/esnext.ts:344:44 TS2345`: Presence of statementsIn[i] in transformUsingDeclarations has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/esnext.ts:384:36 TS2345`: Presence of statementsIn[i] in transformUsingDeclarations has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/esnext.ts:767:34 TS2345`: Presence of statements[i] in countPrologueStatements has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/esnext.ts:767:70 TS2345`: Presence of statements[i] in countPrologueStatements has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1283:39 TS2345`: Presence of statements[i] in transformAndEmitStatements has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1370:35 TS18048`: Presence of variables[i] in transformAndEmitVariableDeclarationList has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1374:70 TS2345`: Presence of variables[i] in transformAndEmitVariableDeclarationList has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1684:46 TS2532`: Presence of initializer.declarations[0] in transformAndEmitForInStatement has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/generators.ts:1729:55 TS2532`: Presence of initializer.declarations[0] in visitForInStatement has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/generators.ts:1866:21 TS18048`: Presence of caseBlock.clauses[i] in transformAndEmitSwitchStatement has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1880:25 TS18048`: Presence of caseBlock.clauses[i] in transformAndEmitSwitchStatement has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1881:43 TS18048`: Presence of caseBlock.clauses[i] in transformAndEmitSwitchStatement has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1887:62 TS18048`: Presence of caseBlock.clauses[i] in transformAndEmitSwitchStatement has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:1889:55 TS2345`: The clause-label list must remain aligned with caseBlock.clauses at i; its parallel length and population invariant review remains unfinished.
- `transformers/generators.ts:1889:85 TS18048`: The clause-label list must remain aligned with caseBlock.clauses at i; its parallel length and population invariant review remains unfinished.
- `transformers/generators.ts:1911:27 TS2345`: Presence of clauseLabels[defaultClauseIndex] in transformAndEmitSwitchStatement has not been established; its producer and index invariant review remains unfinished.
- `transformers/generators.ts:1918:27 TS2345`: The clause-label list must remain aligned with caseBlock.clauses at i; its parallel length and population invariant review remains unfinished.
- `transformers/generators.ts:1919:44 TS2532`: Presence of caseBlock.clauses[i] in transformAndEmitSwitchStatement has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2446:48 TS2345`: Presence of blockStack[j] in hasImmediateContainingLabeledBlock has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2469:56 TS2345`: Presence of blockStack[i] in findBreakTarget has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2472:53 TS2345`: Presence of blockStack[i] in findBreakTarget has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2480:48 TS2345`: Presence of blockStack[i] in findBreakTarget has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2499:51 TS2345`: Presence of blockStack[i] in findContinueTarget has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2507:51 TS2345`: Presence of blockStack[i] in findContinueTarget has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2880:63 TS18048`: Presence of withBlockStack[i] in appendLabel has not been established; its loop-bound and populated-list invariant review remains unfinished.
- `transformers/generators.ts:2951:21 TS2532`: The preceding same-slot undefined test handles absence by allocation; preserve this repeated bucket read in its else branch.
- `transformers/generators.ts:2983:57 TS2532`: The block offset, action and block arrays must have corresponding populated slots at blockIndex; that parallel invariant review remains unfinished.
- `transformers/generators.ts:2984:23 TS2322`: The block offset, action and block arrays must have corresponding populated slots at blockIndex; that parallel invariant review remains unfinished.
- `transformers/jsx.ts:309:58 TS2345`: Presence of nonWhitespaceChildren[0] in convertJsxChildrenToChildrenPropAssignment has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/jsx.ts:502:62 TS2345`: Presence of expressions[0] in transformJsxAttributesToExpression has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/esnextAnd2015.ts:100:46 TS2345`: Presence of node.arguments[0] in <callback> has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/esnextAnd2015.ts:163:91 TS2345`: Presence of importsAndRequiresToRewriteOrShim[0] in visitor has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/esnextAnd2015.ts:194:37 TS2345`: Presence of node.arguments[0] in visitImportOrRequireCall has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/esnextAnd2015.ts:196:81 TS2345`: Presence of node.arguments[0] in visitImportOrRequireCall has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/esnextAnd2015.ts:251:22 TS2532`: Presence of importRequireStatements[1].declarationList.declarations[0] in createRequireCall has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/module.ts:2262:13 TS2322`: Presence of moduleInfoMap[getOriginalNodeId(currentSourceFile)] in onEmitNode has not been established; its table population invariant review remains unfinished.
- `transformers/module/module.ts:246:42 TS2345`: Presence of node.arguments[0] in <callback> has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/module.ts:2481:21 TS2322`: arrayFrom(bindingsSet) has optional iterable elements in the built-in model; not an indexed read.
- `transformers/module/module.ts:382:72 TS2532`: Presence of jsonSourceFile.statements[0] in transformAMDModule has not been established; its nonempty-list invariant review remains unfinished.
- `transformers/module/system.ts:1773:13 TS2322`: Presence of moduleInfoMap[id] in onEmitNode has not been established; its table population invariant review remains unfinished.
- `transformers/module/system.ts:1774:13 TS2322`: Presence of exportFunctionsMap[id] in onEmitNode has not been established; its table population invariant review remains unfinished.
- `transformers/module/system.ts:1776:13 TS2322`: Presence of contextObjectMap[id] in onEmitNode has not been established; its table population invariant review remains unfinished.
- `transformers/module/system.ts:289:21 TS2532`: Presence of dependencyGroups[groupIndex] in collectDependencyGroups has not been established; its producer and index invariant review remains unfinished.
- `visitorPublic.ts:424:59 TS2345`: Presence of parameters[i] in addDefaultValueAssignmentsIfNeeded has not been established; its loop-bound and populated-list invariant review remains unfinished.

The occurrence ledger is in [sites.json](../../sites.json). Survey coordinates document the original source; addressing uses parsed expressions and occurrence counts.

Commands (stdout/stderr saved directly to logs):

```sh
node adapt.cjs TREE
bash census.sh TREE OUTPUT
node verify.cjs BEFORE TREE
node adapt.cjs TREE # second run: zero edits
MUTANT_FILE=transformers/destructuring.ts node mutant.cjs BEFORE TREE NEW_MUTANT_TREE
stage3/oracle/run.sh TREE OUTPUT # default suites, adaptations 10 + 30 + 33, no 20
```

The mutant changes transformers/destructuring.ts elements[i]! to (elements[i] ?? 0). Both emitted-JavaScript equality and the site contract reject it with exit 1. No mutant oracle was run. CRLF is preserved because edits insert tokens without rewriting lines.

Default oracle: pass; {'passing': 106367, 'failing': 0, 'pending': 0}; 0 baseline differences; 214.656 seconds. See oracle.json and baseline.diff.
