# `@typescript-eslint/naming-convention`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **19173** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce naming conventions for everything across a codebase

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

19173 in the tree. Showing the first few.

**`LintConfiguration.ts:18`**

```
'structure/network-no-direct-fetch': 'off',
```

> Object Literal Property name `structure/network-no-direct-fetch` must match one of the following formats: camelCase

**`LintConfiguration.ts:54`**

```
'nexus/consistency-no-boolean-outcome': [
```

> Object Literal Property name `nexus/consistency-no-boolean-outcome` must match one of the following formats: camelCase

**`LintConfiguration.ts:95`**

```
'nexus/import-require-path-alias': [
```

> Object Literal Property name `nexus/import-require-path-alias` must match one of the following formats: camelCase

**`ProjectRoot.ts:163`**

```
export const ProjectRoot = resolveProjectRoot();
```

> Variable name `ProjectRoot` must match one of the following formats: camelCase, UPPER_CASE

**`ProjectSettings.tsx:11`**

```
export const ProjectSettings: StructureSettingsInterface = {
```

> Variable name `ProjectSettings` must match one of the following formats: camelCase, UPPER_CASE

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

