import pathlib,subprocess,os,sys
repo=pathlib.Path('/workspace/never-call-report');out=repo/'stage3/census/never-call-392cb60b'
for side in ['parent','head']:
 name='tsc-entry-first-'+side;env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 env.update(LATENT_ENTRY_REACH='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/(name+'.units.jsonl')),LATENT_PREPASS_TIMES=str(out/(name+'.prepasses.log')))
 cmd=['python3','/tmp/never-call-run.py',name,str(repo if side=='parent' else pathlib.Path('/workspace/never-call-head')),str(out),'/workspace/never-call-tools/'+side+'/census-entry','/workspace/never-call-adapted/src/tsc/tsc.ts',str(out/(name+'.jsonl'))]
 r=subprocess.run(cmd,env=env);print(name,'wrapper child exit code:',r.returncode,flush=True)
 for command in [['python3','/tmp/never-call-archive.py'],['python3','/tmp/never-call-checkpoint.py',name]]:
  p=subprocess.run(command);print('post-run child exit code:',p.returncode,flush=True)
  if p.returncode:sys.exit(p.returncode)
 if r.returncode:sys.exit(r.returncode)
