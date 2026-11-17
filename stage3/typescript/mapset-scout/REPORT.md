# Map/Set scout at 1d3efe7b

## Result

The regenerated first-error entry census confirms **35 distinct NotYet sites**, represented by **37 events**. Every one reaches the **generic depth gate** in `internal/lower/container_fit.go:374–381`. None of the traced refusal endpoints compares an object with a Map slot. The reason text is therefore misleading for this group: it also names a conservative fit-proof boundary, not just a structural fake Map.

The source values partition into **23 named predicate functions (e), 5 arrays/NodeArrays (d), and 7 construction-untraced AST inputs/results**. No site is established as (a) a real Map behind a program interface, (b) a structural Map/Set implementation, or (c) a Map wrapper. The seven untraced origins are explicitly retained; their runtime construction is not guessed.

**Road for these 35: F — keep the current refusal pending a ruling on the bounded generic fit proof.** A representation ruling for real Maps behind program interfaces is a separate road, with **0 established sites in this census**. Converting these arrays or predicate functions into Maps would not preserve their intended behavior. This report proposes no lowering implementation and makes no source adaptation.

## Pins, preparation, and exact stream

Compiler/lowering source: `1d3efe7b92236a0bb94aba6bfd5652f6d0f6dfa6`. The scout branch's intervening commits contain only the previous blocker report/evidence. The recorded meter branch is `meter/daily-20261117T130646Z.rUAmui`, commit `e4b34e0e31c6bc661275ad1467bb9a21db805b20`.

`git submodule update --init --recursive` succeeded: cohere `5360753e19e4417675a90e92a1823bea66b13918`, its checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. This checker submodule is distinct from stage3's adapted-source input. Following `stage3/meter/twice-daily.sh:48–52`, the latter was cloned from `https://github.com/microsoft/TypeScript.git`, tag `v6.0.3`, verified at `050880ce59e30b356b686bd3144efe24f875ebc8`, then supplied as `ADAMIC_TYPESCRIPT_SOURCE`. The meter's lockfile-pinned `npm ci --ignore-scripts --no-audit --no-fund` also succeeded. `stage3/apply.sh /workspace/mapset-adapted` then exited **0**. [Preparation logs and hashes](evidence/toolchain.txt) and [apply log](evidence/apply-retry.log.gz) supersede the prior build blocker.

There is no diagnostic latent stream committed in the recorded run. **The file actually used was the regenerated `/workspace/mapset-first.jsonl`**, produced by the meter's entry overlay at `src/tsc/tsc.ts`, not the compiler-directory pass. Its header says `CommonJSBundle`, `first-error`, checker rejected **false**, **0 diagnostics**, **81 resolved sources**. It returned normally with a `relation_report` trailer and **4,062 timed units** (also the recorded entry timing stream's unit count).

The first-error command was:

```sh
CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
  LATENT_TIMES=/workspace/mapset-first-times.jsonl \
  /workspace/mapset-tools/entry-census \
  /workspace/mapset-adapted/src/tsc/tsc.ts /workspace/mapset-first.jsonl
```

Only findings with the exact reason were retained in [first.findings.jsonl.gz](evidence/first.findings.jsonl.gz). This preserves **all 37 complete messages and unit attributions**. The readable [35-site message appendix](evidence/sites.md) lists each site's full message and every unit. The extra two events are both at `utilities.ts:3095:20`: the same generic body is reached under `getTsConfigPropArrayElementValue`, `getPropertyArrayElementValue`, and `getOptionsSyntaxByValue`. Deduplicating `(kind, where, reason, text)`, as the meter does, gives **35**, not 37. No site was dropped to force the target count.

[Commands](evidence/commands.txt), [census summaries](evidence/census-summary.json), [81 source hashes](evidence/adapted-source-sha256.txt), and the filtered full-mode streams make the observations reviewable. All coordinates below refer to the **adapted** tree; pristine-source coordinates in the brief can differ.

