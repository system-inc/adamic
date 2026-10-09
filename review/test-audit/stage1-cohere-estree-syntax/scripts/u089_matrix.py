from pathlib import Path
import subprocess,os,json,time
E=Path('review/test-audit/stage1-cohere-estree-syntax');rows=['TestSyntaxGrammar','TestSyntaxRefusals','TestThroughput','TestCookedSurrogates','TestLossyInputRefusal'];env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/u089/library',ADAMIC_ESTREE_BENCHMARK='1',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4');result={}
for mid in ['clean','M1','M2','M3','M4','P1']:
 Path('/tmp/u089/mutant').write_text(''if mid=='clean'else mid);cells={}
 for row in (rows[:1] if mid=='clean' else rows):
  log=E/f'{mid}-{row}.log';t=time.monotonic()
  with log.open('w')as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^'+row+'$'],env=env,stdout=out,stderr=subprocess.STDOUT)
  data=[]
  for l in log.read_text().splitlines():
   try:data.append(json.loads(l))
   except:pass
  terminal=[d for d in data if d.get('Test')==row and d.get('Action')in('pass','fail','skip')];cooked='panic: test timed out'in log.read_text()or r.returncode==124
  cells[row]=dict(result=terminal[-1]['Action']if terminal and not cooked else 'unknown',seconds=round(time.monotonic()-t,3),exit=r.returncode,cooked=cooked);result[mid]=cells;(E/'matrix.json').write_text(json.dumps(result,indent=2));print(mid,row,cells[row],flush=True)
  if mid=='clean'and r.returncode and not cooked:raise SystemExit('RED INACTIVE BASELINE: STOP')
