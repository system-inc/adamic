import difflib, json, pathlib, subprocess, time
root=pathlib.Path.cwd()
out=pathlib.Path(__file__).parent
source=root/'internal/lower/collections.go'
original=source.read_text()
mutants=[
 ('wrong-member', 'value := ir.StringLength{Value: ir.Read{Local: held, Of: ir.String}}', 'value := ir.CharCodeAt{Value: ir.Read{Local: held, Of: ir.String}, Index: ir.NumberConstant{Value: 0}}', './internal/oracle', '^TestNativeAgreesWithNode_097$/internal/oracle/testdata/string_destructuring.a$', 'stdout differs'),
 ('skip-empty-source', 'return append([]ir.Statement{ir.Declare{Local: held, Value: value}}, declared...), nil', 'if value.Type() == ir.String && len(pattern.AsBindingPattern().Elements.Nodes) == 0 { return declared, nil }; return append([]ir.Statement{ir.Declare{Local: held, Value: value}}, declared...), nil', './internal/oracle', '^TestNativeAgreesWithNode_097$/internal/oracle/testdata/string_destructuring.a$', 'stdout differs'),
 ('missing-fix', "in object destructuring; call methods on the saved string receiver or wrap the call in an arrow", "in object destructuring", './internal/lower', '^TestStringDestructuringMemberDiagnostic$', 'want string member refusal with path and fix'),
]
results=[]
for name,old,new,package,test,catcher in mutants:
 assert original.count(old)==1,(name,original.count(old))
 mutated=original.replace(old,new)
 (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/internal/lower/collections.go',tofile='b/internal/lower/collections.go')))
 replacement=out/(name+'.go.txt');replacement.write_text(mutated)
 overlay=out/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(source):str(replacement.resolve())}}))
 command=['go','test','-overlay',str(overlay),package,'-run',test,'-count=1','-timeout','90s','-v','-parallel','4']
 start=time.monotonic()
 with (out/(name+'.log.txt')).open('w') as log:
  p=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=180,env=__import__('os').environ|{'ADAMIC_GATE_UNCACHED':'1'})
 log=(out/(name+'.log.txt')).read_text()
 caught=p.returncode==1 and catcher in log and '[build failed]' not in log and 'error: ' not in log and 'Sanitizer' not in log
 row=dict(name=name,command=command,exit=p.returncode,seconds=round(time.monotonic()-start,3),catcher=catcher,caught=caught)
 print(row,flush=True);results.append(row)
 (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
 assert caught,name
