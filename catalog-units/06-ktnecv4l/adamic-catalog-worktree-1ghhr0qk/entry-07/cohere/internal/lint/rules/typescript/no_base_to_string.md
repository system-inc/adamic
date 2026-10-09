# `@typescript-eslint/no-base-to-string`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **53** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Require `.toString()` and `.toLocaleString()` to only be called on objects which provide useful information when stringified

## Why this recommendation

Real signal, but it is explicitly disabled in VerifySettings.json (`typescript/no-base-to-string: off`) and several sites are deliberate `String(value)` fallbacks in generic serializers like Csv.ts/Tsv.ts that already branch on typeof object.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Strong Yes** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

53 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/FinanceChart.tsx:86`**

```
return typeof value === 'number' ? moneyCompact(value, { unit: 'Cents' }) : String(value ?? '');
```

> 'value ?? ''' will use Object's default stringification format ('[object Object]') when stringified

**`app/(os-layout)/os/reports/_components/ReportBlockRenderer.tsx:90`**

```
return value === undefined || value === null ? '' : String(value);
```

> 'value' will use Object's default stringification format ('[object Object]') when stringified

**`libraries/structure/command-line/StructureDoctor.ts:171`**

```
current: packageJson[key] ? (isSimple ? String(packageJson[key]) : '(differs)') : '(missing)',
```

> 'packageJson[key]' will use Object's default stringification format ('[object Object]') when stringified

**`libraries/structure/libraries/nexus/source/errors/BaseErrorSerializer.ts:191`**

```
return createBaseErrorData(`Unknown value caught error: ${String(error)}`);
```

> 'error' may use Object's default stringification format ('[object Object]') when stringified

**`libraries/structure/libraries/nexus/source/structured-text/tables/Csv.ts:14`**

```
const stringValue = typeof value === 'object' ? JSON.stringify(value) : String(value);
```

> 'value' will use Object's default stringification format ('[object Object]') when stringified

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

