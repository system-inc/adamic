import json,os,pathlib,shutil,subprocess
out=pathlib.Path(__file__).parent;tree=pathlib.Path('/workspace/scratch/tracing2-restoring-mutant');source=pathlib.Path('/workspace/scratch/tracing-after2');control=pathlib.Path('/workspace/scratch/tracing-control2')
tree.mkdir()
for p in source.iterdir():
 if p.name=='src':shutil.copytree(p,tree/p.name)
 elif p.name!='.git':(tree/p.name).symlink_to(p,target_is_directory=p.is_dir())
before=(source/'src/compiler/tracing.ts').read_bytes();restored=(control/'src/compiler/tracing.ts').read_bytes();assert before!=restored
assert b'fs = require("fs");' in restored and b'declare const require: (specifier: "fs") => any;' in restored
(tree/'src/compiler/tracing.ts').write_bytes(restored)
args=['/workspace/scratch/tracing2-first-stop',str(tree/'src/compiler/core.ts'),str(out/'restoring2-mutant-first-stop.jsonl')]
with (out/'restoring2-mutant.log').open('w') as log:child=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
print('restoring mutant child exit='+str(child.returncode),flush=True)
(out/'restoring2-mutant-command.json').write_text(json.dumps(dict(argv=args,exit=child.returncode),indent=2)+'\n')
if child.returncode:raise SystemExit(child.returncode)
rows=[json.loads(s) for s in (out/'restoring2-mutant-first-stop.jsonl').read_text().splitlines()]
assert all(r['kind']=='Refused' and r['where'].endswith('/tracing.ts:381:1') and r['reason']=='an ambient value declaration' for r in rows),rows
assert (source/'src/compiler/tracing.ts').read_bytes()==before
(out/'restoring2-mutant.json').write_text(json.dumps(dict(restoredControlTracingBytes=True,firstStopRestored='tracing.ts:381:1',originalProductUnchanged=True),indent=2)+'\n')
