"""Audit final evidence invariants; no product implementation or tests are added."""
import pathlib,json,re,os
out=pathlib.Path(__file__).resolve().parent
r=json.loads((out/'REPORT.json').read_text())
assert r['sourceValidation']['mainBytesEqualCanonical'] and r['sourceValidation']['files']==746
for label,count in [('main',0),('A',2),('B',5)]:
 row=r['checks'][label];assert len(row['adamic'])==len(row['solution'])==count
 expected=[(d.get('file'),d.get('line'),d.get('column'),d['code']) for d in row['adamic']]
 actual=[(d.get('file'),d.get('line'),d.get('column'),d['code']) for d in row['solution']]
 assert sorted(expected)==sorted(actual),(label,expected,actual)
assert [r['systemReferenceCounts'][k]['includingDefinition'] for k in ['setTimeout','clearTimeout']]==[18,15]
assert len(r['callers'])==22 and len(r['implementingMembers'])==49
for label in ['A','B']:
 x=r['artifacts'][label];assert x['javascript']['identical'] and x['javascript']['mainCount']==x['javascript']['candidateCount']==9
 assert len(x['partialJavascript'])==8 and not x['typescript.d.ts']['identical'] and not x['typescript.internal.d.ts']['identical']
 assert r['sourceValidation']['candidates'][label]['exactTwoDeclarationEdits']
for label,n in [('main',123),('A',125),('B',102)]:assert r['latentAnalysis'][label]['distinctSites']==n
assert (out/'REPORT.md').read_text().startswith('| Caller')
for p in out.rglob('*'):
 assert p.suffix not in ['.go','.pyc'],str(p)
 if p.suffix in ['.md','.txt','.cjs','.mjs','.py']:
  text=p.read_text();assert '\u2013' not in text and '\u2014' not in text,str(p)
print(json.dumps({'sourceScope':True,'pairedDiagnosticIdentities':True,'apiReferenceCounts':True,'nineJavascriptFiles':True,'declarationDiffs':True,'latentDeduplication':True,'noGoOrPythonCacheArtifacts':True}))
