"""Bounded full-mode pairing for the two compiler timer owners."""
import concurrent.futures,hashlib,json,os,pathlib,signal,subprocess,time
D=pathlib.Path(__file__).resolve().parent
BINARY=pathlib.Path('/workspace/scratch/timer-latent-tip')
TREES={'before':pathlib.Path('/workspace/scratch/timer-main-tip-final'),'after':pathlib.Path('/workspace/scratch/timer-after-tip')}
FILES=['watchPublic.ts','tsbuildPublic.ts']
# The recut changes the commit message only. Verify the entire repository tree.
repo='/workspace/adamic'
def git(*args):return subprocess.check_output(['git','-C',repo,*args],text=True).strip()
assert git('rev-parse','f37253b3^{tree}')==git('rev-parse','108e46471^{tree}')
manifest=json.loads((D/'input-source-sha256.json').read_text())
for label,tree in TREES.items():
 expected=manifest['main' if label=='before' else 'adapted']
 actual={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
 assert actual==expected,(label,'source drift')
provenance=dict(main=git('rev-parse','0aba99914'),adapted=git('rev-parse','f37253b3'),originalAdaptation=git('rev-parse','108e46471'),identicalAdaptationTree=git('rev-parse','f37253b3^{tree}'),binary=str(BINARY),binarySha256=hashlib.sha256(BINARY.read_bytes()).hexdigest(),sources={label:{file:hashlib.sha256((tree/'src/compiler'/file).read_bytes()).hexdigest() for file in FILES} for label,tree in TREES.items()})
(D/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
def run(job):
 label,file=job;tree=TREES[label];name=label+'-'+file.removesuffix('.ts');result=D/(name+'.jsonl');command=[str(BINARY),str(tree/'src/compiler'),str(result)]
 env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),LATENT_ASSERT_NO_OUTPUT='1',GOMAXPROCS='1',GOMEMLIMIT='3GiB')
 begin=time.monotonic();timed_out=False
 with (D/(name+'.log')).open('w') as stream:
  child=subprocess.Popen(command,env=env,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=600)
  except subprocess.TimeoutExpired:
   timed_out=True;os.killpg(child.pid,signal.SIGTERM)
   try:code=child.wait(timeout=5)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);code=child.wait()
 record=dict(command=command,environment={k:env[k] for k in ['LATENT_FULL','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT','GOMAXPROCS','GOMEMLIMIT']},child_exit=code,timed_out=timed_out,seconds=time.monotonic()-begin)
 (D/(name+'.status.json')).write_text(json.dumps(record,indent=2)+'\n')
 print(json.dumps(dict(name=name,**record)),flush=True)
 return code
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:codes=list(pool.map(run,[(label,file) for label in TREES for file in FILES]))
assert codes==[0,0,0,0],codes
