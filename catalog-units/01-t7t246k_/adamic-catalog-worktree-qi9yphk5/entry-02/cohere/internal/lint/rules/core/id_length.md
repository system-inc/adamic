# `id-length`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **862** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce minimum and maximum identifier lengths

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

862 in the tree. Showing the first few.

**`ProjectSettings.tsx:81`**

```
x: {
```

> Identifier name 'x' is too short (< 2)

**`app/(os-layout)/_components/TasksCenter.tsx:727`**

```
.sort(function(a, b) {
```

> Identifier name 'a' is too short (< 2)

**`app/(os-layout)/_components/TasksCenter.tsx:727`**

```
.sort(function(a, b) {
```

> Identifier name 'b' is too short (< 2)

**`app/(os-layout)/_components/TasksCenter.tsx:775`**

```
.sort(function(a, b) {
```

> Identifier name 'a' is too short (< 2)

**`app/(os-layout)/_components/TasksCenter.tsx:775`**

```
.sort(function(a, b) {
```

> Identifier name 'b' is too short (< 2)

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

