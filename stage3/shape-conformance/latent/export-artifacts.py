"""Package the observed census, preserving the original ledger identities."""
import collections,gzip,hashlib,json,pathlib,sys
result_path,map_path,control_path,adapted_root=map(pathlib.Path,sys.argv[1:])
r=json.loads(result_path.read_text());mapped=json.loads(map_path.read_text());control=json.loads(control_path.read_text())
root=pathlib.Path('stage3/shape-conformance');label=r['measurement']
def write(name,value):
 (root/name).write_text(json.dumps(value,indent=2)+'\n')
def compressed(name,text):
 with (root/name).open('wb') as raw:
  with gzip.GzipFile(fileobj=raw,mode='wb',mtime=0,filename='') as output:output.write(text.encode())
for name,value in [('latent-share.json.gz',r),('latent-site-map.json.gz',mapped),('latent-controls.json.gz',control)]:compressed(name,json.dumps(value,separators=(',',':'))+'\n')
ledger_path=pathlib.Path('stage3/interface-downcasts/shape-conformance-sites.json');baseline=json.loads(ledger_path.read_text());lookup={(s['file'],s['start'],s['end']):s for s in baseline}
own=collections.Counter(s['kind'] for s in r['sites'] if s['detail']==['containing function body has checker diagnostics; body skipped'])
summary={k:r[k] for k in ('measurement','checker_rejected','counts','unknown_reasons','host_values','allocation_shapes','limits')}
summary.update(diagnostic_count=len(r['diagnostics']),total_sites=len(r['sites']),source_hashes={str(p.relative_to(adapted_root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((adapted_root/'src/compiler').rglob('*.ts'))},original_ledger_sha256=hashlib.sha256(ledger_path.read_bytes()).hexdigest(),typed_body_skip_sites=dict(own),diagnosed_dependency_sites={k:r['unknown_reasons'][k].get('diagnosed body',0)-own[k] for k in r['counts']})
write('latent-share-summary.json',summary)
frontiers={kind:collections.Counter() for kind in r['counts']}
for s in r['sites']:
 if s['reason']=="flow the graph can't see":
  frontiers[s['kind']].update(set(d.split(' at /tmp/')[0] for d in s['detail']))
write('latent-flow-reasons.json',{'measurement':label,'overlapping_detail_counts':{k:dict(v.most_common()) for k,v in frontiers.items()}})
write('latent-host-values.json',{'measurement':label,'cast_related_host_frontiers':[{k:s[k] for k in ('file','line','column','kind','target_type','reason','host_values','detail')} for s in r['sites'] if s['host_values']],'additional_native_fixture':{'file':'stage3/interface-downcasts/lane3/host.a','value':'readTextFile(...)','fields':{'kind':'string','message':'string'},'observed':'view retained; native stops at kind with unsupported representation; source Node prints true','included_in_2936_sites':False}})
rows=[]
for s in r['sites']:
 old=lookup[(s['file'],s['start'],s['end'])]
 row={'measurement':label,**{k:s[k] for k in ('file','start','end','line','column','kind')},'before':{k:old[k] for k in ('outcome','reason')},'after':s}
 rows.append(json.dumps(row,separators=(',',':')))
compressed('site-comparison.jsonl.gz','\n'.join(rows)+'\n')
write('site-comparison-summary.json',{'measurement':label,'sites':len(rows),'before_certified_free':0,'after_certified_free':sum(c['conforms and ready (free)'] for c in r['counts'].values()),'after_counts':r['counts'],'after_unknown_reasons':r['unknown_reasons'],'note':'Proof coverage measured on checker-rejected program; diagnosed bodies skipped. Unknown is not proof of nonconformance.'})
print(json.dumps({k:summary[k] for k in ('measurement','counts','unknown_reasons','allocation_shapes','typed_body_skip_sites','diagnosed_dependency_sites','host_values')},indent=2))
