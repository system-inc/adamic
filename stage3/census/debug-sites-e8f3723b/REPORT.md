# debug.ts latent site census

Task b5w3ycg. Source/compiler/tool/apply SHA: `e8f3723bd41abcebedabfcc7546fdf104a9bbf08`, the fetched origin/main tip. Both `b4b8c4d7` and `f4e8a4c6` are ancestors, each verified with `git merge-base --is-ancestor` exit 0. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Run and limitation

Setup exit 0 (990.636 seconds). `npm ci --prefix stage3/api` exit 0. One `stage3/apply.sh` tree, exit 0 in 1168.566 seconds. Guarded overlay build exit 0 in 641.462 seconds. The same completed tree and binary were used for the measurement.

Full mode: child exit 0, **50.090 seconds**, **105 units**, zero checker diagnostics. `LATENT_FULL=1`, `LATENT_ONLY_FILE=/workspace/scratch/debug-sites/adapted/src/compiler/debug.ts`, `LATENT_ASSERT_NO_OUTPUT=1`. All other inherited LATENT flags were removed; only LATENT_TIMES was added. Ordinary Load and Lower output guards ran in the actual driver. The raw legacy label `measured on a checker-rejected program` is preserved, despite zero checker diagnostics.

**First-error was not run.** The pinned tool ignores LATENT_ONLY_FILE in first-error mode: `stage3/census/latent/lower.go.txt:266` iterates every loaded file without `latentFileSelected`; full mode checks that guard in `stage3/census/latent/units.go.txt:119`. The pinned driver loads the compiler directory and its imports. Launching it unchanged would violate debug.ts alone. A question requesting approval for the same selection guard in the scratch first-error overlay remained unanswered while the authorized work completed. No selection patch was applied. Thus no first-error elapsed time, measured first-error stop, or claimed first-error unit count is supplied. This report stops that sub-run at the tool limitation, rather than representing the full run as first-error.

The first-error source loop treats namespace Debug as one top-level statement; full mode splits nested function declarations. Full mode was used to supply the requested complete site inventory independently of the blocked first-error run. This is the same guarded overlay generation used by the meter and reference core report, with physical generated Go filenames changed only to `.go.txt`. No production source or tool logic was patched.

The post-apply checkpoint is `1b136f633792e4502b285a3c49d63b0d8a11055b` on `typescript/debug-sites-checkpoint`. The final report branch contains a single commit directly on the pinned main SHA.

## Forecast sites

“Gone” below means no finding at that forecast expression in the full-mode stream, including the complete refusal scan. It does not prove each enclosing function lowers: independent generic bodies are deferred and other findings may stop lowering. All forecast coordinates still identify the expected expressions in this adapted tree.

