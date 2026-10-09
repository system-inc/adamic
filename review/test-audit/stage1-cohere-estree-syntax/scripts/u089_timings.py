from pathlib import Path
import subprocess,os,json,time
E=Path('review/test-audit/stage1-cohere-estree-syntax');env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/u089/library',ADAMIC_ESTREE_BENCHMARK='1',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4',NODE_V8_COVERAGE='/tmp/u089/v8')
rows=['TestSyntaxGrammar','TestSyntaxRefusals','TestSyntaxLibraries','TestTypeMemberLibraryGap','TestThreePortMutants_Setup','TestThreePortMutants family','TestThreePortMutants_ShardProof','TestThroughput','TestCookedSurrogates','TestCookedSurrogateMutant','TestCookedSurrogateLibraryGap','TestLossyInputRefusal','TestLossyInputControl'];result={}
for row in rows:
 pattern='^TestThreePortMutants_[0-9]{3}$'if row.endswith(' family')else'^'+row+'$';runs=[]
 for i in range(3):
  log=E/f'timing-{row.replace(" ","_")}-{i+1}.log';start=time.monotonic()
  with log.open('w')as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',pattern],env=env,stdout=out,stderr=subprocess.STDOUT)
  data=[]
  for l in log.read_text().splitlines():
   try:data.append(json.loads(l))
   except:pass
  terminal=[d for d in data if d.get('Action') in ('pass','fail')and'Test'not in d and 'Elapsed'in d];cooked='panic: test timed out' in log.read_text()or r.returncode==124
  runs.append(dict(seconds=terminal[-1]['Elapsed']if terminal and not cooked else None,wall=round(time.monotonic()-start,3),exit=r.returncode,cooked=cooked));result[row]=runs;(E/'timings.json').write_text(json.dumps(result,indent=2));print(row,i+1,runs[-1],flush=True)
  if r.returncode and not cooked:print('RED BASELINE: STOP',flush=True);raise SystemExit(2)
