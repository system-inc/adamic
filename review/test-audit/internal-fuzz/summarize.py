import pathlib,json,re,statistics,collections
p=pathlib.Path('review/test-audit/internal-fuzz');rows=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]
plan=json.loads((p/'mutation-plan.json').read_text());runs=json.loads((p/'runs.json').read_text())
if (p/'lower-probe-run.json').exists():runs.append(json.loads((p/'lower-probe-run.json').read_text()))
matrix={}; evidence={}
for r in runs:
 events=[]
 for l in pathlib.Path(r['log']).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 status={row:None for row in rows}
 for e in events:
  name=e.get('Test','');root=name.split('/')[0]
  if root in rows and e.get('Action')=='fail':status[root]='fail'
  if name in rows and e.get('Action') in ('pass','skip'):status[name]=e['Action']
 if r['regex']!='.':
  row=r['regex'][1:-1]
  if row in rows and status[row] is None and 'panic:' in pathlib.Path(r['log']).read_text():status[row]='fail'
  if row in rows:matrix.setdefault(r['id'],{})[row]=status[row]
 else:matrix[r['id']]=status
 for row,v in status.items():
  if v!='fail':continue
  message=next((e.get('Output','').strip() for e in events if e.get('Test','').split('/')[0]==row and re.search(r'\w+_test.go:\d+:',e.get('Output',''))),None)
  if message is None:message=next((e.get('Output','').strip() for e in events if 'panic:' in e.get('Output','')),'failure observed; inspect log')
  evidence[(r['id'],row)]=dict(command='ADAMIC_BUILD_CACHE_DIR=/tmp/u020/cache/'+r['id']+' ADAMIC_MUTANT='+r['id']+' timeout 120 go test -json -count=1 -timeout 90s ./internal/fuzz/ -run '+r['regex']+' > '+r['log']+' 2>&1',line=message,log=r['log'])
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
seconds={}
for row in rows:
 values=[]
 for n in range(1,4):
  match=re.search(r'ok\s+\S+\s+([0-9.]+)s',(p/f'timing-{row}-{n}.log').read_text());values.append(float(match[1]))
 seconds[row]=statistics.median(values)
# M16 is a supplemental comparison removal rather than a whole statement drop.
ids=[m['id'] for m in plan if m['id']!='M16']
kills={row:[id for id in ids if matrix.get(id,{}).get(row)=='fail'] for row in rows}
files={}
for f in pathlib.Path('internal/fuzz').glob('*_test.go'):
 for n,l in enumerate(f.read_text().splitlines(),1):
  match=re.match(r'func (Test\w+)\(',l)
  if match:files[match[1]]=str(f)+':'+str(n)
