# `react/jsx-props-no-spreading`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **234** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow JSX prop spreading

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

234 in the tree. Showing the first few.

**`app/(os-layout)/_components/detail/TaskDetail.tsx:144`**

```
{...paneFileDropTarget.dropTargetProperties}
```

> Prop spreading is forbidden

**`app/(os-layout)/_components/kingdom/OsKingdomGraph.tsx:98`**

```
<OsKingdomGraphCanvas {...properties} />
```

> Prop spreading is forbidden

**`app/(os-layout)/_components/markdown/FlatMarkdown.tsx:39`**

```
return <p className="mt-3 text-sm/[22px] font-semibold first:mt-0" {...withoutNode(properties)} />;
```

> Prop spreading is forbidden

**`app/(os-layout)/_components/markdown/FlatMarkdown.tsx:46`**

```
<p className="mt-2 text-sm/[22px] first:mt-0" {...withoutNode(properties)} />
```

> Prop spreading is forbidden

**`app/(os-layout)/_components/markdown/FlatMarkdown.tsx:49`**

```
<strong className="font-semibold" {...withoutNode(properties)} />
```

> Prop spreading is forbidden

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

