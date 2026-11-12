"""Final census evidence checks; no compiler or adaptation changes."""
import pathlib,json,gzip,hashlib,collections
from merge import merge,load,key
out=pathlib.Path(__file__).resolve().parent

def read(p):
 with gzip.open(p,'rt') as f:return [json.loads(s) for s in f if s.strip()]
results={}
streams={}
for side in ['base','head']:
 shards=[read(out/side/f'shard-{i}.jsonl.gz') for i in range(8)]
 rows=read(out/side/'latent.jsonl.gz');assert merge(shards)==rows
 scans=lambda shard:{key(f) for r in shard for f in r.get('findings',[]) if f['phase']=='refusal_scan'}
 assert scans(shards[0])==scans(shards[1]), 'repeated full scans disagree'
 assert all(not scans(s) for s in shards[2:]), 'later shard unexpectedly scanned'
 assert rows[0]['status']=='measurement' and not rows[0]['checker_rejected'] and rows[0]['latent_mode']=='first-error'
 assert set(rows[0]['sources'])=={r['file'] for r in rows[1:-1]}
 units=[u for r in rows[1:-1] for u in r['units']]
 attempted=[u for u in units if u['status'] in ('attempted','panic')]
 memberships=collections.defaultdict(list)
 for i,shard in enumerate(shards):
  for r in shard[1:-1]:
   for u in r['units']:memberships[u['where']].append(i)
 for ordinal,u in enumerate(attempted):assert memberships[u['where']]==[ordinal%8],(side,u,memberships[u['where']])
 for u in units:
  if u['status'] not in ('attempted','panic'):assert memberships[u['where']]==list(range(8))
 timings=[]
 for i in range(8):
  timings+=read(out/side/f'shard-{i}-times.jsonl.gz')
  d=json.loads((out/side/f'shard-{i}.time.json').read_text());assert d['exit_code']==0 and not d['timed_out'] and d['wall_seconds']<1800
 assert collections.Counter(t['where'] for t in timings)==collections.Counter(u['where'] for u in attempted)
 inventory=json.loads((out/(side+'-inventory.json')).read_text());assert inventory['total_units']==len(units)
 assert all(r['all_units_present_in_source_order'] for r in inventory['files'])
 results[side]=dict(sources=len(rows[0]['sources']),units=len(units),attempted=len(attempted),status_counts=dict(collections.Counter(u['status'] for u in units)),unit_shards_disjoint_and_complete=True,per_unit_timings_complete=True,independent_stock_catalog_complete=True,raw_archived_stream_equals_remerge=True,all_child_exits_zero_and_below_30_minutes=True)
 streams[side]=rows
assert streams['base'][0]==streams['head'][0]
delta=json.loads((out/'delta.json').read_text());manifest=json.loads((out/'adapted-source-hashes.json').read_text())
assert all(manifest[name]==sha for name,sha in delta['source_hashes'].items())
a=json.loads((out/'unit-analysis.json').read_text());assert len(a['concatenate'])==5 and len(a['getSupportedExtensions'])==5 and len(a['writeTokenText'])==1 and len(a['historical_43'])==43
prov=json.loads((out/'provenance.json').read_text());assert prov['base']=='17c6378de039f08283edea73ee75cfd4ba32a55a' and prov['head']=='3a7ea290fca9e3f3325ffa8206a861b22f441583'
assert hashlib.sha256((out/'overload-proof-head.go.txt').read_bytes()).hexdigest()==prov['head_proof_sha256']
nb=json.loads((out/'concatenate-native-base.json').read_text());nh=json.loads((out/'concatenate-native-head.json').read_text());assert nb==nh and nb['checker_diagnostics']==0
stock=json.loads((out/'concatenate.json').read_text());assert len(stock['results'])==5
for r in stock['results']:
 assert len(r['calls'])==1
 c=r['calls'][0];n=next(x for x in nb['calls'] if x['where']==c['where'] and x['unit']==r['unit']['where'])
 assert n['index']==c['resolved_declaration']['index'] and n['declaration_text']==c['resolved_declaration']['text']
for side in ['base','head']:
 for variant in ['', '-scan']:
  fixture=json.loads((out/(side+variant+'-fixture-validation.json')).read_text());assert fixture['file_unit_finding_ledgers_equal'] and fixture['cross_shard_generic_witnesses_merged']
for name,d in json.loads((out/'archive-manifest.json').read_text()).items():
 p=out/name
 data=gzip.decompress(p.read_bytes()) if p.suffix=='.gz' else p.read_bytes()
 assert len(data)==d['bytes'] and hashlib.sha256(data).hexdigest()==d['sha256']
supervisors=json.loads((out/'supervisor-results.json').read_text());assert len(supervisors)==16
assert all(r['census_child_exit']==r['timing_wrapper_exit']==r['outer_shell_exit']==0 for r in supervisors)
excerpt=json.loads((out/'get-supported-extensions-source.json').read_text());assert excerpt['source_sha256']==manifest[excerpt['source']]
assert not list(out.rglob('*.go'))
results.update(headers_equal=True,requested_unit_group_sizes=[5,5,1],historical_cohort_size=43,native_checker_diagnostics=0,native_resolutions_equal_and_stock_agrees=True,fixture_composition_both_sides_pass=True,archive_hashes_verified=True,no_dot_go_evidence=True)
(out/'validation.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results,indent=2))
