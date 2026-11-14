from pathlib import Path
import shutil
out=Path(__file__).resolve().parent
old=b'const changedFilesInLastPoll: WatchedFileWithUnchangedPolls[] = [];'
new=b'const changedFilesInLastPoll: (WatchedFileWithUnchangedPolls | undefined)[] = [];'
a=(out/'after-tree/src/compiler/sys.ts').read_bytes();assert a.count(new)==1
restored=a.replace(new,old);assert restored==(out/'main-tree/src/compiler/sys.ts').read_bytes()
(out/'sys-restoring-mutant.ts.txt').write_bytes(restored)
shutil.copytree(out/'after-tree/src',out/'restoring-mutant-tree/src')
(out/'restoring-mutant-tree/src/compiler/sys.ts').write_bytes(restored)
print('Restoring mutant exactly matches main sys.ts; fixture: sys-restoring-mutant.ts.txt')
