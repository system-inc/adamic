#!/usr/bin/env python3
"""Pair current guard stream using the archived pair tool and immutable parent evidence."""
import collections,gzip,hashlib,json,runpy
from pathlib import Path
D=Path(__file__).resolve().parent;P=D.parent
H=runpy.run_path(str(P/'pair-guards.py'));load=H['load'];pair=H['pair'];tree=Path('/workspace/roads-adapted');ts=P.parents[2]/'stage3/api/node_modules/typescript'

def read_json(p):
 if not p.exists():p=Path(str(p)+'.gz')
 return json.load(gzip.open(p,'rt') if str(p).endswith('.gz') else open(p))
def save(name,v): (D/name).write_text(json.dumps(v,indent=2)+'\n')
def key(r):
 f=r['finding'];return (tuple(r['identity']),f['kind'],f['reason'],f['text'])
def viewkey(r):return (r['where'],r['source'],r['target'])

obs=load(D/'raw/first-records.jsonl');normal=load(D/'raw/first-normal.jsonl');baseline=load(P/'raw/baseline-parent.jsonl');parentobs=load(P/'raw/guard-parent-records.jsonl')
errors=[r for r in obs if r['kind']=='GuardError'];assert not errors,errors
refs=[r for r in obs if r['kind']=='GuardRefusal'];units=[r for r in obs if r['kind']=='GuardUnit'];assert len(refs)==sum(x['refusals'] for x in units)
assert all('error' not in r['diagnostics'] for r in refs)
assert {r['unit']:r['views'] for r in units}=={r['unit']:r['views'] for r in parentobs if r['kind']=='GuardUnit'}
a=P/'raw/baseline-parent.jsonl';b=D/'raw/first-normal.jsonl'
assert H['raw_findings'](a)==H['raw_findings'](b),'ordinary finding bytes changed'
assert H['non_telemetry_bytes'](a)==H['non_telemetry_bytes'](b),'ordinary census bytes changed'
assert H['only_cache_telemetry_changed'](baseline,normal),'non-cache telemetry changed'
verification={'all_header_file_unit_finding_bytes_identical':True,'only_cache_telemetry_may_differ':True,'ordinary_findings':len(H['normal_findings'](normal)),'ordinary_findings_sha256':hashlib.sha256(H['raw_findings'](b)).hexdigest(),'non_telemetry_bytes_sha256':hashlib.sha256(H['non_telemetry_bytes'](b)).hexdigest(),'entire_stream_bytes_identical':H['raw_stream'](a)==H['raw_stream'](b),'per_unit_view_counts_identical':True}
save('normal-verification.json',verification)
grouped=collections.defaultdict(list)
for r in refs:grouped[r['file']].append({'kind':'NotYet','measurement':normal[0]['measurement'],'phase':'lowering','unit':r['unit'],'where':r['where'],'reason':r['reason'],'text':r['text']})
converted=[normal[0]]+[dict(r,findings=grouped[r['file']]) for r in normal[1:] if 'units' in r]
(D/'head-for-pair.jsonl').write_text(''.join(json.dumps(r,separators=(',',':'))+'\n' for r in converted))
parent=load(P/'guard-parent-for-pair.jsonl');delta=pair.compare(parent,converted,tree,ts)
save('delta.json',delta);(D/'delta.md').write_text(pair.report(delta))
old=read_json(P/'guard-delta.json');oldrefs=read_json(P/'guard-new-records.json');oldunits=read_json(P/'guard-new-unit-stops.json')
oldkeys={key(r) for r in old['raw_findings']['new_findings']};newkeys={key(r) for r in delta['raw_findings']['new_findings']}
oldviews={viewkey(r) for r in oldrefs};newviews={viewkey(r) for r in refs}
old114={r['unit']['where'] for r in oldunits};stops={r['unit'] for r in refs};ordinary={f['unit'] for f in H['normal_findings'](normal) if f.get('phase') in ('lowering','deferred')}
roads={}
for category in ['raw_findings','distinct_reads']:
 roadmap={(r['where'],r['reason'],r['text']):r['road'] for r in refs}
 roads[category]=dict(collections.Counter(roadmap[(x['finding']['where'],x['finding']['reason'],x['finding']['text'])] for x in delta[category]['new_findings']))
summary={'target':'ed3c0354','parent':'16429a30','observed_units':len(units),'views':sum(r['views'] for r in units),'refusal_observations':len(refs),'roads_observations':dict(collections.Counter(r['road'] for r in refs)),'roads_findings':roads,'guard_stopped_units':len(stops),'ordinary_finding_free_guard_stopped_units':len(stops-ordinary),'prior_114_still_guard_stopped':len(old114&stops),'prior_114_cleared':len(old114-stops),'prior_289_exact_findings_remain':len(oldkeys&newkeys),'prior_289_exact_findings_gone':len(oldkeys-newkeys),'new_exact_findings_not_in_289':len(newkeys-oldkeys),'prior_view_sites_still_refuse':len(oldviews&newviews),'new_view_sites_not_in_prior':len(newviews-oldviews),'checker_identical_distinct_object_refusal_observations':sum(bool(r['diagnostics']['checker_identical_distinct_object_pairs']) for r in refs),'split_crossing_refusal_observations':sum(bool(r['diagnostics']['split_crossings_without_literal_separation']) for r in refs),'probe_errors':len(errors),'diagnostic_errors':0,'parent_refusals':sum(r['kind']=='GuardRefusal' for r in parentobs),'pair_counts':{k:{x:len(delta[k][x]) for x in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']} for k in ['raw_findings','distinct_reads']}}
save('summary.json',summary);save('old-114-still-stopped.json',sorted(old114&stops));save('old-114-cleared.json',sorted(old114-stops))
cases=[]
for i,x in enumerate(delta['raw_findings']['new_findings'],1):
 f=x['finding'];matches=[r for r in refs if (r['where'],r['reason'],r['text'])==(f['where'],f['reason'],f['text'])];assert matches
 r=matches[0];cases.append({'id':i,'identity':x['identity'],'exact_prior_finding':key(x) in oldkeys,'view_in_prior_findings':viewkey(r) in oldviews,'observations':len(matches),'owner_units':sorted({r['unit'] for r in matches}),'record':r})
save('cases.json',cases)
save('checker-identical-distinct-object-cases.json',[c for c in cases if c['record']['diagnostics']['checker_identical_distinct_object_pairs']])
save('split-crossing-cases.json',[c for c in cases if c['record']['diagnostics']['split_crossings_without_literal_separation']])
print(json.dumps(summary,indent=2));print(json.dumps(verification,indent=2))
