# `@typescript-eslint/no-unsafe-type-assertion`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **2691** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow type assertions that narrow a type

## Why this recommendation

Catches a defect rather than a preference, but the volume means it needs a plan rather than a switch.

## Violations

2691 in the tree. Showing the first few.

**`ProjectRoot.ts:103`**

```
const packageJson = JSON.parse(NodeFileSystem.readFileSync(packageJsonPath, 'utf8')) as { name?: string };
```

> Unsafe assertion from `any` detected: consider using type guards or a safer assertion

**`app/(os-layout)/_components/TasksCenter.tsx:158`**

```
const slug = urlParameters?.slug as string[] | undefined;
```

> Unsafe type assertion: type 'string[] | undefined' is more narrow than the original type

**`app/(os-layout)/_components/TasksCenter.tsx:649`**

```
const target = event.target as HTMLElement | null;
```

> Unsafe type assertion: type 'HTMLElement | null' is more narrow than the original type

**`app/(os-layout)/_components/detail/TaskDetailComments.tsx:68`**

```
setActiveTab(value as TaskDetailCommentsTabType);
```

> Unsafe type assertion: type 'TaskDetailCommentsTabType' is more narrow than the original type

**`app/(os-layout)/_components/detail/TaskDetailDescription.tsx:59`**

```
if((event.target as HTMLElement).closest('a, button')) return;
```

> Unsafe type assertion: type 'HTMLElement' is more narrow than the original type

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

