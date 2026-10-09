import pathlib,json,statistics,re,subprocess
p=pathlib.Path('review/test-audit/internal-oracle-stage3_front');scope=json.loads((p/'scope.json').read_text());names=scope['rows'];matrix=json.loads((p/'matrix.json').read_text());plan=json.loads((p/'plan.json').read_text());timings=json.loads((p/'timing.json').read_text());witnesses=[names[i] for i in [5,7,8,9]]
def events(file):
 r=[]
 for l in (p/file).read_text().splitlines():
  try:r.append(json.loads(l))
  except:pass
 return r
def seconds(name):
 return statistics.median([next(float(re.search(r'\s([\d.]+)s\s*$',e['Output']).group(1)) for e in events('timing-'+name+'-'+str(i)+'.log') if e.get('Output','').startswith('ok ')) for i in range(3)])
extra='TestNativeAgreesWithNode';allrows=names+[extra];medextra=seconds(extra+'-subset');table=[]
for m in plan['mutants']:
 raw=set(x for r in matrix if r['id']==m['id'] for x in r['failed_rows']);valid=raw if m['id']=='W01' else raw-set(witnesses)-{names[1]};m.update(failed_rows=sorted(raw),verdict_rows=sorted(valid));table.append(m)
files={n:next(str(f) for f in pathlib.Path('internal/oracle').glob('*_test.go') if 'func '+n+'(' in f.read_text()) for n in names}
oracles=[('Node executes source; self-written unmatched numeric-enum panic and exit 70 pin',['external-run','self']),('Subprocess protocol and self leak/sanitizer checks; parent TestFixtures family performs agreement','self'),('Node stdout plus self-written Refused/NotYet type and reason substring; M11 shows version prefix is unchecked',['external-run','self']),('Node plus self-written checked-site count, location and missing-value panic',['external-run','self']),('Node stdout/stderr/exit 70, hand-written pin; no Adamic execution',['external-run','self']),('Node source versus planted native changes; disagreement must detect each change','external-run'),('Self-written exact NotYet message, no outside authority','self'),('Node source versus inserted planted early exits','external-run'),('Node source versus planted native output changes','external-run'),('Node source versus planted native substring changes','external-run'),('Node silently drops out-of-bounds write; self-written Adamic stdout/stderr/exit 70 pins',['external-run','self']),('Independent original Platforms statistics functions on Node; compares source, native, release, JavaScript','external-run')]
report=[]
for i,n in enumerate(names):
 kills=[m['id'] for m in table if n in m['verdict_rows']];unique=[m['id'] for m in table if m['verdict_rows']==[n] and m['id']!='W01'];verdict='witness' if n in witnesses else 'helper' if i==1 else 'cannot-judge' if i==4 else 'sacred' if unique else 'subsumed' if i==10 else 'untrue';probes=['P02'] if n in witnesses else ['P01'] if i in [0,2,3,6] else ['P03','P04'] if i==11 else ['P05'] if i==10 else []
 last=kills[-1] if kills else None;line=None;command=None
 if last:
  rs=[r for r in matrix if r['id']==last and n in r['failed_rows']];r=rs[0];command=r['command'];ev=events(r['label']+'.log');out=[e.get('Output','').strip() for e in ev if e.get('Test','').split('/')[0]==n and ('_test.go:' in e.get('Output',''))];line=out[0] if out else next(e['Output'].strip() for e in ev if e.get('Test','').split('/')[0]==n and 'FAIL' in e.get('Output',''))
 row=dict(test=n,package='internal/oracle',file=files[n],seconds=seconds(n),oracle=oracles[i][0],oracle_kind=oracles[i][1],kills=kills,unique_kills=unique,last_proven_fail=(last+': '+line) if last else None,verdict=verdict,subsumed_by=[extra] if i==10 else [],mutants_in_matrix=13,probe_kills=probes,subsumer_seconds=medextra if i==10 else None,vacuous=False if probes else None,bounded=True,matrix_rows=allrows,evidence=(command+'; '+line) if last else ('go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^'+n+'$: three clean passes; '+('subprocess helper for stage3/fixtures TestFixtures family' if i==1 else 'Node-only body; no permitted Adamic mutant reaches it')))
 if i==1:row['parent']='stage3/fixtures TestFixtures family';row['verdict']=None;row['role']='helper'
 if i==4:row['reason']='Only the external Node oracle runs; mutating it is prohibited.'
 if i==10:row['subsumer_scope']='Selected 15 fixture subcases only; full TestNativeAgreesWithNode median unknown; hint based on 2 mutants.'
 report.append(row)
(p/'report.json').write_text(json.dumps(report,indent=2));(p/'mutant-table.json').write_text(json.dumps(table,indent=2));summary={ 'origin_commit':plan['origin_commit'],'nproc':5,'setup_seconds':0,'npm_seconds':json.loads((p/'initial.json').read_text())[1]['wall_seconds'],'go_switch_build_seconds':json.loads((p/'build.json').read_text()),'timing_command_wall_seconds':sum(x['wall_seconds'] for x in timings),'matrix_wall_seconds':sum(x['wall_seconds'] for x in matrix),'validation_wall_seconds':sum(x['wall_seconds'] for x in json.loads((p/'validation.json').read_text())),'baseline_seconds':92.554842502,'extra_subset_median':medextra};(p/'summary.json').write_text(json.dumps(summary,indent=2));print([(r['test'],r['seconds'],r['kills'],r['unique_kills']) for r in report]);print(summary)
