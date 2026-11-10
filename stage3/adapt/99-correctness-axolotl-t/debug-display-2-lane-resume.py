import concurrent.futures,gzip,hashlib,json,os,pathlib,shutil,subprocess,sys,time
root=pathlib.Path('/workspace/adamic');work=pathlib.Path('/workspace/scratch/debug-display-2-lane-counts-rerun');unit=root/'stage3/adapt/99-correctness-axolotl-t'
plan=json.loads((root/'stage3/lane/shard-plan.json').read_text());units={u['name']:u for u in plan['units']}
prior={}
for line in pathlib.Path('/tmp/debug-display-2-lane-counts-rerun.log').read_text().splitlines():
 if not line.startswith('{'):continue
 row=json.loads(line);assert row['name'] not in prior;assert row['child_exit']==0
 report=json.loads((work/'pieces'/row['name']/'report.json').read_text());assert report['status']=='pass' and not report['errors'] and report['counts']==units[row['name']]['expect']
 prior[row['name']]={**row,'reused_completed_before_environment_refresh':True}
env=dict(os.environ,STAGE3_ADAPTED_TREE=sys.argv[1],STAGE3_SHARD_SECONDS='240',NODE_OPTIONS='--require='+str(unit/'debug-display-2-preload.cjs'),DEBUG_DISPLAY_COUNTS=str(work/'counts.jsonl'),PYTHONDONTWRITEBYTECODE='1')
for key in ['VSCODE_INSPECTOR_OPTIONS','NODE_INSPECTOR_IPC']:env.pop(key,None)
missing=[u for u in plan['units'] if u['name'] not in prior]
print(json.dumps({'verified_completed_shards':len(prior),'unfinished_shards_to_run':len(missing)}),flush=True)
def run(u):
 name=u['name'];piece=work/'pieces'/name
 if piece.exists():shutil.rmtree(piece)
 logpath=work/(name+'.log')
 if logpath.exists():logpath.rename(work/(name+'.interrupted.log'))
 start=time.monotonic()
 with logpath.open('wb') as log:code=subprocess.run(['python3',str(root/'stage3/lane/shard.py'),name,str(work/'pieces')],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 row={'name':name,'child_exit':code,'seconds':time.monotonic()-start,'reused_completed_before_environment_refresh':False};print(json.dumps(row),flush=True);return row
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:resumed=list(pool.map(run,missing))
by_name={**prior,**{r['name']:r for r in resumed}};results=[by_name[u['name']] for u in plan['units']];assert len(results)==398
errors=[];totals={'pass':0,'fail':0,'skip':0};accepted=[]
with gzip.GzipFile(filename=str(work/'accepted-reports.jsonl.gz'),mode='wb',mtime=0) as archive:
 for u,row in zip(plan['units'],results):
  p=work/'pieces'/u['name']/'report.json';report=json.loads(p.read_text());archive.write((json.dumps(report)+'\n').encode())
  if row['child_exit'] or report['status']!='pass' or report['errors'] or report['counts']!=u['expect']:errors.append({'name':u['name'],'result':row,'report':report})
  for k,v in report['counts'].items():totals[k]+=v
  accepted.append({**row,'counts':report['counts'],'upstream_exit':report.get('upstream_exit'),'report_sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
counts={n:0 for n in ['attachFlowNodeDebugInfoWorker','attachNodeArrayDebugInfoWorker','attachDebugPrototypeIfDebug','enableDebugInfo']};records=0
for line in (work/'counts.jsonl').read_text().splitlines():
 r=json.loads(line);records+=1
 for k,v in r['counts'].items():counts[k]+=v
if totals!=plan['counts']:errors.append({'totals':totals,'expected':plan['counts']})
if any(counts.values()):errors.append({'nonzero_counts':counts})
report={'status':'pass' if not errors else 'fail','units':398,'verified_reused_shards':len(prior),'resumed_shards':len(missing),'all_ordinary_guards_pass':not errors,'totals':totals,'counts':counts,'counter_process_records':records,'counter_records_include_interrupted_attempts':True,'errors':errors,'accepted':accepted}
(work/'summary.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps({k:v for k,v in report.items() if k not in ['accepted','errors']}),flush=True);sys.exit(1 if errors else 0)
