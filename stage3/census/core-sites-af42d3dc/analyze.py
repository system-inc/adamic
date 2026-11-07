import json,pathlib,re,collections,subprocess
scratch=pathlib.Path('/workspace/scratch/after-wave');repo=pathlib.Path('/workspace/adamic');dest=repo/'stage3/census/core-sites-af42d3dc'
raw=[json.loads(l) for l in (scratch/'raw.jsonl').read_text().splitlines()]
record=next(r for r in raw if r.get('file','').endswith('/core.ts'))
ast=json.loads((scratch/'ast-map.json').read_text())['adapted']
def lc(where):
 m=re.search(r':(\d+):(\d+)$',where);return tuple(map(int,m.groups())) if m else (0,0)
def at(where):
 if '/core.ts:' not in where:return None
 line,col=lc(where)
 candidates=[n for n in ast if (n['line'],n['column'])==(line,col)]
 return candidates[0] if candidates else None
source=pathlib.Path('/workspace/after-wave-adapted/src/compiler/core.ts').read_text();starts=[0]+[m.end() for m in re.finditer('\n',source)]
def offset(where):
 l,c=lc(where);return starts[l-1]+c-1 if l else -1
units=[]
for i,u in enumerate(record['units']):
 node=at(u['where'])
 if not node:raise RuntimeError(('unmapped unit',u))
 units.append(dict(u,id='U'+str(i+1),adapted_line=node['line'],adapted_column=node['column'],upstream_line=node.get('upstream_line'),display_name=u.get('name') or node['name'] or node['kind'],source_start=node['start'],source_end=node['end'],end_line=node['end_line'],end_column=node['end_column'],source_key=node['key'],type_parameter_count=node['type_parameter_count'],mapping_method=node['mapping_method']))
def site_unit(where):
 if '/core.ts:' not in where:return None
 pos=lc(where)
 candidates=[u for u in units if (u['adapted_line'],u['adapted_column'])<=pos<(u['end_line'],u['end_column'])]
 return min(candidates,key=lambda u:u['source_end']-u['source_start']) if candidates else None
bywhere={u['where']:u for u in units}
findings=[]
for f in record['findings']:
 f=dict(f); owner=site_unit(f['where']);f['site_unit_id']=owner['id'] if owner else None
 f['site_unit_name']=owner['display_name'] if owner else None
 f['recording_unit_id']=bywhere.get(f['unit'],{}).get('id')
 f['in_range']=bool(owner)
 f['instantiation_independent']=f['phase'] in ['refusal_scan','contract_scan'] or bool(owner and (owner['status']!='deferred_generic_instantiation' or not owner.get('body_start')))
 findings.append(f)
for u in units:
 owned=[f for f in findings if f['recording_unit_id']==u['id'] and f['kind'] not in ['Boundary']]
 u['reported_status']=u['status'] if u['status']!='attempted' else ('finding' if owned or any(f['site_unit_id']==u['id'] and f['kind']!='Boundary' for f in findings) else 'lowered')
 u['site_findings']=[i for i,f in enumerate(findings) if f['site_unit_id']==u['id']]
 u['recorded_findings']=[i for i,f in enumerate(findings) if f['recording_unit_id']==u['id']]
start=1
selected=[u for u in units if u['adapted_line']>=start]
# Retain every recording event in rows.json, and group the report's repeated observations by site, kind, exact reason and phase.
groups={}
for i,f in enumerate(findings):
 if not f['in_range'] or f['kind']=='Boundary':continue
 key=tuple(f.get(k) for k in ['where','kind','reason','phase'])
 if key not in groups:groups[key]=dict(f,raw_finding_indexes=[],recording_units=[])
 g=groups[key];g['raw_finding_indexes'].append(i)
 if f['unit'] not in g['recording_units']:g['recording_units'].append(f['unit'])
sites=sorted(groups.values(),key=lambda f:lc(f['where']))
for i,f in enumerate(sites): f['id']='F'+str(i+1)
requested='arrayOf arrayFrom arrayToMap arrayToMultiMap group groupBy extend copyProperties assign maybeBind createMultiMap multiMapAdd multiMapRemove createQueue createSet toArray cast memoizeOne levenshteinWithMax removeMinAndVersionNumbers createGetCanonicalFileName orderedRemoveItem or not takeWhile enumerateInsertsAndDeletes cartesianProductWorker isNodeLikeSystem'.split()
result={'measurement':raw[0],'units':selected,'findings':findings,'sites':sites,'first_walls':[f['id'] for f in sites if f['instantiation_independent']],'deferred_false':[u['id'] for u in selected if u['status']=='deferred_generic_instantiation' and not u['instantiated']],'requested_functions':{name:[u for u in units if u['display_name']==name] for name in requested}}
(dest/'rows.json').write_text(json.dumps(result,indent=2)+'\n')
print('range start',start,'units',len(selected),'statuses',dict(collections.Counter(u['reported_status'] for u in selected)),'sites',len(sites),'kinds',dict(collections.Counter(f['kind'] for f in sites)),'deferred false',len(result['deferred_false']))
for f in sites:print(f['id'],f['where'].split('/core.ts:')[-1],f['site_unit_name'],f['kind'],f['phase'],f['instantiation_independent'],f['reason'])
print('missing requested',[n for n,v in result['requested_functions'].items() if not v])
