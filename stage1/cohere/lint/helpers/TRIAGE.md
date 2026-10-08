# Shared lint helper branch triage

Static audit, 2026-10-08 09:36 UTC. No code changes, builds, gates, commits, or pushes. Remote refs were fetched explicitly; all subsequent reads and merge checks use the pinned SHAs below. The worktree remained detached at b455bb4f and was not used as the inventory truth.

## Decision and counting basis

The requested names resolve to **20 branches**, not 19. Whole-tip merge-tree checks against area `c8f6d74f5387df926ae12d8babd2780564b60097` return **17 clean and 3 conflicted**. Main is `ef3141e9b1152ab51b51497f8ce3a2799449c8a3`. Both main and area pin cohere at `7945d102a6c18dd36adf9114a758ce646e8b2359`.

Main’s `stage1/cohere/lint/helpers/readiness.json` is the original 198-rule ledger (46 initially helper-ready). The follow-up is **`stage1/cohere/lint/helpers/comments/readiness.json`**, not `lint/comments/readiness.json`. It contains all 198 rows, of which 62 have `helper_ready:true` and 136 have `helper_ready:false`. I count only those 136 false rows. A rule is credited only when **all** of its `remaining_helpers` are removed; a helper’s consumer count is not an unblock count. The strict-options placeholder remains a dependency where the ledger says so. The `option_helper_ready` flag is retained in the final rule appendix; the ledger itself permits policy-only/comment-ready rules with that flag false, so it is not imposed as a new second blocker.

The candidate catalog has **356 distinct retained symbols**; **356** occur in this remaining ledger. Its union conditionally removes all helper blockers for **56 of 136 rules**, reaching 118/198 helper-ready, with 80 still blocked. There are **330** uncovered distinct ledger dependencies, including the strict-options placeholder and blocked helper witnesses. These are ceilings under the frozen ledger’s common-AST-adapter assumption, not results of a newly run integrated gate. No branch demonstrates complete findings/fixes parity for every consuming rule after combining these ports. Most helper bodies consume projected AST facts or separately supplied callbacks, sometimes actual Go answers during isolated tests. Those contracts must be wired and gated before the ceiling becomes actual readiness.

Recommendation: assemble package-sized landings from the sources below, carrying their tests and narrow oracle adapters, rather than wholesale merging rule-wave history and megabytes of archived evidence. Keep raw evidence available, but do not duplicate it in each package landing. The base comment bundle is already on area; landing its branch again adds no helpers. A clean merge is a syntactic merge result, not a current compiler/Go-parity gate.

## Merge check and volume

Exact command run for each tip: `git -C /workspace/adamic merge-tree --write-tree c8f6d74f5387df926ae12d8babd2780564b60097 <tipSHA>`. Exit 0 means clean; exit 1 means conflicts. The conflict list is read from CONFLICT messages, not guessed from changed-file overlap. Checks apply to whole tips against the same area, not to sequential package landings. They may write unreachable Git objects, but do not change the checkout, index, refs, or commits.

| Branch | Tip | Authored tip date | Merge | Helper implementation/support files, bytes | Test/oracle/capture program files, bytes | Evidence/log/docs/fixture files, bytes (% of helper bytes) | Other-area diff files |
|---|---|---|---|---|---|---|---|
| `origin/codex/lint-helpers` | `95100eb440b47f3f18e960c6f5b49cadf0dc1d9d` | 2026-10-07T02:04:33Z | clean | 6, 11,782 | 5, 20,752 | 24, 1,358,481 (97.7%) | 0 |
| `origin/codex/lint-helpers-01` | `98f7e5f08c10f148d457e3a1fdde0355d1401f6c` | 2026-10-07T18:12:46Z | clean | 35, 40,850 | 45, 201,445 | 186, 5,245,907 (95.6%) | 0 |
| `origin/codex/lint-helpers-02` | `ff5090b777bc4e27bd51a5b57d9c824670f14b26` | 2026-10-07T13:41:58Z | clean | 44, 72,174 | 80, 320,314 | 446, 4,514,283 (92.0%) | 1489 |
| `origin/codex/lint-helpers-03` | `3e85502d728bcd67b3ed326565bcf09749f2ec1b` | 2026-10-07T10:54:38Z | clean | 68, 50,101 | 112, 428,076 | 423, 5,342,133 (91.8%) | 0 |
| `origin/codex/lint-helpers-04` | `8e0d98a645ae73cc4cf72561c779f257b51b8c9a` | 2026-10-07T11:51:32Z | clean | 75, 47,799 | 139, 514,807 | 462, 15,107,943 (96.4%) | 1076 |
| `origin/codex/lint-helpers-05` | `cd24dd580a8e8d97cdc1cf0aa839b47bf333e9bc` | 2026-10-07T21:16:52Z | clean | 115, 101,621 | 149, 797,082 | 1119, 46,501,447 (98.1%) | 0 |
| `origin/codex/lint-helpers-from-codex-lint-wave1-03` | `5e64fe3cf0d236a2fae7e4f8b645f0d3f33afc29` | 2026-10-07T09:23:53Z | clean | 2, 3,194 | 5, 21,278 | 254, 2,947,773 (99.2%) | 0 |
| `origin/codex/lint-helpers-from-codex-lint-wave1-14` | `7b6a7c28bdf9b9e715a8403cf27703a69afeeb74` | 2026-10-07T09:26:39Z | clean | 2, 5,837 | 6, 29,546 | 92, 348,776 (90.8%) | 0 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-02` | `d62bb9d4df13cc6132f0b4bc2b1748d16934edb7` | 2026-10-07T09:24:52Z | conflict C1 | 6, 5,535 | 10, 30,186 | 94, 3,122,189 (98.9%) | 47 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-04` | `8754bdeafe0ef400be14bfb1d8f1505c0c80c713` | 2026-10-07T09:26:02Z | clean | 7, 6,215 | 16, 48,691 | 72, 728,172 (93.0%) | 0 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-06` | `6f2aff526107895a267896fdcbf182cd67c3c7a6` | 2026-10-07T04:18:14Z | conflict C1 | 2, 1,199 | 14, 35,005 | 28, 171,450 (82.6%) | 47 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-09` | `017f34d7fc64fed98740ab5456f516c8cec2931e` | 2026-10-07T09:25:08Z | clean | 3, 6,022 | 13, 47,390 | 105, 652,005 (92.4%) | 47 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-10` | `48c2db8a342ee27a2630f07d0b0923f4a5da0bca` | 2026-10-07T09:24:20Z | clean | 4, 39,855 | 12, 31,724 | 74, 3,990,019 (98.2%) | 0 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-11` | `a6278560ea1d50cbbd9147b048a0c86b7bdf4478` | 2026-10-07T13:19:39Z | clean | 13, 9,153 | 16, 45,527 | 217, 5,391,573 (99.0%) | 1489 |
| `origin/codex/lint-helpers-from-codex/lint-wave1-15` | `298ecc4cb20069615bb68130f177bd6f97dd7869` | 2026-10-07T14:02:47Z | clean | 7, 24,916 | 13, 33,790 | 205, 11,446,405 (99.5%) | 1489 |
| `origin/codex/lint-helpers-from-lint-wave1-01` | `72fca0ebb2abb157f395c2cf91520c04ea26d794` | 2026-10-07T09:27:46Z | clean | 8, 6,296 | 24, 68,246 | 54, 173,380 (69.9%) | 0 |
| `origin/codex/lint-helpers-from-lint-wave1-05` | `376dd83f2c8c7cbeedc819c03631fdd670090c06` | 2026-10-07T09:22:18Z | conflict C1 | 3, 4,507 | 7, 23,664 | 83, 5,527,761 (99.5%) | 47 |
| `origin/codex/lint-helpers-from-lint-wave1-08` | `bdae8bc99dfe54a06615b20e34084bc9b8ebb06e` | 2026-10-07T12:24:58Z | clean | 4, 3,076 | 5, 13,719 | 61, 233,496 (93.3%) | 0 |
| `origin/codex/lint-helpers-from-lint-wave1-12` | `a4c72791598fb9494e02649b7eb380eab2a04779` | 2026-10-07T13:53:33Z | clean | 9, 8,950 | 21, 63,586 | 225, 7,267,509 (99.0%) | 1489 |
| `origin/codex/lint-helpers-from-lint-wave1-13` | `8d48f3f813a806a08d56dda1a71071af0c8a0674` | 2026-10-07T09:25:17Z | clean | 3, 4,012 | 12, 32,766 | 49, 169,836 (82.2%) | 0 |

Helper volumes are net changed paths against common helper base `95100eb4`; the base branch itself is compared against its comment-claim parent `29990b47`. Bytes are uncompressed Git blob sizes at the tip (a .gz blob is counted compressed), not checkout disk usage. Program fixtures, drivers and capture scripts are test code, while logged observations, JSON/gz corpora, docs and historical .a witnesses are evidence. Support types/tables count as implementation, so these numbers are conservative code upper bounds. Other-area diff files use each tip’s merge-base with area and include inherited upstream changes; they are collateral to avoid in a package extraction, not all authored helper changes. Full line counts are in each branch section.

**C1: all three conflicting branches** are `origin/codex/lint-helpers-from-codex/lint-wave1-02`, `origin/codex/lint-helpers-from-codex/lint-wave1-06`, and `origin/codex/lint-helpers-from-lint-wave1-05`. Every one conflicts in these exact eight files:

- `stage1/cohere/lint/helpers/README.md`
- `stage1/cohere/lint/helpers/comments/testdata/oracle.go`
- `stage1/cohere/lint/helpers/testdata/cases.json.gz`
- `stage1/cohere/lint/helpers/testdata/catalog.json`
- `stage1/cohere/lint/helpers/testdata/coverage.json`
- `stage1/cohere/lint/helpers/testdata/descriptors.json`
- `stage1/cohere/lint/inventory/testdata/engine.go`
- `stage1/cohere/lint/inventory/testdata/engine_test.go`

## Ordered package landing list

This order greedily maximizes the next increase using **retained candidate bodies**, breaking ties by consumer reach. Each row is one helper package. Take the union of the named source slices; no single branch contains a complete JSX, React, imports, structure, or CFG package. “Alone” and “with earlier” refer to conditional helper-ready rules among the 136. The last two columns are a separate **completion forecast** if every missing helper in that package is built fresh as well; these larger numbers are not available from these branches. No credit is given to compiler-gap witnesses, withdrawn reservations, or historical duplicates.

