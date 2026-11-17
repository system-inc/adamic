import pathlib,json,collections,hashlib
p=pathlib.Path(__file__).resolve().parent
mapping=json.loads((p/'caller-mapping.stdout').read_text());traces=[];equal={}
for file in ['namedEvaluation','esDecorators']:
 a=[json.loads(s) for s in (p/f'S1-latent-{file}.jsonl').read_text().splitlines()]
 b=[json.loads(s) for s in (p/f'S1-trace-{file}.jsonl').read_text().splitlines()]
 assert len(a)==len(b)==3 and b[-1]['status']=='relation_report'
 key=lambda f:json.dumps(f,sort_keys=True)
 equal[file]=collections.Counter(map(key,a[1]['findings']))==collections.Counter(map(key,b[1]['findings']))
 assert equal[file],file
 for s in (p/f'S1-trace-{file}.stderr').read_text().splitlines():
  if s.startswith('{'):
   x=json.loads(s)
   if x.get('kind_call_trace'):traces.append(x)
rows=[]
for c in mapping['callers']:
 loc=c['call'];site='/workspace/scratch/namedeval-kind-runtime/adapted-B/'+loc['file']+':'+str(loc['line'])+':'+str(loc['column'])
 obs=[x for x in traces if x['site']==site]
 units=c['unit'];unit='/workspace/scratch/namedeval-kind-runtime/adapted-B/'+units['file']+':'+str(units['line'])+':'+str(units['column'])
 records=[json.loads(s) for s in (p/('S1-latent-'+pathlib.Path(loc['file']).stem+'.jsonl')).read_text().splitlines()]
 findings=[f for f in records[1]['findings'] if f['unit']==unit]
 success=[x for x in obs if x['marked'] and x['checked_delta']>0 and x['error']=='']
 upstream=loc['line']+1 if loc['file'].endswith('esDecorators.ts') else loc['line']
 row={'upstreamLine':upstream,'adaptedCall':loc,'resolvedResult':c['result'],'unit':unit,'unitName':c['unitName'],'trace':obs,'successfulCheckConstructions':len(success),'unitFindings':findings}
 if success:row['verdict']='checked result constructed; whole unit remains blocked later'
 else:
  boundary=[f for f in findings if f['kind']=='Boundary' and f['where']==unit and 'function signature or prologue failed' in f['reason']]
  assert boundary,(c,findings)
  row['verdict']='call not reached: existing signature/prologue stop; no check construction observed'
 rows.append(row)
assert sum(r['successfulCheckConstructions'] for r in rows)==1
assert [r['upstreamLine'] for r in rows if r['successfulCheckConstructions']]==[1062]
rt=pathlib.Path('/workspace/scratch/namedeval-kind-runtime/S1')
result={'canonicalFindingsByteEquivalentAsRecords':equal,'scope':'Auxiliary S1-only observation of IR check construction, not execution of a native tsc program. Original method is called once; original expression/error and all no-output guards preserved. Canonical probes/measurements are not replaced.','traceBinarySha256':hashlib.sha256((rt/'latent-call-trace').read_bytes()).hexdigest(),'traceOverlaySha256':hashlib.sha256((rt/'call-trace/overlay.json').read_bytes()).hexdigest(),'callers':rows}
(p/'caller-observations.json').write_text(json.dumps(result,indent=2)+'\n')
for r in rows:print(r['adaptedCall'],r['verdict'])
print('PASS auxiliary trace findings match canonical S1 for both files')