| Forecast class | Adapted coordinates | Status | Current reason |
| --- | --- | --- | --- |
| 1: assertEachIsDefined marker | 243:59 | Gone | No finding at this expression. |
| 2: first hasProperty(member, ...) | 253:66 | Gone | No finding at this expression. |
| 3: second hasProperty(member, ...) | 253:97 | Gone | No finding at this expression. |
| 4: getFunctionName(test!) in assertEachNode | 267:72 | Present, Refused, refusal_scan | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read |
| 4: getFunctionName(test!) in assertNode | 280:98 | Present, Refused, refusal_scan | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read |
| 4: getFunctionName(test!) in assertNotNode | 293:108 | Present, Refused, refusal_scan | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read |
| 4: getFunctionName(test!) in assertOptionalNode | 307:98 | Present, Refused, refusal_scan | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read |
| 5: assertEachNode marker and repeats | 268:35, 281:35, 294:35, 308:35, 322:35 | Gone at every listed expression | No finding at these expressions. |
| checkEachDefined marker repeat | line 248 (stackCrawlMark &#124;&#124; checkEachDefined) | Gone | No finding at this expression. |

The four class-4 stops remain the first refusal-scan wall. Additional lowering findings include the `never` parameter of assertNever and overload implementation compatibility checks before those bodies. Clearing the forecast marker and hasProperty scans therefore does not imply that these declarations now lower completely.

## Every debug.ts site in source order

Repeated observations at the same location, kind, exact reason and phase share an F ID. There are 53 such debug.ts sites. The raw stream has 112 findings, including repeated observations, Boundary continuation records, and findings reached in dependencies. `rows.json` retains the observations and recording-unit identities.

Full mode rolls back failed statements and resumes; downstream “reading” findings may reflect missing locals after rollback. These are observed census findings, not a proven sequence of ordinary compiler stops. Uninstantiated generic units are not proof of successful lowering.

| Site | Location | Containing declaration | Kind | Exact reason | Phase | Emitting lowering line and deciding function |
| --- | --- | --- | --- | --- | --- | --- |
| F1 | debug.ts:252:33 | assertNever | NotYet | a value of type never | lowering | internal/lower/expression.go:53 `(l *lowering) typeOf` |
| F2 | debug.ts:257:86 | assertEachNode | Refused | overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | lowering | internal/lower/census_small.go:233 `(l *lowering) censusOverload` |
| F3 | debug.ts:267:72 | assertEachNode | Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | refusal_scan | internal/lower/invariance.go:1039 `(l *lowering) wideningRefusal`; internal/lower/invariance.go:1053 `(l *lowering) wideningRefusal` |
| F4 | debug.ts:273:82 | assertNode | Refused | overload 1 of assertNode parameter test cannot be served by implementation parameter test | lowering | internal/lower/census_small.go:233 `(l *lowering) censusOverload` |
| F5 | debug.ts:280:98 | assertNode | Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | refusal_scan | internal/lower/invariance.go:1039 `(l *lowering) wideningRefusal`; internal/lower/invariance.go:1053 `(l *lowering) wideningRefusal` |
| F6 | debug.ts:293:108 | assertNotNode | Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | refusal_scan | internal/lower/invariance.go:1039 `(l *lowering) wideningRefusal`; internal/lower/invariance.go:1053 `(l *lowering) wideningRefusal` |
| F7 | debug.ts:299:78 | assertOptionalNode | Refused | overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test | lowering | internal/lower/census_small.go:233 `(l *lowering) censusOverload` |
| F8 | debug.ts:307:98 | assertOptionalNode | Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | refusal_scan | internal/lower/invariance.go:1039 `(l *lowering) wideningRefusal`; internal/lower/invariance.go:1053 `(l *lowering) wideningRefusal` |
| F9 | debug.ts:352:20 | getFunctionName | Refused | inherited library member name read as an own field | lowering | internal/lower/prototype.go:124 `(l *lowering) prototypeRead` |
| F10 | debug.ts:355:26 | getFunctionName | Refused | inherited library member call read as an own field | lowering | internal/lower/prototype.go:124 `(l *lowering) prototypeRead` |
| F11 | debug.ts:355:26 | getFunctionName | Refused | a method read as a value (toString would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go:282 `(l *lowering) methodValueRefusal` |
| F12 | debug.ts:356:60 | getFunctionName | NotYet | reading text | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F13 | debug.ts:357:20 | getFunctionName | NotYet | reading match | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F14 | debug.ts:362:135 | formatSymbol | NotYet | a template interpolating an object, an array, a map, a function or undefined | lowering | internal/lower/expression.go:1217 `(l *lowering) template` |
| F15 | debug.ts:412:27 | getEnumMembers | NotYet | a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) | lowering | internal/lower/reflect_get.go:111 `(l *lowering) checkDynamicKeyedGetterSyntax` |
| F16 | debug.ts:413:24 | getEnumMembers | NotYet | reading value | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F17 | debug.ts:496:15 | attachFlowNodeDebugInfoWorker | Refused | in that narrows a union of object types (any object with the key passes) | lowering | internal/lower/unknown.go:46 `(l *lowering) inProperty` |
| F18 | debug.ts:537:37 | attachFlowNodeDebugInfo | Refused | a cast the runtime can't check | lowering | internal/lower/cast_proof.go:191 `(l *lowering) castProof` |
| F19 | debug.ts:537:37 | attachFlowNodeDebugInfo | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:191 `(l *lowering) castProof` |
| F20 | debug.ts:540:17 | attachFlowNodeDebugInfo | Refused | Object.setPrototypeOf | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:26 `(l *lowering) objectCallArguments` |
| F21 | debug.ts:553:15 | attachNodeArrayDebugInfoWorker | NotYet | in when a possibly absent spread has no property presence descriptors | lowering | internal/lower/unknown.go:56 `(l *lowering) inProperty`; internal/lower/unknown.go:314 `(l *lowering) presenceHazard` |
| F22 | debug.ts:577:38 | attachNodeArrayDebugInfo | Refused | a cast the runtime can't check | lowering | internal/lower/cast_proof.go:191 `(l *lowering) castProof` |
| F23 | debug.ts:577:38 | attachNodeArrayDebugInfo | Refused | a cast the runtime can't check | refusal_scan | internal/lower/cast_proof.go:191 `(l *lowering) castProof` |
| F24 | debug.ts:580:17 | attachNodeArrayDebugInfo | Refused | Object.setPrototypeOf | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:26 `(l *lowering) objectCallArguments` |
| F25 | debug.ts:596:33 | enableDebugInfo | NotYet | new an Identifier | lowering | internal/lower/object.go:1822 `(l *lowering) newExpression` |
| F26 | debug.ts:597:33 | enableDebugInfo | NotYet | new an Identifier | lowering | internal/lower/object.go:1822 `(l *lowering) newExpression` |
| F27 | debug.ts:600:9 | enableDebugInfo | Refused | Object.defineProperties | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:24 `(l *lowering) objectCallArguments` |
| F28 | debug.ts:617:9 | enableDebugInfo | Refused | Object.defineProperties | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:24 `(l *lowering) objectCallArguments` |
| F29 | debug.ts:671:9 | enableDebugInfo | Refused | Object.defineProperties | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:24 `(l *lowering) objectCallArguments` |
| F30 | debug.ts:774:70 | enableDebugInfo | Refused | an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | refusal_scan | internal/lower/invariance.go:1035 `(l *lowering) wideningRefusal` |
| F31 | debug.ts:874:20 | attachDebugPrototypeIfDebug | Refused | Object.setPrototypeOf | lowering | internal/lower/library_object.go:20 `(l *lowering) objectCallArguments`; internal/lower/library_object.go:26 `(l *lowering) objectCallArguments` |
| F32 | debug.ts:959:15 | formatControlFlowGraph | NotYet | a value of type Record<number, FlowGraphNode> | lowering | internal/lower/expression.go:53 `(l *lowering) typeOf` |
| F33 | debug.ts:973:9 | isFlowSwitchClause | NotYet | a checked predicate target whose check would read past the depth bound (FlowSwitchClause) | lowering | internal/lower/predicates_proof.go:949 `(l *lowering) predicateMembership`; internal/lower/predicates_proof.go:988 `(l *lowering) checkedPredicateTargetCause` |
| F34 | debug.ts:980:39 | hasAntecedents | NotYet | in when a possibly absent spread has no property presence descriptors | lowering | internal/lower/unknown.go:56 `(l *lowering) inProperty`; internal/lower/unknown.go:314 `(l *lowering) presenceHazard` |
| F35 | debug.ts:981:32 | hasAntecedents | NotYet | reading hasAntecedentList | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F36 | debug.ts:982:26 | hasAntecedents | NotYet | reading hasAntecedentList | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F37 | debug.ts:983:20 | hasAntecedents | NotYet | reading antecedent | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F38 | debug.ts:1016:29 | buildGraphNode | NotYet | reading links | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F39 | debug.ts:1017:17 | buildGraphNode | NotYet | reading graphNode | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F40 | debug.ts:1033:18 | buildGraphNode | NotYet | reading graphNode | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F41 | debug.ts:1034:29 | buildGraphNode | Refused | a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | refusal_scan | internal/lower/invariance.go:1039 `(l *lowering) wideningRefusal`; internal/lower/invariance.go:1053 `(l *lowering) wideningRefusal` |
| F42 | debug.ts:1046:20 | buildGraphNode | NotYet | reading graphNode | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F43 | debug.ts:1065:20 | computeLevel | NotYet | returning a property or element assignment | lowering | internal/lower/statements.go:337 `(l *lowering) returnAssignment` |
| F44 | debug.ts:1077:39 | computeColumnWidths | NotYet | a member typed any ([]) related to number at the initializer site | refusal_scan | internal/lower/any_relations.go:43 `(l *lowering) refuseAnyRelation` |
| F45 | debug.ts:1079:17 | computeColumnWidths | NotYet | reading columns | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F46 | debug.ts:1081:20 | computeColumnWidths | NotYet | reading columns | lowering | internal/lower/expression.go:672 `(l *lowering) enumNeverValue` |
| F47 | debug.ts:1152:82 | renderGraph | NotYet | Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | lowering | internal/lower/library_globals.go:29 `(l *lowering) libraryGlobalValue` |
| F48 | debug.ts:1153:48 | renderGraph | Refused | an unproven value assigned to a numeric literal or enum member slot Connection | refusal_scan | internal/lower/invariance.go:1035 `(l *lowering) wideningRefusal` |
| F49 | debug.ts:1165:21 | renderGraph | NotYet | assigning to a NonNullExpression | lowering | internal/lower/assignments.go:65 `(l *lowering) assignment` |
| F50 | debug.ts:1168:21 | renderGraph | NotYet | assigning to a NonNullExpression | lowering | internal/lower/assignments.go:65 `(l *lowering) assignment` |
| F51 | debug.ts:1176:21 | renderGraph | NotYet | assigning to a NonNullExpression | lowering | internal/lower/assignments.go:65 `(l *lowering) assignment` |
| F52 | debug.ts:1219:17 | writeLane | NotYet | assigning to a NonNullExpression | lowering | internal/lower/assignments.go:65 `(l *lowering) assignment` |
| F53 | debug.ts:1251:38 | fill | NotYet | a value of type T | lowering | internal/lower/expression.go:53 `(l *lowering) typeOf` |

## First-wall list

Exactly the refusal-scan sites, in source order, as requested. This deliberately excludes lowering findings and is not the broader criterion used by the reference core report. Full mode records every scan stop; priority and traversal order can affect the ordinary first stop. There are 11.

1. F3, `debug.ts:267:72`, Refused: a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read
2. F5, `debug.ts:280:98`, Refused: a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read
3. F6, `debug.ts:293:108`, Refused: a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read
4. F8, `debug.ts:307:98`, Refused: a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read
5. F11, `debug.ts:355:26`, Refused: a method read as a value (toString would lose its object, and this with it)
6. F19, `debug.ts:537:37`, Refused: a cast the runtime can't check
7. F23, `debug.ts:577:38`, Refused: a cast the runtime can't check
8. F30, `debug.ts:774:70`, Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile
9. F41, `debug.ts:1034:29`, Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read
10. F44, `debug.ts:1077:39`, NotYet: a member typed any ([]) related to number at the initializer site
11. F48, `debug.ts:1153:48`, Refused: an unproven value assigned to a numeric literal or enum member slot Connection

## Reached dependency stops

These are not sites physically in debug.ts. They are retained separately in rows.json and the unmodified raw stream.

- `core.ts:1277:12`, NotYet, lowering: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow)
- `core.ts:1049:82`, NotYet, lowering: a comparator that isn't a function
- `core.ts:1276:47`, NotYet, lowering: a value of type T

