# `no-div-regex`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **1** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow equal signs explicitly at the beginning of regular expressions

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

1 in the tree. 

**`modules/google/email/EmailApi.ts:455`**

```
return Buffer.from(message, 'utf-8').toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
```

> A regular expression literal can be confused with '/='

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

