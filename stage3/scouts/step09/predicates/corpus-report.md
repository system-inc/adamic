# Predicate corpus measurement

This measures roadmap step 09 on compiler/step09-predicates-ahead at 90a78f4959fd4401f187321754b1c070a612b27a. The source is the exact adapted tree behind codex/step09-double-casts at 8a7ab17e, rebuilt with that pin's apply.sh. All 79 inventoried source hashes match. All 580 original predicate annotations with bodies match by file:line:column; the 71 bodyless annotations are excluded. There are 4,383 resolved calls to those bodies within the 79 compiler roots.

The production loader still rejects this corpus with its 320 retained checker diagnostics. A scratch overlay retains those diagnostics for the same latent proof measurement used by the original scout, and disables both normal loader entry points and executable IR output. Production proof and check-construction functions are unchanged. These results measure their gates in the original typed context; backend execution evidence comes from the focused fixtures below.

There are 309 successful logical body proofs. Of those, 18 pass the predicate admission gate and 291 stop while admitting the complete view. In .a mode, 271 bodies have a Refused diagnostic and 291 have a NotYet diagnostic. In .ts mode, 18 unproved predicates construct checks at 54 original calls: 54 true checks and 18 false checks, with 36 false directions unobservable. There are 495 predicates with a confirmed view dependency. The original three proven sites remain proven.

| Original group | Bodies | Logical body proof | Admitted proof | Checked .ts predicates (calls) | Refused .a | NotYet .a | Pending views in .ts |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| direct kind comparison | 283 | 279 | 0 | 0 (0) | 4 | 279 | 283 |
| delegation or composition of predicate calls | 115 | 13 | 1 | 10 (13) | 102 | 12 | 78 |
| kind partitions with control flow or extra conditions | 105 | 3 | 3 | 0 (0) | 102 | 0 | 87 |
| property presence or structural reads | 40 | 10 | 10 | 8 (41) | 30 | 0 | 20 |
| flags and bit masks | 24 | 0 | 0 | 0 (0) | 24 | 0 | 22 |
| other value, generic, assertion or erased claims | 13 | 4 | 4 | 0 (0) | 9 | 0 | 5 |
| Total | 580 | 309 | 18 | 18 (54) | 271 | 291 | 495 |

The .a diagnostic columns partition all 580 with the admitted-proof column. The .ts and .a columns are different source modes and overlap. The .ts partition is 17 bodies with constructed proven calls, 18 with constructed checks, 495 pending on views, five pending on other call construction, 44 without a resolved direct call, and one proved body whose calls remain pending. No pending or unobserved body is counted as a checked or executed pass.

An independent resolved-type query covers all 4,383 calls and identifies 252 object-intersection targets. It moves 39 predicate declarations into the view dependency category without changing any body proof or check construction result. The six groups retain the approved source-AST classification. The structural-read group also includes scalar SyntaxKind comparisons: the group name does not establish a structural proof. Complete original types, optional fields and callable members remain intact.

## Where checks are constructed

Every successful check site follows. The complete 580-row index is [corpus-sites.md](corpus-sites.md); exact bodies, all 4,383 calls, both diagnostic categories and branch/source/target reports are in [evidence/corpus-results.json.gz](evidence/corpus-results.json.gz).

