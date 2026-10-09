import importlib.util,json,pathlib,subprocess
spec=importlib.util.spec_from_file_location('runner','/workspace/markdown-defense-run.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
m.records=json.loads((m.P/'runs.json').read_text());plan=json.loads((m.P/'plan.json').read_text())
for wanted in ['TestMarkdownASTPreprocessing','TestMicromarkInputChunks','TestMarkdownSourceDecoding','TestParserRepresentationProbes','TestNativeMdastConstruction']:
 assert any(r['exit']==0 and r['states'].get(wanted)=='pass' for r in m.records), 'no green baseline '+wanted
mdast=['TestMdastIdentifierWitnesses','TestMdastMalformedEvents_Setup','TestMdastMalformedEventsUnion']+['TestMdastMalformedEvents_00'+str(i) for i in range(3)]+['TestProduct_MarkdownMalformedEventsLowered','TestProduct_MarkdownMalformedEventsNative']
def selector(names):return '^('+'|'.join(names)+')$'
def matrix(name,names,id):
 r=m.run(name,selector(names),mutant=id)
 if r['cooked'] or r['exit']==124:
  missing=[n for n in names if r['states'].get(n) not in ['pass','skip','fail']]
  (m.P/(name+'-unfinished.json')).write_text(json.dumps(missing,indent=2))
  for i,n in enumerate(missing):m.run(name+'-recovery-'+str(i),'^'+n+'$',mutant=id)
 return r
for d in plan:
 original=(m.R/d['file']).read_text();assert original.count(d['old'])==1
 try:
  (m.R/d['file']).write_text(original.replace(d['old'],d['new'],1))
  if d['id']=='D1':matrix('D1-matrix',['TestMarkdownASTPreprocessing','TestParserRepresentationProbes'],'D1')
  if d['id']=='D2':
   products=['TestProduct_TokenizerEvents'+s for s in ['Go','Lowered','NativeSanitized','NativeRelease']]
   matrix('D2-products',products,'D2')
   matrix('D2-chunks',['TestMicromarkInputChunks','TestParserRepresentationProbes'],'D2')
   for group in range(4):
    names=['TestTokenizerEvents_'+str(i).zfill(3) for i in range(group*128,(group+1)*128)]
    if group==0:names.insert(0,'TestTokenizerEventsUnion')
    matrix('D2-events-'+str(group),names,'D2')
  if d['id']=='D3':
   matrix('D3-decode',['TestMarkdownSourceDecoding','TestParserRepresentationProbes'],'D3')
   matrix('D3-mdast-construction',['TestNativeMdastConstruction'],'D3')
   matrix('D3-mdast-other',mdast,'D3')
 finally:(m.R/d['file']).write_text(original)
subprocess.run(['git','diff','--exit-code','--']+[d['file'] for d in plan],cwd=m.R,check=True)
