"""Run storage contract mutants using Go overlays, without editing compiler files."""
from pathlib import Path
import json, os, subprocess, sys
root = Path(__file__).resolve().parents[2]
out = Path(os.environ.get('STORAGE_MUTANTS', '/workspace/scratch/step18-storage-mutants'))
out.mkdir(parents=True, exist_ok=True)
cases = []
def add(name, path, before, after, test):
    source = (root/path).read_text()
    assert before in source, (name, path)
    cases.append((name, path, source.replace(before, after, 1), test))
js = 'internal/javascript/optional_calls.go'
native = 'internal/native/emit_functions.go'
add('receiver-lost', native,
    'call = fmt.Sprintf("adamic_closure_receiver_call(%s, %s, %s, %s, %d)", closure, receiver, packed, count, e.packedArgumentSize(expression))',
    'call = fmt.Sprintf("adamic_closure_receiver_call(%s, %s, %s, %s, %d)", closure, "NULL", packed, count, e.packedArgumentSize(expression))',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/field.a')
add('field-reselected', js, 'const fn = %s;', 'let fn = %s;',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/field.a')
# Re-read after argument effects instead of invoking the selection made before them.
name,path,source,test = cases[-1]
cases[-1] = (name,path,source.replace('if (!(fn instanceof AdamicClosure)', 'fn = object.callback; if (!(fn instanceof AdamicClosure)',1),test)
add('arguments-before-guard', js, 'const fn = %s;', 'const eager = () => object; const fn = %s;',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/field.a')
name,path,source,test=cases[-1]
cases[-1]=(name,path,source.replace('%s%sconst values', 'const premature = [%s]; void eager; void premature; %s%sconst values',1).replace('guard, presence, e.callValues(call.Arguments, call.Spread), method','e.callValues(call.Arguments, call.Spread), guard, presence, e.callValues(call.Arguments, call.Spread), method',1),test)
add('noncallable-no-loud-stop',native,
    'e.line("if (%s) adamic_panic(\\"TypeError: optional call value is not callable\\", 46);", invalid)',
    'e.declarations = append(e.declarations, "#include <stdlib.h>"); e.line("if (%s) exit(0);", invalid)', 'TestStep18StorageTypeLie/.ts')
add('noncallable-stop-before-arguments',native,
    'packed, count := e.closureArguments(expression)',
    'if invalid != "" { e.line("if (%s) adamic_panic(\\"TypeError: optional call value is not callable\\", 46);", invalid) }; packed, count := e.closureArguments(expression)',
    'TestStep18StorageTypeLie/.ts')
add('adamic-noncallable-admitted','internal/lower/optional_callable.go',
    'if held == ir.Union && strings.HasSuffix(', 'if false && held == ir.Union && strings.HasSuffix(',
    'TestStep18StorageTypeLie/.a')

chain = 'internal/lower/optional_chain.go'
add('nullish-chain-arguments',chain,
    'Body: []ir.Statement{ir.Declare{Local: local, Value: current}}, Result:',
    'Body: func() []ir.Statement { body := []ir.Statement{ir.Declare{Local:local, Value:current}}; for _, last := range steps { if last.Kind == ast.KindCallExpression { for _, argument := range last.AsCallExpression().Arguments.Nodes { value, err := l.expression(argument); if err != nil { panic(err) }; body=append(body,ir.Evaluate{Value:value}) } } }; return body }(), Result:',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/optional-method.a')

add('getter-twice', js, 'const fn = %s;', 'void %s; const fn = %s;',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/getter.a')
name,path,source,test=cases[-1]
cases[-1]=(name,path,source.replace('selection, quote(property.Name), guard,','selection, selection, quote(property.Name), guard,',1),test)
add('getter-throw-delayed', native,
    'if property.CallableAccessor {\n\t\t\t\te.closureThrown()\n\t\t\t}',
    'if property.CallableAccessor { /* Mutant delays the getter throw until after the call. */ }',
    'TestNativeAgreesWithNode/docs/step-18/storage-fixtures/getter-throw.a')
add('receiver-contract-erased', 'internal/lower/library_function_expressions.go',
    'if ownerType == nil || !l.iterationShapeFits(l.concrete(ownerType), receiverType) || l.nominalMismatch(l.concrete(ownerType), receiverType, map[[2]*checker.Type]bool{}) != nil || l.widened(l.concrete(ownerType), receiverType, map[[2]*checker.Type]bool{}) != nil {',
    'if false && (ownerType == nil || !l.iterationShapeFits(l.concrete(ownerType), receiverType) || l.nominalMismatch(l.concrete(ownerType), receiverType, map[[2]*checker.Type]bool{}) != nil || l.widened(l.concrete(ownerType), receiverType, map[[2]*checker.Type]bool{}) != nil) {',
    'LOWER:TestStep18CallableReceiverContract')
results=[]
for name,path,source,test in cases:
    if len(sys.argv)>1 and name not in sys.argv[1:]: continue
    directory=out/name;directory.mkdir(exist_ok=True)
    replacement=directory/'mutant.go';replacement.write_text(source)
    overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(replacement)}}))
    package='./internal/oracle'
    if test.startswith('LOWER:'): package='./internal/lower'; test=test.removeprefix('LOWER:')
    command=['go','test','-p','1','-overlay='+str(overlay),package,'-run',test,'-count=1','-timeout','5m','-v']
    with (directory/'test.log').open('w') as log:
        run=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
    output=(directory/'test.log').read_text()
    caught=run.returncode!=0 and '--- FAIL:' in output and not any(x in output for x in ['[build failed]','clang failed:','no tests to run'])
    results.append({'name':name,'caught':caught,'exit':run.returncode,'test':test,'log':str(directory/'test.log')})
    print(name, 'CAUGHT' if caught else 'NOT CREDITED',flush=True)
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
if not results or not all(x['caught'] for x in results): sys.exit(1)
