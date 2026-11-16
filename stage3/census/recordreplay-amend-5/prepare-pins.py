from pathlib import Path
import subprocess as s,tarfile,io
D=Path('/workspace/scratch/b7phwab-amend5');R=D/'work';T=D/'control-tree';P=R/'stage3/adapt/99-wasm-crypto-import/proofs'
def run(*args,**kw):return s.run(list(map(str,args)),check=True,**kw)
def snapshot(ref,name):
 out=D/name;out.mkdir()
 data=s.check_output(['git','-C',str(T),'archive',ref,'src'])
 with tarfile.open(fileobj=io.BytesIO(data)) as f:f.extractall(out,filter='data')
 return out
before=snapshot('refs/stage3/99-wasm-crypto-import-before','before99')
for side in ['before','after']:snapshot('refs/stage3/93-builtin-imports-'+side,'93-'+side)
file=T/'src/compiler/sys.ts';raw=file.read_bytes();needle=b'(process as any).recordreplay';assert raw.count(needle)==1
planned=raw.replace(needle,b'process.recordreplay');(D/'planned-after.ts').write_bytes(planned)
with (D/'scoped-pins.log').open('w') as log:run('node',P/'scoped-hashes.cjs',before/'src/compiler/sys.ts',D/'planned-after.ts',P/'targets.json',stdout=log,stderr=s.STDOUT)
with (D/'adapter-pin.log').open('w') as log:run('node',P.parent/'adapt.cjs','--pin',before,stdout=log,stderr=s.STDOUT)
original=(before/'src/compiler/sys.ts').read_bytes()
assert needle in original,'pin must not apply source edits'
with (D/'composed-adapter.log').open('w') as log:run('node',P.parent/'adapt.cjs',before,stdout=log,stderr=s.STDOUT)
assert (before/'src/compiler/sys.ts').read_bytes()==planned,'composed adapter differs from cast-only plan'
# Return before99 to its genuine snapshot for contracts.
(before/'src/compiler/sys.ts').write_bytes(original)
