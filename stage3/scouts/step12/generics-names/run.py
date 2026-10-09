#!/usr/bin/env python3
"""Source Node oracle, sequential normal/split native builds, source mutants."""
import argparse, json, os, pathlib, subprocess
p=argparse.ArgumentParser(); p.add_argument('--generic',required=True); p.add_argument('--names',required=True); p.add_argument('--out',required=True); a=p.parse_args()
here=pathlib.Path(__file__).resolve().parent; repo=here.parents[3]; out=pathlib.Path(a.out).resolve(); out.mkdir(parents=True,exist_ok=False)
assert pathlib.Path(a.generic).is_file() and pathlib.Path(a.names).is_file(), 'build both scratch compilers first'
manifest=json.loads((here/'manifest.json').read_text()); results=[]
def run(command,stem,env=None):
 with (out/(stem+'.stdout')).open('wb') as stdout, (out/(stem+'.stderr')).open('wb') as stderr:
  process=subprocess.run(command,stdout=stdout,stderr=stderr,env=env,cwd=repo)
 return {'exit':process.returncode,'stdout':(out/(stem+'.stdout')).read_text(),'stderr':(out/(stem+'.stderr')).read_text()}
for w in manifest['witnesses']:
 source=here/w['file']; stem=source.stem; compiler=getattr(a,w['topic']); row={'file':w['file'],'topic':w['topic']}
 row['node']=run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(source)],stem+'.node')
 assert row['node']=={'exit':0,'stdout':w['expected_stdout'],'stderr':''},row
 mutant=out/(stem+'.mutant.a'); text=source.read_text(); old,new=w['mutant']; assert text.count(old)==1
 mutant.write_text(text.replace(old,new,1)); row['mutant_node']=run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(mutant)],stem+'.mutant.node')
 assert row['mutant_node']['exit']==0 and row['mutant_node']['stdout']!=row['node']['stdout'],row
 row['modes']=[]
 for split in ['0','1']:
  env=dict(os.environ,ADAMIC_NATIVE_SPLIT=split,ADAMIC_NATIVE_JOBS='5'); label=stem+'.split'+split; binary=out/label
  mode={'split':split,'build':run([compiler,'build',str(source),'-o',str(binary),'--sanitize'],label+'.build',env)}
  if w['expected_build']:
   assert mode['build']['exit']!=0 and w['expected_build'] in mode['build']['stderr'],mode
  else:
   assert mode['build']['exit']==0,mode
  if mode['build']['exit']==0:
   mode['native']=run([str(binary)],label+'.native',dict(env,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'))
   mode['matches_node']=mode['native']==row['node']; assert mode['matches_node'],mode
   mbinary=out/(label+'.mutant'); mode['mutant_build']=run([compiler,'build',str(mutant),'-o',str(mbinary),'--sanitize'],label+'.mutant.build',env)
   if mode['mutant_build']['exit']==0:
    mode['mutant_native']=run([str(mbinary)],label+'.mutant.native',dict(env,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'))
    mode['mutant_caught']=mode['mutant_native']['exit']==0 and mode['mutant_native']['stderr']=='' and mode['mutant_native']['stdout']!=row['node']['stdout']; assert mode['mutant_caught'],mode
   else: raise AssertionError('mutant must execute, not fail compilation: ' + str(mode))
  row['modes'].append(mode)
 results.append(row)
(out/'report.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps({'witnesses':len(results),'node_pass':len(results),'node_mutants_caught':len(results),'native_matches':sum(m.get('matches_node',False) for r in results for m in r['modes']),'native_mutants_caught':sum(m.get('mutant_caught',False) for r in results for m in r['modes'])}))
