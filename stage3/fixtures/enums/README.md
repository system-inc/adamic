# Enums from TypeScript 6.0.3

Thirteen executable fixtures and one source-only dependency answer the census reason `enum`: 164 declarations in 33 original compiler files. The census counts declarations, not runtime-use frequencies. This sample emphasizes range checks and flags, then collections, strings and runtime reflection; it does not claim a usage-frequency census. Source is v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Each copied function and enum retains upstream statements. Minimal interfaces and explicitly separate helpers supply the standalone contracts. Full enum declarations retain implicit numbering, aliases and composite initializers rather than manufacturing smaller domains.

Read census REPORT.md and its enum rows in data/sites.json at origin/codex/tsc-census (429c1177f0130f785c19cf590d1860513b2ddbfc), and stage3/README.md at origin/codex/stage3-base (8728405135d329efc12c837a7a6c293234abbe1c). Neither tree is present on the main base; neither was copied into this fixture branch.

| Fixture | Real form represented |
| --- | --- |
| 01 | isTokenKind, inclusive FirstToken/LastToken range and boundary driver |
| 02 | isNodeKind, lower-bound range and implicit numeric members |
| 03 | isJSDocNode, inclusive FirstJSDocNode/LastJSDocNode range |
| 04 | isParseTreeNode, NodeFlags mask, combined driver flags |
| 05 | getCombinedLocalAndExportSymbolFlags, SymbolFlags OR and absent exportSymbol |
| 06 | setNodeFlags, NodeFlags OR and AND-complement updates, undefined input |
| 07 | scanner's makeReverseMap and regex helpers, enum-indexed sparse arrays and enum-keyed Map |
| 08 | complete Debug.formatEnum/getEnumMembers bodies, caching, zero, unknown value and overlapping flags |
| 09 | SymbolFlags exclusion/complement expressions, signed All=-1, member-derived initializers |
| 10 | getDeclarationEmitExtensionForPath, string enum branches |
| 11 | formatSyntaxKind with the full formatter, alias names and undefined default; syntax-kind-support.a supplies the genuine namespace import |
| 12 | diagnosticCategoryName, genuine numeric reverse lookup and case conversion |
| 13 | isolated scanner Map lookup, missing key, pure flag combination and const members |

08 and 11 retain `any`, for-in and unchecked dense-array access. They intentionally show the checker barrier rather than rewrite the hardest real case into an easier implementation. The local sorted-array helpers stand for core's sorting contract; they are not claimed as extracted upstream functions. 10's suffix helper supplies the path predicate for these driver inputs. 06 retains the original generic Mutable cast. The namespace import in 11 models debug.ts's import of _namespaces/ts.js; the helper contains the original SyntaxKind enum only and is not an independent fixture.

## Observations

Main base ef3d907ecdc4c771b016f7d9c52372def057a340: all 13 are Checker, TS1294. status.json contains exact main observations and Node stdout/stderr/exit, with no additional fields. Node uses oracle/node.mjs from origin/codex/flag-enums at f7d62772fa8e52fcfae754e047ceb66dba88b782, copied into scratch with its runtime. This runner uses stripTypeScriptTypes mode transform, enabling non-erasable syntax. All 13 Node runs exit 0 with empty stderr.

On that exact flag-enums revision: 01, 02, 03, 10, 12 and 13 compile; every binary agrees with Node on stdout, stderr and exit. No silent miscompile was observed. 04, 05, 06, 07 and 09 are Refused, and 08 and 11 are Checker. See flag-enums-results.json for exact diagnostics and native observations. In particular, real NodeFlags/SymbolFlags declarations include aliases or a signed All member outside the branch's restricted flag-declaration recognizer. 07 is refused because makeReverseMap widens a mutable Map<CharacterCodes, RegularExpressionFlags> to Map<CharacterCodes, number>. 08/11 hit TS2532 on members[0], despite the length guard.

Const-enum member inlining was also observed in generated C for 13: the map uses numeric literal keys/values, `flags` is initialized with `adamic_bitwise_or((0x1p+01), (0x1p+02))`, and the scanner lookup is called with `(0x1.9cp+06)` for CharacterCodes.g. This establishes value-use inlining, not a separate stock-tsc emit comparison. Node's transform runner itself does not prove const-enum erasure.

Commands (all test output redirected to logs):

```sh
bash cloud/setup.sh > /tmp/enums-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Each fixture, from the main repository root:
node --disable-warning=ExperimentalWarning /workspace/scratch/enums-runner/node.mjs stage3/fixtures/enums/<file>.a
go run ./cmd/adamic build stage3/fixtures/enums/<file>.a -o /workspace/scratch/enums-results/<name>.main
# Build the isolated feature revision, then run each fixture:
go build -o /workspace/scratch/enums-flag ./cmd/adamic
/workspace/scratch/enums-flag build stage3/fixtures/enums/<file>.a -o /workspace/scratch/enums-results/<name>.flags
# In the feature checkout:
go test ./internal/lower -run 'Test.*Enum' -count=1 -timeout 10m > /tmp/enums-lower-tests.log 2>&1
```

