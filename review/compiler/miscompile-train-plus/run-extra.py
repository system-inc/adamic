from pathlib import Path
import subprocess, difflib, json, re
root=Path.cwd()
evidence=root/'review/compiler/miscompile-train-plus/extra'
evidence.mkdir(exist_ok=True)
receipts=json.loads((evidence/'results.json').read_text()) if (evidence/'results.json').exists() else []
def run(name, changes, packages, test, want=1, catcher=None):
    if any(r['name']==name for r in receipts): return
    originals={p:Path(p).read_text() for p in changes}
    patch=''
    try:
        for p,f in changes.items():
            before=originals[p]; after=f(before)
            assert after!=before,(name,p)
            patch+=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+p,tofile='b/'+p))
            Path(p).write_text(after)
        (evidence/(name+'.patch')).write_text(patch)
        command=['go','test',*packages,'-run',test,'-count=1','-v','-timeout','90s']
        with (evidence/(name+'.log')).open('w') as log:
            result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=120)
        output=(evidence/(name+'.log')).read_text()
        assert result.returncode==want and '[build failed]' not in output,(name,output)
        if catcher: assert catcher in output,(name,output)
        receipts.append(dict(name=name,command=command,exit=result.returncode,catcher=catcher))
        (evidence/'results.json').write_text(json.dumps(receipts,indent=2)+'\n')
        print(name+': '+('caught' if want else 'control passed'),flush=True)
    finally:
        for p,s in originals.items():Path(p).write_text(s)
def patch_changes(patch):
    text=Path(patch).read_text()
    paths=re.findall(r'^\+\+\+ [ab]/(.*)$',text,re.M)
    originals={p:Path(p).read_text() for p in paths}
    try:
        subprocess.run(['git','apply','--unidiff-zero',patch],check=True,timeout=10)
        changed={p:Path(p).read_text() for p in paths}
    finally:
        for p,s in originals.items():Path(p).write_text(s)
    return {p:lambda s,v=v:v for p,v in changed.items()}
run('spread-method-original-masked',patch_changes('review/compiler/fx6-key-read/spread-method-bypass.diff'),['./internal/lower'],'^TestViewSpreadMethodRefused$',0,'PASS')
print('FINDING: original spread-method mutant survives because creation admission independently refuses it; the adapted fx6 run removes both guards and is caught',flush=True)
run('constructor-revert',patch_changes('review/compiler/fx6-candidates-3/constructor-revert.patch'),['./internal/oracle'],'TestNativeAgreesWithNode/internal/oracle/testdata/(nbody_static_collision|field_write_paths|class_features_static|class_features_static_private).a',catcher='exit codes differ')
unsupported={'internal/lower/interface_cast.go':lambda s:s.replace('\tif err := l.checkViewMembers(node, target); err != nil {\n\t\treturn nil, err\n\t}\n',''), 'internal/lower/view_lazy.go':lambda s:s[:s.index('// Every member must have a complete checked lowering')]} 
run('unsupported-view-revert',unsupported,['./internal/lower','./internal/oracle'],'TestUnsupportedView|TestCheckedViewUntaggedOwnClassData.*Refused|TestCheckedViewUntaggedArray.*Refused|TestCheckedViewV2ArrayArmBoundary',catcher='FAIL')
run('unsupported-stage3-revert',unsupported,['./stage3/fixtures'],'^TestFixturesAssertions$/assertions/09_interface_kind.a',catcher='FAIL')
reader='internal/native/view_unions_read.go'
obj='internal/native/runtime/object.c'
def old_reader(s):
    a='\te.line("else if (%s.kind >= adamic_view_union_string && %s.kind <= adamic_view_union_function) %s = adamic_retain(%s.payload.reference);", snapshot, snapshot, boxed, snapshot)\n\te.line("else %s = NULL;", boxed)'
    b='\te.line("else %s = adamic_retain(%s.payload.reference);", boxed, snapshot)'
    assert s.count(a)==1
    return s.replace(a,b)
def old_runtime(s):
    a='    // Unknown storage is the one exception: its readers decide from the raw slot.\n    if (storage == 0) { value.payload = *slot; }'
    assert s.count(a)==1
    return s.replace(a,'    value.payload = *slot;')
run('snapshot-runtime-alone',{reader:old_reader},['./internal/oracle'],'^TestViewSnapshotMaybeBoolean$',0)
run('snapshot-compiler-alone',{obj:old_runtime},['./internal/oracle'],'^TestViewSnapshotMaybeBoolean$',0)
run('snapshot-runtime-mutant',{reader:old_reader,obj:lambda s:s.replace('        adamic_maybe_boolean boolean =','        value.payload = *slot;\n        adamic_maybe_boolean boolean =')},['./internal/oracle'],'^TestViewSnapshotMaybeBoolean$',catcher='native:')
run('snapshot-compiler-mutant',{reader:old_reader,obj:old_runtime},['./internal/oracle'],'^TestViewSnapshotMaybeBoolean$',catcher='native:')
run('snapshot-reference-mutant',{obj:lambda s:s.replace('    if (storage >= adamic_rep_record) { value.payload.reference = slot->reference; }\n','')},['./internal/oracle'],'^TestViewSnapshotTypedArrayReference$',catcher='view_snapshot_test.go:109: exit codes differ')
# Repeat the recorded numeric zero and six generated-C mutants on current output.
p=Path('internal/lower/callable_union_result_test.go')
s=p.read_text()
run('callable-zero-reference',{},['./internal/lower'],'^TestCallableUnionResultReferenceMutant$',0, 'Native backend stdout')
run('callable-six-reference',{str(p):lambda s:s.replace('callable_union_result_zero.a','callable_union_result_number.a').replace('if !strings.Contains(message, "Native backend stdout") {','if !strings.Contains(message, "Native backend") {')},['./internal/lower'],'^TestCallableUnionResultReferenceMutant$',0,'Native backend')
print('all extra mutant obligations caught',flush=True)
