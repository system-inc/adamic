"""Apply reviewed provenance and write the complete ledger and location report."""
import json
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parent
raw = json.loads(Path(sys.argv[1]).read_text())
overrides = json.loads((ROOT / 'inherited-key-provenance.json').read_text())
used = set()
for row in raw['sites']:
    row['input_sources'] = []
    row['classification_method'] = 'checker_key_domain' if row['classification'] != 'unknown' else 'unknown'
    if row['string_capable'] and row['location_id'] in overrides:
        row.update(overrides[row['location_id']])
        row['classification_method'] = 'reviewed_source'
        used.add(row['location_id'])
assert used == set(overrides), ('Unused provenance locations', set(overrides) - used)
assert not raw['diagnostics'], raw['diagnostics']
counted = [r for r in raw['sites'] if r['counted']]
string = [r for r in counted if r['string_capable']]
raw['totals'] = {
    'all_dynamic_reads_and_membership_checks': len(counted),
    'all_classes': dict(Counter(r['classification'] for r in counted)),
    'string_capable_including_literal_helper_calls': len(string),
    'string_capable_classes': dict(Counter(r['classification'] for r in string)),
    'dynamic_string_keys_only': sum(r['dynamic'] for r in string),
    'non_string_key_domain': sum(not r['string_capable'] for r in counted),
    'by_kind': dict(Counter(r['kind'] for r in counted)),
    'string_by_kind': dict(Counter(r['kind'] for r in string)),
    'for_in_loops': len(raw['loops']),
    'for_in_body_sites': sum(bool(r['for_in']) for r in counted),
    'literal_in_tests_excluded': sum(r['kind'] == 'in' and not r['dynamic'] for r in raw['sites']),
    'non_literal_in_tests': sum(r['kind'] == 'in' and r['counted'] for r in raw['sites']),
    'original_census_sites': sum(r['census_index'] is not None for r in raw['sites']),
    'original_census_dynamic_reads': sum(r['census_index'] is not None and r['counted'] for r in raw['sites']),
    'user_input_sources_overlapping': dict(Counter(s for r in counted if r['classification'] == 'user_input' for s in r['input_sources'])),
    'user_input_source_combinations': dict(Counter(' + '.join(r['input_sources']) for r in counted if r['classification'] == 'user_input')),
}
(ROOT / 'inherited-key-global.json').write_text(json.dumps(raw, indent=2) + '\n')
lines = ['# Global inherited-key census', '',
         'Pinned TypeScript 6.0.3 original compiler code: 77 files. Stock checker: zero diagnostics.', '',
         '**1,041** dynamic element reads and membership checks: **962 fixed key sets/domains excluding prototype names, 36 user input, 43 unknown**.', '',
         'Of these, **145** have string-capable keys (including every helper call, even with a literal key): **66 fixed, 36 user input, 43 unknown**. The other **896** have numeric/bigint/symbol domains that cannot name any of the twelve Object.prototype members.', '',
         'The fully dynamic string-key subset is **130**: **51 fixed, 36 user input, 43 unknown**. Fifteen literal-key hasProperty calls make up the difference.', '',
         'Counts are distinct source operations, not executions or possible panics. A helper call and an operation inside its definition are distinct source sites. A for-in tag is an overlapping context, never an extra row in the total. Nested element accesses sharing a start column have distinct end offsets in JSON.', '',
         '## Requested forms', '',
         '| Form | Sites |', '| --- | ---: |',
         '| Dynamic element reads, including compounds | 993 |',
         '| Of those, string-capable keys | 97 |',
         '| hasProperty calls | 36 on 32 lines |',
         '| getProperty calls | 0 |',
         '| Direct hasOwnProperty.call checks in the core helpers and for-in bodies | 12 |',
         '| Non-literal in tests | 0 |',
         '| Literal in tests, excluded | 10 |',
         '| Actual for-in loops | 27 |',
         '| Read/check sites inside their bodies, already counted above | 57 |',
         '| Original string-index census accesses, fully reconciled | 68 |',
         '| Of those, counted dynamic reads | 45 |', '',
         'The 68 reconciled accesses contain 22 write/delete-only sites and one literal dot-key read. All retain a three-way key classification and their counted flag in JSON. The 45 is now a subset of the global pass, not the answer.', '',
         '## User-input origins', '',
         '| Origin combination | Sites |', '| --- | ---: |']
