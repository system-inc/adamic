from pathlib import Path
import re
base=Path('/workspace/scratch/parse-speed');c=(base/'parse-instrumented.c').read_text()
# All generated functions begin at column zero. Record completed scanner calls, including rescans.
chunks=re.split(r'(?=^static [^\n]+ \{\n)',c,flags=re.M)
for i,t in enumerate(chunks):
 if not re.match(r'static [^\n]*adamic_function_\d+_Scanner_(scan|string|template|rescanGreater|rescanTemplate|rescanSlash|scanJsx)\(',t):continue
 # Only the outer scan/rescan owns the counting interval. Nested helper calls are folded into it.
 arg=re.search(r'\(adamic_object \* (adamic_local_\d+_this)',t).group(1)
 t=t.replace('{\n','{\n\tadamic_profile_scan_begin();\n',1)
 t=re.sub(r'^(\s*)return ([^;]+);',lambda m:m[1]+'adamic_profile_scan_end('+arg+');\n'+m[0],t,flags=re.M)
 t=re.sub(r'^(\s*)return;',lambda m:m[1]+'adamic_profile_scan_end('+arg+');\n'+m[0],t,flags=re.M)
 # Void methods may fall through; unwind immediately before closing the function.
 if t.startswith('static void '):
  p=t.rfind('\n}')
  t=t[:p]+'\n\tadamic_profile_scan_end('+arg+');'+t[p:]
 chunks[i]=t
c=''.join(chunks)
helper='''
extern size_t adamic_profile_string_allocations;
static size_t scan_depth,scan_before,token_calls[5],token_allocations[5];
static void token_report(void) {
 static const char *names[]={"punctuation","keywords","identifiers","literals","other"};
 for(size_t i=0;i<5;i++)fprintf(stderr,"token %s scans=%zu string_allocations=%zu\\n",names[i],token_calls[i],token_allocations[i]);
}
static void adamic_profile_scan_begin(void) {
 static bool registered=false;if(!registered){atexit(token_report);registered=true;}
 if(scan_depth++==0)scan_before=adamic_profile_string_allocations;
}
static void adamic_profile_scan_end(adamic_object *scanner) {
 if(--scan_depth!=0)return;
 adamic_slot_cache cache={NULL,0};
 const adamic_string *kind=adamic_object_field(scanner,"kind",&cache)->reference;
 size_t n=kind->length,row=4;
 if(n>=5&&memcmp(kind->bytes+n-5,"Token",5)==0)row=0;
 else if(n>=7&&memcmp(kind->bytes+n-7,"Keyword",7)==0)row=1;
 else if((n==10&&memcmp(kind->bytes,"Identifier",10)==0)||(n==17&&memcmp(kind->bytes,"PrivateIdentifier",17)==0))row=2;
 else if((n>=7&&memcmp(kind->bytes+n-7,"Literal",7)==0)||(n>=8&&memcmp(kind->bytes,"Template",8)==0))row=3;
 token_calls[row]++;token_allocations[row]+=adamic_profile_string_allocations-scan_before;
}
'''
pos=c.index('\nstatic double adamic_function_')
c=c[:pos]+helper+c[pos:]
c=c.replace('#include <stdio.h>','#include <stdio.h>\n#include <string.h>',1)
(base/'parse-instrumented.c').write_text(c)
p=base/'instrumented-runtime/heap.c';t=p.read_text();t=t.replace('void *adamic_allocate(size_t size, enum adamic_kind kind) {','size_t adamic_profile_string_allocations;\nvoid *adamic_allocate(size_t size, enum adamic_kind kind) {\n if(kind==adamic_kind_string)adamic_profile_string_allocations++;');p.write_text(t)
