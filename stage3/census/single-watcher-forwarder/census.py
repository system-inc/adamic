import os,subprocess,json,time
from pathlib import Path
out=Path(__file__).resolve().parent
repo=Path('/workspace/scratch/single-watcher-forwarder')
for label,root in [('base','/workspace/scratch/single-watcher-before'),('adapted','/workspace/scratch/single-watcher-after'),('mutant','/workspace/scratch/single-watcher-mutant')]:
 d=out/label;d.mkdir(exist_ok=True)
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=root+'/src/compiler/sys.ts',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_TIMES=str(d/'latent-times.jsonl'))
 cmd=['python3',str(out/'run-command.py'),'census-'+label,'timeout','--kill-after=10s','1800',str(out/'latent'),root+'/src/compiler',str(d/'latent.jsonl')]
 started=time.time();rc=subprocess.run(cmd,env=env,cwd=repo).returncode
 (d/'latent-status.json').write_text(json.dumps({'exit_code':rc,'seconds':time.time()-started,'command':cmd,'env':{k:env[k] for k in ['LATENT_FULL','LATENT_ASSERT_NO_OUTPUT','LATENT_ONLY_FILE','CENSUS_MODULE_FORMAT']}},indent=2)+'\n')
 if rc:raise SystemExit(rc)
