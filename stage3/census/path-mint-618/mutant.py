from pathlib import Path
import shutil,difflib
s=Path('/workspace/scratch/path-mint-618');a=s/'lane/adapted-tree';m=s/'mutant'
shutil.copytree(a/'src',m/'src');shutil.copyfile(a/'package.json',m/'package.json')
f=m/'src/compiler/sys.ts';before=f.read_bytes().decode();old='const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames);';new='const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames) as (fileName: string) => Path;'
assert before.count(old)==1
f.write_bytes(before.replace(old,new).encode())
(s/'evidence/restoring-function-cast.patch').write_text(''.join(difflib.unified_diff(before.splitlines(True),f.read_bytes().decode().splitlines(True),fromfile='amended-sys.ts',tofile='restoring-function-cast-sys.ts')))