| Order/package | Source to take (retained symbols by owner) | Candidate/needed symbols | Alone | With earlier | Complete-package alone / with earlier | Conflicts and missing work |
|---|---|---|---:|---:|---|---|
| 1. `ecmascript/imports` | `origin/codex/lint-helpers-05` (3); `origin/codex/lint-helpers-from-codex/lint-wave1-15` (2); `origin/codex/lint-helpers-02` (1); `origin/codex/lint-helpers-03` (1) | 7/8 | 9 | 9 | 9 / 9 | Whole source tips clean. Build fresh 1 missing dependencies; see package gaps below. Projected AST/callback composition requires a package gate. |
| 2. `ecmascript/jsx` | `origin/codex/lint-helpers-02` (3); `origin/codex/lint-helpers-04` (1); `origin/codex/lint-helpers-03` (1); `origin/codex/lint-helpers-05` (1); `origin/codex/lint-helpers-01` (1) | 7/7 | 8 | 17 | 8 / 18 | Whole source tips clean. Projected AST/callback composition requires a package gate. |
| 3. `rules/structure` | `origin/codex/lint-helpers-05` (6); `origin/codex/lint-helpers-04` (3); `origin/codex/lint-helpers-01` (1); `origin/codex/lint-helpers-03` (1); `origin/codex/lint-helpers-from-codex/lint-wave1-11` (1) | 12/23 | 4 | 24 | 4 / 25 | Whole source tips clean. Build fresh 11 missing dependencies; see package gaps below. Projected AST/callback composition requires a package gate. |
| 4. `ecmascript/react` | `origin/codex/lint-helpers-05` (4); `origin/codex/lint-helpers-02` (3); `origin/codex/lint-helpers-01` (2); `origin/codex/lint-helpers-03` (1) | 10/29 | 2 | 34 | 2 / 38 | Whole source tips clean. Build fresh 19 missing dependencies; see package gaps below. Projected AST/callback composition requires a package gate. |
| 5. `ecmascript/property` | `origin/codex/lint-helpers-from-codex/lint-wave1-15` (1); `origin/codex/lint-helpers-02` (1) | 2/3 | 5 | 40 | 5 / 46 | Whole source tips clean. Build fresh 1 missing dependencies; see package gaps below.  |
| 6. `rules/react` | `origin/codex/lint-helpers-05` (2); `origin/codex/lint-helpers-04` (1); `origin/codex/lint-helpers-from-codex/lint-wave1-11` (1) | 4/35 | 0 | 45 | 6 / 65 | Whole source tips clean. Build fresh 31 missing dependencies; see package gaps below. Projected AST/callback composition requires a package gate. |
| 7. `ecmascript/text` | `origin/codex/lint-helpers-03` (2); `origin/codex/lint-helpers-01` (1) | 3/18 | 0 | 49 | 0 / 69 | Whole source tips clean. Build fresh 15 missing dependencies; see package gaps below. Projected AST/callback composition requires a package gate. |
| 8. `ecmascript/module` | `origin/codex/lint-helpers-04` (3) | 3/8 | 2 | 53 | 3 / 80 | Whole source tips clean. Build fresh 5 missing dependencies; see package gaps below.  |
| 9. `ecmascript/nextjs` | `origin/codex/lint-helpers-03` (3) | 3/9 | 1 | 55 | 2 / 88 | Whole source tips clean. Build fresh 6 missing dependencies; see package gaps below.  |
| 10. `ecmascript/regexsyntax` | `origin/codex/lint-helpers-05` (7) | 7/13 | 1 | 56 | 1 / 89 | Whole source tips clean. Build fresh 6 missing dependencies; see package gaps below.  |
| 11. `rules/tailwind` | `origin/codex/lint-helpers-04` (7); `origin/codex/lint-helpers-01` (6); `origin/codex/lint-helpers-03` (5); `origin/codex/lint-helpers-02` (4); `origin/codex/lint-helpers-05` (3); `origin/codex/lint-helpers-from-codex/lint-wave1-04` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-11` (1); `origin/codex/lint-helpers-from-lint-wave1-05` (1) | 29/73 | 0 | 56 | 6 / 95 | Whole-tip: lint-helpers-from-lint-wave1-05 C1. Build fresh 44 missing dependencies; see package gaps below.  |
| 12. `rules/core` | **none: build fresh** | 0/18 | 0 | 56 | 5 / 101 | No branch conflict to resolve. Build fresh 18 missing dependencies; see package gaps below.  |
| 13. `strict option decoding and schema validation` | **none: build fresh** | 0/1 | 0 | 56 | 2 / 107 | No branch conflict to resolve. Build fresh 1 missing dependencies; see package gaps below. Existing strict-options primitive is already on main; missing per-rule decoding/schema coverage is not supplied by re-landing it. |
| 14. `rules/tailwind/collapse` | `origin/codex/lint-helpers-04` (47); `origin/codex/lint-helpers-03` (33); `origin/codex/lint-helpers-05` (29); `origin/codex/lint-helpers-02` (13); `origin/codex/lint-helpers-01` (10); `origin/codex/lint-helpers-from-codex/lint-wave1-04` (5); `origin/codex/lint-helpers-from-lint-wave1-08` (3); `origin/codex/lint-helpers-from-lint-wave1-13` (3); `origin/codex/lint-helpers-from-codex/lint-wave1-09` (3); `origin/codex/lint-helpers-from-codex/lint-wave1-06` (2); `origin/codex/lint-helpers-from-codex-lint-wave1-03` (2); `origin/codex/lint-helpers-from-lint-wave1-05` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-11` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-10` (2); `origin/codex/lint-helpers-from-lint-wave1-12` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-15` (2) | 160/243 | 0 | 56 | 0 / 113 | Whole-tip: lint-helpers-from-codex/lint-wave1-06 C1, lint-helpers-from-lint-wave1-05 C1. Build fresh 83 missing dependencies; see package gaps below. Pick slot05 NewTheme; omit wave08 duplicate. Node-arena clone/remove, exact build counter, UTF-8 file input, and live Tailwind inputs need resolution. |
| 15. `ecmascript/regexp` | `origin/codex/lint-helpers-05` (24); `origin/codex/lint-helpers-01` (9); `origin/codex/lint-helpers-02` (3); `origin/codex/lint-helpers-04` (3); `origin/codex/lint-helpers-from-codex-lint-wave1-14` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-10` (2); `origin/codex/lint-helpers-from-codex/lint-wave1-15` (1) | 44/45 | 0 | 56 | 1 / 117 | Whole source tips clean. Build fresh 1 missing dependencies; see package gaps below. Compile is blocked on nonconstant RegExp; use fresh compiler support, not a handwritten matching substitute. |
| 16. `ecmascript/control_flow_graph` | `origin/codex/lint-helpers-05` (16); `origin/codex/lint-helpers-03` (16); `origin/codex/lint-helpers-02` (11); `origin/codex/lint-helpers-from-lint-wave1-01` (8); `origin/codex/lint-helpers-from-lint-wave1-12` (7); `origin/codex/lint-helpers-from-codex/lint-wave1-02` (4); `origin/codex/lint-helpers-from-codex/lint-wave1-11` (2); `origin/codex/lint-helpers-04` (1) | 65/79 | 0 | 56 | 2 / 120 | Whole-tip: lint-helpers-from-codex/lint-wave1-02 C1. Build fresh 14 missing dependencies; see package gaps below. appendSuccessor recursive graph remains blocked; callback bodies and graph identity/ownership need integrated proof. |
| 17. `rules/typescript` | **none: build fresh** | 0/2 | 0 | 56 | 3 / 123 | No branch conflict to resolve. Build fresh 2 missing dependencies; see package gaps below.  |
| 18. `ecmascript/scope` | **none: build fresh** | 0/4 | 0 | 56 | 0 / 126 | No branch conflict to resolve. Build fresh 4 missing dependencies; see package gaps below.  |
| 19. `ecmascript/regexpattern` | **none: build fresh** | 0/18 | 0 | 56 | 0 / 127 | No branch conflict to resolve. Build fresh 18 missing dependencies; see package gaps below.  |
| 20. `ecmascript/classmembers` | **none: build fresh** | 0/5 | 0 | 56 | 0 / 129 | No branch conflict to resolve. Build fresh 5 missing dependencies; see package gaps below.  |
| 21. `rules/next` | **none: build fresh** | 0/1 | 0 | 56 | 0 / 130 | No branch conflict to resolve. Build fresh 1 missing dependencies; see package gaps below.  |
| 22. `ecmascript/literal` | **none: build fresh** | 0/4 | 0 | 56 | 0 / 131 | No branch conflict to resolve. Build fresh 4 missing dependencies; see package gaps below.  |
| 23. `ecmascript/dotnotation` | **none: build fresh** | 0/12 | 0 | 56 | 1 / 132 | No branch conflict to resolve. Build fresh 12 missing dependencies; see package gaps below.  |
| 24. `ecmascript/directives` | **none: build fresh** | 0/10 | 0 | 56 | 1 / 133 | No branch conflict to resolve. Build fresh 10 missing dependencies; see package gaps below.  |
| 25. `ecmascript/decorators` | **none: build fresh** | 0/3 | 0 | 56 | 0 / 134 | No branch conflict to resolve. Build fresh 3 missing dependencies; see package gaps below.  |
| 26. `ecmascript/consistentreturn` | **none: build fresh** | 0/13 | 0 | 56 | 0 / 135 | No branch conflict to resolve. Build fresh 13 missing dependencies; see package gaps below.  |
| 27. `ecmascript/comments` | **none: build fresh** | 0/2 | 0 | 56 | 0 / 136 | No branch conflict to resolve. Build fresh 2 missing dependencies; see package gaps below.  |

The strict-options row represents seven rule-specific contracts sharing one placeholder, not one missing generic function. Existing generic option/schema primitives already passed their older isolated gate; each missing rule still needs its own defaults/configuration/schema adapter.

For the zero-increment tail, build the fresh high-reach core and strict-options work before spending integration effort on large blocked Tailwind/regexp/CFG families. A future scheduler can use the completion column to reorder fresh work once prerequisites are available. The present source-salvage order keeps the large blocked families out of the critical path.

### Missing dependencies by package

**`ecmascript/imports`** (1): `ecmascript/imports.LocalNameOfDefaultImport`.

**`rules/structure`** (11): `rules/structure.*NetworkFileAnalysis.collect`, `rules/structure.*NetworkFileAnalysis.collectFunctionDeclaration`, `rules/structure.*NetworkFileAnalysis.collectVariableStatement`, `rules/structure.NetworkFileAnalysisFor`, `rules/structure.bodyCallsNetworkService`, `rules/structure.callExpressionCallee`, `rules/structure.findNetworkServiceCalls`, `rules/structure.isAnonymousWrapperCall`, `rules/structure.isCacheInvalidateCall`, `rules/structure.isForwardRefCall`, `rules/structure.parameterName`.

**`ecmascript/react`** (19): `ecmascript/react.FunctionBody`, `ecmascript/react.FunctionParameters`, `ecmascript/react.IsCompilerComponentName`, `ecmascript/react.IsCompilerHookName`, `ecmascript/react.IsComponentOrHookLike`, `ecmascript/react.IsCreateElementCall`, `ecmascript/react.IsReachableRootPosition`, `ecmascript/react.SkipParenthesesUpward`, `ecmascript/react.callsHooksOrCreatesJsx`, `ecmascript/react.containsRef`, `ecmascript/react.containsSubstring`, `ecmascript/react.hasComponentOrHookName`, `ecmascript/react.hasComponentShapedParameters`, `ecmascript/react.hasPrimitiveTypeAnnotation`, `ecmascript/react.inferredFunctionName`, `ecmascript/react.isCompilerHookCallee`, `ecmascript/react.isJsxNode`, `ecmascript/react.isSkippedNestedFunction`, `ecmascript/react.isSpreadParameter`.

**`ecmascript/property`** (1): `ecmascript/property.NameTagged`.

**`rules/react`** (31): `rules/react.DecodeNoMethodSetStateOptions`, `rules/react.attributesOf`, `rules/react.commentValueOf`, `rules/react.enclosingClassOf`, `rules/react.enclosingComponentOf`, `rules/react.enclosingFunctionOf`, `rules/react.functionBodyBlock`, `rules/react.hasValidComponentParameters`, `rules/react.isComponentClass`, `rules/react.isComponentIdentifierName`, `rules/react.isCreateReactClassCall`, `rules/react.isFunctionLike`, `rules/react.isHookIdentifierName`, `rules/react.isJavaScriptIdentifier`, `rules/react.isNonNodeExpression`, `rules/react.isPureComponentBase`, `rules/react.isReactCompiledFunction`, `rules/react.isReactComponentBase`, `rules/react.isRestParameter`, `rules/react.isThisExpression`, `rules/react.isTopLevelCompilationCandidate`, `rules/react.jsxAnnotationIn`, `rules/react.mentionsRef`, `rules/react.parametersOf`, `rules/react.reactFunctionNameOf`, `rules/react.reactPragmaFor`, `rules/react.returnsNonNode`, `rules/react.semanticParentOf`, `rules/react.sortDefaultPropsInitializerOf`, `rules/react.sourceSliceOf`, `rules/react.stylePropObjectUnwrapParentheses`.

**`ecmascript/text`** (15): `ecmascript/text.BestMatch`, `ecmascript/text.GraphemeCount`, `ecmascript/text.MinimumEditDistance`, `ecmascript/text.continuesCluster`, `ecmascript/text.hangulJoins`, `ecmascript/text.hangulLeading`, `ecmascript/text.hangulSyllable`, `ecmascript/text.hangulTrailing`, `ecmascript/text.hangulVowel`, `ecmascript/text.hangulVowelSyllable`, `ecmascript/text.isGraphemeControl`, `ecmascript/text.isGraphemeExtend`, `ecmascript/text.isIndicLinker`, `ecmascript/text.isPictograph`, `ecmascript/text.isRegionalIndicator`.

**`ecmascript/module`** (5): `ecmascript/module.AllDeclaredTypeNames`, `ecmascript/module.DeclaresTypeNamed`, `ecmascript/module.HasDefaultModifier`, `ecmascript/module.IsDefaultExported`, `ecmascript/module.forEachDeclaredTypeName`.

**`ecmascript/nextjs`** (6): `ecmascript/nextjs.IsDocumentPage`, `ecmascript/nextjs.IsInApplicationDirectory`, `ecmascript/nextjs.IsInPagesDirectory`, `ecmascript/nextjs.RouteContractExports`, `ecmascript/nextjs.buildRouteFileContracts`, `ecmascript/nextjs.splitSegments`.

**`ecmascript/regexsyntax`** (6): `ecmascript/regexsyntax.ParseHexUint`, `ecmascript/regexsyntax.ParseRegexCharacterClassWithEnd`, `ecmascript/regexsyntax.hexValue`, `ecmascript/regexsyntax.parseRegexCharacterClass`, `ecmascript/regexsyntax.readClassEscape`, `ecmascript/regexsyntax.readRawClassChar`.

**`rules/tailwind`** (44): `rules/tailwind.*ClassLiteralReader.ClassSegmentsIn`, `rules/tailwind.*ClassLiteralReader.ClassTemplateSegmentsIn`, `rules/tailwind.*ClassLiteralReader.ClassTemplatesIn`, `rules/tailwind.*strictVariantLevel.entry`, `rules/tailwind.*strictVariantLevel.order`, `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.anyCandidateResolves`, `rules/tailwind.atoi`, `rules/tailwind.boundaryAfterHole`, `rules/tailwind.boundaryBeforeHole`, `rules/tailwind.buildClass`, `rules/tailwind.candidateValueText`, `rules/tailwind.classExistsIn`, `rules/tailwind.classOrderKeys`, `rules/tailwind.classesOf`, `rules/tailwind.compareVariantMasks`, `rules/tailwind.compileIgnorePatterns`, `rules/tailwind.composesForRoot`, `rules/tailwind.containsString`, `rules/tailwind.deprecationFor`, `rules/tailwind.endsWithWhitespace`, `rules/tailwind.isArbitraryVariant`, `rules/tailwind.isArrayIndexKey`, `rules/tailwind.isIgnored`, `rules/tailwind.jsKeyOrder`, `rules/tailwind.newStrictVariantLevel`, `rules/tailwind.parseTailwindVersion`, `rules/tailwind.propertyIndexBeyond`, `rules/tailwind.readingFor`, `rules/tailwind.repositoryClassFacts`, `rules/tailwind.resolveClassFactsIn`, `rules/tailwind.segmentsOfTemplate`, `rules/tailwind.sortClassesByKey`, `rules/tailwind.sourceValueAfterRoot`, `rules/tailwind.splitCandidateIn`, `rules/tailwind.splitVariantPrefix`, `rules/tailwind.startsWithWhitespace`, `rules/tailwind.strictClassOrder`, `rules/tailwind.strictVariantsOf`, `rules/tailwind.tailwindAtLeast`, `rules/tailwind.templateExpressionOf`, `rules/tailwind.templatesFrom`, `rules/tailwind.valueIsColorIn`, `rules/tailwind.valueResolutionIn`.

**`rules/core`** (18): `rules/core.appendNodeSignature`, `rules/core.appendTokensBetween`, `rules/core.bodyDefinitelyExits`, `rules/core.codePathRoots`, `rules/core.consistentReturnIsGenerator`, `rules/core.hasSameTokens`, `rules/core.idDenylistIsDestructuringTarget`, `rules/core.idDenylistIsImportAttributeKey`, `rules/core.idDenylistIsImportOptionsObject`, `rules/core.isEmptyBracketLiteral`, `rules/core.isNullOrUndefined`, `rules/core.isSeparateEvaluationContext`, `rules/core.memberAccessObject`, `rules/core.noRestrictedExportsNameText`, `rules/core.numericLiteralSign`, `rules/core.switchStatementExits`, `rules/core.tokenSignature`, `rules/core.tryStatementExits`.

**`strict option decoding and schema validation`** (1): `strict option decoding and schema validation`.

**`rules/tailwind/collapse`** (83): `rules/tailwind/collapse.*Descriptor.axisFor`, `rules/tailwind/collapse.*LoadedDesignSystem.DeclaresFunctionalUtility`, `rules/tailwind/collapse.*LoadedDesignSystem.Variants`, `rules/tailwind/collapse.*RunVariantOrder.IndexOf`, `rules/tailwind/collapse.*RunVariantOrder.VariantBitmask`, `rules/tailwind/collapse.*Table.ArbitraryPropertyReading`, `rules/tailwind/collapse.*Table.Lookup`, `rules/tailwind/collapse.*Table.bareReading`, `rules/tailwind/collapse.*Table.frameworkReading`, `rules/tailwind/collapse.*Theme.Get`, `rules/tailwind/collapse.*UtilityEvaluator.Has`, `rules/tailwind/collapse.*UtilityEvaluator.Reading`, `rules/tailwind/collapse.*VariantRegistry.BuildVariantOrder`, `rules/tailwind/collapse.*VariantRegistry.Compare`, `rules/tailwind/collapse.AxisReadings.readingForType`, `rules/tailwind/collapse.ClassValueResolvesIn`, `rules/tailwind/collapse.CompareVariantBitmasks`, `rules/tailwind/collapse.ComposesFor`, `rules/tailwind/collapse.DeclaredPropertiesFor`, `rules/tailwind/collapse.EmitGapRoot`, `rules/tailwind/collapse.FrameworkFunctionalUtility.Emit`, `rules/tailwind/collapse.FrameworkFunctionalUtility.Reading`, `rules/tailwind/collapse.FrameworkFunctionalUtility.staticValueEmitter`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.Emit`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.ReadingFor`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.readingOf`, `rules/tailwind/collapse.IsColorKeyword`, `rules/tailwind/collapse.ModifierAxisFor`, `rules/tailwind/collapse.PrintVariant`, `rules/tailwind/collapse.ResolveFunctionalUtilityValue`, `rules/tailwind/collapse.SelectorShapeForRoot`, `rules/tailwind/collapse.SelectorShapeOfNodes`, `rules/tailwind/collapse.UtilityBranch.isOneOf`, `rules/tailwind/collapse.UtilityBranch.maskStopIsColor`, `rules/tailwind/collapse.UtilityBranch.shadowIsColor`, `rules/tailwind/collapse.UtilityBranchFor`, `rules/tailwind/collapse.absentValued`, `rules/tailwind/collapse.appendVariantAndNested`, `rules/tailwind/collapse.appendVisibleProperty`, `rules/tailwind/collapse.bareValueHandler`, `rules/tailwind/collapse.borderSideEmitter`, `rules/tailwind/collapse.bracketed`, `rules/tailwind/collapse.cloneNode`, `rules/tailwind/collapse.cloneNodes`, `rules/tailwind/collapse.collectVisibleProperties`, `rules/tailwind/collapse.containsString`, `rules/tailwind/collapse.declarations`, `rules/tailwind/collapse.declareComposing`, `rules/tailwind/collapse.declareComposingMask`, `rules/tailwind/collapse.declareComposingWebkit`, `rules/tailwind/collapse.declareProperties`, `rules/tailwind/collapse.declareProperty`, `rules/tailwind/collapse.declareSorted`, `rules/tailwind/collapse.declareWrapped`, `rules/tailwind/collapse.descriptionForRoot`, `rules/tailwind/collapse.emitFunctionalRoot`, `rules/tailwind/collapse.emitRootWithValue`, `rules/tailwind/collapse.escapeUnderscoreAndSpace`, `rules/tailwind/collapse.escapeUnderscoresAndSpaces`, `rules/tailwind/collapse.gapTypeListFor`, `rules/tailwind/collapse.gradientPositionEmitter`, `rules/tailwind/collapse.isLoneVar`, `rules/tailwind/collapse.isSpaceSeparator`, `rules/tailwind/collapse.maskEdgeEmitter`, `rules/tailwind/collapse.maskGradientEmitter`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.printArbitraryValue`, `rules/tailwind/collapse.printModifier`, `rules/tailwind/collapse.printNodesCss`, `rules/tailwind/collapse.removeNodes`, `rules/tailwind/collapse.resolveArm`, `rules/tailwind/collapse.resolveArmByInferredType`, `rules/tailwind/collapse.resolveNamedValue`, `rules/tailwind/collapse.ringEmitter`, `rules/tailwind/collapse.rotateEmitter`, `rules/tailwind/collapse.scaleEmitter`, `rules/tailwind/collapse.shadowFamilyEmitter`, `rules/tailwind/collapse.toPrintNodes`, `rules/tailwind/collapse.unwrapIsSelector`, `rules/tailwind/collapse.variantsAreEqual`, `rules/tailwind/collapse.visibleDeclarationText`, `rules/tailwind/collapse.withResolvedValue`, `rules/tailwind/collapse.withoutRemoved`.

**`ecmascript/regexp`** (1): `ecmascript/regexp.Compile`.

**`ecmascript/control_flow_graph`** (14): `ecmascript/control_flow_graph.*Block[E].Index`, `ecmascript/control_flow_graph.*Builder[E].Current`, `ecmascript/control_flow_graph.*Builder[E].Emit`, `ecmascript/control_flow_graph.*Builder[E].appendSuccessor`, `ecmascript/control_flow_graph.*PathAnalysis[E].IsCyclic`, `ecmascript/control_flow_graph.*PathAnalysis[E].IsOnEveryFinalPath`, `ecmascript/control_flow_graph.*PathAnalysis[E].countFinalPaths`, `ecmascript/control_flow_graph.*PathAnalysis[E].findFinalPathDominators`, `ecmascript/control_flow_graph.*PathAnalysis[E].findShortestPaths`, `ecmascript/control_flow_graph.*cyclePaths[E].count`, `ecmascript/control_flow_graph.*cyclePaths[E].countToEnd`, `ecmascript/control_flow_graph.AnalyzePaths`, `ecmascript/control_flow_graph.IndexRoots`, `ecmascript/control_flow_graph.newCyclePaths`.

**`rules/typescript`** (2): `rules/typescript.isTypeScriptSourceFile`, `rules/typescript.nonNullAssertionOperatorRange`.

**`ecmascript/scope`** (4): `ecmascript/scope.BodyOf`, `ecmascript/scope.EnclosingFunctionLike`, `ecmascript/scope.NameOf`, `ecmascript/scope.borrowedName`.

**`ecmascript/regexpattern`** (18): `ecmascript/regexpattern.*walker.applyQuantifier`, `ecmascript/regexpattern.*walker.emit`, `ecmascript/regexpattern.*walker.emitAndQuantify`, `ecmascript/regexpattern.*walker.kindFromSource`, `ecmascript/regexpattern.*walker.readEscape`, `ecmascript/regexpattern.*walker.run`, `ecmascript/regexpattern.*walker.walkClass`, `ecmascript/regexpattern.Walk`, `ecmascript/regexpattern.braceQuantifierEnd`, `ecmascript/regexpattern.consumeDigits`, `ecmascript/regexpattern.decodeRune`, `ecmascript/regexpattern.escapeValue`, `ecmascript/regexpattern.extendOctalEscape`, `ecmascript/regexpattern.groupPrologueEnd`, `ecmascript/regexpattern.isAsciiLetter`, `ecmascript/regexpattern.octalEscapeValue`, `ecmascript/regexpattern.quantifierAt`, `ecmascript/regexpattern.unicodeEscapeValue`.

**`ecmascript/classmembers`** (5): `ecmascript/classmembers.ForEachDuplicate`, `ecmascript/classmembers.IsAccessorKind`, `ecmascript/classmembers.IsOverloadSignature`, `ecmascript/classmembers.KeyOf`, `ecmascript/classmembers.MemberName`.

**`rules/next`** (1): `rules/next.urlQueryValue`.

**`ecmascript/literal`** (4): `ecmascript/literal.CookedToRaw`, `ecmascript/literal.cookedBytesProducedBy`, `ecmascript/literal.escapeWidthInStringLiteral`, `ecmascript/literal.producesNoCookedBytes`.

**`ecmascript/dotnotation`** (12): `ecmascript/dotnotation.CompileAllowPattern`, `ecmascript/dotnotation.Listeners`, `ecmascript/dotnotation.checkComputed`, `ecmascript/dotnotation.computedFix`, `ecmascript/dotnotation.continuesAnIdentifier`, `ecmascript/dotnotation.hasCommentBetween`, `ecmascript/dotnotation.isOptionalAccess`, `ecmascript/dotnotation.jsonQuote`, `ecmascript/dotnotation.keywordFix`, `ecmascript/dotnotation.literalKey`, `ecmascript/dotnotation.useBracketsMessage`, `ecmascript/dotnotation.useDotMessage`.

**`ecmascript/directives`** (10): `ecmascript/directives.ParseDisable`, `ecmascript/directives.ParseEnable`, `ecmascript/directives.Recognize`, `ecmascript/directives.endsWord`, `ecmascript/directives.isLineComment`, `ecmascript/directives.parseRuleNames`, `ecmascript/directives.splitDirective`, `ecmascript/directives.splitReason`, `ecmascript/directives.splitScope`, `ecmascript/directives.stripCommentMarkers`.

**`ecmascript/decorators`** (3): `ecmascript/decorators.CallName`, `ecmascript/decorators.HasDecoratorInSet`, `ecmascript/decorators.Of`.

**`ecmascript/consistentreturn`** (13): `ecmascript/consistentreturn.HasValue`, `ecmascript/consistentreturn.IsGenerator`, `ecmascript/consistentreturn.IsScope`, `ecmascript/consistentreturn.Judge`, `ecmascript/consistentreturn.Name`, `ecmascript/consistentreturn.ReportRange`, `ecmascript/consistentreturn.Verb`, `ecmascript/consistentreturn.canRunOffEnd`, `ecmascript/consistentreturn.capitaliseFirst`, `ecmascript/consistentreturn.isExemptFromEndJudgment`, `ecmascript/consistentreturn.judgeReturns`, `ecmascript/consistentreturn.judgeScope`, `ecmascript/consistentreturn.staticName`.

**`ecmascript/comments`** (2): `ecmascript/comments.LeadingRunFor`, `ecmascript/comments.isAdjacentGap`.

## Overlap decisions

Only **one exact cohere symbol has two retained executable implementations** after honoring withdrawals: `rules/tailwind/collapse.NewTheme`. Slot05 `stage1/cohere/lint/helpers/slot05/batch6/new_theme.a:4` and wave08 `stage1/cohere/lint/helpers/from_wave08/theme_new.a:10` both return an empty prefix, freshly allocated writable map, fresh empty order array and zero deadKeys. Their return expressions are structurally equal after substituting the ThemeValue/ThemeEntry type names; their surrounding type/import declarations are not byte-identical. Thus they agree on this constructor contract, including independent instances, but no joint cross-branch graph integration was run.

**Take slot05.** Its claimed reservation predates wave08, and it shares MutableThemeState/ThemeValue with the liveKeys/clearAll/delete/resolveKey family. `stage1/cohere/lint/helpers/slot05/batch6/helper_test.go:23` `TestSlot05NewTheme` injects dead-count, prefix, shared-map and shared-order mutants. These mutants are checked natively in its verify loop; source/emitted baseline comparisons also run. Wave08 `stage1/cohere/lint/helpers/from_wave08/helpers_test.go:99` `TestNewTheme` has three-backend comparison and mutant checking at lines 125–185. Retain that extra mutant idea if composing a new package test, but not its duplicate production constructor.

The following races are **withdrawn**, not extra ready implementations: isIdentifierNamed (01/02/03 yield to 05); FileContextFor (03 yields to 01); isComponentBaseName (01 yields to 03); IsEs6ComponentClass (03 yields to 01); AttributeName (05 yields to 02); ListenerKinds (04 yields to 03); isSpace (02 yields to 03); readClassValues (05 yields to 03); attributeValues (03/04 yield to 01); MatchExactly (03 yields to 05); StringAttributeValue (02 yields to 01); IsEs5ComponentCall (04 yields to 02); isValidThemePrefix/namespaceForVariantRoot (03 yields to 02); breakpointBucket (01 yields to 04); HasVariant/VariantKind (02/15 yield to 04); NewVariantRegistry (08 yields to 03); normalizeValueFunctionArgument (01 yields to wave12); NewUtilityEvaluator/addRepositoryFunctionalRoots (02 yields to 05); ingestUtilityBlock (01 yields to 04); FrameworkStaticReading (wave06 yields to from-codex-lint-wave1-03); nodesFromStaticDeclarations (wave14 yields to wave09); leadingInteger (wave14 yields to wave05); FindEntryPoint (wave05 yields to wave11); decodeControlEscape (02 yields to 05); boundedQuantifierWidth/groupKindOf (05 yields to 01); enter (from-codex/lint-wave1-02 yields to from-lint-wave1-01); makeReturn/makeThrow (02 yields to 03); expr/patternBind (04 yields to 03). ParseVariant, DesignSystemForProgram and stylesheetCollector.loadFile have additional withdrawn reservations; a reservation alone is not code.

For withdrawn locally passing variants, the claims report agreement with their own Go observations, sometimes with source-only or bounded input coverage. That is not a pairwise equality proof across their whole domains. No source exists at many current tips; retained .a.txt files and mutation artifacts are historical evidence only. Prefer the retained owner above rather than resurrecting the duplicate. Per-branch claim citations and archival witness paths are listed below. All branches also inherit the exact common comment/strict-options base; that ancestry is not 20 independent implementations.
## Branch-by-branch symbols and evidence

Every symbol below is cohere-qualified relative to `github.com/system-inc/cohere/internal/lint/`. For example `ecmascript/jsx.ElementParts` denotes that package and Go function. Source locations are at the pinned branch tip, accessed with `git show <tip>:<path>`; they do not imply that today’s detached local checkout contains the file. A function’s Adamic spelling can differ from Go’s. “Blocked” rows are source/probe evidence, not delivered implementations. Proof indices cite the real tests/scripts, including mutation checks in the same test body when the test name does not contain Mutant. Recorded reports are historic observations; none was rerun here.

### origin/codex/lint-helpers

Tip `95100eb440b47f3f18e960c6f5b49cadf0dc1d9d`; merge-base with area `95100eb440b47f3f18e960c6f5b49cadf0dc1d9d`; merge-tree clean (exit 0).

**Coverage verdict:** Already on area: five comment helpers, common strict-options/options-JSON/schema/policy primitives inherited. Comment helpers exercise original consumer inputs and semantic mutants; no new delivery relative to area. No pending rules are unlocked by merging this ancestor again.


Already-landed comment symbols:
- `ecmascript/comments.All` → `stage1/cohere/lint/helpers/comments/all.ts:11`
- `ecmascript/comments.ForFile` → `stage1/cohere/lint/helpers/comments/for_file.ts:23`
- `ecmascript/comments.canBeginAt` → `stage1/cohere/lint/helpers/comments/can_begin_at.ts:2`
- `ecmascript/comments.collectListInteriors` → `stage1/cohere/lint/helpers/comments/collect_list_interiors.ts:52`
- `ecmascript/comments.sortByPosition` → `stage1/cohere/lint/helpers/comments/sort_by_position.ts:2`
Inherited primitives, not additional cohere function ports: `stage1/cohere/lint/helpers/strict_options.ts:4` StrictOptions; `stage1/cohere/lint/helpers/options_json.ts:10` OptionsJson; `stage1/cohere/lint/helpers/option_schema.ts:4` OptionSchema; `stage1/cohere/lint/helpers/policy_message.ts:5` PolicyMessage. They model Go decoding/schema/messages rather than exposing an exact shared cohere helper symbol.

**Proof index:**
- `stage1/cohere/lint/helpers/comments/comments_test.go:89`: TestCommentsMatchCohere @89, TestCommentMutants @100, TestConsumerCommentHelpers @149, TestJsxParserGapIsExplicit @165, TestJsxAdapterGuardMutant @209.
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/helpers_test.go:78`: TestHelpersMatchCohere @78, TestHelperMutants @116, TestMessageRefusalsMatchGo @191, TestKnownGapsAreExplicit @239.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 24 files; +22,479/-0 lines; 1,358,481 blob bytes.
- implementation/support types: 6 files; +314/-0 lines; 11,782 blob bytes.
- test/oracle/capture programs: 5 files; +548/-0 lines; 20,752 blob bytes.

### origin/codex/lint-helpers-01

Tip `98f7e5f08c10f148d457e3a1fdde0355d1401f6c`; merge-base with area `db2ecc00447f9ebe8adecb190f71ac222e5db860`; merge-tree clean (exit 0).

**Coverage verdict:** 30 delivered helper bodies plus three BLOCKED node clone/remove source ports. Initial and later tests use actual Go overlays with selected consumer fixtures and explicit controls; dependencies such as isComponentBase, entity table/hexValue, parser facts and regex compilation remain external. No complete consuming-rule findings parity. Wave11 deliberately skips native/emitted coverage after cycle-capable arena refusal; do not credit its eight source-only mutants. See SLOT01_WAVE11_REPORT.md and slot01_wave11_readiness.json.