oracles={
'TestOneSeedOneProgram':'Self-written same-seed equality and different-seed inequality of generated source.',
'TestGeneratedProgramsCheckAndLower':'Self-written expectation that all 60 generated programs are accepted by load.Load and lower.Lower; checks errors only, never inspects returned IR.',
'TestRegexProgramsPassTheChecker':'Self-written expectation of checker acceptance for 30 generated programs; uses the same full Generate source as the lowering row, not a regex-only corpus.',
'TestOctoberFeaturesAppear':'Self-written substring vocabulary across 200 generated seeds; no generated program execution.',
'TestShrinkKeepsOnlyWhatFails':'Self-written table.delete marker predicate, three-line bound, and original-versus-copy source check. Tests shrinker, not compiler or agreement oracle.',
'TestSharedSliceCutsShare':'Self-written sharing geometry and vocabulary. bytesShared copies an older runtime rule (64-byte minimum, quarter-owner ratio); current runtime uses header/storage ratio 8 and no minimum. It never checks real runtime sharing.',
'TestOwnershipShapes':'Self-written vocabulary/cadence over eight seeds plus checker/lowering success; no runtime ownership observations and no returned IR assertion.',
'TestJudgeReadsThePanicLine':'Self-written synthetic Run outcomes and expected Finding/Checked labels; does not run Node.',
'TestOverridesShapesAndLower':'Self-written effect/argument vocabulary across 50 seeds plus checker/lowering success; no native override execution and no returned IR assertion.',
'TestSourceTreeCutsAtItems':'Self-written exact source strings for six parsed-list deletions; no external specification checked.',
'TestReduceKeepsTheSignature':'Self-written exact non-null refusal prefix and reduced minimum, observed against own compiler; tests reducer preservation, not a witness whose sole purpose is proving the agreement checker can fail. Supplemental weakened keeps comparison survives.',
'TestExecuteCPUHelper':'Only subprocess entry, returns immediately unless parent sets ADAMIC_FUZZ_CPU_HELPER; parent TestExecuteCPULimit.',
'TestExecuteCPULimit':'Self-written TimedOut and exit expectations over Go CPU-helper and shell subprocesses; kernel CPU behavior exercised, no external copied authority. SIGXCPU check asserts timeout classification only.',
'TestFuzzerSharesRuntimeLibrary':'Source Node versus native and JavaScript backend via Try, with byte/exit comparison and sanitizers; also self-written runtime path equality and shared runtime stdout.',
'TestUndefinedNumbersShapes':'Self-written source/shape vocabulary and opt-in expectations across 40 seeds plus checker/lowering success; no execution or returned IR assertion. Vocabulary derives its shape list from same generator, so shared omissions can agree.',
}
output=[]
for row in rows:
 k=kills[row];unique=[id for id in k if sum(matrix[id].get(other)=='fail' for other in rows)==1];subs=[];verdict='untrue'
 if row=='TestExecuteCPUHelper':verdict='helper'
 elif unique:verdict='slow-worthy' if seconds[row]>60 else 'sacred'
 elif k:
  supers=[other for other in rows if other!=row and set(k)<=set(kills[other])]
  if supers:subs=[min(supers,key=lambda o:seconds[o])];verdict='subsumed'
  else:subs=sorted({other for id in k for other in rows if other!=row and id in kills[other]});verdict='overlapping'
 probes=[id for id in matrix if id.startswith('P') and id!='PSource' and matrix[id].get(row)=='fail']
 own={'TestOneSeedOneProgram':['PGenerate'],'TestGeneratedProgramsCheckAndLower':['PLower'],'TestRegexProgramsPassTheChecker':['PGenerate'],'TestOctoberFeaturesAppear':['PGenerate'],'TestShrinkKeepsOnlyWhatFails':['PShrink'],'TestSharedSliceCutsShare':['PGenerate','PShareCuts'],'TestOwnershipShapes':['PLower'],'TestJudgeReadsThePanicLine':['PJudge'],'TestOverridesShapesAndLower':['PLower'],'TestSourceTreeCutsAtItems':['PSourceCut'],'TestReduceKeepsTheSignature':['PReduce'],'TestExecuteCPULimit':['PExecute'],'TestFuzzerSharesRuntimeLibrary':['PTry','PPrepare'],'TestUndefinedNumbersShapes':['PLower']}.get(row,[])
 vacuous=None if not own else all(matrix.get(id,{}).get(row)=='pass' for id in own) if all(matrix.get(id,{}).get(row) in ['pass','fail'] for id in own) else None
 last=evidence.get((k[-1],row)) if k else None
 obj=dict(test=row,package='internal/fuzz',file=files[row],seconds=seconds[row],oracle=oracles[row],oracle_kind=['external-run','self'] if row=='TestFuzzerSharesRuntimeLibrary' else 'self',kills=k,unique_kills=unique,last_proven_fail=(k[-1]+': '+last['line']) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(plan),eligible_mutants= len(ids),probe_kills=probes,subsumer_seconds=seconds[subs[0]] if verdict=='subsumed' else None,vacuous=vacuous,bounded=False,matrix_rows=rows,evidence=last or dict(command='16 planned whole-package runs; see runs.json',line='No production mutant failure observed' if row!='TestExecuteCPUHelper' else 'Parent: TestExecuteCPULimit'),entry_probes=own)
 if row in ['TestGeneratedProgramsCheckAndLower','TestOwnershipShapes','TestOverridesShapesAndLower','TestUndefinedNumbersShapes']:obj['vacuous_entry']='lower.Lower returns nil IR with nil error; vocabulary assertions remain active.'
 if row=='TestExecuteCPULimit':obj['vacuous_subcases']=['saturated: PExecute returns Run{}; this positive subcase still passes']
 if matrix.get('PSource',{}).get(row)=='fail':obj['diagnostic_kills']=['PSource: constructor preparation diagnostic, not a code-entry probe']
 output.append(obj)
(p/'audit.json').write_text(json.dumps(output,indent=2)+'\n')
counts=collections.Counter(x['verdict'] for x in output)
print(dict(counts));print('probes',[(o['test'],o['vacuous']) for o in output]);print('survivors',[id for id in matrix if id.startswith('M') and not any(v=='fail' for v in matrix[id].values())])
summary=['Unit u020: internal/fuzz at 7b18d0576930caca4e22ce2eef92fcf563af52d0.','15 top-level rows, no cross-top-level families, no skips; baseline passed in 18.779s.','15 menu mutants plus one supplemental comparison removal; counts: '+str(dict(counts))+'.','Empty-answer probes recorded separately; lowering vacuity is entry-specific.','Evidence: review/test-audit/internal-fuzz/ on test-audit/internal-fuzz.']
report='\n'.join(summary)+'\n\n'+json.dumps(output,indent=2)+'\n\nMutants (origin/main lines):\n\n| ID | Location | Menu/change | Failed rows |\n|---|---|---|---|\n'
for m in plan:
 failed=[row for row in rows if matrix.get(m['id'],{}).get(row)=='fail'];report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['menu']+': '+m['old'].strip().replace('|','\\|')+' -> '+m['new'].strip().replace('|','\\|')+' | '+', '.join(failed)+' |\n'
report+='\nSurvivor witnesses are in clean-witness.log and M05/M12/M14/M16-witness.log.\n'
(p/'REPORT.md').write_text(report)
