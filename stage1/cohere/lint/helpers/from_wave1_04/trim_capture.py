"""Capture every Run in all four consumer test files, without changing a rule."""
import json, os, re, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
COHERE=ROOT/'cohere'
consumers={"better-tailwindcss/enforce-consistent-class-order","better-tailwindcss/enforce-shorthand-classes","better-tailwindcss/no-conflicting-classes","better-tailwindcss/no-unknown-classes"}
with tempfile.TemporaryDirectory(prefix='wave104-trim-capture-') as temp:
 scratch=Path(temp);replacements={}
 for filename in ['rule_testing.go','program.go']:
  path=COHERE/'internal/lint/testing'/filename;source=path.read_text()
  if filename=='rule_testing.go':
   anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
   replacement='result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result'
  else:
   anchor='return Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile:  sourceFile,\n\t\tcapture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}'
   replacement=anchor.replace('return Result{','result := Result{',1)+'\n RecordAssertedCase(t,result)\n return result'
  assert source.count(anchor)==1,(filename,'capture anchor changed')
  side=scratch/filename;side.write_text(source.replace(anchor,replacement,1));replacements[str(path)]=str(side)
 names=[]
 for slug in ['enforce_consistent_class_order','enforce_shorthand_classes','no_conflicting_classes','no_unknown_classes']:
  path=COHERE/'internal/lint/rules/tailwind'/(slug+'_test.go');source=path.read_text()
  names+=re.findall(r'^func (Test\w+)\(',source,re.M)
  if '/Users/kirkouimet/Projects/ahra/app/_theme/styles' in source:
   side=scratch/path.name;side.write_text(source.replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles','/tmp/wave104-tailwind'));replacements[str(path)]=str(side)
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 records=scratch/'records';env=os.environ|{'COHERE_DOCS_CAPTURE':str(records)}
 (HERE/'evidence').mkdir(exist_ok=True)
 with (HERE/'evidence/trim-capture.log').open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'-count=1','-v','-timeout=10m','./internal/lint/rules/tailwind','-run','^('+'|'.join(names)+')$'],cwd=COHERE,env=env,stdout=log,stderr=subprocess.STDOUT)
 rows={}
 for path in sorted(records.glob('*.jsonl')):
  for line in path.read_text().splitlines():
   raw=json.loads(line);row={key:raw.get(key,raw.get(key.lower()))for key in ['Rule','File','Source']}
   if row['Rule'] not in consumers:continue
   key=(row['Rule'],row['File'],row['Source']);rows[key]={k:row[k]for k in ['Rule','File','Source']}
 captured=[rows[k]for k in sorted(rows)]
 counts={name:sum(row['Rule']==name for row in captured)for name in sorted(consumers)}
 assert all(counts.values()),counts
 (HERE/'trim_fixture_cases.json').write_text(json.dumps(captured,ensure_ascii=True,indent=2)+'\n')
 summary={'upstream_exit':result.returncode,'selected_test_functions':len(names),'captured_sources':counts,'fixture_root_overlay':'/tmp/wave104-tailwind','cohere_pin':subprocess.check_output(['git','rev-parse','HEAD'],cwd=COHERE,text=True).strip()}
 (HERE/'trim_capture.json').write_text(json.dumps(summary,indent=2)+'\n')
 print(json.dumps(summary))
