"""Append C without regenerating or changing any accepted A/B evidence."""
import pathlib,json,hashlib,difflib,gzip,subprocess,re,collections
repo=pathlib.Path('/workspace/adamic');out=repo/'stage3/census/sys-timer-slot';s=pathlib.Path('/workspace/scratch/sys-timer-slot');parent='ef28f442f7cbc83f725da967532f5b06dfc8599e'
labels=['C-a','C-b','C-noinfer-a','C-noinfer-b'];baseSource=(s/'main/src/compiler/sys.ts').read_bytes().decode();baseHashes=json.loads((out/'source-hashes.json').read_text());baseJS=json.loads((s/'main-compiler-js.json').read_text());assert len(baseJS)==9
prefix='setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]): unknown;'
noinfer='setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;'
oldSet='setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;';oldClear='clearTimeout?(timeoutId: any): void;';newClear='clearTimeout?(timeoutId: unknown): void;'
heads=['declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;','declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;']
def load(name):return json.loads((s/name).read_text())
def cell(x):return str(x).replace('|','\\|').replace('\n',' ').replace('\r','')
def norm(where,label):
 text=where.replace(str(s/label)+'/','')
 if label.endswith('-b'):
  m=re.match(r'(src/compiler/sys\.ts:)(\d+)(:\d+)$',text)
  if m and int(m[2])>=57:text=m[1]+str(int(m[2])+2)+m[3]
 return text
def census(label):
 rows=[json.loads(l) for l in (s/(label+'-latent.jsonl')).read_text().splitlines() if l];header=rows[0];r=rows[1];sites={}
 for f in r['findings']:
  if f['kind'] not in ['Refused','NotYet']:continue
  k=(norm(f['where'],label),f['kind'],f['reason'],f['phase'])
  sites.setdefault(k,{'where':k[0],'rawWhere':f['where'],'kind':f['kind'],'reason':f['reason'],'phase':f['phase'],'units':[]})['units'].append(norm(f['unit'],label))
 for x in sites.values():x['units']=sorted(set(x['units']))
 return {'checkerDiagnostics':header['diagnostics'],'checkerRejected':header['checker_rejected'],'mode':header['latent_mode'],'units':len(r['units']),'unitStatuses':dict(collections.Counter(u['status'] for u in r['units'])),'splitCheckerBodies':[u for u in r['units'] if u['status']=='split_checker_body'],'distinctSites':len(sites),'rawEventKinds':dict(collections.Counter(f['kind'] for f in r['findings']))},sites
