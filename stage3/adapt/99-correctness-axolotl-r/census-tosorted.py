"""Ordinary per-root checks, using identical inspect code and batches of at most 20."""
import hashlib,json,os,signal,subprocess,tarfile
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-axolotl-r';scratch=Path('/workspace/scratch/tosorted/step2');raw=scratch/'raw'
before=Path('/workspace/scratch/tosorted/before/src/compiler');after=scratch/'after/src/compiler'
roots=sorted(p.relative_to(after) for p in after.rglob('*.ts'));assert len(roots)==79
assert roots==sorted(p.relative_to(before) for p in before.rglob('*.ts'))
manifest=json.loads((scratch/'ready-for-census.json').read_text())
def frozen():
 assert manifest=={str(p.relative_to(scratch/'after/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((scratch/'after/src').rglob('*')) if p.is_file()}
frozen()
allrows=[]
def git(*args):
 with (raw/'census-push.log').open('a') as log:subprocess.run(['git',*args],cwd=repo,stdout=log,stderr=log,check=True)
def run(side,groupdir,relative):
 tree,binary=(before,'census-batch-base') if side=='37f8314b' else (after,'census-batch')
 paths=[str(tree/p) for p in relative];input=groupdir/(side+'-roots.json');input.write_text(json.dumps(paths,indent=2)+'\n');output=groupdir/(side+'.jsonl')
 env=dict(os.environ,GOMAXPROCS='2')
 with (groupdir/(side+'.log')).open('w') as log:
  process=subprocess.Popen([str(scratch/'bin'/binary),str(input),str(output)],stdout=log,stderr=log,env=env,start_new_session=True)
  try:code=process.wait(timeout=3600)
  except BaseException:
   os.killpg(process.pid,signal.SIGTERM)
   try:process.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
   raise
 assert code==0,(side,code)
 rows=[json.loads(line) for line in output.read_text().splitlines()];assert len(rows)==len(paths)
 assert [r['roots'] for r in rows]==[[p] for p in paths]
 return side,rows
for start in range(0,79,20):
 frozen()
 group=start//20+1;directory=raw/f'census-group-{group}';directory.mkdir();relative=roots[start:start+20]
 with ThreadPoolExecutor(max_workers=2) as pool:results=dict(pool.map(lambda side:run(side,directory,relative),['37f8314b','after']))
 pairs=[{'root':str(p),'before':a,'after':b} for p,a,b in zip(relative,results['37f8314b'],results['after'])];allrows.extend(pairs)
 (out/'tosorted-census.json').write_text(json.dumps(allrows,indent=2)+'\n')
 (directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(directory.iterdir()) if p.is_file()))
 archive=out/f'tosorted-group-{group}.tar.gz'
 with tarfile.open(archive,'w:gz') as tar:tar.add(directory,arcname=f'group-{group}')
 (out/f'tosorted-group-{group}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
 frozen()
 git('add',str(out.relative_to(repo)));git('commit','-m',f'99r: checkpoint toSorted compiler census group {group}');git('push','origin','HEAD:refs/heads/typescript/tosorted-readonly')
 print(f'group {group}: {len(relative)} roots per side, committed and pushed',flush=True)
