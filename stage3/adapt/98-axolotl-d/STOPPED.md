# 98-axolotl-d: stopped before any edit

This directory holds no adapter. Work stopped at row 2, the ruling for
moduleNameResolver 286:9 and 627:9, because the checker shows that the ruled
widening neither closes 286 nor leaves the rest of the program alone. Rows 1, 3
and 4 were not edited, measured, mutated or witnessed. Nothing under
stage3/lane, stage3/fixtures or stage3/meter was touched.

## What was run

- Fresh main `7b9c0fe7e52cfe9cdeb774107389bd6870f41423`, `bash cloud/setup.sh --wasi-sdk`
  (exit 0), stock TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
- `stage3/apply.sh /tmp/ax/before` passed end to end.
- The census tool built from this main, run on `before/src/compiler`, and
  stage3/meter/report.py's own `report()` over its output:
  [own-file-before.json](proofs/own-file-before.json). It agrees with the
  November 1 meter: moduleNameResolver 3, resolutionCache 1, sourcemap 1.
- `adamic types` rooted at moduleNameResolver.ts on the applied tree and on
  two scratch copies made by [probe.py](proofs/probe.py). That script edits by
  exact text and is a probe only; it does not meet the adapter contract. Rerun
  from scratch, it reproduced the three committed diagnostic sets line for line.

## Row 2 against the checker

The ruling widens four members to `?: (T) | undefined`: `ResolvedModuleFull.originalPath`
and `.packageId`, and `ResolvedTypeReferenceDirective.originalPath` and `.packageId`.

Each of the four is needed. Widening only the two originalPath members leaves
both diagnostics naming packageId; widening only the two packageId members
leaves both naming originalPath.

With all four widened ([diagnostics-ruled-four.txt](proofs/diagnostics-ruled-four.txt),
[ruled-four.diff](proofs/ruled-four.diff)):

- 627:9 closes.
- 286:9 stays open and now names `isExternalLibraryImport`
  ([elaboration-ruled-four.txt](proofs/elaboration-ruled-four.txt)). That member
  is declared on the public `ResolvedModule` (types.ts line 8031 in the applied
  tree), which `ResolvedModuleFull` extends, and the literal at 286 stores
  `isExternalLibraryImport: boolean | undefined` from the function's parameter.
- Nine diagnostics appear that were not there: program.ts 2183 and 2206, and
  resolutionCache.ts 777, 1005, 1034, 1053, 1056, 1471 and 1472. Each is the
  widened `ResolvedModuleFull` or `ResolvedTypeReferenceDirective` no longer
  assignable to the @internal `ResolutionWithResolvedFileName` (program.ts line
  1021), whose `packageId?: PackageId` is not widened.

Widening those two as well, six members in all
([diagnostics-six.txt](proofs/diagnostics-six.txt), [six.diff](proofs/six.diff)),
closes 286 and 627 and the program's diagnostic diff is only those two removals.
This was observed on one root's whole program, not the per-file census, so it
is not yet an own-file table. It is also not yet a proof that 286 stores an
undefined `isExternalLibraryImport`; that needs the callers of
createResolvedModuleWithFailedLookupLocations.

## Why this is a stop

The declared API baseline difference for row 2 is exactly the two public
packageId lines. Closing row 2 needs a third public line,
`ResolvedModule.isExternalLibraryImport`, which no ruling covers. Shipping the
four ruled members would leave moduleNameResolver at 2 and add nine diagnostics
in two other files. Shipping six would move a public line nobody admitted.

`ResolutionWithResolvedFileName.packageId` is @internal, so under 20's model it
needs proof that undefined is stored, not a ruling; the originalPath members and
`ResolutionWithFailedLookupLocations.files` are @internal too and would need no
api-sanctions.json row.

By generate-api.cjs's `ownerPath`, which drops the `ts` namespace, sanction rows
would be named `["ResolvedModuleFull", "packageId"]`,
`["ResolvedTypeReferenceDirective", "packageId"]` and, if admitted,
`["ResolvedModule", "isExternalLibraryImport"]`.

## What would let it go on

Either a ruling admitting `ResolvedModule.isExternalLibraryImport` as a third
public line, after which the head does all five rows with the full proofs, or a
narrower head of rows 1, 3 and 4 only (moduleNameResolver 3 to 2, no public API
change) under a different commit message.
