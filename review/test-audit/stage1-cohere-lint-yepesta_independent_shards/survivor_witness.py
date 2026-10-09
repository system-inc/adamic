import pathlib,subprocess,json,shutil,os,time
out=pathlib.Path('review/test-audit/stage1-cohere-lint-yepesta_independent_shards');plan=json.loads((out/'frozen-plan.json').read_text());work=pathlib.Path('/tmp/u113-witness');work.mkdir(exist_ok=True);results=[]
js="""const {jsxNoCommentTextnodes}=await import(process.argv[2]);const answer=[];for(const raw of ['// comment','a\\n// comment','//']){const findings=[];const context={source:raw,kind:()=> 'JsxText',node:()=>({pos:0,end:raw.length}),reportRange:(...args)=>findings.push(args)};jsxNoCommentTextnodes(context,0);answer.push({raw,findings});}console.log(JSON.stringify(answer));"""
(work/'stub.mjs').write_text(js)
f='stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts';original=subprocess.check_output(['git','show','HEAD:'+f],text=True)
for m in plan['mutants'][:3]:
 values=[]
 for side,s in [('before',original),('after',original.replace(m['from'],m['to']))]:
  d=work/(m['id']+'-'+side);d.mkdir(exist_ok=True);(d/'rule.ts').write_text(s);shutil.copyfile('stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/messages.ts',d/'messages.ts')
  cmd=['node',str(work/'stub.mjs'),str(d/'rule.ts')]
  with (out/(m['id']+'-witness-'+side+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  values.append((out/(m['id']+'-witness-'+side+'.log')).read_text())
 results.append({'id':m['id'],'command':'node /tmp/u113-witness/stub.mjs /tmp/u113-witness/'+m['id']+'-{before,after}/rule.ts','before':values[0].strip(),'after':values[1].strip(),'different':values[0]!=values[1]})
cache=pathlib.Path('/home/agent/.cache/adamic-build');bundles=sorted([f for f in cache.rglob('bundle.json') if 'Witnesses' in json.loads(f.read_text())],key=lambda f:f.stat().st_mtime);f=bundles[0];bundle=json.loads(f.read_text());rows=[]
for row in bundle['Rows']:
 fields=row.split('\t');fields[0]=str(f.parent/fields[0]);rows.append('\t'.join(fields))
manifest=work/'manifest.txt';manifest.write_text('\n'.join(rows)+'\n');origport=pathlib.Path(bundle['Port']);newport=work/'unsorted';shutil.copytree(origport,newport,dirs_exist_ok=True);lint=newport/'lint.ts';lint.write_text(lint.read_text().replace('        this.findings.sort(compareFindings);',''))
values=[]
for side,port in [('before',origport),('after',newport)]:
 cmd=['timeout','90','node','oracle/node.mjs',str(port/'main.ts'),'--manifest',str(manifest)]
 with (out/('M4-witness-'+side+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 values.append((out/('M4-witness-'+side+'.log')).read_text())
left,right=values[0].splitlines(),values[1].splitlines();diff=next((f'line {i+1}: before {a!r}; after {b!r}' for i,(a,b) in enumerate(zip(left,right)) if a!=b),None)
results.append({'id':'M4','command':'timeout 90 node oracle/node.mjs <baseline cached port or /tmp/u113-witness/unsorted>/main.ts --manifest /tmp/u113-witness/manifest.txt','different':values[0]!=values[1],'first_difference':diff})
(out/'survivor-witnesses.json').write_text(json.dumps(results,indent=2)+'\n')
