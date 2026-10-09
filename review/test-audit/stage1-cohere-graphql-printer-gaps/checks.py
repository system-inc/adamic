from pathlib import Path
import subprocess,json,difflib,time,os
p=Path('/workspace/adamic/review/test-audit/stage1-cohere-graphql-printer-gaps'); root=Path('/workspace/adamic'); tmp=Path('/workspace/u097-tmp/verify');tmp.mkdir(exist_ok=True)
base=lambda f:subprocess.check_output(['git','show','HEAD:'+f],text=True)
def diff(f,a,b):return ''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
checks=json.loads((p/'check-plan.json').read_text())
for c in checks:
 f=c['file'];a=base(f)
 if c['id']=='W01':
  start=a.index('func printerShardDisagreement('); body=a.index('{',start);end=a.index('\nfunc TestPrinterShardUnion',body);b=a[:body+1]+'\n\treturn nil\n}\n'+a[end:];c['line']=a[:start].count('\n')+1
 elif c['id']=='S01':
  old='make([]printerShard, testPrinterWhitespaceGapShards)';b=a.replace(old,'make([]printerShard, testPrinterWhitespaceGapShards+1)');c['line']=a[:a.index(old)].count('\n')+1
 elif c['id']=='S02':
  old='\tfor _, item := range cases {\n\t\towner := whitespaceOwner(item.id)\n\t\tshards[owner].cases = append(shards[owner].cases, item)\n\t}\n';assert a.count(old)==1;b=a.replace(old,'');c['line']=a[:a.index(old)].count('\n')+1
 else:
  start=a.index('func whitespaceShards(');body=a.index('{',start);end=a.index('\nfunc TestPrinterWhitespacePlanted',body);b=a[:body+1]+'\n\treturn nil\n}\n'+a[end:];c['line']=a[:start].count('\n')+1
 (p/(c['id']+'.diff')).write_text(diff(f,a,b))
(p/'check-plan.json').write_text(json.dumps(checks,indent=2))
plans=json.loads((p/'plan.json').read_text())+ [{'id':'P01','file':'internal/lower/lower.go'}]+checks
res=[]
rows=json.loads((p/'scope.json').read_text())['rows'];regex='^('+'|'.join(rows)+')$'
for c in plans:
 mid=c['id'];d=tmp/mid;d.mkdir(exist_ok=True)
 replacements={}
 for f in ['internal/lower/class.go','internal/lower/lower.go','stage1/cohere/graphql/printer/gaps_test.go','stage1/cohere/graphql/printer/shards_test.go']:
  a=base(f);b=a
  if f==c['file']:
   target=d/'applied';target.write_text(a)
   # apply standalone diff in temporary tree using git apply, independent of switch
   q=d/f;q.parent.mkdir(parents=True,exist_ok=True);q.write_text(a)
   z=subprocess.run(['git','apply','--unsafe-paths','--directory='+str(d),str(p/(mid+'.diff'))],capture_output=True,text=True);assert z.returncode==0,z.stderr
   b=q.read_text()
  out=d/(f.replace('/','_'));out.write_text(b);replacements[str(root/f)]=str(out)
 stub=d/'selector.go';stub.write_text('package lower\n');replacements[str(root/'internal/lower/audit_selector.go')]=str(stub)
 overlay=d/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 cmd=['timeout','90','go','vet','-overlay',str(overlay),'./internal/lower/' if mid.startswith('M') or mid=='P01' else './stage1/cohere/graphql/printer/']
 st=time.monotonic()
 with (p/(mid+'-verify.log')).open('w') as log:z=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 r=dict(id=mid,vet_status=z.returncode,vet_wall=time.monotonic()-st,vet_command=cmd)
 z=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],capture_output=True,text=True);r['apply_status']=z.returncode;r['apply_error']=z.stderr
 if mid in [x['id'] for x in checks]:
  cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',regex];st=time.monotonic()
  with (p/(mid+'.log')).open('w') as log:z=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  r.update(status=z.returncode,wall=time.monotonic()-st,command=cmd,failed_rows=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') in rows],seconds=next((e.get('Elapsed') for e in reversed(events) if e.get('Action') in ['pass','fail'] and 'Test' not in e),None))
 res.append(r);(p/'checks.json').write_text(json.dumps(res,indent=2))
