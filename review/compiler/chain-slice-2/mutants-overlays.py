import json,subprocess,os,signal,time,importlib.util
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-2'
known='internal/lower/known_lowering_gaps.go';runtime='internal/native/runtime/exceptions.c'
mutants=[]
for name,before,after,pattern in [
 ('error-spread','case ast.KindSpreadAssignment:','case ast.KindUnknown:','(ErrorSpread|ErrorSubclassSpread)'),
 ('maybe-setter','case ast.KindSetAccessor:','case ast.KindUnknown:','MaybeSetter'),
 ('long-name','len(plainPropertyName(node)) > 4095','len(plainPropertyName(node)) > 1000000','LongName'),
 ('computed-name','case ast.KindClassDeclaration:','case ast.KindUnknown:','(DerivedSymbol|OverrideSource)'),
 ('constructor-capture','case ast.KindThisKeyword:','case ast.KindUnknown:','(ConstructorArrow|ConstructorNumberArrow)')]:
 source=(root/known).read_text();assert source.count(before)==1
 mutants.append((name,{known:source.replace(before,after)},'./internal/oracle','^TestMiscompile2A'+pattern+'$',['want path-bearing stop','panic: Unhandled case in Node.Text']))
source=(root/runtime).read_text();before='adamic_retain(message == NULL ? &empty_message : message)';assert source.count(before)==1
mutants.append(('optional-error',{runtime:source.replace(before,'adamic_retain(message)').replace('static adamic_string empty_message = ADAMIC_STRING("");\n','')},'./internal/oracle','^TestMiscompile2A(OptionalError|MarkerUndefined)$',['member access within null pointer']))
before='adamic_object_new(message == NULL ? &absent_message_shape : &error_shape)';assert source.count(before)==1
mutants.append(('error-own-message',{runtime:source.replace(before,'adamic_object_new(&error_shape)')},'./internal/native','^TestErrorUndefinedMessageHasNoOwnProperty$',['native ','Node ']))
spec=importlib.util.spec_from_file_location('weak',out/'mutants-2b.py');weak=importlib.util.module_from_spec(spec);spec.loader.exec_module(weak)
for name,path,before,after,pattern,markers in weak.MUTANTS:
 source=path.read_text();assert source.count(before)==1
 mutants.append((name,{str(path.relative_to(root)):source.replace(before,after)},'./internal/oracle',pattern,markers))
results=[]
for name,files,package,pattern,markers in mutants:
 directory=out/('overlay-'+name);directory.mkdir(exist_ok=True)
 mapping={}
 for index,(path,source) in enumerate(files.items()):
  target=directory/(str(index)+'.go.txt');target.write_text(source);mapping[str(root/path)]=str(target)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':mapping}))
 command=['go','test',package,'-overlay='+str(overlay),'-run',pattern,'-count=1','-v','-timeout=90s']
 log=directory/'test.log';start=time.monotonic()
 with log.open('w') as f:
  process=subprocess.Popen(command,stdout=f,stderr=subprocess.STDOUT,start_new_session=True,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',UBSAN_OPTIONS='symbolize=0'))
  try:code=process.wait(timeout=110)
  except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);code=process.wait()
 text=log.read_text();matched=[x for x in markers if x in text]
 caught=code!=0 and code!=-9 and matched and '[build failed]' not in text and 'error: unused' not in text
 if name in ['method-views','union-liveness','weak-callback','error-own-message']:caught=caught and len(matched)==len(markers)
 row={'mutant':name,'exit':code,'caught':bool(caught),'seconds':round(time.monotonic()-start,3),'markers':matched,'command':command}
 results.append(row);(out/'mutants-overlay-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True)
 assert caught,(name,str(log))
