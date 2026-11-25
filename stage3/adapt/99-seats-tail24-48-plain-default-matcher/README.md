Adamic ruled this adaptation on Nov 22 2026 (#bbssdcy, 21:49Z, then family 126; the tail pass's family 48); Compiler accepts the head after its review and census, under the widened split.

Recut at Adamic `9328e8e5671195bcb8e291746f4523430a2080f6` for TypeScript
`050880ce59e30b356b686bd3144efe24f875ebc8`.

Before:

```ts
export function getSubPatternFromSpec(
    spec: string,
    basePath: string,
    usage: "files" | "directories" | "exclude",
    { singleAsteriskRegexFragment, doubleAsteriskRegexFragment, replaceWildcardCharacter }: WildcardMatcher = wildcardMatchers[usage],
): string | undefined
```

After:

```ts
export function getSubPatternFromSpec(
    spec: string,
    basePath: string,
    usage: "files" | "directories" | "exclude",
    matcher: WildcardMatcher = wildcardMatchers[usage],
): string | undefined {
    const { singleAsteriskRegexFragment, doubleAsteriskRegexFragment, replaceWildcardCharacter } = matcher;
    let subpattern = "";
    // Original body continues.
}
```

> **Quick rulings (yes, yes with conditions, or hold):**
>
> - **126 (1): yes,** a plain default plus the destructuring moved to the top of the body. It's the last parameter, and the lookup still runs only on undefined and before the field reads. No new row for a destructured parameter with a default.

The matcher is last, so no later default reads its bindings. The lookup runs
in the callee on omission or explicit `undefined`, before the same three field
reads in the same order. All typed callers supply a `WildcardMatcher` or allow
the default. Explicit `null` still throws `TypeError` before the first original
body statement. Its message may name a different binding. Both versions have
`length === 3`.

The structural pins require one parsed top-level `@internal` implementation,
the exact signature and final parameter, the opening and first original body
statement, and no existing `matcher`, `arguments` or `eval` identifiers in
the body. The adapter accepts LF or CRLF and replays without writing. The
`adapter-table-mentions` gate does not apply because the matcher table is
unchanged and the default uses its original indexed read.

`test.cjs` compares a replay from this adapter's before ref with its after ref
and executes both captured owners in equivalent VM contexts. The contexts
record table and field reads and the first original body operation.

Pinned rulings:

Line 220

> - **Omitted optional or default arguments are absent on every call path; defaults evaluate in the callee; general union function arguments stay NotYet.** Meaning, NotYet. Oct 7 2026, docs; source: docs/nested-functions.md:567-583. Spec: What's in, Values and statements (default and optional parameters).

Line 432

>   - `a function's properties aren't read by a run-time key yet`: stop_function_name.a; lifted by #xdcec18 (Runtime Performance). A function read by a run-time key, one a key read handed out (an array's map, a matchAll iterator's next) or one the program made, stops alike on both backends for every key but `length` and, on a handed-out function, `name`: natively a closure, on the JavaScript backend an AdamicClosure or Node's own function, which stopped as a runtime value before. Ruled in by @system_adamic on #zhrjf2t, Nov 6, on Language's m02 should from #g187f1e. Its `length` answers as Node's does on both backends (#xdcec18, Runtime Performance): natively the count every closure carries, an immortal closure's Node's own and a program's the parameters before the first default or rest, and on the JavaScript backend the function's own (function_length.a, and stop_function_length.a and stop_function_length_iterator.a, whose entries left the list). A handed-out function's `name` answers as Node's does on both backends too (#xdcec18): natively the immortal string its table writes, its key, a constructor's own name (Object, Array, Iterator, Stats), or the empty string for a Stats' anonymous functions, and on the JavaScript backend the function's own (function_name.a). A program function's `name` still stops in this form (stop_function_name.a, an arrow read off an object), as does every other key, until a program's closure carries a name.
