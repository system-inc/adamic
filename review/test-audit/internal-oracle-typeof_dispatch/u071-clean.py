import pathlib,json,re,subprocess,time,os
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-oracle-typeof_dispatch';files=['typeof_dispatch_test.go','typeof_null_test.go','unknown_test.go','view_fields_test.go','wasi_shards_test.go'];tests=[t for f in files for t in re.findall(r'^func (Test\w+)\(', (r/'internal/oracle'/f).read_text(),re.M)];assert len(tests)==13
(e/'requested-tests.json').write_text(json.dumps(tests,indent=2));assert all(t in (e/'list.log').read_text().splitlines() for t in tests)
family={'TestViewFieldReadiness family':['TestNarrowedFieldUsesSharedReadiness','TestViewFieldInheritedStaticReadiness']};groups={t:[t] for t in tests if t not in sum(family.values(),[])};groups.update(family);(e/'row-members.json').write_text(json.dumps(groups,indent=2));runs=[]
def run(cmd,log):
 start=time.monotonic()
 with (e/log).open('w') as out:q=subprocess.run(cmd,cwd=r,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(e/'clean-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True);return q.returncode
pattern='^('+'|'.join(tests)+')$'
assert run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern],'bounded-baseline.log')==0
assert run(['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(e/'slice.cover'),'./internal/oracle/','-run',pattern],'slice-coverage.log')==0
run(['go','tool','cover','-func='+str(e/'slice.cover')],'functions-coverage.txt')
cov=(e/'functions-coverage.txt').read_text();(e/'reached-functions.txt').write_text('\n'.join(l for l in cov.splitlines() if not re.search(r'\s0\.0%$',l) and not l.startswith('total:'))+'\n')
for row,ts in groups.items():
 for i in range(1,4):assert run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(ts)+')$'],'timing-'+row.replace(' ','_')+'-'+str(i)+'.log')==0
