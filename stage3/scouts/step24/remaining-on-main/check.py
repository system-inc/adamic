#!/usr/bin/env python3
"""Each fixture test has a 30-second deadline; oracle executes original source."""
import json, pathlib, subprocess, tempfile, time
HERE=pathlib.Path(__file__).resolve().parent
REPO=HERE.parents[3]
CASES=[('optional-field-presence','present\n','{ read }','{}','TS2375'),('nullable-result-relation','ok\n','values[0]','values[1]','TS2322'),('nullable-argument','ok\n','values[0]','values[1]','TS2345')]
def run(args):
 start=time.monotonic(); p=subprocess.run(args,capture_output=True,text=True,timeout=30)
 return {'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr,'seconds':round(time.monotonic()-start,3)}
results=[]
for name,golden,old,new,code in CASES:
 start=time.monotonic(); source=HERE/(name+'.a')
 baseline=run(['node','--disable-warning=ExperimentalWarning',str(REPO/'oracle/node.mjs'),str(source)])
 assert baseline['exit']==0 and baseline['stdout']==golden,(name,baseline)
 with tempfile.TemporaryDirectory() as tmp:
  mutant=pathlib.Path(tmp)/'mutant.a'; text=source.read_text(); assert text.count(old)==1
  mutant.write_text(text.replace(old,new)); changed=run(['node','--disable-warning=ExperimentalWarning',str(REPO/'oracle/node.mjs'),str(mutant)])
  assert changed['exit']!=0 or changed['stdout']!=golden,(name,changed)
  native=run(['/tmp/step24-remaining-main-adamic','build',str(source),'-o',tmp+'/native'])
  assert native['exit']!=0 and ('error '+code+':') in native['stderr']+native['stdout'],(name,native)
 assert time.monotonic()-start<30
 results.append({'group':name,'oracle':baseline,'mutant':changed,'in_place_compiler':native})
(HERE/'fixtures.json').write_text(json.dumps(results,indent=2)+'\n')
print('PASS: three Node goldens, three semantic mutants caught, three in-place headers checked')
