# `one-var`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **24698** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce variables to be declared either together or separately in functions

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

24698 in the tree. Showing the first few.

**`ProjectRoot.ts:133`**

```
const fileAnchoredRoot = projectRootFromThisFile();
```

> Combine this with the previous 'const' statement

**`ProjectRoot.ts:141`**

```
const walkedFromWorkingDirectory = projectRootByWalkingUp(process.cwd());
```

> Combine this with the previous 'const' statement

**`ProjectRoot.ts:163`**

```
export const ProjectRoot = resolveProjectRoot();
```

> Combine this with the previous 'const' statement

**`app/(os-layout)/AhraChatView.tsx:39`**

```
const dyadMember =
```

> Combine this with the previous 'const' statement

**`app/(os-layout)/AhraChatView.tsx:41`**

```
const dyadProfileId = dyadMember?.profileId ?? null;
```

> Combine this with the previous 'const' statement

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

