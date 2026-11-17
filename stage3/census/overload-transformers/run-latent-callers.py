import pathlib,subprocess,sys,os
repo=pathlib.Path(__file__).resolve().parents[3]
scratch=pathlib.Path('/workspace/scratch/overload-transformers-runtime')
for file in ['ts.ts','classFields.ts']:
 for tree in ['main','candidate']:
  adapted=scratch/(tree+'-adapted')
  selected=adapted/'src/compiler/transformers'/file
  name='latent-'+tree+'-'+file.removesuffix('.ts')
  out=pathlib.Path(__file__).parent/(name+'.jsonl')
  cmd=['env','-u','LATENT_SPECULATIVE','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1','CENSUS_MODULE_FORMAT=CommonJSBundle','LATENT_ONLY_FILE='+str(selected),str(scratch/'latent'),str(adapted/'src/compiler'),str(out)]
  p=subprocess.run([sys.executable,str(scratch/'run.py'),name,*cmd],cwd=repo)
  if p.returncode:sys.exit(p.returncode)
