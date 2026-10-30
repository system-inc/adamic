#!/usr/bin/env python3
"""Follow saved sites across measured ledgers; context matching is not causal proof."""
import gzip,json,collections,sys
from pathlib import Path
old=Path(sys.argv[1]);before=Path(sys.argv[2]);after=Path(sys.argv[3]);output=Path(sys.argv[4])
def read(path):
 opener=gzip.open if path.suffix=='.gz' else open
 with opener(path,'rt') as stream:records=[json.loads(line) for line in stream]
 roots={r['file'].split('/src/compiler/')[0] for r in records[1:]};assert len(roots)==1
 root=next(iter(roots))+'/'
 return [{k:v.replace(root,'') if isinstance(v,str) else v for k,v in f.items()} for r in records[1:] for f in r.get('findings') or [] if f['kind'] in ['NotYet','Refused']]
key=lambda f:tuple(f[k] for k in ['kind','where','reason','text'])
a,b=read(old/'before-latent.jsonl.gz'),read(old/'after-latent.jsonl.gz');aa={key(f) for f in a};added={key(f):f for f in b if key(f) not in aa};assert len(added)==164
known={k for k,f in added.items() if f['reason']=='a value of type Path' or f['reason'].startswith('a primitive brand member ')};assert len(known)==31
previous={k:f for k,f in added.items() if k not in known};assert len(previous)==133
control,treatment=read(before),read(after);current_before={key(f) for f in control};current_after={key(f) for f in treatment};by_where=collections.defaultdict(list);by_unit=collections.defaultdict(list);treatment_contexts=collections.defaultdict(list)
for f in treatment:treatment_contexts[key(f)].append((f['unit'],f['phase']))
for f in control:by_where[f['where']].append(f);by_unit[(f['unit'],f['phase'])].append(f)
rows=[]
for k,f in sorted(previous.items()):
 if k not in current_after:status='absent from current treatment ledger'
 elif k in current_before:status='present in current control and treatment'
 else:status='present only in current treatment'
 at_where=by_where[f['where']];at_unit=[x for context in set(treatment_contexts[k]) for x in by_unit[context]]
 void=lambda x:'void container' in x['reason'] or 'instantiated at void' in x['reason']
 rows.append({**dict(zip(['kind','where','reason','text'],k)),'status':status,'same_location_control_void':[x for x in at_where if void(x)],'same_attempt_control_void':[x for x in at_unit if void(x)],'control_other_at_location':[x for x in at_where if not void(x)]})
for row in rows:
 for field in ['same_location_control_void','same_attempt_control_void','control_other_at_location']:
  unique={key(f):f for f in row[field]}
  row[field+'_count']=len(unique)
  row[field]=[unique[k] for k in sorted(unique)[:3]]
counts=dict(collections.Counter(r['status'] for r in rows));present_only=[r for r in rows if r['status']=='present only in current treatment']
summary={'previous_unclassified_sites':133,'context_examples_capped_at':3,'site_status':counts,'treatment_only_with_control_void_at_same_location':sum(bool(r['same_location_control_void']) for r in present_only),'treatment_only_with_control_void_in_same_attempt':sum(bool(r['same_attempt_control_void']) for r in present_only),'known31_remaining_in_treatment':sum(k in current_after for k in known),'limits':'An absent observation is not proof the construct now lowers; a same-attempt or same-location prior void finding does not prove the later refusal existed with old types. First-error ledgers do not provide that counterfactual. Main also changed between the historical and current runs.'}
output.mkdir(parents=True,exist_ok=True);(output/'previous-133-split.json').write_text(json.dumps({'summary':summary,'sites':rows},indent=2)+'\n');lines=['| Previous site | Reason | Current observation |','|---|---|---|']
for r in rows:lines.append('| '+r['where']+' | '+r['reason'].replace('|','&#124;')+' | '+r['status']+' |')
(output/'previous-133-split.md').write_text('\n'.join(lines)+'\n');print(json.dumps(summary))
