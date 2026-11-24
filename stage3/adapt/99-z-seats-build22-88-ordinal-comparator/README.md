# Family 88 ordinal comparator

Adamic's Nov 22 2026 build ruling in #bbssdcy governs these sorts. The
Nov 23 confirmation approved the directory order and phantom brand. These
are the pinned statements after the preceding TypeScript 6.0.3 adapters at
`6b3b321ff4361ff230fa3c91e51e0d05318a5106`.

- `sys.ts:1886:17` in unit `sys.ts:1451:1` uses `files.sort()` on `string` elements.
- `sys.ts:1887:17` in unit `sys.ts:1451:1` uses `directories.sort()` on `string` elements.
- `watchUtilities.ts:156:51` in unit `watchUtilities.ts:118:1` uses `entries.files.map(getCanonicalFileName).sort()` on `Canonicalized` elements.
- `watchUtilities.ts:157:57` in unit `watchUtilities.ts:118:1` uses `entries.directories.map(getCanonicalFileName).sort()` on `Canonicalized` elements.

The local file and directory arrays start empty as `string[]`. Entries
are pushed before sorting, so each sort operates in place on a dense array.
The watch sorts operate on fresh arrays returned by `map(getCanonicalFileName)`.
The callback is cast to `(name: string) => Canonicalized`, where
`type Canonicalized = string & { __canonicalized: void; };`. The phantom
brand preserves the runtime string value. All four arrays have no
`undefined` element arm.

Each call keeps its receiver and passes
`(a, b) => a < b ? -1 : a > b ? 1 : 0`. Node's default sort applies
ToString and compares UTF-16 code units. ToString preserves these strings,
and relational comparison between strings compares the same code units
(ruling 334). Equal strings are indistinguishable, so stability is
unobservable here (ruling 335). Rulings 322 and 496 record Adamic's refusal
of the comparator-free form.

This directory runs after `99-wasm-crypto-import` and `99-z-family20`.
The crypto adapter verifies its whole-function pin before either `sys.ts` sort changes.

The nearest function-like ancestor must be the named function declaration.
Each local receiver resolves directly in the listing try block to
`const files: string[] = [];` or `const directories: string[] = [];`
at sys.ts:1854 and :1855. The watch binding at watchUtilities.ts:149 is
`const entries = getCachedFileSystemEntries(getDirectoryPath(path));`.

The callback cast at watchUtilities.ts:124 must call the factory imported
at :9 from `"./_namespaces/ts.js"` (:61), using the host parameter
`useCaseSensitiveFileNames: boolean` (:118). TypeScript binds explicit
source and library roots with `noResolve` and `noLib`. The import's alias
must resolve through the namespace export to this pinned core.ts:2384
declaration.

```ts
export function createGetCanonicalFileName(useCaseSensitiveFileNames: boolean): GetCanonicalFileName {
    return useCaseSensitiveFileNames ? identity : toFileNameLowerCase;
}
```

Its resolved type at core.ts:2382 is
`export type GetCanonicalFileName = (fileName: string) => string;`. The
factory returns this string callback before the cast. The callback and
cache casts must use the exact phantom alias at watchUtilities.ts:101.

The host's `function getCachedFileSystemEntries(rootDirPath: Path)`
(watchUtilities.ts:144) reads the pinned
`Map<string, MutableFileSystemEntries | false>` at :123. Its inferred
result is `MutableFileSystemEntries | false | undefined`. After the
existing truthiness guard, the fields resolve to
`readonly files: string[];` and `readonly directories: string[];` in
the interface at :103. The method resolves to stock `Array.map`.
All four sites are checked before either source file is written.

Run the adapter test on an output of the full Stage 3 apply.

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-z-seats-build22-88-ordinal-comparator/test.cjs <fully-adapted-tree>
```

The before snapshot supplies both inputs and the factory's supporting
sources. The after snapshot supplies this adapter's recorded output.
Clean, repeat and partial applies must match the after snapshot. The
clean output must also equal the before bytes with exactly four
comparator insertions, so an earlier adapter's unrelated edit can stay
in both snapshots.

The rejection rows cover scope, brand, binding, same-text shadow,
duplicate, parse error, statement drift, second-file drift, numeric
producer, nested numeric producer, namespace factory, the core return,
the watch entries and callback shadows, a non-string map member and a
shadowed cache lookup. The identity check refuses a factory declared in
`_namespaces/ts.ts` that shadows core's re-export.
Rejections leave both file hashes unchanged. Stock TypeScript
transpiles the original and adapted statements for the VM's order and
array identity checks. The watch cases supply identity and
`name => name.toLowerCase()` callbacks.
