# Void-intersection read census — H2 widening (#jvw7qww)

The single freshly adapted main tree contains **1,400 read sites across 27 files**.
All 1,400 acquire their void intersection from the phantom `__String` brand's
`void & { __escapedIdentifier: void }` union arm. There are **0 other origins**
and **0 direct reads of brand members** such as `value.__escapedIdentifier`.
This is a large surface to take to Adamic before widening the rule.

| Tree | Read sites | Distinct files | Evidence / coverage |
|---|---:|---:|---|
| Main: `6dd21a991c8b69aa7b2e0e0346385b42a59c8653` | **1,400** | **27** | Fresh measurement; all adaptations through `92-tty-host-results` |
| Retype: `typescript/adapt40-brands-undefined` (`07ca4739f3204bdc4cfac437de7f478da83e73f1`) | 1,400 (previous worker's report) | 27 (previous worker's report) | Not rebuilt or independently measured here, per revised step 2; reported coverage through adaptation 40 |

The retype row preserves the supplied prior observation; no retype result artifact
or fresh comparison is claimed. The only measured rows here are `results-main.jsonl.gz`.

## Top shapes (exact checker type text)

The six most frequent type texts cover 1,357 sites; the remaining 14 texts cover
43 sites. Import-qualified spellings remain separate because grouping is by exact
`typeToString` output. Each row below includes five examples.

| Type text | Sites | Five examples (`file:line:column`, expression; role) |
|---|---:|---|
| `__String` | 1111 | `src/compiler/binder.ts:632:34` — `name`; identifier; argument<br>`src/compiler/binder.ts:670:24` — `(isGlobalScopeAugmentation(node as ModuleDeclaration) ? "__global" : &#96;"${moduleN`; AsExpression; return value<br>`src/compiler/binder.ts:676:28` — `escapeLeadingUnderscores(nameExpression.text)`; call result; return value<br>`src/compiler/binder.ts:679:28` — `tokenToString(nameExpression.operator) + nameExpression.operand.text as __String`; AsExpression; return value<br>`src/compiler/binder.ts:693:24` — `getSymbolNameForPrivateIdentifier(containingClassSymbol, name.escapedText)`; call result; return value |
| `__String &#124; undefined` | 164 | `src/compiler/binder.ts:698:20` — `isPropertyNameLiteral(name) ? getEscapedTextOfIdentifierOrLiteral(name) : undefi`; ConditionalExpression; return value<br>`src/compiler/binder.ts:738:126` — `getDeclarationName(node)`; call result; argument<br>`src/compiler/binder.ts:755:22` — `isComputedName ? InternalSymbolName.Computed : isDefaultExport && parent ? Inter`; ConditionalExpression; initializer<br>`src/compiler/binder.ts:756:15` — `isDefaultExport && parent ? InternalSymbolName.Default : getDeclarationName(node`; ConditionalExpression<br>`src/compiler/binder.ts:757:15` — `getDeclarationName(node)`; call result |
| `(string & { __escapedIdentifier: void; }) &#124; (void & { __escapedIdentifier: void; }) &#124; InternalSymbolName` | 31 | `src/compiler/binder.ts:738:107` — `Debug.checkDefined(getDeclarationName(node))`; call result; argument<br>`src/compiler/checker.ts:3241:22` — `isString(nameArg) ? nameArg : (nameArg as Identifier).escapedText`; ConditionalExpression; initializer<br>`src/compiler/checker.ts:3246:73` — `name`; identifier; argument<br>`src/compiler/checker.ts:3248:80` — `name`; identifier; argument<br>`src/compiler/checker.ts:3249:82` — `name`; identifier; argument |
| `import("./types.js").__String` | 28 | `src/compiler/commandLineParser.ts:2514:84` — `getTextOfPropertyName(element.name)`; call result<br>`src/compiler/commandLineParser.ts:3597:182` — `getTextOfPropertyName(rootCompilerOptions[0]!)`; call result<br>`src/compiler/expressionToTypeNode.ts:343:77` — `node.typeName.escapedText`; property access result; operand of ===<br>`src/compiler/expressionToTypeNode.ts:366:63` — `p.name.escapedText`; property access result; operand of ===<br>`src/compiler/expressionToTypeNode.ts:610:58` — `p.name.escapedText`; property access result; operand of === |
| `Identifier &#124; __String` | 12 | `src/compiler/checker.ts:3241:31` — `nameArg`; identifier; argument<br>`src/compiler/checker.ts:3246:79` — `nameArg`; identifier; argument<br>`src/compiler/checker.ts:3257:21` — `nameArg`; identifier; condition<br>`src/compiler/checker.ts:3274:33` — `nameArg`; identifier; operand of &&<br>`src/compiler/checker.ts:3278:95` — `nameArg`; identifier; argument |
| `import("../types.js").__String` | 11 | `src/compiler/factory/utilities.ts:219:29` — `jsxFactory.right.escapedText`; property access result; operand of =<br>`src/compiler/factory/utilities.ts:1573:38` — `node.escapedText`; property access result<br>`src/compiler/transformers/declarations.ts:1471:68` — `p.escapedName`; property access result; argument<br>`src/compiler/transformers/declarations.ts:1666:75` — `input.name.escapedText`; property access result; argument<br>`src/compiler/transformers/jsx.ts:275:83` — `elem.name.escapedText`; property access result; operand of === |

