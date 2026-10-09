from pathlib import Path
import json,subprocess,difflib,hashlib
p=Path('review/test-audit/internal-native-number');(p/'diffs').mkdir(exist_ok=True);plan=[]
def add(id,file,old,new,menu,kind='production'):
 text=subprocess.check_output(['git','show','HEAD:'+file],text=True);assert text.count(old)==1 or (id=="M05" and text.count(old)==2),(id,text.count(old))
 plan.append(dict(id=id,file=file,line=text[:text.index(old)].count('\n')+1,old=old,new=new,menu=menu,kind=kind))
number='internal/native/runtime/number.c';parse='internal/native/runtime/parse.c';dtoa='internal/native/runtime/dtoa.c';cache='internal/buildcache/buildcache.go'
add('M01',number,'append(buffer, 0, "NaN")','append(buffer, 0, "nan")','change constant')
add('M02',number,'-6 < point','-5 < point','off-by-one')
add('M03',number,'int exponent = point - 1;','int exponent = point;','off-by-one')
add('M04',number,'value < 9007199254740992.0','value < 4503599627370496.0','change constant')
add('M05',dtoa,'if ((double_bits(v) & double_significand_mask) == 0) {','if ((double_bits(v) & double_significand_mask) == 1) {','change constant')
add('M06',parse,'radix < 2 || radix > 36','radix < 2 || radix > 35','off-by-one')
add('M07',parse,'fmod(trunc(value), 4294967296.0)','fmod(trunc(value), 4294967295.0)','change constant')
add('M08',parse,'end - cursor >= 8 && memcmp','end - cursor >= 9 && memcmp','off-by-one')
add('M09',parse,'if (digits > 0) {','if (digits > 1) {','off-by-one')
add('M10',cache,'\tfor _, flag := range inputs.Flags {\n\t\tfield("flag", flag)\n\t}','\t// Flag hashing loop dropped.','drop statement')
add('M11',cache,'field("name", inputs.Name)','field("name", "artifact")','change constant')
add('M12',cache,'product := filepath.Join(cache, key)\n\tif _, err = os.Stat(product); err == nil {','product := filepath.Join(cache, key)\n\tif _, err = os.Stat(product); err != nil {','flip condition')
add('E01',number,'size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]) {','size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]) { return 0;','return early','empty-answer')
add('E02',parse,'double adamic_number_parse_int(const adamic_string *text, double radix_value) {','double adamic_number_parse_int(const adamic_string *text, double radix_value) { return 0;','return early','empty-answer')
add('E03',parse,'double adamic_number_parse_float(const adamic_string *text) {','double adamic_number_parse_float(const adamic_string *text) { return 0;','return early','empty-answer')
add('E04',cache,'func Get(inputs Inputs, build func(directory string) error) (string, error) {','func Get(inputs Inputs, build func(directory string) error) (string, error) { return "", nil;','return early','empty-answer')
(p/'plan.json').write_text(json.dumps(plan,indent=2));print('frozen',hashlib.sha256((p/'plan.json').read_bytes()).hexdigest())
for m in plan:
 original=subprocess.check_output(['git','show','HEAD:'+m['file']],text=True);changed=original.replace(m['old'],m['new'],1)
 (p/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
for file in dict.fromkeys(m['file'] for m in plan):
 text=subprocess.check_output(['git','show','HEAD:'+file],text=True)
 for m in [x for x in plan if x['file']==file]:
  old,new=m['old'],m['new'];sel='audit_mutant("'+m['id']+'")' if file.endswith('.c') else 'auditMutant("'+m['id']+'")'
  id=m['id']
  if id.startswith('E'):replacement=old+'\nif ('+sel+') { return 0; }' if file.endswith('.c') else old+'\nif '+sel+' { return "", nil }'
  elif id=='M01':replacement='append(buffer, 0, '+sel+' ? "nan" : "NaN")'
  elif id=='M02':replacement='('+sel+' ? -5 : -6) < point'
  elif id=='M03':replacement='int exponent = point - ('+sel+' ? 0 : 1);'
  elif id=='M04':replacement='value < ('+sel+' ? 4503599627370496.0 : 9007199254740992.0)'
  elif id=='M05':replacement='if ((double_bits(v) & double_significand_mask) == ('+sel+' ? 1u : 0u)) {'
  elif id=='M06':replacement='radix < 2 || radix > ('+sel+' ? 35 : 36)'
  elif id=='M07':replacement='fmod(trunc(value), '+sel+' ? 4294967295.0 : 4294967296.0)'
  elif id=='M08':replacement='end - cursor >= ('+sel+' ? 9 : 8) && memcmp'
  elif id=='M09':replacement='if (digits > ('+sel+' ? 1u : 0u)) {'
  elif id=='M10':replacement='if !'+sel+' {\n'+old+'\n}'
  elif id=='M11':replacement='field("name", auditName(inputs.Name))'
  elif id=='M12':replacement='product := filepath.Join(cache, key)\n\tif _, err = os.Stat(product); (err == nil) != '+sel+' {'
  assert old in text; text=text.replace(old,replacement,1)
 if file.endswith('.c'):text=text.replace('#include "adamic.h"','#include "adamic.h"\n#include "audit_mutant.h"')
 Path(file).write_text(text)
Path('internal/native/runtime/audit_mutant.h').write_text('#ifndef ADAMIC_AUDIT_MUTANT_H\n#define ADAMIC_AUDIT_MUTANT_H\n#include <stdlib.h>\n#include <string.h>\nstatic bool audit_mutant(const char *id) { const char *selected = getenv("ADAMIC_MUTANT"); return selected != NULL && strcmp(selected, id) == 0; }\n#endif\n')
Path('internal/buildcache/audit_mutant.go').write_text('package buildcache\nimport "os"\nfunc auditMutant(id string) bool {return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditName(name string) string {if auditMutant("M11") {return "artifact"};return name}\n')
