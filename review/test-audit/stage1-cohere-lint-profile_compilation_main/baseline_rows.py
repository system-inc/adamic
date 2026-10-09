from pathlib import Path
import subprocess,json,time,os
p=Path('review/test-audit/stage1-cohere-lint-profile_compilation_main')
rows=['TestProfileCompilationUnion','TestProfileCompilationPlantedFailure','TestProfileCompilationBuildLower','TestProfileCompilationBuildC','TestProfileCompilationBuildJavaScript','TestProfileCompilationBuildNative','TestProfileCompilation_Setup','TestProfileArtifacts','TestProfileCompilation_000','TestProfileSnapshotsAgree','TestCommentFoldMutant','TestPositionIndexMutant','TestRegistrationMutant','TestFactoryHooks_Setup','TestNestedOutsideModuleCopy']
# Artifact creation precedes the snapshot consumer. Each run is bounded independently.
order=['TestProfileArtifacts']+[x for x in rows if x!='TestProfileArtifacts']
result=[]
for row in order:
 if row=='TestProfileSnapshotsAgree' and not all(Path('/workspace/u111-profile',name).exists() for name in ['scanner','counted','profiled']):
  result.append(dict(test=row,run=None,status=None,wall=0,seconds=None,cooked=False,errors=[],skipped=False,blocked='enabled but incomplete artifacts after their 90-second timeout'))
  (p/'clean-rows.json').write_text(json.dumps(result,indent=2));continue
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run','^'+row+'$'];st=time.monotonic();f=p/(f'clean-{row}-{i}.log')
  with f.open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  es=[]
  for l in f.read_text().splitlines():
   try:es.append(json.loads(l))
   except:pass
  cooked=any('test timed out' in e.get('Output','') for e in es) or r.returncode==124
  faults=[e.get('Output','').strip() for e in es if e.get('OutputType')=='error']
  seconds=next((e.get('Elapsed') for e in reversed(es) if e.get('Action')=='pass' and 'Test' not in e),None)
  result.append(dict(test=row,run=i,command=cmd,status=r.returncode,wall=time.monotonic()-st,seconds=seconds,cooked=cooked,errors=faults,skipped=any(e.get('Action')=='skip' and e.get('Test')==row for e in es)))
  (p/'clean-rows.json').write_text(json.dumps(result,indent=2))
  if faults and not cooked:
   (p/'RED_BASELINE.txt').write_text(row+'\n'+'\n'.join(faults));raise SystemExit(1)
  if cooked:break
