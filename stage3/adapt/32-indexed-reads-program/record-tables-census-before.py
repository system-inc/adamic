from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
import subprocess,os,json,hashlib,time,tarfile,signal
s=Path('/workspace/scratch/gov-b');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/32-indexed-reads-program';tree=s/'full-before/src/compiler';roots=sorted(tree.rglob('*.ts'));assert len(roots)==79
manifest={str(p.relative_to(s/'full-before/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'full-before/src').rglob('*')) if p.is_file()};allrows=[]
def run(item):
 directory,chunk,paths=item;input=directory/f'roots-{chunk}.json';output=directory/f'rows-{chunk}.jsonl';input.write_text(json.dumps([str(p) for p in paths],indent=2)+'\n')
 with (directory/f'run-{chunk}.log').open('w') as log:
  child=subprocess.Popen([s/'census-batch',input,output],stdout=log,stderr=log,env=dict(os.environ,GOMAXPROCS='1',GOMEMLIMIT='3GiB'),start_new_session=True)
  try:code=child.wait(timeout=1300)
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 assert code==0;rows=[json.loads(l) for l in output.read_text().splitlines()];assert {r['roots'][0] for r in rows}=={str(p) for p in paths};return rows
for start in range(0,79,20):
 begun=time.monotonic();number=start//20+1;directory=s/f'census-before-{number}';directory.mkdir();subset=roots[start:start+20]
 with ThreadPoolExecutor(max_workers=2) as pool:rows=[row for part in pool.map(run,[(directory,i,subset[i::2]) for i in range(2)]) for row in part]
 assert manifest=={str(p.relative_to(s/'full-before/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'full-before/src').rglob('*')) if p.is_file()};seconds=time.monotonic()-begun;assert seconds<1500
 record={'group':number,'roots':len(subset),'seconds':seconds,'rows':sorted(rows,key=lambda r:r['roots'][0]),'sourceManifest':manifest};(directory/'group.json').write_text(json.dumps(record,indent=2)+'\n');allrows+=record['rows'];(s/'raw/census-before.json').write_text(json.dumps(allrows,indent=2)+'\n')
 (directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(directory.iterdir()) if p.is_file()))
 archive=out/f'record-tables-census-before-{number}.tar.gz';tmp=archive.with_suffix('.tmp')
 with tarfile.open(tmp,'w:gz') as tf:tf.add(directory,arcname=directory.name)
 tmp.replace(archive);(out/f'record-tables-census-before-{number}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n');print('completed baseline census group',number,'roots',len(subset),'seconds',seconds,flush=True)
print('COMPLETE',flush=True)
