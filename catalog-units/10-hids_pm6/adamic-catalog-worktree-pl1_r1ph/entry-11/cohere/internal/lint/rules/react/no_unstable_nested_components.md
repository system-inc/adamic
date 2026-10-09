# `react/no-unstable-nested-components`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **18** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow creating unstable components inside components

## Why this recommendation

Components defined during render remount their whole subtree and lose state every render, which is a real bug; the 18 hits include concrete cases like Sidebar.tsx:294 and ArtCard.tsx:82.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

18 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomImageCard.tsx:106`**

```
isSettingActive ? () => <CircleNotchIcon className="size-4 animate-spin" /> : undefined
```

> Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “OsKingdomImageCard” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true

**`app/(os-layout)/_layout/sidebar/Sidebar.tsx:294`**

```
icon={function(iconProperties) {
```

> Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “Sidebar” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true

**`app/(os-layout)/art/_components/ArtCard.tsx:151`**

```
icon={isPushing ? () => <CircleNotch className="size-4 animate-spin" /> : MonitorPlay}
```

> Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ArtCard” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true

**`app/(os-layout)/art/_components/ArtCarousel.tsx:186`**

```
? () => <CircleNotch className="size-4 animate-spin" />
```

> Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ArtCarousel” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true

**`app/(os-layout)/finance/_components/FinanceLedgerTable.tsx:58`**

```
cell: (context: LedgerCellContextInterface) => (
```

> Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “FinanceLedgerTable” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

