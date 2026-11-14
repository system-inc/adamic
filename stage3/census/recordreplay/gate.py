from pathlib import Path
import os,time,subprocess as s,shutil,json,hashlib
D=Path('/workspace/scratch/recordreplay'); R=D/'head'; C=D/'control'; N=D/'candidate'
while not (C/'patch-set.md').exists():
 try: os.kill(899,0)
 except ProcessLookupError: raise RuntimeError('official apply failed before patch table')
 time.sleep(2)
# Wait until the complete official apply exits, including its product archive.
while True:
 try: os.kill(899,0)
 except ProcessLookupError: break
 time.sleep(2)
assert not N.exists()
shutil.copytree(C,N,ignore=shutil.ignore_patterns('.git','node_modules','built'))
(N/'node_modules').symlink_to(C/'node_modules',target_is_directory=True)
f=N/'src/compiler/sys.ts'; before=f.read_bytes(); assert before.count(b'(process as any).recordreplay')==1
nl=b'\r\n' if b'\r\n' in before else b'\n'
augmentation=nl.join([b'',b'declare global {',b'    namespace NodeJS {',b'        interface Process {',b'            recordreplay?: unknown;',b'        }',b'    }',b'}',b''])
f.write_bytes(before.replace(b'(process as any).recordreplay',b'process.recordreplay')+augmentation)
(D/'candidate.sys.before').write_bytes(before);(D/'candidate.sys.after').write_bytes(f.read_bytes())
for label,tree in [('control',C),('candidate',N)]:
 for name,argv in [('stock',['node',str(D/'stock.cjs'),str(tree),str(D/(label+'-stock-data.json'))]),('dts',['node',str(C/'node_modules/hereby/bin/hereby.js'),'dts'])]:
  r=s.run(['python3',str(D/'run.py'),label+'-'+name,*argv],cwd=tree)
  if r.returncode: print(label,name,'exit',r.returncode,flush=True)
 a=tree/'built/local';dest=D/(label+'-built');dest.mkdir()
 for p in a.glob('*.d.ts'):shutil.copyfile(p,dest/p.name)
print('gate stock and declaration build complete',flush=True)
