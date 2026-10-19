import pathlib,subprocess,json,difflib,ast
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-cycle-targets-ac-m2'
tree=ast.parse((root/'review/compiler/loom-r-cycle-targets-ac-m/run-mutants.py').read_text())
specs=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='mutants' for t in n.targets))
specs += [
 ('unbound-callable-array-is-empty','internal/ir/call_targets.go','return local.ArrayElement == 0 || local.ArrayElement == Closure || local.ArrayElement == Union','return false','./internal/ir','^TestClosureTargetsKeepCallableArrayInputsUnknown$','lost unknown input'),
 ('omit-array-element-proof','internal/lower/locals.go','l.result.Locals[local].ArrayElement, _ = l.representation(element)','l.result.Locals[local].ArrayElement = 0','./internal/fuzz','^TestGeneratedProgramsCheckAndLower$/^seeds-001-005$','stage 0 didn\'t lower it'),
 ('omit-adapter-element-proof','internal/lower/callback_widening.go','l.result.Locals[local].ArrayElement = supplied[position-2]','l.result.Locals[local].ArrayElement = 0','./internal/fuzz','^TestGeneratedProgramsCheckAndLower$/^seeds-031-035$','stage 0 didn\'t lower it'),
 ('omit-scalar-operation-proof','internal/ir/call_targets.go','case RegExpCall:\n','case RegExpCall:\n targets := f.contents["array"]; targets.Unknown = true; f.contents["array"] = targets\n','./internal/ir','^TestClosureTargetsIgnoreNonCallableArrayInputs$','contaminated closure slots'),
]
results=[]
def run(name,relative,original,mutated,package,selector,catcher):
 target=out/(name+'.go.txt');target.write_text(mutated)
 (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=relative,tofile=relative)))
 overlay=out/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/relative):str(target)}}))
 command=['timeout','150','go','test','-overlay='+str(overlay),package,'-run',selector,'-count=1','-parallel=4','-timeout=120s','-v']
 with(out/(name+'.log')).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 output=(out/(name+'.log')).read_text();caught=r.returncode==1 and catcher in output and 'build failed'not in output and 'panic: test timed out'not in output
 results.append(dict(mutant=name,command=command,exit=r.returncode,catcher=selector,assertion=catcher,caught=caught));(out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(name,'caught'if caught else'FAILED',flush=True)
 if not caught:raise SystemExit(1)
for name,relative,before,after,package,selector,catcher in specs:
 original=(root/relative).read_text();assert before in original,name
 run(name,relative,original,original.replace(before,after),package,selector,catcher)
# Preserve the prior re-cut's Effects proof too.
relative='internal/ir/call_targets.go';original=(root/relative).read_text();a=original.index('\tcase Effects:');b=original.index('\tcase Defined:',a)
run('omit-effects-result',relative,original,original[:a]+original[b:],'./internal/ir','^TestClosureTargetsPreserveEffectsProducers$','effects lost adapter identity')
# Original admission mutant: restore type matching rather than producer flow.
relative='internal/lower/cycles.go';original=(root/relative).read_text();mutated=original
a=mutated.index('\t\t\tif f.l.result.Locals[node.cell-1].Type == ir.Closure {');b=mutated.index('\t\t\tif proven := f.l.localTypes[node.cell-1]; proven != nil {',a);mutated=mutated[:a]+mutated[b:]
a=mutated.index('\t\t\t// Stored values and calls use the same producer graph.');b=mutated.index('\t\tcase f.l.checker.IsArrayType(proven)',a)
old='''\t\t\tfor _, closure := range f.l.closureRecords {
\t\t\t\tif !f.l.checker.IsTypeAssignableTo(closure.proven, proven) { continue }
\t\t\t\tfor _, local := range f.l.result.Functions[closure.function].Environment { queue = append(queue, cycleNode{cell: local + 1}) }
\t\t\t}
''';mutated=mutated[:a]+old+mutated[b:]
run('type-matched-captures',relative,original,mutated,'./internal/oracle','^TestClosureCycleAmbientCwd$','adamic/cycle-capable')
