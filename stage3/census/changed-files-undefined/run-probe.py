import hashlib,json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
label,tree=sys.argv[1:]; tree=Path(tree)
roots=sorted(tree.joinpath('src/compiler').rglob('*.ts'))
assert len(roots)==79,len(roots)
source=tree/'src/compiler/sys.ts'
env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_') and k!='CENSUS_MODULE_FORMAT'}
env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(source),LATENT_TIMES=str(out/(label+'-times.jsonl')))
command=[str(out/'latent-census'),str(tree/'src/compiler'),str(out/(label+'.jsonl'))]
start=time.monotonic()
with (out/(label+'.log')).open('w') as log:
 r=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
receipt=dict(command=command,flags={k:v for k,v in env.items() if k.startswith('LATENT_')},exit=r.returncode,seconds=time.monotonic()-start,roots=[str(p) for p in roots],sys_sha256=hashlib.sha256(source.read_bytes()).hexdigest())
(out/(label+'-receipt.json')).write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({k:receipt[k] for k in ['exit','seconds','sys_sha256']}),flush=True)
sys.exit(r.returncode)