## Why this reason fires

`containerWhereOtherGoes` walks nested fields, type arguments, array elements, and function parameters/results. The reason constant is at `internal/lower/container_fit.go:26` and has multiple return paths:

* At `:340–350`, a target held as `ir.Map` refuses a known non-Map representation (or an unread intersection). This protects against real structural fake Maps.
* At `:374–381`, **depth > 4**, a generic pair refuses if **either type can reach a Map anywhere**. It does not establish an object-to-Map transfer. The observed 35 belong here.

The scratch diagnostic overlay logs the refusing pair and its callers without changing return values. All 35 have a refusing generic pair beyond depth four; **none has `ir.Map` as the endpoint target**. See [per-site fit paths](evidence/fit-path-summary.json), [utilitiesPublic trace](evidence/trace-utilitiesPublic.jsonl.gz), [utilities trace](evidence/trace-utilities.jsonl.gz), and the seven `replay-*.log.txt` files. The two whole-file diagnostic runs reproduced the official filtered findings **exactly**: 29/29 events for utilitiesPublic, 12/12 for utilities. Instrumentation is recorded in [diagnostic-overlay.patch.gz](evidence/diagnostic-overlay.patch.gz); this is measurement evidence, not a proposed lowering change. No repository `internal/` file was edited.

Examples:

* `utilitiesPublic.ts:1116:35`: the value is the function `isJSDocAugmentsTag`, constructed by its declaration at `factory/nodeTests.ts:1096`. The root pair is `(node: Node) => node is JSDocAugmentsTag` versus `(tag: JSDocTag) => tag is JSDocAugmentsTag`. The walk reaches **`EndOfFileToken` → `Node` at depth 5**, where the first type is generic. No Map value is the argument being converted.
* `utilitiesPublic.ts:625:22`: the root value is a union of binding-pattern NodeArrays. The walk reaches **`readonly BindingElement[]` → `readonly (BindingElement | OmittedExpression)[]` at depth 5**, through array predicate callback types; both are `ir.Array`. A Map representation change would not discharge this guard.
* `parser.ts:472:12`: the comparison concerns an AST result returned by `forEach` and its callback; the refusing pair is **`EndOfFileToken` → `Node` at depth 8**. The concrete allocation of the incoming statement node remains untraced.

**G in the table** means the *reachable* library **Map**, for example `SymbolTable = Map<__String, Symbol>` at adapted `types.ts:6216`. It does **not** mean that the site's function, array, or AST object implements Map. A stock TypeScript source-type walk supplies one concrete reachability witness:

```text
Node.emitNode.externalHelpersModuleName.symbol.members
    -> SymbolTable -> library Map<__String, Symbol>
```

The named interfaces at the actual refusing endpoints are AST/array types, not Map, Set, ReadonlyMap, ReadonlySet, PragmaMap, or ReadonlyPragmaMap. The source graph's Map reachability explains why the generic gate can use this reason. No direct Map interface can honestly be assigned to these endpoint comparisons. [Whole site functions and predicate declarations](evidence/source-context.md) and [construction families](evidence/constructions.md) preserve that distinction.

## Site classification

**F** means stays refused for now; owner ruling needed on bounded generic fit, not permission to store these values as `ir.Map`. For the seven bold **untraced** rows, the source expression and fit endpoint are known, but the input/result has not been followed to a concrete construction. Standard AST constructors are context only, not substituted as guessed origins. Array rows name the traced construction family, which can allocate or preserve an incoming array; they do not claim a unique allocation for every invocation.

