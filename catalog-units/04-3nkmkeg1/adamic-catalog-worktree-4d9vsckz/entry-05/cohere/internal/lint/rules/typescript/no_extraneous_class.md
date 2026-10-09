# `@typescript-eslint/no-extraneous-class`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **15** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow classes used as namespaces

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

15 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/http/HttpStatus.ts:177`**

```
export class HttpStatus {
```

> Unexpected class with only static properties

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:23`**

```
class NoParamClass {
```

> Unexpected class with only a constructor

**`modules/asana/AsanaApi.ts:33`**

```
export class AsanaApi {
```

> Unexpected class with only static properties

**`modules/connected/ConnectedAgentApi.ts:49`**

```
export class ConnectedAgentApi {
```

> Unexpected class with only static properties

**`modules/connected/ConnectedCustomerApi.ts:113`**

```
export class ConnectedCustomerApi {
```

> Unexpected class with only static properties

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

