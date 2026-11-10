import pathlib,subprocess,json,os,sys
side=sys.argv[1]; repo=pathlib.Path('/workspace/never-call-report' if side=='parent' else '/workspace/never-call-head'); out=pathlib.Path('/workspace/never-call-tools')/side; out.mkdir(parents=True,exist_ok=True)
dep=pathlib.Path('/workspace/never-call-report/cohere')
if side=='head':
 p=repo/'cohere'
 if not p.is_symlink():p.rmdir();p.symlink_to(dep,target_is_directory=True)
module=json.loads(subprocess.check_output(['go','mod','edit','-json'],cwd=repo,text=True,env={**os.environ,'GOWORK':'off'}))
reps=[r for r in module['Replace'] if r['New']['Path'].startswith('./cohere') and r['Old']['Path']!='github.com/microsoft/TypeScript/tsc']
work=out/'go.work';work.write_text('go 1.27\nuse (\n'+str(repo)+'\n'+str(dep/'TypeScript/tsc')+'\n)\nreplace (\n'+''.join(r['Old']['Path']+' => '+str(dep/pathlib.Path(r['New']['Path']).relative_to('cohere'))+'\n' for r in reps)+')\n')
env={**os.environ,'GOWORK':str(work)}
for phase,cmd in [('overlay',['python3',str(repo/'stage3/census/latent/make_overlay.py'),str(repo),str(out/'overlay')]),('build',['go','build','-buildvcs=false','-overlay='+str(out/'overlay/overlay.json'),'-o',str(out/'census'),'./stage3/census/latent/tool'])]:
 with (out/(phase+'.log')).open('wb') as f:r=subprocess.run(cmd,cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
 print(side,phase,'child exit code:',r.returncode,flush=True)
 if r.returncode:sys.exit(r.returncode)