Claim: `stage1/cohere/lint/helpers/claims/01.md:1` (203 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/jsx.StringAttributeValue` | `stage1/cohere/lint/helpers/jsx_string_attribute_value.a:4` | retained candidate |
| `ecmascript/react.IsEs6ComponentClass` | `stage1/cohere/lint/helpers/react_es6_component_class.a:13` | retained candidate |
| `ecmascript/react.IsLikelyComponentName` | `stage1/cohere/lint/helpers/react_likely_component_name.a:4` | retained candidate |
| `ecmascript/regexp.boundedQuantifierWidth` | `stage1/cohere/lint/helpers/regexp_bounded_quantifier_width.a:3` | retained candidate |
| `ecmascript/regexp.classAtom.covers` | `stage1/cohere/lint/helpers/regexp_class_atom_covers.a:2` | retained candidate |
| `ecmascript/regexp.decimalEscape` | `stage1/cohere/lint/helpers/regexp_decimal_escape.a:2` | retained candidate |
| `ecmascript/regexp.decodeLegacyOctal` | `stage1/cohere/lint/helpers/regexp_legacy_octal.a:2` | retained candidate |
| `ecmascript/regexp.groupKindOf` | `stage1/cohere/lint/helpers/regexp_group_kind.a:2` | retained candidate |
| `ecmascript/regexp.nonWordClassAtoms` | `stage1/cohere/lint/helpers/regexp_nonword_class_atoms.a:3` | retained candidate |
| `ecmascript/regexp.quantifierWidth` | `stage1/cohere/lint/helpers/regexp_quantifier_width.a:2` | retained candidate |
| `ecmascript/regexp.wordBoundary` | `stage1/cohere/lint/helpers/regexp_word_boundary.a:2` | retained candidate |
| `ecmascript/regexp.wordClassAtoms` | `stage1/cohere/lint/helpers/regexp_word_class_atoms.a:3` | retained candidate |
| `ecmascript/text.decodeEntity` | `stage1/cohere/lint/helpers/text_decode_entity.a:5` | retained candidate |
| `rules/structure.FileContextFor` | `stage1/cohere/lint/helpers/structure_file_context.a:15` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.attributeValues` | `stage1/cohere/lint/helpers/tailwind_attribute_values.a:12` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.classValuesIn` | `stage1/cohere/lint/helpers/tailwind_class_values_in.a:9` | retained candidate |
| `rules/tailwind.DefaultClassLiteralSettings` | `stage1/cohere/lint/helpers/tailwind_default_class_literal_settings.a:9` | retained candidate |
| `rules/tailwind.NewClassLiteralReader` | `stage1/cohere/lint/helpers/tailwind_new_class_literal_reader.a:13` | retained candidate |
| `rules/tailwind.classLiteralFrom` | `stage1/cohere/lint/helpers/tailwind_class_literal_from.a:11` | retained candidate |
| `rules/tailwind.dissectClass` | `stage1/cohere/lint/helpers/tailwind_dissect_class.a:2` | retained candidate |
| `rules/tailwind/collapse.*Node.IsContainer` | `stage1/cohere/lint/helpers/collapse_node_is_container.a:3` | retained candidate |
| `rules/tailwind/collapse.*VariantRegistry.RegisterFrameworkVariants` | `stage1/cohere/lint/helpers/collapse_register_framework_variants.a:4` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.ingestThemeBlock` | `stage1/cohere/lint/helpers/collapse_ingest_theme_block.a:18` | retained candidate |
| `rules/tailwind/collapse.NewTable` | `stage1/cohere/lint/helpers/collapse_new_table.a:27` | retained candidate |
| `rules/tailwind/collapse.cloneNode` | `stage1/cohere/lint/helpers/collapse_clone_node.a:8` | BLOCKED source/probe only |
| `rules/tailwind/collapse.cloneNodes` | `stage1/cohere/lint/helpers/collapse_clone_nodes.a:4` | BLOCKED source/probe only |
| `rules/tailwind/collapse.isHexDigit` | `stage1/cohere/lint/helpers/collapse_is_hex_digit.a:3` | retained candidate |
| `rules/tailwind/collapse.isValidArbitrary` | `stage1/cohere/lint/helpers/collapse_valid_arbitrary.a:1` | retained candidate |
| `rules/tailwind/collapse.isValidNamedValue` | `stage1/cohere/lint/helpers/collapse_valid_named_value.a:1` | retained candidate |
| `rules/tailwind/collapse.normalizeUtilityDefinition` | `stage1/cohere/lint/helpers/collapse_normalize_utility_definition.a:2` | retained candidate |
| `rules/tailwind/collapse.normalizeValueFunctionArguments` | `stage1/cohere/lint/helpers/collapse_normalize_value_function_arguments.a:3` | retained candidate |
| `rules/tailwind/collapse.parseThemeOptions` | `stage1/cohere/lint/helpers/collapse_parse_theme_options.a:3` | retained candidate |
| `rules/tailwind/collapse.removeNodes` | `stage1/cohere/lint/helpers/collapse_remove_nodes.a:4` | BLOCKED source/probe only |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot01_test.go:136`: TestSlot01FileContextMatchesCohere @136, TestSlot01Es6ComponentClassMatchesCohere @141, TestSlot01TailwindDefaultsMatchCohere @146.
- `stage1/cohere/lint/helpers/slot01_wave10_test.go:154`: TestSlot01Wave10Word @154, TestSlot01Wave10NonWord @159, TestSlot01Wave10Boundary @164, TestSlot01Wave10WordFresh @169, TestSlot01Wave10NonWordFresh @174, TestSlot01Wave10BoundaryForward @179, TestSlot01Wave10BoundaryCount @184.
- `stage1/cohere/lint/helpers/slot01_wave11_test.go:172`: TestSlot01Wave11Node @172, TestSlot01Wave11Context @177, TestSlot01Wave11Children @182, TestSlot01Wave11List @187, TestSlot01Wave11ListNil @192, TestSlot01Wave11Remove @197, TestSlot01Wave11RemoveIdentity @202, TestSlot01Wave11RemoveEmpty @207.
- `stage1/cohere/lint/helpers/slot01_wave2_test.go:12`: TestSlot01Wave2FactoryMatchesCohere @12, TestSlot01Wave2MemoMatchesCohere @15, TestSlot01Wave2LiteralMatchesCohere @18, TestSlot01Wave2LiteralShortRangeMutant @21, TestSlot01Wave2FactoryInvalidPatternMutant @24, TestSlot01Wave2FactoryAttributeMutant @27, TestSlot01Wave2MemoUnboundMutant @30, TestSlot01Wave2MemoEmptyMutant @33.
- `stage1/cohere/lint/helpers/slot01_wave3_test.go:18`: TestSlot01Wave3AttributeMatchesCohere @18, TestSlot01Wave3EntityMatchesCohere @160, TestSlot01Wave3EntitySurrogateMutant @163, TestSlot01Wave3ComponentMatchesCohere @168, TestSlot01Wave3AttributeFalseValueMutant @190, TestSlot01Wave3EntityNamedTableMutant @193, TestSlot01Wave3ComponentStrideMutant @198.
- `stage1/cohere/lint/helpers/slot01_wave4_test.go:14`: TestSlot01Wave4StringValueMatchesCohere @14, TestSlot01Wave4HexMatchesCohere @145, TestSlot01Wave4ContainerMatchesCohere @150, TestSlot01Wave4StringDecoderMutant @155, TestSlot01Wave4StringNilFirstMutant @158, TestSlot01Wave4HexDomainRefusal @163.
- `stage1/cohere/lint/helpers/slot01_wave5_test.go:141`: TestSlot01Wave5NamedMatchesCohere @141, TestSlot01Wave5NamedEmptyMutant @144, TestSlot01Wave5ArbitraryMatchesCohere @149, TestSlot01Wave5ArbitraryFinalMutant @152, TestSlot01Wave5ArbitraryEscapeMutant @155, TestSlot01Wave5ThemeMatchesCohere @160, TestSlot01Wave5ThemeLastPrefixMutant @163.
- `stage1/cohere/lint/helpers/slot01_wave6_test.go:155`: TestSlot01Wave6DefinitionMatchesCohere @155, TestSlot01Wave6DefinitionForwardMutant @158, TestSlot01Wave6WalkerMatchesCohere @163, TestSlot01Wave6WalkerPresentMutant @166, TestSlot01Wave6WalkerKindMutant @169, TestSlot01Wave6WalkerBailMutant @172, TestSlot01Wave6WalkerOrderMutant @177, TestSlot01Wave6FrameworkMatchesCohere @186, TestSlot01Wave6FrameworkLastOrderMutant @189, TestSlot01Wave6FrameworkCopyMutant @192, TestSlot01Wave6FrameworkDomainRefusal @197.
- `stage1/cohere/lint/helpers/slot01_wave7_test.go:216`: TestSlot01Wave7ThemeMatchesCohere @216, TestSlot01Wave7ThemePrefixMutant @219, TestSlot01Wave7ThemePropertyMutant @222, TestSlot01Wave7ThemeUnescapeMutant @226, TestSlot01Wave7ThemeErrorStopMutant @229, TestSlot01Wave7ThemeInvalidPrefixMutant @232, TestSlot01Wave7TableMatchesCohere @237, TestSlot01Wave7TableReadingCopyMutant @240, TestSlot01Wave7TablePropertyAliasMutant @243, TestSlot01Wave7TableOrderMutant @246, TestSlot01Wave7TableDescriptorMapMutant @250, TestSlot01Wave7TableFrameworkMutant @253, TestSlot01Wave7TableHeaderCopyMutant @257, TestSlot01Wave7TableBackingShareMutant @260, TestSlot01Wave7DissectMatchesCohere @265, TestSlot01Wave7DissectSuffixMutant @268, TestSlot01Wave7DissectColonMutant @271.
- `stage1/cohere/lint/helpers/slot01_wave8_test.go:141`: TestSlot01Wave8Decimal @141, TestSlot01Wave8Octal @146, TestSlot01Wave8Coverage @151.
- `stage1/cohere/lint/helpers/slot01_wave9_test.go:141`: TestSlot01Wave9Bounded @141, TestSlot01Wave9Quantifier @146, TestSlot01Wave9Group @151.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_LANDING2_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE10_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE11_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE2_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE3_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE4_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE5_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE6_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE7_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE8_REPORT.md:1`, `stage1/cohere/lint/helpers/SLOT01_WAVE9_REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/evidence/slot01-wave7/withdrawn-initial-targeted.log`, `stage1/cohere/lint/helpers/evidence/slot01-wave7/withdrawn-regression.log`, `stage1/cohere/lint/helpers/evidence/slot01-wave7/withdrawn-utility-headers.log`, `stage1/cohere/lint/helpers/evidence/slot01-wave7/withdrawn-utility-initial.log`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 186 files; +47,969/-0 lines; 5,245,907 blob bytes.
- implementation/support types: 35 files; +980/-0 lines; 40,850 blob bytes.
- test/oracle/capture programs: 45 files; +5,142/-0 lines; 201,445 blob bytes.

### origin/codex/lint-helpers-02

Tip `ff5090b777bc4e27bd51a5b57d9c824670f14b26`; merge-base with area `e667e3e1dbdfd1b9125c3256961bfbc8ec31946b`; merge-tree clean (exit 0).

**Coverage verdict:** 39 retained helpers. Initial AttributeName and property.Name corpus derives consumers from readiness, checks every consumer has captured input, and queries real Go AST nodes; their original tests compare source Node/native and their compiled mutants, without an emitted-JS path. Later batch harnesses include emitted JavaScript. CFG statement/condition and Tailwind evaluators test callback orchestration, not independent implementations of callback bodies. ClassLiteralReaderFor has cache/binding controls rather than proof of every complete consuming rule. Archived full-helper reruns do not establish current-area integration.

Claim: `stage1/cohere/lint/helpers/claims/02.md:1` (250 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].condition` | `stage1/cohere/lint/helpers/slot02/batch13/condition.a:13` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].link` | `stage1/cohere/lint/helpers/slot02/batch9/link.a:1` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].linkWithCycleBarrier` | `stage1/cohere/lint/helpers/slot02/batch9/link_with_cycle_barrier.a:7` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].loop` | `stage1/cohere/lint/helpers/slot02/batch10/loop.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].nestedFunction` | `stage1/cohere/lint/helpers/slot02/batch11/nested_function.a:1` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].read` | `stage1/cohere/lint/helpers/slot02/batch10/read.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].setCycleBarrier` | `stage1/cohere/lint/helpers/slot02/batch9/set_cycle_barrier.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].statement` | `stage1/cohere/lint/helpers/slot02/batch13/statement.a:5` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].variableDeclaration` | `stage1/cohere/lint/helpers/slot02/batch12/variable_declaration.a:9` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].variableDeclarationList` | `stage1/cohere/lint/helpers/slot02/batch11/variable_declaration_list.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].write` | `stage1/cohere/lint/helpers/slot02/batch10/write.a:2` | retained candidate |
| `ecmascript/imports.CallExpressionSource` | `stage1/cohere/lint/helpers/slot02/batch7/call_expression_source.a:13` | retained candidate |
| `ecmascript/jsx.AttributeName` | `stage1/cohere/lint/helpers/jsx_attribute_name.a:10` | retained candidate |
| `ecmascript/jsx.HasAttributeNamed` | `stage1/cohere/lint/helpers/slot02/batch7/has_attribute_named.a:5` | retained candidate |
| `ecmascript/jsx.MatchIgnoringCase` | `stage1/cohere/lint/helpers/slot02/batch7/match_ignoring_case.a:30` | retained candidate |
| `ecmascript/property.Name` | `stage1/cohere/lint/helpers/property_name.a:20` | retained candidate |
| `ecmascript/react.IsEs5ComponentCall` | `stage1/cohere/lint/helpers/slot02/batch3/es5_component_call.a:5` | retained candidate |
| `ecmascript/react.IsNamespacedMember` | `stage1/cohere/lint/helpers/slot02/batch2/namespaced_member.a:4` | retained candidate |
| `ecmascript/react.isCreateClassName` | `stage1/cohere/lint/helpers/slot02/batch3/create_class_name.a:1` | retained candidate |
| `ecmascript/regexp.EscapeClassRune` | `stage1/cohere/lint/helpers/slot02/batch8/escape_class_rune.a:3` | retained candidate |
| `ecmascript/regexp.identityEscape` | `stage1/cohere/lint/helpers/slot02/batch8/identity_escape.a:7` | retained candidate |
| `ecmascript/regexp.literalRune` | `stage1/cohere/lint/helpers/slot02/batch8/literal_rune.a:4` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.ClassLiteralsIn` | `stage1/cohere/lint/helpers/slot02/batch2/class_literals_in.a:6` | retained candidate |
| `rules/tailwind.ClassLiteralReaderFor` | `stage1/cohere/lint/helpers/tailwind_reader_for.a:18` | retained candidate |
| `rules/tailwind.SplitClasses` | `stage1/cohere/lint/helpers/slot02/batch2/split_classes.a:7` | retained candidate |
| `rules/tailwind.declineListeners` | `stage1/cohere/lint/helpers/slot02/batch12/decline_listeners.a:26` | retained candidate |
| `rules/tailwind/collapse.*Theme.Resolve` | `stage1/cohere/lint/helpers/slot02/batch12/theme_resolve.a:7` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.ingest` | `stage1/cohere/lint/helpers/slot02/batch11/stylesheet_ingest.a:13` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.ingestCustomVariant` | `stage1/cohere/lint/helpers/slot02/batch6/ingest_custom_variant.a:16` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.loadFile` | `stage1/cohere/lint/helpers/slot02/batch6/stylesheet_load_file.a:17` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.resolveImport` | `stage1/cohere/lint/helpers/slot02/batch6/stylesheet_resolve_import.a:13` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveValueFunction` | `stage1/cohere/lint/helpers/slot02/batch13/resolve_value_function.a:12` | retained candidate |
| `rules/tailwind/collapse.DesignSystem.HasUtility` | `stage1/cohere/lint/helpers/slot02/batch5/design_system_has_utility.a:9` | retained candidate |
| `rules/tailwind/collapse.PropertySort` | `stage1/cohere/lint/helpers/slot02/batch5/property_sort.a:18` | retained candidate |
| `rules/tailwind/collapse.convertUnderscoresToWhitespace` | `stage1/cohere/lint/helpers/slot02/batch4/convert_underscores_to_whitespace.a:2` | retained candidate |
| `rules/tailwind/collapse.findRoots` | `stage1/cohere/lint/helpers/slot02/batch5/find_roots.a:7` | retained candidate |
| `rules/tailwind/collapse.isMathFunctionName` | `stage1/cohere/lint/helpers/slot02/batch3/math_function_name.a:7` | retained candidate |
| `rules/tailwind/collapse.isValidThemePrefix` | `stage1/cohere/lint/helpers/slot02/batch4/valid_theme_prefix.a:1` | retained candidate |
| `rules/tailwind/collapse.namespaceForVariantRoot` | `stage1/cohere/lint/helpers/slot02/batch4/namespace_for_variant_root.a:1` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot02_batch10_test.go:20`: TestSlot02Batch10 @20.
- `stage1/cohere/lint/helpers/slot02_batch11_test.go:20`: TestSlot02Batch11 @20.
- `stage1/cohere/lint/helpers/slot02_batch12_test.go:20`: TestSlot02Batch12 @20.
- `stage1/cohere/lint/helpers/slot02_batch13_test.go:19`: TestSlot02Batch13 @19.
- `stage1/cohere/lint/helpers/slot02_batch2_test.go:20`: TestSlot02Batch2 @20.
- `stage1/cohere/lint/helpers/slot02_batch3_test.go:20`: TestSlot02Batch3 @20.
- `stage1/cohere/lint/helpers/slot02_batch4_test.go:20`: TestSlot02Batch4 @20.
- `stage1/cohere/lint/helpers/slot02_batch5_test.go:20`: TestSlot02Batch5 @20.
- `stage1/cohere/lint/helpers/slot02_batch6_test.go:20`: TestSlot02Batch6 @20.
- `stage1/cohere/lint/helpers/slot02_batch7_test.go:21`: TestSlot02Batch7 @21.
- `stage1/cohere/lint/helpers/slot02_batch8_test.go:20`: TestSlot02Batch8 @20.
- `stage1/cohere/lint/helpers/slot02_batch9_test.go:20`: TestSlot02Batch9 @20.
- `stage1/cohere/lint/helpers/slot02_test.go:152`: TestSlot02AttributeName @152, TestSlot02PropertyName @155, TestSlot02ReaderFor @210.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch10/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch11/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch12/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch13/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch2/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch3/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch4/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch5/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch6/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch7/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch8/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/batch9/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-48838501/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-4e0bfda5/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-area-d65a8f93/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-b4691483/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-bb2ece56/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-c01907a7/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-d3a37422/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-e667e3e1/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing-f8013f0/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/landing/REPORT.md:1`, `stage1/cohere/lint/helpers/slot02/regexp-compile-blocker/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/slot02/landing-b4691483/testdata/OctalEscape.a.txt`, `stage1/cohere/lint/helpers/slot02/landing-d3a37422/testdata/OctalEscape.a.txt`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 446 files; +45,106/-0 lines; 4,514,283 blob bytes.
- implementation/support types: 44 files; +2,425/-0 lines; 72,174 blob bytes.
- test/oracle/capture programs: 80 files; +8,205/-0 lines; 320,314 blob bytes.

### origin/codex/lint-helpers-03

Tip `3e85502d728bcd67b3ed326565bcf09749f2ec1b`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** 63 retained helpers; helper/mutant tests exist for all 21 batches. Consumer fixtures plus actual private-Go overlays and controls are used. Latest batches cover Next path and CFG handoffs with callback boundaries. Tailwind upstream capture/gate is red for missing Kirk-local theme corpus (batch21 report: eight corpus-population failures); isolated fixture projections and explicit controls are not a successful full six-consumer gate. Coverage guards distinguish this from a source match.

Claim: `stage1/cohere/lint/helpers/claims/03.md:1` (364 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].binaryExpression` | `stage1/cohere/lint/helpers/slot03/batch18/binary_expression.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].bindWithDefault` | `stage1/cohere/lint/helpers/slot03/batch17/bind_with_default.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].classLike` | `stage1/cohere/lint/helpers/slot03/batch18/class_like.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].conditionalExpression` | `stage1/cohere/lint/helpers/slot03/batch17/conditional_expression.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].expr` | `stage1/cohere/lint/helpers/slot03/batch19/expression.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].ifStatement` | `stage1/cohere/lint/helpers/slot03/batch18/if_statement.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeReturn` | `stage1/cohere/lint/helpers/slot03/batch16/make_return.a:5` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeThrow` | `stage1/cohere/lint/helpers/slot03/batch16/make_throw.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeUnreachable` | `stage1/cohere/lint/helpers/slot03/batch15/make_unreachable.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeYield` | `stage1/cohere/lint/helpers/slot03/batch16/make_yield.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].memberHeader` | `stage1/cohere/lint/helpers/slot03/batch17/member_header.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].patternBind` | `stage1/cohere/lint/helpers/slot03/batch19/pattern_bind.a:5` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].popJump` | `stage1/cohere/lint/helpers/slot03/batch15/pop_jump.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].pushJump` | `stage1/cohere/lint/helpers/slot03/batch15/push_jump.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].switchStatement` | `stage1/cohere/lint/helpers/slot03/batch19/switch_statement.a:5` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].tryStatement` | `stage1/cohere/lint/helpers/slot03/batch20/try_statement.a:6` | retained candidate |
| `ecmascript/imports.HasPathSegment` | `stage1/cohere/lint/helpers/slot03/batch21/has_path_segment.a:1` | retained candidate |
| `ecmascript/jsx.IsIntrinsicElementNamed` | `stage1/cohere/lint/helpers/slot03/batch2/intrinsic_element_named.a:1` | retained candidate |
| `ecmascript/nextjs.IsDocumentFile` | `stage1/cohere/lint/helpers/slot03/batch20/is_document_file.a:2` | retained candidate |
| `ecmascript/nextjs.lastSeparator` | `stage1/cohere/lint/helpers/slot03/batch20/last_separator.a:2` | retained candidate |
| `ecmascript/nextjs.splitPath` | `stage1/cohere/lint/helpers/slot03/batch21/split_path.a:2` | retained candidate |
| `ecmascript/react.isComponentBaseName` | `stage1/cohere/lint/helpers/slot03/component_base_name.a:2` | retained candidate |
| `ecmascript/text.UnescapeStringLiteralText` | `stage1/cohere/lint/helpers/slot03/batch3/unescape_string_literal_text.a:2` | retained candidate |
| `ecmascript/text.hexValue` | `stage1/cohere/lint/helpers/slot03/batch3/hex_value.a:1` | retained candidate |
| `rules/structure.parameterNodes` | `stage1/cohere/lint/helpers/slot03/batch3/parameter_nodes.a:1` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.readClassValues` | `stage1/cohere/lint/helpers/slot03/batch2/read_class_values.a:3` | retained candidate |
| `rules/tailwind.DesignSystemDeclineMessage` | `stage1/cohere/lint/helpers/slot03/batch9/design_system_decline_message.a:2` | retained candidate |
| `rules/tailwind.ListenerKinds` | `stage1/cohere/lint/helpers/slot03/listener_kinds.a:2` | retained candidate |
| `rules/tailwind.holeEdges` | `stage1/cohere/lint/helpers/slot03/batch2/hole_edges.a:3` | retained candidate |
| `rules/tailwind.isSpace` | `stage1/cohere/lint/helpers/slot03/tailwind_space.a:3` | retained candidate |
| `rules/tailwind/collapse.*LoadedDesignSystem.Theme` | `stage1/cohere/lint/helpers/slot03/batch9/loaded_theme.a:3` | retained candidate |
| `rules/tailwind/collapse.*LoadedDesignSystem.Utilities` | `stage1/cohere/lint/helpers/slot03/batch11/loaded_utilities.a:3` | retained candidate |
| `rules/tailwind/collapse.*VariantRegistry.AttachComparison` | `stage1/cohere/lint/helpers/slot03/batch7/attach_comparison.a:2` | retained candidate |
| `rules/tailwind/collapse.*VariantRegistry.Register` | `stage1/cohere/lint/helpers/slot03/batch7/register.a:2` | retained candidate |
| `rules/tailwind/collapse.*VariantRegistry.nextOrder` | `stage1/cohere/lint/helpers/slot03/batch6/variant_next_order.a:2` | retained candidate |
| `rules/tailwind/collapse.AtRule` | `stage1/cohere/lint/helpers/slot03/batch6/at_rule.a:2` | retained candidate |
| `rules/tailwind/collapse.InferDataType` | `stage1/cohere/lint/helpers/slot03/batch13/infer_data_type.a:2` | retained candidate |
| `rules/tailwind/collapse.NewVariantRegistry` | `stage1/cohere/lint/helpers/slot03/batch7/new_variant_registry.a:2` | retained candidate |
| `rules/tailwind/collapse.ParseCSS` | `stage1/cohere/lint/helpers/slot03/batch9/parse_css.a:4` | retained candidate |
| `rules/tailwind/collapse.StyleRule` | `stage1/cohere/lint/helpers/slot03/batch6/style_rule.a:2` | retained candidate |
| `rules/tailwind/collapse.breakpointGroupOrder` | `stage1/cohere/lint/helpers/slot03/batch5/breakpoint_group_order.a:4` | retained candidate |
| `rules/tailwind/collapse.decodeArbitraryValue` | `stage1/cohere/lint/helpers/slot03/batch8/decode_arbitrary_value.a:4` | retained candidate |
| `rules/tailwind/collapse.gapRootAcceptsModifierOnArbitrary` | `stage1/cohere/lint/helpers/slot03/batch21/gap_root_accepts_modifier.a:1` | retained candidate |
| `rules/tailwind/collapse.hasMathFunction` | `stage1/cohere/lint/helpers/slot03/batch11/has_math_function.a:2` | retained candidate |
| `rules/tailwind/collapse.isAbsoluteSize` | `stage1/cohere/lint/helpers/slot03/batch10/is_absolute_size.a:2` | retained candidate |
| `rules/tailwind/collapse.isAngle` | `stage1/cohere/lint/helpers/slot03/batch12/is_angle.a:2` | retained candidate |
| `rules/tailwind/collapse.isBackgroundPosition` | `stage1/cohere/lint/helpers/slot03/batch14/is_background_position.a:2` | retained candidate |
| `rules/tailwind/collapse.isEscapeTerminator` | `stage1/cohere/lint/helpers/slot03/batch4/escape_terminator.a:1` | retained candidate |
| `rules/tailwind/collapse.isFamilyName` | `stage1/cohere/lint/helpers/slot03/batch13/is_family_name.a:2` | retained candidate |
| `rules/tailwind/collapse.isFollowedByWhitespace` | `stage1/cohere/lint/helpers/slot03/batch4/followed_by_whitespace.a:2` | retained candidate |
| `rules/tailwind/collapse.isGenericName` | `stage1/cohere/lint/helpers/slot03/batch11/is_generic_name.a:2` | retained candidate |
| `rules/tailwind/collapse.isIgnoredThemeKey` | `stage1/cohere/lint/helpers/slot03/batch4/ignored_theme_key.a:1` | retained candidate |
| `rules/tailwind/collapse.isImage` | `stage1/cohere/lint/helpers/slot03/batch14/is_image.a:2` | retained candidate |
| `rules/tailwind/collapse.isLineWidth` | `stage1/cohere/lint/helpers/slot03/batch14/is_line_width.a:2` | retained candidate |
| `rules/tailwind/collapse.isNumber` | `stage1/cohere/lint/helpers/slot03/batch12/is_number.a:3` | retained candidate |
| `rules/tailwind/collapse.isPercentage` | `stage1/cohere/lint/helpers/slot03/batch12/is_percentage.a:3` | retained candidate |
| `rules/tailwind/collapse.isRelativeSize` | `stage1/cohere/lint/helpers/slot03/batch10/is_relative_size.a:2` | retained candidate |
| `rules/tailwind/collapse.isURL` | `stage1/cohere/lint/helpers/slot03/batch10/is_url.a:2` | retained candidate |
| `rules/tailwind/collapse.joinSegments` | `stage1/cohere/lint/helpers/slot03/batch5/join_segments.a:1` | retained candidate |
| `rules/tailwind/collapse.matchesDataType` | `stage1/cohere/lint/helpers/slot03/batch13/matches_data_type.a:2` | retained candidate |
| `rules/tailwind/collapse.recursivelyDecodeArbitraryValues` | `stage1/cohere/lint/helpers/slot03/batch8/recursively_decode_arbitrary_values.a:3` | retained candidate |
| `rules/tailwind/collapse.registerThemeBreakpointVariants` | `stage1/cohere/lint/helpers/slot03/batch8/register_theme_breakpoint_variants.a:3` | retained candidate |
| `rules/tailwind/collapse.splitThemeKey` | `stage1/cohere/lint/helpers/slot03/batch5/split_theme_key.a:1` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot03/batch10_test.go:40`: TestBatch10Helpers @40, TestBatch10Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch11_test.go:40`: TestBatch11Helpers @40, TestBatch11Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch12_test.go:66`: TestBatch12Helpers @66, TestBatch12Mutants @79, TestBatch12ArgumentMutants @130.
- `stage1/cohere/lint/helpers/slot03/batch13_test.go:153`: TestBatch13Helpers @153, TestBatch13Mutants @166.
- `stage1/cohere/lint/helpers/slot03/batch14_test.go:66`: TestBatch14Helpers @66, TestBatch14Mutants @79.
- `stage1/cohere/lint/helpers/slot03/batch15_test.go:69`: TestBatch15Helpers @69, TestBatch15Mutants @82, TestBatch15EmptyStackRefusal @153.
- `stage1/cohere/lint/helpers/slot03/batch16_test.go:67`: TestBatch16Helpers @67, TestBatch16Mutants @80.
- `stage1/cohere/lint/helpers/slot03/batch17_test.go:75`: TestBatch17Helpers @75, TestBatch17Mutants @88.
- `stage1/cohere/lint/helpers/slot03/batch18_test.go:83`: TestBatch18Helpers @83, TestBatch18Mutants @96.
- `stage1/cohere/lint/helpers/slot03/batch19_test.go:95`: TestBatch19Helpers @95, TestBatch19Mutants @108.
- `stage1/cohere/lint/helpers/slot03/batch20_test.go:81`: TestBatch20Helpers @81, TestBatch20Mutants @94.
- `stage1/cohere/lint/helpers/slot03/batch21_test.go:62`: TestBatch21Helpers @62, TestBatch21Mutants @75.
- `stage1/cohere/lint/helpers/slot03/batch2_test.go:52`: TestBatch2Helpers @52, TestBatch2Mutants @65.
- `stage1/cohere/lint/helpers/slot03/batch3_test.go:54`: TestBatch3Helpers @54, TestBatch3Mutants @67.
- `stage1/cohere/lint/helpers/slot03/batch4_test.go:40`: TestBatch4Helpers @40, TestBatch4Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch5_test.go:40`: TestBatch5Helpers @40, TestBatch5Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch6_test.go:40`: TestBatch6Helpers @40, TestBatch6Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch7_test.go:40`: TestBatch7Helpers @40, TestBatch7Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch8_test.go:40`: TestBatch8Helpers @40, TestBatch8Mutants @53.
- `stage1/cohere/lint/helpers/slot03/batch9_test.go:57`: TestBatch9Helpers @57, TestBatch9Mutants @70.
- `stage1/cohere/lint/helpers/slot03/slot03_test.go:96`: TestSlot03HelpersMatchCohere @96, TestSlot03HelperMutants @119, TestConsumerCoverageRejectsMutant @271.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch10/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch11/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch12/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch13/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch14/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch15/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch16/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch17/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch18/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch19/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch2/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch20/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch21/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch3/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch4/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch5/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch6/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch7/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch8/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/batch9/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/landing-evidence/REPORT.md:1`, `stage1/cohere/lint/helpers/slot03/landing-harness/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/slot03/batch17/evidence/withdrawn-search.log.gz`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 423 files; +39,053/-0 lines; 5,342,133 blob bytes.
- implementation/support types: 68 files; +802/-0 lines; 50,101 blob bytes.
- test/oracle/capture programs: 112 files; +9,954/-0 lines; 428,076 blob bytes.

