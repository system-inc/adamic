"""Instrument a scratch generated C unit, never a compiler source file.

Use from a clang wrapper during an explicitly profiled build. Measures only
outermost calls, so recursive parent indexing/walking are not double counted.
"""
from pathlib import Path
import re
import sys

path = Path(sys.argv[1])
text = path.read_text()
pattern = re.compile(r'^static (.+?)(adamic_function_\d+_(?:Parser_file|Rules_run|Rules_links|Rules_walk|Shadow_run|Assignment_run|VoidRule_run|Returns_run|Unbound_run|Volume_run|types|byteOffsets))\((.*?)\) \{', re.M)
wrappers = []
reports = []
helper = """
static uint64_t phase_queries_ns, phase_decoding_ns;
static adamic_string *phase_inspect(double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question) {
 uint64_t started = phase_now();
 adamic_string *answer = adamic_tsgo_inspect(handle, file, start, end, kind, question);
 phase_queries_ns += phase_now() - started;
 return answer;
}
static adamic_string *phase_parts(double handle, const adamic_string *file, double start, double end, const adamic_string *kind) {
 uint64_t started = phase_now();
 adamic_string *answer = adamic_tsgo_type_parts(handle, file, start, end, kind);
 phase_queries_ns += phase_now() - started;
 return answer;
}
"""
text = text.replace('adamic_tsgo_inspect(', 'phase_inspect(').replace('adamic_tsgo_type_parts(', 'phase_parts(')
# Generated C includes the runtime types before its first static declaration.
first = re.search(r'^static ', text, re.M)
assert first, 'Generated static declarations missing'
text = text[:first.start()] + helper + text[first.start():]

def wrap(match):
    returns, name, parameters = match.groups()
    arguments = [] if parameters == 'void' else [re.search(r'(\w+)$', p.strip()).group(1) for p in parameters.split(',')]
    call = name + '_profile_body(' + ', '.join(arguments) + ');'
    body = call if returns.strip() == 'void' else returns + 'result = ' + call
    finish = '' if returns.strip() == 'void' else 'return result;'
    decoding = 'phase_decoding_ns += elapsed;' if name.endswith('_types') else ''
    wrappers.append(f'''static uint64_t {name}_ns, {name}_queries, {name}_decoding;
static {returns}{name}({parameters}) {{
 static unsigned depth;
 uint64_t started = depth++ == 0 ? phase_now() : 0;
 uint64_t queries = phase_queries_ns, decoder = phase_decoding_ns;
 {body}
 if (--depth == 0) {{
  uint64_t elapsed = phase_now() - started;
  {name}_ns += elapsed;
  {name}_queries += phase_queries_ns - queries;
  {name}_decoding += phase_decoding_ns - decoder;
  {decoding}
 }}
 {finish}
}}
''')
    reports.append(f' fprintf(stderr, "native_phase: {name}=%" PRIu64 "\\n", {name}_ns);')
    reports.append(f' fprintf(stderr, "native_nested: {name}_queries=%" PRIu64 " {name}_decoding=%" PRIu64 "\\n", {name}_queries, {name}_decoding);')
    return f'static {returns}{name}_profile_body({parameters}) {{'

text, count = pattern.subn(wrap, text)
assert count == 12, f'Expected exactly twelve generated phase functions, saw {count}'
header = '''#define _POSIX_C_SOURCE 200809L
#include <time.h>
#include <inttypes.h>
#include <stdio.h>
static uint64_t phase_now(void) {
 struct timespec now;
 if (clock_gettime(CLOCK_MONOTONIC, &now) != 0) __builtin_trap();
 return (uint64_t)now.tv_sec * UINT64_C(1000000000) + (uint64_t)now.tv_nsec;
}
'''
report = '\n__attribute__((destructor)) static void phase_report(void) {\n' + '\n'.join(reports) + '\n}\n'
path.write_text(header + text + '\n' + '\n'.join(wrappers) + report)
