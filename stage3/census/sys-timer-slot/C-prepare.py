"""Create C scratch trees from the previously validated canonical main source."""
import pathlib,shutil,json,hashlib,difflib
repo=pathlib.Path('/workspace/adamic');out=repo/'stage3/census/sys-timer-slot';scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');main=scratch/'main'
old=(main/'src/compiler/sys.ts').read_bytes().decode();slot='setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: [...A, ...unknown[]]): unknown;'
changes=[('setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;',slot),('clearTimeout?(timeoutId: any): void;','clearTimeout?(timeoutId: unknown): void;'),('pollScheduled: boolean;','pollScheduled: unknown;'),('let pollScheduled: any;','let pollScheduled: unknown;')]
for before,after in changes:assert old.count(before)==1,before
heads=['declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;','declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;']
basehash=json.loads((out/'source-hashes.json').read_text());metadata={'slot':slot,'baseSourceHashesMatch':True,'variants':{}}
for label,delete in [('C-a',False),('C-b',True)]:
 tree=scratch/label;tree.mkdir()
 for item in main.iterdir():
  if item.name in ['.git','node_modules','built']:continue
  if item.is_dir():shutil.copytree(item,tree/item.name,symlinks=True)
  else:shutil.copy2(item,tree/item.name)
 (tree/'node_modules').symlink_to(main/'node_modules',target_is_directory=True)
 text=old
 for before,after in changes:text=text.replace(before,after)
 if delete:
  for head in heads:assert text.count(head)==1;text=text.replace(head+'\r\n','')
 (tree/'src/compiler/sys.ts').write_bytes(text.encode())
 hashes={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
 different=[k for k in basehash if basehash[k]!=hashes.get(k)];assert different==['src/compiler/sys.ts'];assert set(hashes)==set(basehash)
 (out/(label+'.diff.txt')).write_text(''.join(difflib.unified_diff(old.splitlines(True),text.splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts',n=0)).replace('\r\n','\n'))
 metadata['variants'][label]={'deletedAmbientHeads':delete,'changedSources':different,'sysSha256':hashes['src/compiler/sys.ts'],'lineMappingToMain':'add 2 to source lines >=57' if delete else 'identity'}
(out/'C-source-validation.json').write_text(json.dumps(metadata,indent=2)+'\n');print(json.dumps(metadata),flush=True)
