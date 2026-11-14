# #cyjqp5f: dynamic watch-file interval check on ceadba1d

Implementation branch: typescript/watch-file-check-1532-1. Implementation commit: 8499a514fc1d4bbab019594dfeef0e09fc4b5bbf. Its sole parent is ceadba1d9baf55dc0917eaf27dec69ebada795db (#8581h73). No Loom submission. summary.json is the current proof verdict; earlier interrupted attempts are retained explicitly below.

## Change and ownership

The initial grep covers owners.json, targets.json and every adapt.cjs. Existing after pins are amended in 93-builtin-imports (Node worker), 99-correctness-core-multimap (ignored parent-directory interval), 99-correctness-single-watcher-forwarder (internal aliases and factory result/forwarding), and 99-correctness-timer-closures (the dynamic poller). 99-wasm-crypto-import's downstream whole getNodeSystem before/after pins carry the Node-worker widening. The 93 and wasm whole owner texts are now retained alongside their SHA-256 pins. A single JavaScript-changing ledger line cites #cyjqp5f.

The factory result explicitly takes pollingInterval?: number; general internal interval plumbing takes number | undefined. System and every watchDirectory body are byte-identical. The sole JavaScript change is at the start of createDynamicPriorityPollingWatchFile's watchFile (adapted sys.ts:253-255): the three enum members pass; every other value returns Debug.fail(`Invalid dynamic polling interval: ${defaultPollingInterval}`). The Node worker still writes an own interval key at adapted :1737, including when its value is undefined; ceadba1d supplies Node's catchable TypeError there. No extra Node-worker guard or default is introduced.

## Full census

Unmodified ceadba1d's measurement-only CommonJSBundle tool, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE naming the entire adapted sys.ts. Both runs completed with relation-report trailers and zero checker diagnostics. Full raw streams, timings, result receipts and phase-aware delta are retained. The official full after apply's sys.ts is byte-identical to the measured probe.

All finding changes: GONE refusal_scan/Refused at main-aligned sys.ts:1532:13, "a function taking PollingInterval seen as one taking number | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take". NEW: none. CHANGED: none. The refusal_scan sibling at main-aligned :1533:13 is unchanged. The former :1734:13 Node watchFile NotYet and its Boundary are absent. The guard adds three physical source lines; compare-census.py aligns unchanged lines and retains complete original observations in delta-after.json.

## Counting oracle, control, Node witness and mutant

The complete quiet private counting oracle used eight workers, all runners, no filter, light=false. It completed with 106,366 passing, the sole sanctioned public-API failure, and zero pending. Histogram at entry to the dynamic poller: 250=2, 500=4, 2000=0, undefined=0, every other value=0, total=6. All six have typeof number. This is NONZERO; the members-only observation is not vacuous. The logging copy and six raw records are retained. A separate actual built-compiler control proves the counter fires once with 250 and once with 10; 10 throws the original TypeError. The control never writes the lane log.

The pristine main and adapted built compilers run the real factory in Node with deterministic host services. Registration and close traces are byte-identical for 250, 500 and 2000. For 10 and undefined, main throws TypeError and after throws Debug Failure naming the value. Separately, the three interval-10 symlinkWatching modes (polling file, fsEvents file, polling directory) register and close with byte-identical output on both actual Node compilers. The whole unfiltered lane additionally covers the unchanged symlinkWatching tests; no test filter or skip flag is used.

The named mutant drops only the guard, retaining all widened types. Its stock-emitted sys.js is byte-identical to main in both comment modes. Its actual dynamic-poller owner is executed in Node and restores the TypeError for 10. The mutant stands in for the TypeError, not for the removed refusal.

## Remaining proofs and resource attempts

Stock TypeScript 6.0.3: zero diagnostics over the same 79 roots on main and after. Normal ceadba1d adamic types: zero on both populations. Full stage3/apply.sh passed. Existing core-multimap, single-watcher and timer contracts pass, as do 93/wasm exact-owner replay, zero-write/mtime and drift/duplicate/mixed-owner/import refusal controls. Source inventory is 747 files; 1208 stock per-file JavaScript checks cover both comment modes. Every other source file emits identical JavaScript, and removing the guard alone restores main's complete sys.js. No cast or hidden default bridges the widened interval.

Each upstream tree used its own locked npm ci. Their own node_modules/.bin/hereby dts outputs typescript.d.ts and tsserverlibrary.d.ts match byte for byte. Both pairs are retained compressed, hashes are in declarations.json, and declarations.diff is genuinely empty.

The first counting attempt overlapped full apply and was invalidated by OOM; the quiet retry passed with memory.events unchanged. The first final-lane attempt was also resource-interrupted, had no completed verdict, and left orphan workers; its logs, memory events and explicit cleanup receipt are retained. These are not successful proof runs. The unchanged lane is repeated in a fresh directory with NODE_OPTIONS=--max-old-space-size=1536, eight workers, all runners and no filter. Check summary.json and lane-bounded/ for the final verdict. No lane script, expected count, sanction, timeout, or test was changed. Earlier resource failures are preserved rather than silently replaced.

## Reproduction

Use the implementation SHA on the stated base, Node v24.19.0, Go 1.27.1, recursive shallow submodules, pinned stage3/api npm ci and upstream v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 as ADAMIC_TYPESCRIPT_SOURCE. All writes and caches were in scratch. The baseline adapter inputs are byte-identical to the prior 8e3a784aa source-adaptation inputs; the prior official apply was reused only as an input snapshot, while the compiler, full census, official after apply, both type checks, count and candidate lane were rerun on ceadba1d. The final source snapshot is independently checked against the actual official after apply.

Scripts retain their exact scratch paths and must be relocated together for replay. amend.cjs records the mechanical pin amendments; implementation.patch.gz is the complete one-commit diff. The generated census overlay uses physical .go.txt filenames; production IR output is disabled. run-census.py and compare-census.py reproduce the census. count-run.py plus post-count.py reproduce instrumentation, control and pristine baseline build. scope-stock.cjs, contracts.cjs, witness.cjs and interval-10.cjs reproduce bounded proofs. proof-commands.json and bounded-lane-command.json retain exact commands, environments, exits and elapsed times. retry-lane.py runs the ordinary lane and pushes only after PASS. Upstream bundles used for witnesses contain no private counting instrumentation; their hashes are in built-node-modules.json.
