from pathlib import Path
import subprocess,json
out=Path('review/compiler/lowering-gaps-3/runtime-accessor-guard');results=[]
def test(name,pattern,changes,expected,exitcode):
 saved={p:Path(p).read_bytes() for p in changes}
 try:
  for p,change in changes.items():Path(p).write_text(change(saved[p].decode()))
  with (out/(name+'.log')).open('wb') as log:r=subprocess.run(['timeout','120','go','test','./internal/oracle','-run',pattern,'-count=1','-timeout=90s','-v'],stdout=log,stderr=subprocess.STDOUT)
  caught=r.returncode==exitcode and expected in (out/(name+'.log')).read_text();results.append(dict(name=name,exit=r.returncode,caught=caught));print(results[-1],flush=True)
  assert caught,name
 finally:
  for p,content in saved.items():Path(p).write_bytes(content)
  (out/'independent-proofs.json').write_text(json.dumps(results,indent=2)+'\n')
main=lambda p:subprocess.check_output(['git','show','5e33a17b1:'+p]).decode()
test('accessor-on-main-alone','^TestNativeAgreesWithNode$/internal/oracle/testdata/(accessor_spread_throw_(first|middle|last)|plain_data_spread)\\.a$',{'internal/lower/class_accessors.go':lambda s:main('internal/lower/class_accessors.go').replace('l.result.Functions[accessor.Getter].MayThrow || ',''),'internal/lower/expression.go':lambda s:main('internal/lower/expression.go')},'PASS',0)
pattern='^TestNativeAgreesWithNode$/internal/oracle/testdata/template_nullish\\.a$'
test('main-template-revert',pattern,{'internal/lower/expression.go':lambda s:main('internal/lower/expression.go')},'stage 0 can\'t lower a template',1)
test('main-template-spelling',pattern,{'internal/lower/expression.go':lambda s:s.replace('spelling = "null"','spelling = "undefined"')},'stdout differs',1)
test('main-template-effects',pattern,{'internal/lower/expression.go':lambda s:s.replace('Body: []ir.Statement{ir.Evaluate{Value: value}}, Result: ir.StringConstant{Index: l.constant(spelling)}','Body: []ir.Statement{}, Result: ir.StringConstant{Index: l.constant(spelling)}')},'stdout differs',1)
