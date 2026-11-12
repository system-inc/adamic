import pathlib,json,gzip,hashlib,collections,re
from merge import merge,key
out=pathlib.Path(__file__).resolve().parent

def read(p):
 with gzip.open(p,'rt') as f:return [json.loads(s) for s in f if s.strip()]
shards=[read(out/'main'/f'shard-{i}.jsonl.gz') for i in range(8)];rows=read(out/'main/latent.jsonl.gz');assert merge(shards)==rows
scans=lambda stream:{key(f) for r in stream for f in r.get('findings',[]) if f['phase']=='refusal_scan'}
assert scans(shards[0]) and scans(shards[0])==scans(shards[1]) and all(not scans(r) for r in shards[2:])
assert rows[0]['status']=='measurement' and rows[0]['latent_mode']=='first-error' and not rows[0]['checker_rejected']
assert set(rows[0]['sources'])=={r['file'] for r in rows[1:-1]}
units=[u for r in rows[1:-1] for u in r['units']];attempted=[u for u in units if u['status'] in ('attempted','panic')];membership=collections.defaultdict(list)
for i,stream in enumerate(shards):
 for r in stream[1:-1]:
  for u in r['units']:membership[u['where']].append(i)
for ordinal,u in enumerate(attempted):assert membership[u['where']]==[ordinal%8]
for u in units:
 if u['status'] not in ('attempted','panic'):assert membership[u['where']]==list(range(8))
timings=[]
for i in range(8):
 timings+=read(out/'main'/f'shard-{i}-times.jsonl.gz');d=json.loads((out/'main'/f'shard-{i}.time.json').read_text());assert d['exit_code']==0 and not d['timed_out'] and d['wall_seconds']<1800
assert collections.Counter(t['where'] for t in timings)==collections.Counter(u['where'] for u in attempted)
inv=json.loads((out/'main-inventory.json').read_text());assert inv['total_units']==len(units) and all(r['all_units_present_in_source_order'] for r in inv['files'])
pins=json.loads((out/'pins.json').read_text());prov=json.loads((out/'provenance.json').read_text());assert pins['main']==prov['main']=='219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205' and prov['landed']==pins['landed']
manifest=json.loads((out/'adapted-source-hashes.json').read_text());ids=json.loads((out/'normalization.json').read_text());assert all(manifest[x['file']]==x['sha256'] for x in ids.values())
summary=json.loads((out/'summary.json').read_text());assert summary['units']==len(units)==len({u['where'] for u in units}) and summary['attempted']==len(attempted)
ledger=json.loads((out/'overload-findings.json').read_text());um={u['where']:u for u in units};first={};findings=[]
for r in rows[1:]:
 for f in r.get('findings',[]):
  findings.append(f)
  if f['phase'] in ('lowering','deferred') and f['unit'] in um:first.setdefault(f['unit'],f)
assert summary['no_first_stop_units']==sum(u['status']=='attempted' and u['where'] not in first for u in units)
for kind in ['result','parameter']:
 pattern=re.compile(r'^overload \d+ of (\S+) '+kind+r'(?:\s|$)');matching=[f for f in findings if pattern.match(f['reason'])]
 assert [e['finding'] for e in ledger[kind]['observations']]==matching
 raw={};distinct={}
 for f in matching:
  x=ids[f['where']];fn=pattern.match(f['reason']).group(1);site=(x['file'],x['sha256']);raw[(*site,x['byte_offset'],f['kind'],f['reason'],f['text'])]=fn;distinct[(*site,x['read_offset'],f['reason'])]=fn
 observations=collections.Counter(pattern.match(f['reason']).group(1) for f in matching);raw_counts=collections.Counter(raw.values());distinct_counts=collections.Counter(distinct.values());first_counts=collections.Counter(pattern.match(f['reason']).group(1) for f in first.values() if pattern.match(f['reason']))
 c=summary['categories'][kind];assert c['totals']==dict(observations=len(matching),raw=len(raw),distinct=len(distinct),first_stop_units=sum(first_counts.values()))
 assert len(ledger[kind]['raw'])==len(raw) and len(ledger[kind]['distinct'])==len(distinct)
 for name,v in c['functions'].items():assert (v['observations'],v['raw'],v['distinct'],v['first_stop_units'])==(observations[name],raw_counts[name],distinct_counts[name],first_counts[name])
 for view in ['raw','distinct']:
  indices=[i for entry in ledger[kind][view] for i in entry['observations']];assert sorted(indices)==list(range(len(matching)))
for variant in ['', '-scan']:
 fixture=json.loads((out/('main'+variant+'-fixture-validation.json')).read_text());assert fixture['file_unit_finding_ledgers_equal'] and fixture['cross_shard_generic_witnesses_merged']
for name,d in json.loads((out/'archive-manifest.json').read_text()).items():
 p=out/name;data=gzip.decompress(p.read_bytes()) if p.suffix=='.gz' else p.read_bytes();assert len(data)==d['bytes'] and hashlib.sha256(data).hexdigest()==d['sha256']
supervisors=json.loads((out/'supervisor-results.json').read_text());assert len(supervisors)==8 and all(r['census_child_exit']==r['timing_wrapper_exit']==r['outer_shell_exit']==0 for r in supervisors)
assert not list(out.rglob('*.go'))
result=dict(main=pins['main'],sources=len(rows[0]['sources']),units=len(units),attempted=len(attempted),complete_disjoint_shards=True,complete_unit_timings=True,stock_unit_catalog_complete=True,archive_recomposition_exact=True,archive_hashes_verified=True,normalization_hashes_verified=True,all_matching_observations_enumerated=True,raw_and_distinct_counts_recomputed=True,first_stop_counts_recomputed=True,fixtures_both_variants_pass=True,all_child_and_wrapper_exits_zero=True,all_census_children_below_30_minutes=True,no_dot_go_evidence=True)
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
