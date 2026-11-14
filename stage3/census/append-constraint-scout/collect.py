"""Validate independent file coverage and merge raw records without changing finding text."""
import gzip,hashlib,json,pathlib,sys
folder,adapted,out=map(pathlib.Path,sys.argv[1:]);root=adapted/'src/compiler';out.mkdir(parents=True,exist_ok=True)
manifest=json.loads((folder/'manifest.json').read_text());results=json.loads((folder/'results.json').read_text());expected=set(manifest['source_sha256']);assert len(expected)==79
assert {r['file'] for r in results}==expected
rows=[];headers=[];unique=set()
for result in sorted(results,key=lambda r:r['file']):
 assert result['exit']==0 and result['complete_trailer'] and result['file_records']==1 and result['units']==0,result
 source=root/result['file'];assert hashlib.sha256(source.read_bytes()).hexdigest()==manifest['source_sha256'][result['file']]
 raw=(folder/(result['file'].replace('/','__')+'.jsonl')).read_text();records=[json.loads(x) for x in raw.splitlines()]
 assert records[-1]['status']=='relation_report';headers.append(records[0])
 own=[r for r in records if 'file' in r];assert len(own)==1 and own[0]['file']==str(source) and own[0]['units']==[]
 for row in records:
  for f in row.get('findings',[]):unique.add((f['kind'],f['where'],f['reason'],f['text']))
 rows.extend(records)
text=''.join(json.dumps(r,separators=(',',':'))+'\n' for r in rows)
(out/'scan-all.jsonl').write_text(text)
(out/'scan-all.jsonl.gz').write_bytes(gzip.compress(text.encode(),mtime=0))
(out/'coverage.json').write_text(json.dumps({'manifest':manifest,'summary':json.loads((folder/'summary.json').read_text()),'files':results,'unique_findings':len(unique),'checker_rejected_values':sorted(set(r['checker_rejected'] for r in headers))},indent=2)+'\n')
print(len(results),len(unique))
