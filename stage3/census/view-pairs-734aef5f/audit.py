"""Independent membership and grouping audit of the selected raw probe rows."""
import collections,gzip,hashlib,json,re,subprocess
from pathlib import Path
here=Path(__file__).resolve().parent;raw=here/'raw'
def read(p):
 if not p.exists():return []
 return [json.loads(l) for l in gzip.decompress(p.read_bytes()).decode().splitlines() if l]
summary=json.loads((here/'summary.json').read_text());selected=read(here/'selected-pairs.jsonl.gz');units=read(here/'selected-units.jsonl.gz')
passes=sorted([p.name.removesuffix('-boundary.jsonl.gz') for p in raw.glob('*-boundary.jsonl.gz')],key=lambda n:(n!='first',n))
chosen={};raw_pairs={};unit_maps={}
for name in passes:
 for r in read(raw/(name+'-boundary.jsonl.gz')):chosen[r['file'],r['unit']]=name
 raw_pairs[name]=read(raw/(name+'-pairs.jsonl.gz'))
 unit_maps[name]={(r['file'],r['unit']):r for r in read(raw/(name+'-units.jsonl.gz'))}
seen=set();grouped=collections.defaultdict(list);counts=collections.Counter()
for r in selected:
 name,index=r['pass'],r['pass_index'];key=(r['file'],r['unit']);assert chosen[key]==name
 assert r['unit'].startswith(r['file']+':') and len(r['path'] or [])==r['depth']
 if r['depth']:assert r['path'][-1]==r['field']
 assert key in unit_maps[name]
 assert (name,index) not in seen;seen.add((name,index))
 assert {k:v for k,v in r.items() if k not in ('pass','pass_index')}==raw_pairs[name][index]
 if r['source_nil']:kind='nil_source'
 elif r['target_nil']:kind='nil_target_only'
 elif r['object_pair']:
  related=r['intersection_related'] if r['source_intersection'] or r['target_intersection'] else r['related']
  kind='object_'+('related' if related else 'unrelated')+('_top' if r['depth']==0 else '_nested')
 else:kind='scalar_top' if r['depth']==0 else 'scalar_nested'
 counts[kind]+=1
 if kind.startswith('object_unrelated'):grouped[r['source'],r['target'],tuple(r['path'] or [])].append(r)
expected={(name,i) for name,rs in raw_pairs.items() for i,r in enumerate(rs) if chosen.get((r['file'],r['unit']))==name and (r['file'],r['unit']) in unit_maps[name]}
assert seen==expected
assert summary['total_raw_pairs']==sum(map(len,raw_pairs.values()))
assert summary['total_selected_pairs']==len(selected)==sum(r['pairs_logged'] for r in units)
for key,value in summary['counts'].items():assert counts[key]==value
pairs=json.loads((here/'unrelated-object-pairs.json').read_text());assert len(pairs)==len(grouped)==summary['distinct_unrelated_text_pairs']
for pair in pairs:
 key=pair['source'],pair['target'],tuple(pair['path']);rows=grouped[key]
 assert pair['count']==len(rows)
 actual=collections.Counter((r['file'],r['unit']) for r in rows)
 assert {(u['file'],u['unit']):u['count'] for u in pair['units']}==dict(actual)
 assert pair['files']==sorted({r['file'] for r in rows})
assert sum(p['count'] for p in pairs)==summary['unrelated_object_pairs']
nils=json.loads((here/'nil-source-reference-views.json').read_text());expected_nils=[r for r in selected if r['source_nil'] and r.get('target_reference_fields')]
assert nils==expected_nils and len(nils)==summary['nil_source_with_reference_field']
for u in units:assert not u['probe_error']
manifest=json.loads((here/'source-manifest.json').read_text())
for name,digest in manifest.items():assert hashlib.sha256((Path('/workspace/view-pairs-adapted')/name).read_bytes()).hexdigest()==digest
root=here.parents[2]
assert not (root/'internal/lower/zz_view_pairs_probe.go').exists()
changed=subprocess.check_output(['git','diff','--name-only','734aef5f'],cwd=root,text=True).splitlines();assert all(n.startswith('stage3/census/view-pairs-734aef5f/') for n in changed)
result={'raw_event_membership_passed':True,'selected_population_complete':True,'exact_unrelated_groups_and_units_passed':True,'nil_reference_population_passed':True,'source_hashes_passed':True,'production_probe_absent':True,'scope_passed':True,'selected_pairs':len(selected),'unrelated_object_pairs':summary['unrelated_object_pairs'],'nil_source_with_reference_field':len(nils)}
(here/'independent-validation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
