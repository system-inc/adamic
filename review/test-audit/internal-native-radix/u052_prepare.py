import pathlib,json,difflib,re,subprocess
root=pathlib.Path('/tmp/u052-mutants'); out=pathlib.Path('/workspace/adamic/review/test-audit/internal-native-radix')
base='internal/native/runtime/'
menu=[
('M01','radix.c','0.5 * (radix_next_double(value) - value)','1.0 * (radix_next_double(value) - value)','change constant'),
('M02','radix.c','radix_result("0", 1)','radix_result("1", 1)','change constant'),
('M03','radix.c','radix > 36','radix > 35','off-by-one bound'),
('M04','record.c','return true;\n}\n\nsize_t adamic_record_size','return false;\n}\n\nsize_t adamic_record_size','change constant'),
('M05','record.c','number >= UINT32_MAX','number > UINT32_MAX','off-by-one bound'),
('M06','record.c','return a < b ? -1 : a > b ? 1 : 0;','return a < b ? 1 : a > b ? -1 : 0;','swap comparator outcomes'),
('M07','regexp.c','regex_step_limit = limit;','regex_step_limit = 1;','change option'),
('M08','regexp.c','i->ranges[first].first <= c','i->ranges[first].first < c','off-by-one bound'),
('M09','regexp.c','strcmp(object->shape->names[k], "done")','strcmp(object->shape->names[k], "value")','change constant'),
('M10','string_build_impl.h','memcmp(left->bytes, right->bytes, left->length) == 0','memcmp(left->bytes, right->bytes, left->length) != 0','flip condition'),
('M11','heap.c','release_last(value);','(void)value;','drop statement'),
('M12','regexp.c','*steps >= regex_step_limit','*steps > regex_step_limit','off-by-one bound')]
original={f:(root/base/f).read_text() for _,f,*_ in menu}
# Inventory is deliberately conservative, including runtime helper definitions compiled alongside reached APIs.
inv=[]
for f in (root/base).glob('*'):
 if f.suffix not in ['.c','.h']:continue
 txt=f.read_text()
 for m in re.finditer(r'(?m)^(?:static |__attribute__\(\(noinline\)\) static )?(?:[\w*]+\s+)+([A-Za-z_]\w*)\s*\([^;{}]*\)\s*\{',txt):
  inv.append({'file':str(f.relative_to(root)),'line':txt[:m.start()].count('\n')+1,'function':m.group(1)})
(out/'runtime-function-inventory.json').write_text(json.dumps(inv,indent=2))
metadata=[]
for ident,f,a,b,kind in menu:
 assert original[f].count(a)==1,(ident,a,original[f].count(a))
 text=original[f].replace(a,b)
 diff=''.join(difflib.unified_diff(original[f].splitlines(True),text.splitlines(True),fromfile='a/'+base+f,tofile='b/'+base+f))
 (out/(ident+'.diff')).write_text(diff)
 metadata.append({'id':ident,'file':base+f,'line':original[f][:original[f].index(a)].count('\n')+1,'before':a,'after':b,'menu':kind})
(out/'mutant-plan.json').write_text(json.dumps(metadata,indent=2))
# Write fixed plan before running any mutant. Switch instrumentation is separate from replayable diffs.
texts=original.copy()
for ident,f,a,b,kind in menu:
 if ident=='M04':replacement='return !audit_selected("M04");\n}\n\nsize_t adamic_record_size'
 elif ident=='M06':replacement='return audit_selected("M06") ? (a < b ? 1 : a > b ? -1 : 0) : (a < b ? -1 : a > b ? 1 : 0);'
 elif ident=='M07':replacement='regex_step_limit = audit_selected("M07") ? 1 : limit;'
 elif ident=='M11':replacement='if (!audit_selected("M11")) release_last(value);'
 else:replacement='(audit_selected("'+ident+'") ? ('+b+') : ('+a+'))'
 texts[f]=texts[f].replace(a,replacement)
probes=[('P01','radix.c','adamic_string *adamic_number_to_radix(double value, double radix) {','return NULL;'),('P02','regexp.c','bool adamic_regex_test(adamic_object *regex, adamic_string *input) {','return false;'),('P03','regexp.c','adamic_array *adamic_regex_exec(adamic_object *regex, adamic_string *input) {','return NULL;'),('P04','regexp.c','adamic_maybe_boolean adamic_regex_done(adamic_object *object) {','return (adamic_maybe_boolean){false,false};'),('P05','string_build_impl.h','int adamic_string_equal(const adamic_string *left, const adamic_string *right) {','return 0;'),('P06','heap.c','void adamic_release(void *value) {','return;'),('P07','record.c','adamic_value *adamic_record_get(const adamic_record *record, const adamic_string *key) {','return NULL;'),('P08','record.c','adamic_array *adamic_record_keys(const adamic_record *record) {','return NULL;')]
for ident,f,entry,ret in probes:
 assert entry in original[f]
 modified=original[f].replace(entry,entry+'\n '+ret)
 (out/(ident+'.diff')).write_text(''.join(difflib.unified_diff(original[f].splitlines(True),modified.splitlines(True),fromfile='a/'+base+f,tofile='b/'+base+f)))
 texts[f]=texts[f].replace(entry,entry+'\n if (audit_selected("'+ident+'")) { '+ret+' }')
(out/'probe-plan.json').write_text(json.dumps(probes,indent=2))
helper='''\n#include <stdlib.h>\n#include <string.h>\nstatic bool audit_selected(const char *id) {\n static bool loaded; static const char *selected;\n if (!loaded) { selected = getenv("ADAMIC_MUTANT"); loaded = true; }\n return selected != NULL && strcmp(selected,id)==0;\n}\n'''
for f,text in texts.items():
 if f.endswith('.c'):text=text.replace('#include "adamic.h"','#include "adamic.h"'+helper,1)
 # string header is included by string.c, whose helper must precede it.
 (root/base/f).write_text(text)
string=(root/base/'string.c').read_text().replace('#include "adamic.h"','#include "adamic.h"'+helper,1)
(root/base/'string.c').write_text(string)
(out/'switch.diff').write_text(subprocess.check_output(['git','diff'],cwd=root,text=True))
