from pathlib import Path
import subprocess, concurrent.futures, json, time, sys
root=Path('/workspace/adamic')
evidence=Path('/tmp/adamic-ambient-evidence')
phase=sys.argv[1]
paths=sorted(p for p in subprocess.check_output(['git','ls-files','-z'],cwd=root).decode().split('\0') if p.endswith(('.a','.ts')) and not p.endswith('.d.ts') and not p.startswith('cohere/') and 'node_modules' not in p.split('/'))
start=time.time()
if phase=='main':
 (evidence/'sweep-paths.txt').write_text('\n'.join(paths)+'\n')
else:
 assert paths==(evidence/'sweep-paths.txt').read_text().splitlines(), 'tracked paths changed'
previous={r['path']:r for r in map(json.loads,(evidence/'sweep-main.jsonl').read_text().splitlines())} if phase=='new' else {}
def check(path):
 row=previous.get(path,{'path':path})
 try:
  result=subprocess.run([str(evidence/('adamic-'+phase)),'c',path],cwd=root,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=120)
  row[phase]={'exit':result.returncode,'stderr':result.stderr.decode(errors='replace')}
 except subprocess.TimeoutExpired as error:
  row[phase]={'exit':'timeout','stderr':(error.stderr or b'').decode(errors='replace')}
 return row
output=evidence/('sweep-main.jsonl' if phase=='main' else 'sweep.jsonl')
with output.open('w') as log, concurrent.futures.ThreadPoolExecutor(max_workers=2 if phase=='main' else 4) as executor:
 for i,row in enumerate(executor.map(check,paths),1):
  log.write(json.dumps(row)+'\n');log.flush()
  if i%50==0 or i==len(paths):
   (evidence/('sweep-'+phase+'-progress.txt')).write_text(f'{i}/{len(paths)} in {time.time()-start:.1f}s\n')
print(phase,'programs',len(paths),'seconds',round(time.time()-start,1))
if phase=='main': raise SystemExit(0)
rows=[json.loads(line) for line in output.read_text().splitlines()]
new_markers=('Adamic 0.1 refuses an ambient ', 'Adamic 0.1 refuses a global augmentation;', 'Adamic 0.1 refuses a module augmentation or ambient external module;', 'Adamic 0.1 refuses a triple-slash reference ')
changed=[r for r in rows if any(marker in r['new']['stderr'].split('\n')[0] for marker in new_markers) and r['new']['stderr']!=r['main']['stderr']]
(evidence/'sweep-new-refusals.json').write_text(json.dumps(changed,indent=2)+'\n')
(evidence/'sweep-new-refusals.txt').write_text('\n'.join(r['path'] for r in changed)+'\n')
print('new first-stop ambient refusals',len(changed),'timeouts',sum(r[k]['exit']=='timeout' for r in rows for k in ('main','new')))
print((evidence/'sweep-new-refusals.txt').read_text())
