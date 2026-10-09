import json,os,subprocess,time,pathlib
root=pathlib.Path(__file__).resolve().parents[3]
out=root/'review/compiler/fresh-guards'
tests={6:'TestStatAtimeSelfCycleRemainsUnproven',11:'TestReplacementCallbackEscapeRemainsUnproven',14:'TestFutureBufferOperationRemainsUnknown',15:'TestClassMethodOutsideSelfCycleRemainsUnproven',18:'TestFreshWrites____flow_testdata_mutations_a_5514557d1b47'}
env=os.environ.copy();env['GOMAXPROCS']='4'
rows=[]
def run(label,selector):
 cmd=['go','test','./internal/fresh','-run','^'+selector+'$','-count=1','-parallel=4','-timeout=90s','-json'];started=time.monotonic()
 with (out/(label+'.jsonl')).open('w') as log:
  r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=log,timeout=120)
 row={'label':label,'command':cmd,'exit':r.returncode,'seconds':time.monotonic()-started};rows.append(row);print(row,flush=True)
 return r.returncode
for n,test in tests.items():
 patch=out/f'M{n}.diff'
 subprocess.run(['git','apply','--check',str(patch)],cwd=root,check=True,timeout=15)
 subprocess.run(['git','apply',str(patch)],cwd=root,check=True,timeout=15)
 try:
  assert run(f'M{n}-mutant',test)!=0, f'M{n} survived'
  events=[json.loads(line) for line in (out/f'M{n}-mutant.jsonl').read_text().splitlines() if line.startswith('{')]
  assert any(e.get('Test')==test and e.get('Action')=='fail' for e in events),f'M{n} failed before its guard ran'
 finally:
  subprocess.run(['git','apply','--reverse',str(patch)],cwd=root,check=True,timeout=15)
 assert run(f'M{n}-restored',test)==0,f'M{n} baseline failed'
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
