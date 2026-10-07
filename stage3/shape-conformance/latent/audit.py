"""Independently recount latent rows and check raw diagnostic/body boundaries."""
import collections,copy,json,pathlib,sys
LABEL='measured on a checker-rejected program'
OUTCOMES=('conforms and ready (free)','conforms but not proven ready (readiness checks only)','conforms-if','unknown')
REASONS=('host metadata',"flow the graph can't see",'diagnosed body')

def audit(result,mapped,root,fixtures=False):
 assert result['measurement']==LABEL and result['checker_rejected'] and result['analysis_only']
 assert result['diagnostics']
 lookup={(s['file'],s['start'],s['end']):s for s in mapped}
 assert len(result['sites'])==len(mapped)==len(lookup)
 counts={k:dict.fromkeys(OUTCOMES,0) for k in ('tagged','untagged')}
 reasons={k:collections.Counter() for k in counts}
 seen=set()
 diagnostics=collections.defaultdict(list)
 for d in result['diagnostic_sites']:
  assert d['measurement']==LABEL
  try:file=str(pathlib.Path(d['file']).relative_to(root))
  except ValueError:continue
  diagnostics[file].append(d)
 for row in result['sites']:
  key=(row['file'],row['start'],row['end']);assert key not in seen;seen.add(key)
  origin=lookup[key];assert row['text']==origin['text'] and row['kind']==origin['kind']
  assert row['measurement']==LABEL and row['outcome'] in OUTCOMES
  counts[row['kind']][row['outcome']]+=1
  if row['outcome']=='unknown':assert row['reason'] in REASONS;reasons[row['kind']][row['reason']]+=1
  if row['reason']=='diagnosed body' and 'diagnostic_causes' in row:
   assert row['diagnostic_causes'],'diagnosed site lacks provenance'
   for cause in row['diagnostic_causes']:
    assert cause['scope'] in ('own function','dependency') and cause['diagnostics'],'empty diagnostic cause'
    assert all(d in {raw['text'] for raw in result['diagnostic_sites']} for d in cause['diagnostics']),'invented diagnostic provenance'
  if row.get('host_values'):assert row['outcome']=='unknown','host value certified free'
  if row['outcome']!='unknown':assert row['allocation_sites'] and all(site<0 for site in row['allocation_sites']),'vacuous allocation proof'
  if row['outcome']=='conforms-if':assert row['fields'] and all(f['name'] and f['reason'] for f in row['fields'])
  bodies=(origin.get('adapted') or {}).get('bodies',[])[:1]
  diagnosed=any(d['start']<b['end'] and (d['end']>b['start'] or d['start']>=b['start']) for b in bodies for d in diagnostics[row['file']])
  if diagnosed:assert row['outcome']=='unknown' and row['reason']=='diagnosed body','diagnosed body was analyzed'
 assert counts==result['counts']
 for kind in counts:assert dict(reasons[kind])==result['unknown_reasons'][kind]
 if not fixtures:assert len(seen)==2936 and sum(counts['tagged'].values())==1758 and sum(counts['untagged'].values())==1178
 else:
  expected={'read':OUTCOMES[0],'stageRead':OUTCOMES[1],'wrong':OUTCOMES[2],'hostRead':'unknown','diagnosed':'unknown','incoming':'unknown','callback':'unknown','forged':'unknown','hostField':'unknown','aliased':'unknown','nestedClean':OUTCOMES[0],'nestedCapture':'unknown','propertyRead':'unknown','stringKeyRead':'unknown','arrayRead':'unknown','storeRead':'unknown','opaqueRead':'unknown','diagnosedStoreRead':'unknown','hostThroughField':'unknown','augmentedArray':'unknown','closedCallback':OUTCOMES[0],'returnedCallback':OUTCOMES[0],'mixedCallback':OUTCOMES[2],'hostCallback':'unknown','genericCallback':'unknown','diagnosedCallback':'unknown','genericRead':OUTCOMES[0],'genericWrong':OUTCOMES[2],'genericHostRead':'unknown','genericStagedRead':OUTCOMES[1],'genericDiagnosedRead':'unknown','genericJoinedRead':'unknown','genericRestRead':'unknown','genericSpreadRead':'unknown','genericSpreadJoinedRead':'unknown'}
  by_owner={s['adapted']['owner']:s for s in result['sites']};assert set(by_owner)==set(expected)
  for owner,outcome in expected.items():assert by_owner[owner]['outcome']==outcome,(owner,by_owner[owner])
  for owner,reason in {'hostRead':'host metadata','hostField':'host metadata','diagnosed':'diagnosed body','opaqueRead':REASONS[1],'diagnosedStoreRead':'diagnosed body','hostThroughField':'host metadata','incoming':'diagnosed body','nestedCapture':'diagnosed body','callback':REASONS[1],'forged':REASONS[1],'aliased':REASONS[1],'hostCallback':'host metadata','genericCallback':REASONS[1],'diagnosedCallback':'diagnosed body','genericHostRead':'host metadata','genericDiagnosedRead':'diagnosed body','genericJoinedRead':'host metadata','genericRestRead':REASONS[1],'genericSpreadRead':REASONS[1],'genericSpreadJoinedRead':REASONS[1]}.items():assert by_owner[owner]['reason']==reason,(owner,by_owner[owner])
  assert any(f['name']=='ready' and f['declared']=='number' and f['expected']=='boolean' for f in by_owner['mixedCallback']['fields'])
  assert 'JSON.parse' in by_owner['hostCallback']['host_values']
  assert 'JSON.parse' in by_owner['genericHostRead']['host_values']
  assert any(f['name']=='ready' and 'number' in f['declared'] and f['expected']=='boolean' for f in by_owner['genericWrong']['fields'])
  assert any(f['name']=='ready' and f['reason']=='initialization not proven at the cast' for f in by_owner['genericStagedRead']['fields'])
  assert 'array intrinsic and augmented-field certificates unavailable' in by_owner['augmentedArray']['detail']
  assert 'JSON.parse' in by_owner['hostRead']['host_values'] and 'JSON.parse' in by_owner['hostField']['host_values']
  assert any(f['name']=='ready' and f['declared']=='number' and f['expected']=='boolean' for f in by_owner['wrong']['fields'])
  assert any(f['name']=='ready' and f['reason']=='initialization not proven at the cast' for f in by_owner['stageRead']['fields'])
 return counts

if __name__=='__main__':
 result=json.loads(pathlib.Path(sys.argv[1]).read_text());mapped=json.loads(pathlib.Path(sys.argv[2]).read_text());root=pathlib.Path(sys.argv[3]);fixture='--fixtures' in sys.argv
 print(json.dumps({'measurement':LABEL,'status':'PASS','counts':audit(result,mapped,root,fixture)},indent=2))
