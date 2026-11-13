#!/usr/bin/env python3
import json,pathlib,subprocess,time,sys
here=pathlib.Path(__file__).resolve().parent
install=pathlib.Path('/workspace/adamic/stage3/typescript/node_modules')
receipts=json.loads((here/'build-exits.json').read_text()) if (here/'build-exits.json').exists() else []
for label in (sys.argv[1:] or ['before','after']):
 tree=pathlib.Path('/workspace/scratch/polling-'+label)
 (tree/'node_modules').symlink_to(install,target_is_directory=True)
 for task in ['local','tests']:
  start=time.monotonic()
  with (here/(label+'-build-'+task+'.log')).open('w') as log:
   try: code=subprocess.run(['node',str(install/'hereby/bin/hereby.js'),task],cwd=tree,stdout=log,stderr=subprocess.STDOUT,timeout=1700).returncode
   except subprocess.TimeoutExpired: code=124
  receipts.append(dict(tree=label,task=task,exit=code,seconds=time.monotonic()-start))
  (here/'build-exits.json').write_text(json.dumps(receipts,indent=2)+'\n')
  print(label,task,code,flush=True)
  if code: sys.exit(code)
 # Shard views supply their own install link.
 (tree/'node_modules').unlink()
if all(pathlib.Path('/workspace/scratch/polling-'+label+'/built/local/typescript.d.ts').exists() for label in ['before','after']):
 assert pathlib.Path('/workspace/scratch/polling-before/built/local/typescript.d.ts').read_bytes()==pathlib.Path('/workspace/scratch/polling-after/built/local/typescript.d.ts').read_bytes()
 (here/'public-declarations-identical.txt').write_text('built/local/typescript.d.ts: byte-identical before and after\n')
