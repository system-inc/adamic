import pathlib,json,gzip,collections,re,hashlib,difflib,sys
out=pathlib.Path(__file__).resolve().parents[1];follow=out/'continuation'
def load(path):
 if not path.exists():return json.loads(gzip.decompress(pathlib.Path(str(path)+'.gz').read_bytes()))
 return json.loads(path.read_text())
def stream(path):
 if not path.exists():path=pathlib.Path(str(path)+'.gz')
 if not path.exists():return []
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as src:
  result=[]
  for line in src:
   try:result.append(json.loads(line))
   except ValueError:pass
  return result
roots={'original':'/workspace/scratch/baobab-adapted','original-rules':'/workspace/scratch/baobab-adapted','c':'/workspace/scratch/baobab-road2-c','c-public':'/workspace/scratch/baobab-road2-c-public'}
def normalize(value,label):
 if isinstance(value,str):
  value=value.replace(roots[label],'<tree>')
  if label not in ['original','original-rules']:
   value=re.sub(r'(<tree>/src/compiler/types.ts:)(\d+)(:)',lambda m:m[1]+str(int(m[2])-3 if int(m[2])>=4294 else int(m[2]))+m[3],value)
  return value
 if isinstance(value,list):return [normalize(v,label) for v in value]
 if isinstance(value,dict):return {k:normalize(v,label) for k,v in value.items()}
 return value

def key(value):return json.dumps(value,sort_keys=True)
def delta(old,new):
 a={key(v):v for v in old};b={key(v):v for v in new}
 return {'baseline_unique':len(a),'variant_unique':len(b),'added':[b[k] for k in sorted(b.keys()-a.keys())],'removed':[a[k] for k in sorted(a.keys()-b.keys())]}
result={'variants':{},'relation_scope':'Full cohere:adamic set. Actual CLI summary determines the walked scope. No rule disable or source rewrite.','comparison_normalization':'Replace scratch-root paths and map unchanged types.ts lines after the three-line interface insertion back to baseline. Ignore raw byte offsets and timing fields.'}
base=load(follow/'original/whole-src-adamic-checker.json')
base_exact=load(follow/'original-exact/whole-src-diagnostics.json')
base_cohere=[r for r in stream(follow/'original-rules/cohere-rules.jsonl') if r.get('kind')=='finding']
def cohere_key(r,label):
 r={k:r.get(k) for k in ['rule','path','line','column','severity','message']}
 r=normalize(r,label)
 if label not in ['original','original-rules'] and r['path'].endswith('/src/compiler/types.ts') and r['line']>=4294:r['line']-=3
 return r
baseline_coverage={s['file']:s for s in load(out/'coverage.json')}
names=['scanner.ts']+load(follow/'variant-files.json')
base_findings=[]
for name in names:
 for row in stream(out/baseline_coverage[name]['stream']):
  for f in row.get('findings') or []:
   base_findings.append(dict(root=name,**normalize({k:f.get(k) for k in ['unit','phase','kind','where','reason','text']},'original')))