| Predicate and original annotation | Original call | Direction records | Source and target |
| --- | --- | --- | --- |
| `isMultiplicativeOperatorOrHigher` `factory/utilities.ts:1181:62` | `factory/utilities.ts:1193:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 350 more ... \| SyntaxKind.Count -> MultiplicativeOperatorOrHigher` |
| `isAdditiveOperatorOrHigher` `factory/utilities.ts:1191:56` | `factory/utilities.ts:1205:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 349 more ... \| SyntaxKind.Count -> AdditiveOperatorOrHigher` |
| `isShiftOperatorOrHigher` `factory/utilities.ts:1203:60` | `factory/utilities.ts:1219:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 346 more ... \| SyntaxKind.Count -> ShiftOperatorOrHigher` |
| `isShiftOperatorOrHigher` `factory/utilities.ts:1203:60` | `utilities.ts:4931:58` | true checked, false unobservable | `BinaryOperator -> ShiftOperatorOrHigher` |
| `isRelationalOperatorOrHigher` `factory/utilities.ts:1217:58` | `factory/utilities.ts:1231:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 348 more ... \| SyntaxKind.Count -> RelationalOperatorOrHigher` |
| `isEqualityOperatorOrHigher` `factory/utilities.ts:1229:56` | `factory/utilities.ts:1242:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 349 more ... \| SyntaxKind.Count -> EqualityOperatorOrHigher` |
| `isBitwiseOperatorOrHigher` `factory/utilities.ts:1240:55` | `factory/utilities.ts:1253:12` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 350 more ... \| SyntaxKind.Count -> BitwiseOperatorOrHigher` |
| `isLogicalOperatorOrHigher` `factory/utilities.ts:1251:55` | `factory/utilities.ts:1258:12` | true checked, false checked | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 351 more ... \| SyntaxKind.Count -> LogicalOperatorOrHigher` |
| `isAssignmentOperatorOrHigher` `factory/utilities.ts:1256:58` | `factory/utilities.ts:1263:12` | true checked, false checked | `SyntaxKind -> AssignmentOperatorOrHigher` |
| `isBinaryOperator` `factory/utilities.ts:1262:46` | `factory/utilities.ts:1268:12` | true checked, false unobservable | `SyntaxKind -> BinaryOperator` |
| `isCompoundAssignment` `transformers/utilities.ts:490:61` | `checker.ts:41057:21` | true checked, false unobservable | `BinaryOperator -> CompoundAssignmentOperator` |
| `isCompoundAssignment` `transformers/utilities.ts:490:61` | `transformers/classFields.ts:1585:29` | true checked, false checked | `AssignmentOperator -> CompoundAssignmentOperator` |
| `isCompoundAssignment` `transformers/utilities.ts:490:61` | `transformers/classFields.ts:1662:13` | true checked, false checked | `AssignmentOperator -> CompoundAssignmentOperator` |
| `isCompoundAssignment` `transformers/utilities.ts:490:61` | `transformers/esDecorators.ts:1803:25` | true checked, false checked | `AssignmentOperator -> CompoundAssignmentOperator` |
| `isCompoundAssignment` `transformers/utilities.ts:490:61` | `transformers/generators.ts:826:17` | true checked, false checked | `BinaryOperator -> CompoundAssignmentOperator` |
| `isKeyword` `utilities.ts:5266:47` | `emitter.ts:2027:13` | true checked, false checked | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `expressionToTypeNode.ts:1249:13` | true checked, false checked | `TypeNodeSyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:2209:13` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:3442:24` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:3458:24` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:7908:18` | true checked, false checked | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:8612:40` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `parser.ts:8690:40` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `transformers/utilities.ts:474:9` | true checked, false checked | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 351 more ... \| SyntaxKind.Count -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `utilities.ts:5277:12` | true checked, false checked | `SyntaxKind -> KeywordSyntaxKind` |
| `isKeyword` `utilities.ts:5266:47` | `utilities.ts:5287:12` | true checked, false unobservable | `SyntaxKind -> KeywordSyntaxKind` |
| `isPunctuation` `utilities.ts:5271:51` | `utilities.ts:5277:32` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 268 more ... \| SyntaxKind.Count -> PunctuationSyntaxKind` |
| `isKeywordOrPunctuation` `utilities.ts:5276:60` | `parser.ts:2493:22` | true checked, false unobservable | `JSDocSyntaxKind -> PunctuationOrKeywordSyntaxKind` |
| `isKeywordOrPunctuation` `utilities.ts:5276:60` | `parser.ts:2549:22` | true checked, false unobservable | `JSDocSyntaxKind -> PunctuationOrKeywordSyntaxKind` |
| `isKeywordOrPunctuation` `utilities.ts:5276:60` | `parser.ts:5775:30` | true checked, false unobservable | `SyntaxKind -> PunctuationOrKeywordSyntaxKind` |
| `isLogicalOrCoalescingBinaryOperator` `utilities.ts:7425:73` | `binder.ts:1948:17` | true checked, false checked | `BinaryOperator -> SyntaxKind.QuestionQuestionToken \| LogicalOperator` |
| `isLogicalOrCoalescingBinaryOperator` `utilities.ts:7425:73` | `checker.ts:40538:21` | true checked, false checked | `BinaryOperator -> SyntaxKind.QuestionQuestionToken \| LogicalOperator` |
| `isLogicalOrCoalescingBinaryOperator` `utilities.ts:7425:73` | `utilities.ts:7431:40` | true checked, false unobservable | `BinaryOperator -> SyntaxKind.QuestionQuestionToken \| LogicalOperator` |
| `isTypeNodeKind` `utilities.ts:8330:51` | `checker.ts:26125:52` | true checked, false unobservable | `KeywordSyntaxKind -> TypeNodeSyntaxKind` |
| `isTypeNodeKind` `utilities.ts:8330:51` | `utilitiesPublic.ts:1816:12` | true checked, false unobservable | `SyntaxKind -> TypeNodeSyntaxKind` |
| `isLiteralKind` `utilitiesPublic.ts:1494:50` | `factory/parenthesizerRules.ts:216:33` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral` |
| `isLiteralKind` `utilitiesPublic.ts:1494:50` | `factory/parenthesizerRules.ts:268:13` | true checked, false checked | `SyntaxKind -> SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral` |
| `isLiteralKind` `utilitiesPublic.ts:1494:50` | `factory/parenthesizerRules.ts:278:33` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral` |
| `isLiteralKind` `utilitiesPublic.ts:1494:50` | `parser.ts:3770:13` | true checked, false checked | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 346 more ... \| SyntaxKind.Count -> SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral` |
| `isLiteralKind` `utilitiesPublic.ts:1494:50` | `utilitiesPublic.ts:1499:12` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral` |
| `isTemplateLiteralKind` `utilitiesPublic.ts:1518:58` | `emitter.ts:2122:59` | true checked, false checked | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 351 more ... \| SyntaxKind.Count -> SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail` |
| `isTemplateLiteralKind` `utilitiesPublic.ts:1518:58` | `parser.ts:2629:13` | true checked, false checked | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 351 more ... \| SyntaxKind.Count -> SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail` |
| `isTemplateLiteralKind` `utilitiesPublic.ts:1518:58` | `parser.ts:3762:22` | true checked, false checked | `SyntaxKind -> SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail` |
| `isTemplateLiteralKind` `utilitiesPublic.ts:1518:58` | `utilitiesPublic.ts:1523:12` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail` |
| `isTemplateLiteralKind` `utilitiesPublic.ts:1518:58` | `utilitiesPublic.ts:1571:54` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 351 more ... \| SyntaxKind.Count -> SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:2808:16` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:3996:13` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:4005:61` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:4227:13` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:4304:16` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:4864:13` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:5301:17` | true checked, false unobservable | `SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| ... 348 more ... \| SyntaxKind.Count -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `parser.ts:7874:16` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |
| `isModifierKind` `utilitiesPublic.ts:1612:52` | `utilitiesPublic.ts:1648:12` | true checked, false unobservable | `SyntaxKind -> SyntaxKind.ConstKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| ... 7 more ... \| SyntaxKind.OverrideKeyword` |

