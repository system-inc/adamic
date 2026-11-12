import pathlib,subprocess,json,os,shutil
out=pathlib.Path(__file__).parent;repo=pathlib.Path('/workspace/adamic');tree=pathlib.Path('/workspace/scratch/tracing-after2');scratch=pathlib.Path('/workspace/scratch/tracing2-contract');rows=[];env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--max-old-space-size=3072')
def run(name,args):
 with (out/(name+'.log')).open('w') as log:child=subprocess.run(args,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(name=name,argv=args,exit=child.returncode));(out/'adapter-replay2-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(name,'child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
run('owner-replay2',['node',str(repo/'stage3/adapt/47-host-errors/meta-spread-pin-witness.cjs'),str(tree),str(scratch),str(out/'owner-replay2.json')])
for name in ['core.ts','performanceCore.ts','tracing.ts']:
 child=subprocess.run(['git','-C',str(tree),'show','refs/stage3/65-temporary-node-builtins-after:src/compiler/'+name],capture_output=True);print('65 snapshot child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
 (scratch/'src/compiler'/name).write_bytes(child.stdout)
run('65-replay2',['node',str(repo/'stage3/adapt/65-temporary-node-builtins/check.cjs'),str(scratch)])
p=out/'binding_mutants.cjs';text=p.read_text().replace('/tracing-after\'','/tracing-after2\'');(out/'binding_mutants2.cjs').write_text(text)
for case in ['clean','0','1','2']:run('bindings2-'+case,['node',str(out/'binding_mutants2.cjs'),case])
