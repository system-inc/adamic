import sys,pathlib,shutil,json,hashlib,difflib
sys.dont_write_bytecode=True
from variants import invoke,core,checkpoint
from measure import run
from concurrent.futures import ThreadPoolExecutor
follow=pathlib.Path(__file__).resolve().parent
base=pathlib.Path('/workspace/scratch/baobab-adapted')
root=pathlib.Path('/workspace/scratch/baobab-road2-c')
target=follow/'c';target.mkdir(parents=True,exist_ok=True)
assert not root.exists()
shutil.copytree(base,root)
p=root/'src/compiler/types.ts';old=p.read_bytes()
needle=b'export interface SourceFileLike {\r\n    readonly text: string;\r\n    /** @internal */\r\n    lineMap?: readonly number[] | undefined;'
replacement=b'/** @internal */\r\nexport interface LineMapSource {\r\n    readonly text: string;\r\n    lineMap?: readonly number[] | undefined;\r\n}\r\n\r\nexport interface SourceFileLike extends LineMapSource {'
assert old.count(needle)==1;p.write_bytes(old.replace(needle,replacement))
p=root/'src/compiler/scanner.ts';old=p.read_bytes();new=old
# Keep scanner line coordinates stable by adding the type import on the existing line.
new=new.replace(b'    type LineAndCharacter,',b'    type LineAndCharacter, type LineMapSource,',1)
for name in ['getLineStarts','getLineAndCharacterOfPosition']:
 needle=('export function '+name+'(sourceFile: SourceFileLike').encode();assert new.count(needle)==1;new=new.replace(needle,needle.replace(b'SourceFileLike',b'LineMapSource'),1)
p.write_bytes(new)
changes={}
for source in sorted((base/'src').rglob('*')):
 if not source.is_file():continue
 rel=source.relative_to(base);a=source.read_bytes();b=(root/rel).read_bytes()
 if a!=b:
  changes[str(rel)]={'before':hashlib.sha256(a).hexdigest(),'after':hashlib.sha256(b).hexdigest()}
  (target/(source.name+'.patch')).write_text(''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='baseline/'+str(rel),tofile='c/'+str(rel))))
assert set(changes)=={'src/compiler/types.ts','src/compiler/scanner.ts'}
(target/'source-diff.json').write_text(json.dumps({'changed_files':changes,'text_and_lineMap_identical':True,'removed_member_internal_comment':True,'scanner_coordinates_preserved':True},indent=2)+'\n')
core('c',root)
checkpoint('road2-c-prepared')
names=['scanner.ts']+json.loads((follow/'variant-files.json').read_text())
def measure(name):
 run('c',root/'src/compiler',[name],1500,False,concurrency=2)
with ThreadPoolExecutor(max_workers=2) as pool:
 for name,_ in zip(names,pool.map(measure,names)):
  checkpoint('road2-c-'+name.replace('/','__'))
print('c workflow child exit=0',flush=True)
