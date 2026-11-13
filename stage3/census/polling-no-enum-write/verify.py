#!/usr/bin/env python3
import hashlib,json,os,pathlib,subprocess,sys,time
here=pathlib.Path(__file__).resolve().parent
repo=pathlib.Path('/workspace/adamic')
before=pathlib.Path('/workspace/scratch/polling-before')
after=pathlib.Path('/workspace/scratch/polling-after')
receipts={}
def run(name,argv):
 start=time.monotonic()
 with (here/(name+'.log')).open('w') as log:
  try: code=subprocess.run(list(map(str,argv)),stdout=log,stderr=subprocess.STDOUT,timeout=1700).returncode
  except subprocess.TimeoutExpired: code=124
 receipts[name]={'command':list(map(str,argv)),'exit':code,'seconds':time.monotonic()-start}
 (here/'verification-exits.json').write_text(json.dumps(receipts,indent=2)+'\n')
 print(name,code,flush=True)
 return code
assert run('owner-contract',['node',repo/'stage3/adapt/99-correctness-polling-no-enum-write/contract.cjs',after])==0
assert run('replay',['node',repo/'stage3/adapt/99-correctness-polling-no-enum-write/adapt.cjs',after])==0
for label,tree in [('after',after),('before',before)]:
 run(label+'-stock',['node',here/'stock.cjs',tree])
 roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 run(label+'-adamic',['/workspace/scratch/adamic','types',*roots])
for label,tree in [('before',before),('after',after)]:
 run(label+'-custom',['node',here/'node-reference.cjs',tree])
 run(label+'-no-vars',['node',here/'node-reference.cjs',tree,'--no-vars'])
def schedule(name):
 text=(here/(name+'.log')).read_text()
 line=next(line for line in text.splitlines() if line.startswith('scheduled '))
 return text,[v['interval'] for v in json.loads(line[len('scheduled '):])]
btext,bdelays=schedule('before-custom');atext,adelays=schedule('after-custom')
assert bdelays==[17,41,89],bdelays
assert adelays==[250,500,2000],adelays
for value in [250,500,2000]:
 assert f'watchFile {value} TypeError:' in btext
 assert f'watchFile {value} ok' in atext
for value in [17,41,89]: assert f'watchFile {value} ok' in btext
assert (here/'before-no-vars.log').read_bytes()==(here/'after-no-vars.log').read_bytes()
(here/'witness-assertions.json').write_text(json.dumps(dict(before=bdelays,after=adelays,noVariables='byte-identical',defaultCallersBefore='throw',defaultCallersAfter='ok'),indent=2)+'\n')
run('after-enum-census',['node',here/'check-no-enum-write.cjs',after,here/'after-census.json'])
# Restore only the original owner in a disposable copy of the adapted source.
import shutil
mutant=pathlib.Path('/workspace/scratch/polling-restoring-mutant')
shutil.copytree(after/'src',mutant/'src',symlinks=False)
pins=json.loads((repo/'stage3/adapt/99-correctness-polling-no-enum-write/owners.json').read_text())[0]
p=mutant/'src/compiler/sys.ts';text=p.read_text();assert text.count(pins['after'])==1
p.write_text(text.replace(pins['after'],pins['before']))
run('restoring-mutant',['node',here/'check-no-enum-write.cjs',mutant,here/'mutant-census.json'])
files=lambda tree:{str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
b,a=files(before),files(after)
changed=[p for p in sorted(b.keys()|a.keys()) if b.get(p)!=a.get(p)]
assert changed==['src/compiler/sys.ts'],changed
bt=(before/changed[0]).read_text();at=(after/changed[0]).read_text()
assert bt.replace(pins['before'],pins['after'])==at
braw=(before/changed[0]).read_bytes();araw=(after/changed[0]).read_bytes()
ending=b'\r\n' if b'\r\n' in braw else b'\n'
bowner=pins['before'].encode().replace(b'\n',ending);aowner=pins['after'].encode().replace(b'\n',ending)
assert braw.count(bowner)==1
assert braw.replace(bowner,aowner)==araw, 'sys.ts bytes outside the pinned owner changed'
(here/'source-scope.json').write_text(json.dumps(dict(before=b,after=a,changed=changed),indent=2)+'\n')
tracked=subprocess.check_output(['git','-C',str(before),'ls-files','-z']).decode().split('\0')
trackedChanges=[]
for name in filter(None,tracked):
 if (before/name).is_file() and (before/name).read_bytes()!=(after/name).read_bytes(): trackedChanges.append(name)
assert trackedChanges==['src/compiler/sys.ts'],trackedChanges
(here/'tracked-file-scope.json').write_text(json.dumps(dict(files=len(tracked)-1,changed=trackedChanges),indent=2)+'\n')
assert (before/'tests/baselines/reference/api/typescript.d.ts').read_bytes()==(after/'tests/baselines/reference/api/typescript.d.ts').read_bytes()
print('source scope and reference baseline pass',flush=True)
sys.exit(0 if all(v['exit']==(1 if k=='restoring-mutant' else 0) for k,v in receipts.items()) else 1)
