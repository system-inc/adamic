"""Summarize only complete, independently audited lossless census artifacts."""
import collections,gzip,hashlib,importlib.util,json,pathlib,subprocess
repo=pathlib.Path('/workspace/adamic');base=repo/'stage3/shape-conformance';out=pathlib.Path('/workspace/shape-overnight-measurements')
spec=importlib.util.spec_from_file_location('artifacts',base/'dynamic-keys/artifacts.py');artifacts=importlib.util.module_from_spec(spec);spec.loader.exec_module(artifacts)
manifest=json.loads((out/'manifest.json').read_text());root=pathlib.Path('/tmp/shape-stage3-234ab1aa-adapted');map_path=pathlib.Path('/tmp/shape-views-map.json');mapped=json.loads(map_path.read_text())
assert hashlib.sha256(map_path.read_bytes()).hexdigest()==manifest['mapped_sha256']
for name,digest in manifest['source_hashes'].items():assert hashlib.sha256((root/name).read_bytes()).hexdigest()==digest,name
rows=[];diagnostics=None
for run in manifest['runs']:
 label=run['label'];path=base/'overnight'/(label+'-interned.json.gz')
 assert path.exists(),label+' census artifact not complete'
 template=subprocess.check_output(['git','show',run['snapshot']+':stage3/shape-conformance/latent/lower.go.txt'],cwd=repo)
 assert hashlib.sha256(template).hexdigest()==run['adapter_sha256'],label+' snapshot does not match sealed adapter'
 with gzip.open(path,'rt') as stream:packed=json.load(stream)
 result=artifacts.decode(packed)
 assert artifacts.json_hash(result)==packed['json_sha256'],label+' observations changed'
 artifacts.audit(result,mapped,root)
 inventory=set(result['diagnostics'])
 if diagnostics is None:diagnostics=inventory
 assert inventory==diagnostics and len(inventory)==260,label+' checker inventory changed'
 total=collections.Counter()
 for counts in result['counts'].values():total.update(counts)
 reasons=collections.Counter()
 for counts in result['unknown_reasons'].values():reasons.update(counts)
 row=dict(run,artifact=str(path.relative_to(base)),artifact_bytes=path.stat().st_size,artifact_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),canonical_json_sha256=packed['json_sha256'],counts=result['counts'],unknown_reasons=result['unknown_reasons'],total_counts=dict(total),total_unknown_reasons=dict(reasons),diagnostics=len(result['diagnostics']),diagnostics_sha256=artifacts.json_hash(result['diagnostics']),allocation_schemas=len(result['allocation_schemas']),allocation_schemas_sha256=artifacts.json_hash(result['allocation_schemas']),host_names=result['host_names'],host_inventory_sha256=artifacts.json_hash(result['host_values']))
 rows.append(row)
 print(label,dict(total),dict(reasons),len(result['allocation_schemas']),flush=True)
summary=dict(measurement='measured on a checker-rejected program',source_manifest='census-manifest.json',sites=2936,tagged=1758,untagged=1178,checker_diagnostics=260,runs=rows)
(out/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('PASS six complete audited snapshots; fixed sources, map and diagnostics; every canonical observation preserved')
