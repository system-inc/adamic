# 98-axolotl-a: required reads in es2017.ts and classFields.ts

Eight indexed reads in `src/compiler/transformers/es2017.ts` and
`src/compiler/transformers/classFields.ts` each gain a `!` at their original
evaluation point. Node erases it; Adamic checks it loudly. Both files go from 6
own-file checker diagnostics to 0. Adaptation 33 recorded these eight
occurrences as declines with unfinished invariant reviews; this head finishes
those reviews and asserts them. Nothing else changes: no read is added, moved,
hoisted, defaulted or turned into an optional chain.

[sites.json](sites.json) records each site: file, function, expression,
occurrence and total, the parent node kind, the class (as 30 and 33 use them),
the action and a one-line invariant. Line and column are documentation in
main's adapted tree, never edit offsets.

Run `node adapt.cjs <tree>` with `NODE_PATH` at stock typescript@6.0.3 (apply
supplies it). The adapter parses current text with stock 6.0.3 and, for each
site, requires the recorded occurrence count, the enclosing function and the
parent node kind, with or without the `!` already present. It refuses a
doubled assertion, a site that drifted or moved, and a tree where some sites
are asserted and others are not. It transpiles every file before and after
with comments kept and removed and refuses any JavaScript change, all before
writing anything. Inserting one character keeps the files' CRLF bytes.

## The invariants

es2017.ts:

- `node.declarations[0]` in visitVariableDeclarationListWithCollidingNames,
  U-endpoint. Only the for-in and for-of visitors pass `hasReceiver` true, each
  after `isVariableDeclarationListWithCollidingName` proved the list nonempty
  with `declarations.some`. The read is dereferenced in place (`.name`), so an
  absent declaration already throws on Node.
- `node.parameters[i]` in transformAsyncFunctionBody, U-loop: `0 <= i <
  node.parameters.length`.
- `outerParameters[i]`, U-parallel: upstream's own `Debug.assert(i <
  outerParameters.length)` runs just before it. The `!` sits on the two reads,
  not on the five uses the checker flagged.

classFields.ts:

- `privateInstanceMethodsAndAccessors[0]` in visitInNewClassLexicalEnvironment,
  U-endpoint: guarded by `some(privateInstanceMethodsAndAccessors)`, which
  without a predicate is `length > 0`.
- `superPath[superPathDepth]` in transformConstructorBodyWorker, U-position.
  The only outside call, in transformConstructorBody, is guarded by
  `superStatementIndices.length`, at depth 0. The worker recurses only when the
  statement at the path entry is a try statement, and
  findSuperStatementIndexPathWorker (transformers/utilities.ts) records a try
  statement's index only after finding a further entry inside its try block;
  the final entry is a super call's expression statement, never a try. So depth
  + 1 is present whenever the worker recurses. One read, three diagnostics.
- `statementsIn[superStatementIndex]`, U-position: every path entry is an index
  `i < statements.length` the path finder took from the same array the worker
  receives at that depth (constructor.body.statements, then each try block's
  statements). Asserting it is needed for 0: with only the path entry asserted,
  the census finds TS2345 at `isTryStatement(superStatement)` (2254:28). It is
  asserted on its own invariant, not to silence that.
- `statementsIn[statementOffset]` and `constructor.body.statements[statementOffset]`,
  U-loop: each is the first statement inside `while (statementOffset <
  ...length)` over a parsed statement NodeArray.

No site was declined. Every reading agreed with the invariant given.

## Before and after

Own-file diagnostics, counted the way the daily meter counts: the census tool
(`go build ./stage3/census/tool`) over each adapted tree's `src/compiler`, then
stage3/meter/report.py's `report()` over its JSONL, attributing each diagnostic
to its primary location and deduplicating repeated observations.

| File | Before | After |
| --- | ---: | ---: |
| transformers/es2017.ts | 6 | 0 |
| transformers/classFields.ts | 6 | 0 |
| Own file, all of src/compiler | 58/79 | 60/79 |

The before tree is main's apply (the Nov 1 meter's 6 and 6, and its 58/79,
reproduced). Its twelve findings are TS2532 at es2017.ts:599:86, TS18048 at
764:38, 765:25, 765:58, 767:76 and 770:44; classFields.ts TS2532 at 1835:21,
TS2538 at 2251:45, TS18048 at 2252:97 and 2253:27, TS2345 at 2297:52 and
2376:56. No other file's count moved.

JavaScript: both files transpiled from the before and after trees with stock
6.0.3, comments kept and removed, are byte-identical (four empty diffs). The
two trees differ in these two files only.

Apply: `stage3/apply.sh` passes end to end with 98 in place; its row is
`98-axolotl-a | 2 | 8 | 8`. Adaptation 90 pins es2017.ts by hash and runs
before 98, so it sees the same bytes as on main; likewise 33, whose ledger
still lists these eight as declines and checks they are unasserted, runs
before 98 and passes. A second run of `adapt.cjs`
on the applied tree reports 0 assertions and leaves both files' bytes
unchanged. Partial-apply and drift probes are refused before any write.

Mutant: dropping the `!` from `superPath[superPathDepth]!` raises classFields.ts
from 0 to 3 own-file diagnostics (TS2538 at 2251:45, TS18048 at 2252:97 and
2253:27), es2017.ts staying at 0 and no other file moving. A second probe,
dropping only the `!` from `statementsIn[superStatementIndex]!`, raises
classFields.ts from 0 to 1 (the TS2345 above). Both ran on scratch copies of
the applied tree; the committed adapter is unchanged.

lane: run as part of wave N (Loom).