mainInfo,mainSites=census('main')
C={'parentEvidenceCommit':parent,'compilerBase':'d15620712c459baf72b3d359c1829df8d0808fda','typescript':'6.0.3','requestedSpelling':prefix,'additionalSpelling':noinfer,'inferenceTrials':load('C-inference.json'),'pins':load('C-pins.json'),'nodeWitness':load('C-node-witness.json'),'mainCensus':mainInfo,'variants':{}}
spans=load('C-declaration-spans.json')['trees'];
api=load('main-public-timers.json');ref=(s/'main/tests/baselines/reference/api/typescript.d.ts').read_text();referenceSystem=[x for x in api['referenceTimers'] if x.get('owner')=='System'];assert len(referenceSystem)==2
commands=[]
for name in ['C-builds','C-builds-alternative','C-checks','C-checks-alternative','C-latent','C-latent-alternative','C-witness']:commands+=load('commands-'+name+'.json')
for label in labels:
 tree=s/label;slot=noinfer if 'noinfer' in label else prefix
 expected=baseSource.replace(oldSet,slot).replace(oldClear,newClear).replace('pollScheduled: boolean;','pollScheduled: unknown;').replace('let pollScheduled: any;','let pollScheduled: unknown;')
 if label.endswith('-b'):
  for head in heads:expected=expected.replace(head+'\r\n','')
 assert (tree/'src/compiler/sys.ts').read_bytes()==expected.encode(),label
 (out/(label+'.diff.txt')).write_text(''.join(difflib.unified_diff(baseSource.splitlines(True),expected.splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts',n=0)).replace('\r\n','\n'))
 hashes={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()};assert set(hashes)==set(baseHashes)
 changed=[name for name in baseHashes if baseHashes[name]!=hashes[name]];assert changed==['src/compiler/sys.ts']
 check=load(label+'-stock.json');assert sorted((d.get('line'),d.get('column'),d['code']) for d in check['adamic'])==sorted((d.get('line'),d.get('column'),d['code']) for d in check['solution'])
 info,sites=census(label);info.update(kept=[sites[k] for k in sorted(set(sites)&set(mainSites))],gone=[mainSites[k] for k in sorted(set(mainSites)-set(sites))],new=[sites[k] for k in sorted(set(sites)-set(mainSites))])
 queried={str(n):[x for x in sites.values() if x['where'].startswith('src/compiler/sys.ts:'+str(n)+':')] for n in [1501,1502,1599,1600]}
 js=load(label+'-compiler-js.json');assert js==baseJS
 decl={'spans':spans[label]};chunks=[];publicSet=next(x['text'] for x in spans[label]['public'] if x['name']=='setTimeout')
 for name in ['typescript.d.ts','typescript.internal.d.ts']:
  before=(s/'main/built/local'/name).read_text();after=(tree/'built/local'/name).read_text();diff=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/built/local/'+name,tofile='b/built/local/'+name));chunks.append(diff)
  formattedSet=next(x['text'] for x in spans[label]['public' if name=='typescript.d.ts' else 'internal'] if x['name']=='setTimeout')
  assert re.sub(r',(?=\])','',re.sub(r'\s+','',formattedSet))==re.sub(r',(?=\])','',re.sub(r'\s+','',slot))
  assert before.replace(oldSet,formattedSet,1).replace(oldClear,newClear,1)==after,name
  decl[name]={'identical':before==after,'diff':diff,'mainSha256':hashlib.sha256((s/'main/built/local'/name).read_bytes()).hexdigest(),'candidateSha256':hashlib.sha256((tree/'built/local'/name).read_bytes()).hexdigest()}
 draft=ref
 for member in sorted(referenceSystem,key=lambda x:x['start'],reverse=True):
  before=oldSet if member['name']=='setTimeout' else oldClear;after=publicSet if member['name']=='setTimeout' else newClear;assert ref[member['start']:member['end']]==before;draft=draft[:member['start']]+after+draft[member['end']:]
 refdiff=''.join(difflib.unified_diff(ref.splitlines(True),draft.splitlines(True),fromfile='a/tests/baselines/reference/api/typescript.d.ts',tofile='b/tests/baselines/reference/api/typescript.d.ts'));chunks.append(refdiff);decl['referenceBaselineDraft']=refdiff
 (out/(label+'-declarations.diff.txt')).write_text('\n'.join(chunks))
 C['variants'][label]={'slot':slot,'deletedAmbientHeads':label.endswith('-b'),'sourceFilesCompared':len(hashes),'changedSources':changed,'sysSha256':hashes['src/compiler/sys.ts'],'checks':check,'javascript':{'identical':True,'count':len(js),'hashes':js,'partialBeforeCompletion':load(label+'-partial-compiler-js.json')},'declarations':decl,'latent':info,'queriedMainLocations':queried,'commands':[c for c in commands if c['label'].startswith(label+'-')]}
probeRows=[json.loads(l) for l in (s/'C-clear-probe-latent.jsonl').read_text().splitlines() if l];C['isolatedCancellationProbe']={'header':probeRows[0],'file':probeRows[1]}
assert not probeRows[0]['checker_rejected'];assert any(f['kind']=='Refused' and 'seen as one taking unknown' in f['reason'] for f in probeRows[1]['findings'])
C['interpretation']='Requested prefix C infers but adds two provider errors in both heads. The separately named full-tuple NoInfer alternative reaches zero stock diagnostics only without the ambient heads. Native cancellation and generic callback-view variance remain lowering blockers despite the four successful runtime witnesses.'
(out/'C.json').write_text(json.dumps(C,indent=2)+'\n')
oldMD=subprocess.check_output(['git','show',parent+':stage3/census/sys-timer-slot/REPORT.md'],cwd=repo).decode();oldJSON=json.loads(subprocess.check_output(['git','show',parent+':stage3/census/sys-timer-slot/REPORT.json'],cwd=repo));assert 'C' not in oldJSON;oldJSON['C']=C
(out/'REPORT.json').write_text(json.dumps(oldJSON,indent=2)+'\n')
lines=['','## C: prefix arguments, private storage and Node heads','','The requested prefix spelling is retained as C-a/C-b. The union spelling and a separately named full-tuple NoInfer alternative are also measured, without changing A or B. This is a post-apply scratch experiment on the same frozen compiler/adapted main as A/B, not an adaptation or Runtime landing.','','| Variant | Slot spelling | Ambient heads | Added stock diagnostics, Adamic / upstream solution | Ordinary compiler / tests exits | Nine compiler JS files identical |','| --- | --- | --- | --- | --- | --- |']
for label,row in C['variants'].items():
 exits={x['label'].removeprefix(label+'-'):x['exit'] for x in row['commands'] if x['label'].endswith(('build-compiler','build-tests'))}
 lines.append('| '+' | '.join(map(cell,[label,'full tuple + NoInfer' if 'noinfer' in label else 'prefix [...A, ...unknown[]]','deleted' if label.endswith('-b') else '41 retained',str(len(row['checks']['adamic']))+' / '+str(len(row['checks']['solution'])),exits,'yes, all 9']))+' |')
lines+=['','1. Slot and inference','','The requested `...args: [...A, ...unknown[]]` spelling infers A as the typed string/queue prefix and as [] for the two callbacks that ignore their labels. The unused labels then fit the unknown suffix. The tried union spelling `...args: A | [...A, ...unknown[]]` also accepts all three callback shapes. Both fail the provider assignment with either Node head. This is a provider compatibility failure, not failed callback inference. C-inference.json records all six spelling/head trials and their resolved call signatures.','','The additional spelling is:','```ts',noinfer,newClear,'```','It infers the complete extra-argument tuple from the arguments because NoInfer prevents the callback from fixing A prematurely. TypeScript permits a callback to read fewer tuple elements, including zero. The small probe succeeds with the @types/node head, and the full C-noinfer-b tree has zero diagnostics in both checker modes. This spelling is not silently substituted for the requested C.','','Every added diagnostic follows. Both checker modes produce the same file/line/column/code identities.']
for label,row in C['variants'].items():
 for mode in ['adamic','solution']:
  for d in row['checks'][mode]:lines.append('\n'+label+' '+mode+': `'+d['file']+':'+str(d['line'])+':'+str(d['column'])+'` TS'+str(d['code'])+': '+cell(d['message']))
lines+=['','2. Private storage and pins','','Both private annotations become unknown in every C variant: main sys.ts:258 (PollingIntervalQueue.pollScheduled) and main sys.ts:470 (the local pollScheduled). In the head-deleted variants their physical lines are 256 and 468. Every other source file remains byte-identical, across all 746 files. The drafts contain exactly these two type edits, the two public type edits, and, for (b), the two ambient declaration deletions.','','The stock AST maps the enclosing functions to createDynamicPriorityPollingWatchFile (251-386) and createFixedChunkSizePollingWatchFile (464-499). Static inspection found no direct annotation or enclosing-owner pin for either polling slot in the active CJS literals and statically required JSON manifests. This is static pin inspection, not an adapter replay. Adaptation 40\'s sys.ts class-rule pin is the unrelated stat variable; 90 has no sys.ts target, 91 targets factory/utilities, and 92/93/99-wasm guard other sys.ts owners. Adaptation 41 rules 13 and 14 explicitly pin the two ambient declarations: (b) intentionally deletes them as the requested Runtime counterfactual. No adaptation or pin is modified. See C-pins.json for the inventory and its limits.','','3. Node providers and the four census locations','','The provider probes below are stock TypeScript checks against each actual contextual System-derived method slot. Deleting the two ambient lines shifts source locations after them by two; every census comparison maps those locations back to main. Raw locations are retained in C.json.','','| Variant | Main location | Actual source line | Member | Fits tsc method slot | Full sys.ts findings at that location |','| --- | --- | --- | --- | --- | --- |']
for label,row in C['variants'].items():
 for probe in row['checks']['headProbes']:
  findings=row['queriedMainLocations'][str(probe['mainLine'])]
  description='; '.join(f['kind']+' / '+f['phase']+': '+f['reason'] for f in findings)
  if not description:description='not observed; getNodeSystem checker-split' if row['latent']['splitCheckerBodies'] else 'no finding recorded at this exact location'
  lines.append('| '+' | '.join(map(cell,[label,'sys.ts:'+str(probe['mainLine']),probe['line'],probe['name'],probe['fits'],description]))+' |')
lines+=['','With 41 retained the source signatures are `(handler: () => void, timeout: number) => NodeJS.Timeout` and `(handle: NodeJS.Timeout | undefined) => void`. Without them the actual declarations resolve to @types/node/web-globals/timers.d.ts:27 and :41: generic setTimeout with MakeVoidParameterOptional<TArgs>, and clearTimeout accepting NodeJS.Timeout | string | number | undefined. C-check.cjs records the declaration paths and complete signatures. Method bivariance admits the clearTimeout assignment; it does not make the requested generic prefix setTimeout assignment fit either provider.','','4. Cancellation witness and remaining lowering refusal','','Node '+C['nodeWitness']['node']+' clearTimeout returned undefined without throwing for the number 2147483646, the string "not-a-timer", a plain object {}, and undefined. The witness starts no timer and leaves no pending process. This establishes the requested ordinary runtime cases; it does not change @types/node\'s declared input domain.','','The requested prefix C-a/C-b streams do not record cancellation findings at the four queried locations because getNodeSystem is checker-split by the setTimeout errors. That absence is not a clearance. The independent C-clear-probe.ts.txt has zero checker diagnostics and records:']
for f in probeRows[1]['findings']:
 if f['kind']=='Refused':lines+=['','`'+f['reason']+'`.']
lines+=['','The deciding message is emitted at internal/lower/invariance.go:1042 and the Refused is returned at :1053 on the frozen compiler. C-noinfer-b removes the checker obstruction and directly measures the native cancellation view inside the full sys.ts census. Its exact four-location observations are in the table above.','','Full sys.ts census against main','','| Variant | Checker diagnostics | Units | Distinct sites | Kept | Gone / not observed | New |','| --- | --- | --- | --- | --- | --- | --- |']
lines.append('| main | 0 | '+str(mainInfo['units'])+' | '+str(mainInfo['distinctSites'])+' | '+str(mainInfo['distinctSites'])+' | 0 | 0 |')
for label,row in C['variants'].items():
 l=row['latent'];lines.append('| '+' | '.join(map(cell,[label,len(l['checkerDiagnostics']),l['units'],l['distinctSites'],len(l['kept']),len(l['gone']),len(l['new'])]))+' |')
lines+=['','The key is (main-mapped location, kind, exact reason, phase). Boundary events are excluded. Every kept, gone and new site is listed with its reason in [C-LATENT.md](C-LATENT.md), with raw events and recording units retained in C.json and the compressed JSONL streams. Checker-split bodies, deleted ambient units and generic deferrals are retained explicitly, so a missing observation is not credited as successful lowering. In particular, requested C is blocked only in getNodeSystem, while C-noinfer-b is checker-clean. The removed private any annotation no longer produces the old sys.ts:470:9 `a value of type any` finding; this is separate from the unchanged callback/host/native boundaries.','','The checker-clean NoInfer variant also adds private-poller NotYet sites at main-mapped 490:9 and 496:37: `a union of differently held members variable a function value captures`, raised at internal/lower/locals.go:208. At 497:25 it adds `a call through a function value whose rest parameter isn\'t an array`, raised at internal/lower/expression.go:1704. These are lowering observations exposed by the type-only storage/tuple changes, not JavaScript changes. All nine new sites, including the four provider variance refusals, are retained in the full ledger.','','JavaScript and declaration diffs','','The ordinary failed builds emit eight compiler JavaScript artifacts; the existing upstream lssl writer supplies the ninth, exactly as documented for A/B. C-noinfer-b emits all nine in its ordinary successful build. Every variant matches every one of the nine main hashes byte for byte. Declaration bundles are generated from the .d.ts files emitted by stock 6.0.3. The complete public/internal and reference-baseline draft hunks are in each C-*-declarations.diff.txt. Only System\'s two timer members move. NoInfer replaces built public 8545-8546, built internal 16663-16664 and reference 8491-8492 in place. For the requested prefix spelling, the real declaration bundler formats the tuple across eight setTimeout lines: old public 8545-8546 becomes 8545-8553, old internal 16663-16664 becomes 16663-16671, and draft reference 8491-8492 becomes 8491-8499. These are actual formatted output hunks, not the one-line source spelling. The two internal private polling annotations and the ambient deletions do not add a declaration delta. Public and internal declarations do move; JavaScript does not.','','Reproduction and scope','','Use the existing setup and frozen canonical/main trees. Run C-inference.cjs, C-prepare.py, C-run.py C-checks/C-builds/C-latent/C-witness, C-alternative.py, C-run.py C-checks-alternative/C-builds-alternative/C-latent-alternative, C-pins.cjs, C-declaration-spans.cjs, and C-report.py. NODE_PATH names the installed stage3/api/node_modules and the shell sources /workspace/adamic-tools/env.sh. C-run.py uses the existing bounded runner, stock check options and stock API solution graph; each child\'s real exit is retained. LATENT_FULL=1, LATENT_ONLY_FILE selects sys.ts, and LATENT_ASSERT_NO_OUTPUT=1 remains enabled. The old compiler binary and all pre-C A/B evidence are unchanged.','','Inference: the prefix contract repairs callback inference and unknown makes both private storage annotations consistent with the opaque return, but neither requested C variant is assignable to the installed setTimeout provider. The explicitly named full-tuple NoInfer alternative reaches zero stock diagnostics and identical JavaScript with the ambient heads removed. The Node cancellation witness supports an unknown host input for the four tested runtime values, while the narrower published Node declaration still produces an Adamic variance refusal. Stock-green C-noinfer-b also records the generic callback-view Refused at main-mapped 1501 and 1599: a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void. Stock-green C-noinfer-b is therefore not a lowering-ready landing.']
text='\n'.join(lines)+'\n';assert '\u2013' not in text and '\u2014' not in text;(out/'REPORT.md').write_text(oldMD+text)
full=['# C full sys.ts site ledger','','All locations below are mapped to main; raw source locations and unit identities remain in C.json. Gone means not observed, with checker eligibility caveats in REPORT.md.']
for label,row in C['variants'].items():
 full+=['','## '+label,'','| State | Main location | Kind | Phase | Exact reason |','| --- | --- | --- | --- | --- |']
 for state in ['kept','gone','new']:
  for site in row['latent'][state]:full.append('| '+' | '.join(map(cell,[state,site['where'],site['kind'],site['phase'],site['reason']]))+' |')
(out/'C-LATENT.md').write_text('\n'.join(full)+'\n')
for p in s.iterdir():
 if not p.is_file():continue
 if p.name.startswith('C-') and p.suffix in ['.json','.jsonl','.log']:
  if p.suffix=='.json':(out/p.name).write_bytes(p.read_bytes())
  else:
   with gzip.open(out/(p.name+'.gz'),'wb') as f:f.write(p.read_bytes())
 if p.name.startswith('commands-C-'):(out/p.name).write_bytes(p.read_bytes())
print('C report rendered')
