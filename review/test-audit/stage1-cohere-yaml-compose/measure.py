import pathlib,json,subprocess,time
p=pathlib.Path('/tmp/u150/evidence');groups=[('TestFormatterMatchesGo',['TestFormatterMatchesGo']),('TestComposeMatchGo',['TestComposeMatchGo']),('TestCSTMatchesGo',['TestCSTMatchesGo']),('TestComposeMutants',['TestComposeMutants']),('TestCSTMutants',['TestCSTMutants']),('TestFileDriverUnion',['TestFileDriverUnion']),('TestFileDriver family',[f'TestFileDriver_{i:03d}' for i in range(8)]),('TestBundledParserDifference',['TestBundledParserDifference']),('TestFileDriver_Setup',['TestFileDriver_Setup'])];(p/'groups.json').write_text(json.dumps(groups,indent=2));rs=[]
for j in range(1,4):
 for i,(g,m) in enumerate(groups):
  s=time.monotonic();cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+'|'.join(m)+')$']
  with (p/f'group-{i}-timing-{j}.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  item=dict(group=g,index=i,run=j,exit=r.returncode,seconds=time.monotonic()-s,command=cmd);rs.append(item);(p/'group-timings.json').write_text(json.dumps(rs,indent=2));print(item,flush=True)
  if r.returncode:raise SystemExit('Baseline row failed or cooked; stop for inspection.')
