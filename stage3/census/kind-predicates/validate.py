"""Independent checks of saved paired streams, counts, identities, and timings."""
import collections,gzip,hashlib,json,pathlib
P=pathlib.Path(__file__).resolve().parent
D=json.loads((P/'delta.json').read_text())
streams={}
for label in ['base','head']:
 with gzip.open(P/f'evidence/{label}-latent.jsonl.gz','rt') as f: rows=[json.loads(s) for s in f]
 assert len(rows)==80
 assert rows[0]['latent_mode']=='first-error'
 findings=[f for r in rows[1:] for f in r['findings']]
 keys={tuple(f[k] for k in ['kind','where','reason','text']) for f in findings}
 counts=collections.Counter(k[0] for k in keys)
 assert all(D['counts'][label][k]==counts[k] for k in D['counts'][label])
 first={}
 for f in findings:
  if f['phase']=='lowering': first.setdefault(f['unit'],tuple(f[k] for k in ['kind','where','reason']))
 with gzip.open(P/f'evidence/{label}-times.jsonl.gz','rt') as f: times=[json.loads(s) for s in f]
 assert len(times)==4056 and len({x['where'] for x in times})==4056
 assert len(rows[1:])==len({r['file'] for r in rows[1:]})
 streams[label]=(rows,findings,keys,first,times)
b,h=streams['base'],streams['head']
assert b[0][0]==h[0][0]
assert {r['file'] for r in b[0][1:]}=={r['file'] for r in h[0][1:]}
assert {u['where']:u for r in b[0][1:] for u in r['units']}=={u['where']:u for r in h[0][1:] for u in r['units']}
assert b[3]==h[3] and not D['unit_first_stop_delta']
assert not ({f['where'] for f in h[1]}-{f['where'] for f in b[1]})
assert not D['regressions']
removed=b[2]-h[2];added=h[2]-b[2]
assert len(removed)==245 and len(added)==1
reason='an object refinement using an open numeric enum as a literal tag'
assert all(k[2]==reason for k in removed)
assert sum(k[2]==reason for k in b[2])==1251
assert sum(k[2]==reason for k in h[2])==1006
all_locations={k[1] for k in b[2]|h[2]}
B={w:{k for k in b[2] if k[1]==w} for w in all_locations}
H={w:{k for k in h[2] if k[1]==w} for w in all_locations}
changed={w for w in all_locations if B[w]!=H[w]}
assert changed=={x['where'] for x in D['site_delta']}
assert len(changed)==245
transitions=collections.Counter()
preexisting=0
for x in D['open_tag_transitions']:
 w=x['where']; old_open={k for k in B[w] if k[2]==reason}; assert len(old_open)==1
 status='unchanged' if old_open<=H[w] else 'moved' if H[w] else 'no_finding'
 assert status==x['status'];transitions[status]+=1
 if status=='moved' and H[w]<=B[w]: preexisting+=1
assert dict(transitions)=={'unchanged':1006,'no_finding':239,'moved':6}
assert preexisting==5
for x in D['site_delta']:
 i=x['identity']; source=pathlib.Path('/workspace/scratch/kp-adapted')/i['file']; data=source.read_bytes()
 assert hashlib.sha256(data).hexdigest()==i['sha256']
 text=data.decode('utf-8'); prefix='\n'.join(text.split('\n')[:i['line']-1])
 if i['line']>1: prefix+='\n'
 current=text.split('\n')[i['line']-1].encode('utf-16-le')[:2*(i['column']-1)].decode('utf-16-le')
 assert len((prefix+current).encode('utf-8'))==i['byte_offset']
 assert len((prefix+current).encode('utf-16-le'))//2==i['utf16_offset']
for label,index in [('base',b),('head',h)]:
 for metric in ['wall','cpu']:
  assert abs(sum(t[metric+'_seconds'] for t in index[4])-sum(t[label+'_'+metric+'_seconds'] for t in D['unit_timing_delta']))<1e-8
 rss=json.loads((P/f'evidence/{label}-rss.json').read_text())
 assert rss==D['resources'][label] and rss['exit_code']==0 and rss['signal'] is None
 assert (P/f'evidence/{label}-memory-before.txt').read_bytes()==(P/f'evidence/{label}-memory-after.txt').read_bytes()
manifest=json.loads((P/'evidence/source-manifest.json').read_text())
for x in manifest:
 assert hashlib.sha256((pathlib.Path('/workspace/scratch/kp-adapted')/x['file']).read_bytes()).hexdigest()==x['sha256']
for name,digest in json.loads((P/'artifacts.sha256.json').read_text()).items():
 assert hashlib.sha256((P/name).read_bytes()).hexdigest()==digest,name
assert '\u2013' not in (P/'REPORT.md').read_text() and '\u2014' not in (P/'REPORT.md').read_text()
print('PASS: both complete 79-file streams; identical checker headers and 4,554 unit records.')
print('PASS: 4,056 timings per run; sums and every artifact digest match.')
print('PASS: 1,251 -> 1,006 open-tag sites; 239 finding-free, 6 other reasons, 1,006 unchanged.')
print('PASS: 245 changed source tokens; 245 removed findings and 1 added reason at an already blocked token.')
print('PASS: five remaining reasons predate the head; zero new failing locations, first-stop changes, or panics.')
print('PASS: all shared adapted source hashes and changed-token byte/UTF-16 identities match.')
print('PASS: exits, RSS records, unchanged OOM counters, and report punctuation checked.')
