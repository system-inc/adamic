#!/usr/bin/env python3
"""Node-held semantic witnesses; compiler outcomes are evidence, not waived errors."""
import json,os,re,subprocess,tempfile
from pathlib import Path
P=Path(__file__).resolve().parent
ROOT=P.parents[3]
EXPECTED={'nullable-argument':'ok\n','nullable-result-relation':'ok\n','nullable-dereference':'ok\n','nullable-iteration':'ok1\n','optional-field-presence':'present\n','caught-unknown':'ok\n','host-declarations':'host\n','host-members':'7\n','host-inference':'ok\n','library-set-shape':'1\n','nullable-index-key':'ok\n','optional-call':'ok\n','overload-use':'ok\n','spread-tuple':'ok1\n'}
mutants=json.loads((P/'mutants.json').read_text()); results={}
logs=P/'evidence'/'fixtures';logs.mkdir(exist_ok=True)
def run(cmd):
 r=subprocess.run(cmd,text=True,capture_output=True)
 return {'command':cmd,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr}
with tempfile.TemporaryDirectory(prefix='checker-stops-mutants-') as tmp:
 for group,expect in EXPECTED.items():
  src=P/'fixtures'/f'{group}.a'; node=run(['node','--disable-warning=ExperimentalWarning',str(ROOT/'oracle/node.mjs'),str(src)])
  assert (node['exit'],node['stdout'],node['stderr'])==(0,expect,''),(group,node)
  mutation=mutants[group];text=src.read_text();assert text.count(mutation['replace'])==1
  mutant=Path(tmp)/f'{group}.a';mutant.write_text(text.replace(mutation['replace'],mutation['with']))
  bad=run(['node','--disable-warning=ExperimentalWarning',str(ROOT/'oracle/node.mjs'),str(mutant)])
  assert (bad['exit'],bad['stdout'],bad['stderr'])!=(0,expect,''),(group,'uncaught mutant')
  compilers={}
  for name,var in [('main','MAIN_ADAMIC'),('topic','TOPIC_ADAMIC')]:
   compiler=os.environ.get(var)
   if compiler:
    outcome=run([compiler,'c',str(src)]); compilers[name]=outcome
    (logs/f'{group}.{name}.log').write_text(outcome['stderr']+outcome['stdout'])
  if 'main' in compilers:
   actual=compilers['main']; match=re.search(r'error TS(\d+):',actual['stderr'])
   header=f'// a-check: type error TS{match[1]}' if match else '// Node behavior witness; see fixtures.json for compiler outcome.'
   if os.environ.get('RECORD_HEADERS')=='1':
    src.write_text(header+'\n'+'\n'.join(src.read_text().splitlines()[1:])+'\n')
   else: assert src.read_text().splitlines()[0]==header,(group,header)
  results[group]={'expected':expect,'node':node,'mutant':bad,'catcher':'Node golden stdout/exit/stderr comparison','compilers':compilers}
(P/'fixtures.json').write_text(json.dumps(results,indent=2)+'\n')
(P/'counts.md').write_text('# Local fixture counts\n\n14 Node witnesses; 14 semantic mutants caught. No oracle-discovery fixtures added.\nNative counted execution is not claimed; compiler results are in fixtures.json.\n')
print('14 Node witnesses passed; 14 semantic mutants caught')
