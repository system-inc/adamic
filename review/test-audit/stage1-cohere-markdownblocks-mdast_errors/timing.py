import pathlib,json,subprocess,time,re
root=pathlib.Path('/workspace/adamic');ev=root/'review/test-audit/stage1-cohere-markdownblocks-mdast_errors';scope=json.loads((ev/'scope.json').read_text());groups=[]
for n in scope:
 if n.startswith('TestWholeDocumentOraclePreflight_') or n=='TestWholeDocumentOraclePreflightUnion':continue
 if re.match(r'TestMarkdownQuoteLayout_\d+$',n)or n=='TestMarkdownQuoteLayoutUnion':continue
 groups.append(dict(test=n,members=[n]))
groups.append(dict(test='TestWholeDocumentOraclePreflight family',members=[n for n in scope if re.match(r'TestWholeDocumentOraclePreflight_\d+$',n)or n=='TestWholeDocumentOraclePreflightUnion']))
groups.append(dict(test='TestMarkdownQuoteLayout family',members=[n for n in scope if re.match(r'TestMarkdownQuoteLayout_\d+$',n)or n=='TestMarkdownQuoteLayoutUnion']))
(ev/'groups.json').write_text(json.dumps(groups,indent=2));records=[]
for group in groups:
 for i in range(1,4):
  label=group['test'].replace(' family','Family')+'-'+str(i);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^('+'|'.join(group['members'])+')$'];start=time.monotonic()
  with(ev/(label+'.log')).open('w')as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(id=label,test=group['test'],command=' '.join(cmd),exit=r.returncode,wall=time.monotonic()-start));(ev/'timing-commands.json').write_text(json.dumps(records,indent=2))
  if r.returncode!=0:print('CLEAN FAILURE OR TIMEOUT',label,r.returncode,flush=True);break
