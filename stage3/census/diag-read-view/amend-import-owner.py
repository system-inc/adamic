from pathlib import Path
import json,hashlib
s=Path('/workspace/scratch/diag-read-view-1');r=s/'repo';binding=r/'stage3/adapt/99-correctness-binding-views';mm=r/'stage3/adapt/99-correctness-core-multimap';d95=r/'stage3/adapt/95-checker-axolotl-h'
h=lambda t:hashlib.sha256(t.encode()).hexdigest()
def add(t):
 assert t.count('    type Diagnostic,')==1
 return t.replace('    type Diagnostic,','    type Diagnostic,\n    type DiagnosticReadView,')
def whole(file):
 t=(s/'main-tree/src/compiler'/file).read_text();return t[:t.index('} from "./_namespaces/ts.js";')+len('} from "./_namespaces/ts.js";')]
def version(before):
 after=add(before);return dict(before=before,after=after,ownerBeforeSha256=h(before),ownerAfterSha256=h(after))
def pin(file,versions):return dict(file=file,owner='namespace import ./_namespaces/ts.js',names=['@import'],beforeCount=1,afterCount=1,**versions[0],**({'variants':versions} if len(versions)>1 else {}))
p=d95/'diagnostic-owners.json';data=json.loads(p.read_text());data=[pin('program.ts',[version(whole('program.ts'))]) if x['names']==['@import'] else x for x in data];p.write_text(json.dumps(data,indent=2)+'\n')
p=binding/'owners.json';data=json.loads(p.read_text());data=[pin('utilitiesPublic.ts',[version(whole('utilitiesPublic.ts'))]) if x.get('file')=='utilitiesPublic.ts' and x.get('names')==['@import'] else x for x in data if not(x.get('file')=='utilities.ts' and x.get('names')==['@import'])];p.write_text(json.dumps(data,indent=2)+'\n')
inputImport=(s/'99-correctness-core-multimap-before-utilities-import.txt').read_text();outputImport=whole('utilities.ts');newpin=pin('utilities.ts',[version(inputImport),version(outputImport)]);(mm/'diagnostic-owners.json').write_text(json.dumps([newpin],indent=2)+'\n')
p=mm/'owners.json';data=json.loads(p.read_text());matched=0
for row in data:
 if row['file']=='src/compiler/utilities.ts' and row['kind']=='import':
  assert row['before']==inputImport and row['after']==outputImport
  row['before']=add(row['before']);row['after']=add(row['after']);row['beforeSha256']=h(row['before']);row['afterSha256']=h(row['after']);matched+=1
assert matched==1;p.write_text(json.dumps(data,indent=2)+'\n')
p=binding/'diagnostic-engine.cjs';t=p.read_text();start=t.index("    if (names[0] === '@import')");end=t.index('\n    return nodes.sort',start);t=t[:start]+"    if (names[0] === '@import') nodes.push(...source.statements.filter(ts.isImportDeclaration).filter(n => n.moduleSpecifier.text === './_namespaces/ts.js'));"+t[end:]
t=t.replace('    const state = spelling === pin.after;', "    const variant = (pin.variants || [pin]).find(v => spelling === v.before || spelling === v.after);\n    assert(variant, `${pin.owner}: owner drift or half adaptation`);\n    const state = spelling === variant.after;")
t=t.replace('state ? pin.after : pin.before','state ? variant.after : variant.before').replace('state ? pin.ownerAfterSha256 : pin.ownerBeforeSha256','state ? variant.ownerAfterSha256 : variant.ownerBeforeSha256').replace('return {start, end, state};','return {start, end, state, variant};').replace('(restoring ? edit.pin.before : edit.pin.after)','(restoring ? edit.variant.before : edit.variant.after)');p.write_text(t)
p=binding/'adapt.cjs';t=p.read_text().replace("...require('../95-checker-axolotl-h/diagnostic-owners.json'),", "...require('../95-checker-axolotl-h/diagnostic-owners.json'), ...require('../99-correctness-core-multimap/diagnostic-owners.json'),");p.write_text(t)
for p in [binding/'README.md',d95/'README.md']:
 t=p.read_text().replace('Diagnostic/DiagnosticReadView import specifiers','complete namespace import').replace('diagnostic type import specifiers','complete namespace import').replace('edited type import specifiers','complete edited import declarations');p.write_text(t)
with (binding/'README.md').open('a') as f:f.write('\nThe utilities.ts namespace import is already pinned whole by\n99-correctness-core-multimap. Its diagnostic-owners.json owns the read-view\nimport amendment at both the binding and final MultiMap boundaries, with exact\nwhole import text/SHA-256 pairs. Binding-views invokes those pins; MultiMap\nretains the type import in its own before/after pins. Replay preserves the\ncurrent known boundary, and restoration removes only the diagnostic type import,\nretaining the baseline MultiMap runtime imports.\n')
with (mm/'README.md').open('a') as f:f.write('\nDiagnostic read-view import amendment (base 8e3a784aa): the whole utilities.ts\nnamespace import before/after pins now retain type DiagnosticReadView.\ndiagnostic-owners.json pins the complete old/new import at both this adaptation\ninput and output. Binding-views invokes these owner pins atomically with the\nother diagnostic edits; no import pin is bypassed. The paired variants preserve\nthe correct MultiMap runtime imports on replay and restoration. No MultiMap\nruntime owner, body or call changes. Proofs: typescript/diag-read-view-evidence.\n')
