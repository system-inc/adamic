# Family 87 named option reads

Accepts: Compiler, under the widened split. The Nov 22 ruling on #bbssdcy approves the named reads, and the Nov 23 answer scopes the receiver audit to tsc's own producers. This applies rows 294 and 305 of `docs/rulings.md` at `cd3d2fbabe64`.

- `moduleNameResolver.ts:1568:11` in `tryLoadModuleUsingPathsIfEligible` holds `state.compilerOptions` once and reads `baseUrl`, then `paths`.
- `moduleNameResolver.ts:1657:11` in `tryLoadModuleUsingBaseUrl` holds the same field once and reads `baseUrl`.

Both helpers keep the original local names, the same optional values, and the identity of the `paths` record. The adapter checks that `heldCompilerOptions` has no identifier occurrence in the source file before editing. The CLI's normalized options and parsed config options store data properties. Ordinary named property reads take the existing checked read path.

`test.cjs` runs the adapter against its saved predecessor and checks that populated paths resolve with the original record identity, along with optional values, read order, and the direct reapplication guard. From the repository root, run:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-87-named-option-reads/test.cjs /path/to/cold-adapted-tree
```
