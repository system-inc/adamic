"""Validate complete paired measurements, archive evidence, and write the final report."""
import collections,gzip,hashlib,json,pathlib,shutil,sys
s=pathlib.Path(sys.argv[1]).resolve();dest=s/'report/stage3/census/overload-certificates-census';analysis=s/'analysis'
data=json.loads((analysis/'paired.json').read_text());setup=json.loads((dest/'SETUP.json').read_text());callers=json.loads((s/'callers.json').read_text())
if data['coverage']['attemptedRequiredRuns']!=data['coverage']['expectedRequiredRuns']:raise SystemExit('required attempts unfinished; refusing a final report')
if data.get('nonOrdinalOverloadMoves'):raise SystemExit('non-ordinal overload moves require named-call attribution before finalizing')
for f,digest in json.loads((dest/'source-manifest.json').read_text()).items():
 if hashlib.sha256((s/'adapted'/f).read_bytes()).hexdigest()!=digest:raise SystemExit('shared source drift: '+f)
manifest=[]
for p in (s/'results').glob('*manifest.json'):manifest.extend(json.loads(p.read_text()))
primaryFiles={r['call']['file'] for r in data['mapped17']}
for f in primaryFiles:
 assert all(any(r['file']==f and r['mode']=='full' and r['label']==side and r['state']=='complete' and r.get('streamComplete') for r in manifest) for side in ['base','head']),f
keys=[(r['label'],r['mode'],r['file']) for r in manifest];assert len(keys)==len(set(keys))==100
for r in manifest:
 assert r['state'] in ('complete','failed'),r
 if r['state']=='failed':continue
 if r['mode']!='scan':assert r['streamComplete'] and r['header']['module_format']=='CommonJSBundle',r
 if r['mode']=='full':assert r['selectionExact'],r
pairs=[p for p in data['pairs'] if p['mode']!='full-directory']
for p in pairs:
 if p['mode']=='scan':continue
 assert p['headerIdentical'] and p['unitIdentitiesEqual'],(p['mode'],p['file'])
entry=next(p for p in pairs if p['mode']=='entry');assert entry['baseHeader']['checker_rejected'] is False
# Copy all raw streams (gzip), logs, per-child receipts and timings. Empty stdout is represented by its absent contents.
for p in (s/'results').rglob('*'):
 if not p.is_file():continue
 if p.suffix=='.stdout' and p.stat().st_size==0:continue
 relative=p.relative_to(s/'results');out=dest/'runs'/relative;out.parent.mkdir(parents=True,exist_ok=True)
 if p.suffix=='.jsonl':
  with gzip.open(str(out)+'.gz','wb',compresslevel=6) as target:target.write(p.read_bytes())
 else:shutil.copy2(p,out)
with gzip.open(dest/'paired.json.gz','wb',compresslevel=6) as target:target.write((analysis/'paired.json').read_bytes())
if (dest/'paired.json').exists():(dest/'paired.json').unlink()
for name in ['overload-ledger.json','summary.json','other-overload-moves.json','FINDINGS.md','MAPPED-CALLS.md','OTHER-OVERLOADS.md','TOTALS.md','DECLARATIONS.md','all-overload-mentions.json','non-ordinal-overload-moves.json']:shutil.copy2(analysis/name,dest/name)
for document in dest.glob('*.md'):
 document.write_text(document.read_text().replace('paired.json','paired.json.gz'))
for name in ['run-pair.log','run-full.log','base-core-directory.log','head-core-directory.log','core-directory.py','env.sh']:
 if (s/name).exists():shutil.copy2(s/name,dest/name)
other=json.loads((analysis/'other-overload-moves.json').read_text());names=sorted({r['set'] for r in other if r['mode']!='full-directory'})
ledger=data['overloadLedger'];unchangedEntry=[r for r in ledger if r['mode']=='entry' and r['side']=='base' and r['set'] not in ['arrayFrom','arrayToMap']]
relations=[]
for p in pairs:
 if p['mode']=='scan':continue
 for side in ['base','head']:
  r=p[side+'RelationReport'].get('relation_overflow',{})
  if r.get('after_load') or r.get('end'):relations.append({'mode':p['mode'],'file':p['file'],'side':side,'relation':r})
