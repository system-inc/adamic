import pathlib, subprocess, json
root = pathlib.Path('.'); out = root/'review/compiler/lowering-gaps-3'
results = [r for r in json.loads((out/"accessor-mutants.json").read_text()) if r["caught"]]
def run(name, edits, expected, package='./internal/oracle', pattern='TestNativeAgreesWithNode/internal/oracle/testdata/accessor_spread_throw'):
 if any(r["name"] == name for r in results): return
 saved = {p: pathlib.Path(p).read_bytes() for p in edits}
 try:
  for p, transform in edits.items(): pathlib.Path(p).write_text(transform(saved[p].decode()))
  with (out/(name+'.log')).open('wb') as log:
   result = subprocess.run(['timeout','180','go','test',package,'-run',pattern,'-count=1','-timeout','90s','-v'],stdout=log,stderr=subprocess.STDOUT)
  text = (out/(name+'.log')).read_text()
  caught = result.returncode == 1 and expected in text
  results.append(dict(name=name,exit=result.returncode,caught=caught,expected=expected))
  print(results[-1],flush=True)
  if not caught: raise RuntimeError(name+' did not fail by its intended check')
 finally:
  for p, content in saved.items(): pathlib.Path(p).write_bytes(content)
  (out/'accessor-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
run('accessor-revert', {'internal/lower/class_accessors.go':lambda s:s.replace('func (l *lowering) checkAccessorSpreads() error {','func (l *lowering) checkAccessorSpreads() error {\n return &NotYet{Where: l.result.Source, What: "spreading an accessor literal whose getter may throw"}\n')}, 'stage 0 can\'t lower spreading an accessor literal whose getter may throw')
run('accessor-release', {'internal/native/runtime/object.c':lambda s:s.replace('\t\tadamic_release(object);\n\t\treturn NULL;', '\t\treturn NULL;')}, 'LeakSanitizer')
run('accessor-edge', {'internal/ir/accessor_spread.go':lambda s:s.replace('return !data(literal.Spread, map[int]bool{})','return false && !data(literal.Spread, map[int]bool{})')}, 'UndefinedBehaviorSanitizer')
run('accessor-broad-edge', {'internal/ir/accessor_spread.go':lambda s:s.replace('if literal.Spread == nil {','if literal.Spread != nil { return true }; if literal.Spread == nil {')}, 'throw edge true, want false', './internal/ir', '^TestObjectSpreadAccessorEffects$')

run('accessor-getter-retain', {'internal/native/runtime/object.c':lambda s:s.replace('if (adamic_exception_pending) return false;', 'if (adamic_exception_pending) return false;\n\t\t\tadamic_retain(object->slots[index].reference);')}, 'LeakSanitizer')
