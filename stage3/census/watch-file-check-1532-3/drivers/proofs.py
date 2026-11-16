import os, pathlib, subprocess as p, json, time, tarfile
s=pathlib.Path('/workspace/scratch/watch1532-recut');r=s/'work';tree=s/'composed';out=s/'proofs';out.mkdir(exist_ok=True)
env=dict(os.environ,NODE_PATH=str(r/'stage3/api/node_modules'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'gotmp'),GOCACHE=str(s/'go-cache'))
rows=[]
def run(name,args,expected=0):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as f: z=p.run(list(map(str,args)),cwd=r,env=env,stdout=f,stderr=p.STDOUT)
 row=dict(name=name,command=list(map(str,args)),exit=z.returncode,expected=expected,seconds=round(time.monotonic()-start,3));rows.append(row);(out/'results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 return z.returncode

def snapshot(d,state):
 dest=out/(d+'-'+state);dest.mkdir(exist_ok=True)
 a=p.Popen(['git','-C',str(tree),'archive','refs/stage3/'+d+'-'+state],stdout=p.PIPE);p.run(['tar','-x','-C',str(dest)],stdin=a.stdout,check=True);a.stdout.close();assert a.wait()==0
 (dest/'.git').write_text('gitdir: '+str(tree/'.git')+'\n');return dest
names=['93-builtin-imports','99-correctness-core-multimap','99-correctness-single-watcher-forwarder','99-correctness-timer-closures','99-wasm-crypto-import']
snaps={d:{k:snapshot(d,k) for k in ['before','after']} for d in names}
for d,v in snaps.items():
 a=r/'stage3/adapt'/d
 run(d+'-first-apply',['node',a/'adapt.cjs',v['before']])
 # Full byte and nanosecond mtime equality on replay, with direct write counting.
 replay="""const fs=require('fs'),path=require('path'),assert=require('assert/strict'),a=require(process.argv[1]),tree=process.argv[2];const files=fs.readdirSync(path.join(tree,'src'),{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.join(e.parentPath,e.name));const snap=()=>files.map(f=>[fs.readFileSync(f),String(fs.statSync(f,{bigint:true}).mtimeNs)]);const before=snap(),write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args)};const result=a.adapt(tree);assert.equal(writes,0);assert.deepEqual(snap(),before);console.log(JSON.stringify({result,writes,bytesAndMtimeEqual:true}));"""
 run(d+'-replay',['node','-e',replay,a/'adapt.cjs',v['before']])
 # Restore predecessor files from the captured ref after the replay proof.
 dest=v['before'];a1=p.Popen(['git','-C',str(tree),'archive','refs/stage3/'+d+'-before'],stdout=p.PIPE);p.run(['tar','-x','-C',str(dest)],stdin=a1.stdout,check=True);a1.stdout.close();assert a1.wait()==0
v=snaps['93-builtin-imports'];a=r/'stage3/adapt/93-builtin-imports/proofs'
run('93-legacy-check',['node',a/'check.cjs',v['before'],v['after'],out/'93-checks.json'],1)
run('93-bindings-and-mutants',['node',a/'bindings.cjs',v['before'],v['after'],out/'93-bindings.json'])
saved=(a/'owner-contract-results.json').read_bytes()
try:run('93-owner-contract-and-mutants',['node',a/'owner-contract.cjs',v['before'],v['after'],out/'93-contract'])
finally:(a/'owner-contract-results.json').write_bytes(saved)
run('93-scope-mutant',['node',a/'scope-mutant.cjs',v['before'],v['after']],1)
v=snaps['99-correctness-core-multimap'];a=r/'stage3/adapt/99-correctness-core-multimap'
run('core-contract-and-mutants',['node',a/'contract.cjs',v['before'],out/'core-contract',out/'core-contract.json'])
with tarfile.open(a/'preparation.tar.gz') as t:
 data=t.extractfile('preparation/bodies.json').read();(out/'core-bodies.json').write_bytes(data)
run('core-body-proof',['node',a/'body-proof.cjs',v['after'],out/'core-bodies.json',out/'core-body-proof.json'])
run('core-runtime-witness',['node',a/'witness.cjs',out/'core-bodies.json',out/'core-witness.json'])
run('core-restoring-mutant',['node',a/'mutant.cjs',v['after']])
run('core-restoring-mutant-rejected',['node',a/'adapt.cjs',v['after']],1)
v=snaps['99-correctness-single-watcher-forwarder'];a=r/'stage3/adapt/99-correctness-single-watcher-forwarder'
run('single-watcher-contract-and-mutants',['node',a/'contract.cjs',v['before']])
v=snaps['99-correctness-timer-closures'];a=r/'stage3/adapt/99-correctness-timer-closures'
run('timer-contract-and-mutants',['node',a/'contract.cjs',v['before']])
run('timer-restored-slot-mutant',['node',a/'contract.cjs',v['before'],'--restore-slot'],1)
run('timer-witness-before',['node',a/'witness.cjs',v['before']])
run('timer-witness-after',['node',a/'witness.cjs',v['after']])
run('timer-request-mutant',['node',a/'witness.cjs',v['after'],'request'],1)
run('timer-label-mutant',['node',a/'witness.cjs',v['after'],'label'],1)
v=snaps['99-wasm-crypto-import'];a=r/'stage3/adapt/99-wasm-crypto-import/proofs'
run('wasm-source-transpile',['node',a/'check.cjs',v['before'],v['after'],out/'wasm-checks.json'])
run('wasm-bindings-and-mutants',['node',a/'bindings.cjs',v['before'],v['after'],out/'wasm-bindings.json'])
run('wasm-scope-mutant',['node',a/'scope-mutant.cjs',v['before'],v['after']],1)
run('wasm-ternary-mutant',['node',a/'ternary-mutant.cjs',v['after']],1)
run('wasm-require-result-contract',['node',a/'require-result-contract.cjs',v['after'],tree])
run('wasm-resolve-host-contract',['node',a/'resolve-host-contract.cjs'])
run('wasm-watch-directory-contract',['node',a/'watch-directory-contract.cjs',v['after']])
run('wasm-watch-directory-witness',['node',a/'watch-directory-witness.cjs',v['after']])
print('FAILURES',json.dumps([x for x in rows if x['exit']!=x['expected']]),flush=True)
