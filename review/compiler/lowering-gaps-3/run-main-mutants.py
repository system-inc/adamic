from pathlib import Path
import subprocess,json,difflib
out=Path('review/compiler/lowering-gaps-3');results=[]
pattern='^TestNativeAgreesWithNode$/internal/oracle/testdata/accessor_spread_throw_(first|middle|last)\\.a$'
def run(name, edits, expected, package='./internal/oracle', select=pattern, green=False):
 saved={p:Path(p).read_bytes() for p in edits}
 try:
  for p,change in edits.items():
   text=change(saved[p].decode()); assert text!=saved[p].decode(),name+' did not edit'
   Path(p).write_text(text)
   (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(saved[p].decode().splitlines(True),text.splitlines(True),fromfile=p,tofile=p)))
  with (out/(name+'.log')).open('wb') as log:
   result=subprocess.run(['timeout','150','go','test',package,'-run',select,'-count=1','-timeout=90s','-v'],stdout=log,stderr=subprocess.STDOUT)
  text=(out/(name+'.log')).read_text();caught=result.returncode==(0 if green else 1) and expected in text
  results.append(dict(name=name,exit=result.returncode,caught=caught,expected=expected));print(results[-1],flush=True)
  if not caught: raise RuntimeError(name+' did not produce intended evidence')
 finally:
  for p,content in saved.items():Path(p).write_bytes(content)
  (out/'main-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
run('main-static-revert',{'internal/lower/class_accessors.go':lambda s:s.replace('staticSpread = staticSpread || !plainDataSpreadSource(literal.Spread)','staticSpread = true')},'spreading in a program with static constructor objects',select='^TestNativeAgreesWithNode$/internal/oracle/testdata/static_constructor_spread_(gap|owned)\\.a$')
run('main-accessor-revert',{'internal/lower/class_accessors.go':lambda s:s.replace('(l.libraryFailure(l.result.Functions[accessor.Getter].Body, map[int]bool{}) != "")','(l.result.Functions[accessor.Getter].MayThrow || l.libraryFailure(l.result.Functions[accessor.Getter].Body, map[int]bool{}) != "")')},'spreading an accessor literal whose getter may throw')
run('main-accessor-release',{'internal/native/runtime/object.c':lambda s:s.replace('\t\t\tadamic_release(object);\n','')},'LeakSanitizer')
run('main-accessor-edge',{'internal/ir/accessor_spread.go':lambda s:s.replace('return !data(literal.Spread, map[int]bool{})','return false && !data(literal.Spread, map[int]bool{})')},'UndefinedBehaviorSanitizer')
run('main-accessor-broad-edge',{'internal/ir/accessor_spread.go':lambda s:s.replace('if literal.Spread == nil {','if literal.Spread != nil { return true }; if literal.Spread == nil {')},'throw edge true, want false','./internal/ir','^TestObjectSpreadAccessorEffects$')
run('main-accessor-getter-retain',{'internal/native/runtime/object.c':lambda s:s.replace('\t\tif (accessor == NULL) {','\t\tif (accessor != NULL) { adamic_retain(object->slots[index].reference); }\n\t\tif (accessor == NULL) {',1)},'LeakSanitizer')
def instrument(s,double=False):
 needle='\t\tif (adamic_thrown != NULL) {\n\t\t\tadamic_release(object);'
 replacement='''\t\tif (adamic_thrown != NULL) {
\t\t\tfor (size_t next = position + 1; next < shape->count; next++) {
\t\t\t\tsize_t unvisited = adamic_public_index(shape, next);
\t\t\t\tif (shape->references[unvisited] && object->slots[unvisited].reference != NULL) {
\t\t\t\t\tadamic_panic("uncopied slot not zero", sizeof "uncopied slot not zero" - 1);
\t\t\t\t}
\t\t\t}
\t\t\tobject->slots[index].reference = adamic_string_allocate(5);
'''
 if double:replacement+='\t\t\tadamic_release(object->slots[index].reference);\n'
 replacement+='\t\t\tadamic_release(object);'
 assert needle in s
 return s.replace(needle,replacement,1)
run('main-accessor-slots-owned-return',{'internal/native/runtime/object.c':lambda s:instrument(s)},'PASS',green=True)
run('main-accessor-return-double-release',{'internal/native/runtime/object.c':lambda s:instrument(s,True)},'AddressSanitizer')