## Remaining stops

For .a admission, 289 proved bodies stop at a checked field alias requiring an optional, accessor, or representation conversion; two stop at a getter in a checked field contract. All 279 proved direct-kind bodies remain pending on those complete views. The other four direct-kind bodies remain refused: isJSDocPropertyLikeTag, isClassElement and isTypeElement have unsupported target contracts; isPropertyName fails the false-direction proof.

The isPropertyName observation is a false-return refusal at utilitiesPublic.ts:1659:5. Its pinned PropertyName declaration includes NoSubstitutionTemplateLiteral and BigIntLiteral, while its body omits both tags. Inferring a restricted producer domain would require separate evidence; this measurement retains the declared domain and refusal.

The exceptional proved body is Debug.assert at debug.ts:213:142. It has the condition-only annotation asserts expression, without a declared target T. Its body admission succeeds, but all 445 resolved calls stop at a checked predicate without a reifiable target. They are recorded as pending calls rather than 445 retired or checked sites. The 142 other proved calls produce 284 proven direction records. There are no successfully constructed assertion checks in this corpus; the focused checked-assertion fixture below still passes.

Other named stops include negative membership over an open tag domain, unreified complete object targets, indirect predicate calls and unknown generic parameter representations. Untagged positive/negative views, optional reads and unbuilt descriptors stay pending, including facilities on another branch. No branch providing them was merged. Callback bodies remain in the 580; indirect invocation through a bodyless callback contract is not a resolved direct call to its producer, so the 44 unobserved bodies supply no execution claim.

