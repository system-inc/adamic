# 99l: declare flatten's result as T[]

On TypeScript 6.0.3, adapted core.ts:381:26 is `result`, the first
argument of `addRange(result, v)` inside `flatten`. The evolving array's
NonNullable<T> elements otherwise reach a writable T[] parameter.
The only source edit is `const result: T[] = [];`. Pushes use an element
upcast, addRange receives T[], and the return already has its declared type.

The adapter uses stock TypeScript 6.0.3 to parse the complete file before
and after planning. The unique top-level flatten owner is pinned by exact
before/after text and owner-scoped SHA-256. Changed, moved, duplicate,
missing, or partly adapted owners are refused before writing. Unrelated
functions are preserved. CRLF is preserved, the source is reread before
writing, and a second run makes zero edits and writes. The complete core.ts
JavaScript must match with comments kept and removed before any write.
`javascriptChanging` is false.

Python's real directory sort puts 99l immediately after 99k. See
[PROOF.md](PROOF.md) for passing checks and [CENSUS.md](CENSUS.md) for
the observed new first stops of all 73 moving roots.
