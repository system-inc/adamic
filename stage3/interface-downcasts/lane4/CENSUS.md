# Lane 4 mixed-union census

Measured with stock TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.
Every one of the 2,936 pinned spans and texts matched; the compiler project has
zero TypeScript diagnostics. Each of the three mixed-union family bits was
cross-checked against the original ledger at every site.

Mixed-union dependencies: **2,868 sites**. Remaining only lane 4: **0**.
Remaining lane 4 plus other families: **2,868**. No lane 4 dependency: **68**.
These columns retain the complete input remaining-family sets. No family is
removed based on merged helper availability. No complete lowering is claimed.

| Exact checker union shape | Tagged sites | Untagged sites | Total sites |
| --- | ---: | ---: | ---: |
| `false \| string[] \| undefined` | 1758 | 1105 | 2863 |
| `false \| VersionPaths \| undefined` | 1758 | 1105 | 2863 |
| `string \| false \| undefined` | 1758 | 1105 | 2863 |
| `__String` | 1758 | 1104 | 2862 |
| `__String \| undefined` | 1758 | 1104 | 2862 |
| `DeclarationName \| undefined` | 1758 | 1104 | 2862 |
| `EntityName \| undefined` | 1758 | 1104 | 2862 |
| `FlowNode \| undefined` | 1758 | 1104 | 2862 |
| `ForInitializer \| undefined` | 1758 | 1104 | 2862 |
| `FunctionExpression \| ArrowFunction \| MethodDeclaration \| GetAccessorDeclaration \| SetAccessorDeclaration \| undefined` | 1758 | 1104 | 2862 |
| `HasJSDoc` | 1758 | 1104 | 2862 |
| `HasLocals \| undefined` | 1758 | 1104 | 2862 |
| `Identifier \| JSDocNamespaceDeclaration \| undefined` | 1758 | 1104 | 2862 |
| `Identifier \| PrivateIdentifier \| NumericLiteral \| BigIntLiteral \| JsxNamespacedName \| StringLiteralLike \| undefined` | 1758 | 1104 | 2862 |
| `Identifier \| StringLiteral \| NumericLiteral \| undefined` | 1758 | 1104 | 2862 |

Counts overlap. The type graph is recursive, so nearly every Node target reaches
the same package-json, symbol and flow contracts. This is a transitive dependency
census, not the frequency of reads. Arrays/tuples and callables stop traversal,
as in lane 1's executable census. Dictionary values are traversed and the
dictionary blocker remains. Printed aliases are preserved; expanded member types
and intersection flags appear in census-summary.json. Each shape is counted
once per site, including root unions, not once per field path.

The top five are not five freely reifiable primitive/object shapes. `__String`
is `(string & { __escapedIdentifier: void }) | (void & { __escapedIdentifier: void })`
plus InternalSymbolName literals (upstream types.ts:6196). The original family
classifier puts these intersections in its primitive category `other`; that is
why they appear here as mixed primitive unions. Runtime string tags cannot
certify those phantom intersections. `VersionPaths.paths` is MapLike<string[]>
(upstream moduleNameResolver.ts:422), so the second shape also needs dictionary
support, beyond lane 1 nullish and lane 2 arrays. Retain those blockers.

Reproduce:

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH=/workspace/adamic/stage3/fixtures/assertions/api/node_modules node stage3/interface-downcasts/lane4/census.cjs /tmp/views-mixed-typescript stage3/interface-downcasts/blocking-families-sites.json stage3/interface-downcasts/lane4 > stage3/interface-downcasts/lane4/logs/census.log 2>&1
```

The scratch TypeScript checkout needs its upstream diagnostic metadata generated
with scripts/processDiagnosticMessages.mjs and the pinned assertions/api npm
lockfile installed and linked as node_modules. No cohere source is copied.
