import json,pathlib,sys
repo=pathlib.Path(sys.argv[1]); out=pathlib.Path(sys.argv[2]); m=json.loads((out/'overlay.json').read_text()); r=m['Replace']
def put(name,s):
 p=out/name.replace('/','_');p.write_text(s);r[str(repo/name)]=str(p)
s=(repo/'internal/lower/lower.go').read_text(); start=s.index('\tfiles := program.Files()'); end=s.index('\n}\n',start)
s=s[:start]+'\treturn nil, fmt.Errorf("latent census: measurement only; no IR output")'+s[end:]
s=s.replace('type lowering struct {','type lowering struct {\n latentDeclarations map[int]*ast.Node\n latentReady map[int]bool\n latentAncestors map[*ast.Node]bool\n latentProgramRecordsPending bool')
put('internal/lower/lower.go',s)
s=(repo/'internal/lower/graph_types.go').read_text()
head='func (f *cycleFinder) cyclicRecord()' in (repo/'internal/lower/graph_records.go').read_text()
needle='\tif site := f.cyclicRecord(); site != nil {' if head else '\treturn flows\n'
assert s.count(needle)==1
s=s.replace(needle,'\trecordCensusFinder = f\n'+needle)
put('internal/lower/graph_types.go',s)
s=pathlib.Path(r[str(repo/'internal/lower/latent_scoped_tail.go')]).read_text();needle='func (l *lowering) latentFinishUnit(unit *LatentUnit, node *ast.Node) {'
s=s.replace(needle,needle+'\n recordCensusObserve(l, unit.Where)')
put('internal/lower/latent_scoped_tail.go',s)
common='''package lower
import (
 "context"
 "fmt"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/stage3/census/moduleformat"
)
var recordCensusFinder *cycleFinder
var recordCensusCalls, recordCensusHits int
// Prepare once with the real whole-program lowering. No partial finder is accepted.
func RecordCensusPrepare(ctx context.Context,p *load.Program) error {
 recordCensusFinder=nil; recordCensusCalls=0;recordCensusHits=0
 if len(p.Files())!=1 { return fmt.Errorf("record census facts require one program entry, got %d",len(p.Files())) }
 c,release:=p.Checker(ctx,p.Files()[0]);defer release()
 _,err:=lowerCheckedWithCycleTypes(p,c,p.Files()[0],nil,false,false,moduleformat.Resolve(ModuleFormatCommonJSBundle,ModuleFormatEcmaScript),nil)
 if recordCensusFinder==nil { return fmt.Errorf("whole-program graph facts unavailable before record check: %v",err) }
 return nil
}
func RecordCensusTotals() (int,int) {return recordCensusCalls,recordCensusHits}
'''
if head:
 common+='''func recordCensusObserve(l *lowering,unit string) {
 f:=recordCensusFinder
 if f==nil {panic("record census: no complete whole-program graph facts")}
 saved:=f.l.graphRecords;defer func(){f.l.graphRecords=saved}()
 for _,site:=range l.graphRecords {
  f.l.graphRecords=[]graphRecordSite{site}
  recordCensusCalls++
  if found:=f.cyclicRecord();found!=nil {
   recordCensusHits++
   latentFindings=append(latentFindings,LatentFinding{Measurement:LatentMeasurementLabel,Unit:unit,Phase:"record_cycle",Kind:"NotYet",Where:l.program.Where(site.node),Reason:"a record its own values can reach again, on a cycle in a graph program: a record can't join a graph region",Text:l.checker.TypeToString(site.record)})
  }
 }
}
'''
else:
 common+='''func recordCensusObserve(l *lowering,unit string) { recordCensusCalls++ /* This commit has no cyclicRecord check. */ }
'''
put('internal/lower/latent_record_census.go',common)
# Dedicated no-output driver: prepare from the entry, then use unchanged latent units.
driver='''package main
import("context";"encoding/json";"os";"fmt";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower")
func main(){
 out,e:=os.Create(os.Args[2]);if e!=nil{panic(e)};defer out.Close();enc:=json.NewEncoder(out)
 p,e:=load.LatentLoad([]string{os.Args[1]});if e!=nil{panic(e)}
 e=lower.RecordCensusPrepare(context.Background(),p)
 if e!=nil{enc.Encode(map[string]any{"status":"blocked","reason":e.Error(),"units_affected":nil,"first_sites":nil});fmt.Fprintln(os.Stderr,e);os.Exit(2)}
 lower.Latent(context.Background(),p,func(r lower.LatentFile){enc.Encode(r)})
 calls,hits:=lower.RecordCensusTotals();enc.Encode(map[string]any{"status":"complete","calls":calls,"hits":hits})
}
'''
put('stage3/census/latent/tool/main.go',driver)
(out/'overlay.json').write_text(json.dumps(m,indent=2)+'\n')
