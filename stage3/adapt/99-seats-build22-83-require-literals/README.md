# Collect require specifier literals

Adamic's Nov 22 2026 ruling at about 00:00 UTC on #bbssdcy, in
`tail-proposals-2330.md`, assigns Adamic acceptance of this family 83
change after review and census because it changes JavaScript.

The stop is `src/compiler/core.ts:675:12` in the
`getRequiresAtTopOfFile` unit at `src/compiler/utilities.ts:10126:1`.
The sole caller, `inferPreference` at `utilities.ts:10098:56`, reads only
`r.arguments[0]` from each returned call before visiting the specifiers.

After the variable-statement guard succeeds, visit its declarations in
declaration-list order and push each initializer's argument-zero literal.
After the expression-statement guard succeeds, push that call's
argument-zero literal. Both branches append to a `StringLiteralLike[]` as
the source-file statement scan reaches them. The scan exits before the
next statement after counting four non-require statements. The caller
receives the ordered literals directly.

The proofs retain the argument's type, as required by the inferred-`any`
and refusal rules in `docs/rulings.md:39` and `:282`. The collector
constructs an array of the proven literal nodes, within the array
distinction in `:333`, and avoids building, concatenating, returning, or
iterating an array of call nodes. The collector is file-private and the
call is inside a nested function body. No exposed declaration changes, so
the public-API limit at `:397` is outside the scope of the edit.

From the repository root, run
`NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-83-require-literals/test.cjs <adapted-tree>`
after the locked package install and a full apply. The test reads the
adapter's pre-edit snapshot from that tree. It checks replay, the emitted
collectors' order and literal identity, the four-statement cutoff, and
empty and malformed inputs. It catches a cutoff of five and a declaration
loop that collects only the first initializer.
