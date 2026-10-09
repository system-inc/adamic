import pathlib,json,subprocess,os,time
p=pathlib.Path('review/test-audit/internal-native-arguments_length'); menu=json.loads((p/'menu.json').read_text())
probe='''package native
import("context";"fmt";"testing";"os/exec";"path/filepath";"github.com/system-inc/adamic/internal/ir";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower")
func TestAuditSurvivorObservation(t *testing.T) {
 t.Run("captured-root",func(t *testing.T){
 p:= &ir.Program{Locals:[]ir.Local{{Name:"root",Type:ir.Object,Function:0,Captured:true},{Name:"field",Type:ir.String,Function:0}}, Functions:[]ir.Function{{Name:"inspect",Body:[]ir.Statement{ir.Declare{Local:1,Value:ir.Property{Object:ir.Read{Local:0,Of:ir.Object},Name:"kind",Of:ir.String}}}}}}
 plan,_:=planElementBorrows(p);fmt.Printf("captured-root borrows=%d\\n",len(plan))
 })
 t.Run("region-eligibility",func(t *testing.T){
 loaded,e:=load.Load([]string{"../oracle/testdata/class_inheritance_memory.a"});if e!=nil{t.Fatal(e)};p,e:=lower.Lower(context.Background(),loaded);if e!=nil{t.Fatal(e)}
 plan:=planRegions(p);wrong:=0;for s:=range plan.statements {if !plan.feedsRegion(*s){wrong++}};fmt.Printf("region marked=%d ineligible-marked=%d\\n",len(plan.statements),wrong)
 })
 t.Run("actual-count",func(t *testing.T){
 loaded,e:=load.Load([]string{"../oracle/testdata/arguments_length_value_count.a"});if e!=nil{t.Fatal(e)};p,e:=lower.Lower(context.Background(),loaded);if e!=nil{t.Fatal(e)}
 binary:=filepath.Join(t.TempDir(),"program");if e=Build(C(p),binary,Options{});e!=nil{t.Fatal(e)};out,e:=exec.Command(binary).CombinedOutput();if e!=nil{t.Fatal(e)};fmt.Printf("actual-count stdout=%q\\n",out)
 })
}
'''
(p/'survivor-probe.go.txt').write_text(probe); f=pathlib.Path('internal/native/audit_observation_test.go');f.write_text(probe)
try:
 for id in ['BASE','M06','M08','M11']:
  env=os.environ.copy();env.pop('ADAMIC_MUTANT',None);env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u045/witness-cache/'+id
  m=next((x for x in menu if x['id']==id),None);base=None
  if m:
   file=pathlib.Path(m['file']);base=file.read_text();assert m['old'] in base;file.write_text(base.replace(m['old'],m['new']))
  with open(p/('survivor-'+id+'.log'),'w') as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestAuditSurvivorObservation$'],stdout=out,stderr=subprocess.STDOUT,env=env)
  print(id,r.returncode,flush=True)
  if m:file.write_text(base)
finally:f.unlink()
