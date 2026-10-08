"""Reproducibly select whole no-parameter, literal-only function declarations.
Functions are lifted from their original files; no identifiers need a checker in this subset.
The manifest records every occurrence, including identical functions in different tests.
"""
import pathlib, re, json
root = pathlib.Path(__file__).resolve().parents[4]
hir = root/'cohere/internal/lint/ecmascript/high_level_intermediate_representation'
groups = [ ('hir', list(hir.glob('*_test.go'))),
 ('upstream', list((root/'cohere/internal/lint/rules/react/conformance/testdata/fixtures').glob('*'))),
 ('stage1', [p for p in (root/'stage1').rglob('*') if p.is_file() and ('testdata' in p.parts or p.name.endswith('_test.go')) and ('jsx' in str(p).lower() or 'react' in str(p).lower())]) ]
pattern = re.compile(r'function\s+([A-Za-z_$][\w$]*)\s*\(\s*\)\s*\{([^{}]*)\}')
literal = r'(?:[0-9]+(?:\.[0-9]+)?|true|false|null)'
body = re.compile(r'\s*(?:(?:'+literal+r')\s*;\s*|;\s*)*(?:return\s*(?:'+literal+r')?\s*;?\s*)?')
cases=[]; manifest=[]
for group, paths in groups:
 for path in sorted(paths):
  if path.suffix in ['.gz','.md','.json'] or path.stat().st_size>2000000: continue
  try: text=path.read_text()
  except (UnicodeError, OSError): continue
  for match in pattern.finditer(text):
   if not body.fullmatch(match[2]): continue
   code=match[0].replace('\n',' ').replace('\r',' ')
   cases.append(code)
   manifest.append({'group':group,'path':str(path.relative_to(root)),'line':text[:match.start()].count('\n')+1,'source':code})
# Additional probes exercise straight-line expression statements and each literal kind.
for code in ['function Empty() {}','function NumberValue() { 1; 2; return 3; }','function useBoolean() { false; return true; }','function Nil() { return null; }','function Bare() { return; }']:
 cases.append(code);manifest.append({'group':'probe','source':code})
destination=pathlib.Path(__file__).parent
(destination/'corpus.txt').write_text('\n'.join(cases)+'\n')
(destination/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print({group:sum(row['group']==group for row in manifest) for group in ['hir','upstream','stage1','probe']})
