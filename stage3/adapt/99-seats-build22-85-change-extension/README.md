# Family 85: changeExtensionString

Adamic ruled this adaptation's shape on Nov 23 2026 (#bbssdcy, family 85); she accepts the head after its review and census.

In her Nov 23 05:57Z comment in #bbssdcy, Adamic confirmed that the separate
internal helpers with redirected callers are the ruled shape. The original-node
helper carries row 398's two internal overloads. She approved the runtime keys
`changeExtensionString` and `getOriginalNodeWithoutPredicate` on `ts` as internal
exports that the `stripInternal` declaration graph hides. She waived row 398's
SDK probe for family 85. The `stripInternal` emit and the lane's `api_check`
stand in.

It applies rulings 12, 14,
127, 397 and 398 against Adamic 0efbdb083a60086fec3a29687a162bbf21105336.
The public and generic declarations stay byte-identical in source. The twin is
internal. The following positions are in the fully adapted baseline.

| Call | Stock TypeScript 6.0.3 resolved signature | Disposition |
| --- | --- | --- |
| `src/compiler/commandLineParser.ts:4243:46` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/commandLineParser.ts:4274:45` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/emitter.ts:575:12` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/emitter.ts:592:12` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/program.ts:1774:57` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/program.ts:4059:26` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/compiler/transformers/utilities.ts:886:25` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/harness/compilerImpl.ts:111:34` | `(path: string, ext: string): string` | Left (other declaration) |
| `src/harness/compilerImpl.ts:225:16` | `(path: string, ext: string): string` | Left (other declaration) |
| `src/harness/fourslashImpl.ts:4698:27` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/server/project.ts:2657:25` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/services/stringCompletions.ts:891:23` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/services/stringCompletions.ts:905:19` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/services/stringCompletions.ts:1226:33` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/services/stringCompletions.ts:1227:56` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/testRunner/transpileRunner.ts:106:42` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
| `src/testRunner/transpileRunner.ts:112:46` | `(path: string, newExtension: string): string` | Moved (T is exactly string) |
