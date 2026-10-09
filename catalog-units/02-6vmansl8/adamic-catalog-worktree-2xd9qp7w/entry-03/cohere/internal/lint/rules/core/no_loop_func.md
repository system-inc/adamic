# `no-loop-func`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **10** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow function declarations that contain unsafe references inside loop statements

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

10 in the tree. Showing the first few.

**`libraries/structure/source/components/tables/columns/kinds/choice/hooks/useChoiceCounts.ts:46`**

```
function(count) {
```

> Function declared in a loop contains unsafe references to variable(s) 'isActive'

**`libraries/structure/source/components/tables/columns/kinds/choice/hooks/useChoiceCounts.ts:53`**

```
function(error: unknown) {
```

> Function declared in a loop contains unsafe references to variable(s) 'isActive'

**`modules/asana/AsanaImport.ts:602`**

```
onCreate: function() {
```

> Function declared in a loop contains unsafe references to variable(s) 'projectsCreated'

**`modules/data/DataRebuild.ts:795`**

```
onStatement(statementSql) {
```

> Function declared in a loop contains unsafe references to variable(s) 'creationsRewritten'

**`modules/data/DataRebuild.ts:802`**

```
onHeader(header, table) {
```

> Function declared in a loop contains unsafe references to variable(s) 'currentHeader', 'statementsApplied'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

