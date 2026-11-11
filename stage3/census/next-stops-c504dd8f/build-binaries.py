import sys
from pathlib import Path
out=Path(__file__).resolve().parent
sys.path.insert(0,str(out))
from run import run
run('build',['go','build','-o','/workspace/scratch/next-stops-census','./stage3/census/tool'])
run('overlay',[sys.executable,'stage3/census/latent/make_overlay.py','/workspace/adamic','/workspace/scratch/next-stops-overlay'])
files=list(map(str,Path('/workspace/scratch/next-stops-overlay').glob('*.go')))
run('overlay-format',['gofmt','-w',*files])
run('entry-overlay',[sys.executable,'stage3/meter/entry_overlay.py','/workspace/scratch/next-stops-overlay','/workspace/scratch/next-stops-entry-overlay'])
files=list(map(str,Path('/workspace/scratch/next-stops-entry-overlay').glob('*.go')))
run('entry-overlay-format',['gofmt','-w',*files])
run('entry-build',['go','build','-buildvcs=false','-overlay=/workspace/scratch/next-stops-entry-overlay/overlay.json','-o','/workspace/scratch/next-stops-entry','./stage3/census/latent/tool'])
run('latent-build',['go','build','-buildvcs=false','-overlay=/workspace/scratch/next-stops-overlay/overlay.json','-o','/workspace/scratch/next-stops-latent','./stage3/census/latent/tool'])
