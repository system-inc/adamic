#!/usr/bin/env python3
import hashlib,json,os,subprocess,sys
from pathlib import Path
sdk='/workspace/adamic-tools/go/bin/go'
root=Path.cwd()
while root.parent!=root and not (root/'go.mod').exists():root=root.parent
module=(root/'go.mod').read_text() if (root/'go.mod').exists() else ''
if module.startswith('module github.com/system-inc/adamic\n'):
 template=Path('/workspace/stage3-scoreboard-tmp/shared.work').read_text()
 work=template.replace('/workspace/stage3-latent-scoreboard',str(root))
 folder=Path('/workspace/stage3-meter-progress-tmp/cached-workspaces');folder.mkdir(exist_ok=True)
 file=folder/(hashlib.sha256(str(root).encode()).hexdigest()[:16]+'.work')
 if not file.exists():
  file.write_text(work)
  (folder/(file.stem+'.json')).write_text(json.dumps(dict(compiler_root=str(root),compiler_sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),dependency_root='/workspace/stage3-diagnosis-main/cohere',workspace=str(file)),indent=2)+'\n')
 os.environ['GOWORK']=str(file)
os.execvpe(sdk,[sdk,*sys.argv[1:]],os.environ)
