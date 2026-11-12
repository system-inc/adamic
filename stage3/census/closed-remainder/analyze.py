#!/usr/bin/env python3
"""Audit paired scoped-tail streams and emit a normalized, lossless unit ledger."""
import collections,gzip,json,pathlib,sys
root=pathlib.Path(sys.argv[1]);dest=pathlib.Path(sys.argv[2]);dest.mkdir(parents=True,exist_ok=True)
def norm(s):
 if isinstance(s,str) and '/src/' in s: return 'src/'+s.split('/src/',1)[1]
 return s
def rows(p):
 text=p.read_text() if p.exists() else gzip.open(str(p)+'.gz','rt').read()
 return [json.loads(s) for s in text.splitlines()]
passes=['checkOverloadResults','forwarderValues','finishClassCalls','finishAccessors','finishToPrimitive','checkIterableViews','exceptions','checkLibrarySpreads','checkAccessorSpreads','checkRecordGuardedReads','checkDynamicKeyedGetters','checkGeneratorRoads','NormalizeAsync','asyncEntrySource','checkBoundArguments','checkViewOptionalWrites','checkViewCopies','checkViewWidenedReads','checkHostMemberWrites','asyncPayloadFailures','checkPromiseConsumption','checkViewLogicalWrites','readiness','borrow','counters','voidDiscards']
runs=root/'runs' if (root/'runs').exists() else root/'census'
results={};statuses=[]
for side in ['base','head']:
 units={};findings=collections.defaultdict(list);tails=collections.defaultdict(list);headers=[]
 for d in sorted((runs/side).iterdir(),key=lambda p:int(p.name)):
  status=json.loads((d/'status.json').read_text());assert status['child_exit_code']==0 and not status['deadline_expired'];statuses.append(dict(side=side,**status))
  rr=rows(d/'census.jsonl');headers.append(rr[0]);assert not rr[0]['checker_rejected'] and rr[0]['latent_mode']=='first-error'
  trailer=rr[-1];assert trailer['status']=='relation_report';assert not trailer['relation_overflow']['after_load'] and not trailer['relation_overflow']['end']
  for r in rr[1:-1]:
   for u in r['units']:
    key=norm(u['where']);assert key not in units;units[key]=u
   for f in r['findings']:
    assert f['phase'] not in ['refusal_scan','contract_scan','deferred_scan'];findings[norm(f['unit'])].append(f)
  for p in rows(d/'passes.jsonl'):tails[norm(p['unit'])].append(p)
 assert all(h['roots']==headers[0]['roots'] and h['sources']==headers[0]['sources'] for h in headers)
 for key in findings: assert key in units,(side,key)
 for key,u in units.items():
  fs=findings[key];assert len(fs)<=1,(side,key,fs)
  ps=tails[key];assert [p['pass'] for p in ps]==passes[:len(ps)],(side,key)
  assert not any('error' in p or 'panic' in p for p in ps[:-1])
  if u['status']=='attempted' and not fs:assert len(ps)==len(passes),(side,key,'tail missing')
  if ps and ('error' in ps[-1] or 'panic' in ps[-1]):assert fs,(side,key)
 results[side]={'units':units,'findings':findings,'passes':tails}
assert set(results['base']['units'])==set(results['head']['units'])
ledger=[];changed=[];regressions=[];panics=[]
for key in results['base']['units']:
 row={'unit':key,'name':results['base']['units'][key].get('name','')}
 for side in ['base','head']:
  fs=results[side]['findings'][key];ps=results[side]['passes'][key]
  row[side]={'status':results[side]['units'][key]['status'],'finding':{k:norm(v) for k,v in fs[0].items()} if fs else None,'passes':[p['pass'] for p in ps]}
 def stop(side):
  f=row[side]['finding'];return (f['kind'],f['where'],f['reason']) if f else None
 if stop('base')!=stop('head'):changed.append(row)
 if stop('base') is None and stop('head') is not None:regressions.append(row)
 if row['head']['finding'] and row['head']['finding']['kind']=='panic' and (not row['base']['finding'] or row['base']['finding']['kind']!='panic'):panics.append(row)
 ledger.append(row)
kind_reason="a read of field kind through type Node, which may hold a checked view's value"
counts=collections.Counter('lowers' if not r['head']['finding'] else 'kind-through-Node' if r['head']['finding']['reason']==kind_reason else 'elsewhere' for r in changed)
summary={'units':len(ledger),'changed_first_stop_units':len(changed),'changed_first_stop_counts':dict(counts),'regressions':len(regressions),'new_panics':len(panics),'status':statuses}
for name,data in [('units',ledger),('changed-first-stops',changed),('regressions',regressions),('new-panics',panics),('summary',summary)]: (dest/(name+'.json')).write_text(json.dumps(data,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='status'},indent=2))
