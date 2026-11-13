"""Audit pairing against exact unchanged source-line mappings and complete runs."""
import collections,difflib,json,pathlib
D=pathlib.Path(__file__).resolve().parent
roots={'before':pathlib.Path('/workspace/scratch/timer-main-tip-final'),'after':pathlib.Path('/workspace/scratch/timer-after-tip')}
x=json.loads((D/'comparison.json').read_text());cache={};checked=0
for file,result in x['files'].items():
 for label in roots:
  status=json.loads((D/(label+'-'+file[:-3]+'.status.json')).read_text());assert status['child_exit']==0 and not status['timed_out']
  raw=[json.loads(l) for l in (D/(label+'-'+file[:-3]+'.jsonl')).read_text().splitlines()]
  assert raw[0]['diagnostics']==[] and raw[0]['latent_mode']=='full'
  trailer=next(r for r in raw if r.get('status')=='relation_report')
  assert trailer['relation_overflow']['after_load']=='' and trailer['relation_overflow']['end']==''
 for pair in result['retainedPairs']:
  for field in ['where','unit']:
   parts={k:pair[k][field].rsplit(':',2) for k in roots};name=parts['before'][0];assert name==parts['after'][0]
   if name not in cache:
    texts={k:(root/name).read_text().splitlines() for k,root in roots.items()};mapping={}
    for a,b,count in difflib.SequenceMatcher(None,texts['before'],texts['after'],autojunk=False).get_matching_blocks():
     for offset in range(count):mapping[b+offset+1]=a+offset+1
    cache[name]=mapping
   assert cache[name][int(parts['after'][1])]==int(parts['before'][1]),pair
   assert parts['before'][2]==parts['after'][2],pair
  assert pair['before']['sourceLine']==pair['after']['sourceLine'],pair
  checked+=1
 assert result['before']['total']==len(result['gone'])+result['retained']
 assert result['after']['total']==len(result['new'])+result['retained']
record={'retainedPairsAudited':checked,'unitAndFindingPositionsMapToSameUnchangedSource':True,'checkerDiagnostics':0,'relationOverflow':False,'allFourChildExits':0,'newFindingClassificationsRequired':sum(len(r['new']) for r in x['files'].values())}
(D/'audit.json').write_text(json.dumps(record,indent=2)+'\n');print(json.dumps(record,indent=2))
