#!/usr/bin/env python3
"""Analyze paired canonical streams. No source changes; positions stripped only for pairing keys."""
import sys,json,re,collections
from pathlib import Path
runs=Path(sys.argv[1]);targets=json.loads(Path(sys.argv[2]).read_text());out=Path(sys.argv[3]);out.mkdir(parents=True,exist_ok=True)
GATE='a generic comparison deeper than 4 positions that can reach a Map'
REAL="a Map or Set interface implemented by a value that isn't a Map or Set"
CAP='a fit walk past 4096 distinct pairs'
NAMES={'src/compiler/utilitiesPublic.ts':'utilitiesPublic','src/compiler/utilities.ts':'utilities','src/compiler/transformers/utilities.ts':'transformers-utilities'}
TW={t['where'] for t in targets if t.get('where')}
def stripped(s):return re.sub(r':\d+:\d+','',s)
def key(f):return f['kind'],f['phase'],stripped(f['reason']),stripped(f['text'])
def dedup(fs):return list({(f['kind'],f['phase'],f['where'],f['reason'],f['text']):f for f in fs}.values())
def readrun(label,name):
 p=runs/(label+'-'+name);rs=[];errors=[]
 if (p/'latent.jsonl').exists():
  for i,l in enumerate((p/'latent.jsonl').read_text().splitlines(),1):
   try:rs.append(json.loads(l))
   except Exception as e:errors.append({'line':i,'error':str(e)})
 status=json.loads((p/'status.json').read_text()) if (p/'status.json').exists() else None
 return {'complete':bool(status and status['exit_code']==0 and rs and rs[-1].get('status')=='relation_report'),'status':status,'parse_errors':errors,'header':rs[0] if rs else None,'events':[f for r in rs for f in r.get('findings',[])],'units':{u['where']:u for r in rs for u in r.get('units',[])},'records':rs}
data={(l,n):readrun(l,n) for l in ['base','head'] for n in NAMES.values()}
def pair_changes(bs,hs):
 bg=collections.defaultdict(list);hg=collections.defaultdict(list)
 for b in bs:bg[key(b)].append(b)
 for h in hs:hg[key(h)].append(h)
 unmatched_b=[];unmatched_h=[];matches=0
 for k in sorted(set(bg)|set(hg)):
  b=bg[k][:];h=hg[k][:]
  for specificity in [2,1]:
   rem=[]
   for x in b:
    idx=next((i for i,y in enumerate(h) if y.get('unit')==x.get('unit') and (specificity==1 or y['where']==x['where'])),None)
    if idx is None:rem.append(x)
    else:h.pop(idx);matches+=1
   b=rem
  n=min(len(b),len(h));matches+=n;unmatched_b+=b[n:];unmatched_h+=h[n:]
 changes=[]
 for b in unmatched_b:
  idx=next((i for i,h in enumerate(unmatched_h) if h['where']==b['where'] and h['phase']==b['phase'] and h.get('unit')==b.get('unit')),None)
  if idx is None:changes.append({'grade':'gone','base':b,'head':None})
  else:changes.append({'grade':'changed','base':b,'head':unmatched_h.pop(idx)})
 for h in unmatched_h:
  grade='new';proof=None
  try:
   fn,line,col=h['where'].rsplit(':',2);raw=Path(fn).read_bytes().splitlines(keepends=True);offset=sum(map(len,raw[:int(line)-1]))+len(raw[int(line)-1].decode().rstrip('\r\n')[:int(col)-1].encode())
   for b in bs:
    if b['kind']=='Boundary' and b.get('unit')==h.get('unit') and b['where'].rsplit(':',2)[0]==fn and b.get('start',-1)<=offset<b.get('end',-1):grade='reached';proof={'base_boundary':b['where'],'start':b['start'],'end':b['end']};break
  except (ValueError,IndexError,FileNotFoundError):pass
  changes.append({'grade':grade,'base':None,'head':h,'reach_proof':proof})
 return {'matched_occurrences':matches,'changes':changes}