site_comparison=load(follow/'c-site-comparison.json')
for label in (sys.argv[1:] or ['c','c-public']):
 own=load(follow/label/'whole-src-adamic-checker.json')
 checker=delta(normalize(base['diagnostics'],'original'),normalize(own['diagnostics'],label))
 exact_label='c-exact' if label=='c' else 'c-public-exact'
 diagnostic_key=lambda d:{k:d.get(k) for k in ['file','line','column','code','category','message']}
 exact=delta([diagnostic_key(normalize(d,'original')) for d in base_exact],[diagnostic_key(normalize(d,label)) for d in load(follow/exact_label/'whole-src-diagnostics.json')])
 cohere_rows=stream(follow/label/'cohere-rules.jsonl')
 cohere=delta([cohere_key(r,'original-rules') for r in base_cohere],[cohere_key(r,label) for r in cohere_rows if r.get('kind')=='finding'])
 cohere_summary=next(r for r in cohere_rows if r.get('kind')=='summary')
 statuses=[];findings=[];units=[];headers=[]
 for name in names:
  stem=name.replace('/','__').removesuffix('.ts');status=load(follow/label/'raw'/(stem+'.status.json'));rows=stream(follow/label/'raw'/(stem+'.jsonl'))
  status['complete_stream']=status['child_exit']==0 and bool(rows) and rows[-1].get('status')=='relation_report'
  statuses.append(status)
  if rows:headers.append(dict(file=name,checker_rejected=rows[0].get('checker_rejected'),diagnostics=rows[0].get('diagnostics')))
  for row in rows:
   for f in row.get('findings') or []:findings.append(dict(root=name,**normalize({k:f.get(k) for k in ['unit','phase','kind','where','reason','text']},label)))
   units.extend(row.get('units') or [])
 lower=delta(base_findings,findings)
 own_scanner=[];own_units=[]
 for row in stream(follow/label/'raw/scanner.jsonl'):
  own_scanner.extend(f for f in row.get('findings') or [] if f.get('unit','').endswith('/scanner.ts:503:1'))
  own_units.extend(u for u in row.get('units') or [] if u.get('where','').endswith('/scanner.ts:503:1'))
 views=[r for r in site_comparison if r['variant']==label]
 helper=[r for r in views if r['callee'] in ['getLineStarts','getLineAndCharacterOfPosition']]
 result['variants'][label]={'stock':load(follow/label/'summary.json'),'stock_exact':load(follow/exact_label/'summary.json'),'adamic_whole_src_diagnostics':len(own['diagnostics']),'adamic_whole_src_delta':checker,'stock_exact_delta':exact,'cohere_summary':cohere_summary,'cohere_delta':cohere,'lowering_delta':lower,'files':statuses,'headers':headers,'view_verdicts':dict(collections.Counter(r['verdict'] for r in views)),'helper_view_verdicts':dict(collections.Counter(r['verdict'] for r in helper)),'core_first_stop':stream(follow/label/'core-first-stop.jsonl')[0],'getLineStarts_findings':own_scanner,'getLineStarts_units':own_units}
 (follow/label/'getLineStarts-unit.json').write_text(json.dumps({'findings':own_scanner,'units':own_units},indent=2)+'\n')
# Verify every src file, including the public-marker-only rerun, against the frozen input.
base_root=pathlib.Path(roots['original'])
for label in (sys.argv[1:] or ['c','c-public']):
 root=pathlib.Path(roots[label]);diff={}
 for p in sorted((base_root/'src').rglob('*')):
  if not p.is_file():continue
  rel=p.relative_to(base_root);other=root/rel;assert other.is_file(),str(rel)
  a,b=p.read_bytes(),other.read_bytes()
  if a!=b:diff[str(rel)]={'before':hashlib.sha256(a).hexdigest(),'after':hashlib.sha256(b).hexdigest()}
 assert set(diff)=={'src/compiler/types.ts','src/compiler/scanner.ts'},diff
 assert not [p for p in (root/'src').rglob('*') if p.is_file() and not (base_root/p.relative_to(root)).exists()]
 result['variants'][label]['source_diff']=diff
(follow/'c-analysis.json').write_text(json.dumps(result,indent=2)+'\n')
for label,r in result['variants'].items():
 print(json.dumps({'label':label,'complete_files':sum(s['complete_stream'] for s in r['files']),'views':r['view_verdicts'],'helper_views':r['helper_view_verdicts'],'native_checker_errors':r['adamic_whole_src_diagnostics'],'native_added':len(r['adamic_whole_src_delta']['added']),'native_removed':len(r['adamic_whole_src_delta']['removed']),'exact_added':len(r['stock_exact_delta']['added']),'cohere_added':len(r['cohere_delta']['added']),'cohere_removed':len(r['cohere_delta']['removed']),'lower_added':len(r['lowering_delta']['added']),'lower_removed':len(r['lowering_delta']['removed'])}))
