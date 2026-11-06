"""Export one committed checkpoint's object data for connector publication."""
import base64,json,subprocess,sys
from pathlib import Path
repo=Path('/workspace/adamic')
def git(*args):return subprocess.check_output(['git',*args],cwd=repo)
sha=sys.argv[1] if len(sys.argv)>1 else git('rev-parse','HEAD').decode().strip()
assert len(sha)==40 and all(c in '0123456789abcdef' for c in sha)
parent=git('rev-parse',sha+'^').decode().strip()
files=[]
for value in git('diff-tree','--no-commit-id','--name-only','-r','-z',parent,sha).split(b'\0'):
 if not value:continue
 path=value.decode();entry=git('ls-tree',sha,'--',path).decode().strip()
 if not entry:files.append(dict(path=path,mode='100644',type='blob',sha=None));continue
 mode,kind,obj=entry.split('\t')[0].split()
 assert kind=='blob'
 files.append(dict(path=path,mode=mode,type=kind,sha=obj,content=base64.b64encode(git('cat-file','blob',obj)).decode()))
print(json.dumps(dict(local_sha=sha,parent=parent,tree=git('rev-parse',sha+'^{tree}').decode().strip(),base_tree=git('rev-parse',parent+'^{tree}').decode().strip(),message=git('show','-s','--format=%B',sha).decode().strip(),files=files)))
