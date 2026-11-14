# #cyjqp5f relocated dynamic check: stopped at the full-census gate

Base 8e3a784aa5a601f57342c07b68f145fe56252c21. A fresh detached scratch worktree was created and recursive shallow submodules initialized. Upstream v6.0.3 remains verified at 050880ce59e30b356b686bd3144efe24f875ebc8. Go 1.27.1 was downloaded into scratch; Go, npm and temporary caches used this attempt's scratch. The previously completed official base apply was reused as the census input: sys-main.ts was verified byte-identical to its final refs/stage3/99-wasm-crypto-import-after blob. The private counting instrumentation was removed before measurement. The baseline source was restored after every probe.

## Ownership first

The initial grep covered every stage3/adapt adapt.cjs, owners.json and targets.json. 99-correctness-single-watcher-forwarder/owners.json whole-pins createSystemWatchFunctions. 99-correctness-timer-closures/owners.json whole-pins createDynamicPriorityPollingWatchFile. The propagated guard was placed inside the latter owner. Internal type propagation also reaches the parent-directory watcher and the Node worker in getNodeSystem, whose whole pins are held by 93 and 99-wasm. No adapters or owner pins were amended because the prerequisite full-census gate failed. No implementation branch or ledger change was created.

## Probes and the complete checker-clean form

The result-only probe replaces the factory result's HostWatchFile reference with an explicit optional-number function type. It clears the outer refusal but introduces TS2322 at sys.ts:1021:9: the factory still returns an enum-only function. Its disappearance of two lowering observations is an artifact of the newly diagnosed body, not a proof of acceptance.

The second probe also widens the nested implementation parameter and checks only at the dynamic-poller call. It introduces three TS2345 diagnostics at candidate sys.ts:1032:61, :1046:21 and :1053:64, in other watch modes whose helper types still require PollingInterval. Its extra gone lowering findings likewise reflect checker-diagnosed body exclusion.

The final propagated probe makes HostWatchFile admit number | undefined, propagates that type through the general polling worker, FsWatch fallback intervals, the parent-directory watcher's ignored interval, and the Node fsWatchFileWorker. It guards defaultPollingInterval at the start of createDynamicPriorityPollingWatchFile's nested watchFile using the three enum values, otherwise returning Debug.fail with `Invalid dynamic polling interval: ${defaultPollingInterval}`. This is the sole runtime change in the source diff. No System declaration or watchDirectory function was edited. Stock tsc 6.0.3 reports zero diagnostics over the same 79 roots for both main and this final probe.

## Mandatory stop: new lowering finding

The complete main/final-probe census, with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE naming the full adapted sys.ts, completed with exit zero in CommonJSBundle mode. Both checker headers have zero diagnostics, so no diagnosed-body exclusion masks this result.

After aligning unchanged source lines, all unique finding changes are:

- GONE, refusal_scan, Refused, main sys.ts:1532:13: "a function taking PollingInterval seen as one taking number | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take".
- NEW, lowering, NotYet, probe sys.ts:1737:13 (main-aligned :1734:13): "node:fs.watchFile with a non-number interval".
- NEW, lowering, Boundary, same site: "failed statement; state rolled back; continued at next statement or declaration", carrying the same non-number-interval lowering error.
- CHANGED: none. The refusal_scan sibling at main sys.ts:1533:13 is unchanged (shifted by three source lines in the probe).

The two new unique findings each occur in three recording units; delta-propagated.json retains every complete observation with its phase, unit, reason, text and spans. Full streams and timing logs are retained, including relation-report trailers. compare-census.py uses unchanged-line alignment (and pairs equal-length replacement blocks), includes phase in finding identity, and records every new/gone/changed finding. This triggers the user's explicit instruction: "If widening it exposes a new refusal-scan or lowering finding anywhere else in sys.ts, stop and report." No additional guard was added at Node's filesystem boundary.

## Proof status

Completed: ownership census; fresh detached base worktree and submodules; pinned API install and Go toolchain; guarded full baseline and three probe censuses; stock 6.0.3 checks over 79 roots; complete raw and line-aligned deltas. The reused baseline official apply passed in the earlier attempt and its receipt remains in check-1532/apply-main.log.gz.

Not run after the mandatory stop: new dynamic-site counting oracle or control, member/non-member fixture, candidate Node equality witness, guard-dropping mutant, adamic types, candidate full apply/contracts, emitted JavaScript comparison, upstream npm ci/hereby dts declaration comparison or empty diff, and candidate whole lane/symlinkWatching cases. The new-site histogram is UNMEASURED, not zero; no vacuous members-only conclusion is claimed. The earlier public-site six calls with interval 10 are not reused as a dynamic-site histogram. No Loom submission.

## Reproduction

The source snapshots, exact probe diffs, stock scripts and generated census overlay are retained. Scripts use the recorded scratch paths; relocate those paths when replaying. make-overlay-txt.py is the base latent/make_overlay.py with generated physical Go filenames suffixed .txt and its territory set to the fresh base worktree. The overlay retains production refusal logic and disables usable IR output. Build with Go 1.27.1 using `go build -buildvcs=false -overlay=OVERLAY/overlay.json -o latent ./stage3/census/latent/tool`. Run each sys.ts snapshot at the same adapted-tree path with the flags recorded in each census result.json, using all 79 compiler roots. stock-propagated.cjs uses the existing strict compiler helper. Run compare-census.py after the streams complete. Intermediate probes are diagnostic evidence only; sys-propagated.ts.gz is the final checker-clean probe stopped by lowering.
