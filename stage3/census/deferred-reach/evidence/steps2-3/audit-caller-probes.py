from pathlib import Path
import json
import sys
prefix=sys.argv[1] if len(sys.argv)>1 else 'caller-probe'
w=Path('/workspace/scratch/abrq61p');rel=lambda s:'src/'+s.split('/src/',1)[1] if '/src/' in s else s
base=[json.loads(x) for x in (w/'base.jsonl').read_text().splitlines()];probe=[json.loads(x) for x in (w/(prefix+'.jsonl')).read_text().splitlines()]
def first(rows):
 out={}
 for row in rows:
  for f in row.get('findings',[]):
   if f.get('phase') in ('lowering','deferred'):out.setdefault(rel(f['unit']),f)
 return out
bf,pf=first(base),first(probe);stages={}
for line in (w/(prefix+'-stages.jsonl')).read_text().splitlines():
 event=json.loads(line);stages.setdefault(rel(event['unit']),[]).append(event)
result=[]
for row in probe:
 for unit in row.get('units',[]):
  where=rel(unit['where']);before=bf.get(where);after=pf.get(where);key=lambda f:tuple(f.get(k) for k in ('phase','kind','where','reason','text')) if f else None
  matching=key(before)==key(after);events=stages.get(where,[])
  # Ignore preparatory errors that were swallowed by registration. A propagated
  # stop must match the observed finding text, not merely be an earlier log error.
  exits=[dict(e,declaration=rel(e['declaration']),unit=where) for e in events if after and e.get('error')==after['text']]
  result.append(dict(unit=where,baseline_stop=before,probe_stop=after,first_stop_matches=matching,matching_error_exits=exits))
(w/(prefix+'-audit.json')).write_text(json.dumps(result,indent=2)+'\n')
print('caller units',len(result),'matching',sum(r['first_stop_matches'] for r in result),'different',sum(not r['first_stop_matches'] for r in result))
print('no finding',[(r['unit']) for r in result if r['baseline_stop'] is None]);print('mismatches',[(r['unit'],r['baseline_stop'],r['probe_stop']) for r in result if not r['first_stop_matches']])
