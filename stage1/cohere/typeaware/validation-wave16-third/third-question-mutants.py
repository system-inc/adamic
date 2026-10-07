from pathlib import Path
import subprocess,json
root=Path('/workspace/adamic');dest=Path('/workspace/wave16-artifacts/third-question-mutants');dest.mkdir(exist_ok=True)
mutants=[
 ('text-flags','text_type_facts.go','out.number(uint64(t.Flags()))','out.number(uint64(t.Flags()) + 1)','text type flags differ'),
 ('text-literal','text_type_facts.go','kind = "string"','kind = "boolean"','literal value differs'),
 ('value-range','value_declaration.go','out.number(uint64(d.End()))','out.number(uint64(d.End()) + 1)','value declaration differs'),
 ('value-kind','value_declaration.go','return fmt.Errorf("value-declaration requires an Identifier")','return nil','value declaration accepted wrong node kind'),
 ('literal-utf8','text_type_facts.go','if !utf8.ValidString(value) {','if false {','non-UTF-8 literal accepted'),
 ('template-utf8','text_type_facts.go','if !utf8.ValidString(text) {','if false {','non-UTF-8 template accepted'),
]
for name,file,old,new,expected in mutants:
 original=root/'bridge/tsgo/checker'/file;source=original.read_text();assert source.count(old)==1,(name,old)
 replacement=dest/(name+'.go');replacement.write_text(source.replace(old,new,1));overlay=dest/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(replacement)}}))
 log=dest/(name+'.log')
 with log.open('wb') as output:result=subprocess.run(['go','test','-overlay',str(overlay),'./bridge/tsgo/checker','-run','^TestWave16TextQuestions$','-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
 data=log.read_text();assert result.returncode==1 and expected in data and '[build failed]' not in data,(name,data)
 print(name,'caught by',expected,flush=True)
