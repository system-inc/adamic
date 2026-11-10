import concurrent.futures,json,os,pathlib,subprocess,sys,time
root=pathlib.Path('/workspace/adamic'); work=pathlib.Path('/workspace/scratch/debug-display-2-lane-counts-rerun');work.mkdir(exist_ok=True)
plan=json.loads((root/'stage3/lane/shard-plan.json').read_text())
env=dict(os.environ,STAGE3_ADAPTED_TREE=sys.argv[1],STAGE3_SHARD_SECONDS='240',NODE_OPTIONS='--require=/workspace/adamic/stage3/adapt/99-correctness-axolotl-t/debug-display-2-preload.cjs',DEBUG_DISPLAY_COUNTS=str(work/'counts.jsonl'))
for k in ['VSCODE_INSPECTOR_OPTIONS','NODE_INSPECTOR_IPC']:env.pop(k,None)
def run(unit):
 name=unit['name'];start=time.monotonic()
 with (work/(name+'.log')).open('wb') as log:
  code=subprocess.run(['python3',str(root/'stage3/lane/shard.py'),name,str(work/'pieces')],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 result={'name':name,'child_exit':code,'seconds':time.monotonic()-start}
 print(json.dumps(result),flush=True)
 return result
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:results=list(pool.map(run,plan['units']))
totals={n:0 for n in ['attachFlowNodeDebugInfoWorker','attachNodeArrayDebugInfoWorker','attachDebugPrototypeIfDebug','enableDebugInfo']}
records=0
for line in (work/'counts.jsonl').read_text().splitlines():
 row=json.loads(line);records+=1
 for name,count in row['counts'].items():totals[name]+=count
report={'units':len(results),'results':results,'process_records':records,'counts':totals}
(work/'summary.json').write_text(json.dumps(report,indent=2)+'\n')
sys.exit(0 if all(r['child_exit']==0 for r in results) and not any(totals.values()) else 1)
