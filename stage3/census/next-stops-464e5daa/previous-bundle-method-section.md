, tsc measured as the CommonJS bundle it ships as

[rulings.md:221](../../docs/rulings.md#L221) binds the CommonJS module values
only when the build declares that format. The censuses had lowered tsc as an
ES program, although it ships as the single CommonJS bundle `_tsc.js`. This is
a measurement-method change, not a compiler improvement. Both readings below
use the same main, `fb2b54da12a02e1550b9f5bf046d3314c77bf657`, and adapted input.
Locations and root names are relative to `src/compiler/`.

| Per-root first stop | Before: EcmaScript | After: CommonJSBundle |
| --- | ---: | ---: |
| Refused, `sys.ts:57:1`: ambient function declaration | 73 | 0 |
| NotYet, `core.ts:2599:15`: process host row measured only in ES mode (#redxes2) | 0 | 73 |
| NotYet, `performanceCore.ts:29:15`: performance host row measured only in ES mode (#redxes2) | 0 | 1 |
| Refused, `scanner.ts:469:12`: method read as a value (`sys.ts` root) | 1 | 0 |
| NotYet, `semver.ts:71:9`: namespace read before initialization (`debug.ts` root) | 1 | 1 |
| NotYet, `commandLineParser.ts:293:35`: enum read before initialization (`diagnosticInformationMap.generated.ts` root) | 1 | 1 |
| NotYet, `diagnosticInformationMap.generated.ts:13:45`: enum read before initialization (`types.ts` root) | 1 | 1 |
| Accepted: files lowering whole | 2 | 2 |

74 of 79 roots move: of the 73 formerly at `sys.ts:57:1`, 72 move to
`core.ts:2599:15`, while `core.ts`'s own root moves to
`performanceCore.ts:29:15`. The `sys.ts` root moves from `scanner.ts:469:12`
to `core.ts:2599:15`, bringing the after group there to 73. Files lowering
whole are unchanged: **`corePublic.ts` and `hostErrors.ts`, 2/79 before and
after**. No native-output success is claimed.

The complete before/after root table, raw observations and sys.ts latent pair
are on branch `typescript/census-commonjs-bundle-evidence`, commit
[`3dd24184b`](https://github.com/system-inc/adamic/tree/3dd24184b9150872269ff035bd0f3814397ac437/stage3/census/commonjs-bundle-evidence)
([report](https://github.com/system-inc/adamic/blob/3dd24184b9150872269ff035bd0f3814397ac437/stage3/census/commonjs-bundle-evidence/README.md)).
