# CanonicalKey phantom brand

Family 127, Compiler accepts under the Nov 22 split. At the Adamic pin, the fully adapted declaration is at `src/compiler/commandLineParser.ts:4191:1`.

The adapter finds the local `CanonicalKey` alias in the syntax tree and pins its full declaration. It replaces the `never` token of `__canonicalKey` with `undefined`. Ruling [row 29](../../../docs/rulings.md#L29) gives a phantom string brand its primitive representation, [row 30](../../../docs/rulings.md#L30) allows the erased cast, and [row 35](../../../docs/rulings.md#L35) specifies `undefined` and the observation audit. The existing mint keeps the path or uses `toFileNameLowerCase(path)`, then casts it to `CanonicalKey`.

Stock TypeScript 6.0.3 `LanguageService.findReferences` over the fully adapted source returned the following references. The checker excluded `CanonicalKey` from the module exports. The only `__canonicalKey` reference was its definition, with zero reads, writes or presence tests. The shipped and reference declarations contain neither name.

```text
CanonicalKey
  src/compiler/commandLineParser.ts 4151 39 reference const wildCardKeyToPath = new Map<CanonicalKey, string>();
  src/compiler/commandLineParser.ts 4153 30 reference const recursiveKeys: CanonicalKey[] = [];
  src/compiler/commandLineParser.ts 4191 6 definition type CanonicalKey = string & { __canonicalKey: never; };
  src/compiler/commandLineParser.ts 4192 76 reference function toCanonicalKey(path: string, useCaseSensitiveFileNames: boolean): CanonicalKey {
  src/compiler/commandLineParser.ts 4193 78 reference return (useCaseSensitiveFileNames ? path : toFileNameLowerCase(path)) as CanonicalKey;
  src/compiler/commandLineParser.ts 4196 97 reference function getWildcardDirectoryFromSpec(spec: string, useCaseSensitiveFileNames: boolean): { key: CanonicalKey; path: string; flags: WatchDirectoryFlags; } | undefined {
__canonicalKey
  src/compiler/commandLineParser.ts 4191 32 definition type CanonicalKey = string & { __canonicalKey: never; };
```