comparisons={};outcomes=[]
for file,name in NAMES.items():
 b,h=data['base',name],data['head',name]
 info={'base_complete':b['complete'],'head_complete':h['complete']}
 if b['complete'] and h['complete']:
  info['all_events']=pair_changes(b['events'],h['events'])
  info['outside_target_events']=pair_changes([f for f in b['events'] if f['where'] not in TW],[f for f in h['events'] if f['where'] not in TW])
  info['outside_target_distinct_findings']=pair_changes([f for f in dedup(b['events']) if f['where'] not in TW],[f for f in dedup(h['events']) if f['where'] not in TW])
 comparisons[name]=info
for t in targets:
 name=NAMES[t['file']];b,h=data['base',name],data['head',name];w=t.get('where');bf=[f for f in b['events'] if f['where']==w and f['kind']!='Boundary'];hf=[f for f in h['events'] if f['where']==w and f['kind']!='Boundary'];r={'id':t['id'],'where':w,'historical_where':t['historical_where'],'function':t['function'],'file':t['file'],'base_findings':bf,'head_findings':hf,'base_gate_confirmed':any(f['reason']==GATE for f in bf),'classification':'unknown','owning_head_units':[]}
 if not h['complete']:r['classification']='not determined: incomplete head run'
 elif hf:
  reasons={f['reason'] for f in hf}
  r['head_reason_categories']=sorted({('4096-pair cap' if x==CAP else 'real Map check' if x==REAL else 'identity bound' if x==GATE else 'later stop') for x in reasons})
  if CAP in reasons:r['classification']='4096-pair cap'
  elif REAL in reasons:r['classification']='real Map check'
  elif GATE in reasons:r['classification']='identity bound'
  else:r['classification']='later stop'
 else:
  owners={f['unit'] for f in bf};r['owning_head_units']=[h['units'][u] for u in owners if u in h['units']]
  remaining=[f for f in h['events'] if f.get('unit') in owners and f['kind']!='Boundary'];r['remaining_owning_unit_findings']=remaining
  if bf and owners and len(r['owning_head_units'])==len(owners):
   base_owner_keys={(f.get('unit'),f['kind'],f['phase'],f['where'],f['reason'],f['text']) for f in b['events'] if f.get('unit') in owners and f['kind']!='Boundary'}
   newly_reached=[f for f in remaining if (f.get('unit'),f['kind'],f['phase'],f['where'],f['reason'],f['text']) not in base_owner_keys]
   r['new_owning_unit_stops']=newly_reached
   r['classification']='later stop in owning attempt' if newly_reached else 'lowers at mapped site'
   r['owning_units_body_clear']=not any(f['phase']=='lowering' for f in remaining)
  else:r['classification']='not established: no mapped head finding'
 outcomes.append(r)
summary={}
for (label,name),r in data.items():
 fs=r['events'];summary[label+'-'+name]={'complete':r['complete'],'status':r['status'],'parse_errors':r['parse_errors'],'checker_rejected':r['header'].get('checker_rejected') if r['header'] else None,'diagnostics':r['header'].get('diagnostics') if r['header'] else None,'events':len(fs),'distinct_findings':len(dedup(fs)),'distinct_gate_sites':len({f['where'] for f in fs if f['reason']==GATE}),'real_Map_events':sum(f['reason']==REAL for f in fs),'cap_events':sum(f['reason']==CAP for f in fs),'kinds':dict(collections.Counter(f['kind'] for f in fs))}
(out/'run-summary.json').write_text(json.dumps(summary,indent=2)+'\n');(out/'finding-comparisons.json').write_text(json.dumps(comparisons,indent=2)+'\n');(out/'site-outcomes.json').write_text(json.dumps(outcomes,indent=2)+'\n')
print('Outcomes:',dict(collections.Counter(r['classification'] for r in outcomes)))
for n,c in comparisons.items():
 if c['base_complete'] and c['head_complete']:print(n,'outside events',dict(collections.Counter(r['grade'] for r in c['outside_target_events']['changes'])),'outside distinct',dict(collections.Counter(r['grade'] for r in c['outside_target_distinct_findings']['changes'])))
