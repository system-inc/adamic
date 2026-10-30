# Phantom brand members become undefined

@system_adamic, October 29, #p7t8tmh: fix the brand types, not the void classifier. A missing brand reads as undefined; void permits a discarded value and therefore makes Node, Symbol and Program appear to hold void through parent and emitNode. Use undefined rather than never, because never properties can reduce unions and change TypeScript's own checking.

History: the October 7 ruling (#bw3xg7c) changed these 36 any-to-undefined replacements to void, using TypeScript's __escapedIdentifier precedent; today's ruling supersedes it.

The 36 declarations have 31 distinct names in eight compiler files and 27 public API lines. Every key and required/optional flag is retained. The sorted-array key includes its literal leading space. The class member remains declare.

The member audit is regenerated at [member-audit.json](evidence/brands-void/member-audit.json); the evidence directory keeps its historical name. The stock parser rejects runtime reads, writes, tests and materialized class fields. Run `brands-undefined-proof.cjs audit TREE EVIDENCE_DIRECTORY` with the pinned stock TypeScript 6.0.3 API on NODE_PATH. Its prepare/before/after modes compare an otherwise identical void baseline against undefined, including all emitted JavaScript and declaration files.

The lane generator reconstructs the same 27 public brand owners from class-rules.json and pinned pristine Git objects. ErrorCallback is the unchanged 28th adaptation-40 API declaration. The generator, normalized sanction and API-reference writer all use undefined.

## Current follow-up: all later pins inventoried

The later-pin refresh is now authorized for every brand-touched file. The complete grep listing and active source-pin inventory are in evidence/brands-undefined-3. Active fixed source hashes occur in 90 (binder.ts, checker.ts, emitter.ts, parser.ts, program.ts, symbolWalker.ts, transformers/es2017.ts), 91 (factory/nodeFactory.ts, utilities.ts), and 92 (sys.ts). Only checker.ts and utilities.ts overlap the 36 brand sites. sys.ts carries no adaptation-40 brand and 92's hashes are unchanged. Other SHA-256 matches are runtime fingerprints or historical proof artifacts, not additional fixed input-file guards.

91's utilities before pin is refreshed from 850e68859cc1d901df641b2128e7e5c7a00ed87e960025c3ca115b8bfe8786b3 to be42957f4ef90d32fcb912717ebe55d8a531820f0724ee56d24d62cc35001269; its after pin changes from fdbb9727fecbe7b7556ccf13ffc79c8dc0ade4eafb8fbf3675adca435124aa3b to e8f6c9b1c9e25531f9a904606886a597e52fa8749d9cbb78ca3cadd09fc6b36b. No 91 offsets are recorded. Undoing only the prototype brand retype independently reproduces both old hashes. 90's earlier checker hash/offset refresh is retained.

One fresh full apply passes end to end. 90, 91 and 92 own-check reruns each write zero files. 91's LanguageService registrant check passes across the complete source/test/script inventory; its scope proof accepts only the two arrow-body edits, including comments-retained and comments-removed JavaScript comparisons. Actual Node witnesses remain [1,0,1,0] before 91 and [1,1,1,1] after 91, matching main's current built runtime; the original early-stop assertion remains red after 91 as intended. Brand member audit and independent source/API retype checks pass. See preparation.json and adapt91/ under the new evidence directory.

Both complete bounded oracle runs now pass the lane's existing per-unit checks: 398 accepted units, 106,366 passing tests, one sanctioned Public APIs acknowledgement failure, zero pending. All 106,367 observed test identities match between runs (including multiplicity). Every unit's task assignment, inventory digest, counts, exit, baseline paths and failure titles is checked by the existing shard runner; the coordinator only releases completed private views after preserving reports and observations. The longest accepted unit is 12.04 seconds on control and 12.94 seconds on treatment, below 30 seconds; four units run concurrently, each with one upstream worker. Cgroup OOM and memory-limit events remain zero. The two oracle phases are sequential and both complete before either census starts.

All ten emitted JavaScript artifacts are byte-identical between control and treatment. Built public typescript.d.ts and its accepted reference each differ in exactly 27 lines, all void-to-undefined brand replacements, with no other API changes. The same 223 composed declarations are sanctioned on each side. Treatment shard bindings are regenerated for the changed sanction runtime hash only; all 398 assignments and expectations are unchanged. See oracle-comparison.json, api-runtime.json and shard-binding-refresh.json.

The required undo-retype mutant is rerun against the completed control byte-comparison output. Test test_undo_retype_rejected_by_lane_api_sanction invokes the real lane API guard and exits 1 for 27 changed declarations; the red line is `changed sanctioned API declaration: TypeAliasDeclaration:["ts","Path"]#0`. Its control check passes with the main sanction. See api-sanction-undo-mutant.json.

The paired first-error censuses complete on all 79 compiler files and pass the meter's existing coverage/count validator. Distinct NotYet/Refused sites fall 12,423 → 2,829 (NotYet 676 → 694; Refused 11,747 → 2,135). Void-container reads fall 9,612 → 0 and generic instantiated-at-void sites fall 80 → 0. Path value failures and the five original primitive-brand failure rows are zero on both sides. Three unrelated nonvoid-brand sites are unchanged: one __canonicalKey and two _isResolvedConfigFileName. Both loads have 144 checker diagnostics; errors and panics stay zero.

There are 211 changed reason counts and 35 new or growing reason categories, containing 100 newly observed distinct sites. Every changed reason is in [the complete table](evidence/brands-undefined-3/census-delta.md), with [all new/growing reasons and examples](evidence/brands-undefined-3/new-or-growing.md) and complete compressed ledgers beside them. Dependency skips grow 1 → 3, newly at tsbuildPublic.ts:295:1 and watch.ts:843:1; these are recorded measurement skips, excluded from refusal totals. Other reason changes are named with examples rather than claimed unchanged.

Of the previous 133 causally unclassified sites, 36 exact observations are absent from the current treatment ledger and 97 remain treatment-only. Every remaining site has a control void stop in the same attempt; 87 have one at the same diagnostic location and the other 10 at a different location within that attempt. These facts support an exposure interpretation but cannot establish that the later failure would have occurred under the old types. Absence is not proof that a construct now lowers; main changed between historical and current measurements. The complete [133-site table](evidence/brands-undefined-3/previous-133-split.md) and bounded structured context examples preserve these limits. All 31 previously known Path/primitive-brand failures are absent. The current 100 added sites comprise those 97 and three further sites: optional call at builder.ts:985:5; union/optional field read at commandLineParser.ts:3275:21; and ResolvedConfigFileName value at program.ts:1290:19.

The canonical member audit is refreshed from the fully applied tree: 31 names, 36 declarations, zero reads/writes/tests for each. Its previous snapshot is retained under evidence/brands-undefined/member-audit-oct29.json. No compiler source is changed. The census binary uses main's compiler and overlay without a cherry-pick, with LATENT_FULL=0 and LATENT_ASSERT_NO_OUTPUT=1; the oracle pair finishes before either measurement starts.

The following two sections retain the previous stops as history; current measurements are in evidence/brands-undefined-3.

## October 30 follow-up on main a108bf788a

Main now contains the symbol-keyed-brand refusal (204ee480) and required-undefined primitive phantom support (a108bf78). The retype is reapplied on a108bf788ab0cf570e025bdd7d605d5fd26ba958 without a compiler patch or measurement-only cherry-pick. Evidence from the October 29 stop is retained below and in evidence/brands-undefined; it is historical, including its then-uncommitted status and its primitive-brand refusal result.

This round refreshes adaptation 90's checker.ts before/after hashes and 18 declaration/edit offsets (+5 characters for the earlier brand annotation), retaining all 90 replacement declarations. The measured before hash is 750160c12c5f293beecdc21474e0240c1b4c3a468f1d3e93b27f533caf6e8a3b and after hash is 207c91ad384d14d0cdc1a356e4daba2bbb09b13330d0297cf6ac2232123827f1. Current measurements are saved separately under evidence/brands-undefined-2.

The planned bounded oracle uses every command in lane/shards.json via measure_shards.py and the adapted-tree product path; run.sh itself has no sharded command-line mode in this checkout. Updating the sanction changes the shard runtime identity and would require rebinding the treatment manifest without changing assignments or expectations; the treatment never reached that step. No census ran beside an oracle.

### October 30 stop: adaptation 91 also pins a retyped source

Control apply exits 0. Retyped apply passes adaptation 90 but stops at adaptation 91: `AssertionError: target source moved: src/compiler/utilities.ts`. Its `proofs/targets.json` beforeSha256 is `850e68859cc1d901df641b2128e7e5c7a00ed87e960025c3ca115b8bfe8786b3`; the observed hash is `be42957f4ef90d32fcb912717ebe55d8a531820f0724ee56d24d62cc35001269`. Undoing only `_prototypePropertyAssignmentBrand: undefined` to void restores the pin exactly. Adaptation 91 is unchanged, under the instruction to stop if anything else moves. See [adapt91-stop.json](evidence/brands-undefined-2/adapt91-stop.json).

Rerunning adaptation 90's own adapter on the tree it produced passes: zero files written, 69 declarations checked, JavaScript identical. Its hashes and offset repair are recorded in adapt90-pin.json and adapt90-own-rerun.json.

The control shard run was stopped with 152 of 398 units complete, all 152 accepted. Those units observed 34,031 passing tests, one sanctioned API failure and zero pending; these are partial counts, not a full oracle result. The retyped oracle did not start. Initial scratch-only launches were discarded after a product-layout mistake and /tmp storage exhaustion; the retained partial run used the workspace volume and removed completed private views after saving reports and observations, without changing any manifest task or acceptance check. The byte-comparison unit passes the existing 223-declaration API guard. Its report and diff are retained in evidence/brands-undefined-2.

The census binary was built from main with the meter's overlay procedure and no cherry-pick. Neither measurement was started because apply hit the new stop condition. There is no new before/after reason table and no new causal split of the previous 133 sites. The October 29 table remains historical and cannot establish the new base's result.

The real control API serves as the undo-retype mutant against the undefined sanction. `sanction-undo-mutant.py prove CONTROL_BYTE_UNIT EVIDENCE` first accepts the main sanction, then executes the existing lane API guard against the undefined sanction. Test `test_undo_retype_rejected_by_lane_api_sanction` records exit 1 and 27 changed declarations; its red line is `changed sanctioned API declaration: TypeAliasDeclaration:["ts","Path"]#0`. See api-sanction-undo-mutant.json and its complete text output.

This follow-up is saved as one WIP commit on typescript/adapt40-brands-undefined-2. A green end-to-end apply, complete oracle pair and new census result are not claimed.

## Archived October 29 stop

No compiler source or lowering behavior is changed. This proposal is blocked on current main; the completed measurements below are findings, not a green landing claim.

## Apply and gate blocker

Pinned main is 519819228d84d2a8eefce34a6bf76b79c7eeff63. The requested setup passed in 549.314 seconds. Baseline apply passed. Retyped apply and the full lane both fail in adaptation 90: `AssertionError: target declaration or source moved: src/compiler/checker.ts`. That adapter pins the complete source hash and byte offsets, so the earlier `_symbolLinksBrand` annotation changes its input. Updating adaptation 90 is outside the requested edit scope. The lane exits 1 before oracle execution; its report is retained in evidence/brands-undefined.

For the controlled census, the completed baseline tree is copied and exactly the 36 source and 27 API annotations are retyped. `retype-check.cjs` independently projects owners from the pristine pinned source, compares every compiler source byte and the complete API reference, and catches the undo-retype mutant. See retype.json, retype-mutant.json and api-retype.diff in evidence/brands-undefined. The regenerated sanction has exactly 27 normalized VoidKeyword-to-UndefinedKeyword changes, with all owner identities, other declarations and counts unchanged. All 72 existing lane guard tests pass.

The controlled tree installs and type-checks through `npm run build`, both exit 0. A standalone full eight-worker oracle was attempted concurrently with the censuses; a worker exited and the cgroup recorded an OOM kill. No completed oracle pass/fail counts or baseline-equivalence claim can be made from that attempt. Its report and test log are retained.

## C layout observation

The interface is `interface P { readonly _b: undefined; x: number }`. The cast `{ x: 7 } as P` is refused as an unchecked cast and emits no C. The proven upcast `{ _b: undefined, x: 7 } as P`, followed by an exported `p.x` read, emits a uniform `adamic_object`, not a named P struct. Its shape has two members, `_b` and `x`: `_b` occupies reference slot 0 initialized to NULL; `x` occupies number slot 1. The actual source and emitted C are retained as probe.ts and probe.c. This establishes the slot for the explicitly supplied brand; it does not establish a layout for the refused absent-brand cast. Lowering is unchanged.

## Primitive-brand control

The current `phantomField` predicate in internal/lower/view_unions_primitive_brands.go accepts void, or optional undefined. It does not accept required undefined. The identical small programs `type Brand = number & { readonly _b: TYPE }; export const value = 7 as Brand` lower with TYPE=void (exit 0) and refuse TYPE=undefined (exit 1): `a primitive brand member _b whose type is not void`. Both inputs and outputs are retained under evidence/brands-undefined. This contradicts the premise that required void and undefined brand members are treated alike; no compiler fix is attempted.

## Latent census

Both binaries use the same scratch compiler worktree, with 55c1c38fbb0840fbd4c2b93579c8722d4c1df35b cherry-picked only there. Each uses the meter's overlay generation, gofmt and Go build command, with LATENT_FULL=0 and LATENT_ASSERT_NO_OUTPUT=1. See provenance.json for exact pins, commands and binary hashes. The comparison counts unique (kind, where, reason, text) sites, as the meter does.


The first-error census completes on all 79 files: **12,826 → 3,273 distinct NotYet/Refused sites** (NotYet 811 → 840; Refused 12,015 → 2,433). Void-container element reads fall **9,610 → 0**; generic instantiations at void fall **76 → 0**. Both loads retain 144 checker diagnostics. The meter's existing lowering_summary validator accepts both complete ledgers. Error and panic sites remain zero; checker-diagnosed dependency skips grow 1 → 2, with the added skip at watch.ts:843:1 from executeCommandLine.ts:1011:1.

**Stop condition met:** 244 refusal-reason counts change and 52 reasons are new or grow. [Every changed reason](evidence/brands-undefined/census-delta.md), [every new or growing reason with example locations](evidence/brands-undefined/new-or-growing.md), and [the structured examples](evidence/brands-undefined/census-delta.json) are retained, together with both compressed raw ledgers. Tree roots are removed from embedded source locations in where/reason/text before comparison; the meter's distinct-site tuple is otherwise unchanged.

Five primitive-brand refusals grow 0 → 1 each: __pathBrand (types.ts:23:20), __compilerOptionsKey (moduleNameResolver.ts:975:33), __modeAwareCacheKey (moduleNameResolver.ts:1113:33), __incrementalBuildInfoFileIdBrand (builder.ts:1074:42), and __incrementalBuildInfoFileIdListIdBrand (builder.ts:1076:48). All say the member's type is not void. Path value refusals grow 0 → 26, and unchecked casts grow 536 → 554 (examples builder.ts:1392:49, 1403:77 and 2270:37). These observations include first-error exposure; the census alone does not classify every growing reason as a semantic regression.

Implementation stops here under the requested rule. The branch remains an uncommitted proposal, with no push and no compiler change. Adaptation 90 is unchanged. Its input hash matches again when only the checker brand retype is undone; see apply-blocker.json. A green apply/lane and unchanged full oracle baselines are not claimed.