## Method and reproducibility

- Stock TypeScript **6.0.3**, required from the lockfile-pinned `stage3/api` install.
- Source pin: TypeScript `v6.0.3`, `050880ce59e30b356b686bd3144efe24f875ebc8`.
- `stage3/apply.sh` from the main commit above built one complete adapted tree.
  Its product key was `7044f75a637960aedc26a704a05ad044f5055601db565853132d87aa1072d47a`.
  No compiler or adaptation code was changed. The requested setup completed successfully.
- `stage3/census/tool/main.go` calls `load.Load`; its options come from
  `internal/load/load.go:compilerOptions`. The script parses the adapted
  `src/compiler/tsconfig.json` for roots and inherited configuration, then overrides
  the census settings: strict (including strict bind/call/apply and unknown catch
  variables), unchecked indexed access, exact optional properties, verbatim modules,
  ESNext/Bundler, forced modules, ES2024 and its library, no ambient types,
  importing TS extensions, no emit, and erasable syntax disabled.
- It visits expression nodes in every program source under `src/compiler/**/*.ts`,
  obtains `getTypeAtLocation`, and looks through unions for an intersection with
  a constituent carrying `TypeFlags.Void`. Plain void alone does not qualify.
  Declaration names, property-name tokens, and type-only syntax are not reads;
  shorthand property values are included.
- It excludes expression-statement discards, `void` operands, assignment left sides,
  and callees of calls whose results are discarded. Parentheses are transparent
  when determining those contexts. Parenthesized expression wrappers are queried
  but do not count a second time: `refuseVoidRead` explicitly skips them. Counting
  those two wrappers as separate expression nodes gives 1,402 rather than 1,400.
- Other expression wrappers and nested expression nodes remain separate sites,
  so this is an expression-site count, not a count of statements or runtime executions.
  This follows the requested type-and-syntax inventory; it is not a lowering run or
  an assertion that all sites are reachable before other refusals. No meter refusal
  was added. The current exact-void brand-member exception was not changed.
- Each JSONL row records location, expression text capped at 80 characters, untruncated
  checker type text, syntactic role, direct brand-member declarations, and matching
  intersection origins. All matching origins are the same branded void arm.

```bash
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
git fetch origin main typescript/adapt40-brands-undefined
# Use the recorded main commit's stage3 files; later main may have different adaptations.
npm ci --ignore-scripts --no-audit --no-fund --prefix stage3/api > /tmp/void-npm.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/void-upstream > /tmp/void-upstream.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/void-upstream bash stage3/apply.sh /tmp/void-main-adapted > /tmp/void-apply.log 2>&1
node stage3/census/void-intersections/count.cjs /tmp/void-main-adapted stage3/census/void-intersections/results-main.jsonl.gz > /tmp/void-count.log 2>&1
```

Validation: Node syntax check; gzip/JSON decoding; 1,400 unique location/expression
rows; type-group totals and 27-file total; every origin checked against the branded
void arm. The initial 1,402-row checkpoint differs only by the two transparent
parenthesis wrappers. Rows and script were committed and pushed before this README;
the corrected rows were also pushed before this README was written.