### origin/codex/lint-helpers-04

Tip `8e0d98a645ae73cc4cf72561c779f257b51b8c9a`; merge-base with area `bb2ece564842c4b2f909b9f75c27e74c2efa4f29`; merge-tree clean (exit 0).

**Coverage verdict:** 66 retained helpers. ElementParts and early packages have Go/source/native tests; later wave harnesses add emitted JavaScript and semantic mutants. Every source tip merges cleanly today, but package composition is not proven. Build/utility Compile use caller-owned identity, clone/walk/remove callbacks; DecodeCompilerRuleOptions intentionally accepts only the empty object using external decode/key-sort facts. resolveBareArgument was not reached by the captured upstream suites, so only direct Go controls hold that path. expr/patternBind are withdrawn to 03.

Claim: `stage1/cohere/lint/helpers/claims/04.md:1` (557 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.Build` | `stage1/cohere/lint/helpers/slot04_wave23/build.a:5` | retained candidate |
| `ecmascript/jsx.ElementParts` | `stage1/cohere/lint/helpers/jsx_element_parts.a:8` | retained candidate |
| `ecmascript/module.HasExportModifier` | `stage1/cohere/lint/helpers/slot04_wave10/has_export_modifier.a:2` | retained candidate |
| `ecmascript/module.IsExported` | `stage1/cohere/lint/helpers/slot04_wave10/is_exported.a:3` | retained candidate |
| `ecmascript/module.IsExportedByName` | `stage1/cohere/lint/helpers/slot04_wave10/is_exported_by_name.a:2` | retained candidate |
| `ecmascript/regexp.checkGroupConstruct` | `stage1/cohere/lint/helpers/slot04_wave13/group_construct.a:13` | retained candidate |
| `ecmascript/regexp.checkGroupQuantifier` | `stage1/cohere/lint/helpers/slot04_wave13/group_quantifier.a:2` | retained candidate |
| `ecmascript/regexp.errNothingToRepeat` | `stage1/cohere/lint/helpers/slot04_wave13/nothing_to_repeat.a:1` | retained candidate |
| `rules/react.DecodeCompilerRuleOptions` | `stage1/cohere/lint/helpers/slot04_wave22/compiler_rule_options.a:4` | retained candidate |
| `rules/structure.HasJsxOrReactHookCalls` | `stage1/cohere/lint/helpers/slot04_wave20/has_jsx_or_react_hook_calls.a:3` | retained candidate |
| `rules/structure.descendsForJsxSearch` | `stage1/cohere/lint/helpers/slot04_wave20/descends_for_jsx_search.a:1` | retained candidate |
| `rules/structure.searchForJsxOrHook` | `stage1/cohere/lint/helpers/slot04_wave20/search_for_jsx_or_hook.a:4` | retained candidate |
| `rules/tailwind.ClassLiteralSettings.key` | `stage1/cohere/lint/helpers/tailwind_settings_key.a:4` | retained candidate |
| `rules/tailwind.classTokensIn` | `stage1/cohere/lint/helpers/slot04_wave21/class_tokens_in.a:4` | retained candidate |
| `rules/tailwind.classTokensOf` | `stage1/cohere/lint/helpers/slot04_wave21/class_tokens_of.a:3` | retained candidate |
| `rules/tailwind.classValuesUnder` | `stage1/cohere/lint/helpers/slot04_wave2/class_values_under.a:3` | retained candidate |
| `rules/tailwind.collectClassValues` | `stage1/cohere/lint/helpers/slot04_wave2/collect_class_values.a:3` | retained candidate |
| `rules/tailwind.compiledClassLiteralReader` | `stage1/cohere/lint/helpers/tailwind_compiled_reader.a:7` | retained candidate |
| `rules/tailwind.loadDesignSystemThrough` | `stage1/cohere/lint/helpers/slot04_wave9/load_design_system_through.a:24` | retained candidate |
| `rules/tailwind/collapse.*Theme.PrefixKey` | `stage1/cohere/lint/helpers/slot04_wave6/theme_prefix_key.a:3` | retained candidate |
| `rules/tailwind/collapse.*Theme.ResolveWith` | `stage1/cohere/lint/helpers/slot04_wave18/resolve_with.a:5` | retained candidate |
| `rules/tailwind/collapse.*Theme.variableReference` | `stage1/cohere/lint/helpers/slot04_wave16/variable_reference.a:3` | retained candidate |
| `rules/tailwind/collapse.*UtilityEvaluator.Compile` | `stage1/cohere/lint/helpers/slot04_wave22/compile_candidate.a:8` | retained candidate |
| `rules/tailwind/collapse.*UtilityEvaluator.compile` | `stage1/cohere/lint/helpers/slot04_wave22/compile.a:4` | retained candidate |
| `rules/tailwind/collapse.*VariantRegistry.Has` | `stage1/cohere/lint/helpers/slot04_wave6/variant_registry_has.a:2` | retained candidate |
| `rules/tailwind/collapse.*stylesheetCollector.ingestUtilityBlock` | `stage1/cohere/lint/helpers/slot04_wave9/ingest_utility_block.a:11` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.modifierAsValue` | `stage1/cohere/lint/helpers/slot04_wave16/modifier_value.a:3` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveArbitraryArgument` | `stage1/cohere/lint/helpers/slot04_wave16/arbitrary_argument.a:4` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveBareArgument` | `stage1/cohere/lint/helpers/slot04_wave21/bare_argument.a:2` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveDeclaration` | `stage1/cohere/lint/helpers/slot04_wave19/resolve_declaration.a:1` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveThemeArgument` | `stage1/cohere/lint/helpers/slot04_wave18/theme_argument.a:2` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.resolveValueFunctions` | `stage1/cohere/lint/helpers/slot04_wave19/resolve_value_functions.a:3` | retained candidate |
| `rules/tailwind/collapse.*utilityEvaluation.walk` | `stage1/cohere/lint/helpers/slot04_wave19/walk.a:2` | retained candidate |
| `rules/tailwind/collapse.Comment` | `stage1/cohere/lint/helpers/slot04_wave5/comment.a:3` | retained candidate |
| `rules/tailwind/collapse.Declaration` | `stage1/cohere/lint/helpers/slot04_wave5/declaration.a:3` | retained candidate |
| `rules/tailwind/collapse.DesignSystem.HasVariant` | `stage1/cohere/lint/helpers/slot04_wave7/has_variant.a:2` | retained candidate |
| `rules/tailwind/collapse.DesignSystem.Prefix` | `stage1/cohere/lint/helpers/slot04_wave6/design_system_prefix.a:2` | retained candidate |
| `rules/tailwind/collapse.DesignSystem.VariantCompoundsWith` | `stage1/cohere/lint/helpers/slot04_wave7/variant_compounds_with.a:3` | retained candidate |
| `rules/tailwind/collapse.DesignSystem.VariantKind` | `stage1/cohere/lint/helpers/slot04_wave7/variant_kind.a:2` | retained candidate |
| `rules/tailwind/collapse.IsColor` | `stage1/cohere/lint/helpers/slot04_wave17/color.a:5` | retained candidate |
| `rules/tailwind/collapse.IsLength` | `stage1/cohere/lint/helpers/slot04_wave17/length.a:3` | retained candidate |
| `rules/tailwind/collapse.IsNamedColor` | `stage1/cohere/lint/helpers/slot04_wave17/named_color.a:3` | retained candidate |
| `rules/tailwind/collapse.IsPositiveInteger` | `stage1/cohere/lint/helpers/slot04_wave15/positive_integer.a:4` | retained candidate |
| `rules/tailwind/collapse.Walk` | `stage1/cohere/lint/helpers/slot04_wave8/walk.a:3` | retained candidate |
| `rules/tailwind/collapse.breakpointBucket` | `stage1/cohere/lint/helpers/slot04_wave5/breakpoint_bucket.a:3` | retained candidate |
| `rules/tailwind/collapse.formatJavaScriptNumber` | `stage1/cohere/lint/helpers/slot04_wave14/format_number.a:1` | retained candidate |
| `rules/tailwind/collapse.hasAnyPrefix` | `stage1/cohere/lint/helpers/slot04_wave11/any_prefix.a:1` | retained candidate |
| `rules/tailwind/collapse.isBackgroundSize` | `stage1/cohere/lint/helpers/slot04_wave18/background_size.a:1` | retained candidate |
| `rules/tailwind/collapse.isBlank` | `stage1/cohere/lint/helpers/slot04_wave3_space/blank.a:2` | retained candidate |
| `rules/tailwind/collapse.isBracketed` | `stage1/cohere/lint/helpers/slot04_wave11/bracketed.a:1` | retained candidate |
| `rules/tailwind/collapse.isDigit` | `stage1/cohere/lint/helpers/slot04_wave11/digit.a:1` | retained candidate |
| `rules/tailwind/collapse.isJavaScriptSpace` | `stage1/cohere/lint/helpers/slot04_wave3_space/javascript_space.a:2` | retained candidate |
| `rules/tailwind/collapse.isMultipleOf` | `stage1/cohere/lint/helpers/slot04_wave14/multiple_of.a:4` | retained candidate |
| `rules/tailwind/collapse.isQuotedLiteral` | `stage1/cohere/lint/helpers/slot04_wave12/quoted_literal.a:1` | retained candidate |
| `rules/tailwind/collapse.isValidSpacingMultiplier` | `stage1/cohere/lint/helpers/slot04_wave15/spacing_multiplier.a:3` | retained candidate |
| `rules/tailwind/collapse.isValueSeparator` | `stage1/cohere/lint/helpers/slot04_wave3_space/value_separator.a:2` | retained candidate |
| `rules/tailwind/collapse.modFloat` | `stage1/cohere/lint/helpers/slot04_wave14/mod_float.a:1` | retained candidate |
| `rules/tailwind/collapse.normalizeValueFunctionNodes` | `stage1/cohere/lint/helpers/slot04_wave9/normalize_value_function_nodes.a:14` | retained candidate |
| `rules/tailwind/collapse.peekByte` | `stage1/cohere/lint/helpers/slot04_wave4/peek_byte.a:3` | retained candidate |
| `rules/tailwind/collapse.replaceValueNode` | `stage1/cohere/lint/helpers/slot04_wave12/replace_value_node.a:7` | retained candidate |
| `rules/tailwind/collapse.roundTripsAsJavaScriptNumber` | `stage1/cohere/lint/helpers/slot04_wave15/round_trip.a:3` | retained candidate |
| `rules/tailwind/collapse.sortedKeys` | `stage1/cohere/lint/helpers/slot04_wave4/sorted_keys.a:4` | retained candidate |
| `rules/tailwind/collapse.topOfStack` | `stage1/cohere/lint/helpers/slot04_wave4/top_of_stack.a:2` | retained candidate |
| `rules/tailwind/collapse.unionNodeSets` | `stage1/cohere/lint/helpers/slot04_wave12/union_node_sets.a:5` | retained candidate |
| `rules/tailwind/collapse.walkNodes` | `stage1/cohere/lint/helpers/slot04_wave8/walk_nodes.a:3` | retained candidate |
| `rules/tailwind/collapse.writeValueCss` | `stage1/cohere/lint/helpers/slot04_wave8/write_value_css.a:3` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot04_test.go:61`: TestSlot04ElementParts @61, TestSlot04ElementPartsMutant @71, TestSlot04ElementShapes @94, TestSlot04SettingsKey @118, TestSlot04SettingsKeyMutant @133, TestSlot04CompiledReader @186, TestSlot04CompiledReaderMutant @201, TestSlot04ConsumerCoverage @285.
- `stage1/cohere/lint/helpers/slot04_wave10/helpers_test.go:115`: TestExportsGoNodeNativeJavaScript @115, TestCompilingMutants @133, TestConsumerCoverage @217.
- `stage1/cohere/lint/helpers/slot04_wave11/helpers_test.go:116`: TestBytesGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @221.
- `stage1/cohere/lint/helpers/slot04_wave12/helpers_test.go:116`: TestSurgeryGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @232.
- `stage1/cohere/lint/helpers/slot04_wave13/helpers_test.go:116`: TestAssertionsGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @221.
- `stage1/cohere/lint/helpers/slot04_wave14/helpers_test.go:116`: TestNumericGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @221.
- `stage1/cohere/lint/helpers/slot04_wave15/helpers_test.go:116`: TestIntegerGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @235.
- `stage1/cohere/lint/helpers/slot04_wave16/helpers_test.go:118`: TestUtilityResolutionGoNodeNativeJavaScript @118, TestCompilingMutants @139, TestConsumerCoverage @227.
- `stage1/cohere/lint/helpers/slot04_wave17/helpers_test.go:116`: TestColorLengthGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @227.
- `stage1/cohere/lint/helpers/slot04_wave18/helpers_test.go:116`: TestNestedThemeGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @227.
- `stage1/cohere/lint/helpers/slot04_wave19/helpers_test.go:116`: TestUtilityTraversalGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @227.
- `stage1/cohere/lint/helpers/slot04_wave2/wave2_test.go:92`: TestWave2GoNodeNative @92, TestWave2CompilingMutants @109, TestWave2ConsumerCoverage @203.
- `stage1/cohere/lint/helpers/slot04_wave20/helpers_test.go:116`: TestJsxSearchGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @224.
- `stage1/cohere/lint/helpers/slot04_wave21/helpers_test.go:116`: TestClassTokenBareGoNodeNativeJavaScript @116, TestCompilingMutants @137, TestConsumerCoverage @228.
- `stage1/cohere/lint/helpers/slot04_wave22/helpers_test.go:116`: TestActualGoComparisons @116, TestCompilingMutants @132, TestConsumerCoverage @229.
- `stage1/cohere/lint/helpers/slot04_wave23/helpers_test.go:114`: TestActualGoComparisons @114, TestCompilingMutants @149, TestConsumerCoverage @241.
- `stage1/cohere/lint/helpers/slot04_wave3_space/space_test.go:115`: TestSpaceGoNodeNativeJavaScript @115, TestSpaceCompilingMutants @139, TestSpaceConsumerCoverage @229.
- `stage1/cohere/lint/helpers/slot04_wave4/bytes_test.go:116`: TestBytesGoNodeNativeJavaScript @116, TestBytesCompilingMutants @140, TestBytesConsumerCoverage @228, TestSortedKeys @282, TestSortedKeysCompilingMutants @295.
- `stage1/cohere/lint/helpers/slot04_wave5/strings_test.go:116`: TestStringsGoNodeNativeJavaScript @116, TestStringsCompilingMutants @140, TestStringsConsumerCoverage @232.
- `stage1/cohere/lint/helpers/slot04_wave6/helpers_test.go:116`: TestStringsGoNodeNativeJavaScript @116, TestStringsCompilingMutants @140, TestStringsConsumerCoverage @230, TestPrefixKeyShortKeyRefusalAndMutant @281.
- `stage1/cohere/lint/helpers/slot04_wave7/helpers_test.go:116`: TestStringsGoNodeNativeJavaScript @116, TestStringsCompilingMutants @140, TestStringsConsumerCoverage @231.
- `stage1/cohere/lint/helpers/slot04_wave8/helpers_test.go:116`: TestStringsGoNodeNativeJavaScript @116, TestStringsCompilingMutants @140, TestStringsConsumerCoverage @232.
- `stage1/cohere/lint/helpers/slot04_wave9/helpers_test.go:118`: TestHelpersGoNodeNativeJavaScript @118, TestCompilingMutants @136, TestLoaderBodyMatchesPinnedGo @184, TestConsumerCoverage @251.
- `stage1/cohere/lint/helpers/slot04_wave16/testdata/generate.py:1`: owned helper/semantic mutation validation script; comparison assertions at .

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_20261007/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_area_d65a8f93/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_b4691483/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_b84a9d93/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_b8fb957/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_d3a37422/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_landing_f8013f0/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_regexp_gap/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_unified_landing/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave10/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave11/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave12/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave13/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave14/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave15/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave16/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave17/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave18/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave19/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave2/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave20/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave21/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave22/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave23/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave3_space/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave4/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave5/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave6/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave7/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave8/REPORT.md:1`, `stage1/cohere/lint/helpers/slot04_wave9/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 462 files; +578,991/-0 lines; 15,107,943 blob bytes.
- implementation/support types: 75 files; +856/-0 lines; 47,799 blob bytes.
- test/oracle/capture programs: 139 files; +13,199/-0 lines; 514,807 blob bytes.

### origin/codex/lint-helpers-05

Tip `cd24dd580a8e8d97cdc1cf0aa839b47bf333e9bc`; merge-base with area `65017b318da1995237ff3ea2c80f59b055b39ac3`; merge-tree clean (exit 0).

**Coverage verdict:** 95 retained helpers plus BLOCKED regexp.Compile witness. Latest landing_7945d102 evidence supersedes earlier batch34/LANDING.md pin-drift failure: source tests now pin current cohere, and the claim reports 34 package observations/467 own semantic mutants. Inspect current test loops: most verify variants run the mutated native binary only, while unmodified baselines compare Go/source Node/emitted JS/native. Do not restate this as 467 three-backend mutant catches. Ports remain projected/callback contracts. Nonconstant RegExp construction remains explicitly refused; no Compile body or readiness credit.

Claim: `stage1/cohere/lint/helpers/claims/05.md:1` (517 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].accessOrCall` | `stage1/cohere/lint/helpers/slot05/batch20/access_or_call.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].doStatement` | `stage1/cohere/lint/helpers/slot05/batch19/do_statement.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].firstThrowableFork` | `stage1/cohere/lint/helpers/slot05/batch16/first_throwable_fork.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].forInOfStatement` | `stage1/cohere/lint/helpers/slot05/batch20/for_in_of_statement.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].forStatement` | `stage1/cohere/lint/helpers/slot05/batch20/for_statement.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].forkOptionalChain` | `stage1/cohere/lint/helpers/slot05/batch17/fork_optional_chain.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].labeledStatement` | `stage1/cohere/lint/helpers/slot05/batch19/labeled_statement.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeBreak` | `stage1/cohere/lint/helpers/slot05/batch14/make_break.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].makeContinue` | `stage1/cohere/lint/helpers/slot05/batch15/make_continue.a:3` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].parameter` | `stage1/cohere/lint/helpers/slot05/batch16/parameter.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].patternReads` | `stage1/cohere/lint/helpers/slot05/batch18/pattern_reads.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].patternWrites` | `stage1/cohere/lint/helpers/slot05/batch18/pattern_writes.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].statements` | `stage1/cohere/lint/helpers/slot05/batch15/statements.a:1` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].updateExpression` | `stage1/cohere/lint/helpers/slot05/batch16/update_expression.a:1` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].visitUnknown` | `stage1/cohere/lint/helpers/slot05/batch17/visit_unknown.a:1` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].whileStatement` | `stage1/cohere/lint/helpers/slot05/batch19/while_statement.a:2` | retained candidate |
| `ecmascript/imports.BindingsOf` | `stage1/cohere/lint/helpers/slot05/import_bindings.a:12` | retained candidate |
| `ecmascript/imports.ImportedNameOf` | `stage1/cohere/lint/helpers/slot05/batch9/imported_name_of.a:3` | retained candidate |
| `ecmascript/imports.NormalizedFileName` | `stage1/cohere/lint/helpers/slot05/batch3/imports_normalized_file_name.a:5` | retained candidate |
| `ecmascript/jsx.MatchExactly` | `stage1/cohere/lint/helpers/slot05/batch3/jsx_match_exactly.a:2` | retained candidate |
| `ecmascript/react.IsHookCall` | `stage1/cohere/lint/helpers/slot05/batch17/react_is_hook_call.a:2` | retained candidate |
| `ecmascript/react.IsHookName` | `stage1/cohere/lint/helpers/slot05/batch3/react_is_hook_name.a:159` | retained candidate |
| `ecmascript/react.isComponentBase` | `stage1/cohere/lint/helpers/slot05/react_component_base.a:8` | retained candidate |
| `ecmascript/react.isIdentifierNamed` | `stage1/cohere/lint/helpers/slot05/react_identifier_named.a:6` | retained candidate |
| `ecmascript/regexp.*RegExp.Test` | `stage1/cohere/lint/helpers/slot05/batch23/regexp_test.a:3` | retained candidate |
| `ecmascript/regexp.Canonicalize` | `stage1/cohere/lint/helpers/slot05/batch13/canonicalize.a:1` | retained candidate |
| `ecmascript/regexp.CaseClass` | `stage1/cohere/lint/helpers/slot05/batch13/case_class.a:2` | retained candidate |
| `ecmascript/regexp.CaseEquivalenceGroups` | `stage1/cohere/lint/helpers/slot05/batch12/case_equivalence_groups.a:2` | retained candidate |
| `ecmascript/regexp.CaseEquivalents` | `stage1/cohere/lint/helpers/slot05/batch12/case_equivalents.a:2` | retained candidate |
| `ecmascript/regexp.buildCaseTables` | `stage1/cohere/lint/helpers/slot05/batch21/build_case_tables.a:8` | retained candidate |
| `ecmascript/regexp.caseExtras` | `stage1/cohere/lint/helpers/slot05/batch21/case_extras.a:6` | retained candidate |
| `ecmascript/regexp.caseTables` | `stage1/cohere/lint/helpers/slot05/batch12/case_tables.a:2` | retained candidate |
| `ecmascript/regexp.classAtom.write` | `stage1/cohere/lint/helpers/slot05/batch11/class_atom_write.a:2` | retained candidate |
| `ecmascript/regexp.classAtoms` | `stage1/cohere/lint/helpers/slot05/batch23/class_atoms.a:2` | retained candidate |
| `ecmascript/regexp.countGroups` | `stage1/cohere/lint/helpers/slot05/batch9/regexp_count_groups.a:4` | retained candidate |
| `ecmascript/regexp.decodeControlEscape` | `stage1/cohere/lint/helpers/slot05/batch10/decode_control_escape.a:17` | retained candidate |
| `ecmascript/regexp.decodeEscape` | `stage1/cohere/lint/helpers/slot05/batch23/decode_escape.a:3` | retained candidate |
| `ecmascript/regexp.decodeNumericEscape` | `stage1/cohere/lint/helpers/slot05/batch22/decode_numeric_escape.a:4` | retained candidate |
| `ecmascript/regexp.decodePropertyEscape` | `stage1/cohere/lint/helpers/slot05/batch22/decode_property_escape.a:4` | retained candidate |
| `ecmascript/regexp.expandsOnUppercase` | `stage1/cohere/lint/helpers/slot05/batch11/expands_on_uppercase.a:1` | retained candidate |
| `ecmascript/regexp.isDecimalDigit` | `stage1/cohere/lint/helpers/slot05/batch9/regexp_is_decimal_digit.a:3` | retained candidate |
| `ecmascript/regexp.joinRanges` | `stage1/cohere/lint/helpers/slot05/batch21/join_ranges.a:4` | retained candidate |
| `ecmascript/regexp.namedBackreference` | `stage1/cohere/lint/helpers/slot05/batch10/named_backreference.a:3` | retained candidate |
| `ecmascript/regexp.namedGroupOpener` | `stage1/cohere/lint/helpers/slot05/batch10/named_group_opener.a:3` | retained candidate |
| `ecmascript/regexp.readClass` | `stage1/cohere/lint/helpers/slot05/batch22/read_class.a:3` | retained candidate |
| `ecmascript/regexp.rewrite` | `stage1/cohere/lint/helpers/slot05/batch24/rewrite.a:12` | retained candidate |
| `ecmascript/regexp.wordCharacters` | `stage1/cohere/lint/helpers/slot05/batch11/word_characters.a:2` | retained candidate |
| `ecmascript/regexp.writeClass` | `stage1/cohere/lint/helpers/slot05/batch13/write_class.a:3` | retained candidate |
| `ecmascript/regexsyntax.AllHexDigits` | `stage1/cohere/lint/helpers/slot05/batch24/all_hex_digits.a:3` | retained candidate |
| `ecmascript/regexsyntax.ClassEnd` | `stage1/cohere/lint/helpers/slot05/batch26/class_end.a:4` | retained candidate |
| `ecmascript/regexsyntax.IsHexDigit` | `stage1/cohere/lint/helpers/slot05/batch24/is_hex_digit.a:2` | retained candidate |
| `ecmascript/regexsyntax.ParseRegexFlags` | `stage1/cohere/lint/helpers/slot05/batch25/parse_regex_flags.a:4` | retained candidate |
| `ecmascript/regexsyntax.PatternAndFlags` | `stage1/cohere/lint/helpers/slot05/batch25/pattern_and_flags.a:4` | retained candidate |
| `ecmascript/regexsyntax.RegexFlags.UV` | `stage1/cohere/lint/helpers/slot05/batch25/regex_flags_uv.a:2` | retained candidate |
| `ecmascript/regexsyntax.SkipPatternEscape` | `stage1/cohere/lint/helpers/slot05/batch26/skip_pattern_escape.a:7` | retained candidate |
| `rules/react.isEs5ComponentCallStrict` | `stage1/cohere/lint/helpers/slot05/batch18/react_es5_component_call_strict.a:1` | retained candidate |
| `rules/react.isReactComponentBaseName` | `stage1/cohere/lint/helpers/slot05/batch26/is_react_component_base_name.a:7` | retained candidate |
| `rules/structure.IsLikelyReactComponent` | `stage1/cohere/lint/helpers/slot05/batch33/is_likely_react_component.a:3` | retained candidate |
| `rules/structure.isJsxValue` | `stage1/cohere/lint/helpers/slot05/batch32/is_jsx_value.a:3` | retained candidate |
| `rules/structure.isNetworkServiceHookCall` | `stage1/cohere/lint/helpers/slot05/batch33/is_network_service_hook_call.a:3` | retained candidate |
| `rules/structure.jsxAfterParentheses` | `stage1/cohere/lint/helpers/slot05/batch32/jsx_after_parentheses.a:4` | retained candidate |
| `rules/structure.returnArgumentLooksLikeJsx` | `stage1/cohere/lint/helpers/slot05/batch32/return_argument_looks_like_jsx.a:5` | retained candidate |
| `rules/structure.typeReferenceName` | `stage1/cohere/lint/helpers/slot05/batch33/type_reference_name.a:3` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.calleeValues` | `stage1/cohere/lint/helpers/slot05/batch2/tailwind_callee_values.a:6` | retained candidate |
| `rules/tailwind.*ClassLiteralReader.variableValues` | `stage1/cohere/lint/helpers/slot05/batch2/tailwind_variable_values.a:6` | retained candidate |
| `rules/tailwind.trimDelimiters` | `stage1/cohere/lint/helpers/slot05/batch27/trim_delimiters.a:3` | retained candidate |
| `rules/tailwind/collapse.*Table.addRepositoryFunctionalRoots` | `stage1/cohere/lint/helpers/slot05/batch7/table_add_repository_functional_roots.a:6` | retained candidate |
| `rules/tailwind/collapse.*Table.addThemeNamespaces` | `stage1/cohere/lint/helpers/slot05/batch7/table_add_theme_namespaces.a:4` | retained candidate |
| `rules/tailwind/collapse.*Theme.Entries` | `stage1/cohere/lint/helpers/slot05/batch4/theme_entries.a:9` | retained candidate |
| `rules/tailwind/collapse.*Theme.KeysInNamespaces` | `stage1/cohere/lint/helpers/slot05/batch4/theme_keys_in_namespaces.a:6` | retained candidate |
| `rules/tailwind/collapse.*Theme.ResolveValue` | `stage1/cohere/lint/helpers/slot05/batch6/theme_resolve_value.a:9` | retained candidate |
| `rules/tailwind/collapse.*Theme.clearAll` | `stage1/cohere/lint/helpers/slot05/batch5/theme_clear_all.a:9` | retained candidate |
| `rules/tailwind/collapse.*Theme.compactKeyOrder` | `stage1/cohere/lint/helpers/slot05/batch5/theme_compact_key_order.a:3` | retained candidate |
| `rules/tailwind/collapse.*Theme.delete` | `stage1/cohere/lint/helpers/slot05/batch5/theme_delete.a:4` | retained candidate |
| `rules/tailwind/collapse.*Theme.liveKeys` | `stage1/cohere/lint/helpers/slot05/batch4/theme_live_keys.a:12` | retained candidate |
| `rules/tailwind/collapse.*Theme.resolveKey` | `stage1/cohere/lint/helpers/slot05/batch6/theme_resolve_key.a:7` | retained candidate |
| `rules/tailwind/collapse.FrameworkFunctionalUtility.Description` | `stage1/cohere/lint/helpers/slot05/batch29/functional_description.a:2` | retained candidate |
| `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.Description` | `stage1/cohere/lint/helpers/slot05/batch29/multi_description.a:2` | retained candidate |
| `rules/tailwind/collapse.LoadDesignSystem` | `stage1/cohere/lint/helpers/slot05/batch8/load_design_system.a:22` | retained candidate |
| `rules/tailwind/collapse.NewTheme` | `stage1/cohere/lint/helpers/slot05/batch6/new_theme.a:4` | retained candidate |
| `rules/tailwind/collapse.NewUtilityEvaluator` | `stage1/cohere/lint/helpers/slot05/batch7/new_utility_evaluator.a:6` | retained candidate |
| `rules/tailwind/collapse.ParseCandidate` | `stage1/cohere/lint/helpers/slot05/batch8/parse_candidate.a:3` | retained candidate |
| `rules/tailwind/collapse.ParseVariant` | `stage1/cohere/lint/helpers/slot05/batch8/parse_variant.a:3` | retained candidate |
| `rules/tailwind/collapse.bareValuePredicate` | `stage1/cohere/lint/helpers/slot05/batch28/bare_value_predicate.a:11` | retained candidate |
| `rules/tailwind/collapse.bareValueTransform` | `stage1/cohere/lint/helpers/slot05/batch28/bare_value_transform.a:1` | retained candidate |
| `rules/tailwind/collapse.borderSideDescription` | `stage1/cohere/lint/helpers/slot05/batch31/border_side_description.a:4` | retained candidate |
| `rules/tailwind/collapse.colorArm` | `stage1/cohere/lint/helpers/slot05/batch30/color_arm.a:2` | retained candidate |
| `rules/tailwind/collapse.isFontStretchPercentage` | `stage1/cohere/lint/helpers/slot05/batch29/font_stretch_percentage.a:3` | retained candidate |
| `rules/tailwind/collapse.isPositiveInteger` | `stage1/cohere/lint/helpers/slot05/batch27/is_positive_integer.a:13` | retained candidate |
| `rules/tailwind/collapse.isStrictPositiveInteger` | `stage1/cohere/lint/helpers/slot05/batch27/is_strict_positive_integer.a:2` | retained candidate |
| `rules/tailwind/collapse.isValidOpacityValue` | `stage1/cohere/lint/helpers/slot05/batch28/is_valid_opacity_value.a:2` | retained candidate |
| `rules/tailwind/collapse.maskStopDescription` | `stage1/cohere/lint/helpers/slot05/batch31/mask_stop_description.a:3` | retained candidate |
| `rules/tailwind/collapse.resolveArmColor` | `stage1/cohere/lint/helpers/slot05/batch31/resolve_arm_color.a:3` | retained candidate |
| `rules/tailwind/collapse.themeArm` | `stage1/cohere/lint/helpers/slot05/batch30/theme_arm.a:2` | retained candidate |
| `rules/tailwind/collapse.widthArm` | `stage1/cohere/lint/helpers/slot05/batch30/width_arm.a:2` | retained candidate |
| `ecmascript/regexp.Compile` | `stage1/cohere/lint/helpers/slot05/batch34/gap_test.go:17` TestDynamicCompilationBoundary | BLOCKED; no helper body |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot05/batch10/helper_test.go:23`: TestSlot05ControlEscape @23, TestSlot05NamedBackreference @27, TestSlot05NamedGroupOpener @30.
- `stage1/cohere/lint/helpers/slot05/batch11/helper_test.go:23`: TestSlot05ClassAtomWrite @23, TestSlot05WordCharacters @26, TestSlot05ExpandsOnUppercase @29.
- `stage1/cohere/lint/helpers/slot05/batch12/helper_test.go:23`: TestSlot05CaseTables @23, TestSlot05CaseEquivalents @26, TestSlot05CaseEquivalenceGroups @29.
- `stage1/cohere/lint/helpers/slot05/batch13/helper_test.go:23`: TestSlot05Canonicalize @23, TestSlot05CaseClass @26, TestSlot05WriteClass @30.
- `stage1/cohere/lint/helpers/slot05/batch14/helper_test.go:23`: TestSlot05MakeBreak @23.
- `stage1/cohere/lint/helpers/slot05/batch15/helper_test.go:23`: TestSlot05Statements @23, TestSlot05StatementOrder @26, TestSlot05MakeContinue @29.
- `stage1/cohere/lint/helpers/slot05/batch16/helper_test.go:23`: TestSlot05UpdateExpression @23, TestSlot05Parameter @26, TestSlot05FirstThrowableFork @29.
- `stage1/cohere/lint/helpers/slot05/batch17/helper_test.go:23`: TestSlot05VisitUnknown @23, TestSlot05ForkOptionalChain @26, TestSlot05IsHookCall @29.
- `stage1/cohere/lint/helpers/slot05/batch18/helper_test.go:23`: TestSlot05PatternReads @23, TestSlot05PatternWrites @26, TestSlot05Es5Strict @29.
- `stage1/cohere/lint/helpers/slot05/batch19/helper_test.go:23`: TestSlot05WhileStatement @23, TestSlot05DoStatement @26, TestSlot05LabeledStatement @29.
- `stage1/cohere/lint/helpers/slot05/batch2/helper_test.go:21`: TestSlot05CalleeValues @21, TestSlot05CalleeArguments @24, TestSlot05VariableValues @27.
- `stage1/cohere/lint/helpers/slot05/batch20/helper_test.go:23`: TestSlot05ForStatement @23, TestSlot05ForInOfStatement @26, TestSlot05AccessOrCall @29.
- `stage1/cohere/lint/helpers/slot05/batch21/helper_test.go:21`: TestJoinRanges @21, TestCaseExtras @24, TestBuildCaseTables @27.
- `stage1/cohere/lint/helpers/slot05/batch22/helper_test.go:21`: TestProperty @21, TestNumeric @24, TestReadClass @27.
- `stage1/cohere/lint/helpers/slot05/batch23/helper_test.go:21`: TestDecodeEscape @21, TestClassAtoms @24, TestRegExpTest @27.
- `stage1/cohere/lint/helpers/slot05/batch24/helper_test.go:21`: TestRewrite @21, TestIsHexDigit @38, TestAllHexDigits @41.
- `stage1/cohere/lint/helpers/slot05/batch25/helper_test.go:21`: TestParseRegexFlags @21, TestRegexFlagsUV @24, TestPatternAndFlags @27.
- `stage1/cohere/lint/helpers/slot05/batch26/helper_test.go:21`: TestSkipPatternEscape @21, TestClassEnd @24, TestReactComponentBaseName @27.
- `stage1/cohere/lint/helpers/slot05/batch27/helper_test.go:21`: TestPositiveInteger @21, TestStrictPositiveInteger @24, TestTrimDelimiters @27.
- `stage1/cohere/lint/helpers/slot05/batch28/helper_test.go:21`: TestBareValuePredicate @21, TestBareValueTransform @33, TestValidOpacityValue @39.
- `stage1/cohere/lint/helpers/slot05/batch29/helper_test.go:21`: TestFunctionalDescription @21, TestMultiDescription @39, TestFontStretchPercentage @54.
- `stage1/cohere/lint/helpers/slot05/batch3/helper_test.go:20`: TestSlot05MatchExactly @20, TestSlot05NormalizedFileName @23, TestSlot05IsHookName @26.
- `stage1/cohere/lint/helpers/slot05/batch30/helper_test.go:21`: TestColorArm @21, TestThemeArm @32, TestWidthArm @43.
- `stage1/cohere/lint/helpers/slot05/batch31/helper_test.go:21`: TestBorder @21, TestMask @38, TestResolve @60.
- `stage1/cohere/lint/helpers/slot05/batch32/helper_test.go:21`: TestValue @21, TestAfter @28, TestReturns @35.
- `stage1/cohere/lint/helpers/slot05/batch33/helper_test.go:21`: TestComponent @21, TestType @34, TestNetwork @41.
- `stage1/cohere/lint/helpers/slot05/batch34/gap_test.go:17`: TestDynamicCompilationBoundary @17.
- `stage1/cohere/lint/helpers/slot05/batch4/helper_test.go:21`: TestSlot05LiveKeys @21, TestSlot05ThemeEntries @24, TestSlot05KeysInNamespaces @27.
- `stage1/cohere/lint/helpers/slot05/batch5/helper_test.go:23`: TestSlot05ClearAll @23, TestSlot05CompactKeyOrder @29, TestSlot05DeleteThemeKey @36.
- `stage1/cohere/lint/helpers/slot05/batch6/helper_test.go:23`: TestSlot05NewTheme @23, TestSlot05ResolveThemeKey @28, TestSlot05ResolveThemeValue @37.
- `stage1/cohere/lint/helpers/slot05/batch7/helper_test.go:23`: TestSlot05FunctionalRoots @23, TestSlot05ThemeNamespaces @26, TestSlot05UtilityEvaluator @29.
- `stage1/cohere/lint/helpers/slot05/batch8/helper_test.go:23`: TestSlot05Candidate @23, TestSlot05Variant @26, TestSlot05Loader @29.
- `stage1/cohere/lint/helpers/slot05/batch9/helper_test.go:23`: TestSlot05ImportedNameOf @23, TestSlot05DecimalDigit @27, TestSlot05CountGroups @30.
- `stage1/cohere/lint/helpers/slot05_test.go:35`: TestSlot05Identifier @35, TestSlot05Imports @38, TestSlot05ComponentBase @41.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/LANDING_REPORT_2.md:1`, `stage1/cohere/lint/helpers/slot05/LANDING_REPORT_3.md:1`, `stage1/cohere/lint/helpers/slot05/LANDING_REPORT_4.md:1`, `stage1/cohere/lint/helpers/slot05/LANDING_REPORT_5.md:1`, `stage1/cohere/lint/helpers/slot05/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch10/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch11/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch12/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch13/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch14/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch15/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch16/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch17/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch18/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch19/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch2/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch20/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch21/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch22/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch23/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch24/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch25/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch26/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch27/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch28/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch29/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch3/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch30/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch31/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch32/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch33/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch34/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch34/recheck_78dca9077/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch4/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch5/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch6/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch7/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch8/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/batch9/REPORT.md:1`, `stage1/cohere/lint/helpers/slot05/landing_7945d102/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/slot05/batch2/evidence/withdrawn-dispatcher-experiment.log`, `stage1/cohere/lint/helpers/slot05/batch2/evidence/withdrawn-factory-attempt.log`, `stage1/cohere/lint/helpers/slot05/batch4/evidence/withdrawn-escape-experiment.log`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-attribute-corpus.sha256`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-attribute-coverage.json`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-comparisons.log`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-extra-mutants.log`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-source-corpus.sha256`, `stage1/cohere/lint/helpers/slot05/batch9/evidence/withdrawn-source-coverage.json`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 1119 files; +441,273/-5 lines; 46,501,447 blob bytes.
- test/oracle/capture programs: 149 files; +20,891/-1 lines; 797,082 blob bytes.
- implementation/support types: 115 files; +1,887/-0 lines; 101,621 blob bytes.

### origin/codex/lint-helpers-from-codex-lint-wave1-03

Tip `5e64fe3cf0d236a2fae7e4f8b645f0d3f33afc29`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Two retained helpers (Theme.Add, FrameworkStaticReading) plus BLOCKED appendSuccessor recursive graph source. validate.py and validate-static.py instrument actual calls from six consumer suites and controls; separate repository-corpus tests are unavailable. Semantic mutants are byte-compared on source Node/emitted JS/native. appendSuccessor stops at cycle-capable lowering on a real self-edge, with source-only output mutant; no four-consumer native readiness.

Claim: `stage1/cohere/lint/helpers/claims/codex-lint-wave1-03.md:1` (68 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].appendSuccessor` | `stage1/cohere/lint/helpers/slot_wave1_03/gaps/successor-mutant.a:6` | BLOCKED source/probe only |
| `rules/tailwind/collapse.*Theme.Add` | `stage1/cohere/lint/helpers/slot_wave1_03/theme_add.a:19` | retained candidate |
| `rules/tailwind/collapse.FrameworkStaticReading` | `stage1/cohere/lint/helpers/slot_wave1_03/framework_static_reading.a:28` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot_wave1_03/helpers_test.go:13`: TestThemeAddMatchesCohere @13, TestFrameworkStaticReadingMatchesCohere @38.
- `stage1/cohere/lint/helpers/slot_wave1_03/validate-static.py:15`: comparison/mutant assertion @15, comparison/mutant assertion @38, comparison/mutant assertion @48, comparison/mutant assertion @54.
- `stage1/cohere/lint/helpers/slot_wave1_03/validate.py:15`: comparison/mutant assertion @15, comparison/mutant assertion @40, comparison/mutant assertion @53, comparison/mutant assertion @63.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot_wave1_03/REPORT.md:1`, `stage1/cohere/lint/helpers/slot_wave1_03/STATIC_REPORT.md:1`, `stage1/cohere/lint/helpers/slot_wave1_03/SUCCESSOR_REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 254 files; +11,736/-0 lines; 2,947,773 blob bytes.
- implementation/support types: 2 files; +71/-0 lines; 3,194 blob bytes.
- test/oracle/capture programs: 5 files; +314/-0 lines; 21,278 blob bytes.

### origin/codex/lint-helpers-from-codex-lint-wave1-14

Tip `7b6a7c28bdf9b9e715a8403cf27703a69afeeb74`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Two retained regex decoders, decodeFixedHex and decodeUnicodeEscape, with actual private-Go four-consumer targeted paths, controls and compiling three-backend semantic mutants; HEX_REPORT.md explicitly records zero leaf invocations in every original suite, followed by targeted real-rule probes in validate_hex.py/validate_unicode.py. leadingInteger and nodesFromStaticDeclarations duplicates are removed from active delivery and archived as .a.txt; their old six-consumer attempts recorded zero live calls and missing Tailwind inputs. Do not credit that history.

Claim: `stage1/cohere/lint/helpers/claims/codex-lint-wave1-14.md:1` (86 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/regexp.decodeFixedHex` | `stage1/cohere/lint/helpers/slot14/regexp_decode_fixed_hex.a:50` | retained candidate |
| `ecmascript/regexp.decodeUnicodeEscape` | `stage1/cohere/lint/helpers/slot14/regexp_decode_unicode_escape.a:18` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/slot14/validate.py:55`: comparison/mutant assertion @55, comparison/mutant assertion @63.
- `stage1/cohere/lint/helpers/slot14/validate_hex.py:121`: comparison/mutant assertion @121.
- `stage1/cohere/lint/helpers/slot14/validate_nodes.py:32`: comparison/mutant assertion @32, comparison/mutant assertion @38.
- `stage1/cohere/lint/helpers/slot14/validate_unicode.py:122`: comparison/mutant assertion @122.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/AREA_LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/HEX_REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/RETAINED_REPORT.md:1`, `stage1/cohere/lint/helpers/slot14/UNICODE_REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/slot14/withdrawn_leading_integer.a.txt`, `stage1/cohere/lint/helpers/slot14/withdrawn_nodes_from_static_declarations.a.txt`, `stage1/cohere/lint/helpers/slot14/withdrawn_nodes_runner.a.txt`, `stage1/cohere/lint/helpers/slot14/withdrawn_runner.a.txt`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 92 files; +3,528/-0 lines; 348,776 blob bytes.
- test/oracle/capture programs: 6 files; +404/-0 lines; 29,546 blob bytes.
- implementation/support types: 2 files; +124/-0 lines; 5,837 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-02

