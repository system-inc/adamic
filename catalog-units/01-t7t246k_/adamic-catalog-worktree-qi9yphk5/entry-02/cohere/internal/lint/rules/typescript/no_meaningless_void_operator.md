# `@typescript-eslint/no-meaningless-void-operator`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **11** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow the `void` operator except when used to discard a value

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

11 in the tree. Showing the first few.

**`libraries/structure/source/components/markdown/Markdown.tsx:180`**

```
void button;
```

> void operator shouldn't be used on undefined; it should convey that a return value is being ignored

**`modules/phi/PhiSocialMediaCommandLineInterface.ts:106`**

```
if(!rest[0]) return void console.log('Usage: ahra phi social show <postId>');
```

> void operator shouldn't be used on void; it should convey that a return value is being ignored

**`modules/phi/PhiSocialMediaCommandLineInterface.ts:148`**

```
return void console.log('Usage: ahra phi social upload <postId> --image <path>');
```

> void operator shouldn't be used on void; it should convey that a return value is being ignored

**`modules/phi/PhiSocialMediaCommandLineInterface.ts:159`**

```
return void console.log('Usage: ahra phi social assign <postId> <persona> [--reason "..."]');
```

> void operator shouldn't be used on void; it should convey that a return value is being ignored

**`modules/phi/PhiSocialMediaCommandLineInterface.ts:180`**

```
if(!rest[0]) return void console.log('Usage: ahra phi social untarget <targetId>');
```

> void operator shouldn't be used on void; it should convey that a return value is being ignored

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

