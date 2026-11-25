Tail family 69 (#bbssdcy, Adamic's Nov 24 21:50:15Z answer, the third road): `loadModuleFromSpecificNodeModulesDirectory` tests the root package JSON content for `undefined` before `hasProperty(..., "exports")`, in place of `?? emptyArray`. Compiler accepts this head under Adamic's widened split of Nov 24 19:29:49Z, on its value identity, after its review, the hazard walk over tsc's whole program and Loom's unit census.

The stock empty array has no own `exports` key, so the missing-content path answers false in both versions. Present content reaches the same call with the same receiver and lookup order. The stock `readJson` normalizes parsed null to an empty object before caching successful package info. The call's inferred `T` moves from `never[] | PackageJsonPathFields` to `PackageJsonPathFields`, which lets the detached road answer under docs/rulings.md:312 and :316.

Value identity is claimed for tsc's own program and for callers within the public declarations. Two differences are reachable only by a JavaScript caller outside them, and Adamic ruled both outside the contract (#bbssdcy, Nov 25): a caller that defines an own `exports` key on the shared `emptyArray` the API hands out (for example through `getOptionsDiagnostics()`) makes stock answer "has exports" for a missing root package JSON while this version never asks the array; and a caller that puts `{ packageJsonContent: null }` into a resolution cache through `setPackageJsonInfo`, a runtime member the public `PackageJsonInfoCache` omits, gets `false` from stock's `??` where this version's `hasProperty(null, ...)` throws a `TypeError`.

Unit 1, `loadModuleFromExportsOrImports`, still has `T = object` and stays stopped. The whole-program resolver walk removed unit 2's finding at `core.ts:1290:12` and left its two findings for "a union or optional field read in a program with record storage". Unit 1 retains "hasOwnProperty through an object view (an Error may be hidden by the view)".

Before:

```ts
!hasProperty((rootPackageInfo = getPackageJsonInfo(packageDirectory, !nodeModulesDirectoryExists, state))?.contents.packageJsonContent ?? emptyArray, "exports")
```

After adding `let rootPackageJsonContent: PackageJsonPathFields | undefined;`:

```ts
!((rootPackageJsonContent = (rootPackageInfo = getPackageJsonInfo(packageDirectory, !nodeModulesDirectoryExists, state))?.contents.packageJsonContent) !== undefined && hasProperty(rootPackageJsonContent, "exports"))
```

- `adapter-structural-pins` uses the function, the local declaration and the exact own-property call and operand, with each node unique. The test checks the adapter's snapshot refs.
- `adapter-output-test` compares stock resolution with the before and after resolver functions on the same hosts.
- `adapter-table-mentions` is inapplicable because the adapter edits a resolver condition.
