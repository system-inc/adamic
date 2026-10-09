from pathlib import Path
import os,json,subprocess,time
root=Path('/workspace/adamic');out=root/'review/test-defend/stage1-typescript-parser-compiler_expressions_shards'
while not Path('/tmp/defend-parser/port-done').exists():time.sleep(1)
names=[l.strip() for l in Path('/tmp/defend-parser/list.log').read_text().splitlines() if l.startswith('Test')];seen={}
for p in Path('/tmp/defend-parser').glob('D3-group*.log'):
 for l in p.read_text().splitlines():
  try:e=json.loads(l)
  except:continue
  n=e.get('Test','')
  if n in names and e.get('Action') in ['pass','skip']:seen[n]=e['Action']
  if n=='TestGeneratedExpressionsAgree' and e.get('Action')=='fail':seen[n]='fail'
pending=[n for n in names if n not in seen];(out/'D3-pending-before-rerun.json').write_text(json.dumps(pending,indent=2));serial=[n for n in pending if n in ['TestJsxNameBoundaryRejections','TestJsxScannerMutants']];parallel=[n for n in pending if n not in serial]
base=os.environ.copy();base['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u030/typescript';base['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-parser/cache/D3';p=root/'stage1/typescript/parser/parser.ts';s=p.read_text();old="new ParseNode('PrivateIdentifier', old.pos, old.end, [])";new="new ParseNode('Identifier', old.pos, old.end, [])";assert s.count(old)==1;results=[]
try:
 p.write_text(s.replace(old,new))
 for i,rows in enumerate(([parallel] if parallel else [])+[[n] for n in serial]):
  regex='^('+'|'.join(rows)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run',regex];start=time.monotonic()
  with open('/tmp/defend-parser/D3-rerun'+str(i)+'.log','w') as f:r=subprocess.run(cmd,env=base,stdout=f,stderr=subprocess.STDOUT)
  results.append(dict(rows=rows,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd)));(out/'D3-adaptive-runs.json').write_text(json.dumps(results,indent=2));print(i,len(rows),r.returncode,results[-1]['wall'],flush=True)
finally:p.write_text(s)
Path('/tmp/defend-parser/adaptive-done').write_text('done')
