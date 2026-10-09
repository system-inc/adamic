# `max-depth`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **176** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum depth that blocks can be nested

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

176 in the tree. Showing the first few.

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:146`**

```
if(componentPrefix) {
```

> Blocks are nested too deeply (5). Maximum allowed is 4

**`libraries/structure/code-quality/lint/rules/BoundaryNoProjectThemeValueRule.ts:148`**

```
if(keys.length > 0) {
```

> Blocks are nested too deeply (6). Maximum allowed is 4

**`libraries/structure/libraries/nexus/code-quality/lint/rules/ConsistencyNoPropertyAliasRule.ts:65`**

```
if(arrayParent.callee.type === 'Identifier' && isHookName(arrayParent.callee.name)) {
```

> Blocks are nested too deeply (5). Maximum allowed is 4

**`libraries/structure/libraries/nexus/code-quality/lint/rules/ConsistencyNoPropertyAliasRule.ts:69`**

```
if(
```

> Blocks are nested too deeply (5). Maximum allowed is 4

**`libraries/structure/code-quality/lint/rules/ConsistencyOrganizeImportsRule.ts:570`**

```
if(comment.type === 'Line') {
```

> Blocks are nested too deeply (5). Maximum allowed is 4

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