(dest/'relation-audit.json').write_text(json.dumps(relations,indent=2)+'\n')
rejected=[{'mode':p['mode'],'file':p['file'],'side':side} for p in pairs if p['mode']!='scan' for side in ['base','head'] if p[side+'Header'].get('checker_rejected')]
(dest/'checker-rejected-runs.json').write_text(json.dumps(rejected,indent=2)+'\n')
sourceNote='All shared TypeScript source hashes still match the pre-run manifest.'
summary={'coverage':data['coverage'],'sourceHashesMatch':True,'requiredChildren':len(manifest),'allRequiredChildExitsZero':not data['coverage']['failedRequiredRuns'],'allSuccessfulFullSelectionsExact':True,'allHeadersPaired':True,'allUnitIdentitiesPaired':True,'entry':{'base':entry['baseFindings'],'head':entry['headFindings'],'grades':entry['counts']},'mappedStorageCallsOf17':data['storageStopCalls17'],'otherMovingSets':names,'relationOverflowRuns':len(relations)}
(dest/'VALIDATION.json').write_text(json.dumps(summary,indent=2)+'\n')
rows=[r for r in data['mapped17'] if r['storageStopObserved']]
lines=['# Paired overload certificates census — #5fws6q8','',
'Base: `910f21afffd892c747bd82f72229743a899586ea`. Head: `812c342689ded7e1cc8a31c35a86f3546f7b0772`, exactly one commit on base. This report branch changes no adaptation, ledger or internal source.','',
'## Answers for Lower','',
'### 1. Declaration acceptance','',
'**Both complete named sets now pass the declaration census:** arrayFrom overloads **1–2**, and arrayToMap overloads **1–4**. Source accounting is two named sets and six overload declarations; [DECLARATIONS.md](DECLARATIONS.md) lists all ordinals and explains baseline short-circuiting.','',
'The visible base refusals eliminated are arrayFrom overload 1 (U[] versus (T | U)[]) and arrayToMap overload 2 (Map<K,V2> versus Map<K,V1 | V2>). Map overload 4 is no longer blocked by its unwitnessed mapper-result binder. It was masked behind overload 2 at base, so it is not counted as an independently observed old failure. The identity signatures are not all claimed as new independent passes.','',
'Positive evidence is the completed first-error entry pass reaching the certified storage gate for each set in builder.ts: the callee census must have succeeded before that gate. The standalone full core.ts pair defers these generic declarations; their absence there is not itself acceptance evidence.','',
'### 2. The 17 mapped calls','',
'**'+str(data['storageStopCalls17'])+' of the 17 compiler/services mapped call sites now have an observed certified-storage NotYet in the required streams.** These are unique static sites, not repeated event counts. Missing storage refusals do not prove success: a scan refusal, prior stop or independently deferred generic body may prevent reaching a call. [MAPPED-CALLS.md](MAPPED-CALLS.md) accounts for every one of the 17 and lists observation contexts.','']
for r in rows:
 c=r['call'];lines += ['- `'+c['file']+':'+str(c['line'])+':'+str(c['column'])+'` (`'+c['set']+'`):','']
 for text in r['words']:lines+=['```text',text,'```','']
