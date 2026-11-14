"""Check the amendment against its gate source and official pre-99 input."""
from pathlib import Path
import hashlib,json,os,subprocess as s,sys
D=Path(sys.argv[1]).resolve();R=Path(__file__).resolve().parents[5];P=Path(__file__).resolve().parent
adapter=R/'stage3/adapt/99-wasm-crypto-import/adapt.cjs';tree=D/'amend-proof';expected=D/'candidate2';control=D/'control'
h=lambda b:hashlib.sha256(b).hexdigest()
def inventory(t):return {str(p.relative_to(t)):h(p.read_bytes()) for p in (t/'src').rglob('*') if p.is_file()}
assert inventory(tree)==inventory(expected), 'adapter differs from gate source'
a=inventory(control);b=inventory(tree);changed=sorted(k for k in a.keys()|b.keys() if a.get(k)!=b.get(k));assert changed==['src/compiler/sys.ts','src/recordreplay.d.ts'],changed
f=tree/'src/compiler/sys.ts';d=tree/'src/recordreplay.d.ts';before=f.read_bytes();decl=d.read_bytes();stamp=(f.stat().st_mtime_ns,d.stat().st_mtime_ns)
s.run(['node',str(adapter),str(tree)],check=True)
assert (f.stat().st_mtime_ns,d.stat().st_mtime_ns)==stamp, 'rerun wrote source'
assert f.read_bytes()==before and d.read_bytes()==decl
# A foreign ambient type must not be silently accepted on a rerun.
d.write_bytes(decl.replace(b'unknown',b'number'));mutant=d.read_bytes();mt=d.stat().st_mtime_ns
p=s.run(['node',str(adapter),str(tree)],capture_output=True,text=True)
assert p.returncode and 'recordreplay declaration changed' in p.stderr,p.stderr
assert f.read_bytes()==before and d.read_bytes()==mutant and d.stat().st_mtime_ns==mt
d.write_bytes(decl)
# Recreate the official before-owner input and plant an out-of-owner edit.
prefix=s.check_output(['git','-C',str(control),'show','refs/stage3/99-wasm-crypto-import-before:src/compiler/sys.ts'])
f.write_bytes(prefix);d.unlink();mt=f.stat().st_mtime_ns
script="const a=require(process.argv[1]),r=require(process.argv[2]);const old=r.edits;r.edits=(...args)=>[...old(...args),{start:args[0].text.length,end:args[0].text.length,text:'\\nexport const extra = true;\\n'}];a.adapt(process.argv[3]);"
p=s.run(['node','-e',script,str(adapter),str(adapter.parent/'recordreplay.cjs'),str(tree)],capture_output=True,text=True)
assert p.returncode and 'edit outside getNodeSystem and pinned import edits' in p.stderr,p.stderr
assert f.read_bytes()==prefix and f.stat().st_mtime_ns==mt and not d.exists()
# An owner mutation must fail before either file is written.
f.write_bytes(prefix.replace(b'(process as any).recordreplay',b'(process as any).recordreplaY'));mutant=f.read_bytes();mt=f.stat().st_mtime_ns
p=s.run(['node',str(adapter),str(tree)],capture_output=True,text=True)
assert p.returncode and 'uninventoried owner' in p.stderr,p.stderr
assert f.read_bytes()==mutant and f.stat().st_mtime_ns==mt and not d.exists()
f.write_bytes(before);d.write_bytes(decl)
assert inventory(tree)==inventory(expected)
stock=json.loads((D/'control-stock-data.json').read_text());after=json.loads((D/'candidate2-stock-data.json').read_text());assert not stock['diagnostics'] and not after['diagnostics'];assert stock['javascript']==after['javascript'];assert stock['roots']==after['roots']==79
public={}
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 x=(D/'control-built'/name).read_bytes();y=(D/'candidate2-built'/name).read_bytes();assert x==y;public[name]={'bytes':len(y),'sha256':h(y)}
result={'status':'pass','source_changes':changed,'exact_adapter_matches_gate':True,'idempotent_no_write':True,'declaration_mutant_refused_before_write':True,'out_of_scope_edit_mutant_refused_before_write':True,'owner_mutant_refused_before_write':True,'stock_version':stock['version'],'stock_diagnostics':0,'compiler_roots':79,'javascript_files_identical':len(stock['javascript']),'public_declarations':public,'sys_sha256':h(before),'augmentation_sha256':h(decl)}
(P/'results.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