Tip `d62bb9d4df13cc6132f0b4bc2b1748d16934edb7`; merge-base with area `b6b1538b0cebc4ba6741ac34f1aedb60293c1d06`; merge-tree conflicted (exit 1), C1 eight files above.

**Coverage verdict:** Four retained CFG helpers: markFinal/markThrown/snapshotForks/restoreForks. Scripts instrument all four original consumer suites, then add live rule probes where originals make zero calls; snapshot original no-unreachable-loop has zero calls, its probe has eight. Alias snapshots and shared-frame ordering have explicit controls/mutants; all three backends compare bytes. Whole-tip merge conflicts C1. Stable arena and full graph integration remain external. enter withdrawn.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-02.md:1` (63 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].markFinal` | `stage1/cohere/lint/helpers/from_wave1_02/final/mark_final.a:23` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].markThrown` | `stage1/cohere/lint/helpers/from_wave1_02/thrown/mark_thrown.a:25` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].restoreForks` | `stage1/cohere/lint/helpers/from_wave1_02/forks/restore_forks.a:5` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].snapshotForks` | `stage1/cohere/lint/helpers/from_wave1_02/forks/snapshot_forks.a:9` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave1_02/final/verify.py:47`: comparison/mutant assertion @47, comparison/mutant assertion @53, comparison/mutant assertion @55.
- `stage1/cohere/lint/helpers/from_wave1_02/forks/verify_restore.py:42`: comparison/mutant assertion @42, comparison/mutant assertion @43, comparison/mutant assertion @52, comparison/mutant assertion @53, comparison/mutant assertion @55.
- `stage1/cohere/lint/helpers/from_wave1_02/forks/verify_snapshot.py:41`: comparison/mutant assertion @41, comparison/mutant assertion @42, comparison/mutant assertion @50, comparison/mutant assertion @51, comparison/mutant assertion @53.
- `stage1/cohere/lint/helpers/from_wave1_02/thrown/verify.py:49`: comparison/mutant assertion @49, comparison/mutant assertion @55, comparison/mutant assertion @57.
- `stage1/cohere/lint/helpers/from_wave1_02/verify_joint.py:18`: comparison/mutant assertion @18, comparison/mutant assertion @22.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/DELIVERY.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/final/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/forks/DELIVERY.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/forks/RESTORE_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/forks/SNAPSHOT_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_02/thrown/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/from_wave1_02/evidence/withdrawn-enter/coverage.json`, `stage1/cohere/lint/helpers/from_wave1_02/evidence/withdrawn-enter/verification.log`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 94 files; +39,894/-0 lines; 3,122,189 blob bytes.
- test/oracle/capture programs: 10 files; +410/-0 lines; 30,186 blob bytes.
- implementation/support types: 6 files; +141/-0 lines; 5,535 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-04

Tip `8754bdeafe0ef400be14bfb1d8f1505c0c80c713`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Seven helpers, actual Go over original fixture domains and extensive scanner/suffix/vector/fraction controls, plus clean-executing mutants. ProjectRoot and loadDesignSystemForProgram use supplied Program/filesystem/load behavior: this is not the synchronized DesignSystemForProgram cache. All-rule end-to-end findings/fixes are not tested by these isolated contracts.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-04.md:1` (146 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind.loadDesignSystemForProgram` | `stage1/cohere/lint/helpers/from_wave1_04/load_design_system_for_program.a:13` | retained candidate |
| `rules/tailwind.projectRootOf` | `stage1/cohere/lint/helpers/from_wave1_04/project_root.a:24` | retained candidate |
| `rules/tailwind/collapse.isFraction` | `stage1/cohere/lint/helpers/from_wave1_04/is_fraction.a:3` | retained candidate |
| `rules/tailwind/collapse.isVector` | `stage1/cohere/lint/helpers/from_wave1_04/is_vector.a:3` | retained candidate |
| `rules/tailwind/collapse.numberWithSuffix` | `stage1/cohere/lint/helpers/from_wave1_04/number_with_suffix.a:2` | retained candidate |
| `rules/tailwind/collapse.scanNumber` | `stage1/cohere/lint/helpers/from_wave1_04/scan_number.a:3` | retained candidate |
| `rules/tailwind/collapse.trimLeadingJavaScriptSpace` | `stage1/cohere/lint/helpers/from_wave1_04/trim_leading_javascript_space.a:3` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave1_04/fraction_test.go:44`: TestFractionMatchesGo @44, TestFractionMutants @58.
- `stage1/cohere/lint/helpers/from_wave1_04/helpers_test.go:205`: TestProjectRootMatchesGo @205, TestProjectRootMutants @218.
- `stage1/cohere/lint/helpers/from_wave1_04/load_test.go:150`: TestLoadDesignSystemForProgramMatchesGo @150, TestLoadDesignSystemForProgramMutants @162.
- `stage1/cohere/lint/helpers/from_wave1_04/scan_test.go:47`: TestScanNumberMatchesGo @47, TestScanNumberMutants @61.
- `stage1/cohere/lint/helpers/from_wave1_04/suffix_test.go:44`: TestNumberWithSuffixMatchesGo @44, TestNumberWithSuffixMutants @58.
- `stage1/cohere/lint/helpers/from_wave1_04/trim_test.go:103`: TestTrimLeadingJavaScriptSpaceMatchesGo @103, TestTrimLeadingJavaScriptSpaceMutants @117.
- `stage1/cohere/lint/helpers/from_wave1_04/vector_test.go:44`: TestVectorMatchesGo @44, TestVectorMutants @58.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/FRACTION_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/SCAN_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/SUFFIX_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/TRIM_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_04/VECTOR_REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 72 files; +7,802/-0 lines; 728,172 blob bytes.
- test/oracle/capture programs: 16 files; +1,136/-0 lines; 48,691 blob bytes.
- implementation/support types: 7 files; +128/-0 lines; 6,215 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-06

Tip `6f2aff526107895a267896fdcbf182cd67c3c7a6`; merge-base with area `f8013f0baac41ddc340d76f83bddde38536a8f07`; merge-tree conflicted (exit 1), C1 eight files above.

**Coverage verdict:** Two retained helpers, Table.addRepositoryStatics and attachFrameworkVariantComparisons; actual Go table/comparison facts on six fixture families, plus controls and eight semantic mutants. Comparison functions and table allocation are explicit dependencies. Whole-tip conflicts C1. NewTheme and FrameworkStaticReading withdrawn and their prior isolated green observations are historical only.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-06.md:1` (36 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind/collapse.*Table.addRepositoryStatics` | `stage1/cohere/lint/helpers/from-wave1-06/repository_statics.a:5` | retained candidate |
| `rules/tailwind/collapse.attachFrameworkVariantComparisons` | `stage1/cohere/lint/helpers/from-wave1-06/attach_variant_comparisons.a:2` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from-wave1-06/attach_test.go:66`: TestAttachComparisons @66.
- `stage1/cohere/lint/helpers/from-wave1-06/repository_test.go:155`: TestRepositoryStatics @155.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-06/DELIVERY.md:1`, `stage1/cohere/lint/helpers/from-wave1-06/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-06/REPOSITORY_REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-06/STATIC_REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 28 files; +1,842/-0 lines; 171,450 blob bytes.
- test/oracle/capture programs: 14 files; +925/-0 lines; 35,005 blob bytes.
- implementation/support types: 2 files; +27/-0 lines; 1,199 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-09

Tip `017f34d7fc64fed98740ab5456f516c8cec2931e`; merge-base with area `b6b1538b0cebc4ba6741ac34f1aedb60293c1d06`; merge-tree clean (exit 0).

**Coverage verdict:** Three bounded candidates, not consumer-ready: nodesFromStaticDeclarations/propertySort/ParseValue. Readiness explicitly says confirmed dependency removals 0 and every original consumer recorded zero actual calls because Tailwind/Kirk inputs are absent. Nineteen semantic mutants are caught on three backends in bounded Go comparisons. TestStylesheetLoaderByteInputGap proves readTextFile loses invalid UTF-8; loader withdrawn. Current test files also have old 715ba94f pin guards and need actual oracle migration, not a hash-only edit.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-09.md:1` (53 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind/collapse.ParseValue` | `stage1/cohere/lint/helpers/from_wave1_09/parse_value.a:14` | retained candidate |
| `rules/tailwind/collapse.nodesFromStaticDeclarations` | `stage1/cohere/lint/helpers/from_wave1_09/nodes_from_static_declarations.a:20` | retained candidate |
| `rules/tailwind/collapse.propertySort` | `stage1/cohere/lint/helpers/from_wave1_09/property_sort.a:16` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave1_09/file_gap_test.go:12`: TestStylesheetLoaderByteInputGap @12.
- `stage1/cohere/lint/helpers/from_wave1_09/helper_test.go:22`: TestStaticDeclarationNodesAgree @22.
- `stage1/cohere/lint/helpers/from_wave1_09/sort_test.go:14`: TestPrivatePropertySortAgrees @14.
- `stage1/cohere/lint/helpers/from_wave1_09/value_test.go:14`: TestValueParserAgrees @14.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_09/LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_09/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_09/file_loader_gap/BLOCKER.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 105 files; +7,414/-0 lines; 652,005 blob bytes.
- test/oracle/capture programs: 13 files; +1,258/-0 lines; 47,390 blob bytes.
- implementation/support types: 3 files; +151/-0 lines; 6,022 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-10

Tip `48c2db8a342ee27a2630f07d0b0923f4a5da0bca`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Four retained helpers. Math spacing/CSS unescape and modifierGroup/simpleFold each have real Go queries, four consumer fixture families/boundaries, and explicit compiling semantic mutant tests. Byte/rune/valid-Unicode interfaces are bounded contracts; helper parity is not full rule implementation.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-10.md:1` (39 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/regexp.modifierGroup` | `stage1/cohere/lint/helpers/from-wave1-10/regexp_modifier_group.a:10` | retained candidate |
| `ecmascript/regexp.simpleFold` | `stage1/cohere/lint/helpers/from-wave1-10/regexp_simple_fold.a:268` | retained candidate |
| `rules/tailwind/collapse.addWhitespaceAroundMathOperators` | `stage1/cohere/lint/helpers/from-wave1-10/math_operator_whitespace.a:3` | retained candidate |
| `rules/tailwind/collapse.unescapeCSSIdentifier` | `stage1/cohere/lint/helpers/from-wave1-10/css_identifier_unescape.a:2` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from-wave1-10/fold_test.go:98`: TestSimpleFold @98, TestSimpleFoldMutant @109.
- `stage1/cohere/lint/helpers/from-wave1-10/helper_test.go:149`: TestMathWhitespace @149, TestMathWhitespaceMutant @163.
- `stage1/cohere/lint/helpers/from-wave1-10/modifier_test.go:97`: TestModifierGroup @97, TestModifierGroupMutant @108.
- `stage1/cohere/lint/helpers/from-wave1-10/unescape_test.go:53`: TestCSSIdentifierUnescape @53, TestCSSIdentifierUnescapeMutant @67.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-10/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 74 files; +4,109/-0 lines; 3,990,019 blob bytes.
- implementation/support types: 4 files; +397/-0 lines; 39,855 blob bytes.
- test/oracle/capture programs: 12 files; +799/-0 lines; 31,724 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-11

Tip `a6278560ea1d50cbbd9147b048a0c86b7bdf4478`; merge-base with area `e667e3e1dbdfd1b9125c3256961bfbc8ec31946b`; merge-tree clean (exit 0).

**Coverage verdict:** Seven helpers. TestFindEntryPointMatchesGoWithMutant and sibling tests invoke owned validate scripts with actual private-Go overlays, consumer fixture inputs, source Node/emitted JS/native and output-only mutants. Filesystem existence, optional node projection and CSS/value arena dependencies are caller-owned. It includes parked wave rule history outside helper scope; extract only the package slice.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-11.md:1` (44 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.IsRoot` | `stage1/cohere/lint/helpers/from_wave1_11/cfg_is_root.a:2` | retained candidate |
| `ecmascript/control_flow_graph.isDestructuringTarget` | `stage1/cohere/lint/helpers/from_wave1_11/cfg_is_destructuring_target.a:4` | retained candidate |
| `rules/react.skipParenthesesOptional` | `stage1/cohere/lint/helpers/from_wave1_11/react_skip_parentheses_optional.a:3` | retained candidate |
| `rules/structure.parameterTypeNode` | `stage1/cohere/lint/helpers/from_wave1_11/structure_parameter_type_node.a:3` | retained candidate |
| `rules/tailwind.FindEntryPoint` | `stage1/cohere/lint/helpers/from_wave1_11/tailwind_find_entry_point.a:15` | retained candidate |
| `rules/tailwind/collapse.ValueToCss` | `stage1/cohere/lint/helpers/from_wave1_11/tailwind_value_to_css.a:9` | retained candidate |
| `rules/tailwind/collapse.escapeCSSIdentifier` | `stage1/cohere/lint/helpers/from_wave1_11/tailwind_escape_css_identifier.a:10` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave1_11/entry_test.go:10`: TestFindEntryPointMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/escape_test.go:10`: TestEscapeCssIdentifierMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/optional_test.go:10`: TestReactOptionalParenthesesMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/parameter_test.go:10`: TestParameterTypeNodeMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/pattern_test.go:10`: TestCfgIsDestructuringTargetMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/root_test.go:10`: TestCfgIsRootMatchesGoWithMutant @10.
- `stage1/cohere/lint/helpers/from_wave1_11/validate.py:43`: comparison/mutant assertion @43.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_escape.py:50`: comparison/mutant assertion @50.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_optional.py:39`: comparison/mutant assertion @39.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_parameter.py:38`: comparison/mutant assertion @38.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_pattern.py:38`: comparison/mutant assertion @38.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_root.py:36`: comparison/mutant assertion @36.
- `stage1/cohere/lint/helpers/from_wave1_11/validate_value.py:52`: comparison/mutant assertion @52.
- `stage1/cohere/lint/helpers/from_wave1_11/value_test.go:10`: TestValueToCssMatchesGoWithMutant @10.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/AREA_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/ESCAPE_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/NEXT_BLOCKER.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/OPTIONAL_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/PARAMETER_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/PATTERN_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_11/ROOT_REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 217 files; +79,287/-0 lines; 5,391,573 blob bytes.
- test/oracle/capture programs: 16 files; +613/-0 lines; 45,527 blob bytes.
- implementation/support types: 13 files; +188/-0 lines; 9,153 blob bytes.

### origin/codex/lint-helpers-from-codex/lint-wave1-15

Tip `298ecc4cb20069615bb68130f177bd6f97dd7869`; merge-base with area `e667e3e1dbdfd1b9125c3256961bfbc8ec31946b`; merge-tree clean (exit 0).

**Coverage verdict:** Six retained helpers: CSS string/declaration, regexp.parseFlags, SpecifierNode/SourceVisitors/AccessedName. validate scripts check actual Go, all selected consumer fixture families, three output backends and semantic mutants. Fresh sources for import/property are explicit projected-node interfaces. VariantKind withdrawn. DesignSystemForProgram probe fails missing Mutex TS2305; no serial/path-key cache delivered.

Claim: `stage1/cohere/lint/helpers/claims/codex/lint-wave1-15.md:1` (117 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/imports.SourceVisitors` | `stage1/cohere/lint/helpers/wave15/import_source_visitors/source_visitors.a:19` | retained candidate |
| `ecmascript/imports.SpecifierNode` | `stage1/cohere/lint/helpers/wave15/import_specifier/specifier_node.a:6` | retained candidate |
| `ecmascript/property.AccessedName` | `stage1/cohere/lint/helpers/wave15/accessed_property/accessed_name.a:18` | retained candidate |
| `ecmascript/regexp.parseFlags` | `stage1/cohere/lint/helpers/wave15/regexp_flags/parse_flags.a:44` | retained candidate |
| `rules/tailwind/collapse.parseCSSDeclaration` | `stage1/cohere/lint/helpers/wave15/css_declaration/parse_css_declaration.a:48` | retained candidate |
| `rules/tailwind/collapse.parseCSSString` | `stage1/cohere/lint/helpers/wave15/css_string/parse_css_string.a:12` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/wave15/accessed_property/validate.py:32`: comparison/mutant assertion @32, comparison/mutant assertion @38.
- `stage1/cohere/lint/helpers/wave15/css_declaration/validate.py:22`: comparison/mutant assertion @22, comparison/mutant assertion @28.
- `stage1/cohere/lint/helpers/wave15/css_string/validate.py:22`: comparison/mutant assertion @22, comparison/mutant assertion @28.
- `stage1/cohere/lint/helpers/wave15/design_system_cache/validate_blocker.py:13`: comparison/mutant assertion @13.
- `stage1/cohere/lint/helpers/wave15/import_source_visitors/validate.py:36`: comparison/mutant assertion @36, comparison/mutant assertion @42.
- `stage1/cohere/lint/helpers/wave15/import_specifier/validate.py:42`: comparison/mutant assertion @42, comparison/mutant assertion @48.
- `stage1/cohere/lint/helpers/wave15/regexp_flags/validate.py:25`: comparison/mutant assertion @25, comparison/mutant assertion @33.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/accessed_property/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/css_declaration/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/css_string/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/design_system_cache/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/import_source_visitors/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/import_specifier/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-7481e032/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-b28757f3/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-b4691483/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-b84a9d93/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-d3a37422/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-d65a8f93/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-area-e667e3e1/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-b8fb957a/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-c01907a7/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-current/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-evidence/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/landing-f8013f0b/REPORT.md:1`, `stage1/cohere/lint/helpers/wave15/regexp_flags/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 205 files; +174,619/-0 lines; 11,446,405 blob bytes.
- implementation/support types: 7 files; +996/-0 lines; 24,916 blob bytes.
- test/oracle/capture programs: 13 files; +412/-0 lines; 33,790 blob bytes.

### origin/codex/lint-helpers-from-lint-wave1-01

Tip `72fca0ebb2abb157f395c2cf91520c04ea26d794`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Eight CFG helpers on all four upstream consumer suites plus independent/live supplementary controls, source Node/emitted JS/native and 17 comparison-only mutants. Enter, frame and statement/type-argument handoffs assume supplied fact graphs and expression visitor; full AST/graph integration is not covered. Complete rule blocker sets remain open.

Claim: `stage1/cohere/lint/helpers/claims/lint-wave1-01.md:1` (80 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].enter` | `stage1/cohere/lint/helpers/from_wave1_01/enter.a:16` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].enterDisconnected` | `stage1/cohere/lint/helpers/from_wave1_01/enter_disconnected.a:6` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].reachedStatement` | `stage1/cohere/lint/helpers/from_wave1_01/reached_statement.a:16` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].returnFrame` | `stage1/cohere/lint/helpers/from_wave1_01/return_frame.a:2` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].throwFrame` | `stage1/cohere/lint/helpers/from_wave1_01/throw_frame.a:10` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].typeArguments` | `stage1/cohere/lint/helpers/from_wave1_01/type_arguments.a:18` | retained candidate |
| `ecmascript/control_flow_graph.isThrowableIdentifier` | `stage1/cohere/lint/helpers/from_wave1_01/is_throwable_identifier.a:20` | retained candidate |
| `ecmascript/control_flow_graph.throwTarget` | `stage1/cohere/lint/helpers/from_wave1_01/throw_target.a:8` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave1_01/enter_disconnected_test.go:12`: TestDisconnectedEntryMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/enter_test.go:67`: TestEnterMatchesGo @67.
- `stage1/cohere/lint/helpers/from_wave1_01/is_throwable_identifier_test.go:12`: TestThrowableIdentifierMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/reached_statement_test.go:12`: TestStatementHandoffMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/return_frame_test.go:12`: TestReturnFrameMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture.py:33`: TestAdamicEnterControls @33.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_disconnected.py:33`: TestAdamicDisconnectedControls @33, TestArrayCallbackReturnAdamicDisconnected @51, TestConsistentReturnAdamicDisconnected @56, TestNoUnreachableLoopAdamicDisconnected @61, TestRulesOfHooksAdamicDisconnected @68.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_return.py:29`: TestAdamicReturnControls @29.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_statement.py:51`: TestAdamicStatementControls @51.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_target.py:26`: TestAdamicTargetControls @26.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_throw.py:29`: TestAdamicThrowControls @29.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_throwable.py:38`: TestAdamicThrowableControls @38.
- `stage1/cohere/lint/helpers/from_wave1_01/testdata/capture_type_arguments.py:53`: TestAdamicTypeArgumentsControls @53, TestArrayCallbackReturnAdamicTypeArguments @75, TestConsistentReturnAdamicTypeArguments @80, TestNoUnreachableLoopAdamicTypeArguments @85, TestRulesOfHooksAdamicTypeArguments @92.
- `stage1/cohere/lint/helpers/from_wave1_01/throw_frame_test.go:12`: TestThrowFrameMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/throw_target_test.go:12`: TestThrowTargetMatchesGo @12.
- `stage1/cohere/lint/helpers/from_wave1_01/type_arguments_test.go:12`: TestTypeArgumentsMatchesGo @12.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave1_01/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 54 files; +1,799/-0 lines; 173,380 blob bytes.
- test/oracle/capture programs: 24 files; +1,495/-0 lines; 68,246 blob bytes.
- implementation/support types: 8 files; +166/-0 lines; 6,296 blob bytes.

### origin/codex/lint-helpers-from-lint-wave1-05

Tip `376dd83f2c8c7cbeedc819c03631fdd670090c06`; merge-base with area `b6b1538b0cebc4ba6741ac34f1aedb60293c1d06`; merge-tree conflicted (exit 1), C1 eight files above.

**Coverage verdict:** Three retained helpers: leadingInteger/findTailwindPackageRoot/CompareBreakpoints. Owned validate scripts compare actual Go over six fixture families plus POSIX path, integer overflow and direction controls, with compiling semantic mutants on three backends. Whole-tip conflicts C1. DesignSystemForProgram retry still TS2305 missing Mutex and has no cache body/readiness; FindEntryPoint withdrawn.

Claim: `stage1/cohere/lint/helpers/claims/lint-wave1-05.md:1` (41 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind.findTailwindPackageRoot` | `stage1/cohere/lint/helpers/tailwind_find_package_root.a:16` | retained candidate |
| `rules/tailwind/collapse.CompareBreakpoints` | `stage1/cohere/lint/helpers/collapse_compare_breakpoints.a:17` | retained candidate |
| `rules/tailwind/collapse.leadingInteger` | `stage1/cohere/lint/helpers/collapse_leading_integer.a:28` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/wave05/compare_validate.py:52`: comparison/mutant assertion @52.
- `stage1/cohere/lint/helpers/wave05/package_validate.py:63`: comparison/mutant assertion @63.
- `stage1/cohere/lint/helpers/wave05/validate.py:53`: comparison/mutant assertion @53.
- `stage1/cohere/lint/helpers/wave05_test.go:13`: TestWave05LeadingInteger @13, TestWave05PackageRoot @34, TestWave05CompareBreakpoints @55.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/CACHE_BLOCKER.md:1`, `stage1/cohere/lint/helpers/wave05/COMPARE_REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/LEADING_REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/PACKAGE_REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/REPORT.md:1`, `stage1/cohere/lint/helpers/wave05/RULES_LANDING_BLOCKER.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/wave05/gaps/design_system_mutex.a.txt`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 83 files; +192,286/-0 lines; 5,527,761 blob bytes.
- implementation/support types: 3 files; +98/-0 lines; 4,507 blob bytes.
- test/oracle/capture programs: 7 files; +306/-0 lines; 23,664 blob bytes.

### origin/codex/lint-helpers-from-lint-wave1-08

Tip `bdae8bc99dfe54a06615b20e34084bc9b8ebb06e`; merge-base with area `71d7e491b3c9724f7a0e2ee754592149e7f9790b`; merge-tree clean (exit 0).

**Coverage verdict:** Four retained helpers. from_wave08/helpers_test.go consumes all six fixture literal families; helper and mutant comparisons include source Node/emitted JS/native. ClearNamespace, width resolver and variant comparator have callback dependencies; fixture literals are not every complete consuming rule run. NewTheme overlaps slot05 (choose slot05). NewVariantRegistry withdrawn.

Claim: `stage1/cohere/lint/helpers/claims/lint-wave1-08.md:1` (26 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind/collapse.*Theme.ClearNamespace` | `stage1/cohere/lint/helpers/from_wave08/theme_clear_namespace.a:4` | retained candidate |
| `rules/tailwind/collapse.NewTheme` | `stage1/cohere/lint/helpers/from_wave08/theme_new.a:10` | retained candidate |
| `rules/tailwind/collapse.compareBreakpointVariants` | `stage1/cohere/lint/helpers/from_wave08/breakpoint_compare.a:3` | retained candidate |
| `rules/tailwind/collapse.resolveBreakpointWidth` | `stage1/cohere/lint/helpers/from_wave08/breakpoint_width.a:10` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from_wave08/helpers_test.go:99`: TestNewTheme @99, TestClearNamespace @102, TestBreakpointWidth @194, TestBreakpointCompare @198.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave08/BREAKPOINTS_REPORT.md:1`, `stage1/cohere/lint/helpers/from_wave08/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 61 files; +2,379/-0 lines; 233,496 blob bytes.
- implementation/support types: 4 files; +60/-0 lines; 3,076 blob bytes.
- test/oracle/capture programs: 5 files; +321/-0 lines; 13,719 blob bytes.

### origin/codex/lint-helpers-from-lint-wave1-12

Tip `a4c72791598fb9494e02649b7eb380eab2a04779`; merge-base with area `e667e3e1dbdfd1b9125c3256961bfbc8ec31946b`; merge-tree clean (exit 0).

**Coverage verdict:** Nine delivered helpers plus BLOCKED exact nextBuildCount support probe. Actual private-Go observations and semantic mutants exist on all four CFG fixture families. Decorators/typeParameters original fixtures have zero direct calls: eight decorator controls exercise 19 calls, twelve type-parameter controls 32. Those are explicit controls, not silently counted as live original coverage. Normalization now uses shared JS RegExp literals. Exact signed-64 counter probe refuses bigint return lowering, with an intentionally wrong number-counter mutant; no readiness credit.

Claim: `stage1/cohere/lint/helpers/claims/lint-wave1-12.md:1` (56 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `ecmascript/control_flow_graph.*Builder[E].decorators` | `stage1/cohere/lint/helpers/wave12/control_flow_decorators.a:4` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].newBlock` | `stage1/cohere/lint/helpers/wave12/control_flow_new_block.a:16` | retained candidate |
| `ecmascript/control_flow_graph.*Builder[E].typeParameters` | `stage1/cohere/lint/helpers/wave12/control_flow_type_parameters.a:7` | retained candidate |
| `ecmascript/control_flow_graph.isAlwaysTruthyTest` | `stage1/cohere/lint/helpers/wave12/control_flow_is_always_truthy_test.a:10` | retained candidate |
| `ecmascript/control_flow_graph.isBreakableStatement` | `stage1/cohere/lint/helpers/wave12/control_flow_is_breakable_statement.a:6` | retained candidate |
| `ecmascript/control_flow_graph.labelsOf` | `stage1/cohere/lint/helpers/wave12/control_flow_labels_of.a:9` | retained candidate |
| `ecmascript/control_flow_graph.normalizeBigIntLiteral` | `stage1/cohere/lint/helpers/wave12/control_flow_normalize_bigint_literal.a:2` | retained candidate |
| `rules/tailwind/collapse.nextBuildCount` | `stage1/cohere/lint/helpers/wave12/gaps/atomic-counter.a:4` | BLOCKED source/probe only |
| `rules/tailwind/collapse.normalizeValueFunctionArgument` | `stage1/cohere/lint/helpers/wave12/collapse_normalize_value_function_argument.a:10` | retained candidate |
| `rules/tailwind/collapse.segment` | `stage1/cohere/lint/helpers/wave12/collapse_segment.a:4` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/wave12/bigint_test.go:27`: TestBigIntNormalizationMatchesGo @27, TestBigIntNormalizationMutants @38.
- `stage1/cohere/lint/helpers/wave12/block_test.go:12`: TestControlFlowBlockAllocation @12.
- `stage1/cohere/lint/helpers/wave12/breakable_test.go:27`: TestBreakableStatementMatchesGo @27.
- `stage1/cohere/lint/helpers/wave12/counter_gap_test.go:13`: TestCounterExactPrimitiveGap @13.
- `stage1/cohere/lint/helpers/wave12/decorators_test.go:40`: TestDecoratorsMatchesGo @40.
- `stage1/cohere/lint/helpers/wave12/labels_test.go:28`: TestLabelsOfMatchesGo @28, TestLabelsNilRefused @99.
- `stage1/cohere/lint/helpers/wave12/normalization_test.go:91`: TestNormalizationMatchesGo @91, TestNormalizationMutants @102.
- `stage1/cohere/lint/helpers/wave12/segment_test.go:15`: TestSegmentMatchesGo @15, TestSegmentMutants @26, TestSegmentSeparatorRefusals @65.
- `stage1/cohere/lint/helpers/wave12/truthy_test.go:28`: TestAlwaysTruthyTestMatchesGo @28, TestAlwaysTruthyNilRefused @99.
- `stage1/cohere/lint/helpers/wave12/type_parameters_test.go:40`: TestTypeParametersMatchesGo @40.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/wave12/CURRENT-LANDING.md:1`, `stage1/cohere/lint/helpers/wave12/REPORT.md:1`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 225 files; +214,103/-0 lines; 7,267,509 blob bytes.
- test/oracle/capture programs: 21 files; +1,375/-0 lines; 63,586 blob bytes.
- implementation/support types: 9 files; +211/-0 lines; 8,950 blob bytes.

### origin/codex/lint-helpers-from-lint-wave1-13

Tip `8d48f3f813a806a08d56dda1a71071af0c8a0674`; merge-base with area `d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898`; merge-tree clean (exit 0).

**Coverage verdict:** Three retained helpers ParseAtRule/parseModifier/NodeStylesheetResolver; Go helpers over six consumer literal families and byte/Unicode/alias/path controls, with three-backend compiling semantic mutants. Allocation/decoder/validation callbacks are actual Go facts in the isolated tests. NodeStylesheetResolver is pure POSIX resolution, not disk or Node execution. Current helper tests guard the old 715ba94f pin; migrate and revalidate before landing. NewTheme withdrawn.

Claim: `stage1/cohere/lint/helpers/claims/lint-wave1-13.md:1` (28 lines).

| Cohere package/function | Implementation at tip | Status |
|---|---|---|
| `rules/tailwind/collapse.NodeStylesheetResolver` | `stage1/cohere/lint/helpers/from-wave1-13/resolver/node_stylesheet_resolver.a:20` | retained candidate |
| `rules/tailwind/collapse.ParseAtRule` | `stage1/cohere/lint/helpers/from-wave1-13/at-rule/parse_at_rule.a:11` | retained candidate |
| `rules/tailwind/collapse.parseModifier` | `stage1/cohere/lint/helpers/from-wave1-13/modifier/parse_modifier.a:10` | retained candidate |

**Proof index:**
- `stage1/cohere/lint/helpers/comments/testdata/capture.py:1`: comparison/mutant assertion @1.
- `stage1/cohere/lint/helpers/from-wave1-13/at-rule/helper_test.go:20`: TestParseAtRule @20.
- `stage1/cohere/lint/helpers/from-wave1-13/modifier/helper_test.go:20`: TestParseModifier @20.
- `stage1/cohere/lint/helpers/from-wave1-13/resolver/helper_test.go:20`: TestNodeStylesheetResolver @20.

Archived reports / limits: `stage1/cohere/lint/helpers/REPORT.md:1`, `stage1/cohere/lint/helpers/comments/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/CONTINUATION_REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/LANDING_REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/at-rule/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/modifier/REPORT.md:1`, `stage1/cohere/lint/helpers/from-wave1-13/resolver/REPORT.md:1`.

Withdrawn witness paths: `stage1/cohere/lint/helpers/from-wave1-13/evidence/main.a.txt`, `stage1/cohere/lint/helpers/from-wave1-13/evidence/new_theme.a.txt`.

Volume (changed helper files; additions/deletions are textual lines; binary fixture line counts excluded):
- evidence/logs/docs/fixtures: 49 files; +3,263/-0 lines; 169,836 blob bytes.
- test/oracle/capture programs: 12 files; +1,002/-0 lines; 32,766 blob bytes.
- implementation/support types: 3 files; +78/-0 lines; 4,012 blob bytes.

## Rule-level arithmetic appendix

These are the 56 candidate-complete helper blocker sets. A blank option-ready flag is not implied: the original flag is shown. All require the common AST adapter, callback integration and a fresh parity/mutant gate. No findings or fix byte comparisons were performed in this triage.

| Rule | Option helper ready in ledger | First candidate package landing that completes blockers |
|---|---|---|
| `@next/next/google-font-display` | false | 7. ecmascript/text |
| `@next/next/google-font-preconnect` | false | 7. ecmascript/text |
| `@next/next/next-script-for-ga` | false | 7. ecmascript/text |
| `@next/next/no-css-tags` | false | 7. ecmascript/text |
| `@next/next/no-head-import-in-document` | false | 9. ecmascript/nextjs |
| `@next/next/no-img-element` | false | 2. ecmascript/jsx |
| `@next/next/no-location-assign-relative-destination` | false | 1. ecmascript/imports |
| `@next/next/no-styled-jsx-in-document` | false | 9. ecmascript/nextjs |
| `@next/next/no-sync-scripts` | false | 2. ecmascript/jsx |
| `@typescript-eslint/no-import-type-side-effects` | false | 1. ecmascript/imports |
| `@typescript-eslint/no-useless-empty-export` | false | 8. ecmascript/module |
| `base/consistency-no-bare-throw` | false | 1. ecmascript/imports |
| `complexity` | true | 5. ecmascript/property |
| `grouped-accessor-pairs` | true | 5. ecmascript/property |
| `nexus/boundary-no-internal-import` | false | 1. ecmascript/imports |
| `nexus/boundary-no-nexus-outside-import` | false | 1. ecmascript/imports |
| `nexus/boundary-no-project-import` | true | 1. ecmascript/imports |
| `nexus/consistency-no-screaming-snake-case` | true | 8. ecmascript/module |
| `nexus/import-no-forbidden-source` | false | 1. ecmascript/imports |
| `nexus/localization-no-untranslated-value` | false | 5. ecmascript/property |
| `no-prototype-builtins` | false | 5. ecmascript/property |
| `no-useless-escape` | true | 10. ecmascript/regexsyntax |
| `react-hooks/config` | true | 6. rules/react |
| `react/forbid-dom-props` | true | 2. ecmascript/jsx |
| `react/jsx-no-duplicate-props` | true | 2. ecmascript/jsx |
| `react/jsx-no-script-url` | true | 2. ecmascript/jsx |
| `react/jsx-no-target-blank` | true | 2. ecmascript/jsx |
| `react/no-danger` | true | 2. ecmascript/jsx |
| `react/no-deprecated` | false | 1. ecmascript/imports |
| `react/no-direct-mutation-state` | false | 6. rules/react |
| `react/no-render-return-value` | false | 5. ecmascript/property |
| `react/no-string-refs` | true | 4. ecmascript/react |
| `react/no-this-in-sfc` | false | 6. rules/react |
| `react/no-typos` | false | 4. ecmascript/react |
| `react/no-unknown-property` | true | 2. ecmascript/jsx |
| `react/prefer-es6-class` | true | 6. rules/react |
| `react/require-optimization` | true | 4. ecmascript/react |
| `react/require-render-return` | false | 6. rules/react |
| `react/state-in-constructor` | true | 4. ecmascript/react |
| `structure/boundary-no-project-theme-value` | false | 8. ecmascript/module |
| `structure/consistency-require-organized-imports` | false | 1. ecmascript/imports |
| `structure/import-require-react-namespace` | false | 4. ecmascript/react |
| `structure/network-no-direct-fetch` | false | 3. rules/structure |
| `structure/network-no-forbidden-import` | false | 3. rules/structure |
| `structure/next-no-page-state` | false | 4. ecmascript/react |
| `structure/next-require-api-parameter-name` | false | 3. rules/structure |
| `structure/react-component-no-display-name` | false | 4. ecmascript/react |
| `structure/react-component-no-forward-ref` | false | 4. ecmascript/react |
| `structure/react-component-no-separate-named-export` | false | 4. ecmascript/react |
| `structure/react-component-require-named-export` | false | 8. ecmascript/module |
| `structure/react-component-require-properties-parameter` | false | 4. ecmascript/react |
| `structure/react-element-no-anchor` | false | 3. rules/structure |
| `structure/react-element-no-horizontal-rule` | false | 3. rules/structure |
| `structure/react-hook-require-result-naming` | false | 3. rules/structure |
| `structure/storage-no-direct-local-storage` | false | 3. rules/structure |
| `yoda` | true | 5. ecmascript/property |

The other 80 rule rows still have these exact dependencies after all retained candidates; names are relative to the cohere lint root, except the literal strict-options placeholder:

- `@eslint-community/eslint-comments/require-description`: `ecmascript/directives.ParseDisable`, `ecmascript/directives.ParseEnable`, `ecmascript/directives.Recognize`, `ecmascript/directives.endsWord`, `ecmascript/directives.isLineComment`, `ecmascript/directives.parseRuleNames`, `ecmascript/directives.splitDirective`, `ecmascript/directives.splitReason`, `ecmascript/directives.splitScope`, `ecmascript/directives.stripCommentMarkers`.
- `@next/next/inline-script-id`: `ecmascript/imports.LocalNameOfDefaultImport`.
- `@next/next/no-before-interactive-script-outside-document`: `ecmascript/nextjs.IsDocumentPage`, `ecmascript/nextjs.IsInApplicationDirectory`.
- `@next/next/no-document-import-in-page`: `ecmascript/nextjs.IsDocumentPage`.
- `@next/next/no-head-element`: `ecmascript/nextjs.IsInApplicationDirectory`.
- `@next/next/no-html-link-for-pages`: `ecmascript/regexp.Compile`.
- `@next/next/no-page-custom-font`: `ecmascript/module.HasDefaultModifier`, `ecmascript/module.IsDefaultExported`.
- `@next/next/no-typos`: `ecmascript/nextjs.IsInPagesDirectory`, `ecmascript/nextjs.splitSegments`, `ecmascript/text.BestMatch`, `ecmascript/text.MinimumEditDistance`.
- `@next/next/no-unwanted-polyfillio`: `rules/next.urlQueryValue`.
- `@typescript-eslint/no-dupe-class-members`: `ecmascript/classmembers.ForEachDuplicate`, `ecmascript/classmembers.IsAccessorKind`, `ecmascript/classmembers.IsOverloadSignature`, `ecmascript/classmembers.KeyOf`, `ecmascript/classmembers.MemberName`, `ecmascript/property.NameTagged`.
- `@typescript-eslint/no-empty-object-type`: `ecmascript/regexp.Compile`.
- `@typescript-eslint/no-non-null-asserted-optional-chain`: `rules/typescript.nonNullAssertionOperatorRange`.
- `@typescript-eslint/no-this-alias`: `rules/typescript.isTypeScriptSourceFile`.
- `@typescript-eslint/no-unsafe-function-type`: `ecmascript/module.DeclaresTypeNamed`, `ecmascript/module.forEachDeclaredTypeName`.
- `@typescript-eslint/triple-slash-reference`: `rules/typescript.isTypeScriptSourceFile`.
- `array-callback-return`: `ecmascript/control_flow_graph.*Builder[E].appendSuccessor`.
- `base/security-require-context-access`: `ecmascript/decorators.CallName`, `ecmascript/decorators.HasDecoratorInSet`, `ecmascript/decorators.Of`.
- `better-tailwindcss/enforce-canonical-classes`: `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.candidateValueText`, `rules/tailwind.sourceValueAfterRoot`, `rules/tailwind.splitCandidateIn`, `rules/tailwind/collapse.nextBuildCount`.
- `better-tailwindcss/enforce-consistent-class-order`: `rules/tailwind.*ClassLiteralReader.ClassTemplateSegmentsIn`, `rules/tailwind.*strictVariantLevel.entry`, `rules/tailwind.*strictVariantLevel.order`, `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.anyCandidateResolves`, `rules/tailwind.atoi`, `rules/tailwind.buildClass`, `rules/tailwind.classOrderKeys`, `rules/tailwind.classesOf`, `rules/tailwind.compareVariantMasks`, `rules/tailwind.deprecationFor`, `rules/tailwind.isArbitraryVariant`, `rules/tailwind.isArrayIndexKey`, `rules/tailwind.jsKeyOrder`, `rules/tailwind.newStrictVariantLevel`, `rules/tailwind.parseTailwindVersion`, `rules/tailwind.propertyIndexBeyond`, `rules/tailwind.readingFor`, `rules/tailwind.segmentsOfTemplate`, `rules/tailwind.sortClassesByKey`, `rules/tailwind.strictClassOrder`, `rules/tailwind.strictVariantsOf`, `rules/tailwind.tailwindAtLeast`, `rules/tailwind/collapse.*Descriptor.axisFor`, `rules/tailwind/collapse.*LoadedDesignSystem.DeclaresFunctionalUtility`, `rules/tailwind/collapse.*LoadedDesignSystem.Variants`, `rules/tailwind/collapse.*RunVariantOrder.IndexOf`, `rules/tailwind/collapse.*RunVariantOrder.VariantBitmask`, `rules/tailwind/collapse.*Table.ArbitraryPropertyReading`, `rules/tailwind/collapse.*Table.Lookup`, `rules/tailwind/collapse.*Table.bareReading`, `rules/tailwind/collapse.*Table.frameworkReading`, `rules/tailwind/collapse.*UtilityEvaluator.Reading`, `rules/tailwind/collapse.*VariantRegistry.BuildVariantOrder`, `rules/tailwind/collapse.*VariantRegistry.Compare`, `rules/tailwind/collapse.AxisReadings.readingForType`, `rules/tailwind/collapse.ClassValueResolvesIn`, `rules/tailwind/collapse.CompareVariantBitmasks`, `rules/tailwind/collapse.FrameworkFunctionalUtility.Emit`, `rules/tailwind/collapse.FrameworkFunctionalUtility.Reading`, `rules/tailwind/collapse.FrameworkFunctionalUtility.staticValueEmitter`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.Emit`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.ReadingFor`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.readingOf`, `rules/tailwind/collapse.ModifierAxisFor`, `rules/tailwind/collapse.PrintVariant`, `rules/tailwind/collapse.ResolveFunctionalUtilityValue`, `rules/tailwind/collapse.appendVariantAndNested`, `rules/tailwind/collapse.bareValueHandler`, `rules/tailwind/collapse.bracketed`, `rules/tailwind/collapse.cloneNode`, `rules/tailwind/collapse.cloneNodes`, `rules/tailwind/collapse.declareComposing`, `rules/tailwind/collapse.declareComposingMask`, `rules/tailwind/collapse.declareComposingWebkit`, `rules/tailwind/collapse.declareProperties`, `rules/tailwind/collapse.declareProperty`, `rules/tailwind/collapse.declareSorted`, `rules/tailwind/collapse.declareWrapped`, `rules/tailwind/collapse.descriptionForRoot`, `rules/tailwind/collapse.escapeUnderscoreAndSpace`, `rules/tailwind/collapse.escapeUnderscoresAndSpaces`, `rules/tailwind/collapse.isLoneVar`, `rules/tailwind/collapse.isSpaceSeparator`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.printArbitraryValue`, `rules/tailwind/collapse.printModifier`, `rules/tailwind/collapse.printNodesCss`, `rules/tailwind/collapse.removeNodes`, `rules/tailwind/collapse.resolveArm`, `rules/tailwind/collapse.resolveArmByInferredType`, `rules/tailwind/collapse.resolveNamedValue`, `rules/tailwind/collapse.toPrintNodes`, `rules/tailwind/collapse.unwrapIsSelector`, `rules/tailwind/collapse.variantsAreEqual`, `rules/tailwind/collapse.withoutRemoved`.
- `better-tailwindcss/enforce-consistent-important-position`: `rules/tailwind.splitVariantPrefix`.
- `better-tailwindcss/enforce-consistent-variable-syntax`: `rules/tailwind.splitVariantPrefix`.
- `better-tailwindcss/enforce-consistent-variant-order`: `rules/tailwind.DesignSystemForProgram`, `rules/tailwind/collapse.*LoadedDesignSystem.Variants`, `rules/tailwind/collapse.*RunVariantOrder.IndexOf`, `rules/tailwind/collapse.*VariantRegistry.BuildVariantOrder`, `rules/tailwind/collapse.*VariantRegistry.Compare`, `rules/tailwind/collapse.appendVariantAndNested`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.variantsAreEqual`.
- `better-tailwindcss/enforce-shorthand-classes`: `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.anyCandidateResolves`, `rules/tailwind.classExistsIn`, `rules/tailwind.splitVariantPrefix`, `rules/tailwind/collapse.*LoadedDesignSystem.DeclaresFunctionalUtility`, `rules/tailwind/collapse.*UtilityEvaluator.Reading`, `rules/tailwind/collapse.ClassValueResolvesIn`, `rules/tailwind/collapse.ResolveFunctionalUtilityValue`, `rules/tailwind/collapse.bareValueHandler`, `rules/tailwind/collapse.cloneNode`, `rules/tailwind/collapse.cloneNodes`, `rules/tailwind/collapse.descriptionForRoot`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.removeNodes`, `rules/tailwind/collapse.resolveArm`, `rules/tailwind/collapse.resolveArmByInferredType`, `rules/tailwind/collapse.resolveNamedValue`.
- `better-tailwindcss/no-concatenated-classes`: `rules/tailwind.*ClassLiteralReader.ClassTemplatesIn`, `rules/tailwind.boundaryAfterHole`, `rules/tailwind.boundaryBeforeHole`, `rules/tailwind.endsWithWhitespace`, `rules/tailwind.startsWithWhitespace`, `rules/tailwind.templateExpressionOf`, `rules/tailwind.templatesFrom`.
- `better-tailwindcss/no-conflicting-classes`: `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.composesForRoot`, `rules/tailwind.repositoryClassFacts`, `rules/tailwind.resolveClassFactsIn`, `rules/tailwind.valueIsColorIn`, `rules/tailwind.valueResolutionIn`, `rules/tailwind/collapse.*Theme.Get`, `rules/tailwind/collapse.*UtilityEvaluator.Has`, `rules/tailwind/collapse.ComposesFor`, `rules/tailwind/collapse.DeclaredPropertiesFor`, `rules/tailwind/collapse.EmitGapRoot`, `rules/tailwind/collapse.FrameworkFunctionalUtility.Emit`, `rules/tailwind/collapse.FrameworkFunctionalUtility.staticValueEmitter`, `rules/tailwind/collapse.FrameworkMultiDeclarationUtility.Emit`, `rules/tailwind/collapse.IsColorKeyword`, `rules/tailwind/collapse.SelectorShapeForRoot`, `rules/tailwind/collapse.SelectorShapeOfNodes`, `rules/tailwind/collapse.UtilityBranch.isOneOf`, `rules/tailwind/collapse.UtilityBranch.maskStopIsColor`, `rules/tailwind/collapse.UtilityBranch.shadowIsColor`, `rules/tailwind/collapse.UtilityBranchFor`, `rules/tailwind/collapse.absentValued`, `rules/tailwind/collapse.appendVisibleProperty`, `rules/tailwind/collapse.borderSideEmitter`, `rules/tailwind/collapse.cloneNode`, `rules/tailwind/collapse.cloneNodes`, `rules/tailwind/collapse.collectVisibleProperties`, `rules/tailwind/collapse.containsString`, `rules/tailwind/collapse.declarations`, `rules/tailwind/collapse.declareComposing`, `rules/tailwind/collapse.declareComposingMask`, `rules/tailwind/collapse.declareComposingWebkit`, `rules/tailwind/collapse.declareProperties`, `rules/tailwind/collapse.declareProperty`, `rules/tailwind/collapse.declareSorted`, `rules/tailwind/collapse.declareWrapped`, `rules/tailwind/collapse.emitFunctionalRoot`, `rules/tailwind/collapse.emitRootWithValue`, `rules/tailwind/collapse.gapTypeListFor`, `rules/tailwind/collapse.gradientPositionEmitter`, `rules/tailwind/collapse.maskEdgeEmitter`, `rules/tailwind/collapse.maskGradientEmitter`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.removeNodes`, `rules/tailwind/collapse.ringEmitter`, `rules/tailwind/collapse.rotateEmitter`, `rules/tailwind/collapse.scaleEmitter`, `rules/tailwind/collapse.shadowFamilyEmitter`, `rules/tailwind/collapse.visibleDeclarationText`, `rules/tailwind/collapse.withResolvedValue`.
- `better-tailwindcss/no-deprecated-classes`: `rules/tailwind.classesOf`, `rules/tailwind.containsString`.
- `better-tailwindcss/no-duplicate-classes`: `rules/tailwind.*ClassLiteralReader.ClassTemplateSegmentsIn`, `rules/tailwind.segmentsOfTemplate`.
- `better-tailwindcss/no-unknown-classes`: `rules/tailwind.DesignSystemForProgram`, `rules/tailwind.anyCandidateResolves`, `rules/tailwind.classExistsIn`, `rules/tailwind.compileIgnorePatterns`, `rules/tailwind.isIgnored`, `rules/tailwind/collapse.*LoadedDesignSystem.DeclaresFunctionalUtility`, `rules/tailwind/collapse.*UtilityEvaluator.Reading`, `rules/tailwind/collapse.ClassValueResolvesIn`, `rules/tailwind/collapse.ResolveFunctionalUtilityValue`, `rules/tailwind/collapse.bareValueHandler`, `rules/tailwind/collapse.cloneNode`, `rules/tailwind/collapse.cloneNodes`, `rules/tailwind/collapse.descriptionForRoot`, `rules/tailwind/collapse.nextBuildCount`, `rules/tailwind/collapse.removeNodes`, `rules/tailwind/collapse.resolveArm`, `rules/tailwind/collapse.resolveArmByInferredType`, `rules/tailwind/collapse.resolveNamedValue`.
- `better-tailwindcss/no-unnecessary-whitespace`: `rules/tailwind.*ClassLiteralReader.ClassSegmentsIn`, `rules/tailwind.classesOf`, `rules/tailwind.segmentsOfTemplate`.
- `boundaries/dependencies`: `strict option decoding and schema validation`.
- `consistent-return`: `ecmascript/consistentreturn.HasValue`, `ecmascript/consistentreturn.IsGenerator`, `ecmascript/consistentreturn.IsScope`, `ecmascript/consistentreturn.Judge`, `ecmascript/consistentreturn.Name`, `ecmascript/consistentreturn.ReportRange`, `ecmascript/consistentreturn.Verb`, `ecmascript/consistentreturn.canRunOffEnd`, `ecmascript/consistentreturn.capitaliseFirst`, `ecmascript/consistentreturn.isExemptFromEndJudgment`, `ecmascript/consistentreturn.judgeReturns`, `ecmascript/consistentreturn.judgeScope`, `ecmascript/consistentreturn.staticName`, `ecmascript/control_flow_graph.*Builder[E].appendSuccessor`.
- `constructor-super`: `rules/core.isSeparateEvaluationContext`.
- `dot-notation`: `ecmascript/dotnotation.CompileAllowPattern`, `ecmascript/dotnotation.Listeners`, `ecmascript/dotnotation.checkComputed`, `ecmascript/dotnotation.computedFix`, `ecmascript/dotnotation.continuesAnIdentifier`, `ecmascript/dotnotation.hasCommentBetween`, `ecmascript/dotnotation.isOptionalAccess`, `ecmascript/dotnotation.jsonQuote`, `ecmascript/dotnotation.keywordFix`, `ecmascript/dotnotation.literalKey`, `ecmascript/dotnotation.useBracketsMessage`, `ecmascript/dotnotation.useDotMessage`.
- `id-length`: `ecmascript/text.GraphemeCount`, `ecmascript/text.continuesCluster`, `ecmascript/text.hangulJoins`, `ecmascript/text.hangulLeading`, `ecmascript/text.hangulSyllable`, `ecmascript/text.hangulTrailing`, `ecmascript/text.hangulVowel`, `ecmascript/text.hangulVowelSyllable`, `ecmascript/text.isGraphemeControl`, `ecmascript/text.isGraphemeExtend`, `ecmascript/text.isIndicLinker`, `ecmascript/text.isPictograph`, `ecmascript/text.isRegionalIndicator`, `rules/core.idDenylistIsDestructuringTarget`, `rules/core.idDenylistIsImportAttributeKey`, `rules/core.idDenylistIsImportOptionsObject`, `strict option decoding and schema validation`.
- `nexus/import-require-path-alias`: `strict option decoding and schema validation`.
- `no-constant-condition`: `rules/core.numericLiteralSign`, `strict option decoding and schema validation`.
- `no-control-regex`: `ecmascript/regexpattern.*walker.applyQuantifier`, `ecmascript/regexpattern.*walker.emit`, `ecmascript/regexpattern.*walker.emitAndQuantify`, `ecmascript/regexpattern.*walker.kindFromSource`, `ecmascript/regexpattern.*walker.readEscape`, `ecmascript/regexpattern.*walker.run`, `ecmascript/regexpattern.*walker.walkClass`, `ecmascript/regexpattern.Walk`, `ecmascript/regexpattern.braceQuantifierEnd`, `ecmascript/regexpattern.consumeDigits`, `ecmascript/regexpattern.decodeRune`, `ecmascript/regexpattern.escapeValue`, `ecmascript/regexpattern.extendOctalEscape`, `ecmascript/regexpattern.groupPrologueEnd`, `ecmascript/regexpattern.isAsciiLetter`, `ecmascript/regexpattern.octalEscapeValue`, `ecmascript/regexpattern.quantifierAt`, `ecmascript/regexpattern.unicodeEscapeValue`, `ecmascript/regexsyntax.ParseHexUint`, `ecmascript/regexsyntax.ParseRegexCharacterClassWithEnd`, `ecmascript/regexsyntax.hexValue`, `ecmascript/regexsyntax.parseRegexCharacterClass`, `ecmascript/regexsyntax.readClassEscape`, `ecmascript/regexsyntax.readRawClassChar`.
- `no-dupe-class-members`: `ecmascript/classmembers.ForEachDuplicate`, `ecmascript/classmembers.IsAccessorKind`, `ecmascript/classmembers.IsOverloadSignature`, `ecmascript/classmembers.KeyOf`, `ecmascript/classmembers.MemberName`, `ecmascript/property.NameTagged`.
- `no-dupe-else-if`: `rules/core.appendNodeSignature`, `rules/core.appendTokensBetween`, `rules/core.hasSameTokens`, `rules/core.isEmptyBracketLiteral`, `rules/core.tokenSignature`.
- `no-empty-function`: `rules/core.consistentReturnIsGenerator`.
- `no-extra-boolean-cast`: `strict option decoding and schema validation`.
- `no-regex-spaces`: `ecmascript/literal.CookedToRaw`, `ecmascript/literal.cookedBytesProducedBy`, `ecmascript/literal.escapeWidthInStringLiteral`, `ecmascript/literal.producesNoCookedBytes`, `ecmascript/regexpattern.*walker.applyQuantifier`, `ecmascript/regexpattern.*walker.emit`, `ecmascript/regexpattern.*walker.emitAndQuantify`, `ecmascript/regexpattern.*walker.kindFromSource`, `ecmascript/regexpattern.*walker.readEscape`, `ecmascript/regexpattern.*walker.run`, `ecmascript/regexpattern.*walker.walkClass`, `ecmascript/regexpattern.Walk`, `ecmascript/regexpattern.braceQuantifierEnd`, `ecmascript/regexpattern.consumeDigits`, `ecmascript/regexpattern.decodeRune`, `ecmascript/regexpattern.escapeValue`, `ecmascript/regexpattern.extendOctalEscape`, `ecmascript/regexpattern.groupPrologueEnd`, `ecmascript/regexpattern.isAsciiLetter`, `ecmascript/regexpattern.octalEscapeValue`, `ecmascript/regexpattern.quantifierAt`, `ecmascript/regexpattern.unicodeEscapeValue`, `ecmascript/regexsyntax.ParseHexUint`, `ecmascript/regexsyntax.ParseRegexCharacterClassWithEnd`, `ecmascript/regexsyntax.hexValue`, `ecmascript/regexsyntax.parseRegexCharacterClass`, `ecmascript/regexsyntax.readClassEscape`, `ecmascript/regexsyntax.readRawClassChar`.
- `no-restricted-exports`: `ecmascript/regexp.Compile`, `strict option decoding and schema validation`.
- `no-restricted-imports`: `ecmascript/regexp.Compile`, `rules/core.noRestrictedExportsNameText`.
- `no-this-before-super`: `rules/core.bodyDefinitelyExits`, `rules/core.switchStatementExits`, `rules/core.tryStatementExits`.
- `no-unreachable-loop`: `ecmascript/control_flow_graph.*Builder[E].Current`, `ecmascript/control_flow_graph.*Builder[E].appendSuccessor`, `ecmascript/control_flow_graph.IndexRoots`, `rules/core.codePathRoots`.
- `no-useless-call`: `rules/core.appendNodeSignature`, `rules/core.appendTokensBetween`, `rules/core.hasSameTokens`, `rules/core.isEmptyBracketLiteral`, `rules/core.isNullOrUndefined`, `rules/core.memberAccessObject`, `rules/core.tokenSignature`.
- `object-shorthand`: `strict option decoding and schema validation`.
- `prefer-spread`: `rules/core.appendNodeSignature`, `rules/core.appendTokensBetween`, `rules/core.tokenSignature`.
- `react-hooks/error-boundaries`: `rules/react.isComponentIdentifierName`, `rules/react.isHookIdentifierName`.
- `react-hooks/incompatible-library`: `ecmascript/react.FunctionBody`, `ecmascript/react.FunctionParameters`, `ecmascript/react.IsCompilerComponentName`, `ecmascript/react.IsCompilerHookName`, `ecmascript/react.IsComponentOrHookLike`, `ecmascript/react.IsReachableRootPosition`, `ecmascript/react.SkipParenthesesUpward`, `ecmascript/react.callsHooksOrCreatesJsx`, `ecmascript/react.containsRef`, `ecmascript/react.containsSubstring`, `ecmascript/react.hasComponentOrHookName`, `ecmascript/react.hasComponentShapedParameters`, `ecmascript/react.hasPrimitiveTypeAnnotation`, `ecmascript/react.inferredFunctionName`, `ecmascript/react.isCompilerHookCallee`, `ecmascript/react.isJsxNode`, `ecmascript/react.isSkippedNestedFunction`, `ecmascript/react.isSpreadParameter`.
- `react-hooks/rules-of-hooks`: `ecmascript/control_flow_graph.*Block[E].Index`, `ecmascript/control_flow_graph.*Builder[E].Emit`, `ecmascript/control_flow_graph.*Builder[E].appendSuccessor`, `ecmascript/control_flow_graph.*PathAnalysis[E].IsCyclic`, `ecmascript/control_flow_graph.*PathAnalysis[E].IsOnEveryFinalPath`, `ecmascript/control_flow_graph.*PathAnalysis[E].countFinalPaths`, `ecmascript/control_flow_graph.*PathAnalysis[E].findFinalPathDominators`, `ecmascript/control_flow_graph.*PathAnalysis[E].findShortestPaths`, `ecmascript/control_flow_graph.*cyclePaths[E].count`, `ecmascript/control_flow_graph.*cyclePaths[E].countToEnd`, `ecmascript/control_flow_graph.AnalyzePaths`, `ecmascript/control_flow_graph.newCyclePaths`.
- `react-hooks/void-use-memo`: `rules/react.enclosingFunctionOf`, `rules/react.hasValidComponentParameters`, `rules/react.isComponentIdentifierName`, `rules/react.isFunctionLike`, `rules/react.isHookIdentifierName`, `rules/react.isNonNodeExpression`, `rules/react.isReactCompiledFunction`, `rules/react.isRestParameter`, `rules/react.isTopLevelCompilationCandidate`, `rules/react.mentionsRef`, `rules/react.parametersOf`, `rules/react.reactFunctionNameOf`, `rules/react.returnsNonNode`.
- `react/default-props-match-prop-types`: `rules/react.enclosingClassOf`, `rules/react.functionBodyBlock`, `rules/react.semanticParentOf`, `rules/react.sortDefaultPropsInitializerOf`.
- `react/jsx-key`: `rules/react.attributesOf`, `rules/react.commentValueOf`, `rules/react.isJavaScriptIdentifier`, `rules/react.jsxAnnotationIn`, `rules/react.reactPragmaFor`, `rules/react.sourceSliceOf`.
- `react/jsx-props-no-spread-multi`: `ecmascript/property.NameTagged`.
- `react/no-access-state-in-setstate`: `rules/react.enclosingComponentOf`, `rules/react.isComponentClass`, `rules/react.isCreateReactClassCall`, `rules/react.isReactComponentBase`.
- `react/no-arrow-function-lifecycle`: `rules/react.stylePropObjectUnwrapParentheses`.
- `react/no-children-prop`: `ecmascript/react.IsCreateElementCall`.
- `react/no-did-mount-set-state`: `rules/react.DecodeNoMethodSetStateOptions`.
- `react/no-did-update-set-state`: `rules/react.DecodeNoMethodSetStateOptions`.
- `react/no-set-state`: `rules/react.isComponentClass`, `rules/react.isCreateReactClassCall`, `rules/react.isReactComponentBase`.
- `react/no-unused-class-component-methods`: `rules/react.isThisExpression`.
- `react/no-unused-state`: `rules/react.isComponentClass`, `rules/react.isCreateReactClassCall`, `rules/react.isReactComponentBase`.
- `react/no-will-update-set-state`: `rules/react.DecodeNoMethodSetStateOptions`.
- `react/prefer-stateless-function`: `rules/react.isPureComponentBase`.
- `react/void-dom-elements-no-children`: `ecmascript/react.IsCreateElementCall`.
- `structure/consistency-require-matching-file-name`: `rules/structure.isAnonymousWrapperCall`.
- `structure/network-no-invalidate-cache-in-on-success`: `rules/structure.callExpressionCallee`, `rules/structure.isCacheInvalidateCall`.
- `structure/network-require-hook-options-parameter`: `rules/structure.*NetworkFileAnalysis.collect`, `rules/structure.*NetworkFileAnalysis.collectFunctionDeclaration`, `rules/structure.*NetworkFileAnalysis.collectVariableStatement`, `rules/structure.NetworkFileAnalysisFor`, `rules/structure.bodyCallsNetworkService`, `rules/structure.findNetworkServiceCalls`, `rules/structure.parameterName`.
- `structure/network-require-hook-request-suffix`: `rules/structure.*NetworkFileAnalysis.collect`, `rules/structure.*NetworkFileAnalysis.collectFunctionDeclaration`, `rules/structure.*NetworkFileAnalysis.collectVariableStatement`, `rules/structure.NetworkFileAnalysisFor`, `rules/structure.bodyCallsNetworkService`, `rules/structure.findNetworkServiceCalls`.
- `structure/network-require-hook-variables-type`: `rules/structure.*NetworkFileAnalysis.collect`, `rules/structure.*NetworkFileAnalysis.collectFunctionDeclaration`, `rules/structure.*NetworkFileAnalysis.collectVariableStatement`, `rules/structure.NetworkFileAnalysisFor`, `rules/structure.bodyCallsNetworkService`, `rules/structure.findNetworkServiceCalls`, `rules/structure.parameterName`.
- `structure/next-no-near-miss-route-export`: `ecmascript/nextjs.RouteContractExports`, `ecmascript/nextjs.buildRouteFileContracts`, `ecmascript/nextjs.splitSegments`, `ecmascript/text.BestMatch`, `ecmascript/text.MinimumEditDistance`.
- `structure/next-require-page-default-export`: `ecmascript/module.HasDefaultModifier`.
- `structure/react-component-no-const-assignment`: `rules/structure.isForwardRefCall`.
- `structure/react-component-no-destructuring`: `ecmascript/scope.BodyOf`, `ecmascript/scope.EnclosingFunctionLike`, `ecmascript/scope.NameOf`, `ecmascript/scope.borrowedName`.
- `structure/react-component-require-properties-type-suffix`: `ecmascript/module.AllDeclaredTypeNames`, `ecmascript/module.forEachDeclaredTypeName`.
- `structure/react-hook-no-destructuring`: `ecmascript/scope.NameOf`, `ecmascript/scope.borrowedName`.
- `structure/react-hook-no-properties-in-dependencies`: `ecmascript/scope.NameOf`, `ecmascript/scope.borrowedName`.
- `structure/react-hook-require-effect-comment`: `ecmascript/comments.LeadingRunFor`, `ecmascript/comments.isAdjacentGap`.

## What this audit did and did not prove

Observed: explicit remote ref enumeration; exact tips; current main and area inventory/submodule pins; committed implementation declarations/test assertions/claim withdrawals; current whole-tip merge-tree exits and conflict paths; measured Git blob sizes; deterministic set arithmetic. Inferred recommendations: package source selection, constructor semantic agreement within the stated representation, and conditional helper-ready counts. Not observed: a new Node/native/emitted-JS helper run, a consuming-rule findings/fixes gate, pairwise merged-package tests, current compiler support closure, or a live Tailwind corpus re-provisioning. Old successful logs are not relabeled as current gates. A fresh integration test must fail compiling semantic mutants by output comparison, not by pin guard, compile refusal, crash, or omitted-input skip.

Repository state: detached checkout unchanged, no commits and no pushes. Only remote-tracking refs/unreachable merge objects and scratch audit artifacts were written; this report is the sole requested workspace deliverable.
