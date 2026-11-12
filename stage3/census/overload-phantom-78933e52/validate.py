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
a=json.loads((out/'unit-analysis.json').read_text());assert len(a['getPathComponents'])==11 and len(a['getSupportedExtensions'])==5
prov=json.loads((out/'provenance.json').read_text());c=json.loads((out/'composition.json').read_text());assert prov['base']==c['base'] and prov['head']==c['head'] and prov['composition']==c
assert c['main']=='0c7052759671347775615af5b4a8028dd8688af8' and c['path']=='a3f67b011b1a810c282da390f80a3180df1be34b'
assert c['step2']=='3a7ea290fca9e3f3325ffa8206a861b22f441583' and c['step2b']=='694834e95c3dd79bc4d350f4560d5a16b175f0ad' and c['step3']=='78933e52b0abd0da2fce91e73b4b71c1b36dc4e4'
for file,sha in prov['go_evidence_sha256'].items():assert hashlib.sha256((out/file).read_bytes()).hexdigest()==sha
interface=json.loads((out/'path-components-interface.json').read_text());assert interface['source_sha256']==manifest[interface['source']]
for side in ['base','head']:
 for variant in ['', '-scan']:
  fixture=json.loads((out/(side+variant+'-fixture-validation.json')).read_text());assert fixture['file_unit_finding_ledgers_equal'] and fixture['cross_shard_generic_witnesses_merged']
for name,d in json.loads((out/'archive-manifest.json').read_text()).items():
 p=out/name;data=gzip.decompress(p.read_bytes()) if p.suffix=='.gz' else p.read_bytes();assert len(data)==d['bytes'] and hashlib.sha256(data).hexdigest()==d['sha256']
supervisors=json.loads((out/'supervisor-results.json').read_text());assert len(supervisors)==16 and all(r['census_child_exit']==r['timing_wrapper_exit']==r['outer_shell_exit']==0 for r in supervisors)
assert not list(out.rglob('*.go'))
results.update(headers_equal=True,requested_unit_group_sizes=[11,5],compositions_and_pins_verified=True,path_interface_present=True,fixture_composition_both_sides_pass=True,archive_hashes_verified=True,no_dot_go_evidence=True)
(out/'validation.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results,indent=2))
