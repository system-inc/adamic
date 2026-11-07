# getAllKeys reach, task b5w3ycg

Complete: all 398 lane units passed their ordinary guards in 20 pieces.
There were **6566 getAllKeys calls**, all from the harness
caller at `src/harness/fourslashImpl.ts:461`, and **0 runtime
reads of the getAllKeys property on the ts namespace objects**.
**6566 distinct selected lane cases** reached the helper.
No other caller was observed. No adaptation or implementation change was made.
Only this report is committed.

Base: origin/main `837da8ed9a306db11ca0f0a39f396111048c5040`.
Pinned TypeScript: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
The requested historical census base was
`af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`; its cited F10 report is not
present on this origin/main. This measurement uses the completed unchanged
full apply from the frozen base above. It is not a rerun of that census.

## Calls, reads and cases

| Observed caller, source-mapped from emitted code | Calls |
| --- | ---: |
| `at memoWrap (adapted-tree/src/harness/fourslashImpl.ts:461:29)` | 6566 |
| Any other caller | 0 |

| Lane runner | Harness calls | Distinct cases reaching the caller |
| --- | ---: | ---: |
| fourslash | 6322 | 6322 |
| fourslash-server | 244 | 244 |

The calls-per-case histogram is `{'1': 6566}`. Every observed
call was matched to that unit's actually selected test identity, not merely
to a file name. The native inventory has 6,323 whole-file tasks, including
`fourslash.ts`, which the runner explicitly excludes from test execution.
Server files also have a separate tsserver-log baseline case; those log cases
do not construct another TestState or call this helper.

The source spelling is `ts.getAllKeys(ls)`. The built harness has
`const keys = getAllKeys(ls);` at uninstrumented `built/local/run.js:215031`:
the bundler resolved the namespace access to a direct function call. Therefore
executions of the source call do not entail runtime ts-property reads. The
property-read count measures executed getters, including computed and aliased
access, rather than inferring a read from the TypeScript spelling.

All five emitted namespace getter sites were observed independently of the
two function-body entry counters:

| Runtime namespace getter site, before observation | Reads |
| --- | ---: |
| `run.js:197692` | 0 |
| `run.js:201006` | 0 |
| `run.js:210171` | 0 |
| `typescript.js:655` | 0 |
| `typescript.js:184761` | 0 |

The same helper is present in emitted `typescript.js:3055` and `run.js:791`.
The complete observed caller stacks identify the harness call above. Positive
seeds outside the oracle detected exactly five calls and four namespace reads,
including a computed property read, two calls through one saved alias, and
VM createContext, runInNewContext and Script.runInNewContext paths. The seeds
also checked inherited nonenumerable names and an object with null prototype.
They are excluded from every lane count in this report.

## What the harness passes and does with the keys

`TestState` obtains an adapter at fourslashImpl.ts:386 and calls
`memoWrap(languageServiceAdapter.getLanguageService(), this)` at line 389.
NativeLanguageServiceAdapter.getLanguageService returns
`ts.createLanguageService(...)` (pinned harnessLanguageService.ts:456); the server
adapter returns its `SessionClient` (line 803).

The observed chains contain exactly two descriptor shapes:

| Argument | Calls | Own names at successive non-null chain levels | Returned unique names |
| --- | ---: | --- | ---: |
| Native language service | 6322 | 73 + 12 | 85 |
| Server SessionClient | 244 | 7 + 105 + 12 | 123 |

The native chain is `language-service object -> Object.prototype -> null`.
The server chain is
`SessionClient instance -> SessionClient.prototype -> Object.prototype -> null`.
The server instance's seven own fields are host, sequence, lineMaps, messages,
getCombinedCodeFix, applyCodeActionCommand and mapCode. SessionClient.prototype
contributes nonenumerable methods, not just the instance's own fields.
Object.prototype contributes its twelve own nonenumerable names; constructor
is deduplicated in the server chain because SessionClient.prototype already
contributed that name. Symbols are excluded by the helper itself.

For every call, the raw observation retains each chain level's ordered own
string names and descriptor kinds/enumerability, and the returned ordered key
list. Independent reconciliation reconstructs first-occurrence order from the
saved descriptors and compares every returned element. All calls agree, with
no missing return or order/projection mismatch.

