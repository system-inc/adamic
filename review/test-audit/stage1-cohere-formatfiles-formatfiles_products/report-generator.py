import pathlib,json,statistics,re,subprocess
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_products');scope=json.loads((p/'scope.json').read_text());names=scope['names'];groups=scope['groups'];matrix=json.loads((p/'matrix.json').read_text());plan=json.loads((p/'plan.json').read_text());rowlabels=[names[0],names[1],names[2],'TestProduct_FormatfilesNative family']
def group(n):
 if n in names[3:]:return rowlabels[3]
 if n.startswith('TestThePortParsesAsGoCohereDoes'):return 'TestThePortParsesAsGoCohereDoes family'
 return n
def ev(label):
 es=[]
 for l in (p/(label+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
def secs(label):return next(float(re.search(r'\s([\d.]+)s\s*$',e['Output']).group(1)) for e in ev(label) if e.get('Output','').startswith('ok '))
rows=[];mrows=rowlabels+['TestThePortParsesAsGoCohereDoes family','TestFormatfilesShardPlantedDisagreement','TestExtAgreesWithGo']
for i,members in enumerate(groups):
 label=rowlabels[i];times=[secs('timing-'+('family' if i==3 else members[0])+'-'+str(j)) for j in range(3)];kills=sorted(set(r['id'] for r in matrix if r['id'].startswith('S') and any(n in r['failed_tests'] for n in members)));prods=sorted(set(r['id'] for r in matrix if r['id'].startswith('M') and any(n in r['failed_tests'] for n in members)));probe=['P02','P03','P05','P04'][i];observed=[r for r in matrix if r['id']==probe];vacuous=not any(n in r['failed_tests'] for r in observed for n in members) if observed else None;last=kills[-1] if kills else None;unique=[]
 for k in kills:
  fail=set(group(n) for r in matrix if r['id']==k for n in r['failed_tests'])
  if fail=={label}:unique.append(k)
 line=None;command=None
 if last:
  r=next(r for r in matrix if r['id']==last and any(n in r['failed_tests'] for n in members));command=r['command'];outputs=[e['Output'].strip() for e in ev(r['label']) if e.get('Test','').split('/')[0] in members and e.get('Output') and ('_test.go:' in e['Output'])];line=next((l for l in outputs if 'failed' in l or 'build Go oracle' in l),outputs[0])
 row=dict(test=label,package='stage1/cohere/formatfiles',file='stage1/cohere/formatfiles/formatfiles_products_test.go',members=members,seconds=statistics.median(times),trials=times,oracle='Self: successful product preparation and internal nonempty binary-path checks; no executable output comparison.' if i!=0 else 'Self: successful go test -c product construction; does not run Go cohere or check that its returned binary exists.',oracle_kind='self',kills=kills,production_kills=prods,unique_kills=unique,last_proven_fail=(last+': '+line) if last else None,verdict='setup-check' if kills else 'untrue',subsumed_by=[],mutants_in_matrix=12,probe_kills=[] if vacuous else [probe],subsumer_seconds=None,vacuous=vacuous,bounded=False,matrix_rows=mrows,evidence=(command+'; '+line) if last else None,own_probe=probe)
 rows.append(row)
(p/'report.json').write_text(json.dumps(rows,indent=2));table=[]
for m in plan:
 observations=[r for r in matrix if r['id']==m['id']];entry=m.copy();entry['failed_tests']=sorted(set(n for r in observations for n in r['failed_tests']));entry['failed_rows']=sorted(set(group(n) for n in entry['failed_tests']));table.append(entry)
(p/'mutant-table.json').write_text(json.dumps(table,indent=2));summary=dict(commit=scope['commit'],setup_seconds=0,nproc=5,baseline_binary_seconds=secs('baseline'),control_binary_seconds=secs('control'),timing_command_wall_seconds=sum(x['wall_seconds'] for x in json.loads((p/'timing.json').read_text())),matrix_wall_seconds=sum(x['wall_seconds'] for x in matrix),validation_wall_seconds=sum(x['wall_seconds'] for x in json.loads((p/'validation.json').read_text())),standalone_builds={m['id']:json.loads((p/(m['id']+'-validation.log')).read_text()) for m in plan if m['file'].endswith('.ts')});(p/'summary.json').write_text(json.dumps(summary,indent=2));print([(r['test'],r['seconds'],r['kills'],r['vacuous']) for r in rows]);print(summary)
