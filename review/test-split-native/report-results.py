import json,re
from pathlib import Path
root=Path(__file__).resolve().parent
rows=[json.loads(s) for s in (root/'cold/results.jsonl').read_text().splitlines()]
latest={r['test']:r for r in rows}
latest.pop('TestRecordReadMutants',None)
expected=['TestArtifactInputsAndReuse','TestShardSelection','TestRetainedSplitCoverage','TestRetainedTopLevelCoverage']
for prefix,n in [('TestRegExpBytecodeRandomNodeUnit',40),('TestRecordMutantsUnit',6),('TestSplitTSGoAgreesUnit',2),('TestWASIUnit',36)]:expected += [prefix+f'{i:02}' for i in range(n)]
expected += ['TestRecordReadMutants/'+n for n in ['prototype-membership-restored','missing-read-silent','own-read-checked-as-missing']]
assert set(expected) <= set(latest),set(expected)-set(latest)
for n in expected:
 r=latest[n];assert r['exit']==0 and r['result']['Action']=='pass' and r['wall']<60,(n,r)
leaf_rows=[]
for n in expected:
 r=latest[n];file=root/'cold'/(n.replace('/','__')+'.jsonl');events=[]
 for line in file.read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError:continue
 terminal={e['Test']:e for e in events if e.get('Test') and e['Action'] in ['pass','fail','skip']}
 leaves={name:e for name,e in terminal.items() if not any(o.startswith(name+'/') for o in terminal)}
 for name,e in leaves.items():
  assert e['Action']=='pass' and e.get('Elapsed',0)<60,(name,e)
  leaf_rows.append(dict(test=name,elapsed=e.get('Elapsed',0),root_elapsed=r['result']['Elapsed'],wall=r['wall'],selector=r['command']))
(root/'results.json').write_text(json.dumps(dict(roots=len(expected),leaves=len(leaf_rows),max_wall=max(r['wall'] for r in leaf_rows),max_root=max(r['root_elapsed'] for r in leaf_rows),measurements=leaf_rows),indent=2)+'\n')
text='| Test leaf | Leaf seconds | Root with setup seconds | Invocation seconds |\n| --- | ---: | ---: | ---: |\n'
for r in leaf_rows:text+=f"| {r['test']} | {r['elapsed']:.2f} | {r['root_elapsed']:.2f} | {r['wall']:.2f} |\n"
(root/'TIMINGS.md').write_text(text)
print('roots',len(expected),'leaves',len(leaf_rows),'max invocation',max(r['wall'] for r in leaf_rows))
