#!/usr/bin/env python3
"""Keep observation counts, and adapt guard-only records to the existing pair tool."""
import collections,gzip,hashlib,json,sys
from pathlib import Path
root=Path(__file__).resolve().parent;repo=root.parents[2]
sys.path.insert(0,str(repo/'stage3/census/pair'));import pair

def load(p):
 op=gzip.open if str(p).endswith('.gz') else open
 with op(p,'rt') as f:return [json.loads(x) for x in f if x.strip()]
def path(name):
 p=root/'raw'/name
 return p if p.exists() else Path(str(p)+'.gz')
def normal_findings(rows):
 return [f for row in rows for f in row.get('findings',[]) or []]
def raw_findings(p):
 op=gzip.open if str(p).endswith('.gz') else open
 chunks=[]
 with op(p,'rt') as stream:
  for line in stream:
   marker='"findings":'
   if marker not in line:continue
   start=line.index(marker)+len(marker)
   value,end=json.JSONDecoder().raw_decode(line,start)
   chunks.append(line[start:end].encode())
 return b'\n'.join(chunks)
def raw_stream(p):
 return gzip.open(p,'rb').read() if str(p).endswith('.gz') else p.read_bytes()
def main():
 run=sys.argv[1] if len(sys.argv)>1 else 'guard'
 tree=Path('/workspace/roads-adapted');ts=repo/'stage3/api/node_modules/typescript'
 observations={};normal={};summary={};converted={};verification={}
 for side in ['parent','head']:
  observations[side]=load(path(f'{run}-{side}-records.jsonl'))
  normal[side]=load(path(f'{run}-{side}-normal.jsonl'))
  data=observations[side];errors=[r for r in data if r['kind']=='GuardError'];assert not errors,errors
  units=[r for r in data if r['kind']=='GuardUnit'];findings=[r for r in data if r['kind']=='GuardRefusal']
  assert sum(r['refusals'] for r in units)==len(findings)
  grouped=collections.defaultdict(list)
  for r in findings:
   grouped[r['file']].append({'kind':'NotYet','measurement':normal[side][0]['measurement'],'phase':'lowering','unit':r['unit'],'where':r['where'],'reason':r['reason'],'text':r['text']})
  rows=[normal[side][0]]
  for r in normal[side][1:]:
   if 'units' in r:rows.append(dict(r,findings=grouped[r['file']]))
  converted[side]=rows
  (root/f'{run}-{side}-for-pair.jsonl').write_text(''.join(json.dumps(r,separators=(',',':'))+'\n' for r in rows))
  summary[side]={'observed_units':len(units),'views_evaluated':sum(r['views'] for r in units),'raw_guard_refusals':len(findings),'units_with_refusals':sum(r['refusals']>0 for r in units),'probe_errors':len(errors),'roads':dict(collections.Counter(r['road'] for r in findings)),'graph_types_nonempty_units':sum(r['graph_types_count']!=0 for r in units)}
  if run=='guard':
   baseline=load(path(f'baseline-{side}.jsonl'));a=normal_findings(baseline);b=normal_findings(normal[side])
   enc=lambda x:json.dumps(x,separators=(',',':'),ensure_ascii=False).encode()
   assert enc(a)==enc(b),f'{side}: ordinary findings differ'
   exact_a=raw_findings(path(f'baseline-{side}.jsonl'));exact_b=raw_findings(path(f'{run}-{side}-normal.jsonl'))
   assert exact_a==exact_b,f'{side}: raw findings bytes differ'
   ba=raw_stream(path(f'baseline-{side}.jsonl'));nb=raw_stream(path(f'{run}-{side}-normal.jsonl'))
   verification[side]={'findings_byte_identical':True,'findings_count':len(a),'baseline_findings_sha256':hashlib.sha256(exact_a).hexdigest(),'seam_findings_sha256':hashlib.sha256(exact_b).hexdigest(),'entire_normal_stream_byte_identical':ba==nb,'baseline_unit_inventory_equal':[r.get('units') for r in baseline if 'units' in r]==[r.get('units') for r in normal[side] if 'units' in r]}
 delta=pair.compare(converted['parent'],converted['head'],tree,ts)
 (root/f'{run}-delta.json').write_text(json.dumps(delta,indent=2)+'\n')
 (root/f'{run}-delta.md').write_text(pair.report(delta))
 normal_stops={f['unit'] for f in normal_findings(normal['parent']) if f.get('phase') in ('lowering','deferred')}
 guard_new_units=[r for r in delta['raw_findings']['unit_first_stop_delta'] if r['base'] is None and r['head'] is not None]
 new_lowered_units=[r for r in guard_new_units if r['head']['unit'] not in normal_stops]
 summary['unit_stop_reading']={'new_guard_stopped_units':len(guard_new_units),'new_guard_stops_in_normally_finding_free_units':len(new_lowered_units),'new_guard_stops_in_already_normally_blocked_units':len(guard_new_units)-len(new_lowered_units),'changed_guard_first_stop_in_already_guard_blocked_units':sum(r['base'] is not None and r['head'] is not None for r in delta['raw_findings']['unit_first_stop_delta'])}
 (root/f'{run}-new-unit-stops.json').write_text(json.dumps(new_lowered_units,indent=2)+'\n')
 (root/f'{run}-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
 if verification:(root/'normal-findings-verification.json').write_text(json.dumps(verification,indent=2)+'\n')
 (root/f'{run}-new-records.json').write_text(json.dumps([r for r in observations['head'] if r['kind']=='GuardRefusal' and any(r['where']==n['finding']['where'] and r['reason']==n['finding']['reason'] and r['text']==n['finding']['text'] for n in delta['raw_findings']['new_findings'])],indent=2)+'\n')
 print(json.dumps({'summary':summary,'verification':verification,'deltas':{k:{x:len(delta[k][x]) for x in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']} for k in ['raw_findings','distinct_reads']}},indent=2))
if __name__=='__main__':main()
