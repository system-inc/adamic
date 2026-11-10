import concurrent.futures,json,os,pathlib,subprocess,sys,time,hashlib,gzip
root=pathlib.Path('/workspace/adamic');work=pathlib.Path('/workspace/scratch/debug-display-final-lane');work.mkdir(exist_ok=True);plan=json.loads((root/'stage3/lane/shard-plan.json').read_text());env=dict(os.environ,STAGE3_ADAPTED_TREE=sys.argv[1],STAGE3_SHARD_SECONDS='240')
for key in ['NODE_OPTIONS','DEBUG_DISPLAY_COUNTS','VSCODE_INSPECTOR_OPTIONS','NODE_INSPECTOR_IPC']:env.pop(key,None)
def run(unit):
 name=unit['name'];start=time.monotonic()
 with (work/(name+'.log')).open('wb') as log:code=subprocess.run(['python3',str(root/'stage3/lane/shard.py'),name,str(work/'pieces')],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 result={'name':name,'wrapper_child_exit':code,'seconds':time.monotonic()-start};print(json.dumps(result),flush=True);return result
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:results=list(pool.map(run,plan['units']))
totals={'pass':0,'fail':0,'skip':0};accepted=[];errors=[]
with gzip.open(work/'ordinary-reports.jsonl.gz','wt') as archive:
 for unit,result in zip(plan['units'],results):
  p=work/'pieces'/unit['name']/'report.json';report=json.loads(p.read_text());archive.write(json.dumps(report)+'\n')
  if result['wrapper_child_exit'] or report['status']!='pass' or report['errors'] or report['counts']!=unit['expect']:errors.append({'unit':unit['name'],'result':result,'errors':report['errors']})
  for key,value in report['counts'].items():totals[key]+=value
  accepted.append({**result,'counts':report['counts'],'upstream_child_exit':report.get('upstream_exit'),'report_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'inputs_hash':report.get('inputs_hash')})
if totals!=plan['counts']:errors.append({'totals':totals,'expected':plan['counts']})
(work/'summary.json').write_text(json.dumps({'status':'pass' if not errors else 'fail','units':len(results),'totals':totals,'errors':errors,'accepted':accepted,'instrumented':False},indent=2)+'\n');sys.exit(1 if errors else 0)
