"""Plant one change at a time, require its catcher, restore original bytes."""
import difflib,json,pathlib,subprocess,time
root=pathlib.Path('/workspace/adamic'); out=root/'review/compiler/generic-values-next/lint-chain-refresh'
results=[]
def mutant(name,path,old,new,package,pattern,needle,count):
    path=root/path; saved=path.read_bytes(); text=saved.decode(); assert text.count(old)==1
    changed=text.replace(old,new)
    (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+str(path.relative_to(root)),tofile='b/'+str(path.relative_to(root)))))
    start=time.monotonic()
    try:
        path.write_text(changed)
        with (out/(name+'.log')).open('w') as log:
            result=subprocess.run(['timeout','90','go','test',package,'-run',pattern,'-count=1','-timeout','85s','-v'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
        output=(out/(name+'.log')).read_text()
        assert result.returncode==1,(name,result.returncode)
        assert output.count(needle)==count,(name,output)
        row=dict(name=name,exit=result.returncode,seconds=round(time.monotonic()-start,3),catcher=needle,count=count)
        results.append(row); print(json.dumps(row),flush=True)
    finally: path.write_bytes(saved)
    (out/'mutants-results.json').write_text(json.dumps(results,indent=2)+'\n')
mutant('widening-guard-mutant','internal/lower/unknown.go','known && held == ir.Closure {','known && held == ir.Closure && false {','./internal/oracle','^TestGenericValue(ObjectWideningRefused|UnknownWideningRefused|AddressOnlyMapMutant|AddressOnlyArrayMutant)$','must stop before emission: <nil>',4)
mutant('json-layout-mutant','internal/native/runtime/library_language.c','static adamic_object json = {.heap = {0, adamic_kind_object, 0}, .shape = &json_shape};','static adamic_closure json = {.heap = {0, adamic_kind_object, 0}};','./internal/native','^TestGenericValueLibraryIdentityViews$','--- FAIL: TestGenericValueLibraryIdentityViews',1)
mutant('json-shape-mutant','internal/native/runtime/library_language.c','static const adamic_shape json_shape = {.count = 0};','static const adamic_shape json_shape = {.count = 1};','./internal/native','^TestGenericValueLibraryIdentityViews$','--- FAIL: TestGenericValueLibraryIdentityViews',1)
for package,pattern in [('./internal/oracle','^TestGenericValue'),('./internal/native','^TestGenericValueLibraryIdentityViews$')]:
    with (out/('restored-'+package.split('/')[-1]+'.log')).open('w') as log:
        result=subprocess.run(['timeout','90','go','test',package,'-run',pattern,'-count=1','-timeout','85s','-v'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
    assert result.returncode==0
