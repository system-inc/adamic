"""Accept earlier measurements only against byte-identical clean-apply src."""
import pathlib,json,hashlib
scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');out=pathlib.Path('/workspace/adamic/stage3/census/sys-timer-slot')
def manifest(label):
 root=scratch/label
 return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
canonical=manifest('canonical');main=manifest('main');assert canonical==main,{'onlyCanonical':sorted(set(canonical)-set(main)),'different':[k for k in canonical if canonical[k]!=main.get(k)]}
r={'canonicalApplyExit':0,'canonicalHasNodeModules':(scratch/'canonical/node_modules').exists(),'files':len(canonical),'mainBytesEqualCanonical':True,'candidates':{}}
assert not r['canonicalHasNodeModules']
a='setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;';b='clearTimeout?(timeoutId: any): void;'
base=(scratch/'canonical/src/compiler/sys.ts').read_bytes()
for label,signature in [('A','setTimeout?(callback: (...args: unknown[]) => void, ms: number, ...args: unknown[]): unknown;'),('B','setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A): unknown;')]:
 m=manifest(label);changed=[k for k in canonical if canonical[k]!=m.get(k)];assert changed==['src/compiler/sys.ts'],changed;assert set(m)==set(canonical)
 expected=base.replace(a.encode(),signature.encode()).replace(b.encode(),b'clearTimeout?(timeoutId: unknown): void;');actual=(scratch/label/'src/compiler/sys.ts').read_bytes();assert expected==actual
 r['candidates'][label]={'changed':changed,'exactTwoDeclarationEdits':True}
(out/'source-validation.json').write_text(json.dumps(r,indent=2)+'\n');(out/'source-hashes.json').write_text(json.dumps(canonical,indent=2)+'\n');print(json.dumps(r))
