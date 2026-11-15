from pathlib import Path
import shutil
r=Path('/workspace/scratch/write-file-slot-run')
main=r/'main'
for label in ['a','b']:
 t=r/label
 shutil.copytree(main,t,dirs_exist_ok=True,ignore=shutil.ignore_patterns('.git','node_modules'))
 f=t/'src/compiler/sys.ts'
 s=f.read_text()
 if label=='a':
  needle='        const nodeSystem: NodeResolutionSystem = {'
  assert s.count(needle)==1
  s=s.replace(needle,'''        const originalWriteFile = writeFile;
        const wrappedWriteFile: System["writeFile"] = (path, data, writeBom) =>
            writeFileEnsuringDirectories(
                path,
                data,
                !!writeBom,
                (path, data, writeByteOrderMark) => originalWriteFile.call(nodeSystem, path, data, writeByteOrderMark),
                path => nodeSystem.createDirectory(path),
                path => nodeSystem.directoryExists(path),
            );
'''+needle)
  assert s.count('            writeFile,')==1
  s=s.replace('            writeFile,','            writeFile: wrappedWriteFile,')
  needle='''    if (sys) {
        // patch writefile to create folder before writing the file
        patchWriteFileEnsuringDirectory(sys);
    }
'''
  assert s.count(needle)==1
  s=s.replace(needle,'')
 else:
  needle='    sys.writeFile = (path, data, writeBom) =>'
  assert s.count(needle)==1
  s=s.replace(needle,'''    const writeFileSlot: { writeFile: System["writeFile"] } = sys;
    writeFileSlot.writeFile = (path, data, writeBom) =>''')
 f.write_text(s)
print('private candidates prepared')
t=r/'hidden'
shutil.copytree(main,t,dirs_exist_ok=True,ignore=shutil.ignore_patterns('.git','node_modules'))
f=t/'src/compiler/sys.ts';s=f.read_text();needle='    sys.writeFile = (path, data, writeBom) =>';assert s.count(needle)==1
s=s.replace(needle,'    const unusedWriteFile: System["writeFile"] = (path, data, writeBom) =>');f.write_text(s)
print('hidden-stop probe prepared (store removed; not a candidate)')
