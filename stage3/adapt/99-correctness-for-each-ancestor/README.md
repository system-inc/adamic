# forEachAncestorDirectory: string callback contract

Drop the internal Path overload, retain the string overload, and give the implementation the same `<T>(directory: string, callback: (directory: string) => T | undefined): T | undefined` signature. The loop uses `getDirectoryPath(directory)` directly, with no `as P`. This is candidate A from the shape measurements. It preserves the loop, callback bodies and emitted JavaScript bytes; it adds no cast, `any`, callback union or method-bivariance spelling.

The two previously Path-resolved calls do not require the brand:

- `src/compiler/watchUtilities.ts:296`, `clearFirstAncestorEntry` within `createCachedDirectoryStructureHost`: the callback passes `ensureTrailingDirectorySeparator(ancestor)` to `Map<string, MutableFileSystemEntries | false>.delete`. Both operations accept string; the result is `true | undefined`.
- `src/typingsInstallerCore/typingsInstaller.ts:258`, `TypingsInstaller.installPackage`: the callback combines the directory with `package.json`, checks `fileExists`, and returns the directory. Its result, or `projectRootPath`, becomes cwd for `installWorker`, whose cwd parameter is string (line 552). There is no downstream Path requirement.

All five calls therefore use the string overload. The typingsInstaller caller is outside the 79 compiler roots and is checked separately with the extended caller program. The function is absent from TypeScript 6.0.3's public typescript.d.ts and the sanctioned API inventory.

`owners.json` pins the complete top-level owner, including Path Traversal, its documentation, all overloads and the implementation, before and after, by exact LF text and SHA-256. The adapter refuses changed, half-adapted, missing, moved and duplicate owners, verifies the complete planned owner and full-file stock transpile equality before its sole write, rereads the original input before writing, and preserves CRLF. Replay makes zero edits and writes, preserving bytes and mtime. `contract.cjs BEFORE_TREE` exercises these properties and unrelated-owner preservation.

No existing adaptation owns this complete function. Adaptation 30 pins indexed reads elsewhere in path.ts, including getPathFromPathComponents's root read, but no read in this loop. `99-correctness-path-from-components` pins only its own named owner. Neither is amended. Full stage-by-stage apply, rather than replaying an early adapter over every later adaptation, is the integration gate.

Evidence, actual output comparisons, stock/Adamic checks, guarded full-mode censuses, restoring mutant, replay and complete stage3/lane results are recorded on [typescript/for-each-ancestor-evidence](https://github.com/system-inc/adamic/tree/typescript/for-each-ancestor-evidence/stage3/census/for-each-ancestor-adapt). The prior shape measurements are on [typescript/for-each-ancestor](https://github.com/system-inc/adamic/tree/typescript/for-each-ancestor/stage3/census/for-each-ancestor).
