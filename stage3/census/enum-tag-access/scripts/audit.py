import json,pathlib,collections,math
R=pathlib.Path('/workspace/scratch'); d=json.loads((R/'comparison.json').read_text()); s=d['summary']
raw={x:[json.loads(y) for y in (R/f'{x}-results/latent.jsonl').read_text().splitlines()] for x in ['base','head']}
for side,rows in raw.items():
 assert rows[0]['latent_mode']=='first-error'; assert len(rows)==80
 counts=collections.Counter(); unique=set(); first={}
 for row in rows[1:]:
  for f in row['findings']:
   key=(f['kind'],f['where'],f['reason'],f['text'])
   if key not in unique: counts[f['kind']]+=1; unique.add(key)
   if f['phase']=='lowering': first.setdefault(f['unit'],(f['kind'],f['where'],f['reason'],f['text']))
 assert dict(counts)==s[side+'_totals']
 timing=[json.loads(y) for y in (R/f'{side}-results/times.jsonl').read_text().splitlines()]
 assert len(timing)==len({x['where'] for x in timing})==4056
 for key in ['wall_seconds','cpu_seconds']: assert math.isclose(sum(t[key] for t in timing),s['timing'][side][key],rel_tol=0,abs_tol=1e-8)
 raw[side]=(rows,first)
def flattened(side): return [f for r in raw[side][0][1:] for f in r['findings']]
bf,hf=flattened('base'),flattened('head')
bw,hw={f['where'] for f in bf},{f['where'] for f in hf}
assert len(hw-bw)==s['site_regressions']==54
reason='an object refinement using an open numeric enum as a literal tag'
bopen={f['where'] for f in bf if f['reason']==reason}; hopen={f['where'] for f in hf if f['reason']==reason}
assert len(bopen-hopen)==74 and len(hopen-bopen)==s['new_open_findings']==55
assert not any(f['kind']=='panic' for f in bf+hf)
assert s['base_open']==1006
assert sum(s['open_outcomes'].values())==1006
assert sum(f['reason']=='an object refinement using an open numeric enum as a literal tag' for r in raw['head'][0][1:] for f in r['findings'])==s['head_open']
b,h=raw['base'][1],raw['head'][1]
assert sum(b.get(k)!=h.get(k) for k in b.keys()|h.keys())==s['moved_units']
for row in d['base_open_sites']:
 file,sha,offset=row['identity']; where_file,line,col=row['where'].rsplit(':',2)
 assert file==where_file
 data=(R/'adapted'/file).read_bytes(); prefix=data[:offset].decode('utf-8'); lines=prefix.split('\n')
 assert len(lines)==int(line)
 assert len(lines[-1].encode('utf-16-le'))//2+1==int(col)
 assert row['outcome']=='unchanged' or not any(f['reason']==row['base']['reason'] for f in row['head'])
a=json.loads((R/'accessor-summary.json').read_text()); assert a['matched_findings']>=a['unique_sites']
assert len(set(a['sites']))==a['unique_sites']
for name,n in a['by_name'].items(): assert len({r['where'] for r in a['trace'] if r['name']==name})==n
for file in ['REPORT.md','DELTA.json']:
 text=(R/'report/stage3/census/enum-tag-access'/file).read_text(); assert '\u2013' not in text and '\u2014' not in text
print('PASS: raw totals, 79-file coverage, 4056 unique timed units per side, time sums, 1006-site partition, independent first-stop accounting, source-byte/UTF-16 position round trips, accessor attribution, punctuation.')
