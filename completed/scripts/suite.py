from pathlib import Path
import subprocess as s,os,json,time,tarfile,io,shutil
r=Path(__file__).resolve().parent;repo=Path('/workspace/scratch/watch-file-1532-merge');tree=r/'after';runner=r/'runproof.py'
while not (r/'after-apply.exit').exists():time.sleep(1)
assert (r/'after-apply.exit').read_text().strip()=='0'
snaps=r/'snapshots';snaps.mkdir();failures=[]
def snapshot(name,state):
 d=snaps/(name+'-'+state);d.mkdir()
 data=s.check_output(['git','-C',str(tree),'archive','refs/stage3/'+name+'-'+state,'src'])
 with tarfile.open(fileobj=io.BytesIO(data)) as t:t.extractall(d,filter='data')
 return d
names=['93-builtin-imports','99-correctness-core-multimap','99-correctness-single-watcher-forwarder','99-correctness-timer-closures','99-wasm-crypto-import']
pairs={n:(snapshot(n,'before'),snapshot(n,'after')) for n in names}
(r/'snapshots.json').write_text(json.dumps({n:list(map(str,p)) for n,p in pairs.items()},indent=2)+'\n')
def run(name,cmd,expected=0):
 code=s.run(['python3',str(runner),name,str(expected),*map(str,cmd)],cwd=repo).returncode
 if code:failures.append(name)
 return code
node=lambda p:['node',repo/'stage3/adapt'/p]
# Single-step contracts and their embedded red cases.
b,a=pairs['99-correctness-core-multimap'];run('multimap-contract',node('99-correctness-core-multimap/contract.cjs')+[b,r/'multimap-contract-scratch',r/'multimap-contract.json'])
b,a=pairs['99-correctness-single-watcher-forwarder'];run('single-watcher-contract',node('99-correctness-single-watcher-forwarder/contract.cjs')+[b])
b,a=pairs['99-correctness-timer-closures'];run('timer-contract',node('99-correctness-timer-closures/contract.cjs')+[b]);run('timer-restored-slot-mutant',node('99-correctness-timer-closures/contract.cjs')+[b,'--restore-slot'],1)
run('timer-witness',node('99-correctness-timer-closures/witness.cjs')+[tree]);run('timer-label-mutant',node('99-correctness-timer-closures/witness.cjs')+[tree,'label'],1);run('timer-request-mutant',node('99-correctness-timer-closures/witness.cjs')+[tree,'request'],1)
b,a=pairs['93-builtin-imports'];report=repo/'stage3/adapt/93-builtin-imports/proofs/owner-contract-results.json';saved=report.read_bytes()
try:run('builtin-owner-contract',node('93-builtin-imports/proofs/owner-contract.cjs')+[b,a,r/'builtin-owner-contract']);shutil.copy2(report,r/'builtin-owner-contract.json')
finally:report.write_bytes(saved)
b,a=pairs['99-wasm-crypto-import'];run('crypto-check',node('99-wasm-crypto-import/proofs/check.cjs')+[b,a,r/'crypto-check.json']);run('crypto-bindings',node('99-wasm-crypto-import/proofs/bindings.cjs')+[b,a,r/'crypto-bindings.json'])
run('crypto-scope-mutant',node('99-wasm-crypto-import/proofs/scope-mutant.cjs')+[b,a],1);run('crypto-ternary-mutant',node('99-wasm-crypto-import/proofs/ternary-mutant.cjs')+[a],1)
run('crypto-require-result-contract',node('99-wasm-crypto-import/proofs/require-result-contract.cjs')+[tree]);run('crypto-watch-directory-contract',node('99-wasm-crypto-import/proofs/watch-directory-contract.cjs')+[tree]);run('crypto-watch-directory-witness',node('99-wasm-crypto-import/proofs/watch-directory-witness.cjs')+[tree])
# Mutant generators exit zero; the unmodified validator must then exit one.
m=r/'mutant-tree';shutil.copytree(tree/'src',m/'src');sysfile=m/'src/compiler/sys.ts';core=m/'src/compiler/core.ts';sysraw=sysfile.read_bytes();coreraw=core.read_bytes()
for label,script,mode in [('require-result','require-result-mutant.cjs',None),('resolve-host','resolve-host-mutant.cjs',None),('watch-directory-restore','watch-directory-mutant.cjs','restore'),('watch-directory-unnormalized','watch-directory-mutant.cjs','unnormalized'),('watch-directory-optional-contract','watch-directory-mutant.cjs','optional-contract')]:
 try:
  run('crypto-'+label+'-mutate',node('99-wasm-crypto-import/'+script)+[m]+([mode] if mode else []))
  run('crypto-'+label+'-mutant-red',node('99-wasm-crypto-import/adapt.cjs')+[m],1)
 finally:
  shutil.rmtree(m/'src');shutil.copytree(tree/'src',m/'src')
run('multimap-mutant-generate',node('99-correctness-core-multimap/mutant.cjs')+[m]);run('multimap-mutant-red',node('99-correctness-core-multimap/adapt.cjs')+[m],1)
# The directory's own instrumenter extracts old/free bodies, then its witness compares them.
mm=r/'multimap-instrument';b,a=pairs['99-correctness-core-multimap'];shutil.copytree(b/'src',mm/'src')
run('multimap-instrument',node('99-correctness-core-multimap/instrument.cjs')+[mm,r/'multimap-bodies.json']);run('multimap-witness',node('99-correctness-core-multimap/witness.cjs')+[r/'multimap-bodies.json',r/'multimap-witness.json']);run('multimap-body-proof',node('99-correctness-core-multimap/body-proof.cjs')+[tree,r/'multimap-bodies.json',r/'multimap-body-proof.json'])
# 95h's existing four independent red controls at its immediate checkpoint.
p=snapshot('95-checker-axolotl-h','after');run('multimap-95h-contract',node('99-correctness-core-multimap/contract95.cjs')+[p,r/'multimap-95h-contract.json'])
(r/'suite-result.json').write_text(json.dumps({'failures':failures,'allExpected':not failures},indent=2)+'\n');print('SUITE FAILURES',failures,flush=True)
