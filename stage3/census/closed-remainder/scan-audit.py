#!/usr/bin/env python3
"""Join separately measured ordinary scans before isolated body/tail first stops."""
import collections,gzip,json,pathlib,sys
root=pathlib.Path(sys.argv[1]);dest=pathlib.Path(sys.argv[2]);dest.mkdir(parents=True,exist_ok=True)
norm=lambda s:'src/'+s.split('/src/',1)[1]
def normalize(f):return {k:norm(v) if isinstance(v,str) and '/src/' in v else v for k,v in f.items()}
def stop(f):return (f['kind'],f.get('where',''),f['reason'],f['phase']) if f else None
scanroot=root/'scans' if (root/'scans').exists() else root/'ordinary-scans'
def read(p):
 text=p.read_text() if p.exists() else gzip.open(str(p)+'.gz','rt').read()
 return [json.loads(s) for s in text.splitlines()]
def scans(side):
 files={};ranges=[]
 for start in ['0','1500','3000']:
  d=scanroot/side/start;s=json.loads((d/'status.json').read_text());assert s['child_exit_code']==0 and not s['deadline_expired'];ranges.append([s['start'],s['end']]);rr=read(d/'census.jsonl');assert not rr[0]['checker_rejected'] and rr[0]['latent_mode']=='first-error'
  assert rr[-1]['status']=='relation_report' and not rr[-1]['relation_overflow']['end'] and not rr[-1]['relation_overflow']['after_load']
  for row in rr[1:-1]:
   assert not row['units'];fs=[normalize(f) for f in row['findings']]
   assert all(f['phase'] in ['refusal_scan','contract_scan','deferred_scan'] for f in fs)
   if not fs:continue
   if row['file'] in files:assert files[row['file']]==fs,('boundary-file scan drift',side,row['file'])
   files[row['file']]=fs
 assert ranges==[[0,1500],[1500,3000],[3000,4600]]
 byunit=collections.defaultdict(list)
 for fs in files.values():
  for f in fs:byunit[f['unit']].append(f)
 return byunit
base=scans('base');head=scans('head');p=root/'analysis/units.json';body=json.loads(p.read_text()) if p.exists() else None
# Published units are an array gzip, not JSONL.
if not p.exists():body=json.loads(gzip.open(root/'units.json.gz','rt').read())
rows=[];changed=[];regressions=[];newpanics=[]
for r in body:
 x={'unit':r['unit'],'name':r['name'],'body_base':r['base'],'body_head':r['head'],'base_scans':base[r['unit']],'head_scans':head[r['unit']]}
 x['base']=base[r['unit']][0] if base[r['unit']] else r['base']['finding'];x['head']=head[r['unit']][0] if head[r['unit']] else r['head']['finding'];rows.append(x)
 if stop(x['base'])!=stop(x['head']):changed.append(x)
 if not x['base'] and x['head']:regressions.append(x)
 if x['head'] and x['head']['kind']=='panic' and (not x['base'] or x['base']['kind']!='panic'):newpanics.append(x)
kind="a read of field kind through type Node, which may hold a checked view's value"
summary={'unit_count':len(rows),'changed_first_stops':len(changed),'destinations':dict(collections.Counter('lowers' if not r['head'] else 'kind through Node' if r['head']['reason']==kind else 'elsewhere' for r in changed)),'regressions':len(regressions),'new_panics':len(newpanics),'identical_unit_scans':dict(base)==dict(head),'scan_only_owner_locations_not_in_body':sorted((set(base)|set(head))-{r['unit'] for r in body})}
for name,data in [('ordered-units',rows),('ordered-changed-first-stops',changed),('ordered-regressions',regressions),('ordered-new-panics',newpanics),('ordered-summary',summary)]:
 text=json.dumps(data,indent=2)+'\n'
 if name=='ordered-units':(dest/(name+'.json.gz')).write_bytes(gzip.compress(text.encode(),mtime=0))
 else:(dest/(name+'.json')).write_text(text)
print(json.dumps(summary,indent=2))