| Site | Owning unit(s) | Interface | Value class | Construction file:line | Road |
|---|---|---|---|---|---|
| `utilitiesPublic.ts:625:22` | `utilitiesPublic.ts:623:1` (isEmptyBindingPattern) | G: Map, transitive | (d) `node.elements`: binding-pattern NodeArray | factory/nodeFactory.ts:2780 / :2801 → createNodeArray :1169 | F |
| `utilitiesPublic.ts:1035:30` | `utilitiesPublic.ts:1034:1` (getModifiers) | G: Map, transitive | **untraced** — `node: HasModifiers` | **untraced** (input/result; no allocation established) | F |
| `utilitiesPublic.ts:1082:31` | `utilitiesPublic.ts:1080:1` (getJSDocTypeParameterTagsWorker) | G: Map, transitive | **untraced** — `param.parent` | **untraced** (input/result; no allocation established) | F |
| `utilitiesPublic.ts:1111:37` | `utilitiesPublic.ts:1110:1` (hasJSDocParameterTags) | G: Map, transitive | (e) function `isJSDocParameterTag` | factory/nodeTests.ts:1148 | F |
| `utilitiesPublic.ts:1116:35` | `utilitiesPublic.ts:1115:1` (getJSDocAugmentsTag) | G: Map, transitive | (e) function `isJSDocAugmentsTag` | factory/nodeTests.ts:1096 | F |
| `utilitiesPublic.ts:1121:34` | `utilitiesPublic.ts:1120:1` (getJSDocImplementsTags) | G: Map, transitive | (e) function `isJSDocImplementsTag` | factory/nodeTests.ts:1180 | F |
| `utilitiesPublic.ts:1126:35` | `utilitiesPublic.ts:1125:1` (getJSDocClassTag) | G: Map, transitive | (e) function `isJSDocClassTag` | factory/nodeTests.ts:1104 | F |
| `utilitiesPublic.ts:1131:35` | `utilitiesPublic.ts:1130:1` (getJSDocPublicTag) | G: Map, transitive | (e) function `isJSDocPublicTag` | factory/nodeTests.ts:1112 | F |
| `utilitiesPublic.ts:1136:35` | `utilitiesPublic.ts:1135:1` (getJSDocPublicTagNoCache) | G: Map, transitive | (e) function `isJSDocPublicTag` | factory/nodeTests.ts:1112 | F |
| `utilitiesPublic.ts:1141:35` | `utilitiesPublic.ts:1140:1` (getJSDocPrivateTag) | G: Map, transitive | (e) function `isJSDocPrivateTag` | factory/nodeTests.ts:1116 | F |
| `utilitiesPublic.ts:1146:35` | `utilitiesPublic.ts:1145:1` (getJSDocPrivateTagNoCache) | G: Map, transitive | (e) function `isJSDocPrivateTag` | factory/nodeTests.ts:1116 | F |
| `utilitiesPublic.ts:1151:35` | `utilitiesPublic.ts:1150:1` (getJSDocProtectedTag) | G: Map, transitive | (e) function `isJSDocProtectedTag` | factory/nodeTests.ts:1120 | F |
| `utilitiesPublic.ts:1156:35` | `utilitiesPublic.ts:1155:1` (getJSDocProtectedTagNoCache) | G: Map, transitive | (e) function `isJSDocProtectedTag` | factory/nodeTests.ts:1120 | F |
| `utilitiesPublic.ts:1161:35` | `utilitiesPublic.ts:1160:1` (getJSDocReadonlyTag) | G: Map, transitive | (e) function `isJSDocReadonlyTag` | factory/nodeTests.ts:1124 | F |
| `utilitiesPublic.ts:1166:35` | `utilitiesPublic.ts:1165:1` (getJSDocReadonlyTagNoCache) | G: Map, transitive | (e) function `isJSDocReadonlyTag` | factory/nodeTests.ts:1124 | F |
| `utilitiesPublic.ts:1170:35` | `utilitiesPublic.ts:1169:1` (getJSDocOverrideTagNoCache) | G: Map, transitive | (e) function `isJSDocOverrideTag` | factory/nodeTests.ts:1128 | F |
| `utilitiesPublic.ts:1175:35` | `utilitiesPublic.ts:1174:1` (getJSDocDeprecatedTag) | G: Map, transitive | (e) function `isJSDocDeprecatedTag` | factory/nodeTests.ts:1136 | F |
| `utilitiesPublic.ts:1180:35` | `utilitiesPublic.ts:1179:1` (getJSDocDeprecatedTagNoCache) | G: Map, transitive | (e) function `isJSDocDeprecatedTag` | factory/nodeTests.ts:1136 | F |
| `utilitiesPublic.ts:1190:35` | `utilitiesPublic.ts:1189:1` (getJSDocThisTag) | G: Map, transitive | (e) function `isJSDocThisTag` | factory/nodeTests.ts:1156 | F |
| `utilitiesPublic.ts:1195:35` | `utilitiesPublic.ts:1194:1` (getJSDocReturnTag) | G: Map, transitive | (e) function `isJSDocReturnTag` | factory/nodeTests.ts:1152 | F |
| `utilitiesPublic.ts:1200:35` | `utilitiesPublic.ts:1199:1` (getJSDocTemplateTag) | G: Map, transitive | (e) function `isJSDocTemplateTag` | factory/nodeTests.ts:1164 | F |
| `utilitiesPublic.ts:1204:35` | `utilitiesPublic.ts:1203:1` (getJSDocSatisfiesTag) | G: Map, transitive | (e) function `isJSDocSatisfiesTag` | factory/nodeTests.ts:1184 | F |
| `utilitiesPublic.ts:1210:40` | `utilitiesPublic.ts:1208:1` (getJSDocTypeTag) | G: Map, transitive | (e) function `isJSDocTypeTag` | factory/nodeTests.ts:1160 | F |
| `utilitiesPublic.ts:1229:84` | `utilitiesPublic.ts:1228:1` (getJSDocType) | G: Map, transitive | (e) function `isJSDocTypeTag` | factory/nodeTests.ts:1160 | F |
| `utilitiesPublic.ts:1357:28` | `utilitiesPublic.ts:1355:1` (getEffectiveConstraintOfTypeParameter) | G: Map, transitive | **untraced** — `node.parent` | **untraced** (input/result; no allocation established) | F |
| `utilitiesPublic.ts:2638:76` | `utilitiesPublic.ts:2637:1` (hasRestParameter) | G: Map, transitive | (d) `s.parameters`: parameter array | factory/nodeFactory.ts:1875 / :5173 → createNodeArray :1169 | F |
| `utilities.ts:3095:20` | `utilities.ts:3113:1` (getTsConfigPropArrayElementValue)<br>`utilities.ts:12296:1` (getPropertyArrayElementValue)<br>`utilities.ts:12304:1` (getOptionsSyntaxByValue) | G: Map, transitive | (d) `objectLiteral?.properties`: NodeArray | factory/nodeFactory.ts:2871 → createNodeArray :1169 | F |
| `utilities.ts:4429:82` | `utilities.ts:4426:1` (getNamespaceDeclarationNode) | G: Map, transitive | (e) function `isNamespaceImport` | factory/nodeTests.ts:870 | F |
| `utilities.ts:4780:34` | `utilities.ts:4779:1` (getJSDocOverloadTags) | G: Map, transitive | (e) function `isJSDocOverloadTag` | factory/nodeTests.ts:1132 | F |
| `factory/utilities.ts:1614:9` | `factory/utilities.ts:1611:1` (createAccessorPropertyBackingField) | G: Map, transitive | (d) `modifiers`: visited modifier NodeArray | transformers/classFields.ts:940 → visitorPublic.ts:223 / :328 → factory/nodeFactory.ts:1169 | F |
| `parser.ts:472:12` | `parser.ts:469:1` (isFileProbablyExternalModule) | G: Map, transitive | **untraced** — AST node returned by `forEach` / `isAnExternalModuleIndicatorNode` | **untraced** (input/result; no allocation established) | F |
| `parser.ts:496:17` | `parser.ts:495:1` (hasModifierOfKind) | G: Map, transitive | (d) `node.modifiers`: modifier NodeArray | factory/nodeFactory.ts:7137 (asNodeArray) → createNodeArray :1169; parser.ts:2567 | F |
| `transformers/utilities.ts:126:25` | `transformers/utilities.ts:124:1` (containsDefaultReference) | G: Map, transitive | **untraced** — `node`: named import/export bindings | **untraced** (input/result; no allocation established) | F |
| `transformers/utilities.ts:161:25` | `transformers/utilities.ts:153:1` (getImportNeedsImportStarHelper) | G: Map, transitive | **untraced** — `node.importClause.namedBindings` | **untraced** (input/result; no allocation established) | F |
| `transformers/utilities.ts:175:118` | `transformers/utilities.ts:173:1` (getImportNeedsImportDefaultHelper) | G: Map, transitive | **untraced** — `node.importClause.namedBindings` | **untraced** (input/result; no allocation established) | F |

