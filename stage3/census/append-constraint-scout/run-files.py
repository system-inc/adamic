"""Independent scan-only processes; checkpoint every file. No native body lowering."""
import concurrent.futures,hashlib,json,os,pathlib,signal,subprocess,sys,time
binary,adapted,out=map(pathlib.Path,sys.argv[1:]);root=adapted/'src/compiler';files=sorted(root.rglob('*.ts'));assert len(files)==79
out.mkdir(parents=True,exist_ok=True);started=time.monotonic()
manifest={'source_sha256':{str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in files},'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'workers':4,'timeout_seconds_per_file':1500}
(out/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
def run(p):
 name=str(p.relative_to(root));key=name.replace('/','__');env=dict(os.environ,LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(p));t=time.monotonic();timeout=False
 with (out/(key+'.log')).open('w') as log:
  child=subprocess.Popen([str(binary),str(root),str(out/(key+'.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=1500)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait();code=124;timeout=True
 rows=[]
 try:rows=[json.loads(x) for x in (out/(key+'.jsonl')).read_text().splitlines()]
 except (ValueError,FileNotFoundError):pass
 records=[r for r in rows if 'file' in r]
 r={'file':name,'exit':code,'seconds':round(time.monotonic()-t,3),'timeout':timeout,'complete_trailer':bool(rows and rows[-1].get('status')=='relation_report'),'file_records':len(records),'units':sum(len(r.get('units',[])) for r in records),'findings':sum(len(r.get('findings',[])) for r in records)}
 (out/(key+'.result.json')).write_text(json.dumps(r,indent=2)+'\n');return r
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for future in concurrent.futures.as_completed([pool.submit(run,p) for p in files]):
  r=future.result();results.append(r);(out/'results.json').write_text(json.dumps(sorted(results,key=lambda r:r['file']),indent=2)+'\n');print(len(results),r,flush=True)
(out/'summary.json').write_text(json.dumps({'seconds':round(time.monotonic()-started,3),'files':len(results),'all_complete':all(r['exit']==0 and r['complete_trailer'] and r['file_records']==1 and r['units']==0 for r in results)},indent=2)+'\n')
