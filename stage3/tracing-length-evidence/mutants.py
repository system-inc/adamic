from pathlib import Path
from run import run,raw,root
import os,json,shutil
source=Path('/workspace/scratch/tracing-length-after');pins=json.loads((root/'stage3/adapt/99-correctness-tracing-length/pins.json').read_text())
for p in pins:
 side='mutant-'+p['owner'];tree=raw/side
 shutil.copytree(source/'src',tree/'src',symlinks=True)
 file=tree/'src/compiler/tracing.ts';text=file.read_bytes().decode();assert text.count(p['after'])==1
 file.write_bytes(text.replace(p['after'],p['before']).encode())
 env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(file),LATENT_ASSERT_NO_OUTPUT='1')
 run(side+'-census',[raw/'latent',tree/'src/compiler',raw/(side+'-census.jsonl')],env=env)
