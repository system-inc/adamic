import json,pathlib,subprocess,time,sys
root=pathlib.Path.cwd(); evidence=root/'review/compiler/scout-unions-objects'
oracle='./internal/oracle'; differential='TestNativeAgreesWithNode/internal/oracle/testdata/'
cases=[]
for name,old,new,fixture in json.loads((evidence/'object-mutants.json').read_text()):
    cases.append([name,'internal/lower/nullable_objects.go',old.replace('\\n','\n').replace('\\t','\t'),new.replace('\\n','\n').replace('\\t','\t'),oracle,differential+fixture])
cases += [
 ['string-check','internal/lower/boxed_scalar_fields.go','ir.Unary{Operator: ir.Not, Operand: matches}','ir.BooleanConstant{}',oracle,differential+'scout_string_number_stale_field.a'],
 ['string-tag','internal/lower/boxed_scalar_fields.go','name = "string"','name = "number"',oracle,differential+'scout_string_number_fields_locals.a'],
 ['local-nullish','internal/lower/locals.go','observing := comparedWithUndefined(node)', 'if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 { read = ir.Box{Value: ir.StringConstant{Index: l.constant("mutant")}} }\n\t\tobserving := comparedWithUndefined(node)',oracle,differential+'scout_string_number_fields_locals.a'],
 ['regex-adapter','internal/lower/nullable_objects.go','if l.regexReplacementArgument(node) {','if false && l.regexReplacementArgument(node) {',oracle,differential+'regexp_replace/callback.a'],
 ['shared-view','internal/lower/expression.go','if l.nullableObjectViewHazard(node, contextual) {','if false && l.nullableObjectViewHazard(node, contextual) {',oracle,'^TestNullableObjectStorageArray$'],
 ['lookup','internal/lower/expression.go','if value.Type() == ir.Object && l.nullableObjectUnion','if false && value.Type() == ir.Object && l.nullableObjectUnion',oracle,'^TestNullableObjectStorageLookup$'],
 ['local-equality','internal/lower/locals.go',' || unionEqualityObservation(node)','',oracle,differential+'scout_string_number_fields_locals.a'],
 ['object-local-runtime','internal/lower/locals.go','if l.nullableObjectUnion(l.checker.GetTypeOfSymbol(l.symbol(node))) {','if false && l.nullableObjectUnion(l.checker.GetTypeOfSymbol(l.symbol(node))) {',oracle,differential+'scout_nullable_object_stale_local.a'],
 ['object-local','internal/lower/locals.go','if l.nullableObjectUnion(l.checker.GetTypeOfSymbol(l.symbol(node))) {','if false && l.nullableObjectUnion(l.checker.GetTypeOfSymbol(l.symbol(node))) {',oracle,differential+'scout_nullable_objects.a'],
]
(evidence/'mutants.json').write_text(json.dumps(cases,indent=2))
for name,path,old,new,pkg,pattern in cases[int(sys.argv[1]) if len(sys.argv)>1 else 0:]:
    source=(root/path).read_text(); assert old in source, name
    mutant=evidence/(name+'.go.txt'); mutant.write_text(source.replace(old,new,1))
    overlay=evidence/(name+'.overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/path):str(mutant)}}))
    start=time.monotonic()
    with (evidence/(name+'.log')).open('w') as log:
        result=subprocess.run(['timeout','180','go','test','-overlay',str(overlay),pkg,'-run',pattern,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT)
    output=(evidence/(name+'.log')).read_text()
    caught=result.returncode==1 and '--- FAIL:' in output and ('want the inserted check to fire' in output or 'stdout differs' in output or 'exit codes differ' in output or 'want explicit storage view stop, got <nil>' in output or 'stage 0' in output)
    print(name,result.returncode,round(time.monotonic()-start,3),'caught' if caught else 'UNVERIFIED',flush=True)
    if not caught: raise SystemExit(1)
