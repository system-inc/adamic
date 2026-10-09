# `react/no-multi-comp`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **86** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow multiple component definition per file

## Why this recommendation

The codebase is actively moving toward one component per file per the recent 'each file holds one component' commits.

Kirk ruled on 2026-10-01 that this rule is strict and stays on: one component per file. The in-house size-banded rule it was first audited beside, which fired only past 60 lines, was retired and deleted rather than switched off, and `structure/consistency-require-matching-file-name` was added as the other half of the same promise: once a file holds one component, the file is named for it.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

86 in the tree. Showing the first few.

**`app/(os-layout)/_components/detail/TaskDetail.tsx:186`**

```
function TaskDetailDropOverlay() {
```

> Declare only one React component per file

**`app/(os-layout)/_components/detail/TaskDetailAttachmentPreview.tsx:111`**

```
function TaskDetailAttachmentCaption(properties: TaskDetailAttachmentCaptionProperties) {
```

> Declare only one React component per file

**`app/(os-layout)/_components/detail/TaskDetailCommentCards.tsx:95`**

```
function TaskDetailDrillInCard(properties: { comment: TaskCommentInterface; label: string; tone: 'Cyan' | 'Neutral' }) {
```

> Declare only one React component per file

**`app/(os-layout)/_components/detail/TaskDetailCommentCards.tsx:134`**

```
function TaskDetailRulingChipCard(properties: { comment: TaskCommentInterface }) {
```

> Declare only one React component per file

**`app/(os-layout)/_components/detail/TaskDetailCommentItems.tsx:75`**

```
function TaskDetailPlainComment(properties: { comment: TaskCommentInterface }) {
```

> Declare only one React component per file

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

