import importlib.util,json,pathlib
spec=importlib.util.spec_from_file_location("runner","/workspace/markdown-defense-run.py");m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
m.records=json.loads((m.P/'runs.json').read_text())
groups=[('baseline-decode-warm','^TestMarkdownSourceDecoding$',False),('baseline-ast-v8','^TestMarkdownASTPreprocessing$',True),('baseline-chunks-v8','^TestMicromarkInputChunks$',True),('baseline-probes-v8','^TestParserRepresentationProbes$',True),('baseline-events-000-v8','^TestTokenizerEvents_000$',True),('baseline-mdast-construction','^TestNativeMdastConstruction$',False),('baseline-mdast-other','^(TestMdastIdentifierWitnesses|TestMdastMalformedEvents(_Setup|Union|_00[0-2])|TestProduct_MarkdownMalformedEvents(Lowered|Native))$',False)]
for name,selector,v8 in groups:
 if v8:
  directory=m.P/'v8'/name;directory.mkdir(parents=True,exist_ok=True);m.env['NODE_V8_COVERAGE']=str(directory)
 else:m.env.pop('NODE_V8_COVERAGE',None)
 r=m.run(name,selector)
 if r['exit'] and not r['cooked']:raise RuntimeError('RED baseline '+name)
