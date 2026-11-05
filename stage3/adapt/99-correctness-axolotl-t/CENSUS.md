# 99t compiler census

The ordinary compiler and census driver are built from the final main base a7b6da7f, using the meter's ordinary `go build ./stage3/census/tool` procedure. Every compiler TypeScript root is loaded and lowered independently. Both complete 79-root `adamic types` runs exit 0. No compiler or refusal policy is changed. Raw non-TypeScript and whole-program rows are retained but excluded from the 79-root grouping.

The additional selected-file census uses the unmodified generated meter overlay with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and the two exact filenames in LATENT_ONLY_FILE. Both runs load all 79 roots. Both headers have zero diagnostics and checker_rejected:false. Its historical measurement label says "measured on a checker-rejected program"; the actual header controls that interpretation.

## Ordinary per-file results

| Side | Roots | Kind | First stop | Reason |
| --- | ---: | --- | --- | --- |
| before | 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| before | 1 | NotYet | src/compiler/debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| before | 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| before | 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| before | 73 | Refused | src/compiler/core.ts:726:12 | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold |
| before | 2 | accepted | - | - |
| after | 1 | NotYet | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration |
| after | 1 | NotYet | src/compiler/debug.ts:243:59 | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| after | 1 | NotYet | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration |
| after | 1 | NotYet | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| after | 73 | Refused | src/compiler/core.ts:726:12 | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold |
| after | 2 | accepted | - | - |

## Edited-file census lines

| Side | File | Kind and first stop | Reason |
| --- | --- | --- | --- |
| before | src/compiler/debug.ts | NotYet src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| before | src/compiler/expressionToTypeNode.ts | Refused src/compiler/core.ts:726:12 | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold |
| after | src/compiler/debug.ts | NotYet src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| after | src/compiler/expressionToTypeNode.ts | Refused src/compiler/core.ts:726:12 | a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold |

## Every edited site

The following are actual selected-file refusal/lowering events at the old site span or new assertion span, not a claim that an enclosing function lowers. Coordinates are report data, never edit addresses. Complete findings within each touched owner, including any successor walls outside an assertion, are retained in census-results.json. Full raw streams are retained separately.

| Site | Before | After |
| --- | --- | --- |
| CW144 | debug.ts:813:31; Refused src/compiler/debug.ts:813:32: a cast the runtime can't check | debug.ts:816:21; no finding recorded at this span |
| CW145 | debug.ts:813:86; Refused src/compiler/debug.ts:813:87: a cast the runtime can't check | debug.ts:820:21; no finding recorded at this span |
| CW146 | debug.ts:816:25; Refused src/compiler/debug.ts:816:25: a cast the runtime can't check | debug.ts:831:29; no finding recorded at this span |
| CW147 | debug.ts:817:25; Refused src/compiler/debug.ts:817:25: a cast the runtime can't check | debug.ts:835:29; no finding recorded at this span |
| CW148 | debug.ts:824:38; Refused src/compiler/debug.ts:824:39: a cast the runtime can't check | debug.ts:846:29; no finding recorded at this span |
| CW149 | debug.ts:824:83; Refused src/compiler/debug.ts:824:84: a cast the runtime can't check | debug.ts:850:29; no finding recorded at this span |
| CW150 | debug.ts:828:35; Refused src/compiler/debug.ts:828:36: a cast the runtime can't check | debug.ts:858:21; no finding recorded at this span |
| CW151 | debug.ts:829:7; Refused src/compiler/debug.ts:829:8: a cast the runtime can't check | debug.ts:862:21; no finding recorded at this span |
| CW143 | debug.ts:774:33; Refused src/compiler/debug.ts:774:52: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | debug.ts:775:49; no finding recorded at this span |
| CW154 | debug.ts:942:13; Boundary src/compiler/debug.ts:942:13: failed statement; state rolled back; continued at next statement or declaration; Refused src/compiler/debug.ts:942:56: a cast the runtime can't check | debug.ts:982:13; Boundary src/compiler/debug.ts:982:13: failed statement; state rolled back; continued at next statement or declaration; NotYet src/compiler/debug.ts:982:26: reading hasAntecedentList |
| ArrowFunction or FunctionExpression | expressionToTypeNode.ts:913:17; Refused src/compiler/expressionToTypeNode.ts:913:64: an object refinement using an open numeric enum as a literal tag | expressionToTypeNode.ts:913:17; Refused src/compiler/expressionToTypeNode.ts:913:108: an object refinement using an open numeric enum as a literal tag; Refused src/compiler/expressionToTypeNode.ts:913:56: an object refinement using an open numeric enum as a literal tag; Refused src/compiler/expressionToTypeNode.ts:913:66: an object refinement using an open numeric enum as a literal tag |

The separate ModuleResolutionHost probe is not a proof of callback results. The intentionally retained TypeMapper call and the six individually stopped callers remain in source. Ordinary roots can still stop before any of these edited bodies. The adaptation does not claim that the full compiler now lowers.
