import pathlib,json,collections,re,sys,gzip
root=pathlib.Path(__file__).parent; mode=sys.argv[1]
path=root/f'{mode}.jsonl'
lines=path.read_text().splitlines() if path.exists() else gzip.open(str(path)+'.gz','rt').read().splitlines()
events=[]
for line in lines:
 try:events.append(json.loads(line))
 except ValueError:pass
counts=collections.Counter(); top=collections.Counter();skip=[];fail=[];times=[];single={};agreement=None;proofs=[]
for e in events:
 a=e.get('Action');t=e.get('Test');out=e.get('Output','')
 if t and a in ['pass','fail','skip']:
  counts[a]+=1
  if '/' not in t:top[a]+=1
  if a=='skip':skip.append(t)
  if a=='fail':fail.append(t)
 if t=='TestCompilerAndStage1Agree':
  if a=='pass':agreement=e['Elapsed']
  m=re.search(r'(Node|emitted JavaScript|sanitized native) shard (\d+)/(\d+): ([\d.]+) ms',out)
  if m:times.append(dict(backend=m[1],index=int(m[2]),count=int(m[3]),seconds=float(m[4])/1000))
  m=re.search(r'compiler shard (\d+)/(\d+) only file: (.*)',out)
  if m:single[int(m[1])]=m[3].strip()
 if t and t.startswith('TestChild') and a=='output' and any(s in out for s in ['killed and named','silent healthy','separate backstop']):proofs.append(out.strip())
longest=max(times,key=lambda x:x['seconds']) if times else None
if longest:longest=dict(longest,only_input=single.get(longest['index']))
meta=json.loads((root/f'{mode}-wall.json').read_text())
loadpath=root/f'{mode}-load.jsonl'
loadlines=loadpath.read_text().splitlines() if loadpath.exists() else gzip.open(str(loadpath)+'.gz','rt').read().splitlines()
loads=[json.loads(s) for s in loadlines]
burners=loads[-1]['burners'] if loads else []
summary=dict(code_sha='4678c7302cae4ab51e2378156e228e9a771e73b4',mode=mode,command=meta['command'],exit=meta['exit'],test_and_subtest_counts={k:counts[k] for k in ['pass','fail','skip']},top_level_counts={k:top[k] for k in ['pass','fail','skip']},skips_named=skip,failures_named=fail,wall_seconds=meta['wall_seconds'],nproc=meta['nproc'],cpu_max=meta['cpu_max'],load_before=meta['load_before'],load_after=meta['load_after'],load_1m_range=[min(float(s['load'].split()[0]) for s in loads),max(float(s['load'].split()[0]) for s in loads)],agreement_seconds=agreement,longest_shard=longest,shards=times,guard_proofs=proofs,burners_final=burners,all_burners_alive_throughout=all(len(s['burners'])==meta['nproc'] and all(b['alive'] for b in s['burners']) for s in loads) if mode=='loaded' else None,inputs=meta['inputs'])
(root/f'{mode}-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:summary[k] for k in ['mode','exit','test_and_subtest_counts','wall_seconds','agreement_seconds','longest_shard']},indent=2))
