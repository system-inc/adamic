# `no-use-before-define`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **787** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow the use of variables before they are defined

## Why this recommendation

Catches a defect rather than a preference, but the volume means it needs a plan rather than a switch.

## Violations

787 in the tree. Showing the first few.

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

