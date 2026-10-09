import subprocess,time,pathlib,json
out=pathlib.Path('review/test-audit/stage1-cohere-lint-rules-no-unsafe-negation');rs=[]
for name,cmd in [('npm',['npm','ci','--prefix','stage3/api']),('baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/no-unsafe-negation/','-run','.'])]:
 t=time.monotonic()
 with (out/(name+'.log')).open('w') as log:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 rs.append(dict(name=name,command=cmd,exit=p.returncode,wall_seconds=time.monotonic()-t));(out/'initial.json.txt').write_text(json.dumps(rs,indent=2));print(name,p.returncode,rs[-1]['wall_seconds'],flush=True)
 if p.returncode:break
