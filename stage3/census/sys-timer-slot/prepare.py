"""Freeze the applied tree and create two source-only candidate scratch copies."""
import pathlib, shutil, hashlib, difflib, json
repo=pathlib.Path('/workspace/adamic'); scratch=pathlib.Path('/workspace/scratch/sys-timer-slot'); out=repo/'stage3/census/sys-timer-slot'; main=scratch/'main'
a='setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;'
b='clearTimeout?(timeoutId: any): void;'
source=(main/'src/compiler/sys.ts').read_bytes().decode()
assert source.count(a)==source.count(b)==1
candidates={'A':'setTimeout?(callback: (...args: unknown[]) => void, ms: number, ...args: unknown[]): unknown;', 'B':'setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A): unknown;'}
manifest={str(f.relative_to(main)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((main/'src').rglob('*')) if f.is_file()}
(out/'source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
for label,decl in candidates.items():
    dest=scratch/label;dest.mkdir()
    for item in main.iterdir():
        if item.name in ['.git','node_modules','built']:continue
        if item.is_dir():shutil.copytree(item,dest/item.name,symlinks=True)
        else:shutil.copy2(item,dest/item.name)
    (dest/'node_modules').symlink_to(main/'node_modules',target_is_directory=True)
    text=source.replace(a,decl).replace(b,'clearTimeout?(timeoutId: unknown): void;')
    (dest/'src/compiler/sys.ts').write_bytes(text.encode())
    (out/(label+'.diff.txt')).write_text(''.join(difflib.unified_diff(source.splitlines(True),text.splitlines(True),fromfile='a/src/compiler/sys.ts',tofile='b/src/compiler/sys.ts')).replace('\r\n','\n'))
print(json.dumps({'files':len(manifest),'candidates':list(candidates)}),flush=True)
