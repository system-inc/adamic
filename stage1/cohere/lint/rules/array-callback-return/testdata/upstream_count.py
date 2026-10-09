#!/usr/bin/env python3
import os,json,pathlib,subprocess,tempfile
here=pathlib.Path(__file__).resolve().parent;repo=here.parents[5];cohere=repo/'cohere'
with tempfile.TemporaryDirectory(prefix='cfg-rule-count-') as td:
 temp=pathlib.Path(td);capture=temp/'capture';capture.mkdir()
 original=cohere/'internal/lint/testing/rule_testing.go';source=original.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}';assert source.count(anchor)==1
 side=temp/'rule_testing.go';side.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(side)}}))
 p=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/core','-run','^TestArrayCallbackReturn','-count=1'],cwd=cohere,env=dict(os.environ,COHERE_DOCS_CAPTURE=str(capture)),capture_output=True,text=True)
 assert p.returncode==0,p.stdout+p.stderr
 records=[]
 for file in capture.glob('*.jsonl'):
  records.extend(json.loads(line) for line in file.read_text().splitlines())
 own=[r for r in records if r['rule']=='array-callback-return'];unique={json.dumps({k:r.get(k) for k in ['rule','file','source','options']},sort_keys=True) for r in own}
 counts={'upstream_recordings':len(own),'unique_upstream_cases':len(unique),'cohere_pin':subprocess.check_output(['git','rev-parse','HEAD'],cwd=cohere,text=True).strip()}
 (here/'upstream_count.json').write_text(json.dumps(counts,indent=2)+'\n');print(counts)
