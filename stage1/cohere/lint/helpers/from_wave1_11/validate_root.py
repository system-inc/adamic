import argparse,json,subprocess,os,re,shutil,hashlib,posixpath
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];COHERE=ROOT/'cohere'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--package-root',type=Path);p.add_argument('--replay',action='store_true');a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT,env=None):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
if not a.replay:
 consumers=[('array-callback-return','core','array_callback_return','TestArrayCallbackReturn'),('consistent-return','core','consistent_return','TestConsistentReturn'),('no-unreachable-loop','core','no_unreachable_loop','TestNoUnreachableLoop'),('react-hooks/rules-of-hooks','react','rules_of_hooks','TestRulesOfHooks')]
 source=COHERE/'internal/lint/ecmascript/control_flow_graph/roots.go';text=source.read_text();anchor='func IsRoot(';assert text.count(anchor)==1;patched=S/'roots.go';patched.write_text(text.replace(anchor,'func wave11OriginalIsRoot(',1))
 extra=source.parent/'wave11_capture.go';overlay=S/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(patched),str(extra):str(HERE/'capture_root.go.txt')}}));capture=S/'captured.jsonl';env=os.environ|{'WAVE11_ROOT_CAPTURE':str(capture)}
 original=[];coverage={}
 for consumer,pkg,slug,prefix in consumers:
  capture.write_text('');run('consumer-Go-'+slug,['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+pkg,'-run','^'+prefix,'-count=1','-v','-timeout=10m'],COHERE,env)
  records=[json.loads(line) for line in capture.read_text().splitlines()];coverage[consumer]=len(records)
  for row in records:row['consumer']=consumer
  original+=records
 assert all(coverage.values()),coverage
 capture.write_text(''.join(json.dumps(row)+'\n' for row in original))
 # Pure predicate: deduplicate metadata, preserving every distinct actual consumer input.
 unique={(r['kind'],r['initializer']) for r in original}
 rows=[{'kind':kind,'initializer':init} for kind,init in sorted(unique)]
 for kind in range(501):
  for init in [False,True]:rows.append({'kind':kind,'initializer':init})
 casefile=S/'cases.json';casefile.write_text(json.dumps(rows));print('actual helper calls',len(original),'distinct',len(unique),'coverage',coverage,'total cases',len(rows),flush=True)
else:
 original=[json.loads(line) for line in (HERE/'root_capture.jsonl').read_text().splitlines()]
 coverage=json.loads((HERE/'root_capture_summary.json').read_text())['consumers']
 rows=json.loads((HERE/'root_cases.json').read_text());casefile=S/'cases.json';casefile.write_text(json.dumps(rows))
 print('replaying',len(original),'actual consumer calls through',len(rows),'metadata cases',flush=True)
virtual=COHERE/'wave11_root_oracle.go';overlay=S/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'oracle_root.go.txt')}}));run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],COHERE);want=run('Go',[S/'oracle',casefile])
run('build',['go','run',HERE/'build.go',HERE/'main_root.a',S/'native',S/'emitted.mjs'])
def check(label,source,native,js,mutant=False):
 for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,casefile]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,casefile]),('native',[native,casefile])]:
  actual=run(label+'-'+backend,cmd);assert (actual!=want)==mutant,label+' '+backend;print(label,backend,'byte-only catch' if mutant else 'equal',flush=True)
check('baseline',HERE/'main_root.a',S/'native',S/'emitted.mjs')
change=json.loads((HERE/'mutant_root.json').read_text());copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);file=copy/HERE.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1));source=copy/HERE.relative_to(ROOT)/'main_root.a';run('mutant-build',['go','run',HERE/'build.go',source,copy/'native',copy/'emitted.mjs']);check(change['name'],source,copy/'native',copy/'emitted.mjs',True)
(S/'summary.json').write_text(json.dumps({'actualCalls':len(original),'consumers':coverage,'cases':len(rows),'bytes':len(want),'stdoutSha256':hashlib.sha256(want).hexdigest(),'mutant':change['name'],'backends':['Node','emitted','native'],'onlyByteComparison':True},indent=2));print('PASS',len(rows),'cases',len(want),'bytes',flush=True)
