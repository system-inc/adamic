import json,difflib,subprocess
from pathlib import Path
P=Path('review/test-audit/bridge-tsgo-registered_units');S=Path('/tmp/u004');w=[]
def add(id,file,before,after,rows,reason):
 base=subprocess.check_output(['git','show','origin/main:'+file],text=True);assert before in base;new=base.replace(before,after)
 diff=''.join(difflib.unified_diff(base.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(P/(id+'.diff')).write_text(diff)
 q=S/'witness'/id;q.mkdir(parents=True,exist_ok=True);target=q/Path(file).name;target.write_text(new);(q/'overlay.json').write_text(json.dumps({'Replace':{str(Path(file).resolve()):str(target)}}));w.append({'id':id,'file':file,'line':base.count('\n',0,base.index(before))+1,'before':before,'after':after,'rows':rows,'reason':reason,'overlay':str(q/'overlay.json')})
add('W_EQUAL','bridge/tsgo/units_test.go','bytes.Equal(truth, observed)','(bytes.Equal(truth, truth) && bytes.Equal(observed, observed))',['TestBridgeWrongPosition'+f for f in ['Sample','Checker','Parser','Types','Utilities']],'Treat every native answer as agreeing with Go, weakening the agreement comparison.')
add('W_STALE','bridge/tsgo/testdata/api.c',' assert(tsgo_query(handle, file, 14, &answer, &error) == TSGO_HANDLE);',' /* stale-query assertion disabled for witness audit */',['TestBridgeStaleHandle'],'Remove the stale-query assertion the planted failure witnesses.')
add('W_LINK','bridge/tsgo/bridge_test.go','if _, err := lower.Lower(context.Background(), loaded); err == nil {\n\t\tt.Fatal("lowering accepted an unlinked checker call")\n\t}','if _, err := lower.Lower(context.Background(), loaded); false && err == nil {\n\t\tt.Fatal("lowering accepted an unlinked checker call")\n\t}',['TestBridgeLinkage'],'Disable the link refusal comparison in the inner test.')
# A runtime sanitizer setting disables only leak detection, preserving the production code.
w.append({'id':'W_LSAN','file':'bridge/tsgo/units_test.go','line':0,'rows':['TestBridgeOutputFree','TestBridgeRegionOwnership'],'reason':'Disable the witnessed LeakSanitizer check using ASAN_OPTIONS=detect_leaks=0 and LSAN_OPTIONS=detect_leaks=0. No source diff.'})
# Disable AddressSanitizer on the C consumer, retaining the archive instrumentation.
add('W_ASAN','bridge/tsgo/products_test.go','"-O1", "-g", "-fsanitize=address,undefined", "-I",','"-O1", "-g", "-fsanitize=address,undefined", "-fno-sanitize=address", "-I",',['TestBridgeInputLength','TestBridgeOutputLength'],'Disable C consumer address instrumentation. The input overflow occurs in instrumented archive copying, so this may not weaken that checker enough; observe rather than assume.')
for item in w:
 if item['id']=='W_ASAN':
  file=item['file'];base=subprocess.check_output(['git','show','origin/main:'+file],text=True)
  new=base.replace('if name == "tsgo-asan.a" {','if false {').replace('"-O1", "-g", "-fsanitize=address,undefined", "-I",','"-O1", "-g", "-I",')
  (S/'witness'/'W_ASAN'/'products_test.go').write_text(new)
  (P/'W_ASAN.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
  item['reason']='Disable AddressSanitizer in both the archive C boundary and C consumer recipe.'
(P/'witness-plan.json').write_text(json.dumps(w,indent=2))
