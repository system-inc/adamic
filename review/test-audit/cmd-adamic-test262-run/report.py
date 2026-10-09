import pathlib,json,re,statistics,csv
p=pathlib.Path('/tmp/u014');names=[l.strip() for l in (p/'list.log').read_text().splitlines() if l.startswith('Test')];ms=json.loads((p/'mutants.json').read_text());matrix=[]
def ev(file):
 a=[]
 for l in file.read_text().splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
for m in ms:
 a=ev(p/(m['id']+'-matrix.log'));failed=[x['Test'] for x in a if x['Action']=='fail' and x.get('Test') in names];observed=[x['Test'] for x in a if x['Action'] in ['fail','pass','skip'] and x.get('Test') in names];assert set(observed)==set(names),(m['id'],set(names)-set(observed))
 lines={n:[x['Output'].strip() for x in a if x.get('Test','').split('/')[0]==n and '_test.go:' in x.get('Output','')] for n in failed}
 matrix.append(dict(mutant=m['id'],failed=failed,observed=observed,skipped=[x['Test'] for x in a if x['Action']=='skip' and x.get('Test') in names],failing_lines=lines,probe=m['probe']))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
with open(p/'matrix.csv','w') as f:
 w=csv.writer(f);w.writerow(['test']+[m['mutant'] for m in matrix]);
 for n in names:w.writerow([n]+['fail' if n in m['failed'] else 'pass' for m in matrix])
seconds=[float(re.search(r'\bok\s+\S+\s+([\d.]+)s',(p/('timing-'+str(i)+'.log')).read_text()).group(1)) for i in range(3)]
name='TestProgramCPUDeadline';kills=[m['mutant'] for m in matrix if not m['probe'] and name in m['failed']];unique=[m['mutant'] for m in matrix if not m['probe'] and m['failed']==[name]];last=next(m for m in reversed(matrix) if not m['probe'] and name in m['failed']);line=last['failing_lines'][name][0]
row=dict(test=name,package='github.com/system-inc/adamic/cmd/adamic-test262',file='cmd/adamic-test262/run_test.go',seconds=statistics.median(seconds),oracle='Self-written execution-field and literal stdout expectations, using an unmodified C clock() helper. The saturated clean run consumed one CPU second in 2.582271724 wall seconds with a 2-second CPU budget. The spin result alone cannot distinguish CPU exhaustion from an early wall deadline; M3 is caught by the saturated subcase.',oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=last['mutant']+': '+line,verdict='sacred' if unique else 'untrue',subsumed_by=None,mutants_in_matrix=4,probe_kills=[m['mutant'] for m in matrix if m['probe'] and name in m['failed']],subsumer_seconds=None,vacuous=False,bounded=False,matrix_rows=[],evidence='ADAMIC_MUTANT=M3 ADAMIC_TEST262_MEASURE=1 ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > M3-matrix.log 2>&1; '+line,own_entry_probes=['E_RUN_PROGRAM'],vacuous_subcases=[])
(p/'report.json').write_text(json.dumps([row],indent=2));(p/'timings.json').write_text(json.dumps(dict(samples=seconds,median=statistics.median(seconds)),indent=2))
(p/'scope.json').write_text(json.dumps(dict(unit='u014',base=(p/'base.txt').read_text().strip(),rows=[name],file='cmd/adamic-test262/run_test.go',moved=[],vanished=[],family=False,matrix_tests=names,nproc=5,setup='warm env.sh worked, setup skipped',opt_in='ADAMIC_TEST262_MEASURE=1',skipped=[],helpers=['TestCompilerHangHelper','TestRunnerLocationHelper'],reached_functions=['init (startup guard)','runProgram','runCommand','runCommandWithLimit','limitedBuffer.Write','limitedBuffer.String']),indent=2))
print(json.dumps([row],indent=2));print('matrix',[(m['mutant'],m['failed']) for m in matrix]);print('test count',len(names))
