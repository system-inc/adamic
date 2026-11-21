# Readonly family 3 views, third round

Adamic task #k8ed1f2, following Performance's decision of Nov 21 at
03:59Z. These changes address wider views whose consumers need less authority
than their types express.

## Slots

| Cause | File and owner | Type |
|---|---|---|
| 6 | builder.ts, `clearAffectedFilesPendingEmit.state` | `Omit<BuilderProgramState, "program">` |
| 6 | builder.ts, `addToAffectedFilesPendingEmit.state` | `Omit<BuilderProgramState, "program">` |
| 21 | utilities.ts, `createDiagnosticCollection` → `lookup.diagnostics` | `SortedReadonlyArray<Diagnostic> \| undefined` |
| 43 | commandLineParser.ts, `parseConfig.resolutionStack` | `readonly string[]` |
| 43 | commandLineParser.ts, `getExtendedConfig.resolutionStack` | `readonly string[]` |

## Premises

The two builder functions update emit bookkeeping and have no use for
`program`. The narrower view still permits those bookkeeping updates.
The diagnostic lookup rebinds its local and only searches and indexes the
array. The add path works through its own writable local. Both input arrays carry the same sorted marker. The collection keeps them
sorted on its write path. The view preserves the marker and exposes only
readonly array operations. Its added import is type-only. The views escape to the search helpers and the returned diagnostic,
which retains its existing type.

Cause 43 is an amendment to **52-json-values**. The existing
`parseConfig` replacement emits the readonly type and the same validated
batch pins the other end of the recursive call, `getExtendedConfig`.
The assignment in `parseConfig` rebinds the parameter to a fresh array
from `concat`. It leaves the incoming array untouched. The two private
functions pass the view to each other. The public config reader
signatures keep their existing input types.

## Composition

`adapt.cjs` locates each slot by owner and member through stock TypeScript
6.0.3's AST, and validates the expected type before writing any file.
Rerunning it on the adapted tree makes no changes. The lookup's input
type is the text left by `30-indexed-reads`. Adaptation 30 changes an
indexed read in the body and leaves the declaration intact, so this
round has no amendment to 30.

Later adapters pinning the changed owners, after 52 for the stack parameters
and after this directory for builder and lookup, are **none**. The nearby
`99-readonly-family3-views` pin on the diagnostic collection addresses its
separate `getDiagnostics` implementation.

## Left out

Causes 14 and 17 remain outside this adapter. Stock TypeScript resolves all
four sites to `array2: T[]` of the same internal overload,
`core.ts:656:1 concatenate<T>(array1: T[], array2: T[]): T[]`.

- Cause 14, checker.ts:49759:63 and 49765:62. The semantic diagnostics go
  into the overload's writable `Diagnostic[]`. The enclosing
  `getDiagnosticsWorker` return is already readonly.
- Cause 17, checker.ts:32666:21 and 32690:21. The maps go into the overload's
  writable outer array of readonly discriminator tuples.

The implementation returns `array2` by identity when the first array is
empty. A consumer can write through that result. See
`services/completions.ts:4030` for the assignment to `symbols` and
`services/completions.ts:4055` for its `.push`. Changing the shared overload
would change the contract that allows this use.

This round amends [the first family 3 README](../99-readonly-family3-views/README.md)
to withdraw the four cleared-site claims above, using Performance's
J-ntvcw70 evidence at f881d0d5.
