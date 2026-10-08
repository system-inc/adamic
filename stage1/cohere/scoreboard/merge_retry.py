#!/usr/bin/env python3
"""Add one independently rerun missing input without replacing any raw receipt.
The preliminary stream counters are not the canonical score; audit_sweep rebuilds
all scores from the now complete receipt set, including the retry.
"""
import argparse,gzip,hashlib,json,pathlib,re,shutil
p=argparse.ArgumentParser();p.add_argument('manifest');p.add_argument('receipts');p.add_argument('retry');p.add_argument('index',type=int);a=p.parse_args()
source=pathlib.Path(a.receipts);retry=pathlib.Path(a.retry);m=json.load(gzip.open(a.manifest,'rt'));expected=m['files'][a.index]
status=json.loads((retry/'summary.json').read_text())
if not status['complete'] or status['files']!=1 or status['dependency_inputs']!=m['dependency_inputs']:raise SystemExit('retry is incomplete or has different dependency input')
rows=[]
for f in retry.glob('receipts-*.jsonl.gz'):
 rows.extend(json.loads(line) for line in gzip.open(f,'rt'))
if len(rows)!=1 or rows[0]['index']!=0 or rows[0]['path']!=expected:raise SystemExit('retry source differs')
r=rows[0]
if hashlib.sha256(pathlib.Path(expected).read_bytes()).hexdigest()!=r['source_sha256']:raise SystemExit('retry source hash differs')
seen=set();prefix=re.compile(rb'^\{"index":(\d+),"path":("(?:[^"\\]|\\.)*")')
for f in sorted(source.glob('receipts-*.jsonl.gz')):
 for line in gzip.open(f,'rb'):
  match=prefix.match(line)
  if not match:raise SystemExit('receipt prefix differs '+str(f))
  index=int(match[1]);path=json.loads(match[2])
  if index in seen or m['files'][index]!=path:raise SystemExit('duplicate or substituted input '+str(index))
  seen.add(index)
if set(range(len(m['files'])))-seen!={a.index}:raise SystemExit('missing set is not exactly the retry index')
r['index']=a.index;r['retry_scope']={'reason':'first full-tree census exceeded 30-second request budget','request_limit':'5m','go_memory_limit':'4GiB','node_heap_limit':'512MiB','attempt_directory':str(retry)}
summary=source/'summary.json';old=json.loads(summary.read_text())
if old['cohere']!=status['cohere'] or old['port_rule_order']!=status['port_rule_order']:raise SystemExit('retry engine pin/registry differs')
dest=source/('receipts-retry-'+str(a.index)+'.jsonl.gz')
with gzip.open(str(dest)+'.tmp','wt') as target:target.write(json.dumps(r,separators=(',',':'))+'\n')
pathlib.Path(str(dest)+'.tmp').rename(dest)
shutil.copyfile(summary,source/'summary-before-retry.json')
old.update(complete=True,files=len(m['files']),lint_files=old['lint_files']+1,preliminary_counters_superseded=True,retry_index=a.index)
summary.write_text(json.dumps(old,indent=2)+'\n')
(source/'retry-receipt.json').write_text(json.dumps(dict(index=a.index,path=expected,source_sha256=r['source_sha256'],files_verified=len(seen)+1,raw_receipt=str(dest),reason=r['retry_scope']),indent=2)+'\n')
print('receipt set complete:',len(seen)+1,'inputs; retry index',a.index)
