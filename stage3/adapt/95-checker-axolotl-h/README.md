# 95h: truthful internal module-specifier host slots

lane: run as part of a Correctness wave

Current fold result: checks 1 through 6 pass, with the user-authorized and
proved binding normalization. The whole lane and all 398 shards pass; see
[Completed fold validation](#completed-fold-validation).

Base: 95a `ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`.
Only this adaptation directory belongs to H. 95c stays at `7b1e14c56`.
The earlier blocked two-slot attempt is historical evidence at
`a6ce2a25d`; the extended contract below supersedes that attempt.

## Scope, internal audit and 97b ownership

H widens exactly four optional members of the **@internal**
`ModuleSpecifierResolutionHost` interface in src/compiler/types.ts:
`getSymlinkCache`, `readFile`, `getModuleSpecifierCache`, and
`getNearestAncestorDirectoryWithPackageJson`. Each receives an indexed method
type plus undefined, `x?: { x(...): R; }["x"] | undefined`, preserving method
parameter bivariance. All four declarations were resolved to that internal
interface before editing. The interface is absent from the pinned public
`typescript.d.ts` baseline. No public API line is edited.

Accepted 97b is **cli/axolotl-97b**, commit
`27162cab2bb0054b4ed7cad5a9e258f67f8030c1`, directory
`97-cli-axolotl-b`. It owns `directoryExists`, `realpath`,
`getGlobalTypingsCacheLocation`, and `trace`; H never edits them.
54261 requires its global-cache widening too. Apply orders H before 97b.
All canonical census comparisons have 97b on both sides.

Five contextually typed or derived object literals were inventoried across
src/compiler and src/services. The whole return literal at checker.ts:54261
has 14 members; exactly getSymlinkCache, readFile and
getGlobalTypingsCacheLocation may hold undefined as property values.
getPackageJsonInfoCache is always an arrow whose *return* may be undefined.

The entire services createModuleSpecifierResolutionHost literal at
utilities.ts:2454 has 16 members:

| Member | Property value may be undefined? | Owner |
| --- | --- | --- |
| fileExists | No | unchanged |
| getCurrentDirectory | No | unchanged |
| readFile | Yes | H |
| useCaseSensitiveFileNames | No, guarded fallback | unchanged |
| getSymlinkCache | Yes | H |
| getModuleSpecifierCache | Yes | H |
| getPackageJsonInfoCache | No, always an arrow | unchanged |
| getGlobalTypingsCacheLocation | Yes | 97b |
| redirectTargetsMap | No | unchanged |
| getRedirectFromSourceFile | No, always an arrow | unchanged |
| isSourceOfProjectReferenceRedirect | No | unchanged |
| getNearestAncestorDirectoryWithPackageJson | Yes | H |
| getFileIncludeReasons | No | unchanged |
| getCommonSourceDirectory | No | unchanged |
| getDefaultResolutionModeForFile | No | unchanged |
| getModeForResolutionAtIndex | No | unchanged |

The other literals are createProgram, its getEmitHost helper, and the services
getModuleSpecifierResolverHost forwarding spread. Their ModuleSpecifierResolutionHost
undefined stores are owned by 97b, or forward the audited services producer.
No further host member requires widening. Other existing Program/EmitHost
members are inventoried as different destinations and are not changed by H.

## Adapter and independent verifier

Stock TypeScript 6.0.3 through 47-host-errors' compiler() parses and binds the
compiler plus services. The adapter inventories every relevant contextual
literal, all its members, its undefined-valued subset and bound destinations,
and the known forwarding spread. All five writer functions have individual
owner text pins. Each H member has its own ownerBeforeSha256. There is **no
whole-interface or whole-file acceptance pin**, so 97b can modify its own
members independently. Own slot symbols, optionality and unchanged read types
are checked; unrecognized writer, member or owner shapes are refused.

All edits are source slices, planned and parsed before writing types.ts and checker.ts.
Inputs are reread and complete stock JavaScript is compared in both comments
modes before writing. types.ts must remain JavaScript-identical. checker.ts may
change only by the two optional-bind initializers described below. CRLF is
preserved without a printer. --check performs no writes. On the isolated H
fixture a second run reports files 0, declarations 0 and rewrites 0. Full-stack
replay may refuse later pinned owners, as recorded in the fold evidence.

verify.cjs is independent of the planner. It reconstructs H's four parsed
member edits and the two checker initializer edits, compares every src file,
checks CRLF and complete JavaScript, and leaves unrelated 97b edits in the
control intact. types.ts JavaScript is byte-identical. The following measurements
are the historical four-slot run:

| Comments | Bytes | SHA-256 before and after |
| --- | ---: | --- |
| Retained | 139659 | `4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3` |
| Removed | 121398 | `f041d720cfda4ccb137ffc32198ec3d2fce46dbf13404df737dc637e21290639` |

## Exact production census

The unchanged real stage3/census/tool uses production internal/load options.
Each root and identical aggregate observation must agree and count once,
following 95a's compare-census.py. Full diagnostic texts are compared.
Controls are 95a+97b; candidates are 95a+H+97b, before npm installs.

| Profile | Whole program before / after | checker.ts own-file before / after | Removed / added |
| --- | ---: | ---: | ---: |
| tsc entry | 125 / 124 | 17 / 16 | 1 / 0 |
| services/server entry | 3341 / 3339 | 17 / 16 | 2 / 0 |
| all 711 sources | 8488 / 8486 | 23 / 22 | 2 / 0 |

Exactly checker.ts:54261 TS2375 and services/utilities.ts:2454 TS2375 disappear
in profiles that reach both. utilities.ts goes **153 to 152**, not sideways.
No message appears or changes elsewhere. Per-file CSV/JSON tables include
every source file, including zero rows and all src/services files. The all-source
.a importer deliberately includes upstream libraries and harness sources, so
its baseline differs from the normal services/server entry.

The undo mutant restores only readFile, leaving the other three H members and
all 97b edits intact. The source verifier and adapter check reject it, and the
census restores the original checker and services readFile incompatibilities.
No mutant enters the oracle tree. Complete results, exits and raw observations
are retained on the evidence branch.

## Evidence and reproduction

Every file in this directory feeds apply's input key. The fold evidence below
is retained here to keep the shipping branch confined to this directory.
Historical logs, raw censuses, oracle
reports, undo controls, public API checks and composition results live on
[typescript/axolotl-checker-h-evidence](https://github.com/system-inc/adamic/tree/typescript/axolotl-checker-h-evidence/stage3/evidence/95-checker-axolotl-h).
The evidence branch's current report identifies the accepted four-slot run;
Git commit a6ce2a25d preserves the superseded stop report.

Fresh apply is run in a scratch worktree with this directory and the exact
accepted 97b directory added, in numeric order. Canonical commands, with all
test output logged:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95-upstream
bash stage3/apply.sh /workspace/scratch/95hx-apply > /tmp/95hx-apply.log 2>&1
node stage3/adapt/95-checker-axolotl-h/verify.cjs /workspace/scratch/95h97-before /workspace/scratch/95hx-apply > /tmp/95hx-apply-source.json 2>&1
/workspace/scratch/95-census /workspace/scratch/95hx-apply/src/tsc/tsc.ts /workspace/scratch/95hx-apply-tsc.jsonl > /tmp/95hx-apply-tsc.log 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95hx-apply /workspace/scratch/95hx-oracle --limit-seconds 720 > /tmp/95hx-oracle.log 2>&1
```

The bounded all-runner oracle completes in **352.424s**, with **106,366
passing, 1 failing, 0 pending**. Install exits 0 (2.422s), build exits 0
(22.992s), tests exit 1 (326.909s), solely for the standing Public APIs
acknowledgement case. There are no timeouts or exceptions. Built/local API
bytes are exactly main's 590,996 bytes, SHA-256
`b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`;
the accepted reference remains main's 588,626 bytes, SHA-256
`f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.
There is no new public API delta. No lane acceptance is inferred. Fresh apply
exits 0 with rows `95-checker-axolotl-h | 1 | 4 | 4` and
`97-cli-axolotl-b | 3 | 12 | 12`; its exact source map equals the oracle
input. The combined B/C/D/E/H/97b tsc census is 110 whole-program and
2 checker own-file diagnostics. Both orders are checked with 97b and with B, C, D and E,
including the full forward/reverse stack; source maps and exact refs are retained.

Setup passed in 12.003s: Node 0.022s, Go 0.023s, submodules 0.071s, clang
0.157s, tsc 0.162s, Markdown 0.204s, Go build 11.863s, cache warm
11.973s. nproc=5, CPU quota 4. Every build/test shell sources
/workspace/adamic-tools/env.sh. No Go, lane, fixture, meter, checked-in patch
table or other adaptation source is changed by the shipping branch.

Not covered: native tsc compilation/lowering, browser or ESLint integration,
the whole repository Go gate, or lane acceptance. None is claimed by this head.

## maybeBind fold ledger: task kpp0f8s

Adamic's ruling on task kpp0f8s makes `maybeBind(X, X.m)` become
`X.m?.bind(X)` when X is one identifier, as a stage 3 JavaScript-changing
adaptation. Correctness's 99-correctness-axolotl-x owns 44 other sites. Under
the pinned-owner rule, 95h owns the remaining two initializers in
`createBasicNodeBuilderModuleSpecifierResolutionHost` in checker.ts:

- `maybeBind(host, host.getSymlinkCache)` becomes `host.getSymlinkCache?.bind(host)`.
- `maybeBind(host, host.getGlobalTypingsCacheLocation)` becomes `host.getGlobalTypingsCacheLocation?.bind(host)`.

The row retains the exact original owner SHA-256
`9171aad5d35e92052c682093a2f2b7fb0ad49a63787488f78d0aa74161d9b9b1`
and literal SHA-256
`3f47dd64f51c5c74ad8263d40713a28e424f4d28859e714238b8a2417cbae3d7`.
It accepts both complete before and complete after states. An after owner is
normalized by AST initializer slices back to the original owner pin. A mixed
state refuses as half-adapted; any foreign literal edit refuses at its hash.
The callee resolves through the type checker to core.ts's maybeBind declaration,
and both receiver identifiers must resolve to the same symbol. Only the two
inventoried non-optional property accesses qualify.

`mayStoreUndefined`, `hostMember` and `destinationOwner` stay unchanged. Both
forms must have the same pinned value type: `(() => SymlinkCache) | undefined`
and `(() => string | undefined) | undefined`. The planned after tree is rebound
and checked before writes. Both forms evaluate host, then host.m, then host;
a function binds, undefined yields undefined, and another truthy value throws
TypeError at bind in both forms.

The return value now records `javascriptChanging: true` and per-file
`javascriptIdentical`: true for types.ts, false for checker.ts. Complete checker
JavaScript is compared in both comment modes against exactly the two rewrites
within that owner. The independent verifier checks the same source and emitted
scope without using the planner. Historical type-only claims above concern
the original slot widening; they do not classify this fold.

## Historical fold validation and stop point

The initial stop at ad79f64ac was not a claim that the task's done condition
was met. The completed continuation below supersedes that stop. Base is origin/main `7e3c0682da96cbe01f7ae9b9820f1bd8f1993671`.
Only this adaptation directory changes. The TypeScript input is exactly
`050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3).

| Required check | Result |
| --- | --- |
| 1. Full apply before and after, both with --build-product | Both exit 0 |
| 2. Paired full-tree replay, continuing after refusals | Pass; every outcome matches except 95h; both 95h runs refuse at createProgram |
| 3. Adamic types, all 79 compiler roots after | Exit 0; zero diagnostics |
| 4. All nine emitted JavaScript files and both public declarations | FAIL for JavaScript scope; declarations byte-identical |
| 5. Mutants | Pass; both half-adapted states and changed third member refuse without writes |
| 6. Lane and all oracle shards | Not run; stopped at check 4 |

The full source comparison finds only the two checker initializers changed;
both trees already contain the four slot repairs. The isolated prefix test
also verifies the four declaration repairs, both bound value types, idempotence,
--check, and the core.ts symbol guard with a foreign imported callee. The
independent source verifier passes both the prefix and full-tree comparisons.

The ordinary build uses esbuild with target es2020 and node14.17. All emitted
changes within createBasicNodeBuilderModuleSpecifierResolutionHost match exactly
esbuild's lowering of the two authorized rewrites, including its temporary
variables and nested temporary rename. But both _tsc.js and typescript.js also
rename the binder's unrelated local maybeBind2 function to maybeBind3, at its
declaration and its two calls. Those hunks are outside the pinned owner. The
other seven JavaScript files are byte-identical. No allowance for this outside
rename was added, and no unrelated source or build setting was changed to hide
it. This fails the requested strict emitted-output scope, so the oracle and
shard acceptance are not claimed.

All hunks are in [fold-evidence/emitted.diff](fold-evidence/emitted.diff).
[fold-evidence/emitted-comparison.json](fold-evidence/emitted-comparison.json)
records every emitted file and both declarations. The declarations have the
same before/after SHA-256 values:

- typescript.d.ts: c4edc2ad37cf1afeb4704ccb4bacf7c76b9f538a8e5685e4b276e2e7dd52796b
- tsserverlibrary.d.ts: cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd

[fold-evidence/checks.json](fold-evidence/checks.json) is the check ledger.
Paired replay records, exact raw stdout/stderr and their diff are retained
alongside compressed command logs. The before createProgram refusal expects
505bb77211eb4818ad15ed0701812c6c51eb06059becdd73923a5f3e73b69152 and observes
729a032f2ab8b7149432d2725cdc921422210dd6850b6cf450999fae56863854; the after
refusal has exactly the same expected/actual hashes. No earlier checker-owner
refusal occurs. Product keys also hash later evidence additions; the completed
source trees remain the ones measured here.

Reproduction after the setup in task kpp0f8s, with NODE_PATH set to this
checkout's stage3/api/node_modules:

```sh
# BEFORE_REPO is an isolated checkout of the base above.
# BEFORE and AFTER are complete --build-product applies of the two checkouts.
# PREFIX is main's refs/stage3/95-checker-axolotl-h-before archive.
node stage3/adapt/95-checker-axolotl-h/contract.cjs PREFIX NEW_CONTRACT_OUTPUT
node stage3/adapt/95-checker-axolotl-h/verify.cjs PREFIX NEW_CONTRACT_OUTPUT
node stage3/adapt/95-checker-axolotl-h/verify.cjs BEFORE AFTER
node stage3/adapt/95-checker-axolotl-h/compare.cjs BEFORE AFTER
python3 -B stage3/adapt/95-checker-axolotl-h/replay.py BEFORE_REPO AFTER_REPO BEFORE AFTER NEW_REPLAY_OUTPUT
# compare.cjs currently exits 1 for the outside-owner bundle rename.
```

## Accepted binding normalization, task kpp0f8s continuation

The user accepted esbuild collision renumbering once proved to be only a rename.
Check 4 now passes after normalizing these two bindings in each of _tsc.js and
typescript.js:

| Binding scope | Before | After | Before / after occurrences |
| --- | --- | --- | --- |
| Binder local function | maybeBind2 | maybeBind3 | 3 / 3: declaration and two direct calls |
| getPackageJsonInfoCache arrow temporary | _a | _a2 | 3 / 3: declaration, assignment and call receiver |

Every occurrence resolves through the TypeScript symbol table to its unique
declaration. Corresponding use contexts are identical after renaming. Both
functions have no escape or alias use of the renamed binding, and no .name read.
All nine bundles are scanned for string/template literals and property names
mentioning either spelling of either rename; none exists. The binder's former
name is absent after, and its new name is absent before. The temporary is scoped
to its own arrow, whose complete text is identical after name normalization.

[fold-evidence/rename-map.json](fold-evidence/rename-map.json) lists every rename
and every occurrence, with locations, roles, contexts and counts.
[fold-evidence/normalized-comparison.json](fold-evidence/normalized-comparison.json)
contains the symbol and observation proof, nine-file comparison and unchanged
public declarations. [fold-evidence/normalized-emitted.diff](fold-evidence/normalized-emitted.diff)
shows all remaining hunks: exactly the two initializer rewrites with the ordinary
upstream target's optional-chain lowering. No other rename is allowed.

rename-proof.cjs refuses unexpected uses, name observations, count changes and
missed occurrences. rename-mutant.cjs proves refusal for a string literal, a
property name, a .name read and an unrenamed call. compare.cjs accepts an optional
third argument naming a new directory in which to retain the normalized bundles.
The raw stop evidence above remains available; the normalized result supersedes
its check 4 failure. The completed oracle evidence is below.

## Completed fold validation

Checks 1 through 6 are complete. Check 4 uses only the proved, scope-specific
renames above. After that normalization, all nine emitted JavaScript files
differ only by the two rewrites and their ordinary target lowering. Both public
declarations remain byte-identical. All earlier source, type, ownership, replay
and mutant checks remain applicable: the fresh oracle artifact was compared
byte for byte with the proven after tree, across all 746 source files, all nine
compiler JavaScript files and both public declarations.

The official stage3/lane/run.sh, all 398 unchanged manifest shard commands,
and the official merge_shards.py pass on the same verified artifact. Counts
are exactly Linux main's 106,366 passing, one sanctioned Public APIs failure,
and zero pending. Merge verifies all 106,367 actual test identities with
multiplicity, aggregate counts, the exact failure title and API baseline path,
and the complete raw baseline diff bytes against the whole run.

Oracle inputs were frozen at commit a35f12a5410ab79d759daed08daf8f1cac0c2183
in an isolated checkout. This keeps the broad artifact key stable while the
shipping branch records and pushes each finished evidence piece. The artifact
key is 2bd2e2f4c33d019df5647a0dbc087e13c86974dd4147446d73f5d78bfe855e35.
No lane, shard, expectation, sanction, or other adaptation code was changed.

Preparation took 284.032 seconds and the whole lane took 277.498 seconds.
Each of the ten shard pieces was committed and pushed plainly as it finished.
Their walls are recorded below. At most two shard commands ran concurrently.
Every piece used a 1,440-second outer timeout, and every shard passed the
unchanged 30-second limit. The official merge also completed with exit 0 under
a 1,440-second outer timeout. No test process remains running.

| Piece | Shards, one-based | Seconds |
| --- | --- | ---: |
| 1 | 1 through 40 | 530.833 |
| 2 | 41 through 80 | 524.639 |
| 3 | 81 through 120 | 527.248 |
| 4 | 121 through 160 | 529.911 |
| 5 | 161 through 200 | 530.377 |
| 6 | 201 through 240 | 530.987 |
| 7 | 241 through 280 | 524.010 |
| 8 | 281 through 320 | 524.215 |
| 9 | 321 through 360 | 518.611 |
| 10 | 361 through 398 | 506.334 |

[fold-evidence/oracle/summary.json](fold-evidence/oracle/summary.json) records
piece commits, archive hashes, walls, the frozen input identity and final
counts. [fold-evidence/oracle/merge.json](fold-evidence/oracle/merge.json) is the
unchanged official merge verdict. Per-piece archives retain every shard's
report, observations, raw worker output, logs and baseline diff. The whole-run
profile and raw logs are retained alongside its original lane verdict.

Reproduction uses the unmodified lane from the frozen checkout, with the
pinned stage3/api install linked into it, Node 24.19.0 and the pinned TypeScript
source set in ADAMIC_TYPESCRIPT_SOURCE. Each command's output goes to its own
log. Prepare the artifact, run the whole reference, then use oracle-piece.py
with starts 0, 40, 80, 120, 160, 200, 240, 280, 320 and 360:

```sh
python3 -B FROZEN/stage3/lane/prepare_shards.py ARTIFACT_CACHE
bash FROZEN/stage3/lane/run.sh WHOLE --artifact ARTIFACT
python3 -B stage3/adapt/95-checker-axolotl-h/oracle-piece.py FROZEN ARTIFACT RESULTS EVIDENCE PIECE START
STAGE3_ARTIFACT=ARTIFACT python3 -B FROZEN/stage3/lane/merge_shards.py RESULTS --whole WHOLE --output MERGE_JSON
```

The final ledger is [fold-evidence/checks.json](fold-evidence/checks.json).
Historical stop logs remain intact for review; the normalized comparison and
completed oracle records above are the final results.
