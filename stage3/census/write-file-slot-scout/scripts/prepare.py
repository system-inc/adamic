from pathlib import Path
import shutil, json, hashlib
r=Path('/workspace/scratch/write-file-slot-measure')
main=r/'main'
if not (main/'node_modules').exists(): (main/'node_modules').symlink_to('/workspace/scratch/write-file-slot-typescript/node_modules',target_is_directory=True)
p=Path('src/compiler/sys.ts')
s=(main/p).read_text()
assert 'const writeFileView:' in s
for name in ['a-construction','b-slot']:
 d=r/name
 if d.exists(): shutil.rmtree(d)
 shutil.copytree(main/'src',d/'src')
 for f in ['package.json','package-lock.json']:
  if (main/f).exists(): shutil.copy2(main/f,d/f)
 # adapted dependencies are installed by apply; share only immutable dependencies
 if (main/'node_modules').exists(): (d/'node_modules').symlink_to(main/'node_modules',target_is_directory=True)
 if name=='a-construction':
  before='        const nodeSystem: NodeResolutionSystem = {'
  after='''        const originalWriteFile = writeFile;
        const patchedWriteFile: System["writeFile"] = (path, data, writeBom) =>
            writeFileEnsuringDirectories(
                path,
                data,
                !!writeBom,
                (path, data, writeByteOrderMark) => originalWriteFile.call(nodeSystem, path, data, writeByteOrderMark),
                path => nodeSystem.createDirectory(path),
                path => nodeSystem.directoryExists(path),
            );
        const nodeSystem: NodeResolutionSystem = {'''
  assert s.count(before)==1
  t=s.replace(before,after)
  assert t.count('            writeFile,')==1
  t=t.replace('            writeFile,','            writeFile: patchedWriteFile,')
  before='''    if (sys) {
        // patch writefile to create folder before writing the file
        patchWriteFileEnsuringDirectory(sys);
    }
'''
  assert t.count(before)==1
  t=t.replace(before,'')
 else:
  before='    sys.writeFile = (path, data, writeBom) =>'
  assert s.count(before)==1
  t=s.replace(before,'''    const writeFileSlot: { writeFile: System["writeFile"] } = sys;
    writeFileSlot.writeFile = (path, data, writeBom) =>''')
 (d/p).write_bytes(t.replace('\n','\r\n').encode())
print(json.dumps({n:hashlib.sha256((r/n/p).read_bytes()).hexdigest() for n in ['main','a-construction','b-slot']},indent=2))
