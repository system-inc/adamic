from pathlib import Path
import os,sys,shutil,json,hashlib
sys.path.insert(0,'/workspace/scratch/sys-walls');from run import run,S,ROOT
before=S/'before';after=S/'candidate'
def checked(name,args,cwd=ROOT,env=None):
 code=run(name,args,cwd,env)
 if code:raise SystemExit(code)
if not after.exists():checked('candidate-copy',['cp','-a',before,after])
p=after/'src/compiler/sys.ts';raw=p.read_bytes();text=raw.decode();decls=['declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;','declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;']
for d in decls:
 assert text.count(d)==1
 text=text.replace(d+'\r\n','') if d+'\r\n' in text else text.replace(d+'\n','')
p.write_bytes(text.encode())
(S/'deletion.json').write_text(json.dumps({'before':hashlib.sha256(raw).hexdigest(),'after':hashlib.sha256(p.read_bytes()).hexdigest(),'declarations':decls},indent=2)+'\n')
for side in ['before','candidate']:
 tree=S/side;roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 checked('types-'+side,[S/'bin/adamic','types',*roots])
 checked('timer-types-'+side,[S/'bin/type-probe',tree])
 (tree/'node_modules').symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
 checked('build-'+side,['npm','run','build:compiler'],tree)
files=lambda t:sorted(p.relative_to(t) for p in t.rglob('*') if p.is_file())
afiles=files(before/'built/local');bfiles=files(after/'built/local');assert afiles==bfiles
names=[p for p in afiles if str(p).endswith('.js')];assert len(names)==9,names
names += [Path(n) for n in ['typescript.d.ts','tsserverlibrary.d.ts','typescript.internal.d.ts']]
rows=[]
for n in names:
 a=(before/'built/local'/n).read_bytes();b=(after/'built/local'/n).read_bytes();rows.append(dict(file=str(n),byteIdentical=a==b,before=hashlib.sha256(a).hexdigest(),after=hashlib.sha256(b).hexdigest(),beforeBytes=len(a),afterBytes=len(b)))
(S/'emission-comparison.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(r['byteIdentical'] for r in rows if r['file']!='typescript.internal.d.ts'),rows
print('TIMER MEASUREMENT PASSED',flush=True)
