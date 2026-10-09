# `@typescript-eslint/no-deprecated`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **54** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow using code marked as `@deprecated`

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

54 in the tree. Showing the first few.

**`app/(os-layout)/_components/row/TaskRowTitleEditor.tsx:258`**

```
else if(documentWithCaretApis.caretRangeFromPoint) {
```

> `caretRangeFromPoint` is deprecated

**`app/(os-layout)/_components/row/TaskRowTitleEditor.tsx:259`**

```
const range = documentWithCaretApis.caretRangeFromPoint(event.clientX, probeY);
```

> `caretRangeFromPoint` is deprecated

**`app/(os-layout)/art/_components/ArtCard.tsx:151`**

```
icon={isPushing ? () => <CircleNotch className="size-4 animate-spin" /> : MonitorPlay}
```

> `CircleNotch` is deprecated. Use CircleNotchIcon

**`app/(os-layout)/art/_components/ArtCard.tsx:151`**

```
icon={isPushing ? () => <CircleNotch className="size-4 animate-spin" /> : MonitorPlay}
```

> `MonitorPlay` is deprecated. Use MonitorPlayIcon

**`app/(os-layout)/art/_components/ArtCardMetadataPopover.tsx:97`**

```
trigger={<Button variant="Outline" size="Icon" icon={Info} title="Show metadata" />}
```

> `Info` is deprecated. Use InfoIcon

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

