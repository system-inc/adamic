import subprocess,sys
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/census/for-each-ancestor';scratch=Path('/workspace/scratch/ancestor')
for name in (sys.argv[1:] or list('ABCD')):
 root=scratch/name
 code=subprocess.run([sys.executable,str(out/'run.py'),str(out/(name+'-census')),'env','LATENT_FULL=1','LATENT_ONLY_FILE='+str(root/'src/compiler/path.ts'),'LATENT_ASSERT_NO_OUTPUT=1',str(scratch/'latent'),str(root/'src/compiler'),str(out/(name+'-census.jsonl'))]).returncode
 print('wrapper child exit',name,code,flush=True)
 if code:sys.exit(code)
