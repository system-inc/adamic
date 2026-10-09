"""Run independent placeholder compiler mutants against pinned behavioral tests."""
import json, os, pathlib, subprocess, time
root = pathlib.Path(__file__).resolve().parents[3]
evidence = pathlib.Path(__file__).resolve().parent
cases = [
 ('drop-flow-check','internal/lower/placeholder.go','proven := placeholderProven(use.Value, program, ready, fields)','proven := true','./internal/oracle','^TestPlaceholderUseCheckSavedLeak$'),
 ('retain-alias-proof','internal/lower/readiness.go','if field.name == set.Name {','if false && field.name == set.Name {','./internal/oracle','^TestPlaceholderUseCheckAliasReset$'),
 ('nullish-receiver-twice','internal/lower/expression.go','if _, simple := value.(ir.Read); !simple {','if _, simple := value.(ir.Read); false && !simple {','./internal/oracle','^TestNativeAgreesWithNode$/internal/oracle/testdata/placeholder_nonnull_null_loose[.]a$'),
 ('json-ordinary-assertion','internal/lower/library_json_stringify.go','if l.uninitializedInitializer(n) && l.placeholderDeclaration(n) {','if false && l.uninitializedInitializer(n) && l.placeholderDeclaration(n) {','./internal/oracle','^TestNativeAgreesWithNode$/internal/oracle/testdata/placeholder_nonnull_null_json[.]a$'),
 ('admit-weak-slot','internal/lower/expression.go','return 0, l.notYet(node, "a placeholder slot holding Weak; its target-read semantics need a separate checked boundary")','return ir.Union, nil','./internal/lower','^TestPlaceholderWeakSlotStaysNotYet$'),
 ('omit-refusal-state','internal/lower/placeholder.go','if l.result == nil {\n\t\tl.result = &ir.Program{}\n\t}','if false && l.result == nil {\n\t\tl.result = &ir.Program{}\n\t}','./internal/lower','^TestOptionalWideningSpreadOverwrite$'),
]
results=[]
for name,file,before,after,package,selector in cases:
 data=(root/file).read_text()
 assert before in data, (name, 'mutation missing')
 target=evidence/(name+'.go.txt');target.write_text(data.replace(before,after))
 overlay=evidence/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(target)}}))
 command=['go','test','-overlay',str(overlay),package,'-run',selector,'-count=1','-v']
 start=time.monotonic()
 log=evidence/'logs'/(name+'.log')
 with log.open('w') as output:
  run=subprocess.run(command,cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT)
 text=log.read_text()
 caught=run.returncode!=0 and ('--- FAIL:' in text or 'panic: runtime error' in text) and '[build failed]' not in text and 'clang failed' not in text
 results.append(dict(name=name,command=command,exit=run.returncode,caught=caught,seconds=time.monotonic()-start))
 print(name, 'caught' if caught else 'NOT CAUGHT',flush=True)
(evidence/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(row['caught'] for row in results)
