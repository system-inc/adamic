import pathlib,subprocess,difflib,json
out=pathlib.Path('review/test-audit/internal-oracle-stage3_front');specs=[
('M01','internal/lower/switch.go','Operator: ir.LessOrEqual','Operator: ir.Equal','change an option'),
('M02','internal/lower/switch.go','if len(body) == 0 {\n\t\treturn false','if len(body) == 0 {\n\t\treturn true','change a constant'),
('M03','internal/lower/switch.go','dispatch.Default = []ir.Statement{unmatchedCheck}','/* dropped unmatched enum check */','drop a statement'),
('M04','internal/lower/diagnostics.go',"%s: stage 0 can't lower %s yet","%s: stage 0 can't lower %s now",'change a constant'),
('M05','internal/lower/refusals.go','"the non-null assertion !"','"a non-null assertion !"','change a constant'),
('M06','internal/lower/non_null.go','counts.Checked++','/* dropped checked counter increment */','drop a statement'),
('M07','internal/lower/assignments.go','updated := ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}','updated := ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 2}}','change a constant'),
('M08','internal/native/runtime/typed_array.c','index %.*s is outside an array of length %.*s','index %.*s is beyond an array of length %.*s','change a constant'),
('M09','internal/oracle/testdata/typed_arrays_workers_stats.a','let sum = 0;','let sum = 1;','change a constant'),
('M10','internal/oracle/testdata/typed_arrays_workers_stats.a','Math.ceil(0.95 * count)','Math.ceil(0.50 * count)','change a constant'),
('M11','internal/lower/diagnostics.go','%s: Adamic 0.1 refuses %s; %s','%s: Adamic 0.2 refuses %s; %s','change a constant'),
('M12','internal/native/runtime/typed_array.c','return (double)array->length;','return (double)array->length + 1;','off-by-one a bound'),
('M13','internal/lower/switch_bindings.go','Value: ir.BooleanConstant{Value: false}','Value: ir.BooleanConstant{Value: true}','change a constant')]
files=set(s[1] for s in specs)|{'internal/oracle/oracle_test.go','internal/lower/lower.go'};base={f:subprocess.check_output(['git','show','HEAD:'+f],text=True) for f in files};sw=dict(base);plan=[]
def diff(mid,file,text):
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(base[file].splitlines(True),text.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
for mid,f,old,new,menu in specs:
 count=base[f].count(old);assert count==1 or (mid=='M09' and count==2),(mid,count);loc=base[f].index(old);line=base[f][:loc].count('\n')+1;diff(mid,f,base[f].replace(old,new,1));plan.append(dict(id=mid,file=f,line=line,old=old,new=new,menu=menu,kind='port' if f.endswith('.a') else 'production'))
 if mid=='M01':replace='Operator: auditSwitchOperator("M01", ir.LessOrEqual, ir.Equal)'
 elif mid=='M02':replace='if len(body) == 0 {\n\t\treturn auditMutant("M02")'
 elif mid in ['M03','M06']:replace='if !auditMutant("'+mid+'") { '+old+' }'
 elif mid in ['M04','M11']:replace='" + auditString("'+mid+'", "'+old+'", "'+new+'") + "'
 elif mid=='M05':replace='auditString("M05", '+old+', '+new+')'
 elif mid=='M07':replace='updated := ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: auditNumber("M07", 1, 2)}}'
 elif mid=='M08':replace='" + INVALID + "' # replaced separately below, C literal needs expression
 elif mid=='M09':replace='let sum = auditSelector === "M09" ? 1 : 0;'
 elif mid=='M10':replace='Math.ceil((auditSelector === "M10" ? 0.50 : 0.95) * count)'
 elif mid=='M12':replace='return (double)array->length + (audit_c_mutant("M12") ? 1 : 0);'
 elif mid=='M13':replace='Value: ir.BooleanConstant{Value: auditMutant("M13")}'
 if mid in ['M04','M11']:
  sw[f]=sw[f].replace('"'+old+'"','auditString("'+mid+'", "'+old+'", "'+new+'")',1)
 elif mid=='M08':sw[f]=sw[f].replace('"'+old+'"','(audit_c_mutant("M08") ? "'+new+'" : "'+old+'")',1)
 else:sw[f]=sw[f].replace(old,replace,1)
# Witness weakening: disable the complete guarded comparison.
f='internal/oracle/oracle_test.go';start=base[f].index('func disagreement(');end=base[f].index('\n}\n',start)+3;empty=base[f][:start]+'func disagreement(oracle run, native run) string { return "" }\n'+base[f][end:];diff('W01',f,empty);diff('P02',f,empty);sw[f]=sw[f].replace('func disagreement(oracle run, native run) string {','func disagreement(oracle run, native run) string {\n\tif os.Getenv("ADAMIC_MUTANT") == "W01" || os.Getenv("ADAMIC_MUTANT") == "P02" { return "" }');plan.append(dict(id='W01',file=f,line=716,old='disagreement body',new='return ""',kind='witness-check'))
# Lower nil-answer probe: drop its body in the standalone form to avoid unreachable code.
f='internal/lower/lower.go';start=base[f].index('func Lower(');end=base[f].index('\n}\n',start)+3;empty=base[f][:start]+'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return nil, nil }\n'+base[f][end:]
# Removing Lower orchestration also makes fmt and filepath imports unused.
empty=empty.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','');diff('P01',f,empty);sw[f]=sw[f].replace('func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\tif auditMutant("P01") { return nil, nil }')
f='internal/oracle/testdata/typed_arrays_workers_stats.a';entries=[('P03','function stats(values: Float64Array): StatsInterface {','return {count: 0, sum: 0, minimum: 0, maximum: 0};'),('P04','function summarize(values: Float64Array): StatsBody {','return {count: 0, mean: 0, median: 0, p95: 0, min: 0, max: 0, standardDeviation: 0};')];probes=[dict(id='P01',file='internal/lower/lower.go',line=20,entry='Lower'),dict(id='P02',file='internal/oracle/oracle_test.go',line=716,entry='disagreement')]
for mid,signature,ret in entries:
 line=base[f][:base[f].index(signature)].count('\n')+1;diff(mid,f,base[f].replace(signature,signature+'\n    '+ret));sw[f]=sw[f].replace(signature,signature+'\n    if (auditSelector === "'+mid+'") { '+ret+' }');probes.append(dict(id=mid,file=f,line=line,entry=signature))
sw[f]=sw[f].replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';\nconst auditFile = readTextFile('/tmp/u070-selector');\nconst auditSelector = auditFile.kind === 'Ok' ? auditFile.text : '';")
f='internal/native/runtime/typed_array.c';sw[f]=sw[f].replace('#include <string.h>','#include <string.h>\nstatic bool audit_c_mutant(const char *id) { const char *value = getenv("ADAMIC_MUTANT"); return value != NULL && strcmp(value, id) == 0; }')
(out/'plan.json').write_text(json.dumps(dict(origin_commit=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),mutants=plan,probes=probes),indent=2))
for f,text in sw.items():pathlib.Path(f).write_text(text)
pathlib.Path('internal/lower/audit_mutant.go').write_text('package lower\nimport("os";"github.com/system-inc/adamic/internal/ir")\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc auditString(id, before, after string) string { if auditMutant(id) { return after };return before }\nfunc auditNumber(id string, before, after float64) float64 { if auditMutant(id) { return after };return before }\nfunc auditSwitchOperator(id string, before, after ir.Operator) ir.Operator { if auditMutant(id) { return after };return before }\n')
pathlib.Path('/tmp/u070-selector').write_text('')
