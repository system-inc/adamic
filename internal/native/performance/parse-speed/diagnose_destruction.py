#!/usr/bin/env python3
"""Scratch-only counters and identical destructor clones for the accepted baseline."""
from pathlib import Path
import argparse
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('directory', type=Path)
args = parser.parse_args()
repo = Path(__file__).resolve().parents[4]
base = args.directory.resolve()
source = (base / 'parse.c').read_text()
node = re.search(r'ParseNode_new\([^\n]+\) \{\s*ADAMIC_CHECK_STACK\(\);\s*[^\n]+adamic_object_new\(&(adamic_shape_\d+)\)', source).group(1)
labels = ['null', 'string', 'object', 'array_reference', 'map', 'cell', 'closure', 'map_iterator', 'number', 'boolean', 'weak', 'node', 'array_number']

def extent(text, start):
    opening = text.index('{', start)
    depth = 1
    end = opening + 1
    while depth:
        depth += (text[end] == '{') - (text[end] == '}')
        end += 1
    return end

for mode in ['counter', 'profile']:
    runtime = base / ('destruction-' + mode + '-runtime')
    runtime.mkdir(exist_ok=True)
    files = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', '9cc0d58', 'internal/native/runtime'], cwd=repo, text=True).splitlines()
    for name in files:
        p = Path(name)
        if p.suffix in ['.c', '.h']:
            (runtime / p.name).write_bytes(subprocess.check_output(['git', 'show', '9cc0d58:' + name], cwd=repo))
    c = source
    pos = c.index('\nstatic double adamic_function_')
    c = c[:pos] + '\nstatic const adamic_shape ' + node + ';\nint adamic_profile_type(void *value) {\n adamic_heap *heap = value;\n if(heap == NULL) return 0;\n if(heap->kind == adamic_kind_object && ((adamic_object *)value)->shape == &' + node + ') return 11;\n if(heap->kind == adamic_kind_array && !((adamic_array *)value)->references) return 12;\n return heap->kind;\n}\nextern void adamic_profile_phase(unsigned, const adamic_string *);\n' + c[pos:]
    start = c.index('static double adamic_function_224_run(', c.index('static double adamic_function_224_run(') + 1)
    end = extent(c, start)
    body = c[start:end]
    body = body.replace('ADAMIC_CHECK_STACK();', 'ADAMIC_CHECK_STACK(); adamic_profile_phase(1, adamic_local_1087_row);', 1)
    body = body.replace('adamic_object * adamic_temporary_9660 = adamic_function_222_collect', 'adamic_profile_phase(2, NULL);\n\tadamic_object * adamic_temporary_9660 = adamic_function_222_collect')
    body = body.replace('adamic_object * adamic_local_1113_context = adamic_temporary_9660;', 'adamic_object * adamic_local_1113_context = adamic_temporary_9660;\n\tadamic_profile_phase(3, NULL);')
    body = body.replace('return adamic_temporary_9664;', 'adamic_profile_phase(4, NULL);\n\t\t\treturn adamic_temporary_9664;')
    c = c[:start] + body + c[end:]
    c = c.replace('adamic_start(argc, argv);', 'adamic_start(argc, argv); adamic_profile_phase(0, NULL);')
    (base / ('destruction-' + mode + '.c')).write_text(c)
    heap = (runtime / 'heap.c').read_text()
    if mode == 'counter':
        helper = '''
#include <stdio.h>
extern int adamic_profile_type(void *);
static size_t destroyed[80][5][13], destroy_bytes[80][5][13];
static size_t releases[13][4], retains[13][4], child_drops[13][4];
static size_t allocated[11], node_allocations, live_nodes, peak_nodes;
static unsigned phase, file;
static void report_destruction(void) {
 for(unsigned f=0; f<=file; f++) for(unsigned p=0;p<5;p++) for(unsigned t=0;t<13;t++)
  if(destroyed[f][p][t]) fprintf(stderr,"destroy file=%u phase=%u type=%u count=%zu bytes=%zu\\n",f,p,t,destroyed[f][p][t],destroy_bytes[f][p][t]);
 for(unsigned t=0;t<13;t++) for(unsigned state=0;state<4;state++)
  fprintf(stderr,"references type=%u state=%u release=%zu retain=%zu child=%zu\\n",t,state,releases[t][state],retains[t][state],child_drops[t][state]);
 for(unsigned t=1;t<11;t++) fprintf(stderr,"allocate type=%u count=%zu\\n",t,allocated[t]);
 fprintf(stderr,"nodes allocated=%zu live=%zu peak=%zu\\n",node_allocations,live_nodes,peak_nodes);
}
void adamic_profile_phase(unsigned p,const adamic_string *path) {
 static bool registered;
 if(!registered){atexit(report_destruction);registered=true;}
 if(p==1){file++; fprintf(stderr,"file index=%u path=%.*s\\n",file,(int)path->length,path->bytes);}
 phase=p;
 if(p==4) fprintf(stderr,"file_end index=%u live_nodes=%zu\\n",file,live_nodes);
}
static unsigned state_of(void *value) {
 adamic_heap *h=value;return h==NULL?0:h->references==0?1:h->references==1?3:2;
}
static void record_reference(size_t counts[13][4],void *value) {
 counts[adamic_profile_type(value)][state_of(value)]++;
}
static void record_destroy(void *value) {
 adamic_heap *h=value; int type=adamic_profile_type(value); size_t bytes=0;
 if(h->kind==adamic_kind_object)bytes=sizeof(adamic_object)+((adamic_object *)value)->shape->count*sizeof(adamic_value);
 if(h->kind==adamic_kind_array)bytes=sizeof(adamic_array)+((adamic_array *)value)->capacity*sizeof(adamic_value);
 if(h->kind==adamic_kind_string)bytes=sizeof(adamic_string)+((adamic_string *)value)->capacity;
 destroyed[file][phase][type]++;destroy_bytes[file][phase][type]+=bytes;
 if(type==11)live_nodes--;
}
void adamic_profile_new_object(adamic_object *object) {
 if(adamic_profile_type(object)==11){node_allocations++;live_nodes++;if(live_nodes>peak_nodes)peak_nodes=live_nodes;}
}
'''
        heap = heap.replace('#include <stdlib.h>', '#include <stdlib.h>\n' + helper)
        heap = heap.replace('ADAMIC_COUNT_ALLOCATION();', 'allocated[kind]++;\n\tADAMIC_COUNT_ALLOCATION();')
        heap = heap.replace('ADAMIC_COUNT_RETAIN();', 'record_reference(retains,value);\n\tADAMIC_COUNT_RETAIN();')
        heap = heap.replace('ADAMIC_COUNT_RELEASE();', 'record_reference(releases,value);\n\tADAMIC_COUNT_RELEASE();')
        heap = heap.replace('static void let_go(void *value) {', 'static void let_go(void *value) {\n if(draining)record_reference(child_drops,value);')
        heap = heap.replace('static void free_one(void *value) {', 'static void free_one(void *value) {\n record_destroy(value);')
        p = runtime / 'object.c'
        text = p.read_text().replace('adamic_object *adamic_object_new(', 'extern void adamic_profile_new_object(adamic_object *);\n\nadamic_object *adamic_object_new(',1)
        text = text.replace('\treturn object;', '\tadamic_profile_new_object(object);\n\treturn object;',1)
        p.write_text(text)
    else:
        start = heap.index('static void free_one(')
        end = extent(heap,start)
        original = heap[start:end]
        clones = [original.replace('static void free_one', 'static __attribute__((noinline)) void profile_free_' + name) for name in labels[1:]]
        dispatch = '\nextern int adamic_profile_type(void *);\nvoid adamic_profile_phase(unsigned phase,const adamic_string *path){(void)phase;(void)path;}\nstatic void profile_free_dispatch(void *value) {\n switch(adamic_profile_type(value)){\n' + ''.join('case %d:profile_free_%s(value);break;\n' % (i,name) for i,name in enumerate(labels) if i) + 'default:abort();\n}\n}\n'
        heap = heap[:start] + '\n'.join(clones) + dispatch + heap[end:]
        heap = heap.replace('free_one(freeing[--freeing_count]);', 'profile_free_dispatch(freeing[--freeing_count]);')
    (runtime / 'heap.c').write_text(heap)
print('Prepared scratch diagnostics:', ', '.join(labels))
