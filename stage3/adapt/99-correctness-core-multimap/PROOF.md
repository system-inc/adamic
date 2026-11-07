# MultiMap proof

Base: origin/main `698965fe66aed945d18efa320a594a8a8b4df2f8`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`. Setup follows 99w: source /workspace/adamic-tools/env.sh, bash cloud/setup.sh, npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source. Setup and both full applies exited 0. Production compiler sources and oracle guards are unchanged.

| Proof item | Result |
| --- | --- |
| Completed baseline checker census | 27 add calls, 5 remove calls, 24 createMultiMap calls, 36 type references, 70 flow observations; 0 value reads and 0 typed own-key enumerations |
| Approved order | 94-axolotl-property-read, 94-correctness-multimap, 95-checker-axolotl-a; lexical path sort in apply.py |
| Main full apply | Exit 0, all 84 adapters |
| Amended full apply | Exit 0, all 86 adapters; every mandatory semantic guard runs at its own checkpoint |
| Required semantic checkpoints | 97-cli-axolotl-e, 98-axolotl-c and 99-cli-perf-hooks-import pass in the real sequential apply |
| Whole-owner and import contracts | 44 new-adapter checks and 4 95h drift/half-owner checks pass; refusals write nothing |
| 95h immediate checkpoint replay | Exit 0, zero edits |
| Paired replay of every adapter | 84 versus 86; every other outcome identical after validated source-coordinate and owned-fingerprint accounting; new adapters write nothing |
| Adamic types before and after | All 79 sorted compiler roots together, exit 0 and zero diagnostics both times |
| Final source provenance | All 746 files byte-identical to the completed fresh full apply; prototype source proofs are validated |
| Original/free-body equality | Both final free helpers emit exactly the bodies counted on every oracle input |
| Full lane count before final build | 398/398 units; 4,355,837 calls, 0 counterexamples, 0 own-key enumerations |
| Exit audit | 797 recorder-3 process records, all flushed after exit listeners; zero late calls; full rerun count identical |
| Node witnesses | 18 before/after cases pass; 500 lazy/full snapshot cross-checks pass |
| Ordinary compiler and harness builds | Before and after exit 0 |
| Unmodified final stage3/lane/run.sh | Exit 0; 106,366 passes, 1 sanctioned API failure, 0 skips |
| Exact oracle union | All 398 unit identities and test-name multiset match the final whole lane; baseline.diff byte-identical |
| Source scope | 746 files checked; exactly the 94 insertion, 95h two calls/import, and 99 whole-owner/import pins differ |
| Emitted JavaScript scope | All 10 normally emitted .js files checked; only _tsc.js, run.js, typescript.js differ; unowned runtime syntax tokens are identical |
| Public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Full core.ts census | 88 distinct location/kind/reason/phase sites before, 85 after; exactly F12/F13/F14 removed, 0 new sites anywhere in core.ts; 0 checker diagnostics |
| Original-owner mutant | Restoring createMultiMap and its original interface/private-helper dependencies restores all 3 sites |

The two helpers retain the Value suffix because a final rename would touch 95h's pinned createProgram. 94 adds only the helpers. 95h owns both createProgram rewrites. 99 owns every other call plus the plain Map representation and removal of the original private functions. No adaptation edits another pinned owner without the authorized 95h amendment.

Every emitted hunk is in [javascript-hunks.diff](javascript-hunks.diff). javascript-scope.cjs locates complete owners, including esbuild's class-assignment forms and verified _ProjectService/_TestState aliases; it removes only those owners and the two corresponding helper export entries before comparing complete parsed runtime syntax, identifiers, literals and variable declaration kinds. Whitespace and comments remain visible in the raw diff. This is not a blanket identifier or hunk normalization.

Paired replay retains raw reports. Import additions move 458 report coordinates; each remapped location must name the identical complete source line. The 71-writable-views createTypeChecker fingerprint changes solely because its nested mapToTypeNodes is an owned call site; restoring just that pin restores the exact baseline hash. 95h has the same pre-existing later-tree refusal with identical actual/expected hashes and no writes on both sides; its diagnostic line numbers are mapped through unchanged adapter source lines. Its own immediate checkpoint replay succeeds. replay/summary.json, coordinate-proof.json and hash-proof.json retain these checks.

The full site census uses the same LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE=core.ts and LATENT_TIMES procedure as the requested core-sites-4186581e report. Both no-output assertions run. It independently probes the whole file, including deferred generic scans; absence of a site is not a claim of successful whole-program lowering. Historical raw measurement labels are preserved even though diagnostics=[] and checker_rejected=false. Source coordinates are assigned to containing owners and snippets; comparison preserves kind, reason and phase. All raw findings, timings and the .go.txt overlay sources are retained.

The full applies ran before the final compiler build, and the complete count gate preceded it. Intermediate checks are the actual adapter checks, not a claim that every partially adapted upstream tree already has zero ordinary TypeScript diagnostics. Later evidence files change product keys without changing the frozen transformation.

COUNT.md and MEASURE.md explain the operation table, edge witnesses and excluded timeout/disk/preparation attempts. Counts use only the passing recorder-3 union. count-evidence.tar.gz holds raw unit reports, exact observations, process records and input manifests. proof-evidence.tar.gz holds all apply/build/type/census/replay logs, the reference census report, overlays, source snapshots and final comparisons. Each archive has an internal SHA256SUMS and an external .sha256 file; preparation.tar.gz retains the historical setup checkpoint. EVIDENCE.json indexes the completed packs.

Reproduction after setup uses full --build-product applies from the isolated main checkout and this branch, prepare-count.py on a fresh scratch directory, count-all.py for the unchanged lane shards, whole-control.py for the ordinary lane and union comparison, then compare.cjs, body-proof.cjs and javascript-scope.cjs. Proof drivers and archived exact command scripts retain the paths used. All logs go to files. No Go source is added to the working compiler.