The result drives the loop at fourslashImpl.ts:462. For each returned name,
memoWrap assigns `proxy[key]` a closure forwarding to `ls[key](...args)`.
Five names instead use Utils.memoize: getCompletionEntryDetails,
getCompletionEntrySymbol, getQuickInfoAtPosition, getReferencesAtPosition and
getDocumentHighlights. Their memoization key includes script version, active
file, caret, selection end, last marker and arguments. The proxy is returned
as `this.languageService`. Thus the result supplies both own and inherited
names for forwarding, including the server methods stored on the prototype.

## Full-lane observation and guards

The unchanged `stage3/lane/run.sh` oracle inventory was executed through the
same 398-unit shard plan used by the getOwnValues record-table measurement,
in 20 pieces of 20 units (18 in the last piece), with four units concurrent.
The lane runtime, shard plan, API parser, baseline sanction and expected counts
were not edited. Each unit retained its ordinary identity, API, baseline,
count and upstream-exit guards. The measurement allowed 90 seconds per unit,
as the referenced measurement did; this is not a new 30-second gate timing
claim. Every unit completed without timeout.

Totals: **106366 passing, 1 expected
failing, 0 skipped**. The one permitted failure is the
ordinary Public APIs declaration acknowledgement. There is no other failure.
All 398 unit reports passed; every selected identity was checked exactly once.

| Piece | Units | Oracle pass / fail / skip | Calls | Namespace reads | Seconds | Piece child exit |
| --- | ---: | --- | ---: | ---: | ---: | ---: |
| 1 | 20 | 534 / 1 / 0 | 2 | 0 | 57.217 | 0 |
| 2 | 20 | 493 / 0 / 0 | 1 | 0 | 39.704 | 0 |
| 3 | 20 | 3748 / 0 / 0 | 316 | 0 | 60.289 | 0 |
| 4 | 20 | 5736 / 0 / 0 | 407 | 0 | 66.278 | 0 |
| 5 | 20 | 6154 / 0 / 0 | 379 | 0 | 65.627 | 0 |
| 6 | 20 | 6856 / 0 / 0 | 382 | 0 | 63.331 | 0 |
| 7 | 20 | 6694 / 0 / 0 | 393 | 0 | 70.236 | 0 |
| 8 | 20 | 6612 / 0 / 0 | 389 | 0 | 69.925 | 0 |
| 9 | 20 | 6399 / 0 / 0 | 439 | 0 | 62.331 | 0 |
| 10 | 20 | 6666 / 0 / 0 | 415 | 0 | 64.317 | 0 |
| 11 | 20 | 6633 / 0 / 0 | 437 | 0 | 63.980 | 0 |
| 12 | 20 | 7162 / 0 / 0 | 414 | 0 | 66.963 | 0 |
| 13 | 20 | 7061 / 0 / 0 | 400 | 0 | 62.874 | 0 |
| 14 | 20 | 7631 / 0 / 0 | 442 | 0 | 66.709 | 0 |
| 15 | 20 | 7697 / 0 / 0 | 459 | 0 | 67.359 | 0 |
| 16 | 20 | 6726 / 0 / 0 | 509 | 0 | 64.189 | 0 |
| 17 | 20 | 6891 / 0 / 0 | 423 | 0 | 62.667 | 0 |
| 18 | 20 | 6408 / 0 / 0 | 359 | 0 | 64.401 | 0 |
| 19 | 20 | 226 / 0 / 0 | 0 | 0 | 39.369 | 0 |
| 20 | 18 | 39 / 0 / 0 | 0 | 0 | 31.160 | 0 |

Observation changed only the scratch emitted JavaScript: entry and return
counters in the two named getAllKeys functions, plus counters in their five
namespace export getters. Every original function return value was returned
unchanged. No implementation body or adaptation was replaced. The preload
associated calls with Mocha runnable identities and upstream worker tasks,
and supplied observer callbacks to VM contexts. It did not replace
Object.getPrototypeOf or Object.getOwnPropertyNames. Prototype descriptors
were inspected without invoking their accessors.

