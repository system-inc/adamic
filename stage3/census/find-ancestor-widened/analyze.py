import json,gzip,re,collections,html,hashlib,itertools
from pathlib import Path
out=Path(__file__).resolve().parent
sites=json.loads((out/'sites.json').read_text());root=Path('/workspace/scratch/fa-adapted/src');cache={}
def pos(file,line,col):
 if file not in cache:
  offsets=[0]
  for text in (root/file).read_bytes().decode().splitlines(keepends=True):offsets.append(offsets[-1]+len(text.encode('utf-16-le'))//2)
  cache[file]=offsets
 return cache[file][line-1]+col-1
results=[]
for side in ['parent','head']:
 selected=[];rows=[[] for s in sites]
 with gzip.open(out/(side+'-stderr.txt.gz'),'rt') as f:
  for number,line in enumerate(f,1):
   if not line.startswith('probe-widened\t'):continue
   parts=line.rstrip('\n').split('\t');assert len(parts)>=4,(number,parts)
   m=re.fullmatch(r'/workspace/scratch/fa-adapted/src/(.*):(\d+):(\d+)',parts[2])
   if not m:continue
   for index,s in enumerate(sites):
    if m[1]!=s['file']:continue
    offset=pos(s['file'],int(m[2]),int(m[3]))
    if s['start']<=offset<s['end']:
     row=dict(stderr_line=number,line=line.rstrip('\n'));rows[index].append(row);selected.append(dict(site=s['site'],**row))
 with gzip.open(out/(side+'-filtered-probes.txt.gz'),'wt') as f:
  for row in selected:f.write(row['line']+'\n')
 (out/(side+'-filtered-index.json')).write_text(json.dumps(selected,indent=2)+'\n');results.append(rows)
def fields(line):return dict(p.split('=',1) for p in line.split('\t')[3:] if '=' in p)
def typed(fields):
 return {k:(m[1],m[2]) for k,v in fields.items() if (m:=re.fullmatch(r'#(\d+) (.*)',v))}
def key(line):return tuple(line.split('\t')[1:3])
comparisons=[]
for i,s in enumerate(sites):
 p,h=[[r['line'] for r in side[i]] for side in results];evidence=[]
 for side,seq in [('parent',p),('head',h)]:
  for line in seq:
   t=typed(fields(line))
   for a,b in itertools.combinations(t,2):
    if t[a][1]==t[b][1] and t[a][0]!=t[b][0]:evidence.append(dict(kind='same text, different id within line',side=side,fields=[a,b],line=line))
 # Pair corresponding records by stage and location, preserving occurrence order.
 groups=[]
 for seq in [p,h]:
  g=collections.defaultdict(list)
  for line in seq:g[key(line)].append(line)
  groups.append(g)
 for k in sorted(groups[0].keys()|groups[1].keys()):
  for a,b in zip(groups[0][k],groups[1][k]):
   fa,fb=fields(a),fields(b);ta,tb=typed(fa),typed(fb)
   if 'same' in fa and 'same' in fb and fa['same']!=fb['same']:evidence.append(dict(kind='same changed',parent=a,head=b))
   for field in sorted(ta.keys()&tb.keys()):
    if ta[field][1]==tb[field][1] and ta[field][0]!=tb[field][0]:evidence.append(dict(kind='same text, different id',field=field,parent=a,head=b))
 if not p and not h:verdict='no line'
 elif evidence:verdict='identity'
 else:
  def inputs(line):return '\t'.join(x for x in line.split('\t') if not x.startswith(('err=','assignable=')))
  common=[(a,b) for k in sorted(groups[0].keys()&groups[1].keys()) for a,b in zip(groups[0][k],groups[1][k])]
  matching_inputs=bool(common) and all(inputs(a)==inputs(b) for a,b in common)
  different_results=any(a!=b for a,b in common)
  verdict='further down' if matching_inputs and different_results else ('unchanged' if p==h else 'unclassified')
 comparisons.append(dict(**s,parent=p,head=h,verdict=verdict,identity_evidence=evidence,sequence_same=p==h))
(out/'comparison.json').write_text(json.dumps(comparisons,indent=2)+'\n')
v={s:json.loads((out/(s+'-stream-validation.json')).read_text()) for s in ['parent','head']}
assert all(x['identical'] and x['raw_stream_identical'] for x in v.values()),v
parts=['# Widened probe census','','Task 448a51k. Parent `34035b64` on `555b1645`; head `183fc95b` on `8afae944`. Report only.','','## Method','','Read `440777b5` stage3/census/find-ancestor-probe/REPORT.md and its list, plus c6f6688c\'s findings-only lists. Both probe SHAs passed `go vet ./internal/lower` before census builds; exact logs and child exit codes are in the vet files. One separate checkout of `1060360b` received both complete adaptation directories `99-correctness-core-or-predicate` and `99-correctness-core-multimap` from `eeb87728`, then ran `stage3/apply.sh` against TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. All adapted source hashes match c6f6688c\'s source-hashes.json.','','The own-pin guarded latent and entry overlays, shared unchanged checker, entry root `/workspace/scratch/fa-adapted/src/tsc/tsc.ts`, `LATENT_ASSERT_NO_OUTPUT=1`, `GOMEMLIMIT=5GiB`, `GOGC=70` and separate `LATENT_TIMES` files follow the first probe exactly. `ADAMIC_PROBE_BINDINGS=1` is retained from the first probe and `ADAMIC_PROBE_WIDENED=all` enables the new probe. Neither `LATENT_FULL` nor `LATENT_ONLY_FILE` is set. Both no-output guards remain active. Controllers preserve child exits, stop after 1800 seconds without output or unit progress, and clean process groups.','','The same stock TypeScript AST recipe selects each innermost enclosing call, including nested probe coordinates, with exclusive end bounds. The two hook assignments retain their bounded assignment expressions. The same three unchanged append controls are selected. `sites.json` contains source and bounds; `sites.cjs` reproduces selection. Both whole stderr streams and filtered probe lines are gzipped. Filtered indexes retain stderr line numbers and duplicates. All quoted lines below are verbatim. Table rows align by decision stage, coordinate, and occurrence; the filtered files preserve complete stderr order and duplicates. A further-down verdict means all corresponding input fields match and a result field changes; the changed result can suppress or expose later decision lines, which are also quoted.','','## Finding stream validation','','| Side | Vet exit | Census exit | Findings | Entire JSONL byte-identical to c6f6688c |','| --- | ---: | ---: | ---: | --- |']
for side in ['parent','head']:
 vet=json.loads((out/(side+'-vet.status.json')).read_text());status=json.loads((out/(side+'-entry-run.status.json')).read_text());assert vet['exit']==status['exit']==0
 parts.append(f'| {side} | {vet["exit"]} | {status["exit"]} | {v[side]["actual_findings"]} | {v[side]["raw_stream_identical"]} |')
parts+=['','## Site comparison','']
for i,s in enumerate(comparisons,1):
 parts += [f'### {i}. {s["site"]} ({s["category"]})','',f'Expression: `{s["start_where"]}` through `{s["end_where"]}` (exclusive).','', '<table><thead><tr><th>Parent</th><th>Head</th></tr></thead><tbody>']
 groups=[]
 for seq in [s['parent'],s['head']]:
  g=collections.defaultdict(list)
  for line in seq:g[key(line)].append(line)
  groups.append(g)
 keys=list(dict.fromkeys(key(line) for line in s['parent']+s['head']))
 aligned=[pair for k in keys for pair in itertools.zip_longest(groups[0][k],groups[1][k])]
 for p,h in aligned:
  def cell(line):return '<pre>'+html.escape(line).replace('|','&#124;')+'</pre>' if line is not None else 'No corresponding probe-widened line.'
  parts.append('<tr><td>'+cell(p)+'</td><td>'+cell(h)+'</td></tr>')
 if not s['parent'] and not s['head']:parts.append('<tr><td>No probe-widened line.</td><td>No probe-widened line.</td></tr>')
 parts+=['</tbody></table>','', '**Verdict: '+s['verdict']+'.**','']
counts=collections.Counter(s['verdict'] for s in comparisons)
summary='; '.join(f'{k}: {n} sites' for k,n in counts.items())+'. '
for verdict in counts:
 summary+=verdict.capitalize()+' holds at '+', '.join(s['site'] for s in comparisons if s['verdict']==verdict)+'. '
summary=summary.strip()
parts+=['## Verdict','',summary,'']
(out/'REPORT.md').write_text('\n'.join(parts));(out/'closing-paragraph.txt').write_text(summary+'\n')
print(json.dumps(dict(counts=counts,sites=len(sites)),indent=2))
