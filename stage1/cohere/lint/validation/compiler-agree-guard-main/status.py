import pathlib,json,sys,collections
root=pathlib.Path(__file__).parent
mode=sys.argv[1]; counts=collections.Counter(); active=set(); logs=[]
for line in (root/f'{mode}.jsonl').read_text().splitlines():
 try:e=json.loads(line)
 except ValueError:continue
 test=e.get('Test'); action=e.get('Action')
 if test:
  if action in ('run','cont'):active.add(test)
  if action in ('pause','pass','fail','skip'):active.discard(test)
  if action in ('pass','fail','skip'):counts[action]+=1
 if 'Output' in e and (test=='TestCompilerAndStage1Agree' or test and test.startswith('TestChild')):logs.append(e['Output'].strip())
print(json.dumps(dict(mode=mode,counts=dict(counts),active=sorted(active) if len(sys.argv)==2 else sorted(active)[:3],active_count=len(active),latest_relevant=logs[-8:] if len(sys.argv)==2 else []),indent=2))
