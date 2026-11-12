import json,pathlib
S=pathlib.Path('/workspace/ldpi')
a=[json.loads(l) for l in (S/'combined-shallow/census.jsonl').open()];b=[json.loads(l) for l in (S/'combined-fixed/census.jsonl').open()]
result={'unit_metadata_changes':[],'relation_cache_changes':[],'finding_occurrences_equal':True,'all_unit_status_and_instantiated_flags_equal':True}
for x,y in zip(a[1:-1],b[1:-1]):
 assert x['file']==y['file']
 result['finding_occurrences_equal'] &= x['findings']==y['findings']
 assert x['declarations']==y['declarations']
 for u,v in zip(x['units'],y['units']):
  assert u['where']==v['where']
  result['all_unit_status_and_instantiated_flags_equal'] &= u['status']==v['status'] and u['instantiated']==v['instantiated']
  if u!=v:
   result['unit_metadata_changes'].append({'where':u['where'],'name':u.get('name'),'changed_fields':[k for k in set(u)|set(v) if u.get(k)!=v.get(k)],'base':u,'head':v})
for x,y in zip(a[-1]['relation_overflow']['segmented_reports'],b[-1]['relation_overflow']['segmented_reports']):
 assert x['group']==y['group']
 p=x['report']['cache_sizes'];q=y['report']['cache_sizes']
 for name in p:
  if p[name]!=q[name]:result['relation_cache_changes'].append({'group':x['group'],'cache':name,'before':p[name],'after':q[name],'delta':q[name]-p[name]})
assert result['finding_occurrences_equal'] and result['all_unit_status_and_instantiated_flags_equal']
assert len(result['unit_metadata_changes'])==1 and result['unit_metadata_changes'][0]['changed_fields']==['instantiated_by']
(S/'metadata-delta.json').write_text(json.dumps(result,indent=2,sort_keys=True)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='unit_metadata_changes'},indent=2))
