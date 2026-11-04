import json,pathlib,collections,hashlib
R=pathlib.Path('/workspace/scratch')
def records(p): return [json.loads(s) for s in p.read_text().splitlines()]
head=records(R/'head-results/latent.jsonl'); replay=records(R/'accessor-results/latent.jsonl'); targets=set(json.loads((R/'accessor-files.json').read_text()))
assert head[0]==replay[0]
expected=[f for r in head[1:] for f in r.get('findings',[]) if f['phase']!='lowering' and r['file'] in targets]
actual=[f for r in replay[1:] for f in r.get('findings',[]) if f['phase']!='lowering']
a=json.dumps(expected,sort_keys=True,separators=(',',':')).encode();b=json.dumps(actual,sort_keys=True,separators=(',',':')).encode()
assert a==b, 'selected lowering findings changed in accessor replay'
reason='an object refinement using an open numeric enum as a literal tag'
open_sites={(f['where'],f['unit'],f['phase']) for f in expected if f['reason']==reason}
base=records(R/'base-results/latent.jsonl'); base_open={f['where'] for r in base[1:] for f in r.get('findings',[]) if f['reason']==reason}
p=R/'accessor-results/trace.jsonl'; trace=records(p) if p.exists() else []
trace=[t for t in trace if (t['where'],t['unit'],t['phase']) in open_sites]
by_name=collections.defaultdict(set); by_stage=collections.defaultdict(set); pairs=collections.defaultdict(set)
for t in trace:
 by_name[t['name']].add(t['where']); by_stage[t['stage']].add(t['where']); pairs[(t['name'],t['stage'])].add(t['where'])
sites={t['where'] for t in trace}
def short(v):
 if isinstance(v,str): return v.replace(str(R/'adapted')+'/', '')
 if isinstance(v,list): return [short(x) for x in v]
 if isinstance(v,dict): return {k:short(x) for k,x in v.items()}
 return v
result=dict(unique_sites=len(sites),base_open_sites=len(sites&base_open),by_name={k:len(v) for k,v in sorted(by_name.items())},by_stage={k:len(v) for k,v in sorted(by_stage.items())},name_stage=[dict(name=n,stage=s,sites=len(v)) for (n,s),v in sorted(pairs.items())],sites=short(sorted(sites)),trace=short(trace),selected_files=len(targets),matched_findings=len(expected),canonical_findings_sha256=hashlib.sha256(a).hexdigest(),method='Complete program, original refusal scans for files with head open-tag sites; unit lowering omitted; selected scan findings canonically byte-identical to original head.')
(R/'accessor-summary.json').write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps({k:v for k,v in result.items() if k not in ['trace','sites']},indent=2))
