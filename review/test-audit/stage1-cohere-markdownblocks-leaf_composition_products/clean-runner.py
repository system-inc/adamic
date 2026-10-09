import subprocess,time,json,pathlib
out=pathlib.Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products')
groups=[('products-go','TestProduct_MarkdownLeafGo(Lists|DocLayout)'),('product-lowered','TestProduct_MarkdownLeafLowered'),('products-native','TestProduct_MarkdownLeafNative(Sanitized|Release)'),('list-setup','TestMarkdownListLayout_Setup'),('list-family','TestMarkdownListLayout(Union|_00[0-9]|_01[0-5])'),('list-legacy','TestMarkdownListLayout'),('quote','TestMarkdownQuoteLayout'),('table','TestMarkdownTableLayout'),('code','TestMarkdownCodeBlockLayout'),('html','TestMarkdownHTMLBlockLayout'),('leaf','TestMarkdownLeafComposition'),('root','TestMarkdownRootLayout'),('structure-setup','TestMarkdownStructureLayout_Setup'),('malformed-family','TestMdastMalformedEvents(_00[0-2]|Union)')]
results=[]
for name,pattern in groups:
 for i in range(3):
  start=time.monotonic();cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^('+pattern+')$']
  with (out/f'clean-{name}-{i+1}.log').open('w') as f:p=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
  records=[]
  for line in (out/f'clean-{name}-{i+1}.log').read_text().splitlines():
   try:records.append(json.loads(line))
   except:pass
  errors=[r for r in records if r.get('OutputType')=='error']; cooked=any('test timed out' in r.get('Output','') for r in records)
  results.append({'group':name,'round':i+1,'command':cmd,'wall':time.monotonic()-start,'exit':p.returncode,'cooked':cooked,'errors':errors})
  (out/'clean-runs.json').write_text(json.dumps(results,indent=2))
  if errors and not cooked: raise SystemExit('RED CLEAN BASELINE '+name)
  if cooked:break
