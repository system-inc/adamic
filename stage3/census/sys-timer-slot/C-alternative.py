"""Keep requested C intact; measure the full-tuple NoInfer alternative separately."""
import pathlib,shutil,json,hashlib,difflib
scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');out=pathlib.Path('/workspace/adamic/stage3/census/sys-timer-slot');base=(scratch/'main/src/compiler/sys.ts').read_bytes().decode();rows={}
for kind in ['a','b']:
 source=scratch/('C-'+kind);target=scratch/('C-noinfer-'+kind);target.mkdir()
 for p in source.iterdir():
  if p.name in ['built','node_modules','.git']:continue
  if p.is_dir():shutil.copytree(p,target/p.name,symlinks=True)
  else:shutil.copy2(p,target/p.name)
 (target/'node_modules').symlink_to(scratch/'main/node_modules',target_is_directory=True)
 old=(source/'src/compiler/sys.ts').read_bytes().decode();before='setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]): unknown;';after='setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;';assert old.count(before)==1;text=old.replace(before,after);(target/'src/compiler/sys.ts').write_bytes(text.encode())
 (out/(target.name+'.diff.txt')).write_text(''.join(difflib.unified_diff(base.splitlines(True),text.splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts',n=0)).replace('\r\n','\n'))
 rows[target.name]={'sysSha256':hashlib.sha256(text.encode()).hexdigest(),'onlyAdditionalChangeFromRequestedC':after}
(out/'C-alternative-source.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows),flush=True)
