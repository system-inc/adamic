from pathlib import Path
import subprocess,os,time,json,hashlib,sys
r=Path('/workspace/scratch/binding-views-resolution');name=sys.argv[1];tree=r/name;out=r/'evidence'/name;out.mkdir(parents=True,exist_ok=True)
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=r/'repo',text=True).strip();assert sha=='61636df700acf25afe229345128bfa4c193fda23'
file='resolutionCache.ts';env=os.environ|dict(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler'/file),CENSUS_MODULE_FORMAT='CommonJSBundle')
command=[str(r/'latent'),str(tree/'src/compiler'),str(out/(file+'.jsonl'))];start=time.monotonic()
with (out/'run.log.txt').open('w') as log:p=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
result=dict(commit=sha,command=command,environment={k:env[k] for k in ['LATENT_FULL','LATENT_ASSERT_NO_OUTPUT','LATENT_ONLY_FILE','CENSUS_MODULE_FORMAT']},exit=p.returncode,mode='full',seconds=time.monotonic()-start,binary_sha256=hashlib.sha256((r/'latent').read_bytes()).hexdigest(),input_sha256=hashlib.sha256((tree/'src/compiler'/file).read_bytes()).hexdigest())
(out/'run.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
assert p.returncode==0;d=[json.loads(x) for x in (out/(file+'.jsonl')).read_text().splitlines()];assert d[-1]['status']=='relation_report';assert d[0]['latent_mode']=='full' and d[0]['module_format']=='CommonJSBundle'
