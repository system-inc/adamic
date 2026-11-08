"""Build only guarded entry drivers, each from its own compiler pin."""
import os,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
for side,tree in [('parent','/tmp/lmv-report'),('head','/tmp/lmv-head')]:
 env=dict(os.environ,GOWORK='/tmp/lmv-'+side+'.work')
 phases=[('overlay',['python3','stage3/census/latent/make_overlay.py',tree,'/tmp/lmv-'+side+'-overlay']),('entry-overlay',['python3','stage3/meter/entry_overlay.py','/tmp/lmv-'+side+'-overlay','/tmp/lmv-'+side+'-entry']),('build',['go','build','-p=2','-buildvcs=false','-overlay=/tmp/lmv-'+side+'-entry/overlay.json','-o','/tmp/lmv-'+side+'-census','./stage3/census/latent/tool'])]
 for phase,cmd in phases:
  rc=subprocess.run([sys.executable,str(out/'run.py'),side+'-'+phase,tree,*cmd],env=env).returncode
  if rc:raise SystemExit(rc)
