# `func-style`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **8495** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce the consistent use of either `function` declarations or expressions assigned to variables

## Why this recommendation

Stylistic and enormous: the volume says the codebase has a different convention, not that it is wrong.

## Violations

8495 in the tree. Showing the first few.

**`ProjectRoot.ts:69`**

```
function projectRootFromThisFile(): string | undefined {
```

> Expected a function expression

**`ProjectRoot.ts:83`**

```
function projectRootByWalkingUp(startingDirectory: string): string | undefined {
```

> Expected a function expression

**`ProjectRoot.ts:99`**

```
function directoryIsProjectRoot(candidateDirectory: string): boolean {
```

> Expected a function expression

**`ProjectRoot.ts:127`**

```
function resolveProjectRoot(): string {
```

> Expected a function expression

**`ProjectRoot.ts:173`**

```
export function projectPath(...pathSegments: string[]): string {
```

> Expected a function expression

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

