"""Capture real program inputs and original fixture filenames for all consumers."""
import json, os, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
COHERE=ROOT/'cohere'
consumers={r['rule'] for r in json.loads((HERE.parent/'readiness.json').read_text())['remaining'] if any(h.endswith('tailwind.projectRootOf') for h in r['remaining_helpers'])}
with tempfile.TemporaryDirectory(prefix='wave104-project-root-') as scratch:
 scratch=Path(scratch);capture=scratch/'records';capture.mkdir()
 replacements={}
 path=COHERE/'internal/lint/testing/program.go';source=path.read_text()
 source=source.replace('import (','import (\n "encoding/json"',1)
 anchor='options := optionsFor(directory)'
 assert source.count(anchor)==1
 hook='''if captureDir:=os.Getenv("ADAMIC_WAVE104_CAPTURE");captureDir!="" {
 options:=ruleContext.Program.Options();config:="";if options!=nil{config=options.ConfigFilePath}
 output,err:=os.CreateTemp(captureDir,"program-*.json");if err!=nil{t.Fatal(err)}
 err=json.NewEncoder(output).Encode(map[string]any{"Rule":subject.Name,"File":subjectFileName,"Source":files[subjectFileName],"Config":config,"Cwd":ruleContext.Program.GetCurrentDirectory(),"Present":options!=nil});if err!=nil{t.Fatal(err)};if err=output.Close();err!=nil{t.Fatal(err)}
 }
 '''
 side=scratch/'typed.go';side.write_text(source.replace(anchor,hook+anchor,1));replacements[str(path)]=str(side)
 path=COHERE/'internal/lint/testing/rule_testing.go';source=path.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}';assert source.count(anchor)==1
 side=scratch/'simple.go';side.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result',1));replacements[str(path)]=str(side)
 for filename in ['enforce_consistent_class_order_test.go','no_unknown_classes_test.go']:
  path=COHERE/'internal/lint/rules/tailwind'/filename
  side=scratch/filename;side.write_text(path.read_text().replace('/Users/kirkouimet/Projects/ahra/app/_theme/styles','/tmp/wave104-tailwind'))
  replacements[str(path)]=str(side)
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 env=os.environ|{'ADAMIC_WAVE104_CAPTURE':str(capture),'COHERE_DOCS_CAPTURE':str(scratch/'docs')}
 pattern='^Test(EnforceCanonicalClasses|EnforceConsistentClassOrder|EnforceConsistentVariantOrder|EnforceShorthandClasses|NoConflictingClasses|NoUnknownClasses|CanonicalMessage|CanonicalProposes|CanonicalCollapse|CanonicalLogical|ClassOrderOptions|ClassOrderMessage|ClassOrderFix|ClassOrderReports|ClassOrderDimensions|ClassOrderVariantDimensions|ClassOrderDepth|ClassOrderReadings|ClassOrderRoots|ClassOrderMarkers|ClassOrderSortsTemplate|ClassOrderLeavesTemplate|ConflictingClassesPropose|UnknownIgnore|KnownRootWithUnknown)'
 with (HERE/'evidence/capture.log').open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'-count=1','-v','-timeout=10m','./internal/lint/rules/tailwind','-run',pattern],cwd=COHERE,env=env,stdout=log,stderr=subprocess.STDOUT)
 rows=[]
 for file in sorted(capture.glob('*.json')):
  row=json.loads(file.read_text())
  if row['Rule'] in consumers:row['Origin']='real-program';rows.append(row)
 actual={row['Rule']for row in rows}
 # The shorthand suite uses syntax-only Run with a nil Program. Its filenames
 # are independent fallback controls, not claimed runtime helper invocations.
 for file in sorted((scratch/'docs').glob('*.jsonl')):
  for line in file.read_text().splitlines():
   row=json.loads(line)
   name=row.get('Rule',row.get('rule'))
   if name in consumers and name not in actual:
    filename=row.get('File',row.get('file'));source=row.get('Source',row.get('source'))
    rows.append({'Rule':name,'File':filename,'Source':source,'Config':'','Cwd':str(Path(filename).parent),'Present':False,'Origin':'nil-program-fixture filename control'})
 missing=consumers-{row['Rule']for row in rows}
 assert not missing,sorted(missing)
 (HERE/'cases.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
 print('original Go suite exit',result.returncode,'captured',len(rows),'cases from',len(consumers),'consumers; real program consumers',len(actual))
 if result.returncode:print('Original rule assertions failed; helper input capture is not a passing whole-rule gate. See evidence/capture.log.')
