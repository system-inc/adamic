from pathlib import Path
from run import run,raw,root
import os,json,shutil,sys
source=Path('/workspace/scratch/tracing-length-after');pins=json.loads((root/'stage3/adapt/99-correctness-tracing-length/pins.json').read_text())
for p in pins:
 if len(sys.argv)>1 and p['owner'] not in sys.argv[1:]:continue
 side='mutant-'+p['owner'];tree=raw/side
 shutil.copytree(source/'src',tree/'src',symlinks=True)
 file=tree/'src/compiler/tracing.ts';text=file.read_bytes().decode();assert text.count(p['after'])==1
 old,new={'pop':('eventStack.splice(eventStack.length - 1);','eventStack.length--;'),'popAll':('eventStack.splice(0);','eventStack.length = 0;')}[p['owner']]
 assert p['after'].count(old)==1
 restored=p['after'].replace(old,new)
 file.write_bytes(text.replace(p['after'],restored).encode())
 (raw/(side+'-source-edit.json')).write_text(json.dumps({'owner':p['owner'],'onlyLiteralRestored':{'from':old,'to':new},'explicitRangeErrorGuardRetained':True},indent=2)+'\n')
 env=dict(os.environ,LATENT_FULL='1',LATENT_ONLY_FILE=str(file),LATENT_ASSERT_NO_OUTPUT='1')
 run(side+'-census',[raw/'latent',tree/'src/compiler',raw/(side+'-census.jsonl')],env=env)
