"""Copy Memory's probe into the guarded scratch overlay, adding measurement context."""
import json,subprocess
from pathlib import Path
root=Path('/workspace/view-pairs');scratch=Path('/workspace/scratch/vqzty8t');here=Path(__file__).resolve().parent;overlay=scratch/'overlay'
original=subprocess.check_output(['git','show','86157bea:internal/lower/zz_view_pairs_probe.go'],cwd=root,text=True)
(here/'memory-probe-original.go.txt').write_text(original)
s=original.replace('"sync"','"sync"\n "fmt"\n "github.com/microsoft/TypeScript/tsc/shim/ast"\n "github.com/system-inc/adamic/internal/ir"')
s=s.replace('type viewPairRecord struct {','''type viewPairRecord struct {
 Unit string `json:"unit"`
 File string `json:"file"`
 Path []string `json:"path"`
 SourceIntersection bool `json:"source_intersection"`
 ObjectPair bool `json:"object_pair"`
 TargetIntersection bool `json:"target_intersection"`
 IntersectionRelated bool `json:"intersection_related"`
 SourceNil bool `json:"source_nil"`
 TargetNil bool `json:"target_nil"`
 TargetReferenceFields []string `json:"target_reference_fields,omitempty"`
 TargetObjectReferenceFields []string `json:"target_object_reference_fields,omitempty"`
 TargetUnknownFields []string `json:"target_unknown_fields,omitempty"`
''')
s=s.replace('var pair func(source, target *checker.Type, field string, depth int, cast, narrows bool)','var pair func(source, target *checker.Type, field string, path []string, depth int, cast, narrows bool)')
s=s.replace('pair = func(source, target *checker.Type, field string, depth int, cast, narrows bool)', 'pair = func(source, target *checker.Type, field string, path []string, depth int, cast, narrows bool)')
s=s.replace('records = append(records, viewPairRecord{Program: name, Depth: depth, Field: field, Cast: cast, Narrows: narrows, Skipped: "no source"})','''record:=viewPairRecord{Unit:viewPairsUnit,File:viewPairsFile,Path:path,Program:name,Depth:depth,Field:field,Cast:cast,Narrows:narrows,Skipped:"no source",SourceNil:source==nil,TargetNil:target==nil}
   if source!=nil { record.Source=f.l.checker.TypeToString(f.l.concrete(source)) }
   if target!=nil {
    target=f.l.concrete(target);record.Target=f.l.checker.TypeToString(target)
    record.TargetIntersection=target.Flags()&checker.TypeFlagsIntersection!=0
    for _,field:=range f.fields(f.l.checker.GetNonNullableType(target)) {
     held,known:=viewPairsRepresentation(f.l,f.l.checker.GetTypeOfSymbol(field))
     if !known {record.TargetUnknownFields=append(record.TargetUnknownFields,field.Name);continue}
     if held.IsReference(){ record.TargetReferenceFields=append(record.TargetReferenceFields,field.Name) }
     if held.IsReference()&&held!=ir.String&&held!=ir.Weak {record.TargetObjectReferenceFields=append(record.TargetObjectReferenceFields,field.Name)}
    }
   }
   records=append(records,record)''')
s=s.replace('if seen[key] || depth > 6 {\n\t\t\treturn\n\t\t}', 'if seen[key] {viewPairsDedup++;return};if depth>6{viewPairsDepth++;return}')
s=s.replace('record := viewPairRecord{Program: name,', 'record := viewPairRecord{Unit:viewPairsUnit,File:viewPairsFile,Path:path,SourceIntersection:source.Flags()&checker.TypeFlagsIntersection!=0,ObjectPair:target.Flags()&checker.TypeFlagsIntersection!=0 || source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection)!=0 && target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection)!=0,TargetIntersection:target.Flags()&checker.TypeFlagsIntersection!=0,Program: name,')
s=s.replace('switch {\n\t\tcase source.Flags()', 'if record.TargetIntersection || record.SourceIntersection {record.IntersectionRelated=f.deriveRelated(source,target)}\n  switch {\n\t\tcase source.Flags()')
s=s.replace('property.Name, depth+1, cast, narrows)', 'property.Name,append(append([]string{},path...),property.Name),depth+1,cast,narrows)')
s=s.replace('pair(view.source, view.target, "", 0, view.cast, view.narrows)', 'pair(view.source,view.target,"",nil,0,view.cast,view.narrows)')
s=s.replace('if err != nil {\n\t\treturn\n\t}', 'if err != nil {panic(err)}')
s=s.replace('_ = encoder.Encode(record)', 'if err:=encoder.Encode(record);err!=nil{panic(err)};viewPairsLogged++')
s+='''
var viewPairsUnit,viewPairsFile string
var viewPairsLogged,viewPairsDedup,viewPairsDepth int
func viewPairsRepresentation(l *lowering,t *checker.Type)(held ir.Type,known bool){
 defer func(){if recover()!=nil{held=0;known=false}}()
 return l.representation(t)
}
func (l *lowering) latentViewPairs(unit,file string,full,stopped bool){
 if os.Getenv("ADAMIC_VIEW_PAIRS_LOG")==""{return}
 viewPairsUnit,viewPairsFile=unit,file;viewPairsLogged,viewPairsDedup,viewPairsDepth=0,0,0
 initialViews:=len(l.viewSources);failure:=""
 func(){defer func(){if x:=recover();x!=nil{failure=fmt.Sprint(x)}}();f:=&cycleFinder{l:l,where:map[*checker.Type]*ast.Node{}};f.logViewPairs()}()
 if len(l.viewSources)!=initialViews{failure="probe query changed viewSources"}
 out,err:=os.OpenFile(os.Getenv("ADAMIC_VIEW_PAIRS_UNITS"),os.O_APPEND|os.O_CREATE|os.O_WRONLY,0600);if err!=nil{panic(err)}
 status:="lowered";if stopped{status="stopped"}
 row:=map[string]any{"unit":unit,"file":file,"full_mode":full,"status":status,"view_sources":initialViews,"pairs_logged":viewPairsLogged,"dedup_skipped":viewPairsDedup,"depth_skipped":viewPairsDepth,"probe_error":failure,"graph_types_count":len(l.result.GraphTypes)}
 if err=json.NewEncoder(out).Encode(row);err!=nil{panic(err)};if err=out.Close();err!=nil{panic(err)}
}
'''
p=overlay/'zz_view_pairs_probe.go';p.write_text(s)
manifest=overlay/'overlay.json';data=json.loads(manifest.read_text());data['Replace'][str(root/'internal/lower/zz_view_pairs_probe.go')]=str(p)
# The existing observer has already captured its IR row before this query runs.
count=Path(data['Replace'][str(root/'internal/lower/graph_boundary_census.go')]);t=(here/'counting-hook.go.txt').read_text();needle='if closeErr!=nil{panic(closeErr)}';assert t.count(needle)==1;t=t.replace(needle,needle+'\n l.latentViewPairs(unit,file,full,stopped)');count.write_text(t)
manifest.write_text(json.dumps(data,indent=2)+'\n');(here/'memory-probe-context.go.txt').write_text(s)
print('probe prepared; no production source modified')
