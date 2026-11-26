Adamic, item 14, Nov 24 16:19Z on #bbssdcy: "A yes, on a condition: extend :406's Nov 17 extension to `Node.flags`. tsc's public `updateSourceFile` disproves its `readonly`". Compiler accepts the head after its review and census.

The public `updateSourceFile` writes `flags` on an already constructed node. Under the named extension of `docs/rulings.md:406`, the adapter makes `Node.flags` writable. The checker counted zero narrower declarations across 714 TypeScript files. The table in `evidence/flags-declarations.tsv` uses coordinates in the applied tree at f7de5d65. The summary is in `evidence/flags-count.txt`. `ADAPTED_BASELINE_TREE` is the full tree that `bash stage3/apply.sh <out>` writes at f7de5d65 with this adapter's directory absent. From the repository root, after installing the locked packages with `npm ci --ignore-scripts --no-audit --no-fund` in `stage3/api`, reproduce both with `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-scout24-writable-node-flags/evidence/count-flags.cjs "$PWD" "$ADAPTED_BASELINE_TREE" stage3/adapt/99-seats-scout24-writable-node-flags/evidence`. Rules 96 and 265 still govern writable views and unproven writes through `Mutable<T>`.

The adapter pins the exact property signature and the statements it edits. It plans every change before writing. Each of the 20 casts below is the receiver of a single assignment to `flags`. Removing the cast preserves the receiver, assignment operator and right-hand side. These are coordinates in the applied tree at the pin.

- `src/compiler/binder.ts:1101`: `(node as Mutable<Node>).flags &= ~NodeFlags.Unreachable;`
- `src/compiler/binder.ts:1109`: `(node as Mutable<Node>).flags |= NodeFlags.Unreachable;`
- `src/compiler/binder.ts:1799`: `(node.label as Mutable<Node>).flags |= NodeFlags.Unreachable;`
- `src/compiler/program.ts:3328`: `(externalHelpersModuleReference as Mutable<Node>).flags &= ~NodeFlags.Synthesized;`
- `src/compiler/program.ts:3329`: `(importDecl as Mutable<Node>).flags &= ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6476`: `(clone as Mutable<T>).flags |= node.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:7477`: `(node as Mutable<Node>).flags |= NodeFlags.Synthesized;`
- `src/compiler/parser.ts:1404`: `(newSourceFile as Mutable<SourceFile>).flags |= sourceFile.flags & NodeFlags.PermanentlySetIncrementalFlags;`
- `src/compiler/parser.ts:1833`: `(node as Mutable<T>).flags |= NodeFlags.Deprecated;`
- `src/compiler/parser.ts:2579`: `(node as Mutable<T>).flags |= contextFlags;`
- `src/compiler/parser.ts:2587`: `(node as Mutable<T>).flags |= NodeFlags.ThisNodeHasError;`
- `src/compiler/parser.ts:4423`: `(node as Mutable<Node>).flags = type.flags;`
- `src/compiler/parser.ts:6384`: `(node as Mutable<NonNullExpression>).flags |= NodeFlags.OptionalChain;`
- `src/compiler/parser.ts:6492`: `(tagExpression as Mutable<Node>).flags |= NodeFlags.OptionalChain;`
- `src/compiler/parser.ts:7460`: `(m as Mutable<Node>).flags |= NodeFlags.Ambient;`
- `src/compiler/parser.ts:8091`: `(m as Mutable<Node>).flags |= NodeFlags.Ambient;`
- `src/compiler/parser.ts:9672`: `(typeNameOrNamespaceName as Mutable<Identifier>).flags |= NodeFlags.IdentifierIsInJSDocNamespace;`
- `src/compiler/utilities.ts:971`: `(node as Mutable<Node>).flags |= NodeFlags.ThisNodeOrAnySubNodesHasError;`
- `src/compiler/utilities.ts:977`: `(node as Mutable<Node>).flags |= NodeFlags.HasAggregatedChildData;`
- `src/compiler/utilities.ts:10764`: `(node as Mutable<T>).flags = newFlags;`

The adapter edits only the direct `(<expr> as Mutable<...>).flags` form and only the 20 statements above. Before writing any bytes, on first apply and on replay, the walk uses the stock TypeScript 6.0.3 checker to find writes, in every source file the program loads under `src/`, whatever its extension, that resolve to `Node`'s `flags` declaration and refuses each unpinned write at its `file:line:column`. It also refuses unpinned `flags` writes with an unresolved target or an `any` receiver, and writes through a nonliteral key whose type could be `"flags"`, including `string` or `any`, on a receiver with a `Node` member. The checker follows base constraints on key and receiver types, including type parameters and conditional, indexed and `keyof` types. A write target wrapped in `as`, `<T>`, `satisfies`, `!` or parentheses is the same write. Reads are outside the walk. Each pin records the exact write, cast-free on replay, and holds its nearest function-like owner's whole text, exactly, on both snapshots, the `-before` text and the expected `-after` text, kept in `owners.json`. Any difference refuses. Each owner is found on every walk, whether or not a write claims it, by its key: its full path from the source file, every ancestor node, each step with its syntax kind, its slot on its parent, and its name where it has one. The pinned key must resolve to exactly one node, and the node holding the pinned write must be it. A missing or ambiguous owner and a pinned write outside its owner's node each refuse, naming both paths. A write's effect depends on everything in its owner that runs before it. An early `return`, `break`, `continue` or `throw`, a reordered statement, a changed condition or a wrap such as a static block or field initializer changes that text. The refusal reports the write's path to its owner, showing each ancestor's kind and slot and the text of any condition, loop header, catch clause, `switch` or `case`. The owner's text decides. Any upstream edit inside one of the owners refuses at the next re-pin, which is when a person should look anyway. A pin holds its write's owner exactly, and what the owner calls or reads from outside is held by tsc's own tests in the lane, not by this guard. "Active" means the reviewed write's own node, the one its pinned path names, not which function callers are handed: a factory object publishing another function, or a renamed export that throws, changes what tsc does, and tsc's own tests in the lane judge that. The walk pins 45 writes to `Node`'s `flags` with their functions, comprising the 20 direct writes above, 14 alias writes, 8 parameter writes and 3 `this` writes below. It also pins one keyed write. The 35 distinct owners total 41,057 bytes before and 40,667 bytes after.

