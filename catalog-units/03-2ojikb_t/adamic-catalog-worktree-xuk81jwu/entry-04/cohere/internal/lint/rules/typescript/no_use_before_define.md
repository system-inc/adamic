# `@typescript-eslint/no-use-before-define`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **786** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of variables before they are defined

## Why this recommendation

The sampled sites are helper components defined below the primary component, which was the layout the house rules prescribed when this was audited, and function/component declarations are hoisted. Since 2026-10-01 `react/no-multi-comp` is strict and those helpers move to files of their own, and this count has not been re-measured since.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

786 in the tree. Showing the first few.

**`ProjectRoot.ts:87`**

```
if(directoryIsProjectRoot(currentDirectory)) return currentDirectory;
```

> 'directoryIsProjectRoot' was used before it was defined

**`app/(os-layout)/_components/detail/TaskDetail.tsx:168`**

```
{paneFileDropTarget.isDraggingFileOver ? <TaskDetailDropOverlay /> : null}
```

> 'TaskDetailDropOverlay' was used before it was defined

**`app/(os-layout)/_components/detail/TaskDetailActions.tsx:161`**

```
if(!verdictIsRuleableInline(action)) {
```

> 'verdictIsRuleableInline' was used before it was defined

**`app/(os-layout)/_components/detail/TaskDetailAttachmentPreview.tsx:46`**

```
<TaskDetailAttachmentCaption filename={properties.attachment.filename} sizeText={properties.sizeText} />
```

> 'TaskDetailAttachmentCaption' was used before it was defined

**`app/(os-layout)/_components/detail/TaskDetailAttachmentPreview.tsx:62`**

```
<TaskDetailAttachmentCaption filename={properties.attachment.filename} sizeText={properties.sizeText} />
```

> 'TaskDetailAttachmentCaption' was used before it was defined

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

