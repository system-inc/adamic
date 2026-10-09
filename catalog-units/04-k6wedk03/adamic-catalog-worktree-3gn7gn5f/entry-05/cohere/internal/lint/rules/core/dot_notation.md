# `dot-notation`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **121** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce dot notation whenever possible

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

121 in the tree. Showing the first few.

**`libraries/structure/command-line/Structure.ts:1012`**

```
env['NEXT_PUBLIC_API_HOST'] = baseApi.base.port
```

> ["NEXT_PUBLIC_API_HOST"] is better written in dot notation

**`libraries/structure/command-line/Structure.ts:1024`**

```
env['STRUCTURE_DEV_ORIGINS'] = allDevHosts.join(',');
```

> ["STRUCTURE_DEV_ORIGINS"] is better written in dot notation

**`libraries/structure/command-line/Structure.ts:1304`**

```
const pathDirs = (process.env['PATH'] ?? '').split(NodePath.delimiter);
```

> ["PATH"] is better written in dot notation

**`libraries/structure/command-line/Structure.ts:1328`**

```
console.log(`  ${diff['key']}: ${diff['current']} → ${diff['desired']}`);
```

> ["key"] is better written in dot notation

**`libraries/structure/command-line/Structure.ts:1328`**

```
console.log(`  ${diff['key']}: ${diff['current']} → ${diff['desired']}`);
```

> ["current"] is better written in dot notation

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

