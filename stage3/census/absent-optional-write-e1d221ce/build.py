"""Build only guarded entry drivers, each from its own compiler pin."""
import os,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
for side,tree in [('parent','/workspace/adamic'),('head','/workspace/aow-head')]:
 env=dict(os.environ,GOWORK='/workspace/aow-'+side+'.work')
 phases=[('overlay',['python3','stage3/census/latent/make_overlay.py',tree,'/workspace/aow-'+side+'-overlay']),('prime-overlay',['python3',str(out/'prime_overlay.py'),'/workspace/aow-'+side+'-overlay']),('entry-overlay',['python3','stage3/meter/entry_overlay.py','/workspace/aow-'+side+'-overlay','/workspace/aow-'+side+'-entry']),('build',['go','build','-p=2','-buildvcs=false','-overlay=/workspace/aow-'+side+'-entry/overlay.json','-o','/workspace/aow-'+side+'-census','./stage3/census/latent/tool'])]
 for phase,cmd in phases:
  rc=subprocess.run([sys.executable,str(out/'run.py'),side+'-'+phase,tree,*cmd],env=env).returncode
  if rc:raise SystemExit(rc)
