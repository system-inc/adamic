import pathlib,subprocess,os,json,sys
repo=pathlib.Path('/workspace/never-call-report');out=repo/'stage3/census/never-call-392cb60b';tree=pathlib.Path('/workspace/never-call-adapted')
for case,mode,source in [('debug-full','full',tree/'src/compiler'),('semver-full','full',tree/'src/compiler'),('tsc-first','first',tree/'src/tsc/tsc.ts')]:
 for side in ['parent','head']:
  name=case+'-'+side; env=dict(os.environ)
  for k in list(env):
   if k.startswith('LATENT_'):del env[k]
  env.update(LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/(name+'.units.jsonl')),LATENT_PREPASS_TIMES=str(out/(name+'.prepasses.log')))
  if mode=='full':env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/('src/compiler/debug.ts' if case=='debug-full' else 'src/compiler/semver.ts')))
  cmd=['python3','/tmp/never-call-run.py',name,str(repo if side=='parent' else pathlib.Path('/workspace/never-call-head')),str(out),str(pathlib.Path('/workspace/never-call-tools')/side/'census'),str(source),str(out/(name+'.jsonl'))]
  r=subprocess.run(cmd,env=env);print(name,'wrapper child exit code:',r.returncode,flush=True)
  with (out/'REPORT.md').open('a') as f:f.write('\nCompleted '+name+'; wrapper exit '+str(r.returncode)+'. See its timing and stream files.\n')
  for command in [['python3','/tmp/never-call-archive.py'],['python3','/tmp/never-call-checkpoint.py',name]]:
   result=subprocess.run(command);print('post-run child exit code:',result.returncode,flush=True)
   if result.returncode:sys.exit(result.returncode)
  if r.returncode:sys.exit(r.returncode)
