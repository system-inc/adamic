Adamic ruled this adaptation on Nov 22 2026 (#bbssdcy, family 38 of the Nov 22 map, family 18 A of the Nov 24 scout pass); Compiler accepts the head after its review and census.

The following declarations are pinned in the adapter's own before snapshot.

| Slot | Declaration | Provider | Reads `this` | Stock strict diagnostics |
| --- | --- | --- | --- | ---: |
| `ModuleSpecifierPreferences.getAllowedEndingsInPreferredOrder` | `src/compiler/moduleSpecifiers.ts:177:5` | Arrow at `src/compiler/moduleSpecifiers.ts:199:44` | No | 0 |
| `EvaluationResolver.evaluateEntityNameExpression` | `src/compiler/types.ts:10652:5` | Shorthand at `src/compiler/checker.ts:1559:9` refers to the closure declared at `src/compiler/checker.ts:47991:5` | No | 0 |
| `EvaluationResolver.evaluateElementAccessExpression` | `src/compiler/types.ts:10653:5` | Shorthand at `src/compiler/checker.ts:1558:9` refers to the closure declared at `src/compiler/checker.ts:48026:5` | No | 0 |

The stock TypeScript reference search includes every TypeScript file under
`src/`, including services and server. No provider flows through `mutateMap`
or the build cache. All three slots also pass together under `--strict`.

The property spelling `name: { name(...): R }["name"]` takes its callable
type from a method signature, keeping bivariant parameters under
`strictFunctionTypes`, as adaptation 32 does. The signatures are erased
during JavaScript emission. See `docs/rulings.md:122` for method values,
`:123` for bound values, `:102` for declared receiver freshness, and
`:406` for changes to the public TypeScript API.

This is the declaration change for the first stops in
`getLocalModuleSpecifier` units 1 to 3 at
`src/compiler/moduleSpecifiers.ts:583:171`, and `createEvaluator` at
`src/compiler/utilities.ts:11352:35`. The bindings still read the callbacks by
name. Loom measures these units and any exposed stops in first-error mode.

Cause B stays in `sys.ts`. Adamic's Nov 24 16:19:55Z ruling, item 16 B,
sends it to the compiler road pending Language. Its test providers include
an arrow reading lexical `this` and a bound value. Cause C stays in the
map callback declarations. Adamic asked to measure it first. Round 1
found nine new mutable-property refusals in `createStateBuildOrder`;
`onDeleteValue` and `onExistingValue` stay as tsc has them.
