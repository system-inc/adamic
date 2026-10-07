import argparse,json,subprocess,os,re,shutil,hashlib,posixpath
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];COHERE=ROOT/'cohere'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--package-root',type=Path);p.add_argument('--replay',action='store_true');a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT,env=None):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,env=env,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
if not a.replay:
 files=['enforce_consistent_class_order','enforce_shorthand_classes','no_conflicting_classes','no_unknown_classes']
 replace={};source=COHERE/'internal/lint/rules/tailwind/collapse/theme.go';text=source.read_text();anchor='func escapeCSSIdentifier(';assert text.count(anchor)==1;patched=S/'escape_parser.go';patched.write_text(text.replace(anchor,'func wave11OriginalEscapeCSSIdentifier(',1));replace[str(source)]=str(patched)
 extra=COHERE/'internal/lint/rules/tailwind/collapse/wave11_capture.go';replace[str(extra)]=str(HERE/'capture_escape.go.txt')
 tests={}
 for slug in files:
  source=COHERE/('internal/lint/rules/tailwind/'+slug+'_test.go');text=source.read_text();tests[slug]=re.findall(r'^func (Test\w+)\(',text,re.M)
  if '/Users/kirkouimet/Projects/ahra/app/_theme/styles' in text:
   patched=S/(slug+'_test.go');patched.write_text(text.replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles',str(a.package_root.resolve())));replace[str(source)]=str(patched)
 replace[str(COHERE/'internal/lint/rules/tailwind/wave11_escape_probe_test.go')]=str(HERE/'consumer_escape_probe.go.txt')
 tests['enforce_shorthand_classes'].append('TestWave11ShorthandCssEscapeProbe')
 overlay=S/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':replace}));capture=S/'captured.jsonl';capture.write_text('');env=os.environ|{'WAVE11_ESCAPE_CAPTURE':str(capture)}
 original=[];coverage={}
 for slug in files:
  capture.write_text('')
  run('consumer-Go-'+slug,['go','test','-overlay='+str(overlay),'./internal/lint/rules/tailwind','-run','^('+'|'.join(tests[slug])+')$','-count=1','-v','-timeout=10m'],COHERE,env)
  records=[json.loads(line) for line in capture.read_text().splitlines()];coverage[slug]=len(records)
  for row in records:row['consumer']=slug
  original+=records
 assert all(coverage.values()),coverage
 capture.write_text(''.join(json.dumps(row,ensure_ascii=True)+'\n' for row in original))
 rows=sorted({row['value'] for row in original})
 actualDistinct=len(rows)
 controls=['','-','--','--0','0','-0','00','é','😀','é😀','😀0','-é0','a😀z','é\x00😀','a.b','a:b','a/b','a b','a\\b','a\nb','--spacing','-9x','9x','__']
 for code in range(128):
  character=chr(code)
  for prefix,suffix in [('', ''),('-', ''),('a',''),('--','z'),('é','😀')]:controls.append(prefix+character+suffix)
 for first in range(128):
  for second in range(128):controls.append(chr(first)+chr(second))
 for depth in [0,1,2,16,128,1024]:controls.append(('x :😀\x00-'+str(depth))*depth)
 rows+=controls
 casefile=S/'cases.json';casefile.write_text(json.dumps(rows,ensure_ascii=True));print('actual helper calls',len(original),'coverage',coverage,'total cases',len(rows),flush=True)
else:
 original=[json.loads(line) for line in (HERE/'escape_capture.jsonl').read_text().splitlines()]
 coverage=json.loads((HERE/'escape_capture_summary.json').read_text())['consumers']
 rows=json.loads((HERE/'escape_cases.json').read_text());casefile=S/'cases.json';casefile.write_text(json.dumps(rows))
 print('replaying',len(original),'actual consumer calls and',len(rows)-len(original),'controls',flush=True)
virtual=COHERE/'wave11_escape_oracle.go';overlay=S/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'oracle_escape.go.txt'),str(COHERE/'internal/lint/rules/tailwind/collapse/wave11_export.go'):str(HERE/'export_escape.go.txt')}}));run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',virtual],COHERE);want=run('Go',[S/'oracle',casefile])
run('build',['go','run',HERE/'build.go',HERE/'main_escape.a',S/'native',S/'emitted.mjs'])
def check(label,source,native,js,mutant=False):
 for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,casefile]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',js,casefile]),('native',[native,casefile])]:
  actual=run(label+'-'+backend,cmd);assert (actual!=want)==mutant,label+' '+backend;print(label,backend,'byte-only catch' if mutant else 'equal',flush=True)
check('baseline',HERE/'main_escape.a',S/'native',S/'emitted.mjs')
change=json.loads((HERE/'mutant_escape.json').read_text());copy=S/'mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True);file=copy/HERE.relative_to(ROOT)/change['file'];text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to'],1));source=copy/HERE.relative_to(ROOT)/'main_escape.a';run('mutant-build',['go','run',HERE/'build.go',source,copy/'native',copy/'emitted.mjs']);check(change['name'],source,copy/'native',copy/'emitted.mjs',True)
(S/'summary.json').write_text(json.dumps({'actualCalls':len(original),'distinctInputs':len({row['value'] for row in original}),'consumers':coverage,'cases':len(rows),'bytes':len(want),'stdoutSha256':hashlib.sha256(want).hexdigest(),'mutant':change['name'],'backends':['Node','emitted','native'],'onlyByteComparison':True},indent=2));print('PASS',len(rows),'cases',len(want),'bytes',flush=True)
