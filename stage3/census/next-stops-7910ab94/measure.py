import sys,os,json,hashlib,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
sys.path.insert(0,str(out))
from run import run,checkpoint
os.environ['ADAMIC_TYPESCRIPT_SOURCE']='/workspace/scratch/typescript-source'
assert subprocess.check_output(['git','-C',os.environ['ADAMIC_TYPESCRIPT_SOURCE'],'rev-parse','HEAD'],text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
run('apply',['bash','stage3/apply.sh','/workspace/scratch/next-stops-adapted'])
checkpoint('apply')
run('build',['go','build','-o','/workspace/scratch/next-stops-census','./stage3/census/tool'])
run('groups',[sys.executable,str(out/'run.py'),'groups'])
run('analysis',[sys.executable,str(out/'analyze.py')])
run('overlay',[sys.executable,'stage3/census/latent/make_overlay.py','/workspace/adamic','/workspace/scratch/next-stops-overlay'])
run('entry-overlay',[sys.executable,'stage3/meter/entry_overlay.py','/workspace/scratch/next-stops-overlay','/workspace/scratch/next-stops-entry-overlay'])
run('entry-build',['go','build','-overlay=/workspace/scratch/next-stops-entry-overlay/overlay.json','-o','/workspace/scratch/next-stops-entry','./stage3/census/latent/tool'])
run('entry-run',[sys.executable,str(out/'entry-run.py')],limit=1830)
