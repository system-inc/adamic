from pathlib import Path
import json, subprocess, re, os
ROOT=Path(__file__).resolve().parents[2]; OUT=Path(os.environ.get('OPTIONAL_CALL_MUTANTS', '/tmp/optional-calls-main-mutants')); OUT.mkdir(parents=True, exist_ok=True)
os.environ['ADAMIC_GATE_UNCACHED']='1'
LOW='internal/lower/optional_callable.go'; CHAIN='internal/lower/optional_chain.go'; NATIVE='internal/native/optional_calls.go'; JS='internal/javascript/optional_calls.go'
cases=[]
def add(name,path,old,new,fixture=None,test=None):
    original=(ROOT/path).read_text(); assert old in original,(name,path,old)
    cases.append((name,{path:original.replace(old,new,1)},fixture,test))
def many(name,edits,fixture=None,test=None):
    contents={}
    for path,old,new in edits:
        original=contents.get(path,(ROOT/path).read_text()); assert old in original,(name,path,old)
        contents[path]=original.replace(old,new,1)
    cases.append((name,contents,fixture,test))
add('callable-reselect',LOW,'call.Closure = read','call.Closure = selected','runtime-order')
add('callable-null-guard',LOW,'Right: ir.IsNull{Value: read}','Right: ir.BooleanConstant{Value: false}','runtime-values')
add('callable-eager-arguments',LOW,'var present ir.Expression = call','var present ir.Expression = call\n eager := []ir.Statement{}\n for _, argument := range call.Arguments { eager = append(eager, ir.Evaluate{Value: argument}) }\n present = ir.Effects{Body: eager, Result: present}','runtime-arguments')
# Eager effects must live outside the conditional, not in its present result.
name,contents,fixture,test=cases[-1]
contents[LOW]=contents[LOW].replace('present = ir.Effects{Body: eager, Result: present}','').replace('Body: []ir.Statement{ir.Declare{Local: local, Value: selected}}, Result:', 'Body: append([]ir.Statement{ir.Declare{Local: local, Value: selected}}, eager...), Result:')
add('callable-result-zero',LOW,'var present ir.Expression = call','var present ir.Expression = call\n if call.Returns == ir.Number { present = ir.NumberConstant{Value: 0} }','runtime-values')
add('callable-narrowing','internal/lower/narrowed.go','if call.Expression == node && call.QuestionDotToken != nil {','if false && call.Expression == node && call.QuestionDotToken != nil {','runtime-order')
add('return-descriptor','internal/lower/expression.go','result = l.checker.GetReturnTypeOfSignature(signatures[0])','result = l.checker.GetTypeAtLocation(node)','runtime-return-descriptor')
add('method-receiver',NATIVE,'e.callSelected(call, closure, receiver, method, exactCount)','e.callSelected(call, closure, "NULL", method, exactCount)','runtime-methods')
add('method-presence',NATIVE,'present += " || " + selected','_ = selected','runtime-bound-methods')
add('method-reselection',JS,'const values = [%s]; return Object.hasOwn','const values = [%s]; const selected = object[PLACEHOLDER]; return Object.hasOwn','runtime-bound-method-selection')
name,contents,fixture,test=cases[-1]
contents[JS]=contents[JS].replace('const fn = object[%s]','let fn = object[%s]').replace('const selected = object[PLACEHOLDER];','fn = object.report;')
add('method-eager-spread',NATIVE,'text, value, owned := e.asideWith(func() string {','if call.Spread != nil { _, _ = e.closureArguments(call) }\n text, value, owned := e.asideWith(func() string {','runtime-bound-method-arguments')
add('method-retain','internal/native/emit_functions.go','fmt.Sprintf("adamic_retain(%s(%s, %s, &%s, &%s))", lookup, receiver, cString(property.Name), e.cache(), method)','fmt.Sprintf("%s(%s, %s, &%s, &%s)", lookup, receiver, cString(property.Name), e.cache(), method)','runtime-bound-method-selection')
add('discarded-reference','internal/native/discarded_closure_result.go','e.line("if (%s) adamic_release(%s.reference);", condition, result)','_ = condition; _ = result','runtime-discarded-reference')
add('chain-short-circuit',CHAIN,'WhenTrue:  present, WhenNot: fit(ir.Undefined{}, of), Of: of,','WhenTrue:  present, WhenNot: present, Of: of,','call-result')
add('chain-receiver-repeat',CHAIN,'present, err := apply(read)','present, err := apply(current)\n _ = read','runtime-chain-index')
add('chain-index-eager',CHAIN,'Body: []ir.Statement{ir.Declare{Local: local, Value: current}}, Result:', 'Body: func() []ir.Statement { body := []ir.Statement{ir.Declare{Local: local, Value: current}}; if step.Kind == ast.KindElementAccessExpression { key, err := l.expression(step.AsElementAccessExpression().ArgumentExpression); if err != nil { panic(err) }; body = append(body, ir.Evaluate{Value:key}) }; return body }(), Result:','runtime-chain-index')
add('chain-method-presence',NATIVE,'e.line("%s = true;", e.localName(call.OptionalPresent-1))','e.line("%s = false;", e.localName(call.OptionalPresent-1))','runtime-chain-method-continuation')
add('chain-effects-leak','internal/native/taste.go','e.releaseScopes(depth)','// Mutant omits effects local releases.','runtime-chain-index')
add('chain-ordinary-read',CHAIN,'return l.defined(node, l.readObjectField(node, ir.Property{Object: object, Name: name, Of: stored, Class: l.classOf(node)})), nil','return l.readObjectField(node, ir.Property{Object: object, Name: name, Of: stored, Class: l.classOf(node)}), nil','runtime-chain-present-undefined')
add('chain-parentheses',CHAIN,'return ir.Defined{Value: value, Message: "TypeError: Cannot read properties of undefined (reading \'" + parent.Name().Text() + "\')"}, nil','return value, nil','runtime-chain-parentheses')
add('chain-write-order','internal/native/emit_statements.go','object := e.value(statement.Object)\n\t\tvalue := e.value(statement.Value)\n\t\t// The object may be undefined','object := e.value(statement.Object)\n\t\te.line("if (%s == NULL) adamic_panic(\\"early write\\", 11);", object)\n\t\tvalue := e.value(statement.Value)\n\t\t// The object may be undefined','runtime-chain-parentheses-write')
for name,path in [('chain-map-visitor','internal/native/emit_maps.go'),('chain-array-visitor','internal/native/emit_arrays.go')]:
    add(name,path,'e.discardClosureResult(ir.CallClosure{}, callback, "", "", call, holds...)','e.line("%s;", call)\n _ = holds','runtime-chain-intrinsics')
