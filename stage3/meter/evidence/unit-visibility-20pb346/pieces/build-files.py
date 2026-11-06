"""Rebuild both file-selection binaries against one pinned compiler base.

Run on integration base d4c6b07 with the implementation commit applied. Baseline
sources are exported under the repo cache so Go resolves the repo module.
"""
from pathlib import Path
import subprocess, tarfile, io
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/20pb346-2')
base='d4c6b07aca1a7c55785226563c2ad79d552c1adf'
before=repo/'.cache/20pb346-pieces-main'
before.mkdir(parents=True,exist_ok=True)
archive=subprocess.check_output(['git','archive',base,'stage3/census/latent'],cwd=repo)
with tarfile.open(fileobj=io.BytesIO(archive)) as files:files.extractall(before,filter='data')
for side,script in (('before',before/'stage3/census/latent/make_overlay.py'),('after',repo/'stage3/census/latent/make_overlay.py')):
 overlay=root/('pieces-'+side+'-overlay')
 subprocess.run(['python3',str(script),str(repo),str(overlay)],cwd=repo,check=True)
 subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(root/('pieces-'+side)),'./stage3/census/latent/tool'],cwd=repo,check=True)