Setup passed in 111s; go/clang/Node/submodules ready at 0s, cache warm at 111s; nproc 5, cgroup quota four CPUs. Feature checkout reused the identical pinned cohere and shim checkout through scratch symlinks after its initial build reported missing submodule files. Lowering enum regressions pass in 1.163s. No Go files or shared fixture harness were changed; the full integration gate and TypeScript's full suite were not run for this fixture-only unit.

## Number mutant

In 13's driver, replace one use of RegularExpressionFlags.Global with its exact number, 2. Node still prints `6\n2\nundefined\n`, exit 0. Feature stage 0 flips Compiles to Refused: adamic/enum-flags, a number outside the proven flag domain. mutant.json preserves the exact diagnostic. This proves the domain check can fail without changing runtime semantics. Main remains Checker/TS1294: replacing a use cannot remove its earlier enum-declaration syntax gate. A flip on main from this one-use mutant is therefore not demonstrated.

## Every executable enum-object indexing site

A scan of all original src/compiler .ts files against all declared enum names found the following 13 executable sites (12 numeric and one string). The examples `E[E.A]` at checker.ts:19309 and `DiagnosticCategory[d.category]` in older discussions are not additional sites: the former is a comment and the latter is the executable types.ts:7321 site listed below. No SyntaxKind[number] occurs in this pinned compiler source.

| Source | Read | Soundness |
| --- | --- | --- |
| moduleNameResolver.ts:1430 | ModuleResolutionKind[moduleResolution] | Defined string for a validated ModuleResolutionKind; arbitrary numeric API input can return undefined |
| moduleNameResolver.ts:1435 | same | Same contract |
| builder.ts:1631 | DiagnosticCategory[diagnostic.category] | Defined for compiler-produced categories; unvalidated external numeric diagnostics need checking |
| checker.ts:4094 | ModuleKind[moduleKind] | Defined for validated ModuleKind; an arbitrary numeric options value has no total-lookup proof |
| checker.ts:46857 | same | Same contract |
| checker.ts:48695 | same | Same contract |
| program.ts:4374 | ModuleKind[moduleKind] | Truthiness guard safely tests absence |
| program.ts:4378 | same | Defined after the preceding guard, assuming enum object unchanged |
| program.ts:4379 | ModuleResolutionKind[moduleKindName as any] | String forward lookup, not numeric reverse lookup; truthiness safely handles a missing name, but the any cast supplies no static proof |
| program.ts:4383 | ModuleResolutionKind[moduleResolution] | Truthiness guard safely tests absence |
| program.ts:4387 | same | Defined after the preceding guard, assuming enum object unchanged |
| program.ts:4557 | ModuleKind[options.module] | Defined for validated ModuleKind; optional/malformed external options are not proven by this lookup alone |
| types.ts:7321 | DiagnosticCategory[d.category] | Defined over the closed declared category domain; demonstrated by fixture 12 |

None of the unguarded sites is sound as a total `number -> string` function. They are sound when their input is proven to be a declared member, and ordinary enum duplicates still return a string (the last declared name). Values and combinations outside an enum's declared members must return string | undefined. Debug's formatter follows a different route and selects the first stable-sorted declaration for duplicate values.

Debug's reflection engine is debug.ts:422-442: for-in enumerates names, enumObject[name] at 433 reads a name back, and typeof value === "number" filters numeric members. formatEnum at 389-418 matches numbers against that list. It safely falls back to value.toString() for unknown values; undefined takes the default zero. Under immutable, unpolluted compiler enum objects and the dense sorted-array contract this has sound runtime results. Its any types, inherited-property iteration and cache immutability assumption are not static soundness proofs.

All adapters to this engine in debug.ts are: formatSyntaxKind 444, formatSnippetKind 448, formatScriptKind 452, formatNodeFlags 456, formatNodeCheckFlags 460, formatModifierFlags 464, formatTransformFlags 468, formatEmitFlags 472, formatSymbolFlags 476, formatTypeFlags 480, formatSignatureFlags 484, formatObjectFlags 488, formatFlowFlags 492, formatRelationComparisonResult 496, formatCheckMode 500, formatSignatureCheckMode 504, and formatTypeFacts 508. There is also a direct formatEnum call at 540 (FlowFlags debug property); the remaining-SymbolFlags expression at 628 calls formatSymbolFlags, the adapter listed above. These adapters use the same runtime contracts; they are not separate numeric enum-object index expressions.

