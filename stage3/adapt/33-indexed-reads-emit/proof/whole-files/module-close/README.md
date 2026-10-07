# transformers/module/module.ts at zero

Followup all-code latent census **2 -> 0** (initial target file was 4).
currentModuleInfo now truthfully includes undefined: the unconditional emit
notification receives missing cache entries for skipped scripts. The cache read
is not asserted. The three emission getExports reads retain their existing
optional access. No initializer is added or changed.

Twenty-two required receiver reads have individual required-context ledger
entries. All eleven reader functions are reachable exclusively from the
transformation root, not from either emission root. verify-context.cjs constructs
a conservative symbol-resolved reference graph with stock TypeScript 6.0.3;
returned delegates and passed callbacks count as edges. Symbols distinguish the
local transformModule delegate variable from the outer transformer function.
The only assignments to currentModuleInfo are the transform-root producer and
end reset and the emission-root cache read and reset. The producer precedes the
single delegate call; the reset follows its return. No reachable transform
function recursively invokes transformSourceFile. Upstream's transformNodes
completes source transformation before the printer invokes emission callbacks.
Thus these private transform helpers retain a present producer value throughout
this graph. Assertions stay on original receiver reads, without moving reads,
adding defaults, or replacing absent cache values.

The second diagnostic is an overload contract: arrayFrom inferred an optional
element from Set<Identifier>. Every set insertion is a binding from
ExternalModuleInfo.exportedBindings, whose element type is Identifier. The
explicit arrayFrom<Identifier> argument selects the truthful iterable overload.
Both type edits are AST addressed, shape checked and planned before writes.
All stock JavaScript bytes and CRLF are preserved, and idempotence passes.
The required-context ! -> ?? 0 mutant fails JavaScript and site-contract checks.
The default oracle and owner/overload/protocol mutants accompany this proof.

Commands use source-only tree /tmp/emit33-close-meter-source, checker
/tmp/emit33-close-meter and before snapshot
/tmp/emit33-close-source-before-module-close. Census output is
/tmp/emit33-module-close-census; default oracle output is
/tmp/emit33-module-close-oracle. NODE_PATH pins stock typescript@6.0.3.

Default oracle: **106367 passing**, zero failing/pending, empty baseline
diff, 213.165s. Exact mechanically checked API exceptions
remain 20's 189 plus 40's 28 lines. Removing the optional owner restores exactly
TS2322 at line 2262; widening the explicit element argument restores exactly
TS2322 at line 2481. Both also fail adapter shape checks. Referencing a required
transform helper from onEmitNode fails the protocol graph proof. All mutants
restore the source-only input in finally blocks. Final idempotence passes.