The module-load audit saw 398 instrumented helper
loads. The only uninstrumented helper-bearing module loaded was the pinned
stage3/api parser in the ordinary external declaration guard; it was not an
oracle compiler instance. All 797 observed processes have the
start/exit evidence recorded in the raw trace; incomplete-process records:
`[]`. Function-entry counts, returned observations and
process-exit counters reconcile exactly. No unknown caller or unattributed
case remains.

## References in the complete pinned and adapted trees

The compiler source's only spelling is the declaration, at pinned
`src/compiler/core.ts:1299`, adapted `src/compiler/core.ts:1309`.
The sole real source caller is the unchanged harness reference at line 461.
`src/compiler/_namespaces/ts.ts:4` reexports core with export-star;
`src/harness/_namespaces/ts.ts:3` reexports that compiler namespace.
The typescript entrypoint similarly exports its ts namespace. There is no
literal `ts["getAllKeys"]` read or another named helper caller in the checkout.
Runtime-computed property access is covered by the independent getter counts.

The complete search includes hidden and ignored files, source, scripts,
harness, tests/cases, all baselines, shipped lib and the completed build.
Only .git data and dependency installs are excluded. Every matching pinned
file and line follows. A parenthesized count records multiple spellings on
one line; an export getter is a static definition/reference, not evidence
that the getter executed. Scripts have no occurrence.

| Pinned checkout path | Matching lines |
| --- | --- |
| `lib/lib.dom.d.ts` | 22227, 22229, 22231, 22409, 22411, 22413 |
| `lib/lib.webworker.d.ts` | 6790, 6792, 6794, 6972, 6974, 6976 |
| `lib/typescript.js` | 642 (2 spellings), 3038, 184594 (2 spellings), 199392 |
| `src/compiler/core.ts` | 1299 |
| `src/harness/fourslashImpl.ts` | 461 |
| `src/lib/dom.generated.d.ts` | 22211, 22213, 22215, 22393, 22395, 22397 |
| `src/lib/webworker.generated.d.ts` | 6774, 6776, 6778, 6956, 6958, 6960 |
| `tests/baselines/reference/genericClassesRedeclaration.errors.txt` | 18, 32, 45, 63, 77, 90 |
| `tests/baselines/reference/genericClassesRedeclaration.js` | 10, 22, 33, 49, 61, 72 |
| `tests/baselines/reference/genericClassesRedeclaration.symbols` | 25, 26 (2 spellings), 87, 88 (2 spellings), 147, 148 (2 spellings), 216, 217 (2 spellings), 278, 279 (2 spellings), 338, 339 (2 spellings) |
| `tests/baselines/reference/genericClassesRedeclaration.types` | 18, 19, 102, 103, 186, 187, 270, 271, 354, 355, 438, 439 |
| `tests/baselines/reference/parserRealSource4.errors.txt` | 31, 45, 130, 131 (2 spellings) |
| `tests/baselines/reference/parserRealSource4.js` | 27, 41, 126, 127 (2 spellings), 326, 403, 404 (2 spellings) |
| `tests/baselines/reference/parserRealSource4.symbols` | 55, 56 (2 spellings), 111, 112 (2 spellings), 374, 375 (2 spellings), 377 (2 spellings), 378, 379 (2 spellings), 383 (2 spellings), 385 (2 spellings), 389 (2 spellings) |
| `tests/baselines/reference/parserRealSource4.types` | 76, 77, 173, 174, 661, 662, 665 (2 spellings), 666 (2 spellings), 668, 670, 672, 680, 684, 686, 694 |
| `tests/cases/compiler/genericClassesRedeclaration.ts` | 8, 20, 31, 47, 59, 70 |
| `tests/cases/conformance/parser/ecmascript5/parserRealSource4.ts` | 26, 40, 125, 126 (2 spellings) |

The adapted tree has all these same paths and line lists except for core.ts
below. The other fifteen matching pinned files are byte-identical. Existing adaptations also change harness types and unrelated harness code; its getAllKeys reference stays at line 461. Its
additional build outputs and changed core declaration position are:

