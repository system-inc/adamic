# Remaining source adaptations

This unit is partial. Its any owner edits remove 11 explicit-any tokens at ten
reviewed source lines in six files. No compiler source is edited. The current
main base is ef3141e9b1152ab51b51497f8ce3a2799449c8a3; the additional disposition
table is stage3/notyet-table/TABLE.md at dc6b1529ae9d2a2210672e105c8bb6374619a59d.

The main latent census observes 37 direct any blockers, all NotYet, not Refused:
30 values, six return types and one field store. The any-only tree has 36.
The stock parser counts 135 explicit tokens before and 124 after. These are
different measurements: the latent tool skips checker-diagnosed bodies and
stops at each eligible unit's first lowering error. The historical table's 85
value sites are not reproduced on this current main and this latent tool.

rules.json records every exact owner edit and its runtime domain. Scanner's
unused state is undefined, as both callers supply. String.fromCodePoint and
Function.name use their existing library owners. ProgramHost's readFile
projection uses its declared BuilderProgram bound; the full host helper retains
the caller's T to preserve factory callback variance. NodeArray callbacks carry
Node, pragma argument names carry string, and the key-only mapped value is never.
No new assertion, alias for any, or runtime statement is introduced.

An exploratory direct Array.at replacement failed because the selected library
has no at declaration. It was rejected. A broad ProgramHost<BuilderProgram>
helper failed two caller checks; the final helper instead retains its generic T.
The retained upstream build and stock 6.0.3 compiler both have zero diagnostics.

All 79 compiler source files are compared by an independent reconstruction;
each of six touched files emits identical standalone JavaScript. All ten built
JavaScript files and the public API are byte-identical to main. Four internal
declaration artifacts change through the reviewed type edits. There are no new
fixtures and no counts.md change.

The real scanner-state mutant restores any at scanner.ts:952:103 in an isolated
copy of the adapted tree. Its complete census has 37 direct any blockers again,
with exactly that one extra any site. The string-state mutant produces TS2345
at the two existing callers. A second adapter pass edits zero sites.

Source probes show main compiles any-array storage and length-only parameters;
the current census has zero `an array of any` sites. This storage kind is dropped
from adaptation. This does not establish that reading arbitrary any elements is
supported. Any values, returns and calls still fail their probes. Debugger and
void are refused; with is rejected by the checker and has zero compiler-source
syntax sites. Variance rows are left to their owner.

All tests and measurements write logs. Source the environment printed by setup
before running commands. NODE_PATH selects the stock 6.0.3 API cache.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adaptation41-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh NEW_TREE > apply.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/adapt/41-explicit-any-remaining/prove.cjs MAIN_TREE ANY_ONLY_TREE proof.json > proof.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" NEW_OVERLAY > overlay.log 2>&1
gofmt -w NEW_OVERLAY/*.go
go build -buildvcs=false -overlay=NEW_OVERLAY/overlay.json -o NEW_BINARY ./stage3/census/latent/tool > census-build.log 2>&1
python3 stage3/census/latent/audit.py NEW_BINARY > audit.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 NEW_BINARY NEW_TREE/src/compiler census.jsonl > census.log 2>&1
python3 stage3/adapt/41-explicit-any-remaining/census-report.py census.jsonl NEW_TREE report.json > recount.log 2>&1
bash stage3/oracle/run.sh NEW_TREE NEW_ORACLE > oracle.log 2>&1
bash stage3/lane/run.sh NEW_LANE > lane.log 2>&1
```

The source identity proof is specific to the any-only tree. Void adaptations
have a separate evaluation/return proof and full oracle comparison. See the
final report for commands, completed measurements, decisions and remaining work.

Historical result before the October 8 ruling: 15 void expressions removed, lane PASS, and full oracle identical to
main with its one sanctioned API failure. See [REPORT.md](REPORT.md),
[RANKED.md](RANKED.md), [REMAINING.md](REMAINING.md), [DEBUGGER.md](DEBUGGER.md),
[ANY-RETURNS.md](ANY-RETURNS.md), [ANY-CALLS.md](ANY-CALLS.md) and [DROPPED.md](DROPPED.md).

To reconstruct the any-only input, use adapter commit 07aacdf9 in a separate
checkout and apply its pipeline. Batch 4 now retains the type edits and drops all void rewrites. The earlier
any-only source identity proof refers to its separately retained tree.
The batch pipeline is independently exercised by the landing lane.

## October 8 void ruling

Void was ruled a compiler lesson on October 8. The compiler's area-next admits
it with Node agreement. Tsc's source stays as written: all 15 void rewrites are
dropped outright from adaptation 41, while its type edits remain active.
Any void site still stopping after area-next lands is a compiler bug.

The void adapter, rewrite rules and adapter proof are removed. Historical
receipts remain under evidence/ and in REPORT.md; they describe the superseded
proposal, not the current adaptation. The removed proposal's 15 source sites
are listed below for provenance. These coordinates precede later adaptations.

| Site | Source | Expression |
|---|---|---|
| 1 | src/compiler/checker.ts:9062 | `getTypeFromTypeReference(existing)` |
| 2 | src/compiler/checker.ts:9293 | `getTypeFromImportTypeNode(existing)` |
| 3 | src/compiler/checker.ts:9395 | `getInternalSymbolName(symbol, baseName)` |
| 4 | src/compiler/checker.ts:9605 | `getPropertiesOfType(getTypeOfSymbol(symbol))` |
| 5 | src/compiler/checker.ts:21958 | `instantiateType(sourceRestType &#124;&#124; targetRestType, reportUnreliableMarkers)` |
| 6 | src/compiler/checker.ts:23539 | `mappedKeys.push(instantiateType(nameType, appendTypeMapping(targetType.mapper, getTypeParameterFromMappedType(targetType), t)))` |
| 7 | src/compiler/checker.ts:48699 | `resolveExternalModuleName(node, node.moduleSpecifier, /*ignoreErrors*/ undefined, Diagnostics.Cannot_find_module_or_type_declarations_for_side_effect_import_of_0)` |
| 8 | src/compiler/checker.ts:50792 | `resolveExternalModuleSymbol(fileSymbol)` |
| 9 | src/compiler/checker.ts:51078 | `checkExpressionCached(node)` |
| 10 | src/compiler/moduleNameResolver.ts:610 | `diagnostics.push(diag)` |
| 11 | src/compiler/moduleNameResolver.ts:1838 | `diagnostics.push(diag)` |
| 12 | src/compiler/moduleNameResolver.ts:3277 | `diagnostics.push(diag)` |
| 13 | src/compiler/moduleNameResolver.ts:3382 | `diagnostics.push(diag)` |
| 14 | src/compiler/transformers/declarations.ts:274 | `0` |
| 15 | src/compiler/utilities.ts:9004 | `(file.externalModuleIndicator = combined(file, options))` |

Batch 4 verification: all 23 type-edit rules still apply. The fresh landing lane
passes with 106366 passing, one sanctioned API acknowledgement failure and zero
pending, matching main. All ten built JavaScript artifacts and typescript.d.ts
are byte-identical to main. The prior real build with void rewrites fails the
same comparison on three bundles. Type adapter idempotence and LF reconstruction
pass in an isolated 41 control; both retained anchor mutants are caught. See
[evidence/batch4-ruling.json](evidence/batch4-ruling.json).
