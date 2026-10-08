# moduleSpecifiers.ts at zero

Three findings become zero; whole-tree count is 377 to 374. The source ledger
adds one U-endpoint required read at the original switch selector, and the
closure ledger records three truthful optional tuple unions, one erased map
iterator contract, two private nonempty parameter contracts and two private
factory-result refinements. The exported ModuleSpecifierPreferences interface
retains its array return contract so services can override it.

The preference factory's every normal return is a fresh literal array of one
to four enum values; its only default throws. Private getLocalModuleSpecifier
callers create their own preferences via this factory, and tryGetModuleNameAsNodeModule
creates its own instance locally. No external override reaches these objects.
The six processEnding calls receive these fresh arrays, the root-dirs forwarding
parameter, or a fresh [ending] singleton. The singleton in validateEnding is
reached only after ending has narrowed to Minimal. No forwarding body mutates
the endings array before either indexed read.

nonempty-endings.json records the twelve reviewed AST function bodies. Stock
6.0.3 runtime hashes erase type-only adaptations while rejecting changed runtime
control flow, producers or writes. The external interface's unchanged shape is
checked separately. An empty factory-return mutant is rejected before any
source file is written, including an earlier execute contract planned for repair.
Removing a tuple undefined union restores TS2322 and fails the declaration guard;
changing the selected ! to ?? 0 fails emitted bytes and its occurrence contract.
Strengthening the external interface is rejected by the boundary guard.

All 26 files have identical stock-emitted JavaScript and unchanged CRLF;
the second CLI run makes zero edits. The default oracle passes all 106,367 tests
with no baseline differences. The API owner proof remains exactly 217 sanctioned
lines (20:189, 40:28); no adaptation 70 is present in this integration.
