"""Merge disjoint attempted-unit shards, unioning generic call witnesses."""
import json,pathlib,sys,copy

def load(p):
 return [json.loads(s) for s in pathlib.Path(p).read_text().splitlines() if s.strip()]
def key(x): return json.dumps(x,sort_keys=True,separators=(',',':'))
def witness_key(x):
 s=json.dumps({k:x[k] for k in ("unit","call_site","type_arguments")},ensure_ascii=False,separators=(',',':'))
 for a,b in [("<",r"\u003c"),(">",r"\u003e"),("&",r"\u0026"),("\u2028",r"\u2028"),("\u2029",r"\u2029")]:s=s.replace(a,b)
 return s
def loc(w):
 f,l,c=w.rsplit(':',2);return f,int(l),int(c)
def merge(streams):
 first=streams[0];assert all(s[0]==first[0] for s in streams)
 assert all(s[-1]['status']=='relation_report' for s in streams)
 records={r['file']:copy.deepcopy(r) for r in first[1:-1]}
 assert all(set(records)=={r['file'] for r in s[1:-1]} for s in streams)
 for file,r in records.items():
  units={};findings={};attempts={}
  for stream in streams:
   child=next(x for x in stream[1:-1] if x['file']==file)
   assert {k:v for k,v in r.items() if k not in ('units','findings')}=={k:v for k,v in child.items() if k not in ('units','findings')}
   for u in child['units']:
    w=u['where']
    if u['status'] in ('attempted','panic'): attempts[w]=attempts.get(w,0)+1
    if w not in units: units[w]=copy.deepcopy(u);continue
    v=units[w]
    assert {k:x for k,x in u.items() if k not in ('instantiated','instantiated_by')}=={k:x for k,x in v.items() if k not in ('instantiated','instantiated_by')}
    v['instantiated']=v['instantiated'] or u['instantiated']
    witnesses={key(x):x for x in v.get('instantiated_by',[])+u.get('instantiated_by',[])}
    if witnesses:v['instantiated_by']=sorted(witnesses.values(),key=witness_key)[:5]
   for f in child['findings']:findings.setdefault(key(f),copy.deepcopy(f))
  assert all(n==1 for n in attempts.values()),attempts
  r['units']=sorted(units.values(),key=lambda u:loc(u['where']))
  # Stable grouping restores scan then source-ordered attempts; order within each
  # attempt is retained from its unique shard. Exact repeated scan rows dedup.
  r['findings']=sorted(findings.values(),key=lambda f:(f['phase']!='refusal_scan',loc(f['unit'])))
 trailers=[s[-1] for s in streams]
 assert all(not t['findings'] and not t['relation_overflow']['after_load'] and not t['relation_overflow']['end'] for t in trailers), 'stop: relation overflow or global finding in a shard'
 findings={key(f):f for t in trailers for f in t.get('findings',[])}
 trailer=dict(status='relation_report',measurement=first[-1]['measurement'],findings=list(findings.values()),shard_relation_reports=trailers)
 return [first[0],*records.values(),trailer]
if __name__=='__main__':
 output,*inputs=sys.argv[1:];rows=merge([load(p) for p in inputs]);pathlib.Path(output).write_text(''.join(json.dumps(r,separators=(',',':'))+'\n' for r in rows));print('merged',len(inputs),'shards',sum(len(r.get('units',[])) for r in rows),'units')
