import pathlib,subprocess,json,time,re
p=pathlib.Path('/tmp/u129/evidence');src=pathlib.Path('stage1/cohere/markdownblocks/setup_deadline_markdownblocks_test.go').read_text();names=re.findall(r'func (Test\w+)\(',src)
groups=[('TestMarkdownLayoutSetupHasNoDeadline',[names[0]]),('TestMarkdownLayoutDeadlineWorker',[names[1]]),('TestProduct_MarkdownQuote_build family',names[2:5]),('TestProduct_MarkdownQuote_native family',names[5:7]),('TestProduct_MarkdownStructure family',names[7:11]),('TestProduct_MarkdownTable family',names[11:16]),('TestProduct_MarkdownWhitespace_build family',names[16:19]+[names[21]]),('TestProduct_MarkdownWhitespace_native family',names[19:21]),('TestProduct_MarkdownWhitespace_policy_native',[names[22]]),('TestProduct_MarkdownWhitespace_manifest',[names[23]])]
(p/'groups.json').write_text(json.dumps(groups,indent=2));rs=[]
for index,(group,members) in enumerate(groups):
 for run in range(1,4):
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^('+'|'.join(members)+')$'];s=time.monotonic()
  with (p/f'group-{index}-timing-{run}.log').open('w') as f:r=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT)
  item={'group':group,'index':index,'run':run,'exit':r.returncode,'seconds':time.monotonic()-s,'command':command};rs.append(item);(p/'group-timings.json').write_text(json.dumps(rs,indent=2));print(item,flush=True)
  if r.returncode:
   events=[json.loads(l) for l in (p/f'group-{index}-timing-{run}.log').read_text().splitlines()]
   if not any('test timed out after' in e.get('Output','') for e in events):raise SystemExit('RED BASELINE, stop')
   break
