"""Export upstream core tests through overlays of the four owned rule test files."""
import json,pathlib,subprocess,sys,os
repo=pathlib.Path(sys.argv[1]).resolve();out=pathlib.Path(sys.argv[2]).resolve();out.mkdir(parents=True,exist_ok=True)
rules=repo/'cohere/internal/lint/rules/core';replace={}
for name in ['no_obj_calls_test.go','no_obj_calls_corpus_test.go','no_object_constructor_test.go','no_promise_executor_return_test.go']:
 original=rules/name;target=out/name;text=original.read_text().replace('rule_testing.RunTypedWithOptions(', 'wave13MoreCaptureOptions(').replace('rule_testing.RunTyped(', 'wave13MoreCapture(');
 if 'rule_testing.' not in text: text=text.replace('\t"github.com/system-inc/cohere/internal/lint/testing"\n','')
 target.write_text(text);replace[str(original)]=str(target)
helper=out/'wave13_more_capture_test.go'
helper.write_text(r'''package core
import("crypto/sha256";"encoding/hex";"encoding/json";"os";"path/filepath";"strings";"testing";"github.com/system-inc/cohere/internal/lint/rule";rule_testing "github.com/system-inc/cohere/internal/lint/testing")
func wave13MoreSave(t *testing.T,subject rule.Rule,name,text string,options any){
 encoded,_:=json.Marshal(options);sum:=sha256.Sum256([]byte(t.Name()+"\n"+name+"\n"+text+"\n"+string(encoded)));directory:=filepath.Join(os.Getenv("WAVE13_MORE_CAPTURE"),hex.EncodeToString(sum[:8]));if err:=os.MkdirAll(directory,0755);err!=nil{t.Fatal(err)}
 path:=filepath.Join(directory,strings.TrimPrefix(name,"/"));if err:=os.MkdirAll(filepath.Dir(path),0755);err!=nil{t.Fatal(err)};if err:=os.WriteFile(path,[]byte(rule_testing.FixtureText(text)),0644);err!=nil{t.Fatal(err)}
 config:=`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"moduleDetection":"force","types":[]},"include":["**/*.ts","**/*.tsx"]}`
 os.WriteFile(filepath.Join(directory,"tsconfig.json"),[]byte(config),0644);os.WriteFile(filepath.Join(directory,"roots.manifest"),[]byte(path+"\n"),0644)
 info,_:=json.Marshal(map[string]any{"test":t.Name(),"rule":subject.Name,"options":options,"subject":name});os.WriteFile(filepath.Join(directory,"case.json"),info,0644)
}
func wave13MoreCapture(t *testing.T,subject rule.Rule,name,text string)rule_testing.Result{wave13MoreSave(t,subject,name,text,nil);return rule_testing.RunTyped(t,subject,name,text)}
func wave13MoreCaptureOptions(t *testing.T,subject rule.Rule,name,text string,options any)rule_testing.Result{wave13MoreSave(t,subject,name,text,options);return rule_testing.RunTypedWithOptions(t,subject,name,text,options)}
''')
replace[str(rules/helper.name)]=str(helper);overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
with (out/'export.log').open('wb') as log:
 result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lint/rules/core','-run','^Test(NoObjCalls|NoObjectConstructor|NoPromiseExecutorReturn|DecodeNoPromiseExecutorReturnOptions)','-count=1','-v','-timeout=15m'],cwd=repo/'cohere',env=dict(os.environ,WAVE13_MORE_CAPTURE=str(out/'controls')),stdout=log,stderr=subprocess.STDOUT)
if result.returncode:raise SystemExit(result.returncode)
print('exported',len(list((out/'controls').glob('*/case.json'))),'upstream controls')