## Main and checks

After the corpus measurement, git fetch origin and git merge --no-edit origin/main returned 'Already up to date.'. Current main 031a1259bc7973934792dc6cb1bd4074fc2204b9 is already an ancestor of 90a78f49. No new train or worker branch was merged, and main and area branches were not pushed or merged into.

Focused checks after that merge:

- go test ./internal/lower -run 'Predicate|TestConditionAssertionAdmission|TestEveryNeedsCallbackEffects' -count=1 -v: PASS 8.347s.
- go test ./internal/oracle -run '^TestCheckedPredicate(Oracle|CountsAreRecorded)$|^TestPredicateKind(ProofOracle|OptionalReadPending|CountsAreRecorded)$' -count=1 -v: PASS 10.214s; optional_read is explicitly SKIP pending on the named optional-field conversion. Source Node, sanitized native, release native, JavaScript, leak and balance checks run through the existing focused fixtures.
- go test ./cmd/adamic -run '^TestExplainChecksOutput$' -count=1: PASS 1.697s.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts: PASS 38.535s. No count row changes; this measurement adds no fixture.
- Corpus join, source-hash/body/call audit, and all fourteen independent measurement mutants pass.

Six corpus corruptions are caught: dropped body, dropped call, invented proof, failed call labelled checked, pending view labelled proven, and changed source hash. Five independent shard corruptions are caught: dropped shard, duplicated call, changed body proof, call moved to another shard, and dropped checker diagnostic. Three target-metadata corruptions are also caught: dropped metadata, changed call coordinate and a non-boolean view flag. Their exact assertion catchers are retained in the three mutant logs.

The focused oracle also reruns five removed-check mutants (true, unused true, false, boolean-aliased false and assertion) and three default-false proof mutants (enum, immutable alias and optional target), caught in native and JavaScript against Node. The focused lower run retains the six lying kind-body refusals and readonly-getter negative. The independent compiler getter-purity and open-domain mutations were recorded in the prior item; they were not reinstalled in this measurement-only item.

Setup succeeds with GOPROXY=https://proxy.golang.org|direct. Timing lines: Node 0.024s; Go 0.026s; markdown skipped step 0.009s, ready 0.069s; submodules 0.070s; clang 0.156s; go build 36.044s; test binaries deferred 36.154s; cache warm 36.155s; done 36.183s. nproc is 5, cpu.max is 400000 100000; Go 1.27.1, Node 24.19.0, clang 20.1.8. All outputs are saved directly to logs and retained compressed under evidence/corpus-*. No full gate or whole-package test was run. No cohere code was copied.

## Design questions

- How should a proved condition-only asserts expression call bypass target-T reification and count its proof? Debug.assert supplies a concrete 445-call pending case; the current ruling names asserts x is T.
- Which complete untagged and overlapping-tag contracts can be tested negatively without consuming a view or panicking during the membership probe?
- How should receiver/this, indirect calls, generic overloads and escaping callbacks preserve argument identity and reifiable contracts?
- Which generic, recursive and callable descriptors carry enough producer proof for the complete target?
- What representation conversion admits optional/accessor aliases while preserving missing versus present undefined? It blocks 289 logically proved corpus bodies.
- Which effects and getters invalidate tag facts, and which tag/flag/producer domains justify both directions? isPropertyName demonstrates why an omitted target tag needs evidence rather than an annotation.

