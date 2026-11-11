import pathlib,json,gzip
p=pathlib.Path(__file__).resolve().parent
labels=['original','c-public','c-public-unstripped','c-public-readonly-sdk-control']
def rows(path):
 if not path.exists():path=pathlib.Path(str(path)+'.gz')
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as src:
  result=[]
  for line in src:
   try:result.append(json.loads(line))
   except ValueError:pass
  return result
result={'scope':'Independent emitted-public-SDK caller, separate from the 34 original compiler views','conclusion':'No tested exact (c) public API meets no-new-refusal-anywhere: optional cache lineMap is visible in LineMapSource but stripped from public SourceFile. A readonly slot alone does not repair hiding.','variants':{}}
for label in labels:
 directory=p/label;cohere=rows(directory/'public-adamic-cohere.jsonl');low=rows(p/('sdk-'+label)/'raw/consumer.jsonl')
 status=json.loads((p/('sdk-'+label)/'raw/consumer.status.json').read_text());assert status['child_exit']==0 and low[-1]['status']=='relation_report'
 result['variants'][label]={'checker':json.loads((directory/'sdk-adamic-checker.json').read_text()),'cohere_findings':[r for r in cohere if r.get('kind')=='finding'],'cohere_summary':next(r for r in cohere if r.get('kind')=='summary'),'guarded_header':low[0],'guarded_findings':[f for r in low for f in r.get('findings') or []],'guarded_status':status,'source':(directory/'public-adamic-consumer.ts').read_text()}
(p/'c-sdk-analysis.json').write_text(json.dumps(result,indent=2)+'\n')
for label,r in result['variants'].items():print(json.dumps({'label':label,'checker_errors':len(r['checker']['diagnostics']),'cohere_findings':len(r['cohere_findings']),'guarded_refusals':[f['reason'] for f in r['guarded_findings'] if f['kind']=='Refused']}))
