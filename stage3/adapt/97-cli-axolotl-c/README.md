# 97c: close() writes undefined only into the watcher field that may be absent

Axolotl (#axolotl), task #qg1ry2t. This adaptation is JavaScript-changing: it adds a check before one write. It changes no public API, since the one declaration it widens is an interface local to `createWatchProgram`'s body. It edits only src/compiler/watchPublic.ts. Adamic ruled the repair on #axochk.

lane: run as part of the Axolotl wave Correctness cuts it into; Loom's run of that wave is the lane's proof.

## The cause

`createWatchProgram` caches each source file in `sourceFilesCache` as a `HostFileInfo`, one of:

- `FilePresentOnHost`: `version: string`, `sourceFile`, and a required `fileWatcher: FileWatcher`.
- `FilePresenceUnknownOnHost`: `version: false` and an optional `fileWatcher?: FileWatcher`.
- `FileMissingOnHost`: `false`.

`close()` clears the cache with `clearMap(sourceFilesCache, value => { if (value && value.fileWatcher) { value.fileWatcher.close(); value.fileWatcher = undefined; } })`. Past `value &&`, `value` is `FilePresentOnHost | FilePresenceUnknownOnHost`. Under exactOptionalPropertyTypes neither arm admits a written `undefined`: one field is required and the other optional without `| undefined`. That is D143, TS2412 at watchPublic.ts:568:17, "Type 'undefined' is not assignable to type 'FileWatcher' with 'exactOptionalPropertyTypes: true'."

## The repair

- The write becomes `if (isFilePresenceUnknownOnHost(value)) value.fileWatcher = undefined;`, the file's own one-line style for a single-statement `if`. `isFilePresenceUnknownOnHost` is the cache's existing predicate in the same body.
- `FilePresenceUnknownOnHost.fileWatcher?: FileWatcher` becomes `fileWatcher?: FileWatcher | undefined` (inventory row I10).
- `FilePresentOnHost.fileWatcher` stays required and untouched. No other declaration moves.

No cast, no assertion. A guard alone doesn't clear the row: with the field declared `fileWatcher?: FileWatcher`, the guarded write still reports the same TS2412 at 568:57. Adamic ruled the guard plus that one widening, and declined `delete` and dropping the write.

## Why it's sound

- **The guard.** `isFilePresenceUnknownOnHost` is true exactly when `version` is a boolean, which is the unknown variant. That includes a present entry whose version `nextSourceFileVersion` set to `false`, which keeps its `sourceFile` and `fileWatcher`. Unknown entries get the same write as before. Present entries no longer get it, so their required field is never written undefined.
- **The present entry keeps a closed watcher, and nothing can see it.** `clearMap` (utilities.ts:8172) runs the callback over every entry and then calls `map.clear()` straight away. `sourceFilesCache` is `createWatchProgram`'s own closure state, and its entries never leave the closure: readers return `sourceFile` or `version`, never the entry. So no code reads a present entry's `fileWatcher` after its watcher is closed.
- **Every watcher is still closed once.** The `close()` call comes before the guard and is untouched.
- **The widening is honest.** It says what close() really stores in an unknown entry, and every other read of that field already tests it for truthiness (watchPublic.ts:774, 780 and 834).

## Proof

- **The shape.** The adapter parses with stock TypeScript 6.0.3 and finds the site by its AST, with no line numbers and no regex:
  - the one `createWatchProgram` with a body;
  - in that body, `FilePresentOnHost`, whose `fileWatcher` must be the required `FileWatcher`;
  - the one `FilePresenceUnknownOnHost` in the file, exactly `version: false` and an optional `fileWatcher` typed `FileWatcher` (before) or `FileWatcher | undefined` (after);
  - `isFilePresenceUnknownOnHost(hostSourceFile: FileMayBePresentOnHost): hostSourceFile is FilePresenceUnknownOnHost`;
  - the file's one `clearMap(sourceFilesCache, …)`, a statement of `close()`. Its callback must be `value => { if (value && value.fileWatcher) { … } }`, whose branch is exactly `value.fileWatcher.close();` followed by the write, spelled `value.fileWatcher = undefined;` (before) or `if (isFilePresenceUnknownOnHost(value)) value.fileWatcher = undefined;` (after);
  - exactly one write of `undefined` into a `fileWatcher` in the whole file.

  The declaration and the write must agree, both before or both after, or the file is refused as partly adapted. The adapter plans both edits, re-reads its result as the adapted shape, checks that only those bytes moved and that every CRLF stayed, and then writes once.
- **Refusals**, each probed on a scratch copy, each exit 1 with no write: the widening without the guard ("partly adapted: the declaration is after, the write before"), the guard without the widening, `isFileMissingOnHost(value)` as the guard, `!isFilePresenceUnknownOnHost(value)`, the guard with braces, `FilePresentOnHost.fileWatcher` widened, `undefined | FileWatcher`, a second write of undefined into a `fileWatcher` (after line 835's close), and a third statement in the callback's branch.
- **Source:** the adapter changes two lines and keeps CRLF endings:
  - 429: `fileWatcher?: FileWatcher;` to `fileWatcher?: FileWatcher | undefined;`
  - 568: `value.fileWatcher = undefined;` to `if (isFilePresenceUnknownOnHost(value)) value.fileWatcher = undefined;`
- **Diagnostics,** from an in-memory whole-`src/compiler` check with Adamic's options (strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess) and stock TypeScript 6.0.3: 141 before and 140 after. The one removed is watchPublic.ts:568:17 TS2412, and nothing new appears anywhere. `adamic types` on Golem, adapted tree at main 7b9c0fe7: whole program 144 to 143, watchPublic.ts own-file 10 to 9, the one removed exactly 568:17, nothing new. A second run edits nothing.
- **JavaScript:** `ts.transpileModule` of watchPublic.ts before and after (ES2020, CommonJS) gives 41,577 and 41,637 bytes. The only difference is the added test: `value.fileWatcher = undefined;` becomes `if (isFilePresenceUnknownOnHost(value))` followed by the same write. The widening is type-only and emits nothing.
- **Idempotence:** a second run edits nothing and leaves the file's SHA-256 unchanged. An input already holding both edits is accepted with no write.
- **Witnesses:** `node witnesses.cjs <watchPublic.ts before> <watchPublic.ts after>` lifts the `clearMap` callback over `sourceFilesCache` out of `close()` in each file, with the cache's two predicates, transpiles them, and runs the callback through stock TypeScript's own `clearMap`. The cache holds four entries: present with a watcher, unknown with a watcher, invalidated (a present entry whose version `nextSourceFileVersion` set to false, `sourceFile` and watcher kept), and missing (`false`). It records each watcher's `close()` count, each entry's state as `map.clear()` starts, and the final size:

  | Entry | close() before / after | `fileWatcher` before clear, before / after |
  | --- | --- | --- |
  | present, `version: "1"` | 1 / 1 | undefined / its watcher |
  | unknown, `version: false` | 1 / 1 | undefined / undefined |
  | invalidated, `version: false` | 1 / 1 | undefined / undefined |
  | missing, `false` | none / none | `false` / `false` |

  The map ends with size 0 both times. The only difference in state is the present entry's `fileWatcher`, and the witness requires exactly that difference.
- **Mutants:**
  - Dropping the guard and keeping the widening is the original JavaScript with the widened type. The checker does not catch it: TypeScript accepts a write to a union once one arm admits `undefined`, so the check reads 140 with 568 clear, the same as the repair. The adapter refuses the shape as partly adapted. Run against it, the witnesses fail, because the present entry's watcher is written undefined again (no differences, where one is required).
  - `!isFilePresenceUnknownOnHost(value)` brings back a diagnostic, TS2322 "Type 'undefined' is not assignable to type 'FileWatcher'." at 568:58, and the witnesses fail: both unknown entries keep their watchers.
  - `isFileMissingOnHost(value)` gives the checker a clean 140, since the predicate narrows `value` to `never`. The witnesses fail: no entry is written.

Not yet run: the sharded lane oracle. It runs in Loom as part of the wave, per Correctness' rule on #axochk.