| Adapted path | Matching lines |
| --- | --- |
| `built/local/compiler/core.d.ts` | 432 |
| `built/local/lib.dom.d.ts` | 22227, 22229, 22231, 22409, 22411, 22413 |
| `built/local/lib.webworker.d.ts` | 6790, 6792, 6794, 6972, 6974, 6976 |
| `built/local/run.js` | 791, 197692 (2 spellings), 201006 (2 spellings), 210171 (2 spellings), 215031 |
| `built/local/typescript.internal.d.ts` | 7348 |
| `built/local/typescript.js` | 655 (2 spellings), 3055, 184761 (2 spellings), 199584 |
| `src/compiler/core.ts` | 1309 |

The browser-library spellings name unrelated IndexedDB IDBIndex and
IDBObjectStore methods and their documentation. The parserRealSource4 and
genericClassesRedeclaration test families name unrelated collection methods;
their JavaScript, error, symbol and type baselines repeat those names. They
are not compiler-helper callers. This inventory retains them explicitly.

## Declaration and runtime export surface

The completed uninstrumented build has
`function getAllKeys(obj: object): string[];` in
`built/local/typescript.internal.d.ts:7348` and the intermediate
`built/local/compiler/core.d.ts:432`. It has no occurrence in either public
package surface, typescript.d.ts or tsserverlibrary.d.ts. The latter reexports
the public typescript package. The shipped pinned public declarations and
public API reference baseline likewise have no occurrence. There is no
checked-in typescript.internal.d.ts in the pristine pin; that file is a build
output. Its internal-only package declaration follows the @internal tag and
the declaration bundler's WriteTarget.Internal handling.

Both pinned shipped lib/typescript.js and the fresh built typescript.js export
getAllKeys on the runtime ts object. The pinned shipped namespace getters are
at 642 and 184594, the body at 3038, and its generated CommonJS export
annotation at 199392. Node inspection confirmed own=true, type=function,
getter=`() => getAllKeys`; the positive seeds separately exercised the fresh
build's exported function. Public declaration omission does not remove the
runtime export.

## Remaining lowering stop

Reading `internal/lower/library_object.go` establishes that
`Object.getPrototypeOf` is explicitly Refused by objectCallArguments.
`Object.getOwnPropertyNames` is not admitted: it has no handled case and falls
through to the final `l.notYet(node, "Object."+name)`, producing
`NotYet: Object.getOwnPropertyNames` for this non-record object parameter.
The record path (`internal/lower/records.go:396`) does not admit that operation
either. `internal/lower/control.go:78-100` lowers the do/while condition before
its body. This explains why the condition's getPrototypeOf refusal precedes
the body's enumeration stop. Admitting only getPrototypeOf would still leave
the getOwnPropertyNames stop at the start of the body.

`pushIfUnique` is an ordinary generic TypeScript helper, not an Object library
operation with a named admission case in library_object.go. Its pinned body
at core.ts:1003 calls contains, then array.push if absent. Generic helper calls
use ordinary instantiation (`internal/lower/expression.go:1389` and generic.go);
array.push is admitted (`internal/lower/library_array.go:13,25`). No named
pushIfUnique refusal exists in the lowerer. At this site its inferred element
is string. This source reading does not prove an end-to-end successful
specialization of its whole transitive body, and no isolated native-lowering specialization probe was run.
It does prove that getOwnPropertyNames remains an earlier definite stop.

## Provenance, restoration and retained evidence

The earlier approval block was resolved by the user's explicit clarification:
unchanged full apply and ordinary compiler/harness builds are authorized
measurement setup; no adaptation or implementation change may land.
The initial stopped-state report and the post-apply checkpoint were plainly
pushed on this same branch. This report replaces their pending conclusions.

Sourced `/workspace/adamic-tools/env.sh` and followed the setup in
`stage3/adapt/99-correctness-axolotl-w/PROOF.md`. All logs went to files under
`/workspace/scratch/getallkeys`; none was piped into head or tail. Wrappers
record the child's actual exit separately from timeout state, and clean its
process group. Ordinary setup/build phase results are:

| Phase | Child exit | Seconds |
| --- | ---: | ---: |
| setup-resumed | 0 | 301.036 |
| apply | 0 | 588.272 |
| build | 0 | 41.085 |
| harness | 0 | 1.761 |
| instrument | 0 | 5.624 |
| seed | 0 | 0.923 |
| lane-pieces | 0 | 1216.693 |

