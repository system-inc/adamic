import gzip,hashlib,json,pathlib,subprocess
out=pathlib.Path(__file__).parent;root=pathlib.Path('/workspace/overload-adapted')
j=lambda p:json.loads(p.read_text())
expected=j(out/'source-hashes.json')
actual={p.relative_to(root).as_posix():hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
assert actual==expected,'adapted source changed'
results={'source_files_unchanged':len(actual),'runs':{},'archives':[]}
for name in ['debug-parent','debug-head','entry-parent','entry-head']:
 status=j(out/(name+'.status.json'));assert status['child_exit']==0 and status['stopped'] is None,(name,status)
 for r in j(out/(name+'-streams.json')):
  p=out/r['file'];data=p.read_bytes();raw=gzip.decompress(data)
  assert hashlib.sha256(data).hexdigest()==r['compressed_sha256'] and len(data)==r['compressed_bytes']
  assert hashlib.sha256(raw).hexdigest()==r['sha256'] and len(raw)==r['bytes']
  results['archives'].append(r['file'])
 with gzip.open(out/(name+'.jsonl.gz'),'rt') as f:rows=[json.loads(x) for x in f]
 assert rows[0]['status']=='measurement' and not rows[0]['checker_rejected'] and not rows[0]['diagnostics']
 assert rows[-1]['status']=='relation_report','missing complete trailer'
 results['runs'][name]={'rows':len(rows),'sources':len(rows[0]['sources']),'units':sum(len(r.get('units',[])) for r in rows),'raw_events':sum(len(r.get('findings',[])) for r in rows),'complete_trailer':True}
for name in ['debug','entry']:
 d=j(out/(name+'-paired.json'));assert d['normalization'] is True
 for source,sha in d['source_hashes'].items():assert expected[source]==sha,(source,sha)
 results[name+'_pair_source_hashes']=len(d['source_hashes'])
for label,sha in [('parent','1d84074f'),('head','8fabe4c2')]:
 tree='/workspace/overload-census' if label=='parent' else '/workspace/overload-head'
 p=subprocess.run(['git','-C',tree,'diff','--quiet',sha,'--','internal','stage3/adapt','stage3/census/latent','stage3/census/pair','stage3/meter']);assert p.returncode==0
 results[label+'_production_clean']=True
assert j(out/'pair-tests.status.json')['child_exit']==0
report=(out/'REPORT.md').read_text();assert '\u2013' not in report and '\u2014' not in report
assert not list(out.glob('*.go')),'Go evidence must use .go.txt'
assert not list(out.glob('*.jsonl')),'unarchived stream remains'
(out/'verification.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
