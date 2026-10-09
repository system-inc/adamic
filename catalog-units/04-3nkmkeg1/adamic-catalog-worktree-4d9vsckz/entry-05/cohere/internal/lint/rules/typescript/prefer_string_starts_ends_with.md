# `@typescript-eslint/prefer-string-starts-ends-with`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **7** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce using `String#startsWith` and `String#endsWith` over other equivalent methods of checking substrings

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

7 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/ip/IpAddress.ts:56`**

```
if(/^10\./.test(ipAddress)) {
```

> Use 'String#startsWith' method instead

**`libraries/structure/libraries/nexus/source/protocols/ip/IpAddress.ts:66`**

```
if(/^192\.168\./.test(ipAddress)) {
```

> Use 'String#startsWith' method instead

**`libraries/structure/libraries/nexus/source/protocols/ip/IpAddress.ts:71`**

```
if(/^127\./.test(ipAddress)) {
```

> Use 'String#startsWith' method instead

**`libraries/structure/libraries/nexus/source/protocols/ip/IpAddress.ts:76`**

```
if(/^169\.254\./.test(ipAddress)) {
```

> Use 'String#startsWith' method instead

**`modules/pensieve/PensieveSoulCommandLineInterface.ts:86`**

```
if(/^```/.test(line)) {
```

> Use 'String#startsWith' method instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

