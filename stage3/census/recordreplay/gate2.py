from pathlib import Path
import os,time,subprocess as s,shutil,json,hashlib
D=Path('/workspace/scratch/recordreplay');R=D/'head';C=D/'control';N=D/'candidate2'
while 'gate lower observations complete' not in (D/'gate-lower.log').read_text():time.sleep(2)
shutil.copytree(C,N,ignore=shutil.ignore_patterns('.git','node_modules','built'))
(N/'node_modules').symlink_to(C/'node_modules',target_is_directory=True)
f=N/'src/compiler/sys.ts';before=f.read_bytes();nl=b'\r\n' if b'\r\n' in before else b'\n'
# Use an existing blank import line, preserving every later source line.
needle=b'import * as _crypto from "node:crypto";'+nl+nl;assert before.count(needle)==1
text=before.replace(b'(process as any).recordreplay',b'process.recordreplay').replace(needle,b'import * as _crypto from "node:crypto";'+nl+b'import type {} from "../recordreplay.js";'+nl)
f.write_bytes(text)
(N/'src/recordreplay.d.ts').write_text('export {};\n\ndeclare global {\n    namespace NodeJS {\n        interface Process {\n            recordreplay?: unknown;\n        }\n    }\n}\n')
for name,argv in [('stock',['node',str(D/'stock.cjs'),str(N),str(D/'candidate2-stock-data.json')]),('dts',['node',str(C/'node_modules/hereby/bin/hereby.js'),'dts']),('types',[str(D/'adamic'),'types',str(N/'src/compiler/sys.ts')])]:
 s.run(['python3',str(D/'run.py'),'candidate2-'+name,*argv],cwd=N)
a=N/'built/local';dest=D/'candidate2-built';dest.mkdir()
for p in a.glob('*.d.ts'):shutil.copyfile(p,dest/p.name)
env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(N/'src/compiler/sys.ts'))
s.run(['python3',str(D/'run.py'),'candidate2-sys-full',str(D/'latent-head'),str(N/'src/compiler/sys.ts'),str(D/'candidate2-sys-full.jsonl')],cwd=R,env=env)
print('gate2 observations complete',flush=True)
