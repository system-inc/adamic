# `object-shorthand`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1736** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or disallow method and property shorthand syntax for object literals

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1736 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:141`**

```
positionFor: function(index: number, count: number): number {
```

> Expected method shorthand

**`app/(os-layout)/_components/TasksCommandMenu.tsx:86`**

```
onSelected: function() {
```

> Expected method shorthand

**`app/(os-layout)/_components/detail/TaskDetail.tsx:100`**

```
onFiles: function(files) {
```

> Expected method shorthand

**`app/(os-layout)/_components/detail/TaskDetailAttachmentItem.tsx:92`**

```
onSelected: function() {
```

> Expected method shorthand

**`app/(os-layout)/_components/detail/TaskDetailAttachmentItem.tsx:101`**

```
onSelected: function() {
```

> Expected method shorthand

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

