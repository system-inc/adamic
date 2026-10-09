from pathlib import Path
import subprocess,difflib
root=Path.cwd();evidence=root/'review/compiler-program-region-lowering'
def run(name,path,change,test,diagnostic):
 p=root/path;original=p.read_text();altered=change(original);assert altered!=original,name
 (evidence/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),altered.splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
 try:
  p.write_text(altered)
  package='./internal/lower' if 'ShapeKey' in test or 'EntryPairs' in test else './internal/oracle'
  with (evidence/(name+'.log')).open('w') as log:result=subprocess.run(['go','test',package,'-run','^'+test+'$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT)
  output=(evidence/(name+'.log')).read_text();assert result.returncode!=0 and diagnostic in output,(name,output)
  print(name+': caught: '+diagnostic,flush=True)
 finally:p.write_text(original)
run('acyclic-leaf','internal/lower/program_region.go',lambda s:s.replace('case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member','case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member\n\t\tfor _,field:=range value.Fields {if field.Name=="text"{value.ProgramRegion=true}}',1),'TestProgramRegionOwnership','regions 3, want 2')
run('cyclic-member','internal/lower/program_region.go',lambda s:s.replace('import (','import (\n "strings"',1).replace('case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member','case ir.ObjectLiteral:\n\t\tvalue.ProgramRegion = member\n\t\tif strings.Contains(sourceExpression(node),"root${1}"){value.ProgramRegion=false}',1),'TestProgramRegionOwnership','regions 1, want 2')
run('reuse-member','internal/native/reuse.go',lambda s:s.replace('e.line("bool %s = %s;", unique, uniquelyHeld(source))','e.line("bool %s = %s;", unique, "true")',1),'TestProgramRegionMemberMapperStorage','Program member array storage was reused')
run('field-only-size','internal/native/program_region.go',lambda s:s.replace('adamic_object_size(%s->shape->count)','sizeof(adamic_object) + %s->shape->count * sizeof(uintptr_t)'),'TestProgramRegionOptionalStorage','heap-buffer-overflow')
run('shape-key','internal/lower/program_region_shapes.go',lambda s:s.replace('return quote(all) + "|" + quote(required)','_ = quote\n return strings.Join(all,"\\x00")+"\\x01"+strings.Join(required,"\\x00")'),'TestProgramRegionShapeKeySeparators','distinct property shapes collide')
run('entry-pair-publication','internal/lower/program_region.go',lambda s:s.replace('if entryPairs && len(plan.types) != 0 {','if false && entryPairs && len(plan.types) != 0 {').replace('l.programPlan != nil && len(l.programPlan.types) != 0 && node.Kind == ast.KindCallExpression','l.programPlan != nil && false && len(l.programPlan.types) != 0 && node.Kind == ast.KindCallExpression'),'TestProgramRegionEntryPairsRefused','want unpublished-pair refusal')
