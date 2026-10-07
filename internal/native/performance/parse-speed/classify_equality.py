#!/usr/bin/env python3
"""After instrument.py and token_instrument.py, split unchanged equality bodies."""
from pathlib import Path
p = Path('/workspace/scratch/parse-speed/instrumented-runtime/string_build_impl.h')
t = p.read_text()
a = t.index('int adamic_string_equal(')
b = t.index('\nadamic_string adamic_string_empty', a)
body = '''
 if (left == NULL || right == NULL) {return left == right;}
 return left->length == right->length && (left->length == 0 || memcmp(left->bytes,right->bytes,left->length)==0);
'''
helpers = ''.join('__attribute__((noinline)) int adamic_profile_' + name + '_equal(const adamic_string *left,const adamic_string *right) {' + body + '}\n' for name in ['kind', 'other'])
part = t[a:b].replace('equal_counts[row][0]++;', 'int answer=row==0?adamic_profile_kind_equal(left,right):adamic_profile_other_equal(left,right);\n equal_counts[row][0]++;')
for value in ['left==right', '0', '1', 'memcmp(left->bytes,right->bytes,left->length)==0']:
    part = part.replace('return ' + value + ';', 'return answer;')
p.write_text(t[:a] + helpers + part + t[b:])
