from pathlib import Path
import os,subprocess as s,time,json,hashlib
r=Path(__file__).resolve().parent
for side in ['before','after']:
 while not (r/(side+'-apply.exit')).exists():time.sleep(1)
 assert (r/(side+'-apply.exit')).read_text().strip()=='0',side
before=r/'before';after=r/'after';names=sorted(p.relative_to(before/'src') for p in (before/'src').rglob('*') if p.is_file());assert names==sorted(p.relative_to(after/'src') for p in (after/'src').rglob('*') if p.is_file())
changed=[str(p) for p in names if (before/'src'/p).read_bytes()!=(after/'src'/p).read_bytes()];assert changed==['compiler/sys.ts'],changed
f=before/'src/compiler/sys.ts';saved=f.read_bytes();stamp=f.stat().st_mtime_ns;results=[]
try:
 for side in ['before','after']:
  data=(r/side/'src/compiler/sys.ts').read_bytes();f.write_bytes(data)
  env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
  env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(f),LATENT_TIMES=str(r/('census-'+side+'-times.jsonl')),CENSUS_MODULE_FORMAT='CommonJSBundle',GOMAXPROCS='2',GOMEMLIMIT='6GiB')
  cmd=[str(r.parent/'latent'),str(before/'src/compiler'),str(r/('census-'+side+'.jsonl'))]
  code=s.run(['python3',str(r/'runproof.py'),'census-'+side,'0',*cmd],cwd='/workspace/scratch/watch-file-1532-merge',env=env).returncode
  results.append({'side':side,'sourceSha256':hashlib.sha256(data).hexdigest(),'environment':{k:v for k,v in env.items() if k.startswith('LATENT_') or k in ['CENSUS_MODULE_FORMAT','GOMAXPROCS','GOMEMLIMIT']}})
  assert f.read_bytes()==data,'census modified source'
  assert code==0,side
finally:
 f.write_bytes(saved);os.utime(f,ns=(f.stat().st_atime_ns,stamp))
 (r/'census-provenance.json').write_text(json.dumps({'sameCompilerPopulation':True,'changedSourceFiles':changed,'sourceRestored':f.read_bytes()==saved,'binarySha256':hashlib.sha256((r.parent/'latent').read_bytes()).hexdigest(),'runs':results},indent=2)+'\n')
print('CENSUS PAIR COMPLETE',flush=True)
