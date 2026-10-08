#!/usr/bin/env python3
"""Independent bounded audit/export of per-file byte scores. No source execution.
Raw receipts and byte verification remain the evidence; exports keep every cell,
nonzero finding count, output digest, observed refusal and original divergence.
"""
import argparse,base64,collections,gzip,hashlib,json,pathlib
p=argparse.ArgumentParser();p.add_argument('manifest');p.add_argument('receipts');p.add_argument('out');p.add_argument('--corrections');a=p.parse_args()
root=pathlib.Path(__file__).resolve().parents[3];source=pathlib.Path(a.receipts);out=pathlib.Path(a.out);out.mkdir(parents=True,exist_ok=True)
with gzip.open(a.manifest,'rt') if a.manifest.endswith('.gz') else open(a.manifest) as f:m=json.load(f)
if m['dependency_inputs']!={'go':'uninstalled','node':'uninstalled'}:raise SystemExit('this run is explicitly dependency-uninstalled')
descriptors=[]
for f in (root/'stage1/cohere/lint/rules').glob('*/rule.json'):
 d=json.loads(f.read_text());descriptors.append(d)
descriptors.sort(key=lambda d:(d.get('order',1000000),d['name']))
port=[d['name'] for d in descriptors];families={d['name']:d['upstreamPackage'] for d in descriptors}
corrections={};verified=0
if a.corrections:
 for f in pathlib.Path(a.corrections).glob('verified-*.gz'):
  receipt=json.loads(pathlib.Path(str(f)+'.receipt.json').read_text());verified+=receipt['files']
  for line in gzip.open(f,'rt'):
   r=json.loads(line);corrections[r['index']]=r
per_rule={};per_family={};causes=collections.Counter();seen=set();node_findings=0;bytes_total=0;lint_files=0;divergence_count=0

def stdout(e):return base64.b64decode(e['stdout_bytes']) if e.get('stdout_bytes') else e.get('stdout','').encode('utf-8')
def answers(e):
 data=stdout(e);pieces=data.split(b'\n');lines=[v+b'\n' for v in pieces[:-1]]+[pieces[-1]];result=collections.defaultdict(bytes);active=None
 for i,line in enumerate(lines):
  if line.startswith((b'skipped ',b'refused ')):
   fields=line.split(b' ',2)
   if len(fields)>1:result[fields[1].decode()]+=line
   active=None;continue
  if i+1<len(lines) and lines[i+1].startswith(b'  '):active=lines[i+1][2:].split(b'  ',1)[0].decode()
  if line.startswith((b'fixed\t',b'case ',b'rejected ')):active=None
  if active:result[active]+=line
 return result

def state(x,y,fmt=False):
 for e in [x,y]:
  if e.get('error') or e.get('exit_code',0)!=0:
   cause='execution fault'
   if 'invalid corpus' in x.get('error','') or 'typescript/parser' in y.get('stack','') or 'typescript/scanner' in y.get('stack','') or y.get('phase')=='parser':cause='parser'
   return 'blocked',cause
 xx,yy=stdout(x),stdout(y)
 if not fmt and any(line.startswith(b'skipped ') for line in xx.split(b'\n')+yy.split(b'\n')):return 'blocked','checker fact'
 if xx.startswith(b'error\t') or yy.startswith((b'error\t',b'refused ')):return 'blocked','formatter refusal'
 return ('agree','') if xx==yy else ('diverge','')

def bump(target,key,s,c,g,n):
 t=target.setdefault(key,dict(checks=0,agree=0,diverge=0,blocked=0,go_findings=0,node_findings=0,blocked_causes={},oracle_census_statuses={},oracle_finding_files=0));t['checks']+=1;t[s]+=1;t['go_findings']+=g;t['node_findings']+=n;t['oracle_finding_files']+=int(g>0)
 if c:t['blocked_causes'][c]=t['blocked_causes'].get(c,0)+1

def add(name,family,s,c,g=0,n=0):
 bump(per_rule,name,s,c,g,n);bump(per_family,family,s,c,g,n)
 if c:causes[c]+=1

def fixes(e):
 e=e.copy();e['stdout_bytes']=base64.b64encode(b''.join(line+b'\n' for line in stdout(e).split(b'\n') if line.startswith((b'fixed\t',b'rejected ',b'unconverged\t')))).decode();e['stdout']=''
 if b'fixed\t' not in stdout(e) and not e.get('error'):e['error']='missing fixed-source protocol record'
 return e

def brief(e):
 return dict(stdout_sha256=hashlib.sha256(stdout(e)).hexdigest(),stdout_bytes=len(stdout(e)),exit_code=e.get('exit_code'),elapsed_ns=e.get('elapsed_ns'),wall_elapsed_ns=e.get('wall_elapsed_ns'),phase=e.get('phase'),error=e.get('error','')[:400],stack=e.get('stack','')[:1800])
