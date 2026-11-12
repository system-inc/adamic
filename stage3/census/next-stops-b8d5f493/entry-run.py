import json,os,sys,time
from pathlib import Path
out=Path('/workspace/adamic/stage3/census/next-stops-b8d5f493');sys.path.insert(0,str(out));from run import run
p=out/'apply-status.json'
while not p.exists():time.sleep(1)
assert json.loads(p.read_text())['exit_code']==0
os.environ.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'entry-times.jsonl'))
run('entry-census',['/workspace/scratch/next-stops-entry','/workspace/scratch/next-stops-adapted/src/tsc/tsc.ts',str(out/'entry-latent.jsonl')],limit=1780)