add('chain-array-map-storage',CHAIN,'Result: result, CallbackType:','Result: func() ir.Type { _ = result; return ir.Number }(), CallbackType:','runtime-chain-intrinsics')
# Evaluate the already-lowered callback factory before the guard; no duplicate lambda lowering.
add('chain-callback-eager',CHAIN,'present, err := apply(read)','present, err := apply(read)\n eager := []ir.Statement{}\n if i+1 < len(steps) && steps[i+1].Kind == ast.KindCallExpression && step.Kind == ast.KindPropertyAccessExpression && step.Name().Text() == "forEach" { if visit, ok := present.(ir.Effects); ok { for _, statement := range visit.Body { if evaluated, ok := statement.(ir.Evaluate); ok { if call, ok := evaluated.Value.(ir.MapForEach); ok { eager = append(eager, ir.Evaluate{Value: call.Callback}) } } } } }','runtime-chain-intrinsics')
name,contents,fixture,test=cases[-1]
contents[CHAIN]=contents[CHAIN].replace('Body: []ir.Statement{ir.Declare{Local: local, Value: current}}, Result:','Body: append([]ir.Statement{ir.Declare{Local: local, Value: current}}, eager...), Result:')
add('chain-weak-storage',CHAIN,'if stored, known := l.representation(memberType); !known || stored != ir.Closure || l.weakTarget(memberType) != nil {','if stored, known := l.representation(memberType); !known || (stored != ir.Closure && stored != ir.Weak) {',test='TestStep18OptionalCallableBoundaries/chain_weak_callable_storage')
add('chain-accessor',CHAIN,'if symbol == nil || accessorSymbol(symbol) || l.accessorNames[member.Name().Text()] {','if symbol == nil {',test='TestStep18OptionalCallableBoundaries/chain_getter_selection')
import sys
results=json.loads((OUT/"results.json").read_text()) if any(flag in sys.argv for flag in ("--retry", "--inputs-only")) else []
for name,contents,fixture,test in cases:
    if "--inputs-only" in sys.argv: continue
    if "--retry" in sys.argv and any(r["name"] == name and r["caught"] for r in results): continue
    results=[r for r in results if r["name"] != name]
    directory=OUT/name; directory.mkdir(exist_ok=True)
    replacements={}
    for n,(path,data) in enumerate(contents.items()):
        target=directory/(str(n)+'.go'); target.write_text(data); replacements[str(ROOT/path)]=str(target)
    overlay=directory/'overlay.json'; overlay.write_text(json.dumps({'Replace':replacements}))
    package='./internal/lower' if test else './internal/oracle'
    selected=test or 'TestNativeAgreesWithNode/docs/step-18/fixtures/'+fixture+'.a'
    if name in ('discarded-reference', 'chain-map-visitor', 'chain-array-visitor'):
        selected='TestStep18DiscardedReferencesBalance/'+fixture+'.a'
    command=['go','test','-p','1','-overlay='+str(overlay),package,'-run',selected,'-count=1','-timeout','5m','-v']
    with (directory/'test.log').open('w') as output: run=subprocess.run(command,cwd=ROOT,stdout=output,stderr=subprocess.STDOUT)
    log=(directory/'test.log').read_text()
    build_failure=any(s in log for s in ('[build failed]','clang failed:','error: no member','syntax error:'))
    caught=run.returncode!=0 and '--- FAIL:' in log and not build_failure
    results.append(dict(name=name,fixture=fixture,test=test,exit=run.returncode,caught=caught,build_failure=build_failure,log=str(directory/'test.log')))
    (OUT/'results.json').write_text(json.dumps(results,indent=2)+'\n')
    print(name,'CAUGHT' if caught else 'NOT CREDITED',flush=True)
