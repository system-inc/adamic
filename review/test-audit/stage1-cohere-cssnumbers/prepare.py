import pathlib,json,subprocess,difflib,re
out=pathlib.Path('review/test-audit/stage1-cohere-cssnumbers');root=pathlib.Path('/tmp/u082');base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip();(out/'base.txt').write_text(base+'\n')
names=[x for x in (root/'discovery.log').read_text().splitlines() if x.startswith('Test')]
groups=[('TestProduct_CSSNumbersGoOracle','^TestProduct_CSSNumbersGoOracle$',[names[0]]),('TestProduct_CSSNumbersGoAnswers','^TestProduct_CSSNumbersGoAnswers$',[names[1]]),('TestProduct_CSSNumbersLowered family','^TestProduct_CSSNumbers(Port|Mutant[012])Lowered$',[x for x in names if x.startswith('TestProduct_') and x.endswith('Lowered')]),('TestProduct_CSSNumbersNative family','^TestProduct_CSSNumbers((Port|Mutant[012])Native|FastNative)$',[x for x in names if x.startswith('TestProduct_') and x.endswith('Native')]),('TestCSSNumbers family','^TestCSSNumbers_[0-9]{3}$',[x for x in names if re.fullmatch('TestCSSNumbers_[0-9]{3}',x)]),('TestCSSNumbersUnion','^TestCSSNumbersUnion$',['TestCSSNumbersUnion']),('TestCSSNumbersPlantedDisagreement','^TestCSSNumbersPlantedDisagreement$',['TestCSSNumbersPlantedDisagreement']),('TestCSSNumbers_Setup','^TestCSSNumbers_Setup$',['TestCSSNumbers_Setup'])]
assert sorted(x for _,_,members in groups for x in members)==sorted(names)
(out/'scope.json').write_text(json.dumps([dict(test=n,pattern=p,members=m) for n,p,m in groups],indent=2)+'\n')
files=['stage1/cohere/cssnumbers/numbers.ts','stage1/cohere/cssnumbers/main.ts','stage1/cohere/cssstrings/strings.ts']
inventory=[]
for f in files:
 s=pathlib.Path(f).read_text()
 for i,line in enumerate(s.splitlines(),1):
  if re.match(r'(export )?function ',line):inventory.append({'file':f,'line':i,'declaration':line})
(out/'functions.json').write_text(json.dumps(inventory,indent=2)+'\n')
file=files[0];s=pathlib.Path(file).read_text();menu=[('M1',"return 'Q';","return 'q';",'change constant'),('M2','code <= 57','code < 57','off-by-one upper digit bound'),('M3','end > dot + 1','end > dot + 2','off-by-one fraction-trimming bound'),('M4','code >= 128','code > 128','off-by-one identifier bound')]
records=[]
for mid,old,new,kind in menu:
 assert s.count(old)==1
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new,1).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 records.append(dict(id=mid,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
(out/'menu.json').write_text(json.dumps(records,indent=2)+'\n')
(out/'code-and-oracle.txt').write_text('CODE UNDER TEST: numbers.ts and the composed cssstrings port executed by main.ts, compiled natively and to JavaScript.\nORACLE: Go cohere actual CSS numeric/string printer, Prettier 3.9.6, byte-for-byte differential checks. No oracle source is mutated.\nThe fixed four-mutant menu was saved before any mutation run. functions.json lists the port and driver functions reachable by the corpus.\nConstruction and planted-disagreement checks receive separate permitted harness probes.\n')
