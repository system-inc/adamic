import os,pathlib,subprocess as p,time,json
s=pathlib.Path('/workspace/scratch/watch1532-recut');r=s/'work';t=s/'composed';env=dict(os.environ,NODE_PATH=str(r/'stage3/api/node_modules'),ADAMIC_TYPESCRIPT_SOURCE=str(s/'typescript'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'gotmp'),GOCACHE=str(s/'go-cache'),STAGE3_PRODUCT_STORE=str(s/'fresh-products'))
proc=pathlib.Path('/proc/1451/cmdline')
while proc.exists() and b'apply.py' in proc.read_bytes():time.sleep(2)
p.run(['git','-C',str(t),'rev-parse','--verify','refs/stage3/99-wasm-crypto-import-before'],check=True)
with (s/'logs/wasm-pin.log').open('w') as f:z=p.run(['node',str(r/'stage3/adapt/99-wasm-crypto-import/adapt.cjs'),'--pin',str(t)],cwd=r,env=env,stdout=f,stderr=p.STDOUT)
print('wasm pin exit',z.returncode,flush=True);assert z.returncode==0
start=time.monotonic()
with (s/'logs/fresh-apply.log').open('w') as f:z=p.run([str(r/'stage3/apply.sh'),'--build-product',str(s/'fresh')],cwd=r,env=env,stdout=f,stderr=p.STDOUT)
(s/'logs/fresh-apply-result.json').write_text(json.dumps(dict(exit=z.returncode,seconds=time.monotonic()-start,command=[str(r/'stage3/apply.sh'),'--build-product',str(s/'fresh')],productStore=env['STAGE3_PRODUCT_STORE']),indent=2)+'\n');print('fresh apply exit',z.returncode,flush=True)