## Construction and class counts

| Runtime/construction class | Distinct sites | Evidence and road |
|---|---:|---|
| (a) Real Map/Set held under a program interface | 0 established | No observed endpoint assigned here. Real-Map representation ruling would be separate. |
| (b) Class/object implementing Map/Set | 0 established | No FakeMap-shaped value established. |
| (c) Wrapper around a real Map | 0 established | No wrapper established. |
| (d) NodeArray/array | 5 | Array construction families traced; F. |
| (e) Other: named predicate function | 23 | Exact function declaration recorded per site; F. |
| Untraced construction: AST input/result | 7 | No runtime allocation guessed; F. |
| **Total** | **35** | **37 events, 35 distinct sites.** |

For arrays, `factory/nodeFactory.ts:1169` creates or preserves ordinary arrays, adds NodeArray metadata, and returns them. Binding patterns assign these at `:2780`/`:2801`; object-literal properties at `:2871`; JSDoc signature parameters at `:5173`. `parser.ts:2567` delegates to this factory. The backing-field modifier path is `classFields.ts:940–941` or `esDecorators.ts:1605–1607` → `visitNodes` (`visitorPublic.ts:223`) → `visitArrayWorker` (`:328`, `[]`/`slice` or original array) → NodeArray wrapping/preservation. These are array operations, not Map wrapper constructors.

