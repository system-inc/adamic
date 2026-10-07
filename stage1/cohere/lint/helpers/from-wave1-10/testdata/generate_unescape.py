"""Parse every consumer's Go test strings, then generate independent scanner controls."""
import json,subprocess,itertools,gzip
from pathlib import Path
here=Path(__file__).resolve().parent
root=here.parents[5]
ledger=json.loads((here.parents[1]/'readiness.json').read_text())
symbol='github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.unescapeCSSIdentifier'
rules=[r['rule'] for r in ledger['remaining'] if symbol in r['remaining_helpers']]
files=list((root/'cohere/internal/lint/rules/tailwind').glob('*_test.go'))
# The actual rule descriptor's exported name identifies its upstream test file.
names={'enforce-canonical-classes':'enforce_canonical_classes_test.go','enforce-consistent-class-order':'enforce_consistent_class_order_test.go','enforce-consistent-variant-order':'enforce_consistent_variant_order_test.go','enforce-shorthand-classes':'enforce_shorthand_classes_test.go','no-conflicting-classes':'no_conflicting_classes_test.go','no-unknown-classes':'no_unknown_classes_test.go'}
selected=[root/'cohere/internal/lint/rules/tailwind'/names[r.split('/')[1]] for r in rules]
for f in selected: assert f.exists(), f
scratch=Path('/tmp/wave10-helper-literals.go');scratch.write_text((here/'literals.go.txt').read_text())
binary='/tmp/wave10-helper-literals'
subprocess.run(['go','build','-o',binary,str(scratch)],cwd=root,check=True)
raw=json.loads(subprocess.check_output([binary,*map(str,selected)],cwd=root))
values=[]; coverage=[]
for rule,file in zip(rules,selected):
 strings=raw[str(file)]
 coverage.append({'rule':rule,'file':str(file.relative_to(root)),'literals':len(strings)})
 values.extend(strings)
 # Exercise scanner branches even when the original class string contains no math function.
 values.extend('calc('+s+')' for s in strings)
values.extend('\\'+s for s in list(values))
for point in range(0x110001):
    values.append('\\'+format(point,'x')+' ')
for spelling in ['0','00','000000','0000000','110000','ffffff','d800','dfff','D800','DFFF','10ffff','1f600','abcdef','AbCdEf','a','1234567','fffffff','000041','000041f']:
    for tail in ['', ' ', '\t', '\n', '\r', '\f', '\v', '\u00a0', '  ', '\r\n', 'g', 'f', '\\', '\\41', '😀']:
        values.append('a\\'+spelling+tail+'b')
values.extend(['', '\\', '\\\\', '\\😀', '\\é', '\\g', '\\\n', '\\\r\n', '\\1f600abc', 'a😀\\41b'])
values=list(dict.fromkeys(values))
with gzip.GzipFile(filename=str(here/'unescape_cases.json.gz'),mode='wb',mtime=0) as f: f.write((json.dumps(values,ensure_ascii=True)+'\n').encode())
(here/'unescape_coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
print('cases',len(values),'consumers',len(coverage),'literals',sum(r['literals'] for r in coverage))