## Limits

This does not cover all 164 enum declarations, every scalar enum call site, mutations of exported enum objects, or diagnostic host implementations. The complete reflection fixture cannot pass the current checker unchanged. Real flag and generic Map blockers are preserved for the implementation worker. Signed-mask behavior has a Node observation but no native run because the feature branch refuses that declaration. The numeric reverse-lookup inventory is a declared-name source scan with the Debug generic engine audited separately; dynamic aliases other than that engine were not exhaustively analyzed. No frequency claim beyond the declaration census is made. The full gate, stock-tsc emit inlining comparison and a main outcome flip remain unperformed for the reasons above.


## Implementation follow-up

On `codex/flag-enums`, after merging this branch at `671e1fd`, the original
thirteen fixtures have the following observations. None of their source bodies,
drivers or contracts were changed. Historical `status.json` and
`flag-enums-results.json` remain intact; `implementation-results.json` records
the new exact observations.

| Outcome | Before | After |
| --- | ---: | ---: |
| Compiles and agrees byte for byte with Node | 6 | 7 |
| Refused | 5 | 4 |
| Checker | 2 | 2 |

04 now compiles: direct same-enum aliases preserve the non-negative flag domain,
and a const binding iterating a fresh inline array of proven flags keeps that
proof through a shorthand field. Numeric shape alone still cannot opt in.
The seven compiling fixtures also pass the source Node, JavaScript backend,
native release, ASan/UBSan and leak oracle.

| Remaining fixture | Why its original contract remains blocked |
| --- | --- |
| 05 local/export flags | SymbolFlags contains All = -1, complement initializers and member-derived masks. It is outside the approved non-negative flag domain. The OR return is an arbitrary number relative to its closed member union. A signed mask domain needs its own decision; alternatively use number for open masks and validate when returning a closed enum member. |
| 06 setNodeFlags | Mutable<T> permits writing a full NodeFlags into a flags slot that a caller can constrain to one narrower member. Removing readonly also opens writes through a covariant view. For example, T can have flags: NodeFlags.Let, while newFlags is NodeFlags.Const. Keep an invariant NodeFlags field contract or copy into a new full-domain node rather than claiming the result remains T. |
| 07 regex array/map | Map<CharacterCodes, RegularExpressionFlags> cannot become writable Map<CharacterCodes, number>; that view could store 99 into an enum slot. ReadonlyMap is a safe reading contract, but the original result[value] writes also create a sparse array, outside Adamic's dense-array contract. Use a readonly source view and a Map for the reverse table. |
| 09 symbol exclusions | It has the same signed SymbolFlags declaration and open OR return as 05. Its negative complement outputs are correct numbers on Node, but do not establish a non-negative flag domain. Keep complements in number or design a distinct signed-mask domain. |
| 08 Debug formatter | TS2532 rejects members[0] despite the length guard. Array length alone does not prove a present element in general JavaScript arrays. The body also uses any for reflection and cache keys. Narrow a checked first-element read and give reflection an explicit typed contract; disabling noUncheckedIndexedAccess or admitting any would remove soundness checks. |
| 11 SyntaxKind formatter | The same unchecked reads and any reflection/cache contract as 08, plus (ts as any).SyntaxKind. Use a named typed enum import and checked element reads. It remains Checker, not an enum numeric-reverse-lookup limitation. |

`TestStage3EnumBoundaries` asserts each diagnostic. `enums_names.a` separately
proves typed enum-object name enumeration, including aliases, reverse-key order
and string enums. The permanent ordering mutant compiles and finishes cleanly,
then fails the independent Node stdout comparison.

The merge retained enum and accessor initialization hooks, enum and nominal
class invariance checks, enum and definite-assignment refusals, and both counts
row sets. [docs/enums.md](../../../docs/enums.md#typescript-source-fixtures)
records validation commands, timings and the six compiler mutants.

## Numeric enum ruling follow-up

Numeric enums are now open numbers, with checked unreachable branches when the
checker narrows them to never. The unchanged fixture set advances from 7 Node
matches, 4 Refused and 2 Checker to 9 Node matches, 1 checked stop, 1 Refused and
2 Checker. Ten fixtures compile. `open-results.json` records fresh observations
for source Node, generated JavaScript and sanitized native, plus successful-run
leak checks. The full oracle also checks native release.

05 and 09 now match Node, including signed masks. 07 passes numeric Map widening
but reaches the existing sparse-array boundary: writing index 1 into an empty
array stops at 70 in both backends. 06 still cannot prove the generic Mutable<T>
write contract. 08 and 11 still fail TS2532 and retain unproven any/index contracts.
Enum name enumeration is supported; it is separate from numeric reverse lookup,
which remains string | undefined. See docs/enums.md for the ruling and mutants.
