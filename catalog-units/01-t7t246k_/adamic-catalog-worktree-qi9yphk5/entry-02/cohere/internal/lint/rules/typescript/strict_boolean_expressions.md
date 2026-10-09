# `@typescript-eslint/strict-boolean-expressions`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **6163** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow certain types in boolean expressions

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

6163 in the tree. Showing the first few.

**`ProjectRoot.ts:129`**

```
if(environmentRoot && directoryIsProjectRoot(environmentRoot)) {
```

> Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly

**`ProjectRoot.ts:134`**

```
if(fileAnchoredRoot && directoryIsProjectRoot(fileAnchoredRoot)) return fileAnchoredRoot;
```

> Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly

**`ProjectRoot.ts:136`**

```
if(fileAnchoredRoot) {
```

> Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly

**`ProjectRoot.ts:138`**

```
if(walkedFromFile) return walkedFromFile;
```

> Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly

**`ProjectRoot.ts:142`**

```
if(walkedFromWorkingDirectory) return walkedFromWorkingDirectory;
```

> Unexpected nullable string value in conditional. Please handle the nullish/empty cases explicitly

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

