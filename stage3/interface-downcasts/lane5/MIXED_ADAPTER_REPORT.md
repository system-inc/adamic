Built: ranked explicit scalar/boxed callable adapter candidates; admission remains refused.
Commits: this checkpoint contains inventory and planning only; predicates evidence b131a36d and 1191e3f7.
Commands/results: Node TypeScript AST inventory passes: static 5 pairs/10 reads; Unknown fallback 13 pairs/69 reads.
Mutants: no new executable check added; predicates worker's unsafe union-admission mutant remains the evidence for retaining refusal.
Limits: aliases, producer-side variance and exact reachability remain unmeasured; no additional certified pairs or reads.

The adapter family is inserted at each affected pair's existing read-count rank,
not appended behind lower-read families. The static first entry is rank 69,
Map<Path, string | false>.get (4 reads), followed by rank 94 (3 reads).
The nested callback at Map<string, string | number>.forEach is rank 265 (1 read).
The Unknown-fallback first entry is NodeFactory.createNumericLiteral, rank 53
(32 reads); this upper-bound inventory also includes JSON.stringify, rank 87
(19 reads). These inventories overlap and must not be added together.

Reproduce with Node and the independently obtained, pinned TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8 library:

```
node stage3/interface-downcasts/lane5/rank-mixed-callable-adapters.cjs /path/to/typescript/lib/typescript.js
```

The generated JSON includes the original source witness for each candidate.
The AST walk inspects parameters and results of callable signatures, including
nested callbacks. It recognizes explicit number/boolean versus string/object/array
representations. It deliberately does not resolve named aliases or infer producer
signatures. Thus this is a candidate subset, not a complete adapter-demand census.

Read b131a36d:stage3/census/predicates/parser-stops/REPORT.md and investigation
1191e3f74b1e60a62aa1eb947b0bc45136522ffc. Matching union signatures passed Node,
JavaScript and sanitized native in the prototype, but () => number assigned to
() => number | string segfaulted in adamic_union_to_string. Results need boxing,
as do compatible parameter views. Blanket union admission must remain refused.
Adapters must cover named forwarders, arrows, fields, method adapters and aliases,
preserve ownership on normal and exceptional returns, and handle missing/default
arguments using their incoming representation. Parameter adaptation must obey
variance; a wider view cannot invoke a narrower producer with an unsupported value.

No compiler admission changed in this checkpoint. The candidate certification
ledger remains 2 pairs/11 reads, with 306 pairs/1492 reads remaining in the static
inventory. The October 12, 18:00 UTC estimate is retained as the working estimate;
complete adapter demand cannot be sized until aliases and producer signatures are
measured. Stored erased-marker implementation is still uncommitted and is not part
of this inventory-only push.
