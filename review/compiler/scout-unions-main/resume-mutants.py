import json,pathlib,subprocess,time,sys
root=pathlib.Path.cwd(); evidence=root/'review/compiler/scout-unions-main'
cases=json.loads((evidence/'nullable-string-mutants.json').read_text())
cases += [['class-zero','internal/lower/class.go','return ir.MaybeOf{Of: ir.MaybeBoolean}','return ir.MaybeOf{Of: ir.MaybeBoolean, Value: ir.BooleanConstant{}}','./internal/oracle','TestNativeAgreesWithNode/internal/oracle/testdata/scout_class_boolean_fields.a'],['class-optional','internal/lower/class_inheritance.go','of != ir.MaybeBoolean || member.PostfixToken() != nil','of != ir.MaybeBoolean','./internal/lower','TestScoutClassBooleanKeepsUnsafeViewsRefused']]
for name,old,new,pattern in json.loads((evidence/'boxed-mutants.json').read_text()): cases.append([name,'internal/lower/boxed_scalar_fields.go',old,new,'./internal/oracle',pattern])
for name,path,old,new,pkg,pattern in cases[int(sys.argv[1]) if len(sys.argv)>1 else 0:]:
    source=(root/path).read_text(); assert old in source, name
    mutant=evidence/('resume-'+name+'.go.txt'); mutant.write_text(source.replace(old,new,1))
    overlay=evidence/('resume-'+name+'.overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/path):str(mutant)}}))
    start=time.monotonic()
    with (evidence/('resume-'+name+'.log')).open('w') as log:
        result=subprocess.run(['timeout','180','go','test','-overlay',str(overlay),pkg,'-run',pattern,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT)
    output=(evidence/('resume-'+name+'.log')).read_text()
    caught=result.returncode==1 and '--- FAIL:' in output and ('stdout' in output or 'must remain' in output or 'stage 0' in output or 'unexpectedly' in output or 'got <nil>' in output)
    print(name,result.returncode,round(time.monotonic()-start,3),'caught' if caught else 'UNVERIFIED',flush=True)
    if not caught: raise SystemExit(1)
