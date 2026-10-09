import pathlib,subprocess,os,time,json
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/cmd-adamic-test262-cache';env=os.environ.copy();env['ADAMIC_TEST262_MEASURE']='1';menu=json.loads((e/'menu.json').read_text());bases={f:subprocess.check_output(['git','show','origin/main:cmd/adamic-test262/'+f],cwd=r).decode() for f in ['cache.go','compiler.go','run.go']};runs=[]
def run(cmd,log):
 start=time.monotonic()
 with (e/log).open('w') as f:q=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));print(log,q.returncode,flush=True);(e/'finish-runs.json').write_text(json.dumps(runs,indent=2));return q.returncode
w=r/'cmd/adamic-test262/u012_observation_test.go'
try:
 for f,s in bases.items():assert (r/'cmd/adamic-test262'/f).read_text()==s
 w.write_text((e/'survivor-witness.go.txt').read_text())
 for mid,test in [('clean-limit','TestU012LimitObservation'),('M17','TestU012LimitObservation'),('clean-context','TestU012ContextObservation'),('M20','TestU012ContextObservation')]:
  m=next((x for x in menu if x['id']==mid),None)
  if m:
   f=m['file'].split('/')[-1];(r/m['file']).write_text(bases[f].replace(m['old'],m['new'],1))
  run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^'+test+'$'],'witness-'+mid+'.log')
  if m:(r/m['file']).write_text(bases[f])
finally:
 for f,s in bases.items():(r/'cmd/adamic-test262'/f).write_text(s)
 if w.exists():w.unlink()
run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],'final-baseline.log')
run(['go','vet','./cmd/adamic-test262/'],'final-vet.log')
run(['git','diff','--check'],'diff-check.log')

for test in ["TestEditCacheSeparation","TestLoweringSourceEdit"]:
 for i in range(3):run(["timeout","120","go","test","-count=1","-timeout","90s","./cmd/adamic-test262/","-run","^"+test+"$"],test+"-timing-"+str(i)+".log")
