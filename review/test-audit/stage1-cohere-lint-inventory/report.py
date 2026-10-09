import pathlib,json,statistics,re,subprocess,shlex,time
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/u116/evidence'); pkg='stage1/cohere/lint/inventory'; rows=['TestInventoryEngine family','TestInventoryEngineUnion','TestInventoryEngineShardAssignment','TestInventoryEngineShardMutant']; members=[f'TestInventoryEngine_{i:03d}' for i in range(8)]
def events(id):
 out=[]
 for line in (ev/(id+'.log')).read_text().splitlines():
  try: out.append(json.loads(line))
  except ValueError: pass
 return out
def fail(id): return [e['Test'] for e in events(id) if e['Action']=='fail' and 'Test' in e]
def cmd(id):
 rs=json.loads((ev/'runs.json').read_text()); r=next(x for x in rs if x['id']==id)
 return ' '.join(k+'='+shlex.quote(v) for k,v in r.get('environment',{}).items())+' '+shlex.join(r['command'])+' > '+id+'.log 2>&1'
def evidence(id,test):
 es=[e['Output'].strip() for e in events(id) if e.get('Test')==test and 'Output'in e and (': case 'in e['Output'] or 'caught by []' in e['Output'] or 'invalid corpus accepted' in e['Output'] or 'repeated case' in e['Output'] or 'enumerated 7' in e['Output'])]
 if not es: es=[e['Output'].strip() for e in events(id) if e.get('Test')==test and 'Output'in e and ('--- FAIL:'in e['Output'])]
 return es[0] if es else 'no observed failure'
catalog=json.loads((ev/'catalog.json').read_text()); next(x for x in catalog if x['id']=='S1')['line']=105
next(x for x in catalog if x['id']=='M6')['supplemental']=True
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
mat=[]
for item in catalog:
 id=item['id']; raw=fail(id); grouped=sorted(set('TestInventoryEngine family' if x in members else x for x in raw)); eligible=[x for x in grouped if x!='TestInventoryEngineShardMutant'] if item['kind']=='production' and id!='M6' else []
 mat.append({'id':id,'kind':item['kind'],'raw_failed_tests':raw,'grouped_failed_rows':grouped,'production_kills':eligible,'ignored_witness_precondition_failures':[x for x in grouped if x=='TestInventoryEngineShardMutant'] if item['kind'] in ['production','probe'] else [],'package_seconds':next((e.get('Elapsed') for e in reversed(events(id)) if e['Action']in ['fail','pass'] and 'Test'not in e),None)})
(ev/'matrix.json').write_text(json.dumps(mat,indent=2))
med={}
for name in ['family','union','assignment','witness']:
 vals=[next(e['Elapsed'] for e in reversed(events(f'timing-{name}-{i}')) if e['Action']=='pass' and 'Test'not in e) for i in range(1,4)]; med[name]=statistics.median(vals)
 (ev/(name+'-timing.json')).write_text(json.dumps({'samples':vals,'median':med[name]},indent=2))
objects=[]
for idx,name in enumerate(rows):
 o={'test':name,'package':pkg,'file':'stage1/cohere/lint/inventory/'+('main_test.go' if idx==1 else 'engine_shards_test.go'),'seconds':med[['family','union','assignment','witness'][idx]],'oracle':'','oracle_kind':'self','kills':[],'unique_kills':[],'last_proven_fail':None,'verdict':['sacred','setup-check','setup-check','witness'][idx],'subsumed_by':[],'mutants_in_matrix':6,'probe_kills':[],'subsumer_seconds':None,'vacuous':None,'bounded':False,'matrix_rows':rows}
 if idx==0:
  o.update({'members':members,'oracle':'Self-written dependency, ranking, string, path and denominator assertions. TestRegisteredCorpusControl executes Go cohere no-debugger, then checks self-written counts, not diagnostic identity; a different one-diagnostic failure could satisfy this count-only control. No outside authority was copied or checked.','oracle_kind':['self','external-run'],'kills':['M'+str(i) for i in [1,2,3,4,5,7]],'unique_kills':['M'+str(i) for i in [1,2,3,4,5,7]],'supplemental_kills':['M6'],'probe_kills':['P'+str(i) for i in range(1,8)],'vacuous':False,'vacuous_subcases':['TestInventoryEngine_002 (empty shard)','TestInventoryEngine_007 (empty shard)','TestInventoryEngine_005 (empty shard)','TestInventoryEngine_006 (empty shard)'],'entry_probes':{'trace':'P1','rankings':'P2','countText':'P3','hasSelector':'P4','familyPassed':'P5','measured':'P6','measureCorpus':'P7'},'last_proven_fail':'M7: '+evidence('M7','TestInventoryEngine_000'),'evidence':cmd('M7')+'; '+evidence('M7','TestInventoryEngine_000')})
 else:
  id=['','S2','S1','W1'][idx]; o['oracle']=['','Self-written eight-entry top-level table and exact live shard-union checks.','Self-written stable hash assignment, growth, empty/repeated-corpus rejection checks.','Self-written expected singleton failure at TestUnknownFrequencyIsNotZero, with matching failure text and shard. W1 masks the subprocess error, so the expected catch disappears.'][idx]; o['last_proven_fail']=id+': '+evidence(id,name); o['evidence']=cmd(id)+'; '+evidence(id,name)
  if idx<3:o['construction_kills']=[id]
  else:o['weakened_check_kills']=['W1']
 objects.append(o)