## Whole adapted units containing class 4

The four independent full-mode implementation units are reproduced whole below. The enclosing namespace Debug is also a full-mode unit and would be one top-level first-error unit; its entire adapted text, plus these four units, is in [class4-units.md](class4-units.md). The complete original-byte adapted file is [debug.ts.txt](debug.ts.txt). Overload signature units contain no class-4 expression.

### assertEachNode, debug.ts:262:5

```typescript
export function assertEachNode(nodes: readonly Node[] | undefined, test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction) {
        if (shouldAssertFunction(AssertionLevel.Normal, "assertEachNode")) {
            assert(
                test === undefined || every(nodes, test),
                message || "Unexpected node.",
                () => `Node array did not pass test '${getFunctionName(test!)}'.`,
                stackCrawlMark || assertEachNode,
            );
        }
    }
```

### assertNode, debug.ts:275:5

```typescript
export function assertNode(node: Node | undefined, test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction) {
        if (shouldAssertFunction(AssertionLevel.Normal, "assertNode")) {
            assert(
                node !== undefined && (test === undefined || test(node)),
                message || "Unexpected node.",
                () => `Node ${formatSyntaxKind(node?.kind)} did not pass test '${getFunctionName(test!)}'.`,
                stackCrawlMark || assertNode,
            );
        }
    }
```

