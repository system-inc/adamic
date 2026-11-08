"""Build only guarded entry drivers, each from its own compiler pin."""
import os,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
for side,tree in [('parent','/tmp/fa-report'),('head','/tmp/fa-head')]:
 env=dict(os.environ,GOWORK='/tmp/fa-'+side+'.work')
 phases=[('overlay',['python3','stage3/census/latent/make_overlay.py',tree,'/tmp/fa-'+side+'-overlay']),('entry-overlay',['python3','stage3/meter/entry_overlay.py','/tmp/fa-'+side+'-overlay','/tmp/fa-'+side+'-entry']),('build',['go','build','-p=2','-buildvcs=false','-overlay=/tmp/fa-'+side+'-entry/overlay.json','-o','/tmp/fa-'+side+'-census','./stage3/census/latent/tool'])]
 for phase,cmd in phases:
  rc=subprocess.run([sys.executable,str(out/'run.py'),side+'-'+phase,tree,*cmd],env=env).returncode
  if rc:raise SystemExit(rc)