Each of the fourteen writes through fresh factory nodes is pinned by its function and binding initializer. A changed binding or missing pin refuses. Coordinates refer to the final applied tree at the pin.

- `src/compiler/factory/nodeFactory.ts:1225`: `node.flags |= NodeFlags.IdentifierHasExtendedUnicodeEscape;`
- `src/compiler/factory/nodeFactory.ts:2826`: `node.flags |= NodeFlags.OptionalChain;`
- `src/compiler/factory/nodeFactory.ts:2891`: `node.flags |= NodeFlags.OptionalChain;`
- `src/compiler/factory/nodeFactory.ts:2961`: `node.flags |= NodeFlags.OptionalChain;`
- `src/compiler/factory/nodeFactory.ts:3668`: `node.flags |= NodeFlags.OptionalChain;`
- `src/compiler/factory/nodeFactory.ts:4168`: `node.flags |= flags & NodeFlags.BlockScoped;`
- `src/compiler/factory/nodeFactory.ts:4442`: `node.flags |= flags & (NodeFlags.Namespace | NodeFlags.NestedNamespace | NodeFlags.GlobalAugmentation);`
- `src/compiler/factory/nodeFactory.ts:6130`: `node.flags |= flags;`
- `src/compiler/factory/nodeFactory.ts:6202`: `node.flags |= source.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6217`: `node.flags |= source.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6409`: `clone.flags |= node.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6418`: `clone.flags |= node.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6433`: `clone.flags |= node.flags & ~NodeFlags.Synthesized;`
- `src/compiler/factory/nodeFactory.ts:6442`: `clone.flags |= node.flags & ~NodeFlags.Synthesized;`

The eight writes in `binder.ts` through `Mutable<...>` parameters and the three writes through `this: Mutable<Node>` in the `Node`, `Token` and `Identifier` functions of `utilities.ts` resolve to the same symbol, and each is pinned by its function and parameter declaration.

- `src/compiler/binder.ts:1027`: `node.flags &= ~(NodeFlags.ReachabilityAndEmitFlags | NodeFlags.ContainsThis);`
- `src/compiler/binder.ts:1029`: `node.flags |= NodeFlags.HasImplicitReturn;`
- `src/compiler/binder.ts:1030`: `node.flags |= NodeFlags.HasExplicitReturn;`
- `src/compiler/binder.ts:1034`: `node.flags |= NodeFlags.ContainsThis;`
- `src/compiler/binder.ts:1037`: `node.flags |= emitFlags;`
- `src/compiler/binder.ts:1063`: `node.flags = seenThisKeyword ? node.flags | NodeFlags.ContainsThis : node.flags & ~NodeFlags.ContainsThis;`
- `src/compiler/binder.ts:2342`: `node.flags |= NodeFlags.ExportContext;`
- `src/compiler/binder.ts:2345`: `node.flags &= ~NodeFlags.ExportContext;`
- `src/compiler/utilities.ts:8513`: `this.flags = NodeFlags.None;`
- `src/compiler/utilities.ts:8527`: `this.flags = NodeFlags.None;`
- `src/compiler/utilities.ts:8539`: `this.flags = NodeFlags.None;`

One element write whose constrained key could be `"flags"` is pinned by its text and its function, `cloneNode`.

- `src/compiler/factory/nodeFactory.ts:6485`: `clone[key] = node[key];`

`NodeObject` and `TokenOrIdentifierObject` in `services.ts` declare their own writable `flags` class fields, so their writes are outside the walk, and this edit opens nothing there.
A nonliteral key on an `any` receiver, such as `(x as any)[k] = v` with `k: string`, is outside the rule, which requires a Node member for nonliteral keys. Writes through `Object.assign(node, { flags })`, `Reflect.set` or `Object.defineProperty` are also outside the rule because they have no syntactic write target. Both forms already type-checked before this head.

From the repository root, run the refusal test on the snapshots in `$APPLIED_TREE`, a full tree produced by `stage3/apply.sh`: `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-scout24-writable-node-flags/test.cjs "$APPLIED_TREE"`.

No matching `as Mutable<...>).flags` site was left in this tree. Other `Mutable<T>` casts retain their existing behavior, including the named stay for `setParent` and the separate `escapedText` work.

The public API sanction is one member line. Regenerating `stage3/lane/sanctioned-api.json` moves the proof count in `stage3/lane/test_check.py` from 286 to 287. The same test pins the new member's owner. The manifest changes the runtime hash recorded in `stage3/lane/shard-plan.json`.

The complete apply runs this adapter with the locked TypeScript 6.0.3 parser. A second invocation on an already adapted tree makes zero edits.
