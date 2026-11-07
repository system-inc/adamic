import json,os,shutil,signal,subprocess,time
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');raw=root/'raw';out=repo/'stage3/adapt/99-correctness-core-takewhile';begin=time.monotonic()
while not (raw/'apply-after.log').read_text().rstrip().endswith('/workspace/scratch/takewhile/after'):
    if time.monotonic()-begin>1400: raise SystemExit('Complete adapted apply did not arrive within bound')
    time.sleep(2)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='',NODE_DISABLE_COMPILE_CACHE='1');env.pop('TW_CALL_DIR',None);phases={}
def phase(name,args,cwd=repo):
    begin=time.monotonic()
    with (raw/(name+'.log')).open('w') as log:
        child=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=1400)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    phases[name]={'exit':code,'seconds':time.monotonic()-begin};(raw/'proof-preparation.json').write_text(json.dumps(phases,indent=2)+'\n');assert code==0,(name,code);print(name,code,flush=True)
(root/'after/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
phase('build-after',['npm','run','build'],root/'after');(root/'after/node_modules').unlink()
phase('comparison',['node',str(out/'compare.cjs'),str(root/'before'),str(root/'after'),str(raw)])
phase('actual-witness',['node',str(out/'witness.cjs'),str(root/'before'),str(root/'after')])
mutant=root/'mutant';shutil.copytree(root/'after/src',mutant/'src');pin=json.loads((out/'owners.json').read_text())[0];file=mutant/'src/compiler/core.ts';text=file.read_bytes().decode();ending='\r\n' if '\r\n' in text else '\n';a=pin['after'].replace('\n',ending);b=pin['before'].replace('\n',ending);assert text.count(a)==1;file.write_bytes(text.replace(a,b).encode());assert file.read_bytes()==(root/'before/src/compiler/core.ts').read_bytes()
phases['mutant']={'originalCoreBytesRestored':True};(raw/'proof-preparation.json').write_text(json.dumps(phases,indent=2)+'\n');print('READY',flush=True)
