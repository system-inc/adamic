#!/usr/bin/env python3
"""Count actual inline charCodeAt paths in an isolated runtime/C snapshot."""
import argparse
from pathlib import Path
import re
import shutil

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('directory', type=Path)
args = parser.parse_args()
repo = Path(__file__).resolve().parents[4]
base = args.directory.resolve()
runtime = base / 'character-runtime'
runtime.mkdir(exist_ok=True)
for p in (repo/'internal/native/runtime').iterdir():
    if p.suffix in ['.h','.c']:
        shutil.copy2(p,runtime/p.name)
p = runtime/'adamic.h'
s = p.read_text()
s = s.replace('static inline double adamic_string_char_code_at(', 'extern void adamic_profile_read(const adamic_string *, unsigned);\nstatic inline double adamic_string_char_code_at(',1)
start=s.index('static inline double adamic_string_char_code_at(')
end=s.index('\nadamic_string *adamic_string_trim',start)
part=s[start:end]
part=part.replace('return (double)(unsigned char)string->bytes[(size_t)position];','adamic_profile_read(string,0); return (double)(unsigned char)string->bytes[(size_t)position];')
part=part.replace('return (double)string->index->view[(size_t)position];','adamic_profile_read(string,1); return (double)string->index->view[(size_t)position];')
part=part.replace('return adamic_string_char_code(string, position);','adamic_profile_read(string,2); return adamic_string_char_code(string, position);')
p.write_text(s[:start]+part+s[end:])
c=(base/'runtime-parse.c').read_text()
helper='''
#include <stdio.h>
#include <stdlib.h>
static size_t profile_reads[2][3];
static const adamic_string *profile_source;
void adamic_profile_read(const adamic_string *s,unsigned path){profile_reads[s==profile_source][path]++;}
static void profile_read_report(void){
 for(unsigned source=0;source<2;source++)fprintf(stderr,"char_read source=%u ascii=%zu view=%zu fallback=%zu\\n",source,profile_reads[source][0],profile_reads[source][1],profile_reads[source][2]);
}
'''
pos=c.index('\nstatic double adamic_function_')
c=c[:pos]+helper+c[pos:]
c=c.replace('adamic_start(argc, argv);','adamic_start(argc, argv); atexit(profile_read_report);')
# The source variable is assigned from readTextFile's text in the run function.
match=re.search(r'adamic_string \* adamic_local_1112_source = [^;]+;',c)
assert match
c=c[:match.end()]+'\n profile_source=adamic_local_1112_source;'+c[match.end():]
# Count-mode path only: clear the borrowed profiling pointer after per-file cleanup.
start=c.index('static double adamic_function_224_run(',c.index('static double adamic_function_224_run(')+1)
count=c.index('if (adamic_local_1088_count)',start)
first_return=c.index('return ',count)
c=c[:first_return]+'profile_source=NULL; '+c[first_return:]
(base/'character-counter.c').write_text(c)
print('Prepared isolated ASCII/view/fallback counters')
