import pathlib,subprocess,os,json,time
root=pathlib.Path(__file__).resolve().parents[3];out=root/'review/compiler/fresh-guards';patch=out/'M18.diff';env=os.environ.copy();env['GOMAXPROCS']='4';rows=[]
def run(label):
 cmd=['go','test','./internal/fresh','-run','TestFreshWrites|TestFreshCorpus','-count=1','-parallel=4','-timeout=90s','-json'];start=time.monotonic()
 with (out/(label+'.jsonl')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=f,timeout=120)
 rows.append(dict(label=label,command=cmd,exit=r.returncode,seconds=time.monotonic()-start));return r.returncode
subprocess.run(['git','apply',str(patch)],cwd=root,check=True,timeout=15)
try:assert run('M18-family-mutant')!=0
finally:subprocess.run(['git','apply','--reverse',str(patch)],cwd=root,check=True,timeout=15)
assert run('M18-family-restored')==0
(out/'family-results.json').write_text(json.dumps(rows,indent=2)+'\n')
print(rows)