letters={'agree':'A','diverge':'D'};blocked={'checker fact':'C','parser':'P','formatter refusal':'F','unported rule':'U','execution fault':'E'}
files=sorted(source.glob('receipts-*.jsonl.gz'))
for chunk in files:
 archive=out/('files-'+chunk.name);divfile=out/('divergences-'+chunk.name)
 with gzip.open(archive,'wt') as target,gzip.open(divfile,'wt') as divergent:
  for line in gzip.open(chunk,'rt'):
   r=json.loads(line);i=r['index']
   if i in seen or i>=len(m['files']) or m['files'][i]!=r['path']:raise SystemExit('duplicate or substituted input '+str(i))
   seen.add(i);bytes_total+=r['bytes'];correction=corrections.get(i)
   if correction:
    if correction['source_sha256']!=r['source_sha256']:raise SystemExit('correction source differs')
    r['go_lint']=correction.get('go_lint');r['go_format']=correction['go_format']
   compact=dict(index=i,path=r['path'],source_sha256=r['source_sha256'],bytes=r['bytes']);diff=[]
   if bool(r.get('go_lint')) != (pathlib.Path(r['path']).suffix in {'.ts','.tsx','.js','.jsx'}):raise SystemExit('lint eligibility differs at '+str(i))
   if r.get('go_lint'):
    census=r['census']
    if len(census)!=492 or len({row['rule'] for row in census})!=492:raise SystemExit('oracle registry census cardinality differs at '+str(i))
    if any(row['files']!=int(row['findings']>0) or row['findings']<0 for row in census):raise SystemExit('oracle finding count differs at '+str(i))
    lint_files+=1;x,y=r['go_lint'],r['node_lint'];xx,yy=answers(x),answers(y);states='';counts={}
    for name in port:
     u,v=x.copy(),y.copy();u['stdout_bytes']=base64.b64encode(xx[name]).decode();v['stdout_bytes']=base64.b64encode(yy[name]).decode();u['stdout']='';v['stdout']='';s,c=state(u,v);g=xx[name].count(b'\nrange ');n=yy[name].count(b'\nrange ');add(name,families[name],s,c,g,n);states+=letters.get(s,blocked.get(c,'E'))
     if g or n:counts[name]=[g,n]
     if s=='diverge':diff.append(name)
    s,c=state(x,y);add('lint/all-fixes','fixes',s,c);compact['all_fixes']=letters.get(s,blocked.get(c,'E'))
    fx,fy=fixes(x),fixes(y);s,c=state(fx,fy);add('lint/syntax-fixes','fixes',s,c)
    if s=='diverge':diff.append('lint/syntax-fixes')
    compact.update(lint_states=states,findings=counts,go_lint=brief(x),node_lint=brief(y),syntax_fixes=letters.get(s,blocked.get(c,'E')))
    census_counts={};statuses=collections.Counter()
    for row in r['census']:
     name=row['rule']
     if name in families:continue
     family='typescript' if name.startswith('@typescript-eslint/') else 'next' if name.startswith('@next/next/') else name.split('/')[0] if '/' in name else 'core'
     add(name,family,'blocked','unported rule',row['findings']);t=per_rule[name];t['oracle_census_statuses'][row['status']]=t['oracle_census_statuses'].get(row['status'],0)+1;statuses[row['status']]+=1
     if row['findings']:census_counts[name]=row['findings']
    compact.update(unported_findings=census_counts,unported_oracle_statuses=dict(statuses))
   family='typescript' if pathlib.Path(r['path']).suffix in {'.ts','.tsx','.js','.jsx'} else {'.md':'markdown','.json':'json','.yaml':'yaml','.css':'css'}[pathlib.Path(r['path']).suffix]
   x,y=r['go_format'],r['node_format'];s,c=state(x,y,True);add('format/'+family,'format',s,c);compact.update(format_state=letters.get(s,blocked.get(c,'E')),go_format=brief(x),node_format=brief(y))
   if s=='diverge':diff.append('format/'+family)
   compact['divergences']=diff;target.write(json.dumps(compact,ensure_ascii=True,separators=(',',':'))+'\n')
   if diff:
    r['divergences']=diff;r.pop('census',None);divergent.write(json.dumps(r,ensure_ascii=True,separators=(',',':'))+'\n');divergence_count+=len(diff)
 print('audited',len(seen),'divergences',divergence_count,flush=True)
complete=len(seen)==len(m['files']);missing=sorted((dict(rule=name,findings=t['go_findings'],files=t['oracle_finding_files'],status='complete' if set(t['oracle_census_statuses'])=={'complete'} else 'incomplete') for name,t in per_rule.items() if name not in families and not name.startswith(('format/','lint/'))),key=lambda r:(-r['findings'],r['rule']))
totals={k:sum(t[k] for t in per_rule.values()) for k in ['checks','agree','diverge','blocked','go_findings','node_findings']}
engine=json.loads((source/'summary.json').read_text())
if engine['port_rule_order']!=port:raise SystemExit('registry rule order differs from measured receipt')
summary=dict(engine_pins={'source_base':engine['base'],'cohere':engine['cohere'],'node':'source port'},total=totals,complete=complete,dependency_inputs=m['dependency_inputs'],files=len(seen),expected_files=len(m['files']),lint_files=lint_files,bytes=bytes_total,port_rule_order=port,per_rule=per_rule,per_family=per_family,blocked_cells_by_cause=dict(causes),unported_ranked=missing,divergences=divergence_count,oracle_byte_verified_files=verified,oracle_corrected_files=len(corrections),missing_indices=sorted(set(range(len(m['files'])))-seen))
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
if not complete:raise SystemExit('incomplete audit; missing indices explicitly reported')
if a.corrections and verified!=len(m['files']):raise SystemExit('oracle byte verification incomplete')
