import json,os,subprocess,tempfile,gzip,collections,hashlib
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[5];COHERE=ROOT/'cohere'
with tempfile.TemporaryDirectory() as scratch:
 tmp=Path(scratch);replace={};source=COHERE/'internal/lint/ecmascript/decorators/decorators.go';text=source.read_text()
 for name in ['CallName','HasDecoratorInSet','Of']:text=text.replace('func '+name+'(', 'func adamicOriginal'+name+'(',1)
 side=tmp/'decorators.go';side.write_text(text);replace[str(source)]=str(side)
 wrappers=tmp/'wrappers.go';wrappers.write_text('''//go:build lintoracle
package decorators
import "github.com/microsoft/TypeScript/tsc/shim/ast"
func CallName(node *ast.Node) string {value:=adamicOriginalCallName(node);recordAdamicAst("CallName",node,nil,value);return value}
func HasDecoratorInSet(node *ast.Node,names map[string]struct{}) bool {value:=adamicOriginalHasDecoratorInSet(node,names);recordAdamicAst("HasDecoratorInSet",node,names,value);return value}
func Of(node *ast.Node) []*ast.Node {value:=adamicOriginalOf(node);recordAdamicAst("Of",node,nil,value);return value}
''');replace[str(source.parent/'adamic_wrappers.go')]=str(wrappers);replace[str(source.parent/'adamic_recorder.go')]=str(HERE/'recorder.go')
 replace[str(source.parent/'adamic_controls_test.go')]=str(HERE/'controls.go')
 # Capture every source/options run, including tests asserting diagnostics directly.
 harness=COHERE/'internal/lint/testing/rule_testing.go';h=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}';assert h.count(anchor)==1;h=h.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result')
 side=tmp/'rule_testing.go';side.write_text(h);replace[str(harness)]=str(side)
 typed=COHERE/'internal/lint/testing/program.go';h=typed.read_text()
 original="return Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile:  sourceFile,\n\t\tcapture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}"
 assert h.count(original)==1
 h=h.replace(original,original.replace('return Result{','result := Result{')+'\n RecordAssertedCase(t,result)\n return result')
 side=tmp/'program.go';side.write_text(h);replace[str(typed)]=str(side)
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 fixture=tmp/'fixtures';fixture.mkdir();counts={}
 for label,packages,pattern in [('calls',['./internal/lint/rules/base','./internal/lint/rules/core'],'Test'),('controls',['./internal/lint/ecmascript/decorators'],'TestAdamicDecoratorControls')]:
  arenas=[];calls=[]
  for part,package in enumerate(packages):
   records=tmp/(label+str(part)+'.jsonl')
   with (tmp/(label+str(part)+'.log')).open('w') as log:
    result=subprocess.run(['go','test','-tags=lintoracle','-overlay='+str(overlay),package,'-run',pattern,'-count=1','-timeout=10m'],cwd=COHERE,env=os.environ|{'ADAMIC_AST_CAPTURE':str(records),'COHERE_DOCS_CAPTURE':str(fixture)},stdout=log,stderr=subprocess.STDOUT)
   if result.returncode:print((tmp/(label+str(part)+'.log')).read_text());result.check_returncode()
   offset=len(arenas)
   for line in records.read_text().splitlines():
    row=json.loads(line)
    if 'Nodes' in row:row['ID']+=offset;arenas.append(row)
    else:row['Arena']+=offset;calls.append(row)
  assert {r['Name'] for r in calls}=={'CallName','HasDecoratorInSet','Of'}
  (HERE/(label+'.json.gz')).write_bytes(gzip.compress(json.dumps({'Arenas':arenas,'Calls':calls},ensure_ascii=True).encode(),mtime=0))
  counts[label]={'arenas':len(arenas),'calls':len(calls),'symbols':dict(collections.Counter(r['Name'] for r in calls))};print(label,counts[label],flush=True)
 rows=[]
 for f in fixture.glob('*.jsonl'):
  rows.extend(json.loads(line) for line in f.read_text().splitlines() if '"rule":"base/security-require-context-access"' in line[:110])
 unique={json.dumps([r['rule'],r['file'],r['options'],r['source']]):r for r in rows if r['rule']=='base/security-require-context-access'}
 (HERE/'rule-cases.json.gz').write_bytes(gzip.compress(json.dumps(list(unique.values()),ensure_ascii=True).encode(),mtime=0));counts['ruleCases']=len(unique)
 counts['goPin']=subprocess.check_output(['git','rev-parse','HEAD'],cwd=COHERE).decode().strip();counts['goSourceSha256']=hashlib.sha256(source.read_bytes()).hexdigest()
 (HERE/'coverage.json').write_text(json.dumps(counts,indent=2)+'\n');print('rule cases',len(unique))
