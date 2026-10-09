# `no-useless-concat`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **4** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unnecessary concatenation of literals or template literals

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

4 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/security/cryptography/CryptographyObjectOperations.test.ts:26`**

```
'{"settings":{"notifications":true,"theme":"dark"},' + '"user":{"age":30,"name":"John"}}',
```

> Unexpected string concatenation of literals

**`libraries/structure/source/api/graphql/GraphQlOperationsMetadataPlugin.ts:582`**

```
values: [${type.values.map((value) => '\n    ' + `'${value}'`).join(',')}\n  ],
```

> Unexpected string concatenation of literals

**`modules/kingdom/control4/KingdomControl4MediaCommandLineInterface.ts:307`**

```
'Usage: ahra kingdom play <room> <spotify-name>\n' + '  Example: ahra kingdom play office kirk',
```

> Unexpected string concatenation of literals

**`modules/planetscale/PlanetScaleApi.ts:579`**

```
`pscale database restore-dump ${database} ${toBranch}${orgFlag}` + ` --dir ${tmpDirectory} --data-only`,
```

> Unexpected string concatenation of literals

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