Observers have been removed from the measurement tree. All 746 adapted source
files match the frozen source manifest. Both instrumented JavaScript files
have been restored byte for byte to the ordinary build, and the final
reference inventory's original hashes agree. There are no active measurement
processes. No Go file was authored. No compiler product, observer, driver,
source change, archive or log is committed; the branch carries this report.

| Emitted file | Ordinary-build SHA-256, restored | Observed-build SHA-256 |
| --- | --- | --- |
| `run.js` | `f895c2d608d08172bd5edd9976313b61fca1b4c5f7cd48a8e5eab26f4c0e1168` | `5df3155eda206fae89528b85172890ec15506e7c6224631ece47780470ab1d74` |
| `typescript.js` | `c43d14deadf479577e391bd537947c555d8abbcf5c8cc070df0e04dd707210d5` | `6b1d2cf85b2e41c44bc9b5ab7ed106e2d823548368117baa5e30a0e5dca6a314` |

The raw per-process traces, per-unit ordinary reports and observations,
per-piece summaries, actual-exit ledgers, exact observer/instrumenter/driver
scripts, complete reference inventory, source manifest and seed evidence
remain in `/workspace/scratch/getallkeys`.
Independent reconciliation reopened every piece and every trace, verified
all selected identities and counters, and reconstructed the chains and
ordered returned names. Aggregate raw-event SHA-256, hashing relative path,
NUL and file bytes in piece/unit/file order:
`db3430c66717f75a3db4c4a09e5adfcdf82b92353ebdbfcc7f65dc45fb8c5e6e`.

Scratch-only evidence archive: `/workspace/scratch/getallkeys/evidence.tar.gz`, 5486 files, 5222461 bytes; SHA-256 `91747813f98de3835452d0a79fae76d3e33139814aa2d6b984075b6f3147680a`. The archive is not committed. Frozen 746-file source-manifest SHA-256: `2e2f4a48630dda0ed4a3e8e98062027bc095ba3ad642cc03866c3a3edbc7feb7`.

## Observed argument descriptor inventories

### Native language service descriptor inventory

Shape SHA-256: `182e498be4d7f42227751e871e5dd83a678b71dafac187a5bf76c3b6226fe124`.

Level 0: 73 own string names, 73 enumerable; types {'function': 73}. Ordered names:

```text
dispose, cleanupSemanticCache, getSyntacticDiagnostics, getSemanticDiagnostics
getRegionSemanticDiagnostics, getSuggestionDiagnostics, getCompilerOptionsDiagnostics, getSyntacticClassifications
getSemanticClassifications, getEncodedSyntacticClassifications, getEncodedSemanticClassifications, getCompletionsAtPosition
getCompletionEntryDetails, getCompletionEntrySymbol, getSignatureHelpItems, getQuickInfoAtPosition
getDefinitionAtPosition, getDefinitionAndBoundSpan, getImplementationAtPosition, getTypeDefinitionAtPosition
getReferencesAtPosition, findReferences, getFileReferences, getDocumentHighlights
getNameOrDottedNameSpan, getBreakpointStatementAtPosition, getNavigateToItems, getRenameInfo
getSmartSelectionRange, findRenameLocations, getNavigationBarItems, getNavigationTree
getOutliningSpans, getTodoComments, getBraceMatchingAtPosition, getIndentationAtPosition
getFormattingEditsForRange, getFormattingEditsForDocument, getFormattingEditsAfterKeystroke, getDocCommentTemplateAtPosition
isValidBraceCompletionAtPosition, getJsxClosingTagAtPosition, getLinkedEditingRangeAtPosition, getSpanOfEnclosingComment
getCodeFixesAtPosition, getCombinedCodeFix, applyCodeActionCommand, organizeImports
getEditsForFileRename, getEmitOutput, getNonBoundSourceFile, getProgram
getCurrentProgram, getAutoImportProvider, updateIsDefinitionOfReferencedSymbols, getApplicableRefactors
getEditsForRefactor, getMoveToRefactoringFileSuggestions, toLineColumnOffset, getSourceMapper
clearSourceMapperCache, prepareCallHierarchy, provideCallHierarchyIncomingCalls, provideCallHierarchyOutgoingCalls
toggleLineComment, toggleMultilineComment, commentSelection, uncommentSelection
provideInlayHints, getSupportedCodeFixes, preparePasteEditsForFile, getPasteEdits
mapCode
```

