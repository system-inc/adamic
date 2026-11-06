#!/usr/bin/env python3
"""Run at most four resumable pieces. Preserve and push any failed piece, then stop."""
import hashlib,io,json,os,pathlib,signal,subprocess,sys,tarfile,time
p=pathlib.Path(__file__).resolve().parent
plan=json.loads((p/'plan.json').read_text())['files']
# Small entry/namespace files first, then compiler files; each eventual file appears once.
plan=sorted(plan,key=lambda n:(n.startswith('src/compiler/'),n))
groups=[json.loads(x.read_text()) for x in (p/'pieces').glob('*.json')] if (p/'pieces').exists() else []
completed={r['file'] for g in groups for r in g.get('records',[g])}
remaining=[n for n in plan if n not in completed]
batches=[];batch=[];size=0
for name in remaining:
 weight=(pathlib.Path('/workspace/v3-main-adapted-complete')/name).stat().st_size
 if batch and (len(batch)>=5 or size+weight>200000):batches.append(batch);batch=[];size=0
 batch.append(name);size+=weight
if batch:batches.append(batch)
for names in batches[:int(sys.argv[1]) if len(sys.argv)>1 else 4]:
 name=','.join(names)
 stem=names[0].replace('/','_').removesuffix('.ts') if len(names)==1 else 'group_'+hashlib.sha256('\n'.join(names).encode()).hexdigest()[:12]
 logpath=pathlib.Path('/workspace/scratch')/('v3-piece-'+stem+'.log')
 with logpath.open('w') as log:
  process=subprocess.Popen(['python3',str(p/'run-piece.py'),*names],stdout=log,stderr=log,start_new_session=True)
  try:code=process.wait(timeout=24*60)
  except BaseException:
   os.killpg(process.pid,signal.SIGTERM)
   try:process.wait(timeout=5)
   except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
   code=124
 if code:
  scratch=pathlib.Path('/workspace/scratch/v3-pieces')/stem
  files={x.name:x.read_bytes() for x in scratch.iterdir() if x.is_file()} if scratch.exists() else {}
  files['runner.log']=logpath.read_bytes()
  files['failure.json']=json.dumps(dict(file=name,returncode=code,status='failed piece; excluded from totals'),indent=2).encode()
  files['SHA256SUMS']=''.join(f'{hashlib.sha256(d).hexdigest()}  {n}\n' for n,d in sorted(files.items())).encode()
  with tarfile.open(p/'stopped-piece.tar.gz','w:gz') as t:
   for n,d in sorted(files.items()):
    i=tarfile.TarInfo(n);i.size=len(d);i.mtime=0;t.addfile(i,io.BytesIO(d))
  (p/'STOPPED.md').write_text(f'# Piece stopped\n\nFile: {name}. Runner exit: {code}.\n\nRaw outputs and the precise error are in stopped-piece.tar.gz/runner.log. This piece is excluded from closure totals.\n')
  subprocess.run(['python3',str(p/'update-report.py')],check=True)
  subprocess.run(['git','add',str(p)],check=True)
  subprocess.run(['git','commit','-m','census: retain failed views-v3 price piece '+name],check=True)
  subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/views-v3-price'],check=True)
  print('STOPPED '+name+' exit '+str(code),flush=True)
  raise SystemExit(code)
 print(logpath.read_text(),flush=True)
print('Batch complete',flush=True)
