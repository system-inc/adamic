import pathlib,json,re,collections
root=pathlib.Path.cwd();events=[]
for l in pathlib.Path('/tmp/wave2-07-gate/stage1-cohere-lint.jsonl').open():
 try:events.append(json.loads(l))
 except:continue
rows=[]
for p in sorted((root/'stage1/cohere/lint/rules').glob('*/rule.json')):
 d=json.loads(p.read_text());m={k.lower():v for k,v in json.loads((p.parent/'mutant.json').read_text()).items()};test='TestMutants/'+re.sub(r'\s','_',m['name']);logs=[e.get('Output','') for e in events if e.get('Test')==test];s=''.join(logs);backends=[]
 for b in ['Node','emitted JavaScript','native']:
  if ' caught on '+b+':' in s:backends.append(b)
 rows.append({'rule':d['name'],'slug':p.parent.name,'mutant':m['name'],'test':test,'caught':backends,'native_canary':'sanitized native canary ' in s,'run_event':any(e['Action']=='cont' and e.get('Test')==test for e in events),'terminal':[e['Action'] for e in events if e.get('Test')==test and e['Action'] in ['pass','fail','skip']]})
out=pathlib.Path('/tmp/wave2-07-gate/mutant-coverage.json');out.write_text(json.dumps(rows,indent=2));complete=[r for r in rows if 'Node' in r['caught'] and 'emitted JavaScript' in r['caught']];print(len(complete),'complete two-backend catches of',len(rows));print('canaries',[r['rule'] for r in rows if r['native_canary']]);print('failures',[(r['rule'],r['terminal']) for r in rows if 'fail' in r['terminal']])