### assertNotNode, debug.ts:288:5

```typescript
export function assertNotNode(node: Node | undefined, test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction) {
        if (shouldAssertFunction(AssertionLevel.Normal, "assertNotNode")) {
            assert(
                node === undefined || test === undefined || !test(node),
                message || "Unexpected node.",
                () => `Node ${formatSyntaxKind(node!.kind)} should not have passed test '${getFunctionName(test!)}'.`,
                stackCrawlMark || assertNotNode,
            );
        }
    }
```

### assertOptionalNode, debug.ts:302:5

```typescript
export function assertOptionalNode(node: Node | undefined, test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction) {
        if (shouldAssertFunction(AssertionLevel.Normal, "assertOptionalNode")) {
            assert(
                test === undefined || node === undefined || test(node),
                message || "Unexpected node.",
                () => `Node ${formatSyntaxKind(node?.kind)} did not pass test '${getFunctionName(test!)}'.`,
                stackCrawlMark || assertOptionalNode,
            );
        }
    }
```

## Evidence and reproduction

- [full.jsonl](full.jsonl): complete raw stream, unmodified. [full-times.jsonl](full-times.jsonl): per-unit timing stream. [full-result.json](full-result.json): child exit and elapsed time.
- [rows.json](rows.json): sites, recording-unit identities, full record and reached dependency sites. [emitters.json](emitters.json): exact compiler source text and enclosing function declarations.
- [run.py](run.py): exact commands, flags, logging, no-output watchdog and process-group cleanup. [overlay.tar.gz](overlay.tar.gz): exact built scratch overlay.
- [provenance.json](provenance.json), [adapted-source-manifest.json](adapted-source-manifest.json), [adapted-product.manifest.json](adapted-product.manifest.json), and [validation.json](validation.json) pin inputs and confirm unchanged source.
- Logs and result files retain setup, dependency installation, apply and build evidence. `SHA256SUMS` covers the final evidence files.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1  # child exit recorded separately
npm ci --prefix stage3/api > npm.log 2>&1  # child exit recorded separately
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/debug-sites/TypeScript
python3 stage3/census/debug-sites-e8f3723b/run.py apply > apply-wrapper.log 2>&1
python3 stage3/census/debug-sites-e8f3723b/run.py build > build-wrapper.log 2>&1
python3 stage3/census/debug-sites-e8f3723b/run.py full > full-wrapper.log 2>&1
node stage3/census/debug-sites-e8f3723b/ast-map.cjs > ast-map.log 2>&1
python3 stage3/census/debug-sites-e8f3723b/analyze.py > analysis.log 2>&1
python3 stage3/census/debug-sites-e8f3723b/map-emitters.py > emitter-map.log 2>&1
```

Apply requires a new output directory. The recorded invocation ran once. Reproduction uses the exact pinned source revision; current main may have moved. All processes launched for this task completed; no census child was left running.
