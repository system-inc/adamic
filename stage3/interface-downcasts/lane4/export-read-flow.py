#!/usr/bin/env python3
"""Join lane 3 graph evidence to the stock-checker read inventory by exact span."""
import gzip,hashlib,json,pathlib,sys
root=pathlib.Path(sys.argv[1]);raw=pathlib.Path(sys.argv[2]);lane=pathlib.Path(sys.argv[3])
measured=json.loads(raw.read_bytes());queries={q['key']:q for q in measured['read_receivers']}
rows=json.loads(gzip.decompress((lane/'read-demand-sites.json.gz').read_bytes()));source={};matched=0;known=0;overlap=0
positions={}
for row in rows:positions.setdefault(row['file'],set()).update([row['start'],row['end']])
byte_offsets={}
for name,spans in positions.items():
 data=(root/name).read_bytes().decode('utf-8').encode('utf-16-le');source[name]=data
 previous=0;byte=0;mapping={}
 for position in sorted(spans):
  byte+=len(data[2*previous:2*position].decode('utf-16-le').encode());mapping[position]=byte;previous=position
 byte_offsets[name]=mapping
for row in rows:
 if row['file'] not in source:source[row['file']]=(root/row['file']).read_bytes().decode('utf-8').encode('utf-16-le')
 data=source[row['file']]
 start=byte_offsets[row['file']][row['start']];end=byte_offsets[row['file']][row['end']]
 key=f"{row['file']}:{start}:{end}";query=queries.get(key)
 if query:
  matched+=1;known+=not query['graph_unknown'];overlap+=bool(query['viewed_allocation_overlap'])
  row['shared_graph']={k:v for k,v in query.items() if k!='key'}
  assert query['retain_check'],'unexpected erasure in unknown-view corpus'
  row['flow_reason']='Shared graph queried; Unknown cast frontier retains tag/read check even for known receiver allocations'
 else:
  row['shared_graph']={'unmapped':True,'retain_check':True}
  row['flow_reason']='No shared-graph query for this syntax (including destructuring); Unknown retains tag/read check'
(lane/'read-demand-sites.json.gz').write_bytes(gzip.compress((json.dumps(rows)+'\n').encode(),mtime=0))
counts={'source_commit':'050880ce59e30b356b686bd3144efe24f875ebc8','shared_flow_commit':'c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b','measurement':'shared graph on checker-rejected unadapted program; conservative Unknown view reachability','diagnostics':len(measured['diagnostics']),'allocation_shapes':measured['allocation_shapes'],'graph_queries':len(queries),'graph_unknown_queries':sum(q['graph_unknown'] for q in queries.values()),'graph_known_queries':sum(not q['graph_unknown'] for q in queries.values()),'queries_overlapping_viewed_allocations':sum(bool(q['viewed_allocation_overlap']) for q in queries.values()),'inventory_reads':len(rows),'matched_inventory_receivers':matched,'unmapped_inventory_receivers':len(rows)-matched,'matched_known_allocation_receivers':known,'matched_overlapping_viewed_allocations':overlap,'retained_inventory_read_checks':len(rows),'proven_non_viewed_receivers':0,'casts':measured['counts'],'cast_unknown_reasons':measured['unknown_reasons'],'source_hashes':{name:hashlib.sha256((root/name).read_bytes()).hexdigest() for name in sorted(source)}}
(lane/'read-flow-summary.json').write_text(json.dumps(counts,indent=2)+'\n')
demand=json.loads((lane/'read-demand-summary.json').read_text())
demand['flow_dependency']='Reuses lane 3 c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b ReachingAllocations through make-read-flow-overlay.py; Unknown view propagation retains checks'
demand['shared_flow']={key:counts[key] for key in ['graph_queries','matched_inventory_receivers','unmapped_inventory_receivers','matched_known_allocation_receivers','matched_overlapping_viewed_allocations','proven_non_viewed_receivers']}
(lane/'read-demand-summary.json').write_text(json.dumps(demand,indent=2)+'\n')
with raw.open('rb') as reader,(lane/'read-flow-result.json.gz').open('wb') as output:
 with gzip.GzipFile(filename='',mode='wb',fileobj=output,mtime=0) as writer:
  while chunk:=reader.read(1024*1024):writer.write(chunk)
print(json.dumps(counts,indent=2))
