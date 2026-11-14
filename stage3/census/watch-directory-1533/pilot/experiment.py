from pathlib import Path
import shutil,hashlib,json
root=Path(__file__).parent
base=root/'baseline'
candidate=root/'candidate'
candidate.mkdir()
shutil.copytree(base/'src',candidate/'src')
p=candidate/'src/compiler/sys.ts'
b=p.read_bytes()
old=b'function watchDirectory(directoryName: string, callback: DirectoryWatcherCallback, recursive: boolean, options: WatchOptions | undefined): FileWatcher'
new=old.replace(b'recursive: boolean',b'recursive: boolean | undefined')
assert b.count(old)==1
p.write_bytes(b.replace(old,new))
(root/'experiment.json').write_text(json.dumps({'before_sha256':hashlib.sha256(b).hexdigest(),'after_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'before':old.decode(),'after':new.decode()},indent=2)+'\n')
