# `no-await-in-loop`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **363** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow `await` inside of loops

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

363 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksNewTaskDialog.tsx:102`**

```
await uploadAttachment(created.id, file);
```

> Unexpected `await` inside a loop

**`app/(os-layout)/_hooks/useTaskAttachmentUploadRequest.ts:34`**

```
const response = await networkService.request(
```

> Unexpected `await` inside a loop

**`app/(os-layout)/os/sensation/_components/SensationsDashboard.tsx:144`**

```
await postObservationAction(observationId, action);
```

> Unexpected `await` inside a loop

**`baselines/DumpEslintConfig.ts:41`**

```
resolvedConfigurations[targetFile] = await esLint.calculateConfigForFile(targetFile);
```

> Unexpected `await` inside a loop

**`libraries/structure/libraries/nexus/source/coordination/BackoffTask.ts:68`**

```
const result = await this.task();
```

> Unexpected `await` inside a loop

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