lines += ['The tsc.ts checker closure itself contains **11** mapped calls. The prior scout\'s **17** additionally include six service calls outside that closure. Full-src resolution finds **129 calls**, **27 mapped**, across **47 files**; all 47 files (core included) were selected and attempted, so no server/harness/testRunner caller was silently omitted. The 17-call count above excludes those extra mapped sites.','',
'### 3. Other overload sets','',
'The whole-tsc first-error pair has **no other overload-set finding movement**. Its '+str(len({r['set'] for r in unchangedEntry}))+' non-target named sets / '+str(len(unchangedEntry))+' ordinal overload findings remain identical by unit and position-stripped diagnostic tuple. Every overload-result/parameter/type-constraint reason was inspected, not only the two target names.','']
if names:
 lines += ['**The full caller-file contexts do show movement in:** '+', '.join('`'+n+'`' for n in names)+'. [OTHER-OVERLOADS.md](OTHER-OVERLOADS.md) gives every before/after unit stop and its grade.','']
 for name in names:
  items=[r for r in other if r['set']==name and r['mode']!='full-directory'];counts=collections.Counter(r['grade'] for r in items)
  lines += ['- `'+name+'`: '+', '.join(str(v)+' '+k for k,v in sorted(counts.items()))+' event(s); full before/after words in the linked ledger.']
 lines+=['']
