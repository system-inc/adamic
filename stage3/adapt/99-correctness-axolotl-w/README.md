# 99w: sortAndDeduplicate default comparer

Task #b5w3ycg. Base and compiler: main a28effb2c.
Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

The only edit changes `compareStringsCaseSensitive as any as Comparer<T>` to
`compareStringsCaseSensitive as unknown as Comparer<T>`. javascriptChanging is false.
Adamic's ruling permits this spelling for Optimizer's per-overload proof; an
`as any` remains refused on its own. This adaptation changes no compiler policy.

The complete owner, including both overloads, was read before editing. The overloads are:

```ts
export function sortAndDeduplicate(array: readonly string[]): SortedReadonlyArray<string>;
export function sortAndDeduplicate<T>(array: readonly T[], comparer: Comparer<T>, equalityComparer?: EqualityComparer<T>): SortedReadonlyArray<T>;
```

Only the first overload permits omission of the comparer, and it fixes the
element type to string. The generic overload requires Comparer<T>. Overloads
are unchanged. The adapter pins the entire three-declaration group by exact
text and owner-scoped before/after SHA-256, never by a whole-file hash.
Changed, nested, duplicate, missing and half-adapted owners fail before any write.
Stock TypeScript 6.0.3 parses the whole file. Whole core.ts transpilation must
be byte-identical with comments retained and removed before a write is allowed.
Replay changes zero files and preserves source bytes and mtime.

The pre-edit git grep of stage3/adapt found no existing owner pin. 40/41
evidence names program.ts's sortAndDeduplicateDiagnostics; 99k's emitted
JavaScript witnesses contain the function but do not pin it.

Verification and the ordinary main compiler census are recorded in PROOF.md
and CENSUS.md when completed. Logs are retained as files.
