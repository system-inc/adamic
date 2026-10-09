import pathlib,json,subprocess,os,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-json-shards';allrows=json.loads((p/'matrix-rows.txt').read_text())
groups={
'R1':['TestJSONPortShardUnion'],'R2':['TestJSONPortShardDisagreement'],'R3':['TestJSONHashShardsStayStable'],
'R4':[n for n in allrows if n.startswith('TestPortMatchesGoCohere_') and int(n.rsplit('_',1)[1])<2048]+['TestPortMatchesGoCohereUnion'],
'R5':[n for n in allrows if n.startswith('TestPortMatchesGoCohere_') and int(n.rsplit('_',1)[1])>=2048],
'R6':[n for n in allrows if n.startswith('TestUpstreamRepositoryCorpusParity_') and len(n.rsplit('_',1)[1])==4 and n[-1].isdigit()]+['TestUpstreamRepositoryCorpusParityUnion'],
'R7':['TestJSONUpstreamShardDisagreement'], 'R8':['TestUpstreamRepositoryCorpusParity_Setup','TestUpstreamRepositoryCorpusParity'],
'R9':['TestUpstreamRepositoryCorpusParityShardProof'],
'R10':[f'TestUpstreamRepositoryCorpusParity_{i:03}' for i in range(16)],'R11':['TestProduct_JSONUpstreamOracle']}
(p/'timing-selectors.txt').write_text(json.dumps(groups,indent=2)+'\n')
env=os.environ.copy();env.update(ADAMIC_JSON_PRETTIER='/tmp/u104/prettier',ADAMIC_JSON_BENCH='1')
for row,names in groups.items():
 sel='^('+'|'.join(names)+')$'
 for repeat in range(1,4):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',sel];start=time.monotonic()
  with (p/f'{row}-time{repeat}.log').open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  record={'row':row,'repeat':repeat,'command':cmd,'wall_seconds':time.monotonic()-start,'exit':r.returncode}
  with (p/'timing-runs.txt').open('a') as f:f.write(json.dumps(record)+'\n')
  print(row,repeat,r.returncode,round(record['wall_seconds'],3),flush=True)
