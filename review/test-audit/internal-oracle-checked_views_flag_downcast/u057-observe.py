import pathlib,subprocess,os,json
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-oracle-checked_views_flag_downcast'
bridge='''package lower
import("fmt"; "github.com/system-inc/adamic/internal/ir")
func U057DemandObservation() string {
 receiver:=ir.Read{Local:0,Of:ir.Object}
 p:=&ir.Program{Locals:[]ir.Local{{Name:"receiver",Type:ir.Object}},ViewOrigins:[]ir.Expression{receiver},CheckedFields:map[string]bool{"field":true},ViewContractTypes:map[int]ir.ViewContractID{1:1},ViewContracts:[]ir.ViewContract{{Kind:ir.ViewUnknown,Unsupported:"dictionary",Name:"dictionary"}},Main:[]ir.Statement{ir.Evaluate{Value:ir.Property{Object:receiver,Name:"field",Of:ir.Object,ViewTypeID:1,ViewWhere:"probe:1"}}}}
 return fmt.Sprint((&lowering{result:p}).checkLazyViewReads())
}
'''
test='''package oracle
import("testing";"os";"path/filepath";"github.com/system-inc/adamic/internal/lower")
func TestU057Observation(t *testing.T){
 t.Log("demand",lower.U057DemandObservation())
 path:=filepath.Join(t.TempDir(),"literal.a")
 source:=`interface Base { readonly kind: string }
interface Child extends Base { readonly kind: "child" }
const raw: {kind:string} = {kind:"child"};
const source:Base=raw;
const view=source as Child;
raw.kind="other";
console.log(view.kind);`
 if err:=os.WriteFile(path,[]byte(source),0600);err!=nil{t.Fatal(err)}
 program,err:=lowered(t,path);if err!=nil{t.Log("lowering",err);return}
 t.Logf("node %#v",onNode(t,path));t.Logf("javascript %#v",onJavaScriptBackend(t,program))
}
'''
b=r/'internal/lower/u057_observation.go';t=r/'internal/oracle/u057_observation_test.go'
b.write_text(bridge);t.write_text(test);(e/'observation-bridge.go.txt').write_text(bridge);(e/'observation-test.go.txt').write_text(test)
menu=json.loads((e/'menu.json').read_text())
try:
 for id in ['clean','M1','M4']:
  m=next((x for x in menu if x['id']==id),None);p=r/m['file'] if m else None;old=p.read_text() if p else None
  if p:p.write_text(old.replace(m['old'],m['new']))
  try:
   env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u057/cache/observation-'+id
   with (e/('observation-'+id+'.log')).open('w') as f:subprocess.run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./internal/oracle/','-run','^TestU057Observation$'],cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
  finally:
   if p:p.write_text(old)
finally:b.unlink();t.unlink()
