# `default-case`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **62** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require `default` cases in `switch` statements

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

62 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/activity/OsKingdomActivityEntry.tsx:26`**

```
switch(properties.entry.role) {
```

> Expected a default case

**`app/(os-layout)/os/wisdom/_components/WisdomGateItems.ts:78`**

```
switch(verdict.sourceReceipt.kind) {
```

> Expected a default case

**`app/(os-layout)/os/wisdom/_components/WisdomGateItems.ts:111`**

```
switch(ticket.lifecycle) {
```

> Expected a default case

**`app/(os-layout)/os/wisdom/_components/WisdomShared.ts:165`**

```
switch(verdict.sourceReceipt.kind) {
```

> Expected a default case

**`app/(os-layout)/os/wisdom/_components/WisdomShared.ts:331`**

```
switch(verdict.commitment.kind) {
```

> Expected a default case

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

