import pathlib,re,json,subprocess,time
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');files=['private_generic_mutant','readiness','real_nested_functions','regexp_cycle','regexp_replace','representation_clock_source','runtime_last_index','scanner_nested_overload','scanner_nested_references'];rows=[]
for f in files:
 for n in re.findall(r'^func (Test\w+)\(',pathlib.Path('internal/oracle/'+f+'_test.go').read_text(),re.M):rows.append({'test':n,'file':'internal/oracle/'+f+'_test.go'})
listed=pathlib.Path('/tmp/u068-list.log').read_text().splitlines();assert len(rows)==15 and all(r['test'] in listed for r in rows)
(p/'scope.json').write_text(json.dumps(rows,indent=2));pattern='^('+'|'.join(r['test'] for r in rows)+')$';(p/'scope-pattern.txt').write_text(pattern)
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern];start=time.monotonic()
with (p/'slice-baseline.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
(p/'slice-baseline-run.json').write_text(json.dumps({'command':' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))
