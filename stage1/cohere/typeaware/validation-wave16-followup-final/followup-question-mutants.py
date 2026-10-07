from pathlib import Path
import subprocess,json,os
root=Path('/workspace/adamic');dest=Path('/workspace/wave16-artifacts/followup-question-mutants');dest.mkdir(exist_ok=True)
mutants=[
 ('module-kind','emit_module_kind.go','out.number(uint64(p.Compiler.Options().GetEmitModuleKind()))','out.number(uint64(p.Compiler.Options().GetEmitModuleKind()) + 1)','emit module:'),
 ('signature-name','signature_ancestry.go','out.text(name)','out.text(name + "mutant")','signature ancestry:'),
 ('symbol-id','symbol_origins.go','out.number(p.symbolID(symbol))','out.number(p.symbolID(symbol) + 1)','symbol identity:'),
 ('assignable','listener_type_facts.go','out.yes(known != nil && target != nil && c.IsTypeAssignableTo(c.GetNonNullableType(known), c.GetNonNullableType(target)))','out.yes(!(known != nil && target != nil && c.IsTypeAssignableTo(c.GetNonNullableType(known), c.GetNonNullableType(target))))','assignability:'),
 ('render-flags','render_type_facts.go','out.number(uint64(t.Flags()))','out.number(uint64(t.Flags()) + 1)','number graph:'),
 ('question-suffix','additional_questions.go','if question != mode {','if false {','suffix accepted:'),
]
for name,file,old,new,expected in mutants:
 original=root/'bridge/tsgo/checker'/file;source=original.read_text();assert source.count(old)==1,(name,old)
 replacement=dest/(name+'.go');replacement.write_text(source.replace(old,new,1));overlay=dest/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(replacement)}}))
 log=dest/(name+'.log')
 with log.open('wb') as output:
  result=subprocess.run(['go','test','-overlay',str(overlay),'./bridge/tsgo/checker','-run','^TestWave16QuestionFacts$','-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
 data=log.read_text();assert result.returncode==1 and expected in data and '[build failed]' not in data,(name,data)
 print(name,'caught by',expected,flush=True)