for sources, count in sorted(raw['totals']['user_input_source_combinations'].items()):
    lines.append(f'| {sources} | {count} |')
lines += ['', 'Source occurrences overlap: tsconfig 18, package.json 15, command line 8, source-code module specifiers 2. The disjoint combinations above sum to 36. A key from user input stays in that class even where subsequent syntax filtering excludes prototype names, or an own-property guard guarantees a safe hit.', '',
          '## String-capable sites', '',
          '| Location | Operation | Class | Input source or evidence |', '| --- | --- | --- | --- |']
for row in string:
    key = row['key'].replace('|', '\\|').replace('\n', ' ')
    evidence = (' + '.join(row['input_sources']) + ': ' if row['input_sources'] else '') + row['evidence']
    evidence = evidence.replace('|', '\\|').replace('\n', ' ')
    lines.append(f"| {row['location_id']} | {row['kind']}: `{key}` | {row['classification']} | {evidence} |")
lines += ['', '## for-in body coverage', '', '| Loop location | Receiver | Counted reads/checks |', '| --- | --- | ---: |']
for loop in raw['loops']:
    count = sum(loop['id'] in row['for_in'] for row in counted)
    lines.append(f"| {loop['id']} | `{loop['expression']}` | {count} |")
lines += ['', '## Non-string dynamic reads', '',
          'These are included in the 1,041 global total and in the fixed-key exclusion class. The class includes statically known non-string domains, not only finite string unions. All locations and key/receiver types appear in inherited-key-global.json. Their keys cannot become constructor, toString, or another listed prototype name under ToPropertyKey.', '',
          '| Location | Key type | Read |', '| --- | --- | --- |']
for row in counted:
    if not row['string_capable']:
        expression = row['expression'].replace('\n', ' ').replace('|', '\\|')
        typ = row['key_type'].replace('|', '\\|')
        lines.append(f"| {row['location_id']} | `{typ}` | `{expression}` |")
lines += ['', '## Scope and assumptions', '',
          'This enumerates executable AST in every original src/compiler TypeScript file, including any/generic/finite mapped receivers and nested accesses. It excludes literal element reads, write/delete-only accesses, generated diagnostic code, static dot-property reads, Map.get calls, comments, and JavaScript stored in emit-helper template strings. Emitted JavaScript belongs to compiled programs, not to dynamic reads performed by tsc itself. The lower grep count of for-in text also includes such strings and comments; the AST finds 27 actual loops.', '',
          'Unknown is a complete classification, not a missing census row: generic core helpers, clone helpers, arbitrary caller-supplied diagnostic keys, environment-name callers, and any-valued object comparisons have no proven closed key origin. Fixed-table review describes the pinned internal tables without external mutation. Key classification does not imply that an inherited miss is reachable: own guards, enumeration, and record construction can make the lookup safe.', '',
          'Reproduce using count-inherited-keys.cjs and summarize-inherited-keys.py as documented in README.md. inherited-key-provenance.json records explicit source-review overrides; inherited-key-global.json includes hashes, excluded sites, all 68 census links, loop contexts, zero stock diagnostics, and global totals.']
report = '\n'.join(lines) + '\n'
if (ROOT / 'own-guard-verdicts.json').exists():
    import runpy
    report += runpy.run_path(str(ROOT / 'render-guards.py'))['render']()
(ROOT / 'INHERITED_KEYS.md').write_text(report)
print(json.dumps(raw['totals'], indent=2))
