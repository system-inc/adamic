import csv,hashlib,itertools,json,os,pathlib,signal,subprocess,sys,time
root=pathlib.Path('/workspace/scratch/carry3-runs');source=pathlib.Path('/workspace/scratch/tree/src/compiler');repo=pathlib.Path('/workspace/adamic');names=['binder.ts','emitter.ts','utilities.ts']
for r in csv.DictReader((repo/'stage3/census/latent/evidence/carry-facts/adapted-source-shas.csv').open()):
 path=pathlib.Path('/workspace/scratch/tree')/r['path'];assert '\n' not in str(path);assert hashlib.sha256(path.read_bytes()).hexdigest()==r['sha256']
def last_unit(path):
 try:
  with path.open('rb') as f:
   f.seek(0,2);size=f.tell();f.seek(max(0,size-8192));lines=f.read().splitlines()
  return json.loads(lines[-1]) if lines else None
 except (OSError,ValueError):return None
def run(case,side,binary,full,files):
 out=root/case/side;out.mkdir(parents=True,exist_ok=True);env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_FULL=str(full),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'latent-times.jsonl'))
 if files:env['LATENT_ONLY_FILE']='\n'.join(str(source/n) for n in files)
 (out/'invocation.json').write_text(json.dumps(dict(argv=[binary,str(source),str(out/'latent.jsonl')],environment={k:v for k,v in env.items() if k.startswith('LATENT_')}),indent=2)+'\n')
 start=time.monotonic();peak=0;over=False;where=None
 with (out/'run.log').open('w') as log:
  p=subprocess.Popen([binary,str(source),str(out/'latent.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
  while True:
   pid,status,usage=os.wait4(p.pid,os.WNOHANG)
   if pid:p.returncode=os.waitstatus_to_exitcode(status);break
   try:
    statuslines=pathlib.Path(f'/proc/{p.pid}/status').read_text().splitlines();peak=max(peak,max(int(x.split()[1]) for x in statuslines if x.startswith(('VmRSS:','VmHWM:'))))
   except (FileNotFoundError,ValueError,ProcessLookupError):pass
   if peak>14*1024*1024:
    over=True;where=last_unit(out/'latent-times.jsonl')
    try:os.kill(p.pid,signal.SIGKILL)
    except ProcessLookupError:pass
   time.sleep(.2)
 resources=dict(wall_seconds=time.monotonic()-start,peak_rss_kib=usage.ru_maxrss,observed_peak_rss_kib=peak,exit_code=p.returncode,over_14_gib=over,last_completed_unit=where or last_unit(out/'latent-times.jsonl'))
 (out/'resources.json').write_text(json.dumps(resources,indent=2)+'\n');print(case,side,json.dumps(resources),flush=True)
 if p.returncode:raise SystemExit(1)
 return out/'latent.jsonl'
def identical(a,b,label):
 x=a.read_bytes();y=b.read_bytes()
 if x!=y:
  for n,(u,v) in enumerate(itertools.zip_longest(x.splitlines(),y.splitlines()),1):
   if u!=v:
    info=dict(proof=label,first_differing_line=n,baseline=str(a),changed=str(b));(root/'failure.json').write_text(json.dumps(info,indent=2)+'\n');print('IDENTITY FAILED',json.dumps(info),flush=True);break
  raise SystemExit(2)
 print('IDENTICAL',label,hashlib.sha256(x).hexdigest(),flush=True)
base='/workspace/scratch/carry3-baseline-census';carried='/workspace/scratch/carry3-carried-census';multi='/workspace/scratch/carry3-list-baseline-census'
a=run('first-error-all79','baseline',base,0,[]);b=run('first-error-all79','carried',carried,0,[]);identical(a,b,'first-error all 79, one process per side')
a=run('full-multi','baseline',multi,1,names);b=run('full-multi','carried',carried,1,names);identical(a,b,'full three files, one process per side')
for name in names:
 a=run('full-single-'+name,'baseline',base,1,[name]);b=run('full-single-'+name,'carried',carried,1,[name]);identical(a,b,'full single '+name)
for side in ['baseline','carried']:
 multi_path=root/'full-multi'/side/'latent.jsonl';lines=multi_path.read_bytes().splitlines(keepends=True);multi_records={json.loads(line)['file']:line for line in lines[1:]};assert set(multi_records)=={str(source/n) for n in names}
 combined=[lines[0]]
 for name in names:
  path=root/('full-single-'+name)/side/'latent.jsonl';single=path.read_bytes().splitlines(keepends=True);assert single[0]==lines[0];assert len(single)==2
  if single[1]!=multi_records[str(source/name)]:
   (root/'failure.json').write_text(json.dumps(dict(proof='single-file vs one-process full record',side=side,file=name,single=str(path),multi=str(multi_path)),indent=2)+'\n');print('SINGLE/MULTI FILE RECORD DIFFERS',side,name,flush=True);raise SystemExit(3)
  combined.append(single[1])
 path=root/('full-singles-combined-'+side+'.jsonl');path.write_bytes(b''.join(combined));identical(multi_path,path,'full '+side+' one-process vs three single-file records')
print('ALL PROOFS PASSED',flush=True)
