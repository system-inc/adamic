import concurrent.futures,json,os,re,subprocess,time
from pathlib import Path
rows=[('ADAMIC_TYPESCRIPT_SOURCE','stage1/typescript/parser','TestWholeCompilerAgrees'),('ADAMIC_CSS_LIBRARY ADAMIC_CSS_FIXTURES','stage1/cohere/css','TestThePortParsesAsGoCohereDoes/PostCSS'),('ADAMIC_CSS_PRINTER_LIBRARY','stage1/cohere/css','TestCSSPrinterBoundaryProofs'),('ADAMIC_JSON_PRETTIER','stage1/cohere/json','TestUpstreamNumericSeparatorGap'),('ADAMIC_GRAPHQL_LIBRARY','stage1/cohere/graphql','TestThePortParsesAsGoCohereDoes/as_graphql-js'),('ADAMIC_MEDIA_QUERY_LIBRARY','stage1/cohere/mediaquery','TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser'),('ADAMIC_SELECTOR_LIBRARY','stage1/cohere/selector','TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars'),('ADAMIC_VALUES_LIBRARY','stage1/cohere/values','TestThePortParsesAsGoCohereDoes/as_postcss-values-parser'),('ADAMIC_GITIGNORE_LARGEST','stage1/cohere/gitignore','TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower'),('ADAMIC_CLANG_TSGO_ARCHIVE','internal/native','TestSplitTSGoAgrees'),('ADAMIC_GRAPHQL_PRETTIER','stage1/cohere/graphql/printer','TestPrinterUpstreamPreflight'),('ADAMIC_ESTREE_LIBRARY','stage1/cohere/estree','TestOriginalLibraries'),('ADAMIC_YAML_LIBRARY','stage1/cohere/yaml','TestBundledParserDifference'),('ADAMIC_TS_PRETTIER','stage1/cohere/tsprinter','TestStatementUpstreamDifferences'),('ADAMIC_CYCLE_LEDGER_ROOT ADAMIC_CYCLE_LEDGER_OUTPUT','internal/lower','TestOriginalCycleLedger'),('ADAMIC_CSSNUMBERS_LIBRARY','stage1/cohere/cssnumbers','TestCSSNumbers'),('ADAMIC_CSSSTRINGS_LIBRARY','stage1/cohere/cssstrings','TestCSSStrings'),('ADAMIC_MARKDOWNINLINE_LIBRARY','stage1/cohere/markdowninline','TestMarkdownInline')]
out=Path('/tmp/gate-inputs-proof');out.mkdir(exist_ok=True)
env=dict(os.environ)
remove=set(' '.join(x[0] for x in rows).split())
def run(row):
 variables,package,name=row
 pattern='/'.join('^'+re.escape(p)+'$' for p in name.split('/'))
 args=['go','test','./'+package,'-run',pattern,'-v','-count=1','-timeout=20m']
 records=[]
 for mode in ['without','with']:
  e=dict(env)
  if mode=='without':
   for v in remove:e.pop(v,None)
  start=time.monotonic()
  p=subprocess.run(args,env=e,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=1260,text=True)
  log=out/(package.replace('/','-')+'-'+name.replace('/','-')+'-'+mode+'.log');log.write_text(p.stdout)
  lines=p.stdout.splitlines()
  record=dict(variables=variables,package=package,test=name,mode=mode,command=' '.join(args),exit=p.returncode,seconds=round(time.monotonic()-start,2),runs=[s for s in lines if s.startswith('=== RUN')],verdicts=[s for s in lines if re.match(r'\s*--- (PASS|FAIL|SKIP):',s)],first_failure=None if p.returncode == 0 else next((s.strip() for s in lines if re.search(r'\.go:\d+:',s) and not 'Skip' in s),None),log=str(log))
  records.append(record);print(variables,mode,p.returncode,record['verdicts'][-3:],flush=True)
 return records
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:records=sum(list(pool.map(run,rows)),[])
(out/'results.json').write_text(json.dumps(records,indent=2)+'\n')