Level 1: 12 own string names, 0 enumerable; types {'function': 11, 'accessor': 1}. Ordered names:

```text
constructor, __defineGetter__, __defineSetter__, hasOwnProperty
__lookupGetter__, __lookupSetter__, isPrototypeOf, propertyIsEnumerable
toString, valueOf, __proto__, toLocaleString
```

### Server SessionClient descriptor inventory

Shape SHA-256: `47f09afbe40b3822dedfb612a8fd1f3a947d34701750d3257d21ac4c294124ae`.

Level 0: 7 own string names, 7 enumerable; types {'object': 3, 'number': 1, 'function': 3}. Ordered names:

```text
host, sequence, lineMaps, messages
getCombinedCodeFix, applyCodeActionCommand, mapCode
```

Level 1: 105 own string names, 0 enumerable; types {'function': 105}. Ordered names:

```text
constructor, onMessage, writeMessage, getLineMap
lineOffsetToPosition, positionToOneBasedLineOffset, convertCodeEditsToTextChange, processRequest
processResponse, configure, setFormattingOptions, setCompilerOptionsForInferredProjects
openFile, closeFile, createChangeFileRequestArgs, changeFile
toLineColumnOffset, getQuickInfoAtPosition, getProjectInfo, getCompletionsAtPosition
getCompletionEntryDetails, getCompletionEntrySymbol, getNavigateToItems, getFormattingEditsForRange
getFormattingEditsForDocument, getFormattingEditsAfterKeystroke, getDefinitionAtPosition, getDefinitionAndBoundSpan
getTypeDefinitionAtPosition, getSourceDefinitionAndBoundSpan, getImplementationAtPosition, findReferences
getReferencesAtPosition, getFileReferences, getEmitOutput, getSyntacticDiagnostics
getSemanticDiagnostics, getSuggestionDiagnostics, getRegionSemanticDiagnostics, getDiagnostics
getCompilerOptionsDiagnostics, getRenameInfo, getSmartSelectionRange, findRenameLocations
decodeNavigationBarItems, getNavigationBarItems, decodeNavigationTree, getNavigationTree
decodeSpan, decodeLinkDisplayParts, getNameOrDottedNameSpan, getBreakpointStatementAtPosition
getSignatureHelpItems, getDocumentHighlights, getOutliningSpans, getTodoComments
getDocCommentTemplateAtPosition, isValidBraceCompletionAtPosition, getJsxClosingTagAtPosition, getLinkedEditingRangeAtPosition
getSpanOfEnclosingComment, getCodeFixesAtPosition, provideInlayHints, createFileLocationOrRangeRequestArgs
createFileLocationRequestArgs, createFileRangeRequestArgs, createFileLocationRequestArgsWithEndLineAndOffset, getApplicableRefactors
getMoveToRefactoringFileSuggestions, getEditsForRefactor, organizeImports, getEditsForFileRename
convertCodeEditsToTextChanges, convertChanges, convertTextChangeToCodeEdit, getBraceMatchingAtPosition
configurePlugin, getIndentationAtPosition, getSyntacticClassifications, getSemanticClassifications
getEncodedSyntacticClassifications, getEncodedSemanticClassifications, convertCallHierarchyItem, prepareCallHierarchy
convertCallHierarchyIncomingCall, provideCallHierarchyIncomingCalls, convertCallHierarchyOutgoingCall, provideCallHierarchyOutgoingCalls
getSupportedCodeFixes, preparePasteEditsForFile, getPasteEdits, getProgram
getCurrentProgram, getAutoImportProvider, updateIsDefinitionOfReferencedSymbols, getNonBoundSourceFile
getSourceFile, cleanupSemanticCache, getSourceMapper, clearSourceMapperCache
toggleLineComment, toggleMultilineComment, commentSelection, uncommentSelection
dispose
```

Level 2: 12 own string names, 0 enumerable; types {'function': 11, 'accessor': 1}. Ordered names:

```text
constructor, __defineGetter__, __defineSetter__, hasOwnProperty
__lookupGetter__, __lookupSetter__, isPrototypeOf, propertyIsEnumerable
toString, valueOf, __proto__, toLocaleString
```
