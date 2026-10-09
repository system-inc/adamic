# `complexity`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **326** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce a maximum cyclomatic complexity allowed in a program

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

326 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:149`**

```
export function TasksCenter() {
```

> Function 'TasksCenter' has a complexity of 73. Maximum allowed is 20

**`app/(os-layout)/_components/drag/TasksDragContext.tsx:334`**

```
function(point: { x: number; y: number }) {
```

> Function has a complexity of 25. Maximum allowed is 20

**`app/(os-layout)/_components/kingdom/OsKingdomView.tsx:110`**

```
export function OsKingdomView(properties: OsKingdomViewProperties) {
```

> Function 'OsKingdomView' has a complexity of 25. Maximum allowed is 20

**`app/(os-layout)/_components/messages/MessageCard.tsx:66`**

```
export function MessageCard(properties: MessageCardProperties) {
```

> Function 'MessageCard' has a complexity of 21. Maximum allowed is 20

**`app/(os-layout)/_components/row/TaskAssigneePicker.tsx:44`**

```
export function TaskAssigneePicker(properties: TaskAssigneePickerProperties) {
```

> Function 'TaskAssigneePicker' has a complexity of 24. Maximum allowed is 20

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

