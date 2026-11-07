import json,os,shutil,subprocess
from pathlib import Path
s=Path('/workspace/scratch/elementat-b/step-b');raw=s/'raw';before=Path('/workspace/scratch/elementat-b/before');after=s/'after';env=dict(os.environ,GOMAXPROCS='2');checks={}
def run(name,args):
 with (raw/(name+'.log')).open('w') as log:result=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,env=env,timeout=900)
 checks[name]={'exit':result.returncode};(raw/'probe-checks.json').write_text(json.dumps(checks,indent=2)+'\n');return result.returncode
for label,tree in [('before',before),('after',after)]:
 roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79;(raw/(label+'-roots.json')).write_text(json.dumps(roots,indent=2)+'\n');assert run('types-'+label,[str(s/'bin/adamic'),'types',*roots])==0
mutant=s/'mutant';shutil.copytree(after/'src',mutant/'src');file=mutant/'src/compiler/core.ts';text=file.read_bytes().decode();pins=json.loads(Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-z/owners.json').read_text())[0];newline='\r\n' if '\r\n' in text else '\n';a=pins['after'].replace('\n',newline);b=pins['before'].replace('\n',newline);assert text.count(a)==1;file.write_bytes(text.replace(a,b).encode())
for label,tree,where in [('after',after,'core.ts:1326:26'),('mutant',mutant,'core.ts:1076:100')]:
 output=raw/(label+'-binder.jsonl');assert run(label+'-binder',[str(s/'bin/tool'),str(tree/'src/compiler/binder.ts'),str(output)])==0
 rows=[json.loads(l) for l in output.read_text().splitlines()];assert len(rows)==2 and all(r['kind']=='Refused' and r['where'].endswith(where) and r['reason']=="a cast the runtime can't check" for r in rows),rows
checks['mutantRestores1076']=True;checks['afterReaches1326']=True
code=run('witness',['node',str(s/'witness.cjs'),str(before),str(after)]);assert code==1
witness=json.loads((raw/'witness.log').read_text());assert witness['allObservedClassesEqual'] and not witness['allClassesEqual'];checks['requiredAllClassWitnessEquality']=False
(raw/'probe-checks.json').write_text(json.dumps(checks,indent=2)+'\n');print(json.dumps(checks),flush=True)
