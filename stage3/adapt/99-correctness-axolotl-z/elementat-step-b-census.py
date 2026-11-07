import hashlib,json,os,signal,subprocess,tarfile,time
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b/step-b');out=repo/'stage3/adapt/99-correctness-axolotl-z';after=s/'after/src/compiler';before=Path('/workspace/scratch/elementat-b/before/src/compiler');roots=sorted(p.relative_to(after) for p in after.rglob('*.ts'));assert len(roots)==79
manifest={str(p.relative_to(s/'after/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'after/src').rglob('*')) if p.is_file()};(s/'raw/census-source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
allrows=[]
def git(*args):
 with (s/'census-git.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def run(side,relative,directory,chunk):
 tree=before if side=='before' else after;input=directory/f'{side}-{chunk}-roots.json';output=directory/f'{side}-{chunk}.jsonl';input.write_text(json.dumps([str(tree/p) for p in relative],indent=2)+'\n')
 with (directory/f'{side}-{chunk}.log').open('w') as log:
  child=subprocess.Popen([str(s/'bin/census-batch'),str(input),str(output)],env=dict(os.environ,GOMAXPROCS='1',GOMEMLIMIT='3GiB'),stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=1000)
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 assert code==0,(side,chunk,code)
 rows=[json.loads(l) for l in output.read_text().splitlines()];assert len(rows)==len(relative) and all(len(r['roots'])==1 for r in rows);return side,rows
for start in range(0,79,20):
 begin=time.monotonic();group=start//20+1;directory=s/f'census-{group}';directory.mkdir();relative=roots[start:start+20];indexed={'before':{},'after':{}}
 # Two independent ordinary inspectors at once, paired with the two Node replay jobs.
 with ThreadPoolExecutor(max_workers=2) as pool:
  jobs=[pool.submit(run,side,relative[i::2],directory,i) for side in ['before','after'] for i in range(2)]
  for future in as_completed(jobs):
   side,rows=future.result();indexed[side].update({r['roots'][0]:r for r in rows})
 pairs=[{'root':str(p),'before':indexed['before'][str(before/p)],'after':indexed['after'][str(after/p)]} for p in relative];allrows+=pairs
 assert manifest=={str(p.relative_to(s/'after/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'after/src').rglob('*')) if p.is_file()}
 (directory/'group.json').write_text(json.dumps({'group':group,'roots':len(relative),'seconds':time.monotonic()-begin,'pairs':pairs},indent=2)+'\n');assert time.monotonic()-begin<1400
 (out/'elementat-census.json').write_text(json.dumps(allrows,indent=2)+'\n');(directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(directory.iterdir()) if p.is_file()))
 archive=out/f'elementat-census-{group}.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:tar.add(directory,arcname=f'group-{group}')
 (out/f'elementat-census-{group}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 git('add',str(out.relative_to(repo)));git('commit','-m',f'never cut alone: lands with chg_pyyr04zd, because the census reads its stop\n\n99z: paired ordinary per-file census group {group}, {len(relative)} roots per side');git('push','origin','typescript/elementat-any')
 print('group',group,'pushed',len(relative),'roots per side',flush=True)
print('COMPLETE 79 paired roots',flush=True)
