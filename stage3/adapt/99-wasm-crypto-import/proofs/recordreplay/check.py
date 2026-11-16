"""Re-prove the cast-only amendment; SCRATCH holds control, candidate2 and amend-proof."""
from pathlib import Path
import hashlib,json,os,subprocess as s,sys
D=Path(sys.argv[1]).resolve();R=Path(__file__).resolve().parents[5];P=Path(__file__).resolve().parent
adapter=R/'stage3/adapt/99-wasm-crypto-import/adapt.cjs';tree=D/'amend-proof';expected=D/'candidate2';control=D/'control'
h=lambda b:hashlib.sha256(b).hexdigest()
def inventory(t):return {str(p.relative_to(t)):h(p.read_bytes()) for p in (t/'src').rglob('*') if p.is_file()}
assert inventory(tree)==inventory(expected),'adapter differs from gate source'
a=inventory(control);b=inventory(tree);changed=sorted(k for k in a.keys()|b.keys() if a.get(k)!=b.get(k));assert changed==['src/compiler/sys.ts'],changed
f=tree/'src/compiler/sys.ts';d=tree/'src/recordreplay.d.ts';before=f.read_bytes();stamp=f.stat().st_mtime_ns
assert not d.exists() and b'../recordreplay.js' not in before
s.run(['node',str(adapter),str(tree)],check=True)
assert f.stat().st_mtime_ns==stamp and f.read_bytes()==before,'rerun wrote source'
mutantExits={}
def rejected(name,needle):
 old=f.read_bytes();mt=f.stat().st_mtime_ns
 p=s.run(['node',str(adapter),str(tree)],capture_output=True,text=True)
 mutantExits[name]=p.returncode;print(json.dumps({'mutant':name,'exit':p.returncode}),flush=True)
 assert p.returncode and needle in p.stderr,p.stderr
 assert f.read_bytes()==old and f.stat().st_mtime_ns==mt
try:
 d.write_text('export {};\n');rejected('declaration','adaptation declaration must be absent')
finally:d.unlink()
prefix=s.check_output(['git','-C',str(control),'show','refs/stage3/99-wasm-crypto-import-before:src/compiler/sys.ts'])
try:
 f.write_bytes(prefix)
 script="const a=require(process.argv[1]),r=require(process.argv[2]);const old=r.edits;r.edits=(...args)=>[...old(...args),{start:args[0].text.length,end:args[0].text.length,text:'\\nexport const extra = true;\\n'}];a.adapt(process.argv[3]);"
 mt=f.stat().st_mtime_ns;p=s.run(['node','-e',script,str(adapter),str(adapter.parent/'recordreplay.cjs'),str(tree)],capture_output=True,text=True)
 mutantExits['out_of_scope']=p.returncode;print(json.dumps({'mutant':'out_of_scope','exit':p.returncode}),flush=True)
 assert p.returncode and 'edit outside getNodeSystem and pinned import edits' in p.stderr,p.stderr
 assert f.read_bytes()==prefix and f.stat().st_mtime_ns==mt and not d.exists()
 f.write_bytes(prefix.replace(b'(process as any).recordreplay',b'(process as any).recordreplaY'));rejected('owner','uninventoried owner')
finally:f.write_bytes(before)
assert inventory(tree)==inventory(expected)
result={'status':'pass','mutant_exit_codes':mutantExits,'source_changes':changed,'exact_adapter_matches_gate':True,'idempotent_no_write':True,'adaptationDeclarationAbsent':True}
Path(os.environ.get('RECORDREPLAY_RESULTS',P/'results.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
