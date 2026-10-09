# `@typescript-eslint/no-unsafe-enum-comparison`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **508** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow comparing an enum value with a non-enum value

## Why this recommendation

The codebase declares zero `enum`s, and every sampled violation compares TSESTree AST_NODE_TYPES string unions in the house lint rules, so all 508 hits are false positives.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

508 in the tree. Showing the first few.

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:113`**

```
if(member.type === 'TSPropertySignature' && member.key) {
```

> The two values in this comparison do not have a shared enum type

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:114`**

```
if(member.key.type === 'Identifier') {
```

> The two values in this comparison do not have a shared enum type

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:117`**

```
else if(member.key.type === 'Literal') {
```

> The two values in this comparison do not have a shared enum type

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:138`**

```
if(node.type === 'ExportNamedDeclaration' && node.declaration?.type === 'TSInterfaceDeclaration') {
```

> The two values in this comparison do not have a shared enum type

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:138`**

```
if(node.type === 'ExportNamedDeclaration' && node.declaration?.type === 'TSInterfaceDeclaration') {
```

> The two values in this comparison do not have a shared enum type

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

