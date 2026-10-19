The preserved ad029e0c inventory contains 127 direct relation sites in 11 files, from TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. This is the prior pattern inventory, not a new count of structural relations or a claim that tsc compiles under Adamic. The checker loaded 78 inputs and reported the two diagnostics in typescript-impact-ad029.json. The refusal rule was not narrowed.

| Code shape | Sites |
| --- | ---: |
| `void callback passed to forEach` | 57 |
| `void callback passed to forEachChild` | 15 |
| `void callback passed to forEachType` | 4 |
| `void fallback assigned to an any-returning timer field` | 1 |
| `void return into void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>` | 13 |
| `void return into void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram>` | 5 |
| `void return into number \| void \| Symbol` | 32 |

void callback passed to forEach (57 sites):

```ts
// src/compiler/utilities.ts:8579
forEach(objectAllocatorPatchers, fn => fn(objectAllocator));
// src/compiler/factory/nodeFactory.ts:1164
forEach(nodeFactoryPatchers, fn => fn(factory));
```

void callback passed to forEachChild (15 sites):

```ts
// src/compiler/parser.ts:1313
forEachChild(node, addWorkItem, addWorkItem); // By using a stack above and `unshift` here, we emulate a depth-first preorder traversal
// src/compiler/parser.ts:10112
forEachChild(node, visitNode as (node: Node) => void, visitArray as (nodes: NodeArray<Node>) => void);
```

void callback passed to forEachType (4 sites):

```ts
// src/compiler/checker.ts:14832
forEachType(getLowerBoundOfKeyType(constraintType), addMemberForKeyType);
// src/compiler/checker.ts:14838
forEachType(propNameType, t => addMemberForKeyTypeWorker(keyType, t));
```

void fallback assigned to an any-returning timer field (1 site):

```ts
// src/compiler/watch.ts:679
setTimeout: maybeBind(system, system.setTimeout) || noop,
// src/compiler/core.ts:1794, the fallback declaration used at the sole site
export function noop(_?: unknown): void {}
```

This category has one relation site. The second example line is its fallback definition, not a second counted site.

void return into void | WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> (13 sites):

```ts
// src/compiler/executeCommandLine.ts:579
return sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped);
// src/compiler/executeCommandLine.ts:584
return sys.exit(ExitStatus.Success);
```

void return into void | SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> (5 sites):

```ts
// src/compiler/executeCommandLine.ts:833
return sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped);
// src/compiler/executeCommandLine.ts:839
return sys.exit(ExitStatus.Success);
```

void return into number | void | Symbol (32 sites):

```ts
// src/compiler/binder.ts:2871
return checkContextualIdentifier(node as Identifier);
// src/compiler/binder.ts:2882
return checkPrivateIdentifier(node as PrivateIdentifier);
```

All 127 records, including source/target types and grouped source lines, are in tsc-patterns.json. Callback patterns are early-exit traversal APIs that accept return types wider than void. The command-line patterns forward sys.exit() from watch/solution-builder union-returning functions. The binder pattern forwards void helpers from a number/void/Symbol dispatcher. The timer pattern supplies noop where the optional setTimeout field returns any.
