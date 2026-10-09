import pathlib,json,re,subprocess,time,os
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-oracle-checked_views_flag_downcast';files=['checked_views_flag_downcast_test.go','checked_views_interfaces_test.go','checked_views_objects_test.go','checked_views_optional_read_test.go','checked_views_v2_migration_test.go','checked_views_v2_object_primitive_test.go'];rows=[t for f in files for t in re.findall(r'^func (Test\w+)\(', (r/'internal/oracle'/f).read_text(),re.M)];(e/'requested-rows.json').write_text(json.dumps(rows,indent=2));runs=[]
def run(cmd,log):
 start=time.monotonic()
 with (e/log).open('w') as out:q=subprocess.run(cmd,cwd=r,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(e/'clean-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True);return q.returncode
pattern='^('+'|'.join(rows)+')$'
assert run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern],'bounded-baseline.log')==0
assert run(['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(e/'slice.cover'),'./internal/oracle/','-run',pattern],'slice-coverage.log')==0
run(['go','tool','cover','-func='+str(e/'slice.cover')],'functions-coverage.txt')
for t in rows:
 for i in range(1,4):assert run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/oracle/','-run','^'+t+'$'],'timing-'+t+'-'+str(i)+'.log')==0
