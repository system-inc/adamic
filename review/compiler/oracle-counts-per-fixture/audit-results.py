import pathlib,json,collections,gzip
out=pathlib.Path(__file__).resolve().parent;bylog={};final={};mutants={'missing-counts-mutant','orphan-counts-mutant','changed-count-mutant','crlf-writer-mutant'}
paths=list(out.glob('*.jsonl'))+list((out/'logs-archive').rglob('*.jsonl.gz'))
for p in sorted(paths,key=lambda p:p.stat().st_mtime):
 events=[];alltests=set();label=p.name.removesuffix('.gz').removesuffix('.jsonl')
 for l in (gzip.decompress(p.read_bytes()).decode() if p.suffix=='.gz' else p.read_text()).splitlines():
  try:r=json.loads(l)
  except:continue
  if r.get('Test'):alltests.add(r['Test'])
  if r.get('Test') and r.get('Action') in ['pass','fail','skip']:
   row=dict(package=r.get('Package'),test=r['Test'],action=r['Action'],seconds=r.get('Elapsed',0));events.append(row)
   if label not in mutants:final[(row['package'],row['test'])]=dict(**row,log=p.name)
 leaves=[r for r in events if not any(x.startswith(r['test']+'/') for x in alltests)]
 bylog[p.name]=dict(results=len(events),failed=[r for r in events if r['action']=='fail'],over60=[r for r in leaves if r['action']!='skip' and r['seconds']>=60],maximum_leaf_seconds=max([r['seconds'] for r in leaves]or[0]),expected_mutant=label in mutants)
result=dict(logs=bylog,results=list(final.values()),unresolved=[r for r in final.values() if r['action']=='fail'],over60=[dict(log=k,**v) for k,s in bylog.items() for v in s['over60']]);(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print('test identities',len(final),'unresolved',len(result['unresolved']),'over60',len(result['over60']));print(json.dumps(result['unresolved'][:20],indent=2));print(json.dumps(result['over60'][:10],indent=2))