(ev/'rows.json').write_text(json.dumps(objects,indent=2))
summary=['u116: all 11 live tests audited as four rows at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.','Clean baseline passed in 59.762s; nproc=5; no baseline skips.','Verdicts: one sacred family, two setup checks, one witness.','Six primary mutants, one supplemental mutant and seven entry probes were caught; no survivors.','Evidence: test-audit/stage1-cohere-lint-inventory, review/test-audit/stage1-cohere-lint-inventory/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(objects,indent=2)+'\n```\n\n'
report+='| ID | Origin file:line | Change | Rows failed |\n|---|---|---|---|\n'
for it in catalog:
 m=next(x for x in mat if x['id']==it['id']); change=it['after'].replace('\n',' ').replace('|','\\|'); report+=f"| {it['id']} ({it['kind']}) | {it['file']}:{it['line']} | `{change}` | {', '.join(m['grouped_failed_rows'])} |\n"
report+='\nM6 is supplemental and excluded from kills and uniqueness. Production witness failures in this table are broken preconditions and are excluded from kills and uniqueness. Only W1 decides the witness verdict. Setup edits and probes are separate evidence, not production kills. Uniqueness is package-local over four grouped rows; repository-wide replay remains central.\n\nSurvivors: none among the six primary production mutants or the supplemental M6. No equivalent-candidate claim was needed.\n\n'
report+='The brief caused the following friction and limits:\n\n- This stage1 package tests a Go inventory generator, not an Adamic native port. Its implementation lives under testdata and is compiled as a virtual file inside cohere. A plain go vet of the outer package would miss it; each standalone engine diff was vetted through its real cohere overlay boundary. No native rebuild or compiler cache isolation was needed.\n- The live list contains eight shard wrappers but only eight inner controls. Four shards are empty and pass every probe. Grouping them as one family avoids crediting empty wrappers as independent worthy tests; the empty members are recorded as vacuous subcases. The union has an extra top-level-table assertion, so it stays separate.\n- There is no single semantic entry for the family. Seven entry probes were run, one per reached entry. Every probe fails the family; setup and witness vacuity remain null because those rows do not assert a production answer.\n- The countText entry probe breaks the built-in witness literal anchor. A preliminary switch made that precondition fail even when the probe was off. Those preliminary logs are retained but not used for verdicts. The production matrix was replayed with the anchor restored, after a green switch-clean run. Only P3 retains the anchor-precondition failure, which is excluded from witness evidence.\n- The preliminary path-format mutant replaced a path with its basename, which was not a literal member of the fixed menu. It was replaced before the accepted matrix with M6, changing the path separator constant from colon-space to semicolon-space. M6 was selected after the preliminary results, so it is supplemental and no verdict rests on either path-format version.\n- The construction runner initially found two occurrences of the duplicate guard. It stopped before any construction edit; S1 was then restricted to the partition function by its exact indentation.\n- Production mutants frequently also fail the built-in witness through an unexpected failing control. Those failures do not prove the witness comparison works. Returning nil errors in W1 independently makes its expected caught list empty and proves the witness can fail.\n- The actual registered no-debugger control checks counts, not diagnostic identity. The rest of the engine expected answers are self-written. Executing a real external rule does not independently establish the complete inventory result.\n- Several generator paths are not reached by this package: load, entries, inspectBranches, statusFor, captureTests, restoreCapture, collectFiles, markdown, either main, and outer main.go run/must. They were read but not mutated. This audit judges the present tests, not complete generator coverage.\n- The whole-package cold baseline fit the 90s budget, so every accepted mutant ran the complete outer package. No narrowing, timeout, outer panic or unknown row outcome occurred. Native stage1 builds and other packages were not run.\n- The request refers to warm setup but requires npm ci even for this Go-only package. The install succeeded; no additional node_modules directories are loaded by these tests. The README describes .a checker counts as unknown, while the live inner corpus control requires a completed measurement; this audit used observed current behavior rather than that prose.\n\n'
runs=json.loads((ev/'runs.json').read_text()); measured=sum(x.get('wall',0) for x in runs); builds=[]
for log in ev.glob('*.log'):
 for line in log.read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if 'Output'in e:
   for name,sec in re.findall(r'build (inventory-engine(?:-unknown-frequency-mutant)?) private-overlay ([0-9.]+)s',e['Output']):builds.append({'log':log.name,'product':name,'seconds':float(sec)})
(ev/'builds.json').write_text(json.dumps(builds,indent=2))
report+=f'Setup: warm env.sh, no cloud setup; npm ci reported 429ms. Baseline: 59.762s in the test binary, including the cold overlay build. Timed rows used three independent -count=1 package invocations each: medians family {med["family"]:.3f}s, union {med["union"]:.3f}s, assignment {med["assignment"]:.3f}s, witness {med["witness"]:.3f}s. Recorded subsequent commands consumed {measured:.3f}s of wall time including builds, accepted and preliminary matrices, timings and vet. Per-run build durations are in builds.json; no native products were built. Total session duration is recorded when evidence is published.\n'
(ev/'REPORT.md').write_text(report)
print(json.dumps({'medians':med,'matrix':mat,'rows':objects},indent=2))
