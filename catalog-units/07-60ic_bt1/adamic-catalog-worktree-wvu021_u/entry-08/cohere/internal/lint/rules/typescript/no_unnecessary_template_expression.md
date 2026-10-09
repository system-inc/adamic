# `@typescript-eslint/no-unnecessary-template-expression`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **48** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow unnecessary template expressions

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

48 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:341`**

```
const scopeLabel = scopedToCurrent && projectId !== null ? `${activeProjectTitle ?? 'this project'}` : 'all tasks';
```

> Template literal expression is unnecessary and can be simplified

**`app/(os-layout)/os/wisdom/_components/WisdomGateItems.ts:138`**

```
text: `${work.status.toLowerCase() === 'done' ? 'fix landed' : `fix ${work.status.toLowerCase()}`}`,
```

> Template literal expression is unnecessary and can be simplified

**`libraries/structure/libraries/nexus/source/protocols/http/HttpMethod.ts:15`**

```
export type HttpMethodType = `${HttpMethodKindType}`;
```

> Template literal expression is unnecessary and can be simplified

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:192`**

```
[`PREFIX_${'SUFFIX'}`]: 'value',
```

> Template literal expression is unnecessary and can be simplified

**`libraries/structure/libraries/nexus/source/version-control/Git.ts:130`**

```
`${failure.stderr || 'no stderr'}`,
```

> Template literal expression is unnecessary and can be simplified

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

