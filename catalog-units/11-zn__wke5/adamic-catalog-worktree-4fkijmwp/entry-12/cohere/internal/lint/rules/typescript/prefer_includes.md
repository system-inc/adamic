# `@typescript-eslint/prefer-includes`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **30** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce `includes` method over `indexOf` method

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

30 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/http/user-agent/UserAgent.ts:115`**

```
const hasWebKit = /AppleWebKit\//.test(userAgent);
```

> Use `String#includes()` method with a string instead

**`libraries/structure/libraries/nexus/source/protocols/http/user-agent/UserAgent.ts:146`**

```
else if(/Android/.test(raw)) {
```

> Use `String#includes()` method with a string instead

**`libraries/structure/libraries/nexus/source/protocols/http/user-agent/UserAgentParser.ts:31`**

```
if(/iPhone/.test(userAgent)) {
```

> Use `String#includes()` method with a string instead

**`libraries/structure/libraries/nexus/source/protocols/http/user-agent/UserAgentParser.ts:34`**

```
else if(/iPad/.test(userAgent)) {
```

> Use `String#includes()` method with a string instead

**`libraries/structure/libraries/nexus/source/protocols/http/user-agent/UserAgentParser.ts:37`**

```
else if(/Android/.test(userAgent) && !/Mobile/.test(userAgent)) {
```

> Use `String#includes()` method with a string instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

