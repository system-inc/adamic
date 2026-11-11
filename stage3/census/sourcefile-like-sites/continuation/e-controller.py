import sys,pathlib,shutil,json,hashlib,difflib
sys.dont_write_bytecode=True
from variants import core,checkpoint
from measure import run
from concurrent.futures import ThreadPoolExecutor,as_completed
p=pathlib.Path(__file__).resolve().parent;base=pathlib.Path('/workspace/scratch/baobab-adapted');root=pathlib.Path('/workspace/scratch/baobab-road2-e');target=p/'e';target.mkdir(exist_ok=True)
assert not root.exists();shutil.copytree(base,root)
f=root/'src/compiler/types.ts';old=f.read_bytes();needle=b'/**\r\n * Subset of properties from SourceFile that are used in multiple utility functions\r\n */\r\nexport interface SourceFileLike {'
added=b'/** @internal */\r\nexport interface LineMapSource {\r\n    readonly text: string;\r\n    lineMap?: readonly number[] | undefined;\r\n}\r\n\r\n'
assert old.count(needle)==1;f.write_bytes(old.replace(needle,added+needle,1))
f=root/'src/compiler/scanner.ts';old=f.read_bytes();new=old.replace(b'    type LineAndCharacter,',b'    type LineAndCharacter, type LineMapSource,',1)
needle=b'export function getLineStarts(sourceFile: SourceFileLike)';assert new.count(needle)==1;new=new.replace(needle,needle.replace(b'SourceFileLike',b'LineMapSource'),1)
needle=b'export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter {'
replacement=b'/** @internal */\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;\r\nexport function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter {'
assert new.count(needle)==1;new=new.replace(needle,replacement,1)
new=new.replace(b'export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;',b'export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter; // eslint-disable-line @typescript-eslint/unified-signatures',1)
f.write_bytes(new)
changes={}
for source in sorted((base/'src').rglob('*')):
 if not source.is_file():continue
 rel=source.relative_to(base);a,b=source.read_bytes(),(root/rel).read_bytes()
 if a!=b:
  changes[str(rel)]={'before':hashlib.sha256(a).hexdigest(),'after':hashlib.sha256(b).hexdigest()}
  (target/(source.name+'.patch')).write_text(''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='baseline/'+str(rel),tofile='e/'+str(rel))))
assert set(changes)=={'src/compiler/types.ts','src/compiler/scanner.ts'}
(target/'source-diff.json').write_text(json.dumps({'changed_files':changes,'SourceFileLike_unchanged':True,'SourceFileLike_documentation_preserved':True,'scanner_added_lines':3,'types_added_lines':6,'eslint_suppression_added':True,'function_bodies_unchanged':True},indent=2)+'\n')
core('e',root);checkpoint('road2-e-prepared')
names=['scanner.ts']+json.loads((p/'variant-files.json').read_text())
with ThreadPoolExecutor(max_workers=2) as pool:
 jobs={pool.submit(run,'e',root/'src/compiler',[name],1500,False,2):name for name in names}
 for job in as_completed(jobs):job.result();checkpoint('road2-e-'+jobs[job].replace('/','__'))
print('e census workflow child exit=0',flush=True)
