# typescript.d.ts lines moved by 98-axolotl-d

Stock `npm run build` of two trees made by `stage3/apply.sh` on main `470e16f8`, one without and one with
98-axolotl-d, then `diff before/built/local/typescript.d.ts after/built/local/typescript.d.ts`:

```diff
7311c7311
<         packageId?: PackageId;
---
>         packageId?: PackageId | undefined;
7353c7353
<         packageId?: PackageId;
---
>         packageId?: PackageId | undefined;
```

These are the only two lines that differ. Line 7311 is in `interface ResolvedModuleFull extends ResolvedModule`
and line 7353 is in `interface ResolvedTypeReferenceDirective`, both inside `declare namespace ts`.
They are the two rows of [api-sanctions.json](../api-sanctions.json).

| | Bytes | SHA-256 |
| --- | ---: | --- |
| Without 98 | 590,996 | `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d` |
| With 98 | 591,020 | `2ba55ed602cd7fa0f4793d28c8cdd667959570f3eb146de1c7e30932b28bcfbc` |

The 24 added bytes are the two ` | undefined` insertions, 12 bytes each.

In upstream's checked-in baseline, tests/baselines/reference/api/typescript.d.ts at TypeScript 6.0.3, the
same two declarations are lines 7310 and 7352, both `packageId?: PackageId;`, the `before` text of each row.
The one-line offset comes from earlier adaptations, not from this one.

The other four members this adaptation widens are @internal and never reach typescript.d.ts, so they need no
sanction row: `ResolvedModuleFull.originalPath`, `ResolvedTypeReferenceDirective.originalPath`,
`ResolutionWithResolvedFileName.packageId` (program.ts) and `ResolutionWithFailedLookupLocations.files`
(resolutionCache.ts).
