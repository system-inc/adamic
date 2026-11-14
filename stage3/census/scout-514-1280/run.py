import os,subprocess,time,json,hashlib
from pathlib import Path
out=Path('/workspace/scratch/scout-514-1280/stage3/census/scout-514-1280')
for name,root in [('base','/workspace/scratch/scout-adapted'),('a','/workspace/scratch/scout-a'),('b','/workspace/scratch/scout-b'),('c','/workspace/scratch/scout-c')]:
 d=out/name;env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=root+'/src/compiler/sys.ts',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_TIMES=str(d/'latent-times.jsonl'))
 cmd=['timeout','--kill-after=10s','1800','/workspace/scratch/scout-latent',root+'/src/compiler',str(d/'latent.jsonl')]
 start=time.time();print('START',name,flush=True)
 with (d/'latent.log.txt').open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 status={'command':cmd,'env':{k:env[k] for k in ['LATENT_FULL','LATENT_ASSERT_NO_OUTPUT','LATENT_ONLY_FILE','CENSUS_MODULE_FORMAT']},'exit_code':r.returncode,'seconds':time.time()-start,'binary_sha256':hashlib.sha256(Path('/workspace/scratch/scout-latent').read_bytes()).hexdigest()}
 (d/'latent-status.json').write_text(json.dumps(status,indent=2)+'\n');print('DONE',name,r.returncode,flush=True)
