from pathlib import Path
import shutil,subprocess
scratch=Path('/workspace/lint-cost-drivers');nodes=scratch/'stage1/typescript/parser/nodes.ts';prototype=nodes.read_bytes()
try:
 for name in ['baseline','prototype','all-listeners']:
  nodes.write_bytes(Path('/workspace/lint-cost-nodes-original.txt').read_bytes() if name=='baseline' else prototype)
  dst=Path('/workspace/lint-cost-'+name)
  with (dst/'main.c').open('wb') as out,(dst/'generate.stderr').open('wb') as err:
   subprocess.run(['/workspace/lint-cost-adamic','c',str(dst/'main.ts')],cwd=scratch,stdout=out,stderr=err,check=True)
  for src in (scratch/'internal/native/runtime').iterdir():
   if src.suffix in ['.c','.h']:shutil.copy2(src,dst/src.name)
  with (dst/'harness-build.log').open('wb') as out:subprocess.run(['python3','/workspace/lint-cost-build.py',name],stdout=out,stderr=out,check=True)
  print(name,'rebuilt',flush=True)
finally:nodes.write_bytes(prototype)
