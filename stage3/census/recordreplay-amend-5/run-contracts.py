from pathlib import Path
import subprocess as s,json,shutil,os
D=Path('/workspace/scratch/b7phwab-amend5');R=D/'work';P=R/'stage3/adapt/99-wasm-crypto-import/proofs';tree=D/'final-composed';before=D/'before99';rows=[]
def run(name,args,expected=0):
 with (D/(name+'.log')).open('w') as f:p=s.run(list(map(str,args)),cwd=R,stdout=f,stderr=s.STDOUT)
 row=dict(name=name,command=list(map(str,args)),exit=p.returncode,expected=expected);rows.append(row);(D/'contracts.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 if p.returncode!=expected: raise RuntimeError(name)
run('scope',['node',P/'check.cjs',before,tree,D/'scope.json'])
run('bindings',['node',P/'bindings.cjs',before,tree,D/'bindings.json'])
run('recordreplay-contract',['node',P/'recordreplay-contract.cjs',before,tree/'src/compiler/sys.ts'])
run('recordreplay-node-resolution-mutant',['node',P/'recordreplay-contract.cjs',before,tree/'src/compiler/sys.ts','no-node-resolution-mutant'],1)
run('require-result-contract',['node',P/'require-result-contract.cjs',tree])
run('resolve-host-contract',['node',P/'resolve-host-contract.cjs'])
run('watch-directory-contract',['node',P/'watch-directory-contract.cjs',tree])
run('watch-directory-witness',['node',P/'watch-directory-witness.cjs',tree])
run('scope-mutant',['node',P/'scope-mutant.cjs',before,tree],1)
run('ternary-mutant',['node',P/'ternary-mutant.cjs',tree],1)
# All restoration mutants must be rejected by the adapter before writing.
f=tree/'src/compiler/sys.ts';saved=f.read_bytes()
for name in ['require-result','resolve-host','watch-directory']:
 try:
  run(name+'-restore-mutant',['node',P.parent/(name+'-mutant.cjs'),tree])
  run(name+'-restore-refused',['node',P.parent/'adapt.cjs',tree],1)
 finally:f.write_bytes(saved)
