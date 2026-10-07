"""Source-derived inputs and exhaustive local flag grammar, not rule finding replay."""
from pathlib import Path
import gzip,itertools,json,subprocess,tempfile
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
RULES=ROOT/'cohere/internal/lint/rules'
families={
 '@next/next/no-html-link-for-pages':['next/no_html_link_for_pages_test.go'],
 '@typescript-eslint/no-empty-object-type':['typescript/no_empty_object_type_test.go','typescript/no_empty_object_type_corpus_test.go','typescript/no_empty_object_type_corpus_data_test.go'],
 'no-restricted-exports':['core/no_restricted_exports_test.go'],
 'no-restricted-imports':['core/no_restricted_imports_test.go','core/no_restricted_imports_matcher_test.go'],
}
paths=[str(RULES/file) for files in families.values() for file in files]
with tempfile.TemporaryDirectory(prefix='wave10-modifier-literals-') as directory:
 entry=Path(directory)/'main.go';entry.write_text((HERE/'literals.go.txt').read_text())
 binary=Path(directory)/'literals'
 subprocess.run(['go','build','-o',str(binary),str(entry)],cwd=ROOT,check=True)
 with (HERE/'modifier-literals.log').open('wb') as out:subprocess.run([str(binary),*paths],cwd=ROOT,stdout=out,check=True)
values=json.loads((HERE/'modifier-literals.log').read_text())
cases=set(['','(','(?','(?:','(?-:','(?i-:','(?i-m:body)', '(?ims-s:','(?i--m:', '(?i:😀', '(?😀i:', '(?i\x00:', '(?\ud800:', '(?i:\ud800'])
coverage=[]
for rule,files in families.items():
 count=0
 for file in files:
  literals=values[str(RULES/file)];count+=len(literals)
  for literal in literals:
   cases.add(literal)
   start=literal.find('(?')
   while start>=0:
    cases.add(literal[start:]);start=literal.find('(?',start+2)
 coverage.append({'rule':rule,'files':files,'literals':count})
for size in range(8):
 for chars in itertools.product('ims-',repeat=size):
  text=''.join(chars)
  cases.update(['(?'+text+':','(?'+text+':)payload','(?'+text])
for bad in ['u','g','I',' ', '\t','\n','\x00','é','😀','\udfff']:
 for text in ['','i','-m','ims-']:
  cases.add('(?'+text+bad+':')
encoded=json.dumps(sorted(cases),ensure_ascii=True,separators=(',',':')).encode()
(HERE/'modifier-cases.json.gz').write_bytes(gzip.compress(encoded,mtime=0))
(HERE/'modifier-coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
(HERE/'modifier-literals.log').unlink()
print('sources',len(cases),'queries',len(cases)*16,'json bytes',len(encoded),'consumer literals',sum(r['literals'] for r in coverage))
