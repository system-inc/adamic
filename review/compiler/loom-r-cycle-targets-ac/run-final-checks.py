import json,pathlib,shlex,subprocess,time
root=pathlib.Path(__file__).resolve().parents[3]
p=pathlib.Path(__file__).resolve().parent
commands=[('mutants',['python3',str(p/'run-mutants.py')],600),('admission-mutant',['python3',str(p/'run-admission-mutant.py')],180),('final-vet',['go','vet','./internal/ir','./internal/lower','./internal/oracle'],180)]
records=[]
for name,args,limit in commands:
 began=time.monotonic()
 with (p/(name+'.log')).open('w') as log:
  result=subprocess.run(['timeout',str(limit),*args],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=limit+10)
 records.append({'command':shlex.join(['timeout',str(limit),*args]),'exit':result.returncode,'seconds':round(time.monotonic()-began,3),'log':name+'.log'})
 (p/'final-commands.json').write_text(json.dumps(records,indent=2)+'\n')
 print(name,result.returncode,flush=True)
 assert result.returncode==0,name
