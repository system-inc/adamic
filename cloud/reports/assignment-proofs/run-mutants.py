from pathlib import Path
import subprocess,json,os
root=Path('/workspace/adamic');d=Path('/workspace/scratch/assignment-proofs')
a=(root/'internal/lower/assignment_value.go').read_text()
n=(root/'internal/native/taste.go').read_text()
i=(root/'internal/lower/invariance.go').read_text()
c=(root/'internal/lower/closed_frame_inputs.go').read_text()
specs=[
('proof-origin','internal/lower/assignment_value.go',a.replace('node.AsBinaryExpression().Right)', 'node.AsBinaryExpression().Left)',1),'./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_scanner'),
('rhs-twice','internal/lower/assignment_value.go',a.replace('saved := l.assignmentSnapshot(&body, "assignment_value", value)','if _,call := value.(ir.Call); call { body = append(body, ir.Evaluate{Value: value}) }\n\tsaved := l.assignmentSnapshot(&body, "assignment_value", value)',1),'./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_chain'),
('result-retain','internal/native/taste.go',n.replace('held, retained(result)', 'held, result',1),'./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_order'),
('temporary-cleanup','internal/native/taste.go',n.replace('e.releaseScopes(len(e.scopes) - 1)', '// mutant: omit binding releases',1),'./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_while'),
('reference-order','internal/lower/assignment_value.go',a.replace('body = append(body, writes[0])','if len(body)>1 { body[0],body[len(body)-1]=body[len(body)-1],body[0] }\n\tbody = append(body, writes[0])',1),'./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/assignment_order'),
('effects-writes','internal/lower/closed_frame_inputs.go',c.replace('switch value := node.(type) {','if _,ok:=node.(ir.Effects);ok { return false }; switch value := node.(type) {',1),'./internal/lower','TestAssignmentProofsClosedInputStillInspectsEffects'),
('alias-proof','internal/lower/invariance.go',i.replace('if l.canWrite(own, map[*checker.Type]bool{}) {','if false {',1),'./internal/lower','TestAssignmentProofsDoNotInventMembersOrUniqueAliases'),
]
results=[]
for name,file,source,package,test in specs:
 assert source != (root/file).read_text(),name
 replacement=d/(name+'.go');replacement.write_text(source)
 overlay=d/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(replacement)}},indent=2))
 command=['go','test','-overlay',str(overlay),package,'-run',test,'-count=1','-timeout','10m']
 with (d/(name+'.log')).open('w') as log:
  run=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
 evidence=(d/(name+'.log')).read_text()
 expected={'proof-origin':'adamic/enum-literal','rhs-twice':'stdout differs','result-retain':'AddressSanitizer: heap-use-after-free','temporary-cleanup':'LeakSanitizer: detected memory leaks','reference-order':'stdout differs','effects-writes':'an internal field store was hidden by Effects','alias-proof':'want refusal, got <nil>'}
 assert expected[name] in evidence,(name,evidence[:1000])
 assert 'clang failed' not in evidence,(name,evidence[:1000])
 results.append({'name':name,'command':command,'exit':run.returncode,'caught_by':expected[name]})
 print(name,run.returncode,flush=True)
(d/'mutant-results.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(r['exit']!=0 for r in results),results
