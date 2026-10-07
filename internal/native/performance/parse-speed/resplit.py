#!/usr/bin/env python3
"""Disjoint v2 mechanism split, including DWARF inline runtime source costs."""
import argparse
import collections
import importlib.util
import json
from pathlib import Path
import re
import subprocess
import sys
sys.dont_write_bytecode = True
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('profile', type=Path)
parser.add_argument('c_source', type=Path)
parser.add_argument('runtime_ref')
parser.add_argument('output', type=Path)
args = parser.parse_args()
repo = Path(__file__).resolve().parents[4]
spec = importlib.util.spec_from_file_location('scanner_profile', repo / 'stage1/typescript/scanner/profile.py')
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
j = module.summarize(args.profile)
header = subprocess.check_output(['git', 'show', args.runtime_ref + ':internal/native/runtime/adamic.h'], cwd=repo, text=True).splitlines()
inline = {}
for i, line in enumerate(header):
    match = re.match(r'static inline .*\b(adamic_\w+)\(', line)
    if not match:
        continue
    name = match.group(1)
    depth = 0
    for k in range(i, len(header)):
        inline[k + 1] = name
        depth += header[k].count('{') - header[k].count('}')
        if depth == 0:
            break
source = args.c_source.read_text().splitlines()
line_table = set()
for i, line in enumerate(source):
    if re.search(r'static adamic_array \* adamic_function_\d+_lines\(.*\) \{', line):
        depth = 0
        for k in range(i, len(source)):
            line_table.add(k + 1)
            depth += source[k].count('{') - source[k].count('}')
            if depth == 0:
                break
buckets = collections.Counter()
release = {'destroy_last_reference', 'adamic_release', 'let_go', 'free_one', 'adamic_object_free_children', 'release_field', 'adamic_map_free_children', 'adamic_weak_forget', 'adamic_string_free_index'}
allocation = {'adamic_allocate', 'malloc', 'free', 'realloc', 'calloc', 'take', 'give', 'new_chunk', 'deallocate'}
character = {'adamic_string_char_code', 'adamic_string_bmp_view', 'adamic_string_unit_view', 'adamic_string_units', 'adamic_string_locate', 'adamic_string_units_before', 'usable', 'build', 'width', 'decode', 'unit_at', 'sequence', 'adamic_string_code_point_at'}
field = {'adamic_object_field', 'adamic_object_data_field', 'adamic_object_find', 'adamic_object_write_field', 'adamic_object_check_write', 'adamic_object_callee', 'adamic_static_field'}
for row in j['line_self']:
    fn, file, line, cost = row['function'], row['file'], row['line'], row['instructions']
    inlined = inline.get(line, '') if file == 'adamic.h' else ''
    if fn in release:
        bucket = 'Releases and child destruction'
    elif fn == 'adamic_retain':
        bucket = 'Retains'
    elif fn in allocation:
        bucket = 'Allocation and freeing'
    elif fn == 'adamic_string_equal':
        bucket = 'String equality bodies, kind and other'
    elif file == 'input.c':
        bucket = 'File reading and UTF-8 input decode self'
    elif fn in character or inlined in {'adamic_string_length', 'adamic_string_char_code_at'}:
        bucket = 'Character reads and UTF-16 indexing, including inline'
    elif fn in field or inlined in field or inlined == 'adamic_object_check_data_write':
        bucket = 'Object field and call plumbing, including inline'
    elif fn.startswith('adamic_string_slice') or fn == 'adamic_string_share':
        bucket = 'Substrings'
    elif fn.startswith('adamic_string_') or (file.startswith('string_') and not fn.startswith('adamic_function_')) or fn == 'units_next':
        bucket = 'Other string operations, including token values'
    elif re.search(r'_(Scanner|Speculation)_', fn) or re.search(r'_(isIdentifierStart|isIdentifierPart|isLineBreak|isSpace|arrowAhead|skipTrivia)\b', fn):
        bucket = 'Scanner generated control'
    elif '_Context_mapParents' in fn:
        bucket = 'Parent map generated control'
    elif '_ParseNode_new' in fn or '_Parser_make' in fn:
        bucket = 'Node construction generated control'
    elif file == args.c_source.name and line in line_table:
        bucket = 'Line table generated control'
    elif fn == 'adamic_read_text_file':
        bucket = 'File reading entry self'
    else:
        bucket = 'Remainder: parser, arrays, driver, libc, startup and unattributed inline'
    buckets[bucket] += cost
assert sum(buckets.values()) == j['total'], 'bucket reconciliation failed'
result = dict(total=j['total'], buckets=dict(buckets),
              inline_header_instructions=sum(r['instructions'] for r in j['line_self'] if r['file']=='adamic.h'),
              inclusive=j['inclusive'][:40], header_ref=args.runtime_ref,
              unknown_header_lines=sum(r['instructions'] for r in j['line_self'] if r['file']=='adamic.h' and r['line']==0))
args.output.write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result['buckets'], indent=2))
