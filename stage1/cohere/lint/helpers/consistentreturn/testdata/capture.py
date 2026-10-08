"""Requery every real helper invocation in both consuming Go rule suites.
Only nine independent helpers are instrumented; graph-dependent bodies remain Go.
"""
import json, os, subprocess, tempfile, sys
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
COHERE=ROOT/'cohere'
OUTPUT=Path(sys.argv[1]) if len(sys.argv)>1 else HERE
OUTPUT.mkdir(parents=True,exist_ok=True)
PIN='7945d102a6c18dd36adf9114a758ce646e8b2359'
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=COHERE,text=True).strip()==PIN
assert not subprocess.check_output(['git','status','--porcelain'],cwd=COHERE,text=True).strip()
NAMES=['HasValue','IsGenerator','IsScope','Name','ReportRange','Verb','capitaliseFirst','isExemptFromEndJudgment','staticName']
with tempfile.TemporaryDirectory() as tmp:
 temp=Path(tmp); replace={}
 original=COHERE/'internal/lint/ecmascript/consistentreturn/judgment.go'
 source=original.read_text()
 for name in NAMES:
  assert source.count('func '+name+'(')==1,name
  source=source.replace('func '+name+'(','func adamicRaw'+name+'(')
 side=temp/'judgment.go';side.write_text(source);replace[str(original)]=str(side)
 for file,name in [('capture.go','adamic_capture.go'),('controls.go','adamic_controls_test.go')]:
  replace[str(original.parent/name)]=str(HERE/file)
 harness=COHERE/'internal/lint/testing/rule_testing.go';source=harness.read_text()
 anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 assert source.count(anchor)==1
 side=temp/'testing.go';side.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\nRecordAssertedCase(t,result)\nreturn result'));replace[str(harness)]=str(side)
 program=COHERE/'internal/lint/testing/program.go';source=program.read_text()
 anchor='return Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile:  sourceFile,\n\t\tcapture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}'
 assert source.count(anchor)==1
 side=temp/'program.go';side.write_text(source.replace(anchor,anchor.replace('return Result{','result := Result{')+'\nRecordAssertedCase(t,result)\nreturn result'));replace[str(program)]=str(side)
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 data=temp/'calls.jsonl';env=os.environ|{'ADAMIC_RETURN_CAPTURE':str(data),'COHERE_DOCS_CAPTURE':str(temp/'asserted')}
 for family in ['core','typescript']:
  subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-run','^TestConsistentReturn','-count=1','-timeout=10m'],cwd=COHERE,env=env,stderr=subprocess.STDOUT,check=True)
 live=[json.loads(l) for l in data.read_text().split('\n') if l]
 names=['consistent-return','@typescript-eslint/consistent-return']; asserted={n:set() for n in names}
 for file in (temp/'asserted').glob('*.jsonl'):
  for line in file.read_text().split('\n'):
   if not line:continue
   r=json.loads(line)
   if r['rule'] in asserted:asserted[r['rule']].add(json.dumps([r['file'],r['source'],r.get('options')],sort_keys=True))
 counts={n:len(v) for n,v in asserted.items()};assert all(counts.values()),counts
 subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/ecmascript/consistentreturn','-run','TestAdamicReturnControls','-count=1'],cwd=COHERE,env=env,stderr=subprocess.STDOUT,check=True)
 rows=[json.loads(l) for l in data.read_text().split('\n') if l]
 stats={'upstreamCases':counts,'consumerCalls':{n:sum(r['Kind']==n for r in live) for n in NAMES},'allCalls':{n:sum(r['Kind']==n for r in rows) for n in NAMES}}
 assert all(stats['consumerCalls'].values()),stats
 (OUTPUT/'cases.json').write_text(json.dumps(rows,ensure_ascii=True)+'\n')
 (OUTPUT/'counts.json').write_text(json.dumps(stats,indent=2)+'\n')
 print(json.dumps(stats))
