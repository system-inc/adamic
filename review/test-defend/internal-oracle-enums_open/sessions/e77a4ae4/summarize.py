from pathlib import Path
import json,subprocess,shutil
root=Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-enums_open';plan=json.loads((p/'plan.json').read_text());matrix={};skips=[]
for id in ['clean']+[i for i,_ in plan['variants']]:
 failed=set();passed=set();started=set();details=[];elapsed={}
 for label in ['matrix','general','new-tests']:
  f=p/(id+'-'+label+'.log')
  for l in f.read_text().splitlines():
   try:d=json.loads(l)
   except:continue
   t=d.get('Test');a=d.get('Action');o=d.get('Output','')
   if t and a=='run':started.add(t)
   if t and a=='fail':failed.add(t)
   if t and a=='pass':passed.add(t)
   if t and a=='skip':skips.append({'variant':id,'test':t})
   if t and '.go:' in o:details.append({'test':t,'line':o.strip(),'log':f.name})
   if not t and a in ['pass','fail']:elapsed[label]=d.get('Elapsed')
 matrix[id]={'failed_rows':sorted({t.split('/')[0] for t in failed}),'passed_rows':sorted({t for t in passed if '/' not in t}),'failed_subcases':sorted(failed),'started':sorted(started),'details':details,'binary_seconds':elapsed}
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));(p/'skipped.json').write_text(json.dumps(skips,indent=2))
mutants={i:{'file_line':', '.join(f+':'+str(subprocess.check_output(['git','show','HEAD:'+f],cwd=root,text=True).splitlines().index(next(l for l in subprocess.check_output(['git','show','HEAD:'+f],cwd=root,text=True).splitlines() if a in l))+1) for f,a,b in changes),'change':'; '.join(a+' -> '+b for f,a,b in changes)} for i,changes in plan['variants']}
assigned=[('TestStage3EnumSparseArrayBoundary','TestNumericEnumNeverPinned','not defended',['D2','D3','D4']),('TestFreshWriteProbesStayRefused','TestStage3EnumBoundaries','defended',['D1']),('TestImportCycleRuntimeCalls','TestNumericEnumNeverPinned','not defended',['D5','D6','D7'])]
rows=[]
for name,prior,defense,ids in assigned:
 attempts=[{'mutant':id,**mutants[id],'rows_failed':matrix[id]['failed_rows']} for id in ids]
 found=next(d for d in matrix[ids[0]]['details'] if d['test'].startswith(name))
 rows.append({'test':name,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':prior,'defense':defense,'unique_mutant':'D1 internal/lower/fresh.go:69' if defense=='defended' else None,'attempts':attempts,'evidence':'source /workspace/adamic-tools/env.sh; python3 review/test-defend/internal-oracle-enums_open/run.py; '+found['log']+': '+found['line'],'bounded':True})
(p/'rows.json').write_text(json.dumps(rows,indent=2))
text='Fresh-write location contract defended within the bounded matrix.\nSparse-array and import-cycle rows not defended after three production attempts each.\nAll five added tests included; production sources restored; no test changed.\n\n'+json.dumps(rows,indent=2)+'\n\n'
text+='## Mutants\n\n| id | base file:line | change | failed rows |\n|---|---|---|---|\n'
for id,v in mutants.items():text+='| '+id+' | '+v['file_line']+' | '+v['change'].replace('|','\\|')+' | '+', '.join(matrix[id]['failed_rows'])+' |\n'
text+='\n## Passing rows for D1\n\n'+', '.join(matrix['D1']['passed_rows'])+'. TestNativeAgreesWithNode passed only the selected 14 fixture subcases, listed in matrix.json. Added review rows contain skips, listed in skipped.json; skipped subcases are not evidence of passing.\n\n'
text+='## Findings and limits\n\n'+(p/'NOTES.md').read_text()+'\n\n## Run times\n\n| variant | matrix binary seconds | generic fixtures binary seconds | added tests binary seconds |\n|---|---|---|---|\n'
for id,m in matrix.items():text+='| '+id+' | '+' | '.join(str(m['binary_seconds'].get(k)) for k in ['matrix','general','new-tests'])+' |\n'
text+='\nWall time, vet results, and exact selectors are in commands.json and new-commands.json. D1 is unique among observed rows, not a proof of whole-package or repository uniqueness. Diagnostic coverage is the defense; fresh-cycle acceptance was not mutated. No named prior subsumer caught the aimed faults in its assigned target. D5-D7 are also caught by TestReviewProgramsAgreeWithNode. Other witness failures may be broken preconditions.\n'
(p/'REPORT.md').write_text(text)
for f in ['runner.log','new-runner.log','new.py','summarize.py']:shutil.copy('/tmp/defend-enums/'+f,p/f)
print(json.dumps({id:m['failed_rows'] for id,m in matrix.items()},indent=2))
