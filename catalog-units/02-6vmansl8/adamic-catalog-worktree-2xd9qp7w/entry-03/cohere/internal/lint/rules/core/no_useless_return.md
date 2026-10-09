# `no-useless-return`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **11** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow redundant return statements

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

11 in the tree. Showing the first few.

**`app/(os-layout)/_components/row/TaskRowTitleEditor.tsx:352`**

```
return;
```

> Unnecessary return statement

**`libraries/structure/source/components/forms/fields/markup/InputMarkup.tsx:297`**

```
return;
```

> Unnecessary return statement

**`modules/os/reports/AhraOsReportCommandLineInterface.ts:340`**

```
return;
```

> Unnecessary return statement

**`modules/os/sensation/AhraOsTriggerCommandLineInterface.ts:607`**

```
return;
```

> Unnecessary return statement

**`modules/os/sensation/AhraOsTriggerCommandLineInterface.ts:769`**

```
return;
```

> Unnecessary return statement

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

