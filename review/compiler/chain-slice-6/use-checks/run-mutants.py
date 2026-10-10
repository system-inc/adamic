import json,os,subprocess,time
from pathlib import Path
root=Path.cwd(); out=root/'review/compiler/chain-slice-6/use-checks/mutants';out.mkdir(exist_ok=True)
js='internal/javascript/javascript.go'; c='internal/native/runtime/exceptions.c'
source=(root/js).read_text(); write=next(l for l in source.splitlines() if 'adamicUncaughtWrite(2,' in l)
mutants=[
 ('backend-renders-stack',js,'builder.WriteString(uncaughtRuntime)','// mutant: use Node uncaught stack','./internal/oracle','^TestStep21Uncaught$','uncaught diagnostic'),
 ('backend-drops-line',js,write,'\tvoid line;','./internal/oracle','^TestStep21Uncaught$','uncaught diagnostic'),
 ('native-renders-stack',c,"\tfflush(stderr);",'\tfputs("frame\\n", stderr);\n\tfflush(stderr);','./internal/oracle','^TestStep21Uncaught$','uncaught diagnostic'),
 ('native-drops-line',c,'\tfputs("Uncaught ", stderr);','\tif (false) {\n\tfputs("Uncaught ", stderr);','./internal/oracle','^TestStep21Uncaught$','uncaught diagnostic'),
 ('drop-write-set-test','internal/lower/narrowing_calls.go','if len(writes) != 0 {','if true {','./internal/lower','^TestStep21NonWritingCallControl$','narrow again after the call'),
 ('admit-writing-call','internal/lower/refusals.go','if strings.HasSuffix(l.program.FileName(module), ".a") {','if false {','./internal/lower','^TestStep21WritingCallRefusal$','writing-call refusal: got <nil>')]
mutants += [
 ('whole-narrowed-type','internal/lower/locals.go','if l.narrowingPropertyReceiver(node) {','if false {','./internal/oracle','^TestStep21WritingCallTypeScriptAgreement$','.name agreement'),
 ('native-string-length','internal/native/runtime/union.c','return strcmp(name, "length") == 0 ? adamic_box_number(adamic_string_length((const adamic_string *)value)) : NULL;','return NULL;','./internal/oracle','^TestStep21StoredPrimitiveLengthUseAgrees$','primitive length'),
 ('native-scalar-properties','internal/native/runtime/union.c','if (value->kind == adamic_kind_number || value->kind == adamic_kind_boolean) return NULL;','// mutant: omit scalar absence','./internal/oracle','^TestNarrowedUnionMemberCheck$','stale narrowing use failed'),
 ('numeric-enum-as-finite','internal/lower/narrowing_uses.go',' && !l.openNumericEnumType(l.checker.GetTypeAtLocation(node))','', './internal/oracle','^TestStep21NarrowedFlagsUseAgrees$','stored flags')]
results=json.loads((out/"results.json").read_text()) if (out/"results.json").exists() else []
for name,relative,old,new,package,selection,catcher in mutants:
 if any(r["name"]==name and r["caught"] for r in results):continue
 directory=out/name;directory.mkdir(exist_ok=True);path=root/relative; original=path.read_text();assert original.count(old)==1,(name,old)
 changed=original.replace(old,new,1)
 if name=='native-drops-line':changed=changed.replace("\tfputc('\\n', stderr);","\tfputc('\\n', stderr);\n\t}",1)
 target=directory/(path.name+'.txt');target.write_text(changed)
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(target)}},indent=2)+'\n')
 command=['go','test','-overlay',str(overlay),package,'-run',selection,'-count=1','-v','-timeout','90s'];start=time.monotonic()
 with (directory/'test.log').open('w') as log: result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=110,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
 text=(directory/'test.log').read_text(); caught=result.returncode!=0 and catcher in text and '[build failed]' not in text and 'clang failed' not in text
 results.append(dict(name=name,caught=caught,exit=result.returncode,catcher=catcher,command=command,seconds=round(time.monotonic()-start,3)))
 (out/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(name,caught,flush=True)
 assert caught,name
