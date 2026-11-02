import csv,hashlib,json,os,pathlib,signal,subprocess,time
root=pathlib.Path('/workspace/scratch/carry2-subset');repo=pathlib.Path('/workspace/adamic')
manifest=list(csv.DictReader((repo/'stage3/census/latent/evidence/carry-facts/adapted-source-shas.csv').open()))
for r in manifest:assert hashlib.sha256((pathlib.Path('/workspace/scratch/tree')/r['path']).read_bytes()).hexdigest()==r['sha256'],r['path']
for mode,binary in [('baseline','/workspace/scratch/carry2-baseline-census'),('carried','/workspace/scratch/carry2-carried-census')]:
 for name in ['utilities.ts','binder.ts','emitter.ts']:
  out=root/mode/name;out.mkdir(parents=True,exist_ok=True)
  env=dict(os.environ,LATENT_ONLY_FILE='/workspace/scratch/tree/src/compiler/'+name,LATENT_TIMES=str(out/'latent-times.jsonl'),LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1')
  start=time.monotonic();observed_peak=0;over=False
  with (out/'run.log').open('w') as log:
   child=subprocess.Popen([binary,'/workspace/scratch/tree/src/compiler',str(out/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
   while True:
    pid,status,usage=os.wait4(child.pid,os.WNOHANG)
    if pid:child.returncode=os.waitstatus_to_exitcode(status);break
    try:
     lines=pathlib.Path(f'/proc/{child.pid}/status').read_text().splitlines()
     observed_peak=max(observed_peak,max(int(x.split()[1]) for x in lines if x.startswith(('VmRSS:','VmHWM:'))))
    except (FileNotFoundError,ValueError,ProcessLookupError):pass
    if observed_peak>12*1024*1024:
     over=True
     try:os.kill(child.pid,signal.SIGKILL)
     except ProcessLookupError:pass
    time.sleep(.25)
  resources=dict(wall_seconds=time.monotonic()-start,peak_rss_kib=usage.ru_maxrss,observed_peak_rss_kib=observed_peak,exit_code=child.returncode,over_12_gib=over)
  (out/'resources.json').write_text(json.dumps(resources,indent=2)+'\n');print(mode,name,resources,flush=True)
  if child.returncode:raise SystemExit(1)
  if mode=='carried':
   a=(root/'baseline'/name/'latent.jsonl').read_bytes();b=(out/'latent.jsonl').read_bytes()
   if a!=b:
    import itertools
    for i,(x,y) in enumerate(itertools.zip_longest(a.splitlines(),b.splitlines()),1):
     if x!=y:print('FIRST DIFFERING LINE',name,i,flush=True);break
    raise SystemExit(2)
