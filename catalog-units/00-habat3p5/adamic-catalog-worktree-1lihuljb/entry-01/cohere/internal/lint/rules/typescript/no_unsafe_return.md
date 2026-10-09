# `@typescript-eslint/no-unsafe-return`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **52** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow returning a value with type `any` from a function

## Why this recommendation

4 of 6 sampled sites are the same @cloudflare/workers-types `Buffer: any` artifact; the genuine signal is the JSON.parse returns in Structure.ts/StructureDoctor.ts, which is a smaller real cleanup than 52 suggests.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

52 in the tree. Showing the first few.

**`libraries/structure/command-line/Structure.ts:1584`**

```
return JSON.parse(output);
```

> Unsafe return of a value of type `any`

**`libraries/structure/command-line/StructureDoctor.ts:31`**

```
return JSON.parse(NodeFileSystem.readFileSync(packageJsonPath, 'utf8'));
```

> Unsafe return of a value of type `any`

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.test.ts:89`**

```
expect(() => Buffer.from(exported, 'base64')).not.toThrow();
```

> Unsafe return of a value of type `any`

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.test.ts:98`**

```
expect(() => Buffer.from(exported, 'base64')).not.toThrow();
```

> Unsafe return of a value of type `any`

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyKeyFactory.test.ts:107`**

```
expect(() => Buffer.from(exported, 'base64')).not.toThrow();
```

> Unsafe return of a value of type `any`

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

