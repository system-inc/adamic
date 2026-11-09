import json,gzip,difflib,collections
from pathlib import Path
out=Path(__file__).resolve().parent;prefix='/workspace/scratch/bar-adapted/src/'
def short(v):return v.replace(prefix,'')
def read(side):
 p=out/(side+'-entry-run.jsonl');p=p if p.exists() else Path(str(p)+'.gz')
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as f:return [json.loads(s) for s in f]
data={s:read(s) for s in ['base','head']};info={};rows={};full={};indices={}
for side,records in data.items():
 assert records[0]['status']=='measurement' and records[0]['latent_mode']=='first-error'
 assert records[-1]['status']=='relation_report' and records[-1]['relation_overflow']['per_unit']
 units=[u for r in records if 'file' in r for u in r['units']]
 idx={u['where']:i for i,u in enumerate(units,1)};indices[side]=idx
 attempted=[u for u in units if 'relation_overflow' in u]
 first=next(u for u in units if u.get('relation_overflow'))
 info[side]={'onset_ordinal':idx[first['where']],'onset_attempted_ordinal':next(i for i,u in enumerate(attempted,1) if u.get('relation_overflow')),'onset_unit':short(first['where']),'onset_name':first.get('name'),'catalogued_units':len(units),'probed_units':len(attempted),'files':sum('file' in r for r in records),'diagnostics':records[0]['diagnostics'],'relation_report':records[-1]['relation_overflow']}
 assert records[-1]['relation_overflow']['first_affected_unit']==first['where']
 assert len(attempted)==sum(u['status'] in ['attempted','panic','speculative'] for u in units)
 findings=[f for r in records for f in r.get('findings',[])]
 def key(f):return tuple(short(f.get(k,'')) for k in ['unit','phase','kind','where','reason'])
 full[side]={key(f):f for f in findings}
 rows[side]=sorted(set(' | '.join((*k[:4],k[4][:110])) for k in full[side]))
 (out/('findings-'+side+'.txt')).write_text('\n'.join(rows[side])+'\n')
info['same_header']=data['base'][0]==data['head'][0];assert info['same_header']
assert [(u['where'],'attempted' if u['status'] in ['attempted','panic'] else u['status']) for r in data['base'] if 'file' in r for u in r['units']]==[(u['where'],'attempted' if u['status'] in ['attempted','panic'] else u['status']) for r in data['head'] if 'file' in r for u in r['units']]
(out/'findings.diff').write_text(''.join(difflib.unified_diff([s+'\n' for s in rows['base']],[s+'\n' for s in rows['head']],fromfile='base',tofile='head')))
onset=min(info[s]['onset_ordinal'] for s in ['base','head']);moved=[]
for side,other in [('base','head'),('head','base')]:
 for k in sorted(full[side].keys()-full[other].keys()):
  f=full[side][k];unit=f.get('unit','');ordinal=indices[side].get(unit)
  if unit=='program' and f['phase'].startswith('relation_overflow'):category='relation-limit finding itself'
  else:
   assert ordinal is not None,('unmapped unit',f)
   category='head' if ordinal<onset else 'relation-limit noise'
  moved.append({'side':side,'category':category,'unit_ordinal':ordinal,'row':' | '.join(k),'finding':f})
counts={c:{s:sum(r['side']==s and r['category']==c for r in moved) for s in ['base','head']} for c in ['head','relation-limit noise','relation-limit finding itself']}
info['scan_moved_counts']={c:{s:sum(r['side']==s and r['category']==c and r['finding'].get('phase')=='refusal_scan' for r in moved) for s in ['base','head']} for c in counts};info['moved_counts']=counts;info['finding_list_rows']={s:len(rows[s]) for s in rows};info['exact_finding_rows']={s:len(full[s]) for s in full}
(out/'analysis.json').write_text(json.dumps(info,indent=2)+'\n');(out/'moved-findings.json').write_text(json.dumps(moved,indent=2)+'\n')
(out/'head-rows.txt').write_text('\n'.join(r['side']+' | '+str(r['unit_ordinal'])+' | '+r['row'] for r in moved if r['category']=='head')+'\n')
(out/'relation-limit-noise.txt').write_text('\n'.join(r['side']+' | '+str(r['unit_ordinal'])+' | '+r['row'] for r in moved if r['category']=='relation-limit noise')+'\n')
print(json.dumps(info,indent=2));print((out/'head-rows.txt').read_text())