The 23 function values resolve through the namespace reexports to the exact `factory/nodeTests.ts` declarations in the table. The supplied bodies test node kinds; the function values themselves are closures. Their generic callback comparisons pull AST types into the fit walk. They are not objects implementing Map's methods.

The historical 23 utilitiesPublic survivors are background, not a substitute for this census: **26 sites** are present in this file at the requested pin, and all 26 now have depth-gate traces. This measures the current refusal mechanism without asserting an unverified one-to-one mapping to the older census.

## Roads and representation ruling boundary

| Road | Count among these 35 | Judgment |
|---|---:|---|
| Source adaptation making the value a real Map/Set, or spelling the slot as the library's own Map | 0 established | No such value/slot transfer was found. Making the actual predicates/arrays into Maps would change JavaScript and their intended behavior. |
| Lowering ruling to hold a proven real Map behind a program interface as `ir.Map` | 0 established | Does not address the observed AST/array depth gate. |
| Permanent structural-implementation refusal (FakeMap-style) | 0 established | The reason string alone does not prove this case. |
| **F: keep current refusal pending bounded-fit ruling** | **35** | 23 function sites + 5 array sites + 7 construction-untraced AST sites. |

The requested three Map-specific roads do not cover the measured mechanism. A bounded-fit ruling would need to decide whether recursive generic AST, callback, and array comparisons can be proved to retain their container storage while preserving rejection of genuine non-Map-to-Map transfers. No code is proposed, and no acceptance claim is made. Narrower source views might be investigated separately, but no source workaround was implemented or validated here.

