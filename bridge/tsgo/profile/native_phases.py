"""Instrument a scratch generated C unit, never a compiler source file.

Use from a clang wrapper during an explicitly profiled build. Measures only
outermost calls, so recursive parent indexing/walking are not double counted.
"""
from pathlib import Path
import re
import sys

path = Path(sys.argv[1])
text = path.read_text()
pattern = re.compile(r'^static (.+?)(adamic_function_\d+_(?:Parser_file|Rules_run|Rules_links|Rules_walk|types|byteOffsets))\((.*?)\) \{', re.M)
wrappers = []
reports = []

def wrap(match):
    returns, name, parameters = match.groups()
    arguments = [] if parameters == 'void' else [re.search(r'(\w+)$', p.strip()).group(1) for p in parameters.split(',')]
    call = name + '_profile_body(' + ', '.join(arguments) + ');'
    body = call if returns.strip() == 'void' else returns + 'result = ' + call
    finish = '' if returns.strip() == 'void' else 'return result;'
    wrappers.append(f'''static uint64_t {name}_ns;
static {returns}{name}({parameters}) {{
 static unsigned depth;
 uint64_t started = depth++ == 0 ? phase_now() : 0;
 {body}
 if (--depth == 0) {name}_ns += phase_now() - started;
 {finish}
}}
''')
    reports.append(f' fprintf(stderr, "native_phase: {name}=%" PRIu64 "\\n", {name}_ns);')
    return f'static {returns}{name}_profile_body({parameters}) {{'

text, count = pattern.subn(wrap, text)
assert count == 6, f'Expected exactly six generated phase functions, saw {count}'
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
