from pathlib import Path
import json,gzip,shutil,hashlib,subprocess
s=Path('/tmp/watch-file-check-cead');out=s/'evidence-work/watch-functions-1532-evidence/cead-check-1532';out.mkdir(parents=True,exist_ok=True)
for name in ['scope-stock.cjs','scope-stock.json','contracts.cjs','contracts.json','core-contract.json','99-correctness-core-multimap-contract.json','99-correctness-single-watcher-forwarder-contract.json','99-correctness-timer-closures-contract.json','witness.cjs','witness.json','histogram.json','control.json','dynamic-control.cjs','dynamic-control.jsonl','dynamic-lane.jsonl','declarations.diff','declarations.json','interval-10.cjs','interval-10.json','proof-commands.json','final-proofs.py','retry-lane.py','bounded-lane-command.json','implementation-sha.txt','failed-lane-cleanup.json','push-evidence.py','archive-evidence.py']:
 p=s/name
 if p.exists():shutil.copyfile(p,out/name)
for name in ['baseline-build-dts.log','after-install.log','after-build.log','after-dts.log','scope-stock.log','contracts.log','types-main.log','types-after.log','witness.log','final-proofs.log','whole-lane.log','whole-lane-bounded.log','retry-lane-controller.log','push-implementation.log','memory-events-before-count.txt','memory-events-after-count.txt','memory-events-before-lane.txt','memory-events-before-bounded-lane.txt','memory-events-after-bounded-lane.txt','sys-comments-main.js','sys-comments-after.js','sys-no-comments-main.js','sys-no-comments-after.js']:
 p=s/name
 if p.exists():(out/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for folder in ['count-oracle','count-oracle-oom','lane','lane-bounded']:
 root=s/folder
 if not root.exists():continue
 for p in root.rglob('*'):
  rel=p.relative_to(root)
  if 'adapted-tree' in rel.parts or not p.is_file():continue
  dest=out/folder/(str(rel)+'.gz');dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for name in ['sys-mutant.ts','sys-count.ts']:
 p=s/name
 if p.exists():(out/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for side,root in [('main',s/'baseline-declarations'),('after',s/'after-tree/built/local')]:
 for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
  p=root/name
  if p.exists():dest=out/'declarations'/side;dest.mkdir(parents=True,exist_ok=True);(dest/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
sha=(s/'implementation-sha.txt').read_text().strip();parent=subprocess.check_output(['git','-C',str(s/'work'),'rev-parse',sha+'^'],text=True).strip();assert parent=='ceadba1d9baf55dc0917eaf27dec69ebada795db'
h=json.loads((s/'histogram.json').read_text());raw=[json.loads(x) for x in (s/'dynamic-lane.jsonl').read_text().splitlines()];assert len(raw)==6 and all(x['type']=='number' for x in raw)
modules={}
for name,p in [('main',s/'baseline-typescript.cjs'),('after',s/'after-tree/built/local/typescript.js')]:
 data=p.read_bytes();assert b'ADAMIC_DYNAMIC_COUNT' not in data and b'ADAMIC_WATCH_COUNT' not in data;modules[name]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'privateCountingInstrumentationAbsent':True}
(out/'built-node-modules.json').write_text(json.dumps(modules,indent=2)+'\n')
laneDone=(s/'ALL_PROOFS_OK').exists();laneReceipt=json.loads((s/'lane-bounded/oracle/report.json').read_text()) if (s/'lane-bounded/oracle/report.json').exists() else None
summary={'base':parent,'implementation':sha,'implementationBranch':'typescript/watch-file-check-1532-1','implementationPushed':laneDone,'census':{'new':0,'gone':1,'changed':0,'target1532Gone':True,'sibling1533Unchanged':True,'old1734LoweringStopAbsent':True},'dynamicHistogram':h,'countCase':'nonzero; not a vacuous members-only observation','countControl':'passed; 250 and 10 logged separately; 10 restored baseline TypeError','memberWitness':[250,500,2000],'inadmissibleWitness':['10','undefined'],'guardDroppingMutantRestores':'TypeError (not the refusal)','symlinkInterval10Modes':'byte-identical registration/close for polling file, fsEvents file, polling directory','stockTsc603':'0 diagnostics on 79 roots, both sides','adamicTypes':'0 diagnostics on 79 roots, both sides','fullApply':'passed','ownerContracts':'all 5 directories passed','javascript':'747 source files compared; 1208 per-file emission checks; only guard differs; guard-dropping emitted sys.js equals main','declarations':'both byte-identical; declarations.diff is empty','wholeLane':'PASS' if laneDone else 'bounded-memory full lane still in progress','laneNodeHeapMiB':1536,'laneCounts':laneReceipt.get('counts') if laneReceipt else None,'noLoom':True}
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');(out/'checkpoint.json').write_text(json.dumps({'status':'complete' if laneDone else 'pre-lane proofs complete; bounded lane in progress','base':parent,'implementation':sha,'evidenceRetainsFailedResourceRuns':True},indent=2)+'\n')
(out/'README.md').write_text('''# #cyjqp5f: dynamic watch-file interval check on ceadba1d

Implementation branch: typescript/watch-file-check-1532-1. Implementation commit: '''+sha+'''. Its sole parent is ceadba1d9baf55dc0917eaf27dec69ebada795db (#8581h73). No Loom submission. summary.json is the current proof verdict; earlier interrupted attempts are retained explicitly below.

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
''')
for name in ['make-overlay-txt.py','built-bundle-guard.json','lane-verdict-parser-correction.json','retry-lane.py']:
 p=s/name
 if p.exists():shutil.copyfile(p,out/name)
for name in ['overlay-build.log','census-build.log','adamic-build.log','push-implementation.log']:
 p=s/name
 if p.exists():(out/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
shutil.copytree(s/'overlay',out/'overlay',dirs_exist_ok=True)
assert not list((out/'overlay').glob('*.go'))
versions={'go':subprocess.check_output(['/tmp/watch-file-check-1532/toolchains/go/bin/go','version'],text=True).strip(),'node':subprocess.check_output(['node','--version'],text=True).strip(),'compilerSourceBase':'ceadba1d9baf55dc0917eaf27dec69ebada795db','measurementOverlay':'overlay/overlay.json','binaries':{name:hashlib.sha256((s/name).read_bytes()).hexdigest() for name in ['latent','adamic']}}
(out/'tool-provenance.json').write_text(json.dumps(versions,indent=2)+'\n')
if laneDone:
 with (out/'README.md').open('a') as doc:doc.write('\n## Final verdict\n\nWhole lane PASS on implementation '+sha+': 106,366 passing, one sanctioned public-API failure, zero pending; eight workers, all runners, no filter. Its only baseline diff is api/typescript.d.ts, the declaration guard passes, and memory.events is unchanged during the 1536 MiB Node-heap run. The normal lane exited zero and recorded this implementation SHA. The external controller originally assumed verdict.txt contained only PASS; it actually contains a longer PASS line. The corrected controller validates the structured report status; the original traceback and correction receipt are retained. No lane test was repeated for that parser correction. The implementation branch is pushed, without force.\n\nThe actual built typescript.js bundle also becomes byte-identical to main when only its three-line guard is removed; built-bundle-guard.json retains both hashes and the exact emitted guard.\n')
manifest={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.rglob('*')) if p.is_file() and p.name!='manifest-sha256.json'};(out/'manifest-sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
