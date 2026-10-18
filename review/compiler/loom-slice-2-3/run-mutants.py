import json, os, subprocess, time
from pathlib import Path
root=Path.cwd();out=Path(__file__).parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
env=dict(os.environ,GOMAXPROCS="4",ADAMIC_GATE_UNCACHED="1",UBSAN_OPTIONS="symbolize=0")
cases=[]
def add(name,path,before,after,pkg,pattern,markers):
 source=Path(path).read_text();assert source.count(before)==1,(name,source.count(before))
 cases.append((name,{path:source.replace(before,after)},pkg,pattern,markers))
known="internal/lower/known_lowering_gaps.go"
for name,before,after,pattern in [
 ("error-spread","case ast.KindSpreadAssignment:","case ast.KindUnknown:","(ErrorSpread|ErrorSubclassSpread)"),
 ("maybe-setter","case ast.KindSetAccessor:","case ast.KindUnknown:","MaybeSetter"),
 ("long-name","len(plainPropertyName(node)) > 4095","len(plainPropertyName(node)) > 1000000","LongName"),
 ("computed-name","case ast.KindClassDeclaration:","case ast.KindUnknown:","(DerivedSymbol|OverrideSource)"),
 ("constructor-capture","case ast.KindThisKeyword:","case ast.KindUnknown:","(ConstructorArrow|ConstructorNumberArrow)")]:
 add(name,known,before,after,"./internal/oracle","^TestMiscompile2A"+pattern+"$",["want path-bearing stop"] if name!="computed-name" else ["Unhandled case in Node.Text"])
runtime="internal/native/runtime/exceptions.c"
source=Path(runtime).read_text()
cases.append(("optional-error",{runtime:source.replace("adamic_retain(message == NULL ? &empty_message : message)","adamic_retain(message)").replace('static adamic_string empty_message = ADAMIC_STRING("");\n',"")},"./internal/oracle","^TestMiscompile2A(OptionalError|MarkerUndefined)$",["member access within null pointer"]))
add("error-own-message",runtime,"adamic_object_new(message == NULL ? &absent_message_shape : &error_shape)","adamic_object_new(&error_shape)","./internal/native","^TestErrorUndefinedMessageHasNoOwnProperty$",["native ","Node "])
import importlib.util
spec=importlib.util.spec_from_file_location("weak",root/"review/compiler/chain-slice-2/mutants-2b.py");weak=importlib.util.module_from_spec(spec);spec.loader.exec_module(weak)
for name,path,before,after,pattern,markers in weak.MUTANTS:
 add(name,str(path.relative_to(root)),before,after,"./internal/oracle",pattern,["both backends must stop at"] if name=="weak-callback" else markers)
adapter="internal/lower/callback_widening.go"
add("adaptation-omitted",adapter,"if !changed {","if true || !changed {","./internal/oracle","^TestCallbackWideningNumberEvery(Release|Sanitized)$",["native"])
add("nullable-token-omitted",adapter,"if nullPointer {","if false && nullPointer {","./internal/oracle","^TestCallbackWideningNullableObjectMap(Release|Sanitized)$",["exit"])
add("unrepresented-parameter-admitted",adapter,"if fit(ir.Read{Of: given}, of).Type() != of {","if false && fit(ir.Read{Of: given}, of).Type() != of {","./internal/oracle","^TestCallbackWideningUnrepresentedParameter$",["expected unrepresented parameter refusal"])
add("adapter-target-omitted","internal/ir/call_targets.go","value = effects.Result","_ = effects; break","./internal/ir","^TestCallbackAdapterTargets$",["Unknown:true"])
add("split-numeric-fold","internal/native/self_compare.go",'comparison := fmt.Sprintf("%s(%s, %s)", helper, left, right)','comparison := fmt.Sprintf("%s(%s, %s)", helper, left, right)\n if operandType == ir.Number { comparison = fmt.Sprintf("((void)(%s), (void)(%s), true)", left, right) }',"./internal/native","^TestSplitSelfCompareAgreesWithNode$",["native: <nil>","Node "])
recorded=Path("review/compiler/loom-slice-2-2/callback-originals")
mapping=json.loads((recorded/"mapping.json").read_text())
cases.append(("refusal-revert",{path:(recorded/name).read_text() for path,name in mapping.items()},"./internal/oracle","^TestCallbackWideningRefuse",["--- FAIL:"]))
source=Path("internal/lower/weak_representations.go").read_text()
before='return l.notYet(node, callee.Name().Text()+" callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy")'
cases.append(("weak-callback-isolated",{"internal/lower/weak_representations.go":source.replace(before,"return nil"),"internal/lower/object.go":subprocess.check_output(["git","show","origin/main:internal/lower/object.go"],text=True)},"./internal/oracle","^TestWeakDifferential(Each|Find|Map|Reduce|Some|Sort)$",["both backends unexpectedly admit the pinned callback NotYet","AddressSanitizer: heap-buffer-overflow"]))
add("weak-counter-registration-omitted","internal/oracle/weak_differential_test.go",'[]string{"4ddd17f_weak_union_narrowed", "4ddd17f_weak_single_narrowed"}','[]string{"4ddd17f_weak_single_narrowed"}',"./internal/oracle","^TestCountsAreRecordedTable$",["want the row of"])
for name,files,pkg,pattern,markers in cases:
 folder=out/"mutants"/name;folder.mkdir(parents=True,exist_ok=True);mapping={}
 for i,(path,source) in enumerate(files.items()):
  target=(folder/(str(i)+".go.txt")).resolve();target.write_text(source);mapping[str(root/path)]=str(target)
 overlay=(folder/"overlay.json").resolve();overlay.write_text(json.dumps({"Replace":mapping}))
 command=["timeout","240","go","test",pkg,"-overlay="+str(overlay),"-run",pattern,"-count=1","-v","-timeout=90s","-parallel=4"]
 start=time.monotonic()
 with (folder/"test.log").open("w") as log: ran=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=250)
 text=(folder/"test.log").read_text();caught=ran.returncode==1 and "--- FAIL:" in text and all(x in text for x in markers) and not any(x in text for x in ["[build failed]","test timed out","error: unused","error: variable"])
 row=dict(mutant=name,caught=caught,exit=ran.returncode,seconds=round(time.monotonic()-start,3),command=command,markers=markers)
 with (out/"mutants.jsonl").open("a") as f:f.write(json.dumps(row)+"\n")
 print(json.dumps(row),flush=True)
 if not caught:raise SystemExit(1)