## Reproduce

Rebuild the adapted tree with stage3/apply.sh from 8a7ab17e and verify it through the audit below. Use a fresh scratch directory for the archive and adapted tree. These commands assume that tree is /workspace/scratch/predicates-corpus. The probe can run serially with unset HATCH_SHARDS/HATCH_SHARD, or as the four complete independent shards below. Each call has a stable ordinal; the join asserts exactly one observation per ordinal, identical body results, and identical checker diagnostics. The .a proof probe changes only the loader source-mode query for the original typed tree; source bytes, checker options and original coordinates are preserved. It is not a second source adaptation.

```sh
source /workspace/adamic-tools/env.sh
node stage3/scouts/step09/predicates/group.cjs stage3/scouts/step09/predicates/evidence/inventory.json.gz stage3/scouts/step09/predicates/evidence/main-results.json.gz /tmp/predicates-all-groups.json --all-bodies > /tmp/predicates-all-groups.log 2>&1
python3 stage3/scouts/step09/predicates/measure-overlay.py /workspace/adamic /workspace/scratch/predicates-overlay > /tmp/predicates-overlay.log 2>&1
go build -tags hatch_predicate_measurement -buildvcs=false -overlay=/workspace/scratch/predicates-overlay/overlay.json -o /workspace/scratch/predicates-measure ./stage3/scouts/step09/predicates/measure-probe > /tmp/predicates-build.log 2>&1
HATCH_TARGET_ONLY=1 /workspace/scratch/predicates-measure /workspace/scratch/predicates-corpus > /tmp/predicates-corpus-targets.json 2> /tmp/predicates-corpus-targets.log
predicate_workers=()
for predicate_shard in 0 1 2 3; do
  HATCH_SHARDS=4 HATCH_SHARD="$predicate_shard" /workspace/scratch/predicates-measure /workspace/scratch/predicates-corpus > "/tmp/predicates-corpus-shard-$predicate_shard.json" 2> "/tmp/predicates-corpus-shard-$predicate_shard.log" &
  predicate_workers+=("$!")
done
for predicate_worker in "${predicate_workers[@]}"; do wait "$predicate_worker"; done
python3 stage3/scouts/step09/predicates/merge-corpus-shards.py /tmp/predicates-corpus-current.json /tmp/predicates-corpus-shard-0.json /tmp/predicates-corpus-shard-1.json /tmp/predicates-corpus-shard-2.json /tmp/predicates-corpus-shard-3.json > /tmp/predicates-corpus-merge.log 2>&1
python3 stage3/scouts/step09/predicates/merge-target-metadata.py /tmp/predicates-corpus-current.json /tmp/predicates-corpus-targets.json /tmp/predicates-corpus-current-targets.json > /tmp/predicates-corpus-target-merge.log 2>&1
python3 stage3/scouts/step09/predicates/audit-corpus.py stage3/scouts/step09/predicates/evidence/inventory.json.gz /tmp/predicates-all-groups.json /tmp/predicates-corpus-current-targets.json /workspace/scratch/predicates-corpus /tmp/predicates-corpus-report.json > /tmp/predicates-corpus-audit.log 2>&1
python3 stage3/scouts/step09/predicates/audit-corpus-mutants.py stage3/scouts/step09/predicates/evidence/inventory.json.gz /tmp/predicates-all-groups.json /tmp/predicates-corpus-current-targets.json /workspace/scratch/predicates-corpus > /tmp/predicates-corpus-audit-mutants.log 2>&1
python3 stage3/scouts/step09/predicates/merge-corpus-shards-mutants.py /tmp/predicates-corpus-shard-0.json /tmp/predicates-corpus-shard-1.json /tmp/predicates-corpus-shard-2.json /tmp/predicates-corpus-shard-3.json > /tmp/predicates-corpus-shard-mutants.log 2>&1
python3 stage3/scouts/step09/predicates/merge-target-metadata-mutants.py /tmp/predicates-corpus-current.json /tmp/predicates-corpus-targets.json > /tmp/predicates-corpus-target-mutants.log 2>&1
```
