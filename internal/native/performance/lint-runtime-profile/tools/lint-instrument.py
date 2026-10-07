import sys,shutil,subprocess
from pathlib import Path
base=Path(sys.argv[1]);target=Path(str(base)+'-instrumented');target.mkdir(exist_ok=True)
for p in base.iterdir():
 if p.suffix in ('.c','.h') or p.name in ('compiler.txt','small.txt'):shutil.copy2(p,target/p.name)
def edit(name,before,after):
 p=target/name;s=p.read_text();assert s.count(before)==1,(name,before,s.count(before));p.write_text(s.replace(before,after))
edit('heap.c','#include <stdint.h>','#include <stdint.h>\n#include <stdio.h>\nstatic size_t profile_null,profile_immortal,profile_shared,profile_last,profile_drains,profile_empty;\n__attribute__((destructor)) static void profile_release_report(void){fprintf(stderr,"release null=%zu immortal=%zu shared=%zu last=%zu drain_entries=%zu empty_drains=%zu\\n",profile_null,profile_immortal,profile_shared,profile_last,profile_drains,profile_empty);}')
edit('heap.c','\tADAMIC_COUNT_RELEASE();','\tADAMIC_COUNT_RELEASE();\n if(value==NULL) profile_null++;else if(((adamic_heap*)value)->references==0) profile_immortal++;else if(((adamic_heap*)value)->references==1)profile_last++;else profile_shared++;')
edit('heap.c','\tdraining = true;',' profile_drains++;if(freeing_count==0)profile_empty++;\n\tdraining = true;')
edit('string_build_impl.h','int adamic_string_equal(', 'static size_t profile_equal,profile_same,profile_memcmp;\n__attribute__((destructor)) static void profile_equal_report(void){fprintf(stderr,"equality calls=%zu same_header=%zu memcmp=%zu\\n",profile_equal,profile_same,profile_memcmp);}\nint adamic_string_equal(')
edit('string_build_impl.h','int adamic_string_equal(const adamic_string *left, const adamic_string *right) {', 'int adamic_string_equal(const adamic_string *left, const adamic_string *right) {\n profile_equal++;if(left==right)profile_same++;')
edit('string_build_impl.h','memcmp(left->bytes, right->bytes, left->length)', '(profile_memcmp++,memcmp(left->bytes, right->bytes, left->length))')
edit('map.c','#include <math.h>','#include <math.h>\n#include <stdio.h>\nstatic size_t profile_lookups,profile_probes;\n__attribute__((destructor)) static void profile_map_report(void){fprintf(stderr,"map lookups=%zu probes=%zu\\n",profile_lookups,profile_probes);}')
edit('map.c','static size_t find(const adamic_map *map, adamic_value key) {','static size_t find(const adamic_map *map, adamic_value key) {\n profile_lookups++;')
edit('map.c','\t\tsize_t slot = map->buckets[bucket];','\t\tprofile_probes++;\n\t\tsize_t slot = map->buckets[bucket];')
edit('string_search_impl.h','static unsigned *to_units(const adamic_string *string, size_t *count) {','static size_t profile_unit_buffers,profile_decoded_units;\n__attribute__((destructor)) static void profile_search_report(void){fprintf(stderr,"UTF16 temporary_buffers=%zu decoded_units=%zu\\n",profile_unit_buffers,profile_decoded_units);}\nstatic unsigned *to_units(const adamic_string *string, size_t *count) {\n profile_unit_buffers++;')
edit('string_search_impl.h','\t\t(*count)++;','\t\tprofile_decoded_units++;\n\t\t(*count)++;')
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2','-DADAMIC_COUNT']
units=[str(target/'main.c')]+[str(p) for p in sorted(target.glob('*.c')) if p.name!='main.c']
with (target/'build.log').open('wb') as log:subprocess.run(['clang',*flags,'-o',str(target/'counted'),*units,'-lm'],stdout=log,stderr=log,check=True)
with (target/'stdout').open('wb') as out,(target/'counts.txt').open('wb') as err:subprocess.run([str(target/'counted'),'--manifest',str(target/'compiler.txt'),'--count'],stdout=out,stderr=err,check=True)
print((target/'counts.txt').read_text())
