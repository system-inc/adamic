import argparse,json,subprocess,os,re,shutil,hashlib,posixpath
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];COHERE=ROOT/'cohere'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--package-root',type=Path);p.add_argument('--replay',action='store_true');a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT,env=None):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
if not a.replay:
 files=['enforce_canonical_classes','enforce_consistent_class_order','enforce_consistent_variant_order','enforce_shorthand_classes','no_conflicting_classes','no_unknown_classes']
 replace={};source=COHERE/'internal/lint/rules/tailwind/collapse/value_parser.go';text=source.read_text();anchor='func ValueToCss(';assert text.count(anchor)==1;patched=S/'value_parser.go';patched.write_text(text.replace(anchor,'func wave11OriginalValueToCss(',1));replace[str(source)]=str(patched)
 extra=COHERE/'internal/lint/rules/tailwind/collapse/wave11_capture.go';replace[str(extra)]=str(HERE/'capture_value.go.txt')
 tests={}
 for slug in files:
  source=COHERE/('internal/lint/rules/tailwind/'+slug+'_test.go');text=source.read_text();tests[slug]=re.findall(r'^func (Test\w+)\(',text,re.M)
  if '/Users/kirkouimet/Projects/ahra/app/_theme/styles' in text:
   patched=S/(slug+'_test.go');patched.write_text(text.replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles',str(a.package_root.resolve())));replace[str(source)]=str(patched)
 replace[str(COHERE/'internal/lint/rules/tailwind/wave11_value_probe_test.go')]=str(HERE/'consumer_value_probe.go.txt')
 tests['enforce_canonical_classes'].append('TestWave11CanonicalValueProbe');tests['enforce_shorthand_classes'].append('TestWave11ShorthandValueProbe')
 overlay=S/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':replace}));capture=S/'captured.jsonl';capture.write_text('');env=os.environ|{'WAVE11_VALUE_CAPTURE':str(capture)}
 original=[];coverage={}
 for slug in files:
  capture.write_text('')
  run('consumer-Go-'+slug,['go','test','-overlay='+str(overlay),'./internal/lint/rules/tailwind','-run','^('+'|'.join(tests[slug])+')$','-count=1','-v','-timeout=10m'],COHERE,env)
  records=[json.loads(line) for line in capture.read_text().splitlines()];coverage[slug]=len(records)
  for row in records:row['consumer']=slug
  original+=records
 assert all(coverage.values()),coverage
 capture.write_text(''.join(json.dumps(row,ensure_ascii=True)+'\n' for row in original))
 rows=[r['nodes'] for r in original]
 def node(kind,value,children=None):return {'kind':kind,'value':value,'nodes':children or []}
 rows += [[],None,[node('unknown','ignored',[node('word','hidden')])]]
 values=['','a','a b','\t\n\r','_','\\','é😀','x)(','\x00','calc','url']
 for kind in ['word','separator','function','unknown']:
  for value in values:
   rows.append([node(kind,value)])
   rows.append([node(kind,value,[node('word','left'),node('separator',',  '),node('function','nested',[node('word',value)])])])
 for depth in range(1,65):
  tree=[node('word','leaf')]
  for _ in range(depth):tree=[node('function','f',tree)]
  rows.append(tree)
 casefile=S/'cases.json';casefile.write_text(json.dumps(rows,ensure_ascii=True));print('actual helper calls',len(original),'coverage',coverage,'total cases',len(rows),flush=True)
else:
 original=[json.loads(line) for line in (HERE/'value_capture.jsonl').read_text().splitlines()]
 coverage=json.loads((HERE/'value_capture_summary.json').read_text())['consumers']
 rows=json.loads((HERE/'value_cases.json').read_text());casefile=S/'cases.json';casefile.write_text(json.dumps(rows))
 print('replaying',len(original),'actual consumer calls and',len(rows)-len(original),'controls',flush=True)
virtual=COHERE/'wave11_value_oracle.go';overlay=S/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'oracle_value.go.txt')}}));run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],COHERE);want=run('Go',[S/'oracle',casefile])
run('build',['go','run',HERE/'build.go',HERE/'main_value.a',S/'native',S/'emitted.mjs'])
def check(label,source,native,js,mutant=False):
 for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,casefile]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,casefile]),('native',[native,casefile])]:
  actual=run(label+'-'+backend,cmd);assert (actual!=want)==mutant,label+' '+backend;print(label,backend,'byte-only catch' if mutant else 'equal',flush=True)
check('baseline',HERE/'main_value.a',S/'native',S/'emitted.mjs')
change=json.loads((HERE/'mutant_value.json').read_text());copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);file=copy/HERE.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1));source=copy/HERE.relative_to(ROOT)/'main_value.a';run('mutant-build',['go','run',HERE/'build.go',source,copy/'native',copy/'emitted.mjs']);check(change['name'],source,copy/'native',copy/'emitted.mjs',True)
(S/'summary.json').write_text(json.dumps({'actualCalls':len(original),'consumers':coverage,'cases':len(rows),'bytes':len(want),'stdoutSha256':hashlib.sha256(want).hexdigest(),'mutant':change['name'],'backends':['Node','emitted','native'],'onlyByteComparison':True},indent=2));print('PASS',len(rows),'cases',len(want),'bytes',flush=True)
