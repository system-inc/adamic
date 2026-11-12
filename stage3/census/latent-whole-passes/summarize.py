#!/usr/bin/env python3
import collections,gzip,json,pathlib,sys
base=pathlib.Path(sys.argv[1]);output=pathlib.Path(sys.argv[2])
def rows(p):
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as f:
  for line in f:
   if line.strip():yield json.loads(line)
passes=collections.defaultdict(lambda:dict(calls=0,wall_seconds=0.,peak_rss_kib=0,hwm_growth_kib=0,stops=[],panics=[]))
statuses=[];unit_seconds=0.;count=0;seen=set();units=[];findings=[]
for d in sorted(base.iterdir()):
 if not d.is_dir() or not (d/'status.json').exists(): continue
 statuses.append(json.loads((d/'status.json').read_text()))
 p=d/'passes.jsonl'
 if p.exists():
  for r in rows(p):
   v=passes[r['pass']];v['calls']+=1;v['wall_seconds']+=r['wall_seconds'];v['peak_rss_kib']=max(v['peak_rss_kib'],r['maxrss_kib']);v['hwm_growth_kib']+=r['hwm_growth_kib']
   if 'error' in r:v['stops'].append({k:r[k] for k in ('unit','type','error')})
   if 'panic' in r:v['panics'].append({'unit':r['unit'],'panic':r['panic']})
 p=d/'units.jsonl'
 if p.exists():
  for r in rows(p): unit_seconds+=r['wall_seconds'];count+=1
 p=d/'census.jsonl'
 if p.exists():
  for r in rows(p):
   if 'units' in r:
    for u in r['units']:
     if u['where'] in seen: raise ValueError('overlapping units: '+u['where'])
     seen.add(u['where']);units.append(u)
    findings.extend(r['findings'])
result=dict(statuses=statuses,attempted_unit_time_records=count,unit_wall_seconds=unit_seconds,units=units,passes=dict(passes),findings=findings)
output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:result[k] for k in ('attempted_unit_time_records','unit_wall_seconds')},indent=2))
for name,p in passes.items(): print(name,p['calls'],round(p['wall_seconds'],6),p['peak_rss_kib'],p['hwm_growth_kib'],len(p['stops']),len(p['panics']))
