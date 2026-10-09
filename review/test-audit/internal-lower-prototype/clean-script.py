import pathlib,json,subprocess,time,os
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-lower-prototype';files=['prototype_test.go','proven_relations_test.go','readiness_test.go','regexp_test.go','representation_clock_source_test.go','suppression_directives_test.go'];import re
rows=[t for f in files for t in re.findall(r'^func (Test\w+)\(', (r/'internal/lower'/f).read_text(),re.M)];(e/'requested-rows.json').write_text(json.dumps(rows,indent=2));runs=[]
def run(cmd,log):
 start=time.monotonic()
 with (e/log).open('w') as out:q=subprocess.run(cmd,cwd=r,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(e/'clean-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,flush=True)
run(['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile='+str(e/'slice.cover'),'./internal/lower/','-run','^('+'|'.join(rows)+')$'],'slice-coverage.log')
run(['go','tool','cover','-func='+str(e/'slice.cover')],'functions-coverage.txt')
for t in rows:
 for i in range(1,4):run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+t+'$'],'timing-'+t+'-'+str(i)+'.log')
