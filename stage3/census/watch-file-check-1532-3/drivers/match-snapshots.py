import pathlib,subprocess as p,time,json,os
s=pathlib.Path('/workspace/scratch/watch1532-recut');a=s/'composed';b=s/'fresh';r=s/'work';names=['93-builtin-imports','99-correctness-core-multimap','99-correctness-single-watcher-forwarder','99-correctness-timer-closures','99-wasm-crypto-import']
while p.run(['git','-C',str(b),'rev-parse','--verify','refs/stage3/99-wasm-crypto-import-after'],stdout=p.DEVNULL,stderr=p.DEVNULL).returncode:time.sleep(2)
rows=[]
for d in names:
 for state in ['before','after']:
  ref='refs/stage3/'+d+'-'+state
  old=p.check_output(['git','-C',str(a),'rev-parse',ref],text=True).strip();new=p.check_output(['git','-C',str(b),'rev-parse',ref],text=True).strip();assert old==new,(d,state,old,new);rows.append(dict(directory=d,state=state,preparationTree=old,freshTree=new,equal=True))
(s/'logs/snapshot-match.json').write_text(json.dumps(rows,indent=2)+'\n');print('All 10 proof snapshot trees equal official fresh apply snapshots',flush=True)
