"""Capture actual-Go asserted fixtures; dependency seams are not installed here."""
import json,os,subprocess,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[5];COHERE=ROOT/'cohere';SYMBOLS={'github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.Build','github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E].expr','github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E].patternBind'}
consumers={r['rule'] for r in json.loads((HERE.parents[1]/'readiness.json').read_text())['remaining'] if SYMBOLS.intersection(r['remaining_helpers'])}
with tempfile.TemporaryDirectory(prefix='slot04-cfg-capture-') as scratch:
 scratch=Path(scratch);harness=COHERE/'internal/lint/testing/rule_testing.go';anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}';source=harness.read_text();assert source.count(anchor)==1;side=scratch/'harness.go';side.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result'));overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(harness):str(side)}}));env=os.environ|{'COHERE_DOCS_CAPTURE':str(scratch/'capture')}
 for family,selected in [('core','^Test(ArrayCallbackReturn|ConsistentReturn|NoUnreachableLoop)'),('react','^TestRulesOfHooks')]:
  with (HERE.parent/'evidence'/('capture-'+family+'.log')).open('w') as log:subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-count=1','-timeout=10m','-run',selected],cwd=COHERE,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
 unique={}
 for file in sorted((scratch/'capture').glob('*.jsonl')):
  for line in file.read_text().splitlines():
   row=json.loads(line)
   if row['rule'] in consumers:unique[(row['rule'],row['file'],row['source'])]={'Name':row['rule']+':'+row['file'].split('/')[-1],'Source':row['source']}
 rows=[unique[key] for key in sorted(unique)];assert consumers=={r['Name'].split(':')[0] for r in rows};(HERE/'consumers.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n');print('captured',len(rows),'asserted fixtures across',len(consumers),'consumers')
