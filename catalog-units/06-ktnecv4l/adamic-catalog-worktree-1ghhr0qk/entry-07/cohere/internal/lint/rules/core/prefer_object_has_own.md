# `prefer-object-has-own`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **8** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow use of `Object.prototype.hasOwnProperty.call()` and prefer use of `Object.hasOwn()`

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

8 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Object.ts:18`**

```
if(Object.prototype.hasOwnProperty.call(object, key)) {
```

> Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'

**`libraries/structure/libraries/nexus/source/collections/Object.ts:108`**

```
if(Object.prototype.hasOwnProperty.call(object, key)) {
```

> Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'

**`libraries/structure/source/utilities/next/Next.ts:22`**

```
return candidateState && Object.prototype.hasOwnProperty.call(candidateState, 'urlPathname');
```

> Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'

**`modules/godword/GodwordCommandLineInterface.ts:172`**

```
firstToken !== undefined && Object.prototype.hasOwnProperty.call(this.commands, firstToken);
```

> Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'

**`modules/kingdom/somfy/KingdomSomfyCommandLineInterface.ts:61`**

```
if(Object.prototype.hasOwnProperty.call(ShadeGroups, target.toLowerCase())) {
```

> Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

