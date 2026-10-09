from pathlib import Path
import json,subprocess,difflib
p=Path('review/test-audit/internal-oracle-wasi')
plan=[
('M01','internal/native/native.go','index < len(value)','index < len(value)-1','off-by-one bound','index < len(value)-auditDelta("M01")'),
('M02','internal/native/emit_values.go','return "NAN"','return "0.0"','change constant','return auditText("M02", "NAN", "0.0")'),
('M03','internal/native/emit_values.go','return "HUGE_VAL"','return "0.0"','change constant','return auditText("M03", "HUGE_VAL", "0.0")'),
('M04','internal/native/emit_values.go','return "(-HUGE_VAL)"','return "0.0"','change constant','return auditText("M04", "(-HUGE_VAL)", "0.0")'),
('M05','internal/native/emit_expressions.go','if expression.Present {','if !expression.Present {','flip condition','if expression.Present != auditSelected("M05") {'),
('M06','internal/native/emit_values.go',"strconv.FormatFloat(value, 'x', -1, 64)","strconv.FormatFloat(value, 'x', 0, 64)",'change constant',"strconv.FormatFloat(value, 'x', -1+auditDelta(\"M06\"), 64)"),
('M07','internal/native/native.go','flags = append(flags, "-ffp-contract=off")','flags = append(flags, "-ffp-contract=fast")','change option','flags = append(flags, auditText("M07", "-ffp-contract=off", "-ffp-contract=fast"))'),
('M08','internal/native/native.go','flags = append(flags, "-fno-optimize-sibling-calls")','flags = append(flags, "-foptimize-sibling-calls")','change option','flags = append(flags, auditText("M08", "-fno-optimize-sibling-calls", "-foptimize-sibling-calls"))'),
('M09','internal/native/target.go','return append(flags, "-mexec-model=command")','return append(flags, "-mexec-model=reactor")','change option','return append(flags, auditText("M09", "-mexec-model=command", "-mexec-model=reactor"))')]
base={f:subprocess.check_output(['git','show','HEAD:'+f],text=True) for _,f,*_ in plan}
base['internal/native/emit.go']=subprocess.check_output(['git','show','HEAD:internal/native/emit.go'],text=True)
metadata=[]
for mid,f,old,new,kind,switch in plan:
 assert base[f].count(old)==1,(mid,base[f].count(old))
 mutant=base[f].replace(old,new)
 (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(base[f].splitlines(True),mutant.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 metadata.append(dict(id=mid,file=f,line=base[f][:base[f].index(old)].count('\n')+1,before=old,after=new,menu=kind))
(p/'plan.json').write_text(json.dumps(metadata,indent=2))
# Save the plan and standalone diffs before any mutant execution.
for f,src in base.items():
 result=src
 for mid,mf,old,new,kind,switch in plan:
  if mf==f:result=result.replace(old,switch)
 if f=='internal/native/emit.go':result=result.replace('return cProgram(program, -1)','if auditSelected("P01") { return "" }; return cProgram(program, -1)')
 Path(f).write_text(result)
Path('internal/native/audit_selector.go').write_text('package native\nimport "os"\nfunc auditSelected(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditDelta(id string) int { if auditSelected(id) { return 1 }; return 0 }\nfunc auditText(id, original, changed string) string { if auditSelected(id) { return changed }; return original }\n')
old=base['internal/native/emit.go'];new=old.replace('return cProgram(program, -1)','return ""')
(p/'P01.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/internal/native/emit.go',tofile='b/internal/native/emit.go')))