# Source observations and exact gap diagnoses are independent test inputs.
for name,file,modify,test in [
 ('source-baseline','function.a',lambda text:text.replace("'cleanup'","'cleanup-mutant'"),'TestStep18SourceBaselines/function.a'),
 ('gap-masking','number.a',lambda text:'const barrier: bigint = 1n;\n'+text,'TestStep18RecordedGaps/number.a')]:
    path=ROOT/'docs/step-18/fixtures'/file
    original=path.read_bytes()
    mutated=modify(original.decode()).encode()
    assert mutated != original
    directory=OUT/name; directory.mkdir(exist_ok=True)
    try:
        path.write_bytes(mutated)
        with (directory/'test.log').open('w') as output:
            run=subprocess.run(['go','test','-p','1','./internal/oracle','-run',test,'-count=1','-v'],cwd=ROOT,stdout=output,stderr=subprocess.STDOUT)
    finally:
        path.write_bytes(original)
    log=(directory/'test.log').read_text()
    caught=run.returncode != 0 and '--- FAIL:' in log and '[build failed]' not in log
    results=[row for row in results if row['name'] != name]
    results.append(dict(name=name,caught=caught,exit=run.returncode,log=str(directory/'test.log')))
    print(name,'CAUGHT' if caught else 'NOT CREDITED',flush=True)

# Synthetic records test reader completeness only, without a compiler census claim.
directory=OUT/'census-completeness'; directory.mkdir(exist_ok=True)
rows=[{'measurement':'synthetic reader completeness probe; no compiler observations'}]+[{'file':str(index),'findings':[]} for index in range(79)]
before=directory/'before.jsonl'; after=directory/'after.jsonl'
before.write_text(''.join(json.dumps(row)+'\n' for row in rows)); after.write_text(before.read_text())
command=['python3','docs/step-18/compare-roots.py',str(before),str(after),str(directory/'comparison.json')]
with (directory/'control.log').open('w') as output:
    control=subprocess.run(command,cwd=ROOT,stdout=output,stderr=subprocess.STDOUT)
assert control.returncode == 0
rows.pop(); after.write_text(''.join(json.dumps(row)+'\n' for row in rows))
with (directory/'test.log').open('w') as output:
    run=subprocess.run(command,cwd=ROOT,stdout=output,stderr=subprocess.STDOUT)
caught=run.returncode != 0 and 'AssertionError' in (directory/'test.log').read_text()
results=[row for row in results if row['name'] != 'census-completeness']
results.append(dict(name='census-completeness',caught=caught,exit=run.returncode,log=str(directory/'test.log'),scope='reader completeness only; no acceptance census'))
(OUT/'results.json').write_text(json.dumps(results,indent=2)+'\n')
print(sum(row['caught'] for row in results),'of',len(results),'caught',flush=True)
raise SystemExit(any(not row['caught'] for row in results))
