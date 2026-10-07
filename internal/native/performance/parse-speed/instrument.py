from pathlib import Path
import re,shutil
base=Path('/workspace/scratch/parse-speed');repo=Path('/workspace/adamic')
c=(base/'parse.c').read_text()
literals=re.findall(r'static adamic_string (adamic_string_\d+) = ADAMIC_STRING\("([^"\\]*)"\);',c)
kinds=set(re.findall(r'^\s*Kind(\w+)\s', (repo/'cohere/TypeScript/tsc/internal/ast/kind_generated.go').read_text(), re.M))
names=[name for name,value in literals if value in kinds]
assert names
from collections import Counter
assert max(Counter(value for _,value in literals).values())==1
helper='''
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
static const adamic_string *kind_headers[] = {NAMES};
static int header_order(const void *a,const void *b) {
 uintptr_t x=(uintptr_t)*(const adamic_string *const *)a,y=(uintptr_t)*(const adamic_string *const *)b;
 return x<y?-1:x>y?1:0;
}
bool adamic_profile_kind(const adamic_string *s) {
 uintptr_t key=(uintptr_t)s;
 size_t low=0,high=sizeof kind_headers/sizeof kind_headers[0];
 while(low<high) {size_t mid=low+(high-low)/2;uintptr_t p=(uintptr_t)kind_headers[mid]; if(p<key)low=mid+1;else high=mid;}
 return low<sizeof kind_headers/sizeof kind_headers[0] && kind_headers[low]==s;
}
'''.replace('NAMES',','.join('&'+n for n in names))
pos=c.index('\nstatic ',c.index('static adamic_string')+1)
# Helper must follow all literal declarations, before functions.
pos=c.index('\nstatic double adamic_function_')
c=c[:pos]+helper+c[pos:]
c=c.replace('adamic_start(argc, argv);','adamic_start(argc, argv); qsort(kind_headers,sizeof kind_headers/sizeof kind_headers[0],sizeof kind_headers[0],header_order);')
(base/'parse-instrumented.c').write_text(c)
runtime=base/'instrumented-runtime';runtime.mkdir(exist_ok=True)
for p in (repo/'internal/native/runtime').iterdir():
 if p.suffix in ['.h','.c']:shutil.copy2(p,runtime/p.name)
p=runtime/'string_build_impl.h';t=p.read_text();a=t.index('int adamic_string_equal(');b=t.index('\nadamic_string adamic_string_empty',a)
t=t[:a]+'''
extern bool adamic_profile_kind(const adamic_string *);
static size_t equal_counts[2][6];
static void equal_report(void) {
 for(size_t i=0;i<2;i++) fprintf(stderr,"%s calls=%zu same_header=%zu null=%zu different_length=%zu empty=%zu memcmp=%zu\\n",i==0?"kind":"other",equal_counts[i][0],equal_counts[i][1],equal_counts[i][2],equal_counts[i][3],equal_counts[i][4],equal_counts[i][5]);
}
int adamic_string_equal(const adamic_string *left,const adamic_string *right) {
 static bool registered=false; if(!registered){atexit(equal_report);registered=true;}
 size_t row=adamic_profile_kind(left)&&adamic_profile_kind(right)?0:1;
 equal_counts[row][0]++;equal_counts[row][1]+=left==right;
 if(left==NULL||right==NULL){equal_counts[row][2]++;return left==right;}
 if(left->length!=right->length){equal_counts[row][3]++;return 0;}
 if(left->length==0){equal_counts[row][4]++;return 1;}
 equal_counts[row][5]++;
 return memcmp(left->bytes,right->bytes,left->length)==0;
}
'''+t[b:];p.write_text(t)
print('interned literals',len(literals),'kind headers',len(names))
(base/'literal-summary.txt').write_text(f'{len(literals)} literal headers, {len(names)} kind headers; no duplicate literal contents\n')
