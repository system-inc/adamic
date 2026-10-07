import os, subprocess, json, time, pathlib, re, collections, difflib
D=pathlib.Path('/workspace/gate-logs'); sha='eac1fb47cc64c97f452be3b21ed1eab8b31ca29d'
layout=['TestMarkdownListLayout','TestMarkdownQuoteLayout','TestMarkdownTableLayout','TestMarkdownCodeBlockLayout','TestMarkdownHTMLBlockLayout','TestMarkdownWhitespaceLayout','TestMarkdownLeafComposition','TestMarkdownRootLayout','TestMarkdownStructureLayout']
names=[x for x in (D/'markdownblocks-list.log').read_text().splitlines() if x.startswith('Test')]
groups={'layoutOnce':layout}
for n in names:
 if n not in layout: groups[n]=[n]
(D/'markdownblocks-groups.json').write_text(json.dumps(groups,indent=2)+'\n')
union=[n for g in groups.values() for n in g]
assert collections.Counter(union)==collections.Counter(names) and len(set(union))==len(union)
(D/'markdownblocks-coverage.diff').write_text(''.join(difflib.unified_diff(sorted(names),sorted(union),fromfile='go-test-list',tofile='group-union')))
(D/'markdownblocks-coverage.txt').write_text(f'{len(names)} tests; exact union; each occurs once; empty coverage diff. layoutOnce group: '+', '.join(layout)+'\n')
packages=subprocess.check_output(['go','list','./...'],text=True).splitlines()
extras=list(D.glob('*'))+[D/(p.replace('/','_')+'.solo.jsonl') for p in packages]+[D/('markdownblocks-'+g+'.jsonl') for g in groups]+[D/'runs.json',D/'report.txt']
collector=open(D/'collector.log','w')
c=subprocess.Popen(['bash','/workspace/merge-tree/cloud/integration/gate-logs.sh',sha,str(D/'whole.jsonl')]+[str(p) for p in extras if p.is_file() or p.suffix in ['.jsonl','.json','.txt']],stdout=collector,stderr=subprocess.STDOUT,cwd='/workspace/adamic')
runs=[]
def run(label,args,header=False):
 path=D/(label+'.jsonl'); start=time.monotonic()
 with path.open('w') as out, (D/(label+'.stderr')).open('w') as err:
  if header:
   for key,cmd in [('SHA',['git','rev-parse','HEAD']),('Node',['node','-v']),('Go',['go','version'])]:
    out.write(json.dumps({'Action':'output','Output':key+': '+subprocess.check_output(cmd,text=True)})+'\n')
   out.write(json.dumps({'Action':'output','Output':'ADAMIC_GATE_UNCACHED='+os.environ['ADAMIC_GATE_UNCACHED']+'\n'})+'\n');out.flush()
  rc=subprocess.call(['go','test','-count=1','-timeout','60m','-json']+args,stdout=out,stderr=err)
 elapsed=time.monotonic()-start
 counts=collections.Counter(); outputs=collections.defaultdict(list); failed=[]; skipped=[]; pkgoutputs=collections.defaultdict(list); caches=[]
 for line in path.read_text().splitlines():
  try: e=json.loads(line)
  except ValueError: continue
  p=e.get('Package',''); n=e.get('Test'); a=e.get('Action'); o=e.get('Output','')
  if o:
   pkgoutputs[p].append(o)
   if n: outputs[(p,n)].append(o)
   if 'gate cache:' in o: caches.append(o.rstrip())
  if n and a in ('pass','fail','skip'):
   counts[a]+=1
   if a=='fail':failed.append((p,n))
   if a=='skip':skipped.append((p,n))
 record={'label':label,'args':args,'exit':rc,'seconds':elapsed,'counts':dict(counts),'failed':failed,'skipped':skipped,'cache':caches}
 runs.append(record);(D/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
 print(json.dumps(record),flush=True)
 return record,pkgoutputs,outputs
os.environ['ADAMIC_GATE_UNCACHED']='1'
whole,pout,_=run('whole',['./...'])
resource=[p for p,o in pout.items() if re.search(r'panic: test timed out|signal: killed|no space left on device|exit -1|SIGKILL', ''.join(o),re.I)]
for p in resource:
 solo,output,_=run(p.replace('/','_')+'.solo',[p],True)
 if 'panic: test timed out' in ''.join(output.get(p,[])):
  if p.endswith('/stage1/cohere/markdownblocks'):
   for g,ns in groups.items():run('markdownblocks-'+g,['-run','^('+'|'.join(ns)+')$',p],True)
  else:
   listed=subprocess.check_output(['go','test','-list','.',p],text=True)
   (D/(p.replace('/','_')+'.list.txt')).write_text(listed)
   for n in listed.splitlines():
    if n.startswith('Test'):run(p.replace('/','_')+'.'+n,['-run','^'+n+'$',p],True)
(D/'whole.jsonl.done').touch()
c.wait()
