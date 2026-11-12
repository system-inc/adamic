#!/usr/bin/env python3
"""Validate raw coverage, first-error prefixes, child exits and the negative pin independently."""
import collections,gzip,hashlib,json,pathlib,sys
ROOT=pathlib.Path(__file__).resolve().parent
E=ROOT/'evidence'
PASSES=['checkOverloadResults','forwarderValues','finishClassCalls','finishAccessors','finishToPrimitive','checkIterableViews','exceptions','checkLibrarySpreads','checkAccessorSpreads','checkRecordGuardedReads','checkDynamicKeyedGetters','checkGeneratorRoads','NormalizeAsync','asyncEntrySource','checkBoundArguments','checkViewOptionalWrites','checkViewCopies','checkViewWidenedReads','checkHostMemberWrites','asyncPayloadFailures','checkPromiseConsumption','checkViewLogicalWrites','readiness','borrow','counters','voidDiscards']
def rows(path):
 if not path.exists():return []
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as f:return [json.loads(x) for x in f if x.strip()]
def stream(d):
 status=json.loads((d/'status.json').read_text())
 assert status['child_exit_code']==0 and not status['deadline_expired'],(d,status)
 assert status['wall_seconds']<1800,(d,status)
 r=rows(d/'census.jsonl.gz')
 assert r[0]['status']=='measurement' and r[0]['latent_mode']=='first-error' and not r[0]['checker_rejected'],d
 assert r[-1]['status']=='relation_report',d
 rel=r[-1]['relation_overflow']
 assert not rel['after_load'] and not rel['end'],(d,rel)
 assert rel['cache_sizes']['flaggedPairs']==rel['cache_sizes']['ts2859']==0,(d,rel)
 files=[x for x in r if 'file'in x]
 assert len(files)==len(r[0]['sources'])==len({x['file'] for x in files})==81,d
 assert {x['file'] for x in files}==set(r[0]['sources']),d
 return status,files

def main(mode):
 units={};findings=[];allcalls=[];timed=[];next_start=0;statuses=[]
 for d in sorted((E/'main'/mode).iterdir()):
  if not d.is_dir():continue
  status,files=stream(d);statuses.append(status)
  assert status['start']==next_start,(d,next_start)
  next_start=status['end']
  chunk=[u for f in files for u in f['units']]
  assert len(chunk)==status['end']-status['start'],(d,len(chunk))
  for u in chunk:
   assert u['where'] not in units,u['where']
   units[u['where']]=u
  findings.extend(f for r in files for f in r['findings'])
  calls=rows(d/'passes.jsonl.gz');allcalls.extend(calls)
  t=rows(d/'units.jsonl.gz');timed.extend(t)
  assert len(t)==sum(u['status'] in ('attempted','panic') for u in chunk),(d,len(t))
  byunit=collections.defaultdict(list)
  for p in calls:byunit[p['unit']].append(p)
  for unit,ps in byunit.items():
   assert unit in units and units[unit]['status'] in ('attempted','panic'),unit
   assert [p['pass'] for p in ps]==PASSES[:len(ps)],unit
   failures=[i for i,p in enumerate(ps) if 'error'in p or 'panic'in p]
   assert not failures or failures==[len(ps)-1],unit
   if not failures:assert len(ps)==len(PASSES),unit
   assert all(p['wall_seconds']>=0 and p['hwm_growth_kib']>=0 for p in ps),unit
  if mode=='stock':assert not calls,d
 manifest=rows(E/'tip-unit-ranges.jsonl')
 assert next_start==manifest[-1]['end']==len(units)==4566,(next_start,len(units))
 return units,findings,allcalls,timed,statuses

stock,sf,_,st,ss=main('stock')
proto,pf,calls,pt,ps=main('prototype')
assert set(stock)==set(proto)
assert all(stock[u]['kind']==proto[u]['kind'] for u in stock)
base_stopped={f['unit'] for f in sf if f['phase'] in ('lowering','deferred')}
tails={p['unit'] for p in calls}
assert not tails&base_stopped,sorted(tails&base_stopped)
errors=[p for p in calls if 'error'in p]
panics=[p for p in calls if 'panic'in p]
assert len(errors)+len(panics)==len({p['unit'] for p in errors+panics})
for p in errors+panics:
 assert any(f['unit']==p['unit'] and f['phase']=='whole_pass:'+p['pass'] for f in pf),p
hashes=[];pin_owners=[]
for rev in ['parent','87028b08']:
 for mode in ['stock','prototype']:
  d=E/'pin'/rev/mode
  _,files=stream(d)
  units=[u for f in files for u in f['units']]
  assert len(units)==51
  owner=next(u for u in units if u.get('name')=='createTypeChecker')
  failed=[f for r in files for f in r['findings'] if f['unit']==owner['where']]
  assert len(failed)==1 and failed[0]['phase']=='lowering' and failed[0]['reason']=='a value of type ParameterPropertyDeclaration',failed
  assert not any(p['unit']==owner['where'] for p in rows(d/'passes.jsonl.gz'))
  with gzip.open(d/'census.jsonl.gz','rb') as f:hashes.append(hashlib.sha256(f.read()).hexdigest())
  pin_owners.append(owner['where'])
assert len(set(hashes))==1 and hashes[0]=='c6f1d2498c6fad3f83c89dcb36c0d9f8b45e2828db8461963f32c86937cffa0d'
sites=json.loads((E/'parent-generator-sites.json').read_text())
assert len(sites)==5 and all(x['top']=='createTypeChecker' and x['top_where']=='1489:1' for x in sites)
trees=json.loads((E/'pin-trees.json').read_text());assert trees['87028b08^']==trees['87028b08']
result={'entry_sources':81,'units':len(proto),'tail_units':len(tails),'tail_unit_kinds':dict(collections.Counter(proto[u]['kind'] for u in tails)),
        'new_stopped_units':len(errors),'new_stops_by_pass':dict(collections.Counter(p['pass'] for p in errors)),
        'panics':len(panics),'pass_calls':len(calls),'pass_wall_seconds':sum(p['wall_seconds'] for p in calls),
        'stock_unit_wall_seconds':sum(x['wall_seconds'] for x in st),'prototype_unit_wall_seconds':sum(x['wall_seconds'] for x in pt),
        'stock_peak_rss_kib':max(s['peak_rss_kib'] for s in ss),'prototype_peak_rss_kib':max(s['peak_rss_kib'] for s in ps),
        'requested_positive_pin':'not reproduced; owner stops before tail','pin_stream_sha256':hashes[0],
        'scope':'unit-only first-error; independent whole-file refusal scan omitted on both sides'}
print(json.dumps(result,indent=2,sort_keys=True))