The **separate real-Map issue is real**. `internal/lower/expression.go:216–227` takes the unbranded type's **own symbol**, requires its declaration to be in a library file, and matches its name. `representation` at `:138–140` consequently chooses `ir.Map` for the library's Map/ReadonlyMap/Set/ReadonlySet; the program interfaces `PragmaMap` and `ReadonlyPragmaMap` do not pass that symbol test and are held as `ir.Object`.

At adapted `types.ts:10440` and `:10452` (pristine `:10416` and `:10428`), the pragma interfaces extend ReadonlyMap and Map. Their populated values are real Maps: `processCommentPragmas` assigns `new Map() as PragmaMap` at adapted `parser.ts:10572` (pristine `:10599`), while JSON parsing assigns `emptyMap as ReadonlyPragmaMap` at adapted `parser.ts:1595`, and `core.ts:19` constructs `emptyMap` with `new Map`. These are **context, not additional members of the 35-site table**.

For that separate case, a source slot using the library's own Map type is a **type-only** road if the checker can retain the necessary dependent pragma typing; that has not been validated here. Replacing a genuinely structural implementation or wrapper with a new Map would change JavaScript (including identity/behavior) and needs its own semantic assessment. A lowering ruling would need consistent **`ir.Map` storage and Map method dispatch** across construction, fields, arguments, returns, and callbacks for values proven to be real Maps. It must continue to distinguish structural implementations: extending Map structurally is not itself proof of a real Map. Such a representation would admit genuine Map-backed program interfaces currently held as objects; it must not restore the plain-object-to-Map-slot miscompilation fixed on main. No lowering code is proposed.

## Full-mode comparison

Each official comparison used the **same entry tool and tsc.ts graph**, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, and **one exact absolute `LATENT_ONLY_FILE`**, with no wildcard or directory selection. Every run exited 0, had zero checker diagnostics, and wrote its completion trailer.

| Exact selected file under `/workspace/mapset-adapted/` | First-error events | First-error distinct sites | Full events | Full distinct sites | Additional distinct sites |
|---|---:|---:|---:|---:|---:|
| `src/compiler/utilitiesPublic.ts` | 26 | 26 | 29 | 29 | 3 |
| `src/compiler/utilities.ts` | 5 | 3 | 12 | 7 | 4 |
| `src/compiler/transformers/utilities.ts` | 3 | 3 | 4 | 4 | 1 |
| **Selected-file total** | **34** | **32** | **45** | **40** | **8** |

All first-error sites in these files are present in the full-mode sets. The additional sites are:

* `utilitiesPublic.ts:1339:38`, `:1342:20`, `:2559:23`.
* `utilities.ts:10816:71`, `:11649:41`, `:11715:72`, `:11861:55`.
* `transformers/utilities.ts:127:17`.

Thus first-error mode misses **8 distinct sites of this reason in the three dominant files**, in addition to suppressing repeated events. This is a measured selected-file comparison, not a whole-program full-mode total. The extra sites are not included in the 35-site runtime-class counts.

## Scope and limitations

The report and evidence are the only tracked changes. No adaptation, ledger, compiler source, or `internal/` file changed. Only the requested adapted tree and no-output measurement/diagnostic tools were built; no implementation was built or proposed.

The existing replay worker initially panicked because its entry reach used a different ordering. The scratch worker was changed only to call the meter overlay's entry-first `LatentEntryReach` method; [the entry-ordering patch (one semantic line, plus gofmt)](evidence/replay-entry-order.patch.gz) and successful replay logs disclose that difference. Official counts always come from the unmodified meter entry tool. This replay adjustment does not touch repository sources.

Seven construction traces remain **untraced**, as marked. The bounded-fit decision and the separate Map-backed-interface representation decision both require owner rulings; this scout changes neither.