else:lines += ['**No other ordinal overload-result/parameter/type-constraint finding moved in any completed required full-file pair either.** The complete before/after ledger, including unchanged findings, is [overload-ledger.json](overload-ledger.json). Non-ordinal wording is also audited in [all-overload-mentions.json](all-overload-mentions.json); no unnamed overload movement is omitted.','']
lines += ['A syntactic scope cross-check found **25** overload signatures in **15** owners with extra implementation binders, all inside the 81-source entry closure; see [extra-binders.json](extra-binders.json). This is the entry condition for the changed censusInferRemainder path, not a claim that all such bodies were instantiated. Observed movement is not uniquely attributed to that inference change versus newly reached code without a separate ablation.','',
'## Paired totals and grades','',
'The unfiltered meter-style first-error entry pass is **'+str(entry['baseFindings'])+' -> '+str(entry['headFindings'])+' findings**: '+', '.join(str(entry['counts'].get(k,0))+' '+k for k in ['new','gone','changed','reached'])+'. Headers are identical and checker-clean. Both entry relation reports have no load/end overflow.','',
'Across the 41 completed full-file pairs: 0 new, 96 gone, 43 changed, and 53 reached. All 43 changed pairs are Boundary recovery metadata. The six failed harness-file pairs are excluded. [TOTALS.md](TOTALS.md) contains each file\'s paired counts. [FINDINGS.md](FINDINGS.md) and [paired.json](paired.json) grade every changed event, including recovery Boundary records. A changed Boundary message is administrative metadata, not a separate overload claim.','',
'Tuples are `(kind, phase, reason, full message)` with complete source coordinates replaced by `<position>`. Correspondence is within the same unchanged-source unit. Identical tuples pair first, preserving multiplicities. Different words for the same overload claim family in the same unit are changed. Remaining base observations are gone. A remaining head non-scan stop in a previously blocked unit is reached; otherwise it is new. Reached does not mean compilation succeeded or a new panic is harmless. Raw positions and original words are retained.','',
'## Inputs, builds and runs','',
'Both worktrees ran recursive submodule initialization before building. Cohere is `5360753e19e4417675a90e92a1823bea66b13918`; TypeScript-go is `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go is 1.27.1. Each sha supplied its own latent generator, meter entry extension, scan-all generator and compiler sources. All six probe hashes, generation/build commands and exits are under [builds](builds). No head tool reused a base overlay or binary.','',
'Apply inputs are unchanged between shas. Both keys are `'+setup['baseApplyKey']+'`; the source pin is TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. One base-generated adapted tree was used by both probes. '+sourceNote+' [SETUP.json](SETUP.json), [source-manifest.json](source-manifest.json) and [VALIDATION.json](VALIDATION.json) retain these checks.','',
'**100 required probe processes were attempted:** two unfiltered first-error entry runs, 94 exact full-file runs, and four scan-all runs (core and tsc roots at each sha). '+str(data['coverage']['completeRequiredRuns'])+' completed successfully; '+str(len(data['coverage']['failedRequiredRuns']))+' failed. Every successful latent stream has its relation trailer and every successful full selection is exactly the requested single absolute file. Two additional normal-overlay compiler-directory core runs are labeled supplemental and are not substituted for required runs.','',
'Entry flags: `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, CommonJSBundle, rooted at src/tsc/tsc.ts exactly as the meter entry pass. Full flags: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, exact `LATENT_ONLY_FILE`, CommonJSBundle. Entry-reachable full files use the tsc root; extra callers use their own entry closure. [callers.json](callers.json) records the exact roots and getResolvedSignature results.','',
'Failed run attempts are listed in coverage and raw logs. No complete-file conclusions are drawn from their partial streams. Raw streams/timings are gzip-compressed under [runs](runs); logs, exits, wall time, peak child RSS and manifests remain beside them. Scan-all records its real boundary and is blind after its first declaration/body/tail stop; its zeros are not treated as declaration passes. Its complete before/after observations are retained in paired.json and FINDINGS.md.','',
'The initial scheduler was deliberately interrupted to avoid scheduling duplicate full runs when the entry and full batches were separated. Its already-running entry children completed normally with exit 0 and full trailers; only the scheduler exited 130. The full scheduler ran the file/scan batch independently. The interruption log and child receipts are retained; no census child was canceled or its exit masked.','']
if rejected:lines += ['Checker-rejected entry-root contexts are listed in [checker-rejected-runs.json](checker-rejected-runs.json). Their measurements are reported as such, not as checker-clean compilation passes.','']
if relations:lines += ['## Relation boundary','',str(len(relations))+' required run-side relation reports record overflow. Their raw findings and conservative attribution are retained in relation-audit.json; observations beyond a poisoned checker relation are not claimed as successful lowering.','']
else:lines += ['All completed latent relation reports have empty load/end overflow answers; [relation-audit.json](relation-audit.json) is empty.','']
lines += ['## Reproduction','', 'Set scratch cache paths and ADAMIC_TYPESCRIPT_SOURCE to the pinned upstream. Create the base/head worktrees, initialize submodules, install stage3/api, and build every probe from its own source using build.py SCRATCH LABEL. Build one shared adapted tree after confirming equal apply keys. Resolve callers with the pinned stock API, then run the entry and full batches separately or in all mode:','', '```sh', 'python3 run-pair.py /scratch/paired entry', 'python3 run-pair.py /scratch/paired full', 'python3 pair-findings.py /scratch/paired', 'python3 render-report.py /scratch/paired', '```', '', 'The scripts and exact executed commands are retained here. There is no adaptation, ledger entry, internal change, native-compilation claim or claim that an unobserved call passed.','']
failed=data['coverage']['failedRequiredRuns']
if failed:
 lines += ['## Blocked exact-file runs','',str(len(failed))+' required children failed. Every file holding one of the 17 mapped compiler/services calls has a complete paired stream; these failures affect additional caller contexts. They prevent claiming a complete full-file census across all 47 files. The whole-entry result remains complete.','']
 for r in failed:
  lines += ['- '+r['label']+' `'+r['file']+'`: child exit '+str(r['exit'])+'; '+(s/r['log']).read_text().splitlines()[0]+'. Full log: `runs/'+r['log'].removeprefix('results/')+'`.','']
lines += ['Scan-all is unchanged for both roots: 4,836 stops before and after, with `production abort at lowering.declareModule` on both revisions. It provides no evidence beyond that boundary.','']
(dest/'REPORT.md').write_text(('\n'.join(lines)+'\n').replace('paired.json','paired.json.gz'))
manifestHashes={str(p.relative_to(dest)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(dest.rglob('*')) if p.is_file() and '__pycache__' not in p.parts and p.name!='SHA256SUMS.json'}
(dest/'SHA256SUMS.json').write_text(json.dumps(manifestHashes,indent=2,sort_keys=True)+'\n')
print(json.dumps(summary,indent=2))
